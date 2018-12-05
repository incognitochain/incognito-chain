package transaction

import (
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/privacy-protocol/zero-knowledge"
	"math/big"
)

// CreateTxSalary
// Blockchain use this tx to pay a reward(salary) to miner of chain
// #1 - salary:
// #2 - receiverAddr:
// #3 - privKey:
// #4 - snDerivators:
func CreateTxSalary(
	salary uint64,
	receiverAddr *privacy.PaymentAddress,
	privKey *privacy.SpendingKey,
	snDerivators []*big.Int,
) (*Tx, error) {

	tx := new(Tx)
	// Todo: check
	tx.Type = "Salary"
	// assign fee tx = 0
	tx.Fee = 0
	tx.snDerivators = snDerivators

	// create new output coins with info: Pk, value, SND, randomness, last byte pk, coin commitment
	tx.Proof = new(zkp.PaymentProof)
	tx.Proof.OutputCoins = make([]*privacy.OutputCoin, 1)
	tx.Proof.OutputCoins[0] = new(privacy.OutputCoin)
	tx.Proof.OutputCoins[0].CoinDetails.Value = salary
	tx.Proof.OutputCoins[0].CoinDetails.PublicKey, _ = privacy.DecompressKey(receiverAddr.Pk)
	tx.Proof.OutputCoins[0].CoinDetails.PubKeyLastByte = tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress()[len(tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress()) - 1]
	tx.Proof.OutputCoins[0].CoinDetails.Randomness = privacy.RandInt()

	sndOut := new(big.Int)
	ok := true
	for ok {
		sndOut = privacy.RandInt()
		ok = CheckSNDExistence(snDerivators, sndOut)
	}
	snDerivators = append(snDerivators, sndOut)

	tx.Proof.OutputCoins[0].CoinDetails.SNDerivator = sndOut

	// create coin commitment
	tx.Proof.OutputCoins[0].CoinDetails.CommitAll()

	// sign Tx
	var err error
	tx.SigPubKey = receiverAddr.Pk
	tx.sigPrivKey = *privKey
	err = tx.SignTx(false)
	if err != nil{
		return nil, err
	}


	/*// Create Proof for the joinsplit op
	*//*inputs := make([]*client.JSInput, 2)
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
	}*//*
	return tx, nil*/
	return tx, nil
}

func ValidateTxSalary(
	tx *Tx,
	receiverAddr *privacy.PaymentAddress,
) (bool, error) {

	// check whether output coin's public key is receiver's public key or not
	pubKey, err := privacy.DecompressKey(receiverAddr.Pk)
	if err != nil{
		return false, err
	}
	if tx.Proof.OutputCoins[0].CoinDetails.PublicKey.IsEqual(pubKey) {
		return false, err
	}

	// check whether output coin's SND exists in SND list or not
	if !CheckSNDExistence(tx.snDerivators, tx.Proof.OutputCoins[0].CoinDetails.SNDerivator) {
		return false, err
	}

	// check output coin's coin commitment is calculated correctly
	cmTmp := tx.Proof.OutputCoins[0].CoinDetails.PublicKey
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMul(big.NewInt(int64(tx.Proof.OutputCoins[0].CoinDetails.Value))))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMul(tx.Proof.OutputCoins[0].CoinDetails.SNDerivator))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMul(new(big.Int).SetBytes([]byte{tx.Proof.OutputCoins[0].CoinDetails.PubKeyLastByte})))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMul(tx.Proof.OutputCoins[0].CoinDetails.Randomness))
	if !cmTmp.IsEqual(tx.Proof.OutputCoins[0].CoinDetails.CoinCommitment) {
		return false, nil
	}

 return true, nil
}


