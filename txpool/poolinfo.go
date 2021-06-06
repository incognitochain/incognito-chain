package txpool

type GetMempoolInfoTx struct {
	TxID     string
	LockTime int64
}

func (infoTx *GetMempoolInfoTx) GetTxID() string {
	return infoTx.TxID
}
func (infoTx *GetMempoolInfoTx) GetLockTime() int64 {
	return infoTx.LockTime
}

type MempoolInfoTx interface {
	GetTxID() string
	GetLockTime() int64
}

type GetMempoolInfo struct {
	Size          int
	Bytes         uint64
	Usage         uint64
	MaxMempool    uint64
	MempoolMinFee uint64
	MempoolMaxFee uint64
	ListTxs       []MempoolInfoTx
}

func (info *GetMempoolInfo) GetSize() int {
	return info.Size
}

func (info *GetMempoolInfo) GetBytes() uint64 {
	return info.Bytes
}
func (info *GetMempoolInfo) GetUsage() uint64 {
	return info.Usage
}
func (info *GetMempoolInfo) GetMaxMempool() uint64 {
	return info.MaxMempool
}
func (info *GetMempoolInfo) GetMempoolMinFee() uint64 {
	return info.MempoolMinFee
}
func (info *GetMempoolInfo) GetMempoolMaxFee() uint64 {
	return info.MempoolMaxFee
}
func (info *GetMempoolInfo) GetListTxs() []MempoolInfoTx {
	return info.ListTxs
}

type MempoolInfo interface {
	GetSize() int
	GetBytes() uint64
	GetUsage() uint64
	GetMaxMempool() uint64
	GetMempoolMinFee() uint64
	GetMempoolMaxFee() uint64
	GetListTxs() []MempoolInfoTx
}
