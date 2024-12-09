package ipfs

import (
	"bytes"
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
	"github.com/libp2p/go-libp2p/core/pnet"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
)

// NewInMemoryDatastore provides a sync datastore that lives in-memory only and is not persisted.
func NewInMemoryDatastore() datastore.Batching {
	return dssync.MutexWrap(datastore.NewMapDatastore())
}

var connMgr, _ = connmgr.NewConnManager(100, 400, connmgr.WithGracePeriod(time.Minute))

func NewLibp2pHost(
	ctx context.Context,
	ds datastore.Batching) (host.Host, *dualdht.DHT, error) {
	var ddht *dualdht.DHT
	var err error

	// if node is bootstrap node, use a static priv key to persist node identity to ease the node connections
	priv, err := generatePrivKey()
	if err != nil {
		panic(err)
	}

	listenAddr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4001")
	if err != nil {
		return nil, nil, err
	}

	swarmKeyFile, err := os.ReadFile("./transcode/storage/ipfs/swarm.key")
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, nil, fmt.Errorf("swarm.key file not found (%s)", err)
	} else if err != nil {
		return nil, nil, err
	}

	psk, err := pnet.DecodeV1PSK(bytes.NewReader(swarmKeyFile))
	if err != nil {
		return nil, nil, fmt.Errorf("error loading swarm key: :%s", err)
	}

	// we are creating a LAN network so no need NAT or any security methods
	opts := []libp2p.Option{
		libp2p.Identity(priv),
		libp2p.PrivateNetwork(psk),
		libp2p.ListenAddrs(listenAddr),
		libp2p.ConnectionManager(connMgr),
		//libp2p.Security(libp2ptls.ID, libp2ptls.New),
		//libp2p.Security(noise.ID, noise.New),
		libp2p.Transport(tcp.NewTCPTransport),
		// libp2p.Transport(quic.NewTransport), -- remove QUIC cause QUIC does not support private network
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

func generatePrivKey() (crypto.PrivKey, error) {
	var finalPriv crypto.PrivKey
	var r io.Reader

	r = rand.New(rand.NewSource(time.Now().Unix()))
	finalPriv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
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
