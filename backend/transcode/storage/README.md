kubo.go uses kubo as a library to create a IPFS node but somehow it is producing an error. So I changed the name from kubo.go to kubo.go.skip for later changes.


ERROR:

    # github.com/ipfs/kubo/core/node/libp2p
    /home/sen1or/go/pkg/mod/github.com/ipfs/kubo@v0.31.0/core/node/libp2p/dns.go:9:57: cannot use rslv (variable of type *madns.Resolver) as network.MultiaddrDNSResolver value in argument to libp2p.MultiaddrResolver: *madns.Resolver does not implement network.MultiaddrDNSResolver (missing method ResolveDNSAddr)
