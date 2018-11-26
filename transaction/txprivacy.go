package transaction

import (
	"github.com/ninjadotorg/constant/privacy-protocol"
)

type TxPrivacy struct{
	//Proof zkp.ZKProof

	Version  int8   `json:"Version"`
	Type     string `json:"Type"` // Transaction type
	LockTime int64  `json:"LockTime"`
	Fee      uint64 `json:"Fee"` // Fee applies: always consant

	Descs    []*JoinSplitDesc `json:"Descs"`
	JSPubKey []byte           `json:"JSPubKey,omitempty"` // 64 bytes
	JSSig    []byte           `json:"JSSig,omitempty"`    // 64 bytes

	AddressLastByte byte `json:"AddressLastByte"`

	txId       *common.Hash
	sigPrivKey *privacy.SpendingKey // is always private property of struct

	// this one is a hash id of requested tx
	// and is used inside response txs
	// so that we can determine pair of req/res txs
	// for example, BuySellRequestTx/BuySellResponseTx
	RequestedTxID *common.Hash
}

func (tx * TxPrivacy) CreateTx(
	paymentInfo []*privacy.PaymentInfo,
	inputCoins []*privacy.InputCoin,
	fee uint64,
	noPrivacy bool,
) (*TxPrivacy, error){

	return nil, nil
}
