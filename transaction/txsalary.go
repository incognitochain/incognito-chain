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
) (*Tx, error) {

	tx := new(Tx)
	// Todo: check
	tx.Type = "Salary"
	// assign fee tx = 0
	tx.Fee = 0

	// create new output coins with info: Pk, value, SND, randomness, last byte pk, coin commitment
	tx.Proof = new(zkp.PaymentProof)
	tx.Proof.OutputCoins = make([]*privacy.OutputCoin, 1)
	tx.Proof.OutputCoins[0] = new(privacy.OutputCoin)
	tx.Proof.OutputCoins[0].CoinDetails.Value = salary
	tx.Proof.OutputCoins[0].CoinDetails.PublicKey, _ = privacy.DecompressKey(receiverAddr.Pk)
	tx.Proof.OutputCoins[0].CoinDetails.PubKeyLastByte = tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress()[len(tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress()) - 1]
	tx.Proof.OutputCoins[0].CoinDetails.Randomness = privacy.RandInt()

	//sndOut := new(big.Int)
	sndOut := privacy.RandInt()
	for CheckSNDExistence(sndOut) {
		sndOut = privacy.RandInt()
	}

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

	return tx, nil
}

func ValidateTxSalary(
	tx *Tx,
) bool {

	// check whether output coin's SND exists in SND list or not
	if CheckSNDExistence(tx.Proof.OutputCoins[0].CoinDetails.SNDerivator) {
		return false
	}

	// check output coin's coin commitment is calculated correctly
	cmTmp := tx.Proof.OutputCoins[0].CoinDetails.PublicKey
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMul(big.NewInt(int64(tx.Proof.OutputCoins[0].CoinDetails.Value))))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMul(tx.Proof.OutputCoins[0].CoinDetails.SNDerivator))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMul(new(big.Int).SetBytes([]byte{tx.Proof.OutputCoins[0].CoinDetails.PubKeyLastByte})))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMul(tx.Proof.OutputCoins[0].CoinDetails.Randomness))
	if !cmTmp.IsEqual(tx.Proof.OutputCoins[0].CoinDetails.CoinCommitment) {
		return false
	}

 return true
}


