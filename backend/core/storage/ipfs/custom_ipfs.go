package ipfs

import (
	"context"
	"fmt"
	"log"
	"os"
	"sen1or/lets-live/core/storage"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type CustomStorage struct {
	ipfsNode          *Peer
	bootstrapNodeAddr string
	gateway           string
	ctx               context.Context
}

func NewCustomStorage(ctx context.Context, gateway string, bootstrapNodeAddr string) storage.Storage {
	if len(bootstrapNodeAddr) == 0 {
		log.Panic("missing bootstrap node address")
	}

	storage := &CustomStorage{
		bootstrapNodeAddr: bootstrapNodeAddr,
		ctx:               ctx,
		gateway:           gateway,
	}

	if err := storage.SetupNode(); err != nil {
		log.Panic(err)
	}

	return storage
}

func (s *CustomStorage) AddFile(filePath string) (string, error) {
	file, err := os.Open(filePath)

	if err != nil {
		return "", fmt.Errorf("failed to get file: %s", err)
	}

	fileCid, err := s.ipfsNode.AddFile(s.ctx, file)
	if err != nil {
		return "", fmt.Errorf("failed to add file into ipfs: %s", err)
	}

	return fmt.Sprintf("%s/ipfs/%s", s.gateway, fileCid.String()), nil
}

func (s *CustomStorage) SetupNode() error {
	// create node
	ds := NewInMemoryDatastore()
	host, dht, err := NewLibp2pHost(s.ctx, ds)
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
	// parse the bootstrap node address
	targetAddr, err := multiaddr.NewMultiaddr(s.bootstrapNodeAddr)
	if err != nil {
		return err
	}

	targetInfo, err := peer.AddrInfoFromP2pAddr(targetAddr)
	if err != nil {
		return err
	}

	if err := host.Connect(s.ctx, *targetInfo); err != nil {
		log.Printf("failed to connect to bootstrap node (%s)\n", err)
	} else {
		log.Printf("connected to bootstrap node (%s)\n", s.bootstrapNodeAddr)
	}

	s.ipfsNode = node

	return nil
}
