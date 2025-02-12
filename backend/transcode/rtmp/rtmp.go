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
	usergateway "sen1or/lets-live/transcode/gateway/user/http"
	"sen1or/lets-live/transcode/transcoder"
	"sen1or/lets-live/transcode/watcher"
	"sen1or/lets-live/user/dto"
	"strconv"
	"strings"

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
	Port        int
	Registry    *discovery.Registry
	userGateway *usergateway.UserGateway
	config      config.Config
	vodHandler  watcher.VODHandler
}

func NewRTMPServer(config RTMPServerConfig, userGateway *usergateway.UserGateway) *RTMPServer {
	return &RTMPServer{
		Port:        config.Port,
		Registry:    config.Registry,
		config:      config.Config,
		userGateway: userGateway,
		vodHandler:  config.VODHandler,
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

	userId, err := s.onConnect(streamingKey) // userId is used as publishName

	if err != nil {
		logger.Errorf("stream connection failed: %s", err)
		nc.Close()
		return
	}

	pipeOut, pipeIn := io.Pipe()

	go func() {
		transcoder := transcoder.NewTranscoder(pipeOut, s.config)
		transcoder.Start(userId)
	}()

	w := flv.NewMuxer(pipeIn)

	for {
		pkt, err := c.ReadPacket()
		if err == io.EOF {
			s.onDisconnect(userId)
			return
		}

		if err := w.WritePacket(pkt); err != nil {
			logger.Errorf("failed to write rtmp package: %s", err)
			s.onDisconnect(userId)
			return
		}
	}
}

// check if stream api key exists
// then update the user status to online
// return the user id to be used as publishName
func (s *RTMPServer) onConnect(streamingKey string) (string, error) {
	userInfo, errRes := s.userGateway.GetUserInformation(context.Background(), streamingKey)
	if errRes != nil {
		return "", fmt.Errorf("failed to get user information: %s", errRes.Message)
	}

	updateUserDTO := &dto.UpdateUserRequestDTO{
		Id:       userInfo.Id,
		IsOnline: func(b bool) *bool { return &b }(true), // wtf
	}

	errRes = s.userGateway.UpdateUserLiveStatus(context.Background(), *updateUserDTO)
	if errRes != nil {
		return "", fmt.Errorf("failed to get service connection: %s", errRes.Message)
	}

	// setup the vod creation
	s.vodHandler.OnStreamStart(userInfo.Id.String())

	// make sure there is not any files from the privous streaming session
	if err := removeLiveGeneratedFiles(userInfo.Id.String(), s.config.Transcode.PrivateHLSPath, s.config.Transcode.PublicHLSPath); err != nil {
		return "", fmt.Errorf("failed to remove live generated files: %s", err)
	}

	return userInfo.Id.String(), nil
}

// change the status of user to be not online
func (s *RTMPServer) onDisconnect(userId string) {
	userIdUUID, _ := uuid.FromString(userId)
	updateUserDTO := &dto.UpdateUserRequestDTO{
		Id:       userIdUUID,
		IsOnline: func(b bool) *bool { return &b }(false), // wtf
	}

	// create the VOD playlists and remove the entry
	s.vodHandler.OnStreamEnd(userId, s.config.Transcode.PublicHLSPath, s.config.Transcode.FFMpegSetting.MasterFileName)

	errRes := s.userGateway.UpdateUserLiveStatus(context.Background(), *updateUserDTO)
	if errRes != nil {
		logger.Errorf("failed to get service connection: %s", errRes.Message)
	}

	// should be put on the last line
	removeLiveGeneratedFiles(userId, s.config.Transcode.PrivateHLSPath, s.config.Transcode.PublicHLSPath)
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("error reading file: %s", err)
	}

	err = os.WriteFile(dst, input, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error copying file: %s", err)
	}

	return nil
}

// remove live-generated private files and public files after saving into vods
// TODO: make it better
func removeLiveGeneratedFiles(streamingKey, privatePath, publicPath string) error {
	// remove all folders of public and remove all private content
	paths := []string{
		filepath.Join(privatePath, streamingKey),
		filepath.Join(publicPath, streamingKey, "0"),
		filepath.Join(publicPath, streamingKey, "1"),
		filepath.Join(publicPath, streamingKey, "2"),
	}

	var errList []error

	for _, path := range paths {
		logger.Infof("path is removed", path)
		err := os.RemoveAll(path)
		if err != nil {
			errList = append(errList, fmt.Errorf("failed to remove %s: %w", path, err))
		}
	}

	// remove all m3u8 files
	files, err := filepath.Glob(filepath.Join(publicPath, streamingKey) + "/*.m3u8")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			errList = append(errList, err)
		}
	}

	if len(errList) > 0 {
		return fmt.Errorf("encountered errors: %v", errList)
	}

	return nil
}
