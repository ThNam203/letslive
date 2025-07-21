package rtmp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"path/filepath"
	"sen1or/letslive/transcode/config"
	livestreamdto "sen1or/letslive/transcode/gateway/livestream/dto"
	livestreamgateway "sen1or/letslive/transcode/gateway/livestream/http"
	usergateway "sen1or/letslive/transcode/gateway/user/http"
	"sen1or/letslive/transcode/pkg/discovery"
	"sen1or/letslive/transcode/pkg/logger"
	"sen1or/letslive/transcode/transcoder"
	"sen1or/letslive/transcode/watcher"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nareix/joy5/format/flv"
	"github.com/nareix/joy5/format/flv/flvio"
	"github.com/nareix/joy5/format/rtmp"
)

type RTMPServerConfig struct {
	Context    context.Context
	Port       int
	Registry   *discovery.Registry
	Config     config.Config
	VODHandler watcher.VODHandler
}

type RTMPServer struct {
	ctx               context.Context
	Port              int
	Registry          *discovery.Registry
	userGateway       *usergateway.UserGateway
	livestreamGateway *livestreamgateway.LivestreamGateway
	config            config.Config
	vodHandler        watcher.VODHandler
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
		vodHandler:        config.VODHandler,
	}
}

func (s *RTMPServer) Start() {
	portStr := strconv.Itoa(s.Port)
	server := rtmp.NewServer()

	serverAddr, err := net.ResolveTCPAddr("tcp", ":"+portStr)
	if err != nil {
		logger.Panicf("failed to resolve rtmp addr: %s", err)
	}

	s.listener, err = net.ListenTCP("tcp", serverAddr)
	if err != nil {
		logger.Panicf("rtmp failed to listen: %s", err)
	}
	logger.Infow("rtmp server started")

	server.LogEvent = func(c *rtmp.Conn, nc net.Conn, e int) {
		es := rtmp.EventString[e]
		logger.Debugf("RTMP log event: %s", es)
	}

	server.HandleConn = s.HandleConnection

	for {
		conn, err := s.listener.Accept()

		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				logger.Infow("RTMP listener closed, shutting down accept loop.")
				return
			}

			select {
			case <-s.ctx.Done():
				logger.Infow("context cancelled - shutting down rtmp accept loop.")
				return
			default:
				logger.Errorf("rtmp failed to accept connection: %s", err)
				time.Sleep(1 * time.Second) // Prevent fast spinning on unexpected errors
				continue
			}

		}

		logger.Debugf("RTMP connection accepted from: %s", conn.RemoteAddr())
		// Launch a goroutine to handle the connection using the rtmp library's handler
		go server.HandleNetConn(conn)
	}
}

// TODO: check if on disconnect do we need to manually close nc
func (s *RTMPServer) HandleConnection(c *rtmp.Conn, nc net.Conn) {
	c.LogTagEvent = func(isRead bool, t flvio.Tag) {
		if t.Type == flvio.TAG_AMF0 {
			logger.Infof("RTMP log tag: %+v", t.DebugFields())
		}
	}

	streamingKeyComponents := strings.Split(c.URL.Path, "/")
	streamingKey := streamingKeyComponents[len(streamingKeyComponents)-1]

	streamId, _, err := s.onConnect(streamingKey) // get stream id

	if err != nil {
		logger.Errorf("stream connection failed: %s", err)
		nc.Close()
		return
	}

	pipeOut, pipeIn := io.Pipe()

	var startTime time.Time
	var startTimer = func() {
		startTime = time.Now()
	}

	transcoder := transcoder.NewTranscoder(pipeOut, s.config.Transcode, startTimer)
	defer transcoder.Stop()
	go func() {
		transcoder.Start(streamId)
	}()

	w := flv.NewMuxer(pipeIn)

	for {
		pkt, err := c.ReadPacket()
		if err == io.EOF {
			duration := int64(math.Ceil(time.Now().Sub(startTime).Seconds()) - 7) // TODO: proper duration calculation
			pipeOut.Close()
			pipeIn.Close()
			s.onDisconnect(streamId, duration)
			return
		}

		if err := w.WritePacket(pkt); err != nil {
			logger.Errorf("failed to write rtmp package: %s", err)
			duration := int64(math.Ceil(time.Now().Sub(startTime).Seconds()) - 7)
			pipeIn.Close()
			pipeOut.Close()
			s.onDisconnect(streamId, duration)
			return
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

	streamDTO := &livestreamdto.CreateLivestreamRequestDTO{
		Title:        userInfo.LivestreamInformationResponseDTO.Title,
		UserId:       userInfo.Id,
		Description:  userInfo.LivestreamInformationResponseDTO.Description,
		ThumbnailURL: userInfo.LivestreamInformationResponseDTO.ThumbnailURL,
		Visibility:   "public", // TODO: add to livestream information instead of default to public
	}

	req2Ctx, req2CtxCancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer req2CtxCancel()

	createdLivestream, createErrRes := s.livestreamGateway.Create(req2Ctx, *streamDTO)
	if createErrRes != nil {
		return "", "", fmt.Errorf("failed to create livestream: %s", createErrRes.Message)
	}

	livestreamId := createdLivestream.Id.String()

	// setup the vod creation
	s.vodHandler.OnStreamStart(livestreamId)
	return livestreamId, userInfo.Id.String(), nil
}

func (s *RTMPServer) onDisconnect(streamId string, duration int64) {
	s.vodHandler.OnStreamEnd(streamId, s.config.Transcode.PublicHLSPath, s.config.Transcode.FFMpegSetting.MasterFileName)

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
		logger.Errorf("failed to end livestream: %s", createErrRes.Message)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	cleanUpCtx, cancelCleanUp := context.WithTimeout(context.Background(), 10*time.Second)

	go func() {
		defer wg.Done()
		defer cancelCleanUp()
		<-cleanUpCtx.Done()
		removeLiveGeneratedFiles(streamId, s.config.Transcode.PrivateHLSPath, s.config.Transcode.PublicHLSPath)
	}()
	wg.Wait()
}

// remove live-generated private files and public files after saving into vods
func removeLiveGeneratedFiles(streamingKey, privatePath, publicPath string) error {
	// remove all folders of public and remove all private content
	paths := []string{
		filepath.Join(privatePath, streamingKey),
		filepath.Join(publicPath, streamingKey),
	}

	var errList []error

	for _, path := range paths {
		logger.Infof("path is removed", path)
		err := os.RemoveAll(path)
		if err != nil {
			errList = append(errList, fmt.Errorf("failed to remove %s: %w", path, err))
		}
	}

	return nil
}

func (s *RTMPServer) Shutdown(ctx context.Context) error {
	listener := s.listener
	s.listener = nil // prevent further use

	if listener == nil {
		logger.Warnf("RTMP server listener already closed or not started.")
		return nil
	}

	logger.Infow("shutting down RTMP server listener...")

	// close the listener, this will cause the accept loop in Start() to unblock
	err := listener.Close()

	if err != nil && !errors.Is(err, net.ErrClosed) {
		logger.Errorf("error closing RTMP listener: %v", err)
		return err
	}

	logger.Infow("RTMP server listener closed.")

	// Optionally: Add waiting logic here if you need to ensure active connections are finished.
	// This usually involves tracking active connections and waiting for them to complete.
	// The provided code doesn't track connections actively in the RTMPServer struct itself,
	// relying on the rtmp library and individual handler goroutines.

	return nil
}
