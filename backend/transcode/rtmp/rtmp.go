package rtmp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sen1or/letslive/transcode/config"
	livestreamdto "sen1or/letslive/transcode/gateway/livestream/dto"
	livestreamgateway "sen1or/letslive/transcode/gateway/livestream/http"
	usergateway "sen1or/letslive/transcode/gateway/user/http"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/transcode/live"
	"sen1or/letslive/transcode/storage"
	"strconv"
	"strings"
	"time"

	"github.com/nareix/joy5/format/flv/flvio"
	"github.com/nareix/joy5/format/rtmp"
)

// sessionCloseTimeout bounds Session.Close: the recorder's Finish may take up
// to 60s, plus room to stop ffmpeg and the thumbnailer.
const sessionCloseTimeout = 75 * time.Second

// startupOffsetSeconds is subtracted from the wall-clock stream duration to
// account for the time before the first keyframe starts the recording.
const startupOffsetSeconds = 7

// RTMPServerConfig.Registry is the service discovery registry; LiveRegistry
// is where live streams are published for the web server.
type RTMPServerConfig struct {
	Context      context.Context
	Port         int
	Registry     *discovery.Registry
	LiveRegistry *live.Registry
	Storage      storage.Storage
	Config       config.Config
}

type RTMPServer struct {
	ctx               context.Context
	Port              int
	Registry          *discovery.Registry
	userGateway       *usergateway.UserGateway
	livestreamGateway *livestreamgateway.LivestreamGateway
	config            config.Config
	liveRegistry      *live.Registry
	storage           storage.Storage
	listener          net.Listener
}

func NewRTMPServer(config RTMPServerConfig, userGateway *usergateway.UserGateway, livestreamgateway *livestreamgateway.LivestreamGateway) *RTMPServer {
	return &RTMPServer{
		ctx:               config.Context,
		Port:              config.Port,
		Registry:          config.Registry,
		config:            config.Config,
		userGateway:       userGateway,
		livestreamGateway: livestreamgateway,
		liveRegistry:      config.LiveRegistry,
		storage:           config.Storage,
	}
}

func (s *RTMPServer) Start() {
	portStr := strconv.Itoa(s.Port)
	server := rtmp.NewServer()

	serverAddr, err := net.ResolveTCPAddr("tcp", ":"+portStr)
	if err != nil {
		logger.Panicf(s.ctx, "failed to resolve rtmp addr: %s", err)
	}

	s.listener, err = net.ListenTCP("tcp", serverAddr)
	if err != nil {
		logger.Panicf(s.ctx, "rtmp failed to listen: %s", err)
	}
	logger.Infow(s.ctx, "rtmp server started")

	server.LogEvent = func(c *rtmp.Conn, nc net.Conn, e int) {
		es := rtmp.EventString[e]
		logger.Debugf(s.ctx, "RTMP log event: %s", es)
	}

	server.HandleConn = s.HandleConnection

	for {
		conn, err := s.listener.Accept()

		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				logger.Infow(s.ctx, "RTMP listener closed, shutting down accept loop.")
				return
			}

			select {
			case <-s.ctx.Done():
				logger.Infow(s.ctx, "context cancelled - shutting down rtmp accept loop.")
				return
			default:
				logger.Errorf(s.ctx, "rtmp failed to accept connection: %s", err)
				time.Sleep(1 * time.Second) // Prevent fast spinning on unexpected errors
				continue
			}

		}

		logger.Debugf(s.ctx, "RTMP connection accepted from: %s", conn.RemoteAddr())
		// Launch a goroutine to handle the connection using the rtmp library's handler
		go server.HandleNetConn(conn)
	}
}

// TODO: check if on disconnect do we need to manually close nc
func (s *RTMPServer) HandleConnection(c *rtmp.Conn, nc net.Conn) {
	c.LogTagEvent = func(isRead bool, t flvio.Tag) {
		if t.Type == flvio.TAG_AMF0 {
			logger.Infof(s.ctx, "RTMP log tag: %+v", t.DebugFields())
		}
	}

	streamingKeyComponents := strings.Split(c.URL.Path, "/")
	streamingKey := streamingKeyComponents[len(streamingKeyComponents)-1]

	streamId, _, err := s.onConnect(streamingKey) // get stream id

	if err != nil {
		logger.Errorf(s.ctx, "stream connection failed: %s", err)
		nc.Close()
		return
	}

	startTime := time.Now()
	session := live.NewSession(s.ctx, streamId, live.SessionDeps{
		Registry:  s.liveRegistry,
		Storage:   s.storage,
		Transcode: s.config.Transcode,
	})

	s.readPackets(c, session, streamId)

	// Close first so the VOD master exists before the playback URL is published.
	closeCtx, cancelClose := context.WithTimeout(context.WithoutCancel(s.ctx), sessionCloseTimeout)
	if err := session.Close(closeCtx); err != nil {
		logger.Errorf(s.ctx, "failed to close livestream %s session: %v", streamId, err)
	}
	cancelClose()

	duration := int64(math.Ceil(time.Since(startTime).Seconds())) - startupOffsetSeconds
	s.onDisconnect(streamId, max(duration, 0))
}

// readPackets feeds the connection's packets to session until the connection
// ends or fails.
func (s *RTMPServer) readPackets(c *rtmp.Conn, session *live.Session, streamId string) {
	for {
		pkt, err := c.ReadPacket()
		if err != nil {
			if err != io.EOF {
				logger.Errorf(s.ctx, "failed to read rtmp packet of livestream %s: %v", streamId, err)
			}
			return
		}

		if err := session.HandlePacket(pkt); err != nil {
			if errors.Is(err, live.ErrSessionClosed) {
				return
			}
			logger.Errorf(s.ctx, "failed to handle rtmp packet of livestream %s: %v", streamId, err)
		}
	}
}

// check if stream api key exists
// then update the user status to online
// return the stream id to be used as publishName
func (s *RTMPServer) onConnect(streamingKey string) (streamId string, userId string, err error) {
	reqCtx, reqCtxCancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer reqCtxCancel()

	userInfo, errRes := s.userGateway.GetUserInformation(reqCtx, streamingKey)
	if errRes != nil {
		return "", "", fmt.Errorf("failed to get user information: %s", errRes.Message)
	}

	thumb := userInfo.Data.LivestreamInformationResponseDTO.ThumbnailURL
	// empty string is invalid
	if thumb != nil && *thumb == "" {
		thumb = nil
	}

	streamDTO := &livestreamdto.CreateLivestreamRequestDTO{
		Title:        userInfo.Data.LivestreamInformationResponseDTO.Title,
		UserId:       userInfo.Data.Id,
		Description:  userInfo.Data.LivestreamInformationResponseDTO.Description,
		ThumbnailURL: thumb,
		Visibility:   "public", // TODO: add to livestream information instead of default to public
	}

	req2Ctx, req2CtxCancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer req2CtxCancel()

	createdLivestream, createErrRes := s.livestreamGateway.Create(req2Ctx, *streamDTO)
	if createErrRes != nil {
		return "", "", fmt.Errorf("failed to create livestream: %s", createErrRes.Message)
	}

	livestreamId := createdLivestream.Id.String()

	return livestreamId, userInfo.Data.Id.String(), nil
}

func (s *RTMPServer) onDisconnect(streamId string, duration int64) {
	playbackURL := fmt.Sprintf("%s/%s/index.m3u8", s.config.VODPlaybackUrlPrefix, streamId)
	endDTO := &livestreamdto.EndLivestreamRequestDTO{
		PlaybackURL: &playbackURL,
		EndedAt:     time.Now(),
		Duration:    duration,
	}

	reqCtx, reqCtxCancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer reqCtxCancel()

	createErrRes := s.livestreamGateway.EndLivestream(reqCtx, streamId, *endDTO)
	if createErrRes != nil {
		logger.Errorf(s.ctx, "failed to end livestream: %s", createErrRes.Message)
	}
}

func (s *RTMPServer) Shutdown(ctx context.Context) error {
	listener := s.listener
	s.listener = nil // prevent further use

	if listener == nil {
		logger.Warnf(ctx, "RTMP server listener already closed or not started.")
		return nil
	}

	logger.Infow(ctx, "shutting down RTMP server listener...")

	// close the listener, this will cause the accept loop in Start() to unblock
	err := listener.Close()

	if err != nil && !errors.Is(err, net.ErrClosed) {
		logger.Errorf(ctx, "error closing RTMP listener: %v", err)
		return err
	}

	logger.Infow(ctx, "RTMP server listener closed.")

	// Optionally: Add waiting logic here if you need to ensure active connections are finished.
	// This usually involves tracking active connections and waiting for them to complete.
	// The provided code doesn't track connections actively in the RTMPServer struct itself,
	// relying on the rtmp library and individual handler goroutines.

	return nil
}
