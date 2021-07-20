package pdex

//Contribution ...
type Contribution struct {
	poolPairID      string // only "" for the first contribution of pool
	receiverAddress string // receive nfct
	refundAddress   string // refund pToken
	tokenID         string
	tokenAmount     uint64
	amplifier       uint // only set for the first contribution
	txReqID         string
	shardID         byte
}

func NewContributionWithValue(
	poolPairID, receiverAddress, refundAddress,
	tokenID, txReqID string,
	tokenAmount uint64, amplifier uint,
	shardID byte,
) *Contribution {
	return &Contribution{
		poolPairID:      poolPairID,
		receiverAddress: receiverAddress,
		refundAddress:   refundAddress,
		tokenID:         tokenID,
		tokenAmount:     tokenAmount,
		amplifier:       amplifier,
		txReqID:         txReqID,
		shardID:         shardID,
	}
}

func NewContribution() *Contribution {
	return &Contribution{}
}
