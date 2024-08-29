package rtmpserver

import (
	"fmt"
	"io"
	"log"
	"net"
	"sen1or/lets-live/internal/config"

	"github.com/sirupsen/logrus"
	"github.com/yutopp/go-rtmp"
)

type RTMPServer struct {
	config config.Config
}

func (sv *RTMPServer) Listen() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1936")
	if err != nil {
		log.Fatal("Can't resolve address for RTMP server!")
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal("RTMP server can't listen on address!")
	}

	srv := rtmp.NewServer(&rtmp.ServerConfig{
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
			l := logrus.StandardLogger()
			h := &Handler{
				config: sv.config,
			}

			return conn, &rtmp.ConnConfig{
				Handler: h,

				ControlState: rtmp.StreamControlStateConfig{
					DefaultBandwidthWindowSize: 6 * 1024 * 1024 / 8,
				},

				Logger: l,
			}
		},
	})

	if err := srv.Serve(listener); err != nil {
		log.Fatal("RTMP server can't serve")
	}
}

func StartRTMPService() {
	var config = config.GetConfig()

	server := &RTMPServer{
		config: config,
	}

	go server.Listen()
	fmt.Println("rtmp server listening")

}
