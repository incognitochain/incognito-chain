package pruner

const (
	// stateBloomFilePrefix is the filename prefix of state bloom filter.
	stateBloomFilePrefix = "statebloom"

	// stateBloomFilePrefix is the filename suffix of state bloom filter.
	stateBloomFileSuffix = "bf.gz"

	// stateBloomFileTempSuffix is the filename suffix of state bloom filter
	// while it is being written out to detect write aborts.
	stateBloomFileTempSuffix = ".tmp"

	// rangeCompactionThreshold is the minimal deleted entry number for
	// triggering range compaction. It's a quite arbitrary number but just
	// to avoid triggering range compaction because of small deletion.
	rangeCompactionThreshold = 100000
)
