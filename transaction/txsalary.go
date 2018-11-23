package transaction

import (
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/privacy-protocol/client"
	"github.com/ninjadotorg/constant/common"
)

// CreateTxSalary
// Blockchain use this tx to pay a reward(salary) to miner of chain
// #1 - salary:
// #2 - receiverAddr:
// #3 - rt
// #4 - chainID
func CreateTxSalary(
	salary uint64,
	receiverAddr *privacy.PaymentAddress,
	rt []byte,
	chainID byte,
) (*Tx, error) {
	// Create Proof for the joinsplit op
	inputs := make([]*client.JSInput, 2)
	inputs[0] = CreateRandomJSInput(nil)
	inputs[1] = CreateRandomJSInput(inputs[0].Key)
	dummyAddress := client.GenPaymentAddress(*inputs[0].Key)

	// Create new notes: first one is salary UTXO, second one has 0 value
	outNote := &client.Note{Value: salary, Apk: receiverAddr.Pk}
	placeHolderOutputNote := &client.Note{Value: 0, Apk: receiverAddr.Pk}

	outputs := []*client.JSOutput{&client.JSOutput{}, &client.JSOutput{}}
	outputs[0].EncKey = receiverAddr.Tk
	outputs[0].OutputNote = outNote
	outputs[1].EncKey = receiverAddr.Tk
	outputs[1].OutputNote = placeHolderOutputNote

	// Generate proof and sign tx
	tx, err := CreateEmptyTx(common.TxSalaryType, nil, true)
	if err != nil {
		return nil, err
	}
	tx.AddressLastByte = dummyAddress.Apk[len(dummyAddress.Apk)-1]
	rtMap := map[byte][]byte{chainID: rt}
	inputMap := map[byte][]*client.JSInput{chainID: inputs}

	// NOTE: always pay salary with constant coin
	err = tx.BuildNewJSDesc(inputMap, outputs, rtMap, salary, 0, true)
	if err != nil {
		return nil, err
	}
	err = tx.SignTx()
	if err != nil {
		return nil, err
	}
	return tx, nil
}
