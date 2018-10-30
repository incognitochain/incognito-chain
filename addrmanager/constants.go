package addrmanager

import "time"

const (
	Version = 1

	// DumpAddressInterval is the interval used to dump the address
	// cache to disk for future use.
	DumpAddressInterval = time.Second * 10

	// NewBucketCount is the number of buckets that we spread new addresses
	// over.
	NewBucketCount = 1024

	// TriedBucketCount is the number of buckets we split tried
	// addresses over.
	TriedBucketCount = 64
)
