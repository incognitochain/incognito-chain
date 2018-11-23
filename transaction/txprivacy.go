package transaction

import (
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/privacy-protocol/zero-knowledge"
)

type TxPrivacy struct{
	Proof zkp.ZKProof
}

func (tx * TxPrivacy) CreateTx(
	paymentInfo []privacy.PaymentInfo,
	unspentCoins []privacy.SpendingCoin,
	rt privacy.MerkleRoot,





	){

}




