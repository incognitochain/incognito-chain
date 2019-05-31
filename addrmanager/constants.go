package addrmanager

import "time"

const (
	version = 1

	// dumpAddressInterval is the interval used to dump the address
	// cache to disk for future use.
	dumpAddressInterval = time.Second * 10
)
