package rtmp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sen1or/lets-live/pkg/discovery"
	"sen1or/lets-live/pkg/logger"
	"sen1or/lets-live/transcode/config"
	"sen1or/lets-live/transcode/transcoder"
	"strconv"
	"strings"

	"github.com/nareix/joy5/format/flv"
	"github.com/nareix/joy5/format/flv/flvio"
	"github.com/nareix/joy5/format/rtmp"
)

type RTMPServerConfig struct {
	Port     int
	Registry discovery.Registry
	Config   config.Config
}

type RTMPServer struct {
	Port     int
	Registry discovery.Registry
	config   config.Config
}

func NewRTMPServer(config RTMPServerConfig) *RTMPServer {
	return &RTMPServer{
		Port:     config.Port,
		Registry: config.Registry,
		config:   config.Config,
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
			logger.Debugw("RTMP log tag", t.DebugFields())
		}
	}

	streamingKeyComponents := strings.Split(c.URL.Path, "/")
	streamingKey := streamingKeyComponents[len(streamingKeyComponents)-1]

	//streamInfo, err := s.onConnect(streamingKey)
	//if err != nil {
	//	logger.Errorw("request failed", "err", err)
	//	nc.Close()
	//	return
	//}

	//logger.Infof("GET THE STREAM INFO WITH USERID - %s", streamInfo.UserID)

	//pipeOut, pipeIn := io.Pipe()

	//go func() {
	//	transcoder := transcoder.NewTranscoder(pipeOut, s.config)
	//	transcoder.Start(streamInfo.UserID)
	//}()

	logger.Infof("GET THE STREAM INFO WITH USERID - %s", "1cf65df3-1f9f-4e81-94e5-951a99bcb4ce")

	pipeOut, pipeIn := io.Pipe()

	go func() {
		transcoder := transcoder.NewTranscoder(pipeOut, s.config)
		transcoder.Start("1cf65df3-1f9f-4e81-94e5-951a99bcb4ce")
	}()

	w := flv.NewMuxer(pipeIn)

	for {
		pkt, err := c.ReadPacket()
		if err == io.EOF {
			s.onDisconnect(streamingKey)
			return
		}

		if err := w.WritePacket(pkt); err != nil {
			logger.Errorf("failed to write rtmp package: %s", err)
			s.onDisconnect(streamingKey)
			return
		}
	}
}

type response struct {
	UserID string
}

func (s *RTMPServer) onConnect(streamingKey string) (info *response, err error) {
	logger.Infow("a stream is connected", "stream api key", streamingKey)

	userServerAddress, err := s.Registry.ServiceAddress(context.Background(), "user")
	if err != nil {
		logger.Errorf("failed to get service connection: %s", err)
		return
	}

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/v1/streams/%s/online", userServerAddress, streamingKey), nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	} else if res.StatusCode/100 != 2 {
		buf := new(strings.Builder)
		errMsg, _ := io.Copy(buf, res.Body)
		return nil, errors.New("request failed with status code" + string(res.StatusCode) + ", msg: " + string(errMsg))
	}

	defer res.Body.Close()
	var streamInfo response
	if err := json.NewDecoder(res.Body).Decode(&streamInfo); err != nil {
		return nil, errors.New("failed to decode the response")
	}

	return &streamInfo, nil
}

func (s *RTMPServer) onDisconnect(streamingKey string) {
	logger.Infof(fmt.Sprintf("a stream disconnected with stream key %s", streamingKey))

	userServerAddress, err := s.Registry.ServiceAddress(context.Background(), "user")
	if err != nil {
		logger.Errorf("failed to get service connection: %s", err)
		return
	}

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/v1/streams/%s/offline", userServerAddress, streamingKey), nil)
	if err != nil {
		logger.Errorw("failed to make http request")
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf("http client failed: ", err)
	} else if res.StatusCode/100 != 2 {
		buf := new(strings.Builder)
		errMsg, _ := io.Copy(buf, res.Body)
		logger.Errorf("request failed: %s", string(errMsg))
	}

	defer res.Body.Close()
}
