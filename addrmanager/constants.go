package addrmanager

import "time"

const (
	Version = 1

	// DumpAddressInterval is the interval used to dump the address
	// cache to disk for future use.
	DumpAddressInterval = time.Second * 10
)
