package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

var (
	ipfsNode *Peer
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ds := NewInMemoryDatastore()
	host, dht, err := NewLibp2pHost(ctx, ds)
	if err != nil {
		panic(err)
	}

	ipfsNode, err = NewIPFSNode(ctx, ds, host, dht)
	if err != nil {
		panic(err)
	}

	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", ipfsNode.host.ID().String()))
	addr := ipfsNode.host.Addrs()[0]
	log.Printf("IPFS node run on: %s", addr.Encapsulate(hostAddr))

	select {}
}

// local check lol
func runExample(ctx context.Context, host host.Host) {
	//fileCid, _ := addFileToNode(ctx)
	//getFileFromNode(ctx)
	//GetFileFromCID(ctx, fileCid)

	targetAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWC5uA1QUEXmKnceznhJetH8G4wKrmhgqzXGgTGhZUDvY5")
	if err != nil {
		panic(err)
	}

	targetInfo, err := peer.AddrInfoFromP2pAddr(targetAddr)
	if err != nil {
		panic(err)
	}

	err = host.Connect(ctx, *targetInfo)
	if err != nil {
		panic(err)
	}

	log.Println("Connected to", targetInfo.ID)
	GetFileFromCID(ctx, "QmdbNULb6QosrqAZyKgGWMR6rw1szcKG3vZmuNwi7QMFvq")

}

func GetFileFromCID(ctx context.Context, fileCid string) {
	c, err := cid.Decode(fileCid)
	if err != nil {
		log.Printf("invalid CID: %s\n", err)
		return
	}

	rsc, err := ipfsNode.GetFile(ctx, c)
	if err != nil {
		log.Printf("failed to get file from node: %s\n", err)
		return
	}

	defer rsc.Close()
	log.Println("file successfully retrieved from the node!")
}

func getFileFromNode(ctx context.Context) {
	fileCid := "QmPtU9NDfdxFB2oRiE4Lv37i4zWgVPme7qjTqfhZZ18Z89"
	c, err := cid.Decode(fileCid)
	if err != nil {
		log.Printf("invalid CID: %s\n", err)
		return
	}
	rsc, err := ipfsNode.GetFile(ctx, c)

	if err != nil {
		log.Printf("failed to get file from node: %s\n", err)
		return
	}

	defer rsc.Close()
	log.Println("file successfully retrieved from the node!")
}

func addFileToNode(ctx context.Context) (fileCid string, err error) {
	file, err := os.Open("./example_file_to_be_added.txt")
	if err != nil {
		log.Printf("failed to open file: %s\n", err)
		return "", err
	}

	ipldNode, err := ipfsNode.AddFile(ctx, file)
	if err != nil {
		log.Printf("failed to save file into node: %s\n", err)
		return "", err
	}

	log.Printf("saved a file with cid: %s", ipldNode.Cid().String())
	return ipldNode.String(), nil
}
