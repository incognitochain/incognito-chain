package jsonresult

type CrossShardPoolResult struct {
	PendingBlockHeight []BlockHeights `json:"InexecutableBlock"`
	ValidBlockHeight   []BlockHeights `json:"ExecutableBlock"`
}
type ShardToBeaconPoolResult struct {
	PendingBlockHeight []BlockHeights `json:"InexecutableBlock"`
	ValidBlockHeight   []BlockHeights `json:"ExecutableBlock"`
}
type BlockHeights struct {
	ShardID         byte     `json:"ShardID"`
	BlockHeightList []uint64 `json:"BlockHeightList"`
}
type ShardBlockPoolResult struct {
	ShardID            byte     `json:"ShardID"`
	ValidBlockHeight   []uint64 `json:"ExecutableBlock"`
	PendingBlockHeight []uint64 `json:"InexecutableBlock"`
}

type BeaconBlockPoolResult struct {
	ValidBlockHeight   []uint64 `json:"ExecutableBlock"`
	PendingBlockHeight []uint64 `json:"InexecutableBlock"`
}
