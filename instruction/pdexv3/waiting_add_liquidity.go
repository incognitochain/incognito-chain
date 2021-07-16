package pdexv3

type WaitingAddLiquidity struct {
	PoolPairID          string // only "" for the first contribution of pool
	OtaPublicKeyRefund  string // refund contributed token
	OtaTxRandomRefund   string
	OtaPublicKeyReceive string // receive nfct
	OtaTxRandom         string
	TokenID             string
	TokenAmount         uint64
	Amplifier           uint // only set for the first contribution
	TxReqID             string
	ShardID             byte
}
