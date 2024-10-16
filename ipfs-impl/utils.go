package main

import (
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
)

func SavePrivateKey(privKey crypto.PrivKey, filename string) error {
	data, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("error marshaling private key: %v", err)
	}

	// Write the byte array to a file
	return os.WriteFile(filename, data, 0644)
}

func LoadPrivateKey(filename string) (crypto.PrivKey, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %v", err)
	}

	key, err := crypto.UnmarshalPrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling private key: %v", err)
	}

	return key, nil
}
