package ipfs

import (
	"context"
	"testing"
)

func TestCreateIPFSNode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ipfsStorage := NewIPFSStorage(ctx)

	if ipfsStorage == nil {
		t.Fatal("where is my ipfs storage")
	}

	ipfsStorage.SaveIntoHLSDirectory("/home/sen1or/Desktop/.life/secrets/vaj.txt")
}
