package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

var (
	ipfsNode *Peer
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var bootstrapNodeAddr string

	isBootstrapNode := flag.Bool("b", false, "Use if the node is a bootstrap node")
	flag.StringVar(&bootstrapNodeAddr, "a", "", "The boostrap node address")

	flag.Parse()

	if *isBootstrapNode {
		if err := RunBootstrapNode(ctx); err != nil {
			panic(err)
		}

		//fmt.Println("cai con cac")
	} else {
		if len(bootstrapNodeAddr) == 0 {
			log.Panic("missing bootstrap node address")
		}

		if err := RunNormalNode(ctx, bootstrapNodeAddr); err != nil {
			log.Panic(err)
		}
	}

	select {}
}

func RunBootstrapNode(ctx context.Context) error {
	ds := NewInMemoryDatastore()
	host, dht, err := NewLibp2pHost(ctx, ds, true)
	if err != nil {
		return err
	}

	ipfsNode, err = NewIPFSNode(ctx, ds, host, dht)
	if err != nil {
		return err
	}

	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", ipfsNode.host.ID().String()))
	addr := ipfsNode.host.Addrs()[0]
	log.Println("running as bootstrap node, ignore -a flag if there is any")
	log.Printf("** bootstrap node address: %s\n", addr.Encapsulate(hostAddr))

	// serve files
	http.HandleFunc("/ipfs/{fileCid}", getFileHandler)
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()
	log.Println("start serving files on 8080")

	return nil
}

func getFileHandler(w http.ResponseWriter, req *http.Request) {
	// Enable CORS for everyone
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	// Handle OPTIONS request for CORS preflight
	if req.Method == http.MethodOptions {
		return
	}

	fileName := req.URL.Query().Get("fileName")
	if len(fileName) == 0 {
		http.Error(w, "no file name provided", http.StatusBadRequest)
		return
	}

	fileCidString := req.PathValue("fileCid")
	fileCid, err := cid.Decode(fileCidString)

	fmt.Printf("getting file for cid: %s\n", fileCidString)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, err := ipfsNode.GetFile(req.Context(), fileCid)
	if err != nil {
		http.Error(w, "failed to retrieve file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// set the appropriate headers for file serving
	w.Header().Set("Content-Type", "video/MP2T")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")

	// copy the file content to the response writer
	if _, err := io.Copy(w, file); err != nil {
		http.Error(w, "failed to send file: "+err.Error(), http.StatusInternalServerError)
	}
}

func RunNormalNode(ctx context.Context, bootstrapNodeAddr string) error {
	// parse the bootstrap node address
	targetAddr, err := multiaddr.NewMultiaddr(bootstrapNodeAddr)
	if err != nil {
		return err
	}

	targetInfo, err := peer.AddrInfoFromP2pAddr(targetAddr)
	if err != nil {
		return err
	}

	// create node
	ds := NewInMemoryDatastore()
	host, dht, err := NewLibp2pHost(ctx, ds, false)
	if err != nil {
		return err
	}

	ipfsNode, err = NewIPFSNode(ctx, ds, host, dht)
	if err != nil {
		return err
	}

	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", ipfsNode.host.ID().String()))
	addr := ipfsNode.host.Addrs()[0]
	log.Printf("running as normal with addr: %s, trying to connect with bootstrap node", addr.Encapsulate(hostAddr))

	// connect to bootstrap node
	if err := host.Connect(ctx, *targetInfo); err != nil {
		return err
	}
	log.Printf("connected to bootstrap node (%s)\n", bootstrapNodeAddr)

	return nil
}

// TESTING FUNCTIONS - TODO: write unit tests

// func GetFileFromCID(ctx context.Context, fileCid string) {
// 	c, err := cid.Decode(fileCid)
// 	if err != nil {
// 		log.Printf("invalid CID: %s\n", err)
// 		return
// 	}
//
// 	rsc, err := ipfsNode.GetFile(ctx, c)
// 	if err != nil {
// 		log.Printf("failed to get file from node: %s\n", err)
// 		return
// 	}
//
// 	defer rsc.Close()
// 	log.Println("file successfully retrieved from the node!")
// }
//
// func getFileFromNode(ctx context.Context) {
// 	fileCid := "QmPtU9NDfdxFB2oRiE4Lv37i4zWgVPme7qjTqfhZZ18Z89"
// 	c, err := cid.Decode(fileCid)
// 	if err != nil {
// 		log.Printf("invalid CID: %s\n", err)
// 		return
// 	}
// 	rsc, err := ipfsNode.GetFile(ctx, c)
//
// 	if err != nil {
// 		log.Printf("failed to get file from node: %s\n", err)
// 		return
// 	}
//
// 	defer rsc.Close()
// 	log.Println("file successfully retrieved from the node!")
// }
//
// func addFileToNode(ctx context.Context) (fileCid string, err error) {
// 	file, err := os.Open("./example_file_to_be_added.txt")
// 	if err != nil {
// 		log.Printf("failed to open file: %s\n", err)
// 		return "", err
// 	}
//
// 	ipldNode, err := ipfsNode.AddFile(ctx, file)
// 	if err != nil {
// 		log.Printf("failed to save file into node: %s\n", err)
// 		return "", err
// 	}
//
// 	log.Printf("saved a file with cid: %s", ipldNode.Cid().String())
// 	return ipldNode.String(), nil
// }
