package jsonresult

type GetMiningInfoResult struct {
	Blocks         uint64 `json:"Blocks"`
	CurrentBlockTx int    `json:"CurrentBlockTx"`
	PoolSize       int    `json:"PoolSize"`
	Chain          string `json:"Chain"`
}
