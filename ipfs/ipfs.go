package ipfs

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ipfs/kubo/config"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/ipfs/kubo/core/node/libp2p"
	"github.com/ipfs/kubo/repo/fsrepo"
)

type IPFSStorage struct {
	ipfsApi *iface.CoreAPI
	node    *core.IpfsNode

	ctx context.Context
}

func NewIPFSStorage(ctx context.Context) *IPFSStorage {
	ipfsStorage := &IPFSStorage{}
	ipfsStorage.setup(ctx)

	return ipfsStorage
}

func (s *IPFSStorage) setup(ctx context.Context) {
	ipfsApi, node, err := s.createIPFSInstance()
	if err != nil {
		log.Panic(err)
	}

	s.ipfsApi = ipfsApi
	s.node = node
}

func (s *IPFSStorage) Save() error {
	// TODO
	return nil
}

func createIPFSNode(ctx context.Context, repoPath string) (*iface.CoreAPI, *core.IpfsNode, error) {
	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		return nil, nil, err
	}

	nodeOptions := &core.BuildCfg{
		Online:  true,
		Routing: libp2p.DHTOption,
		Repo:    repo,
	}

	node, err := core.NewNode(ctx, nodeOptions)
	// TODO:
	// node.IsDaemon = true

	if err != nil {
		return nil, nil, err
	}

	coreApi, err := coreapi.NewCoreAPI(node)
	return &coreApi, node, nil
}

func createTempRepo(ctx context.Context) (string, error) {
	repoPath, err := os.MkdirTemp("", "ipfs")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir for ipfs: %s", err)
	}

	cfg, err := config.Init(log.Writer(), 2048)
	if err != nil {
		return "", err
	}

	err = fsrepo.Init(repoPath, cfg)
	return repoPath, err
}

func (s *IPFSStorage) createIPFSInstance() (*iface.CoreAPI, *core.IpfsNode, error) {
	tempPath, err := createTempRepo(s.ctx)
	if err != nil {
		return nil, nil, err
	}

	api, node, err := createIPFSNode(s.ctx, tempPath)
	return api, node, err
}
