package jsonresult

// GetBlockChainInfoResult models the data returned from the getblockchaininfo
// command.
type GetBlockChainInfoResult struct {
	Chain                string   `json:"blockChain"`
	Blocks               int      `json:"Blocks"`
	Headers              int32    `json:"Headers"`
	BestBlockHash        []string `json:"BestBlockHash"`
	Difficulty           uint32   `json:"Difficulty"`
	MedianTime           int64    `json:"MedianTime"`
	VerificationProgress float64  `json:"VerificationProgress,omitempty"`
	Pruned               bool     `json:"Pruned"`
	PruneHeight          int32    `json:"PruneHeight,omitempty"`
	ChainWork            string   `json:"ChainWork,omitempty"`
}
