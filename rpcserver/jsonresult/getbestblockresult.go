package jsonresult

type GetBestBlockResult struct {
	BestBlocks map[string]GetBestBlockItem
}

type GetBestBlockItem struct {
	Height int32
	Hash   string
}

type GetBestBlockHashResult struct {
	BestBlockHashes map[string]string
}
