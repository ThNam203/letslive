package rtmp

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sen1or/lets-live/logger"
	"strconv"
	"strings"

	"github.com/nareix/joy5/format/flv"
	"github.com/nareix/joy5/format/flv/flvio"
	"github.com/nareix/joy5/format/rtmp"
)

type RTMPServer struct {
	Port          int
	MainServerURL string
}

func NewRTMPServer(port int, mainServerURL string) *RTMPServer {
	return &RTMPServer{
		Port:          port,
		MainServerURL: mainServerURL,
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

	if err := s.onConnect(streamingKey); err != nil {
		nc.Close()
		return
	}

	pipeOut, pipeIn := io.Pipe()

	go func() {
		transcoder := NewTranscoder(pipeOut)
		transcoder.Start(streamingKey)
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

func (s *RTMPServer) onConnect(streamingKey string) error {
	logger.Infow("a stream is connected", "stream api key", streamingKey)
	req, err := http.NewRequest(http.MethodPatch, s.MainServerURL+fmt.Sprintf("/v1/streams/%s/online", streamingKey), nil)
	if err != nil {
		logger.Errorw("failed to make http request")
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorw("request failed", "err", err)
		return err
	} else if res.StatusCode/100 != 2 {
		buf := new(strings.Builder)
		errMsg, _ := io.Copy(buf, res.Body)
		logger.Errorw("request failed", "status code", res.StatusCode, "msg", errMsg)
		return errors.New("request failed with status code" + string(res.StatusCode))
	}

	defer res.Body.Close()

	return nil
}

func (s *RTMPServer) onDisconnect(streamingKey string) {
	logger.Infof(fmt.Sprintf("a stream disconnected with stream key %s", streamingKey))

	req, err := http.NewRequest(http.MethodPatch, s.MainServerURL+fmt.Sprintf("/v1/streams/%s/offline", streamingKey), nil)
	if err != nil {
		logger.Errorw("failed to make http request")
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorw("request failed", "err", err)
	} else if res.StatusCode/100 != 2 {
		buf := new(strings.Builder)
		errMsg, _ := io.Copy(buf, res.Body)
		logger.Errorw("request failed", "status code", res.StatusCode, "msg", errMsg)
	}

	defer res.Body.Close()
}
