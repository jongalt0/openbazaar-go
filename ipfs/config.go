package ipfs

import (
	"context"

	dht "gx/ipfs/QmSY3nkMNLzh9GdbFKK5tT7YMfLpf52iUZ8ZRkr29MJaa5/go-libp2p-kad-dht"
	dhtopts "gx/ipfs/QmSY3nkMNLzh9GdbFKK5tT7YMfLpf52iUZ8ZRkr29MJaa5/go-libp2p-kad-dht/opts"
	ds "gx/ipfs/QmUadX5EcvrBmxAV9sE7wUWtWSqxns5K84qKJBixmcT1w9/go-datastore"
	p2phost "gx/ipfs/QmYrWiWM4qtrnCeT3R14jY3ZZyirDNJgwK57q4qFYePgbd/go-libp2p-host"
	routing "gx/ipfs/QmYxUdYY9S6yg5tSPVin5GFTvtfsLauVcr7reHDD3dM8xf/go-libp2p-routing"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	record "gx/ipfs/QmbeHtaBy9nZsW4cHRcvgVY4CnDhXudE2Dr6qDxS7yg9rX/go-libp2p-record"

	ipfscore "github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/repo"
)

var routerCacheURI string

// PrepareIPFSConfig builds the configuration options for the internal
// IPFS node.
func PrepareIPFSConfig(r repo.Repo, routerAPIEndpoint string, testEnable, regtestEnable bool) *ipfscore.BuildCfg {
	routerCacheURI = routerAPIEndpoint
	ncfg := &ipfscore.BuildCfg{
		Repo:   r,
		Online: true,
		ExtraOpts: map[string]bool{
			"mplex":  true,
			"ipnsps": true,
		},
	}

	// regtest and test are never enabled together
	ncfg.Routing = constructRouting
	if regtestEnable {
		ncfg.Routing = constructDHTRouting
	} else if testEnable {
		ncfg.Routing = constructTestnetDHTRouting
	}
	return ncfg
}

func constructRouting(ctx context.Context, host p2phost.Host, dstore ds.Batching, validator record.Validator) (routing.IpfsRouting, error) {
	dhtRouting, err := dht.New(
		ctx, host,
		dhtopts.Datastore(dstore),
		dhtopts.Validator(validator),
	)
	if err != nil {
		return nil, err
	}
	apiRouter := NewAPIRouter(routerCacheURI)
	cachingRouter := NewCachingRouter(dhtRouting, &apiRouter)
	return cachingRouter, nil
}

func constructDHTRouting(ctx context.Context, host p2phost.Host, dstore ds.Batching, validator record.Validator) (routing.IpfsRouting, error) {
	return dht.New(
		ctx, host,
		dhtopts.Datastore(dstore),
		dhtopts.Validator(validator),
	)
}

func constructTestnetDHTRouting(ctx context.Context, host p2phost.Host, dstore ds.Batching, validator record.Validator) (routing.IpfsRouting, error) {
	testnetDHT := protocol.ID("/openbazaar/kad/testnet/1.0.0")
	testnetApp := protocol.ID("/openbazaar/app/testnet/1.0.0")
	dhtRouting, err := dht.New(
		ctx, host,
		dhtopts.Datastore(dstore),
		dhtopts.Validator(validator),
		dhtopts.Protocols(testnetDHT, testnetApp),
	)
	if err != nil {
		return nil, err
	}
	apiRouter := NewAPIRouter(routerCacheURI)
	cachingRouter := NewCachingRouter(dhtRouting, &apiRouter)
	return cachingRouter, nil
}
