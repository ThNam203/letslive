package ipfs

import (
	"os"
	"sen1or/lets-live/internal/config"
	"testing"
)

func TestCreateIPFSNode(t *testing.T) {
	cfg := config.GetConfig()
	ipfsStorage := NewIPFSStorage(cfg.PrivateHLSPath, "http://localhost:8080/")

	file, err := os.OpenFile("./test.txt", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		t.Error("failed to create file")
	}

	file.Write([]byte("please do work for me please, i dont want to have bad grades"))
	file.Close()

	ipfsStorage.SaveIntoHLSDirectory("./test.txt")

	if ipfsStorage == nil {
		t.Fatal("where is my ipfs storage")
	}

	ipfsStorage.SaveIntoHLSDirectory("/home/sen1or/Desktop/.life/secrets/vaj.txt")
}
