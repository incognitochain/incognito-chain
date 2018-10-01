package jsonresult

type GetMiningInfoResult struct {
	Blocks         uint64 `json:"Blocks"`
	CurrentBlockTx int    `json:"CurrentBlockTx"`
	Difficulty     uint32 `json:"Difficulty"`
	PoolSize       int    `json:"PoolSize"`
	Chain          string `json:"Chain"`
}
