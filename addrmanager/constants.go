package addrmanager

import "time"

const (
	// version of addrmanager
	version = 1

	// DumpAddressInterval is the interval used to dump the address
	// cache to disk for future use. Every 10 second, automatically saving all
	// connected address into file to reuse in the future
	dumpAddressInterval = time.Second * 10

	maxLengthPeerPretty = 46
)
