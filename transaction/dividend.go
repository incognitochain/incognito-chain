package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy/client"
)

const MaxDivTxsPerBlock = 1000
const PayoutFrequency = 1000 // Payout dividend every 1000 blocks

type PayoutProposal struct {
	TotalAmount uint64 // total Constant to pay dividend
	PayoutID    uint64 // increasing ID for each type of token
}

type DividendInfo struct {
	TokenHolder client.PaymentAddress
	Amount      uint64
}

func BuildDividendTxs(
	infos []DividendInfo,
	rt []byte,
	chainID byte,
) ([]*Tx, error) {
	numInfos := len(infos)
	txs := []*Tx{}
	for i := 0; i < numInfos; i += 2 {
		// Create Proof for the joinsplit op
		inputs := make([]*client.JSInput, 2)
		inputs[0] = CreateRandomJSInput(nil)
		inputs[1] = CreateRandomJSInput(inputs[0].Key)
		dummyAddress := client.GenPaymentAddress(*inputs[0].Key)

		// Create new notes to send to 2 token holders at the same time
		outNote1 := &client.Note{Value: infos[i].Amount, Apk: infos[i].TokenHolder.Apk}
		outNote2 := &client.Note{Value: infos[i+1].Amount, Apk: infos[i+1].TokenHolder.Apk}
		totalAmount := outNote1.Value + outNote2.Value

		outputs := []*client.JSOutput{&client.JSOutput{}, &client.JSOutput{}}
		outputs[0].EncKey = infos[i].TokenHolder.Pkenc
		outputs[0].OutputNote = outNote1
		outputs[1].EncKey = infos[i+1].TokenHolder.Pkenc
		outputs[1].OutputNote = outNote2

		// Generate proof and sign tx
		tx, err := CreateEmptyTx(common.TxTokenDividendPayout)
		if err != nil {
			return nil, err
		}

		// TODO(@0xbunyip): use DCB hard-coded account here
		tx.AddressLastByte = dummyAddress.Apk[len(dummyAddress.Apk)-1]
		rtMap := map[byte][]byte{chainID: rt}
		inputMap := map[byte][]*client.JSInput{chainID: inputs}

		err = tx.BuildNewJSDesc(inputMap, outputs, rtMap, totalAmount, 0, true)
		if err != nil {
			return nil, err
		}
		err = tx.SignTx()
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
