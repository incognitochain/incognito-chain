package jsonresult

type GetBestBlockResult struct {
	// BestBlocks map[string]GetBestBlockItem `json:"BestBlocks"`
	BestBlocks map[int]GetBestBlockItem `json:"BestBlocks"`
}

type GetBestBlockItem struct {
	Height           uint64 `json:"Height"`
	Hash             string `json:"Hash"`
	TotalTxs         uint64 `json:"TotalTxs"`
	SalaryFund       uint64 `json:"SalaryFund"`
	BasicSalary      uint64 `json:"BasicSalary"`
	SalaryPerTx      uint64 `json:"SalaryPerTx"`
	BlockProducer    string `json:"BlockProducer"`
	BlockProducerSig string `json:"BlockProducerSig"`
}

type GetBestBlockHashResult struct {
	// BestBlockHashes map[byte]string `json:"BestBlockHashes"`
	BestBlockHashes map[int]string `json:"BestBlockHashes"`
}
