package proof

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	zkp "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/privacy/proof/agg_interface"
)

// Paymentproof
type Proof interface {
	GetVersion() uint8

	Init()
	GetInputCoins() []*coin.InputCoin
	GetOutputCoins() []*coin.OutputCoin
	GetAggregatedRangeProof() *agg_interface.AggregatedRangeProof

	SetInputCoins([]*coin.InputCoin)
	SetOutputCoins([]*coin.OutputCoin)

	Bytes() []byte
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error

	Verify(hasPrivacy bool, pubKey key.PublicKey, fee uint64, db database.DatabaseInterface, shardID byte, tokenID *common.Hash, isBatch bool) (bool, error)
}

type ProofV1 = zkp.PaymentProof
type ProofV2 = privacy_v2.PaymentProof
