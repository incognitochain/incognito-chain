package jsonresult

type CrossShardPoolResult struct {
	PendingBlockHeight []BlockHeights `json:"PendingBlockHeight"`
	ValidBlockHeight   []BlockHeights `json:"PendingBlockHeight"`
}
type BlockHeights struct {
	ShardID         byte     `json:"ShardID"`
	BlockHeightList []uint64 `json:"BlockHeightList"`
}
