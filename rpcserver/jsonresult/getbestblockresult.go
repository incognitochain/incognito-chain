package jsonresult

type GetBestBlockResult struct {
	BestBlocks map[string]GetBestBlockItem `json:"BestBlocks"`
}

type GetBestBlockItem struct {
	Height   int32  `json:"Height"`
	Hash     string `json:"Hash"`
	TotalTxs uint64 `json:"TotalTxs"`
}

type GetBestBlockHashResult struct {
	BestBlockHashes map[string]string `json:"BestBlockHashes"`
}
