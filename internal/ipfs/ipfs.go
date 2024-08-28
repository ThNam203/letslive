package ipfs

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/kubo/config"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/ipfs/kubo/core/coreiface/options"
	"github.com/ipfs/kubo/core/node/libp2p"
	"github.com/ipfs/kubo/plugin/loader"
	"github.com/ipfs/kubo/repo/fsrepo"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type IIPFSStorage interface {
	// // Save the file and return its hash string
	//Save(filePath string) (string, error)

	// The function save the file and add into a hls directory
	// Return final hash after adding file hash into directory
	SaveIntoHLSDirectory(filePath string) (string, error)
}

// TODO: add way to check if hls directory exists
type IPFSStorage struct {
	ipfsApi iface.CoreAPI
	node    *core.IpfsNode
	ctx     context.Context

	hlsDirectoryHash string
}

func NewIPFSStorage(ctx context.Context) IIPFSStorage {
	ipfsStorage := &IPFSStorage{
		ctx: ctx,
	}
	ipfsStorage.setup()

	return ipfsStorage
}

func (s *IPFSStorage) setup() {
	ipfsApi, node, err := s.spawnEphemeral()
	if err != nil {
		log.Panic(err)
	}

	s.ipfsApi = *ipfsApi
	s.node = node
	fmt.Println("IPFS storage is running")

	hlsDirectoryHashString, err := s.AddDirectory("./hls")
	if err != nil {
		log.Panicf("failed to add hls directory: %s", err)
	}

	s.hlsDirectoryHash = hlsDirectoryHashString
}

// create and return the directory hash string
func (s *IPFSStorage) AddDirectory(directoryPath string) (string, error) {
	directoryNode, err := getUnixfsNode(directoryPath)
	defer (func() {
		if directoryNode != nil {
			directoryNode.Close()
		}
	})()

	if err != nil {
		return "", fmt.Errorf("failed to create directory: %s", err)
	}

	directoryHash, err := s.ipfsApi.Unixfs().Add(s.ctx, directoryNode)
	if err != nil {
		return "", fmt.Errorf("failed to add directory: %s", err)
	}

	return directoryHash.String(), nil
}

func (s *IPFSStorage) Save(filePath string) (path.Path, error) {
	file, err := getUnixfsNode(filePath)
	defer file.Close()

	if err != nil {
		return nil, fmt.Errorf("failed to get file: %s", err)
	}

	opts := []options.UnixfsAddOption{
		options.Unixfs.Pin(false),
	}

	fileCid, err := s.ipfsApi.Unixfs().Add(s.ctx, file, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to add file into ipfs: %s", err)
	}

	return fileCid, err
}

func (s *IPFSStorage) SaveIntoHLSDirectory(filePath string) (string, error) {
	fileCid, err := s.Save(filePath)
	if err != nil {
		return "", err
	}

	finalHash, err := s.addHashedFileToDirectory(fileCid, s.hlsDirectoryHash, filepath.Base(filePath))
	return finalHash, err
}

// Add the hashed file into "hls" directory hash which is already get added into IPFS storage
func (s *IPFSStorage) addHashedFileToDirectory(fileHash path.Path, directoryToAddTo string, filename string) (string, error) {
	directoryPath, err := path.NewPath(directoryToAddTo)
	if err != nil {
		return "", err
	}

	newDirectoryHash, err := s.ipfsApi.Object().AddLink(s.ctx, directoryPath, filename, fileHash)
	if err != nil {
		return "", err
	}

	return filepath.Join(newDirectoryHash.String(), filename), nil
}

func getUnixfsNode(path string) (files.Node, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	f, err := files.NewSerialFile(path, false, st)
	if err != nil {
		return nil, err
	}

	return f, nil
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

// Must load plugins before setting up everything
func setupPlugins(repoPath string) error {
	plugins, err := loader.NewPluginLoader(repoPath)
	if err != nil {
		return fmt.Errorf("error loading plugins: %s", err)
	}

	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	return nil
}

// Setting buffer to 7.5M
// Explain: https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes
func createTempRepo() (string, error) {
	repoPath, err := os.MkdirTemp("", "ipfs-shell")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir for ipfs: %s", err)
	}

	cfg, err := config.Init(log.Writer(), 2048)
	if err != nil {
		return "", fmt.Errorf("failed to init config file for repo: %s", err)
	}

	err = fsrepo.Init(repoPath, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create ephemeral node: %s", err)
	}
	return repoPath, nil
}

var loadPluginsOnce sync.Once

// Function "spawnEphemeral" Create a temporary just for one run
func (s *IPFSStorage) spawnEphemeral() (*iface.CoreAPI, *core.IpfsNode, error) {
	var onceErr error
	loadPluginsOnce.Do(func() {
		onceErr = setupPlugins("")
	})

	if onceErr != nil {
		return nil, nil, onceErr
	}

	// Create a Temporary Repo
	repoPath, err := createTempRepo()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp repo: %s", err)
	}

	api, node, err := createIPFSNode(s.ctx, repoPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create ipfs node: %s", err)
	}

	return api, node, err
}

func connectToPeers(ctx context.Context, ipfs iface.CoreAPI, peers []string) error {
	var wg sync.WaitGroup
	peerInfos := make(map[peer.ID]*peer.AddrInfo, len(peers))
	for _, addrStr := range peers {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			return err
		}
		pii, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return err
		}
		pi, ok := peerInfos[pii.ID]
		if !ok {
			pi = &peer.AddrInfo{ID: pii.ID}
			peerInfos[pi.ID] = pi
		}
		pi.Addrs = append(pi.Addrs, pii.Addrs...)
	}

	wg.Add(len(peerInfos))
	for _, peerInfo := range peerInfos {
		go func(peerInfo *peer.AddrInfo) {
			defer wg.Done()
			err := ipfs.Swarm().Connect(ctx, *peerInfo)
			if err != nil {
				log.Printf("failed to connect to %s: %s", peerInfo.ID, err)
			}
		}(peerInfo)
	}
	wg.Wait()
	return nil
}
