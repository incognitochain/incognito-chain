package transaction

import (
	"github.com/ninjadotorg/constant/privacy-protocol"
)

type TxPrivacy struct{
	//Proof zkp.ZKProof
}

func (tx * TxPrivacy) CreateTx(
	paymentInfo []*privacy.PaymentInfo,
	inputCoins []*privacy.InputCoin,
	fee uint64,
	noPrivacy bool,
) (*TxPrivacy, error){

	return nil, nil
}
