package addrmanager

import "time"

const (
	version = 1

	// DumpAddressInterval is the interval used to dump the address
	// cache to disk for future use. Every 10 second, saving all
	// connected address into file to reuse in the future
	DumpAddressInterval = time.Second * 10
)
