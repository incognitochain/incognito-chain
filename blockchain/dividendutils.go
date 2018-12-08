package blockchain

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/privacy-protocol/client"
	"github.com/ninjadotorg/constant/transaction"
)

func buildCoinbaseTx(
	pks, tks [][]byte,
	amounts []uint64,
	rt []byte,
	chainID byte,
) (*transaction.Tx, error) {
	// Create Proof for the joinsplit op
	inputs := make([]*client.JSInput, 2)
	inputs[0] = transaction.CreateRandomJSInput(nil)
	inputs[1] = transaction.CreateRandomJSInput(inputs[0].Key)
	dummyAddress := privacy.GeneratePaymentAddress(*inputs[0].Key)

	// Create new notes to send to 2 token holders at the same time
	outNote1 := &client.Note{Value: amounts[0], Apk: pks[0]}
	outNote2 := &client.Note{Value: amounts[1], Apk: pks[1]}
	totalAmount := outNote1.Value + outNote2.Value

	outputs := []*client.JSOutput{&client.JSOutput{}, &client.JSOutput{}}
	outputs[0].EncKey = tks[0]
	outputs[0].OutputNote = outNote1
	outputs[1].EncKey = tks[1]
	outputs[1].OutputNote = outNote2

	// Generate proof and sign tx
	tx, err := transaction.CreateEmptyTx(common.TxNormalType, nil, true)
	if err != nil {
		return nil, err
	}

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
	return tx, nil
}

func buildDividendTxs(
	infos []metadata.DividendInfo,
	rt []byte,
	chainID byte,
	proposal *metadata.DividendProposal,
) ([]*transaction.Tx, error) {
	if len(infos)%2 != 0 { // Add dummy receiver if needed
		infos = append(infos, metadata.DividendInfo{
			TokenHolder: privacy.GeneratePaymentAddress(privacy.GenerateSpendingKey([]byte{})),
			Amount:      0,
		})
	}

	numInfos := len(infos)
	txs := []*transaction.Tx{}
	for i := 0; i < numInfos; i += 2 {
		pks := [][]byte{infos[i].TokenHolder.Pk[:], infos[i+1].TokenHolder.Pk[:]}
		tks := [][]byte{infos[i].TokenHolder.Tk[:], infos[i+1].TokenHolder.Tk[:]}
		amounts := []uint64{infos[i].Amount, infos[i+1].Amount}
		tx, err := buildCoinbaseTx(pks, tks, amounts, rt, chainID)
		// TODO(@0xbunyip): return list of failed txs instead of error
		if err != nil {
			return nil, err
		}
		tx.Metadata = &metadata.Dividend{
			PayoutID: proposal.PayoutID,
			TokenID:  proposal.TokenID,
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
