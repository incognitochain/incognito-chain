package addrmanager

import "time"

const (
	// version of addrmanager
	version = 1

	// file to storage connected peer for reusing when restart node
	dataFile = "peer.json"

	// DumpAddressInterval is the interval used to dump the address
	// cache to disk for future use. Every 60 second, automatically saving all
	// connected address into file to reuse in the future
	dumpAddressInterval = time.Second * 60

	maxLengthPeerPretty = 46
)
