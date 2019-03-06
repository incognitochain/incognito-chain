package jsonresult

// GetBlockChainInfoResult models the data returned from the getblockchaininfo
// command.
type GetBlockChainInfoResult struct {
	ChainName    string                   `json:"ChainName"`
	BestBlocks   map[int]GetBestBlockItem `json:"BestBlocks"`
	ActiveShards int                      `json:"ActiveShards"`
}
