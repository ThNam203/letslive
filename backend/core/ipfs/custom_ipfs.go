package ipfs

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type CustomStorage struct {
	ipfsNode          *Peer
	bootstrapNodeAddr string
	ctx               context.Context
}

func NewCustomStorage(ctx context.Context, bootstrapNodeAddr string) *CustomStorage {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(bootstrapNodeAddr) == 0 {
		log.Panic("missing bootstrap node address")
	}

	storage := &CustomStorage{
		bootstrapNodeAddr: bootstrapNodeAddr,
		ctx:               ctx,
	}

	if err := storage.SetupNode(); err != nil {
		log.Panic(err)
	}

	return storage
}

func (s *CustomStorage) SetupNode() error {
	// parse the bootstrap node address
	targetAddr, err := multiaddr.NewMultiaddr(s.bootstrapNodeAddr)
	if err != nil {
		return err
	}

	targetInfo, err := peer.AddrInfoFromP2pAddr(targetAddr)
	if err != nil {
		return err
	}

	// create node
	ds := NewInMemoryDatastore()
	host, dht, err := NewLibp2pHost(s.ctx, ds, false)
	if err != nil {
		return err
	}

	node, err := NewIPFSNode(s.ctx, ds, host, dht)
	if err != nil {
		return err
	}

	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", node.host.ID().String()))
	addr := node.host.Addrs()[0]
	log.Printf("running as normal with addr: %s, trying to connect with bootstrap node", addr.Encapsulate(hostAddr))

	// connect to bootstrap node
	if err := host.Connect(s.ctx, *targetInfo); err != nil {
		return err
	}
	log.Printf("connected to bootstrap node (%s)\n", s.bootstrapNodeAddr)

	s.ipfsNode = node

	return nil
}
