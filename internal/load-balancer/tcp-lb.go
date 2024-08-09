package loadbalancer

import (
	"context"
	"io"
	"log"
	"net"
	"net/url"
	"sen1or/lets-live/internal/config"
	"sync"
	"time"
)

type TCPLoadBalancer struct {
	backendPool BackendPool
	config      config.LBSetting
}

func NewTCPLoadBalancer(config config.LBSetting) *TCPLoadBalancer {
	backends := make([]Backend, 0)

	for _, address := range config.To {
		url, err := url.Parse(address)
		if err != nil {
			log.Printf("backend address '%s' failed to parse", address)
			continue
		}
		be := NewBackend(url)
		backends = append(backends, *be)
	}

	if len(backends) == 0 {
		log.Panic("no backend found")
	}

	return &TCPLoadBalancer{
		backendPool: *NewBackendPool(backends),
		config:      config,
	}
}

func (lb *TCPLoadBalancer) ListenAndServe() error {
	listener, err := net.Listen("tcp", lb.config.From)

	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		log.Println("accepted a connection")
		if err != nil {
			log.Println("error accepting connection: ", err)
		}

		be, err := lb.backendPool.GetNextBackend()
		if err != nil {
			log.Printf("failed to get backend")
			conn.Close()
			continue
		}

		go connect(be, conn)
	}
}

func connect(be *Backend, incomingConn net.Conn) {
	defer incomingConn.Close()

	outConn, err := net.DialTimeout("tcp", be.url.String(), 5*time.Second)
	if err != nil {
		log.Println("error while dialing to backend: ", err)
		return
	}

	defer outConn.Close()

	ctx := context.Background()

	err = transferData(ctx, incomingConn, outConn)
	if err != nil {
		log.Println("error while transfering data: ", err)
	}
}

func transferData(ctx context.Context, incomingConn net.Conn, outConn net.Conn) error {
	ctx, cancel := context.WithCancel(ctx)
	var err error = nil
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, localErr := io.Copy(incomingConn, outConn)
		setErrorIfHave(localErr, err)
		cancel()
	}()

	go func() {
		defer wg.Done()
		_, localErr := io.Copy(outConn, incomingConn)
		setErrorIfHave(localErr, err)
		cancel()
	}()

	go func() {
		<-ctx.Done()

		incomingConn.Close()
		outConn.Close()
	}()

	wg.Wait()
	return err
}

func setErrorIfHave(localError error, resultError error) {
	if localError != nil && resultError == nil {
		resultError = localError
	}
}
