package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	ipns "github.com/ipfs/boxo/ipns"
	datastore "github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dualdht "github.com/libp2p/go-libp2p-kad-dht/dual"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
)

// NewInMemoryDatastore provides a sync datastore that lives in-memory only and is not persisted.
func NewInMemoryDatastore() datastore.Batching {
	return dssync.MutexWrap(datastore.NewMapDatastore())
}

var connMgr, _ = connmgr.NewConnManager(100, 400, connmgr.WithGracePeriod(time.Minute))

func NewLibp2pHost(
	ctx context.Context,
	ds datastore.Batching,
	isBootstrapHost bool) (host.Host, *dualdht.DHT, error) {
	var ddht *dualdht.DHT
	var err error

	priv, err := generatePrivKey(isBootstrapHost)
	if err != nil {
		panic(err)
	}

	addr1, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4001")
	addr2, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/udp/4001/quic-v1")
	addrs := []multiaddr.Multiaddr{addr1, addr2}

	// we are creating a LAN network so no need NAT or any security methods
	opts := []libp2p.Option{
		libp2p.Identity(priv),
		libp2p.ListenAddrs(addrs...),
		libp2p.ConnectionManager(connMgr),
		//libp2p.Security(libp2ptls.ID, libp2ptls.New),
		//libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultTransports,
		//libp2p.NATPortMap(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			ddht, err = newDHT(ctx, h, ds)
			return ddht, err
		}),
		//libp2p.EnableNATService(),
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, nil, err
	}

	return h, ddht, nil
}

// if the host is a normal node, just create a random priv key
//
// if the host is for a bootstrap node
// we use the same private key to persist the host identity which allows other nodes to connect
//
// if there is no private key ("bootstrap_priv.key" file)
// then generate randomly one and store and load it
func generatePrivKey(isBootstrapHost bool) (crypto.PrivKey, error) {
	var finalPriv crypto.PrivKey
	var r io.Reader

	if !isBootstrapHost {
		var err error
		r = rand.New(rand.NewSource(time.Now().Unix()))
		finalPriv, _, err = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
		if err != nil {
			return nil, err
		}
	} else {
		_, err := os.Stat("./bootstrap_priv.key")

		// generate a random priv key and store it
		if err != nil && errors.Is(err, os.ErrNotExist) {
			r = rand.New(rand.NewSource(time.Now().Unix()))
			priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
			if err != nil {
				return nil, err
			}

			if err := savePrivateKey(priv); err != nil {
				return nil, err
			}
		}

		if finalPriv, err = loadPrivateKey(); err != nil {
			return nil, err
		}
	}

	return finalPriv, nil
}

// see more: DualDHT vs KademliaDHT
func newDHT(ctx context.Context, h host.Host, ds datastore.Batching) (*dualdht.DHT, error) {
	dhtOpts := []dualdht.Option{
		dualdht.DHTOption(dht.NamespacedValidator("pk", record.PublicKeyValidator{})),
		dualdht.DHTOption(dht.NamespacedValidator("ipns", ipns.Validator{KeyBook: h.Peerstore()})),
		dualdht.DHTOption(dht.Concurrency(10)),
		dualdht.DHTOption(dht.Mode(dht.ModeAuto)),
	}

	if ds != nil {
		dhtOpts = append(dhtOpts, dualdht.DHTOption(dht.Datastore(ds)))
	}

	return dualdht.New(ctx, h, dhtOpts...)
}

func savePrivateKey(privKey crypto.PrivKey) error {
	data, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("error marshaling private key: %v", err)
	}

	// Write the byte array to a file
	return os.WriteFile("bootstrap_priv.key", data, 0644)
}

func loadPrivateKey() (crypto.PrivKey, error) {
	data, err := os.ReadFile("bootstrap_priv.key")
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %v", err)
	}

	key, err := crypto.UnmarshalPrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling private key: %v", err)
	}

	return key, nil
}
