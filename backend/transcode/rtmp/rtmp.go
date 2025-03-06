package rtmp

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sen1or/lets-live/pkg/discovery"
	"sen1or/lets-live/pkg/logger"
	"sen1or/lets-live/transcode/config"
	livestreamdto "sen1or/lets-live/transcode/gateway/livestream"
	livestreamgateway "sen1or/lets-live/transcode/gateway/livestream/http"
	userdto "sen1or/lets-live/transcode/gateway/user"
	usergateway "sen1or/lets-live/transcode/gateway/user/http"
	"sen1or/lets-live/transcode/transcoder"
	"sen1or/lets-live/transcode/watcher"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/nareix/joy5/format/flv"
	"github.com/nareix/joy5/format/flv/flvio"
	"github.com/nareix/joy5/format/rtmp"
)

type RTMPServerConfig struct {
	Port       int
	Registry   *discovery.Registry
	Config     config.Config
	VODHandler watcher.VODHandler
}

type RTMPServer struct {
	Port              int
	Registry          *discovery.Registry
	userGateway       *usergateway.UserGateway
	livestreamGateway *livestreamgateway.LivestreamGateway
	config            config.Config
	vodHandler        watcher.VODHandler
}

func NewRTMPServer(config RTMPServerConfig, userGateway *usergateway.UserGateway, livestreamgateway *livestreamgateway.LivestreamGateway) *RTMPServer {
	return &RTMPServer{
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
	var listener net.Listener

	serverAddr, err := net.ResolveTCPAddr("tcp", ":"+portStr)
	if err != nil {
		logger.Panicf("failed to resolve rtmp addr: %s", err)
	}

	listener, err = net.ListenTCP("tcp", serverAddr)
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
		conn, err := listener.Accept()

		if err != nil {
			logger.Errorf("rtmp failed to connect: %s", err)
			continue
		}

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

	streamId, userId, err := s.onConnect(streamingKey) // get stream id

	if err != nil {
		logger.Errorf("stream connection failed: %s", err)
		nc.Close()
		return
	}

	pipeOut, pipeIn := io.Pipe()

	go func() {
		transcoder := transcoder.NewTranscoder(pipeOut, s.config.Transcode)
		transcoder.Start(streamId)
	}()

	w := flv.NewMuxer(pipeIn)

	for {
		pkt, err := c.ReadPacket()
		if err == io.EOF {
			s.onDisconnect(userId, streamId)
			return
		}

		if err := w.WritePacket(pkt); err != nil {
			logger.Errorf("failed to write rtmp package: %s", err)
			s.onDisconnect(userId, streamId)
			return
		}
	}
}

// check if stream api key exists
// then update the user status to online
// return the stream id to be used as publishName
func (s *RTMPServer) onConnect(streamingKey string) (streamId string, userId string, err error) {
	userInfo, errRes := s.userGateway.GetUserInformation(context.Background(), streamingKey)
	if errRes != nil {
		return "", "", fmt.Errorf("failed to get user information: %s", errRes.Message)
	}

	updateUserDTO := &userdto.UpdateUserIsOnlineDTO{
		Id:       userInfo.Id,
		IsOnline: true,
	}

	errRes = s.userGateway.UpdateUserLiveStatus(context.Background(), *updateUserDTO)
	if errRes != nil {
		return "", "", fmt.Errorf("failed to update user live status: %s", errRes.Message)
	}

	streamDTO := &livestreamdto.CreateLivestreamRequestDTO{
		Title:        userInfo.LivestreamInformationResponseDTO.Title,
		UserId:       userInfo.Id,
		Description:  userInfo.LivestreamInformationResponseDTO.Description,
		ThumbnailURL: userInfo.LivestreamInformationResponseDTO.ThumbnailURL,
		Status:       "live",
	}

	createdLivestream, createErrRes := s.livestreamGateway.Create(context.Background(), *streamDTO)
	if createErrRes != nil {
		return "", "", fmt.Errorf("failed to create livestream: %s", createErrRes.Message)
	}

	livestreamId := createdLivestream.Id.String()

	// setup the vod creation
	s.vodHandler.OnStreamStart(livestreamId)
	return livestreamId, userInfo.Id.String(), nil
}

// change the status of user to be not online
func (s *RTMPServer) onDisconnect(userId string, streamId string) {
	userIdUUID, _ := uuid.FromString(userId)
	updateUserDTO := &userdto.UpdateUserIsOnlineDTO{
		Id:       userIdUUID,
		IsOnline: false,
	}

	// create the VOD playlists and remove the entry
	s.vodHandler.OnStreamEnd(streamId, s.config.Transcode.PublicHLSPath, s.config.Transcode.FFMpegSetting.MasterFileName)

	errRes := s.userGateway.UpdateUserLiveStatus(context.Background(), *updateUserDTO)
	if errRes != nil {
		logger.Errorf("failed to get service connection: %s", errRes.Message)
	}

	endedStatus := "ended"
	playbackURL := fmt.Sprintf("http://%s:%d/vods/%s/index.m3u8", s.config.Service.Hostname, s.config.Webserver.Port, streamId)
	endedAt := time.Now()

	updateDTO := &livestreamdto.UpdateLivestreamRequestDTO{
		Id:           uuid.FromStringOrNil(streamId),
		Title:        nil,
		Description:  nil,
		ThumbnailURL: nil,
		Status:       &endedStatus,
		PlaybackURL:  &playbackURL,
		ViewCount:    nil,
		EndedAt:      &endedAt,
	}

	createErrRes := s.livestreamGateway.Update(context.Background(), *updateDTO)
	if createErrRes != nil {
		logger.Errorf("failed to update livestream: %s", createErrRes.Message)
	}

	// should be put on the last line
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func() {
		<-ctx.Done()
		removeLiveGeneratedFiles(streamId, s.config.Transcode.PrivateHLSPath, s.config.Transcode.PublicHLSPath)
	}()
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
