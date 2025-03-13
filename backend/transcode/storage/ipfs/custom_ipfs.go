package ipfs

import (
	"context"
	"fmt"
	"os"
	"sen1or/letslive/transcode/pkg/logger"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type IPFSStorage struct {
	ipfsNode          *Peer
	bootstrapNodeAddr *string
	gateway           string
	ctx               context.Context
}

// If we don't want to connect to bootstrap node, enter a nil value for bootstrapNodeAddr
func NewIPFSStorage(ctx context.Context, gateway string, bootstrapNodeAddr *string) *IPFSStorage {
	if bootstrapNodeAddr != nil && len(*bootstrapNodeAddr) == 0 {
		logger.Panicw("missing bootstrap node address")
	}

	storage := &IPFSStorage{
		bootstrapNodeAddr: bootstrapNodeAddr,
		ctx:               ctx,
		gateway:           gateway,
	}

	if err := storage.SetupNode(); err != nil {
		logger.Panicf("error setting up node: %s", err)
	}

	return storage
}

func (s *IPFSStorage) AddSegment(filePath string, _ string, _ int) (string, error) {
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

func (s *IPFSStorage) AddThumbnail(filePath string, _ string, _ string) (string, error) {
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

func (s *IPFSStorage) SetupNode() error {
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
	logger.Infof("running as normal with addr: %s", addr.Encapsulate(hostAddr))

	if s.bootstrapNodeAddr != nil {
		logger.Infof("trying to connect with bootstrap node")

		// connect to bootstrap node
		// parse the bootstrap node address
		targetAddr, err := multiaddr.NewMultiaddr(*s.bootstrapNodeAddr)
		if err != nil {
			return err
		}

		targetInfo, err := peer.AddrInfoFromP2pAddr(targetAddr)
		if err != nil {
			return err
		}

		if err := host.Connect(s.ctx, *targetInfo); err != nil {
			logger.Errorf("failed to connect to bootstrap node (%s)", err)
		} else {
			logger.Infof("connected to bootstrap node (%s)", s.bootstrapNodeAddr)
		}
	}

	s.ipfsNode = node

	return nil
}
