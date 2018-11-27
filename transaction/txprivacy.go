package transaction

import (
	"fmt"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/privacy-protocol/zero-knowledge"
)

type TxPrivacy struct{
	Version  int8   `json:"Version"`
	Type     string `json:"Type"` // Transaction type
	LockTime int64  `json:"LockTime"`
	Fee      uint64 `json:"Fee"` // Fee applies: always consant

	SigPubKey []byte           `json:"SigPubKey, omitempty"` // 64 bytes
	Sig		    []byte           `json:"Sig, omitempty"`    // 64 bytes
	Proof 		zkp.PaymentProof

	PubKeyLastByte byte `json:"AddressLastByte"`

	txId       *common.Hash
	sigPrivKey *privacy.SpendingKey // is always private property of struct

	// this one is a hash id of requested tx
	// and is used inside response txs
	// so that we can determine pair of req/res txs
	// for example, BuySellRequestTx/BuySellResponseTx
	//RequestedTxID *common.Hash

	// all input of verify function
	// outputcoin []OutputCoin
}

func (tx * TxPrivacy) CreateTx(
	senderSK *privacy.SpendingKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoins []*privacy.InputCoin,
	fee uint64,
	noPrivacy bool,
) (*TxPrivacy, error){

	// Print list of all input coins
	fmt.Printf("List of all input coins before building tx:\n")
	for _, coin := range inputCoins {
		fmt.Printf("%+v\n", coin)
	}

	// Calculate sum of all output coins' value
	var sumOutputValue uint64
	for _, p := range paymentInfo {
		sumOutputValue += p.Amount
		fmt.Printf("[CreateTx] paymentInfo.Value: %+v, paymentInfo.PaymentAddress: %x\n", p.Amount, p.PaymentAddress.Pk)
	}

	// Calculate sum of all input coins' value
	var sumInputValue uint64
	for _, coin := range inputCoins {
		sumInputValue += coin.CoinDetails.Value
	}

	// Calculate over balance, it will be returned to sender
	overBalance := sumInputValue - sumOutputValue - fee

	// Check if sum of input coins' value is at least sum of output coins' value and tx fee
	if overBalance < 0 {
		return nil, fmt.Errorf("Input value less than output value")
	}

	// create sender's key set from sender's spending key
	senderFullKey := cashec.KeySet{}
	senderFullKey.ImportFromPrivateKeyByte((*senderSK)[:])

	// get public key last byte
	pkLastByte := senderFullKey.PaymentAddress.Pk[len(senderFullKey.PaymentAddress.Pk)-1]
	tx.PubKeyLastByte = pkLastByte

	// create new output coins
	outputCoins := make([]*privacy.OutputCoin, len(paymentInfo))

	// create new output coins with info: Pk, value, SND
	for i, pInfo := range paymentInfo{
		outputCoins[i] = new(privacy.OutputCoin)
		outputCoins[i].CoinDetails.Value = pInfo.Amount
		outputCoins[i].CoinDetails.PublicKey, _ = privacy.DecompressKey(pInfo.PaymentAddress.Pk)
		outputCoins[i].CoinDetails.SNDerivator = privacy.RandInt()
	}

	// if overBalance > 0, create a output coin with pk is pk's sender and value is overBalance
	if overBalance > 0{
		changeMoney := new(privacy.OutputCoin)
		changeMoney.CoinDetails.Value = overBalance
		changeMoney.CoinDetails.PublicKey, _ = privacy.DecompressKey(senderFullKey.PaymentAddress.Pk)
		changeMoney.CoinDetails.SNDerivator = privacy.RandInt()
	}

	// create zero knowledge proof of payment


	// encrypt coin details (Randomness)

	// sign tx


	return nil, nil
}

func (tx * TxPrivacy) SignTx(noPrivacy bool) error {
	if noPrivacy{
		//using ECDSA
	} else{
		//using Schnorr
	}

	return nil
}


