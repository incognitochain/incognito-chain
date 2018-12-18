package transaction

import (
	"fmt"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
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
	db database.DatabaseInterface,
) (*Tx, error) {

	tx := new(Tx)
	tx.Type = common.TxSalaryType
	// assign fee tx = 0
	tx.Fee = 0

	var err error
	// create new output coins with info: Pk, value, SND, randomness, last byte pk, coin commitment
	tx.Proof = new(zkp.PaymentProof)
	tx.Proof.OutputCoins = make([]*privacy.OutputCoin, 1)
	tx.Proof.OutputCoins[0] = new(privacy.OutputCoin)
	//tx.Proof.OutputCoins[0].CoinDetailsEncrypted = new(privacy.CoinDetailsEncrypted).Init()
	tx.Proof.OutputCoins[0].CoinDetails = new(privacy.Coin)
	tx.Proof.OutputCoins[0].CoinDetails.Value = salary
	tx.Proof.OutputCoins[0].CoinDetails.PublicKey, err = privacy.DecompressKey(receiverAddr.Pk)
	if err != nil {
		return nil, err
	}
	tx.Proof.OutputCoins[0].CoinDetails.Randomness = privacy.RandInt()

	sndOut := privacy.RandInt()
	for true {
		lastByte := receiverAddr.Pk[len(receiverAddr.Pk)-1]
		chainIdSender, err := common.GetTxSenderChain(lastByte)
		ok, err := CheckSNDerivatorExistence(sndOut, chainIdSender, db)
		if err != nil {
			return nil, err
		}
		if ok {
			sndOut = privacy.RandInt()
		} else {
			break
		}
	}

	tx.Proof.OutputCoins[0].CoinDetails.SNDerivator = sndOut

	// create coin commitment
	tx.Proof.OutputCoins[0].CoinDetails.CommitAll()
	// get last byte
	tx.PubKeyLastByteSender = receiverAddr.Pk[len(receiverAddr.Pk)-1]

	// sign Tx
	tx.SigPubKey = receiverAddr.Pk
	tx.sigPrivKey = *privKey
	err = tx.SignTx(false)
	if err != nil {
		return nil, err
	}

	if len(tx.Proof.InputCoins) > 0 {
		fmt.Println(11111)
	}
	return tx, nil
}

func ValidateTxSalary(
	tx *Tx,
	db database.DatabaseInterface,
) bool {
	// verify signature
	valid, err := tx.VerifySigTx(false)
	if valid == false {
		if err != nil {
			fmt.Printf("Error verifying signature of tx: %+v", err)
		}
		return false
	}

	// check whether output coin's SND exists in SND list or not
	lastByte := tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress()[len(tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress())-1]
	chainIdSender, err := common.GetTxSenderChain(lastByte)
	if ok, err := CheckSNDerivatorExistence(tx.Proof.OutputCoins[0].CoinDetails.SNDerivator, chainIdSender, db); ok || err != nil {
		return false
	}

	// check output coin's coin commitment is calculated correctly
	cmTmp := tx.Proof.OutputCoins[0].CoinDetails.PublicKey
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMul(big.NewInt(int64(tx.Proof.OutputCoins[0].CoinDetails.Value))))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMul(tx.Proof.OutputCoins[0].CoinDetails.SNDerivator))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMul(new(big.Int).SetBytes([]byte{tx.Proof.OutputCoins[0].CoinDetails.GetPubKeyLastByte()})))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMul(tx.Proof.OutputCoins[0].CoinDetails.Randomness))
	if !cmTmp.IsEqual(tx.Proof.OutputCoins[0].CoinDetails.CoinCommitment) {
		return false
	}

	return true
}
