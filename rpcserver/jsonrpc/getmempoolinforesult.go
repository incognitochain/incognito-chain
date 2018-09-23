package jsonrpc

type GetMempoolInfo struct {
	Size          int    `json:"Size"`
	Bytes         uint64 `json:"Bytes"`
	Usage         uint64 `json:"Usage"`
	MaxMempool    uint64 `json:"MaxMempool"`
	MempoolMinFee uint64 `json:"MempoolMinFee"`
	MempoolMaxFee uint64 `json:"MempoolMaxFee"`
}
