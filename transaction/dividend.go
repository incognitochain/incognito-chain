package transaction

import (
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/privacy-protocol/client"
)

const MaxDivTxsPerBlock = 1000
const PayoutFrequency = 1000 // Payout dividend every 1000 blocks

type DividendPayout struct {
	PayoutID uint64
	TokenID  *common.Hash
}

type TxDividendPayout struct {
	*Tx
	DividendPayout
}

func (tx *TxDividendPayout) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	record += fmt.Sprintf("%d", tx.PayoutID)
	record += string(tx.TokenID[:])

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxDividendPayout) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.Tx.ValidateTransaction() {
		return false
	}

	return true
}

func (tx *TxDividendPayout) GetType() string {
	return common.TxDividendPayout
}

type PayoutProposal struct {
	TotalAmount uint64 // total Constant to pay dividend
	PayoutID    uint64 // increasing ID for each type of token
	TokenID     *common.Hash
}

type DividendInfo struct {
	TokenHolder privacy.PaymentAddress
	Amount      uint64
}

func BuildDividendTxs(
	infos []DividendInfo,
	rt []byte,
	chainID byte,
	proposal *PayoutProposal,
) ([]*TxDividendPayout, error) {
	if len(infos)%2 != 0 { // Add dummy receiver if needed
		infos = append(infos, DividendInfo{
			TokenHolder: privacy.GeneratePaymentAddress(privacy.GenerateSpendingKey([]byte{})),
			Amount:      0,
		})
	}

	numInfos := len(infos)
	txs := []*TxDividendPayout{}
	for i := 0; i < numInfos; i += 2 {
		// Create Proof for the joinsplit op
		inputs := make([]*client.JSInput, 2)
		inputs[0] = CreateRandomJSInput(nil)
		inputs[1] = CreateRandomJSInput(inputs[0].Key)
		dummyAddress := privacy.GeneratePaymentAddress(*inputs[0].Key)

		// Create new notes to send to 2 token holders at the same time
		outNote1 := &client.Note{Value: infos[i].Amount, PaymentAddress: infos[i].TokenHolder}
		outNote2 := &client.Note{Value: infos[i+1].Amount, PaymentAddress: infos[i+1].TokenHolder}
		totalAmount := outNote1.Value + outNote2.Value

		outputs := []*client.JSOutput{&client.JSOutput{}, &client.JSOutput{}}
		outputs[0].EncKey = infos[i].TokenHolder.Tk
		outputs[0].OutputNote = outNote1
		outputs[1].EncKey = infos[i+1].TokenHolder.Tk
		outputs[1].OutputNote = outNote2

		// Generate proof and sign tx
		tx, err := CreateEmptyTx(common.TxDividendPayout)
		if err != nil {
			return nil, err
		}

		// TODO(@0xbunyip): use DCB hard-coded account here
		tx.AddressLastByte = dummyAddress.Pk[len(dummyAddress.Pk)-1]
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
		txs = append(txs, &TxDividendPayout{
			DividendPayout: DividendPayout{
				PayoutID: proposal.PayoutID,
				TokenID:  proposal.TokenID,
			},
			Tx: tx,
		})
	}
	return txs, nil
}
