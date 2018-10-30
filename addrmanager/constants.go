package addrmanager

import "time"

const (
	// dumpAddressInterval is the interval used to dump the address
	// cache to disk for future use.
	dumpAddressInterval = time.Second * 10

	// newBucketCount is the number of buckets that we spread new addresses
	// over.
	newBucketCount = 1024

	// triedBucketCount is the number of buckets we split tried
	// addresses over.
	triedBucketCount = 64
)
