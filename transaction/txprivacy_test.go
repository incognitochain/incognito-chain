package transaction

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func TestUnmarshalJSON(t *testing.T) {
	key, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = key.KeySet.InitFromPrivateKey(&key.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	paymentAddress := key.KeySet.PaymentAddress
	responseMeta, err := metadata.NewWithDrawRewardResponse(&common.Hash{})
	tx, err := BuildCoinBaseTxByCoinID(NewBuildCoinBaseTxByCoinIDParams(&paymentAddress, 10, &key.KeySet.PrivateKey, db, responseMeta, common.Hash{}, NormalCoinType, "PRV", 0))
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, tx)
	assert.Equal(t, uint64(10), tx.(*Tx).Proof.GetOutputCoins()[0].CoinDetails.GetValue())
	assert.Equal(t, common.PRVCoinID.String(), tx.GetTokenID().String())

	jsonStr, err := json.Marshal(tx)
	assert.Equal(t, nil, err)
	fmt.Println(string(jsonStr))

	tx1 := Tx{}
	//err = json.Unmarshal(jsonStr, &tx1)
	err = tx1.UnmarshalJSON(jsonStr)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(10), tx1.Proof.GetOutputCoins()[0].CoinDetails.GetValue())
	assert.Equal(t, common.PRVCoinID.String(), tx1.GetTokenID().String())
}

func TestInitTx(t *testing.T) {
	senderKey, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = senderKey.KeySet.InitFromPrivateKey(&senderKey.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	senderPaymentAddress := senderKey.KeySet.PaymentAddress
	senderPublicKey := senderPaymentAddress.Pk

	shardID := common.GetShardIDFromLastByte(senderKey.KeySet.PaymentAddress.Pk[len(senderKey.KeySet.PaymentAddress.Pk)-1])

	// coin base tx to mint PRV
	mintedAmount := 1000
	coinBaseTx, err := BuildCoinBaseTxByCoinID(NewBuildCoinBaseTxByCoinIDParams(&senderPaymentAddress, uint64(mintedAmount), &senderKey.KeySet.PrivateKey, db, nil, common.Hash{}, NormalCoinType, "PRV", 0))

	isValidSanity, err := coinBaseTx.ValidateSanityData(nil)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, isValidSanity)

	// store output coin's coin commitments in coin base tx
	db.StoreCommitments(
		common.PRVCoinID,
		senderPaymentAddress.Pk,
		[][]byte{coinBaseTx.(*Tx).Proof.GetOutputCoins()[0].CoinDetails.GetCoinCommitment().Compress()},
		shardID)

	// get output coins from coin base tx to create new tx
	coinBaseOutput := ConvertOutputCoinToInputCoin(coinBaseTx.(*Tx).Proof.GetOutputCoins())

	// init new tx without privacy
	tx1 := Tx{}
	// calculate serial number for input coins
	coinBaseOutput[0].CoinDetails.SetSerialNumber(privacy.PedCom.G[privacy.PedersenPrivateKeyIndex].Derive(new(big.Int).SetBytes(senderKey.KeySet.PrivateKey),
		coinBaseOutput[0].CoinDetails.GetSNDerivator()))

	// receiver's address
	receiverPaymentAddress, _ := wallet.Base58CheckDeserialize("1Uv3BkYiWy9Mjt1yBa4dXBYKo3az22TeCVEpeXN93ieJ8qhrTDuUZBzsPZWjjP2AeRQnjw1y18iFPHTRuAqqufwVC1vNUAWs4wHFbbWC2")
	// transfer amount
	transferAmount := 5
	hasPrivacy := false
	fee := 1
	err = tx1.Init(
		NewTxPrivacyInitParams(
			&senderKey.KeySet.PrivateKey,
			[]*privacy.PaymentInfo{{PaymentAddress: receiverPaymentAddress.KeySet.PaymentAddress, Amount: uint64(transferAmount)}},
			coinBaseOutput, uint64(fee), hasPrivacy, db, nil, nil,
		),
	)
	if err != nil {
		t.Error(err)
	}

	senderPubKeyLastByte := tx1.GetSenderAddrLastByte()
	assert.Equal(t, senderKey.KeySet.PaymentAddress.Pk[len(senderKey.KeySet.PaymentAddress.Pk)-1], senderPubKeyLastByte)

	actualFee := tx1.GetTxFee()
	assert.Equal(t, uint64(fee), actualFee)

	actualFeeToken := tx1.GetTxFeeToken()
	assert.Equal(t, uint64(0), actualFeeToken)

	unique, pubk, amount := tx1.GetUniqueReceiver()
	assert.Equal(t, true, unique)
	assert.Equal(t, string(pubk[:]), string(receiverPaymentAddress.KeySet.PaymentAddress.Pk[:]))
	assert.Equal(t, uint64(5), amount)

	unique, pubk, amount, coinID := tx1.GetTransferData()
	assert.Equal(t, true, unique)
	assert.Equal(t, common.PRVCoinID.String(), coinID.String())
	assert.Equal(t, string(pubk[:]), string(receiverPaymentAddress.KeySet.PaymentAddress.Pk[:]))

	a, b := tx1.GetTokenReceivers()
	assert.Equal(t, 0, len(a))
	assert.Equal(t, 0, len(b))

	e, d, c := tx1.GetTokenUniqueReceiver()
	assert.Equal(t, false, e)
	assert.Equal(t, 0, len(d))
	assert.Equal(t, uint64(0), c)

	listInputSerialNumber := tx1.ListSerialNumbersHashH()
	assert.Equal(t, 1, len(listInputSerialNumber))
	assert.Equal(t, common.HashH(coinBaseOutput[0].CoinDetails.GetSerialNumber().Compress()), listInputSerialNumber[0])

	isValidSanity, err = tx1.ValidateSanityData(nil)
	assert.Equal(t, true, isValidSanity)
	assert.Equal(t, nil, err)

	isValid, err := tx1.ValidateTransaction(hasPrivacy, db, shardID, nil)
	assert.Equal(t, true, isValid)
	assert.Equal(t, nil, err)

	isValidTxVersion := tx1.CheckTxVersion(1)
	assert.Equal(t, true, isValidTxVersion)

	isValidTxFee := tx1.CheckTransactionFee(0)
	assert.Equal(t, true, isValidTxFee)

	isSalaryTx := tx1.IsSalaryTx()
	assert.Equal(t, false, isSalaryTx)

	actualSenderPublicKey := tx1.GetSender()
	expectedSenderPublicKey := make([]byte, common.PublicKeySize)
	copy(expectedSenderPublicKey, senderPublicKey[:])
	assert.Equal(t, expectedSenderPublicKey, actualSenderPublicKey[:])

	//qual(t, nil, err)err = tx1.ValidateTxWithCurrentMempool(nil)
	//	//assert.E

	err = tx1.ValidateDoubleSpendWithBlockchain(nil, shardID, db, nil)
	assert.Equal(t, nil, err)

	err = tx1.ValidateTxWithBlockChain(nil, shardID, db)
	assert.Equal(t, nil, err)

	isValid, err = tx1.ValidateTxByItself(hasPrivacy, db, nil, shardID)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, isValid)

	metaDataType := tx1.GetMetadataType()
	assert.Equal(t, metadata.InvalidMeta, metaDataType)

	metaData := tx1.GetMetadata()
	assert.Equal(t, nil, metaData)

	info := tx1.GetInfo()
	assert.Equal(t, 0, len(info))

	lockTime := tx1.GetLockTime()
	now := time.Now().Unix()
	assert.LessOrEqual(t, lockTime, now)

	actualSigPubKey := tx1.GetSigPubKey()
	assert.Equal(t, expectedSenderPublicKey, actualSigPubKey)

	proof := tx1.GetProof()
	assert.NotEqual(t, nil, proof)

	isValidTxType := tx1.ValidateType()
	assert.Equal(t, true, isValidTxType)

	isCoinsBurningTx := tx1.IsCoinsBurning()
	assert.Equal(t, false, isCoinsBurningTx)

	actualTxValue := tx1.CalculateTxValue()
	assert.Equal(t, uint64(transferAmount), actualTxValue)

	// store output coin's coin commitments in tx1
	//for i:=0; i < len(tx1.Proof.GetOutputCoins()); i++ {
	//	db.StoreCommitments(
	//		common.PRVCoinID,
	//		tx1.Proof.GetOutputCoins()[i].CoinDetails.GetPublicKey().Compress(),
	//		[][]byte{tx1.Proof.GetOutputCoins()[i].CoinDetails.GetCoinCommitment().Compress()},
	//		shardID)
	//}

	// init tx with privacy
	tx2 := Tx{}

	// prepare input coins
	//outputCoins, err := db.GetOutcoinsByPubkey(common.PRVCoinID, senderPaymentAddress.Pk, shardID)

	//fmt.Printf("outputCoins len: %v\n", len(outputCoins))
	//fmt.Printf("outputCoins: %v\n", outputCoins)
	err = tx2.Init(
		NewTxPrivacyInitParams(
			&senderKey.KeySet.PrivateKey,
			[]*privacy.PaymentInfo{{PaymentAddress: senderPaymentAddress, Amount: uint64(transferAmount)}},
			coinBaseOutput, 1, true, db, nil, nil))
	if err != nil {
		t.Error(err)
	}

	isValidSanity, err = tx2.ValidateSanityData(nil)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, isValidSanity)

	isValidTx, err := tx2.ValidateTransaction(true, db, shardID, &common.PRVCoinID)
	assert.Equal(t, true, isValidTx)
}

func TestInitSalaryTx(t *testing.T) {
	salary := uint64(1000)

	senderKey, _ := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	senderKey.KeySet.InitFromPrivateKey(&senderKey.KeySet.PrivateKey)
	senderPaymentAddress := senderKey.KeySet.PaymentAddress
	receiverAddr := senderPaymentAddress

	tx := new(Tx)
	err := tx.InitTxSalary(salary, &receiverAddr, &senderKey.KeySet.PrivateKey, db, nil)
	assert.Equal(t, nil, err)

	isValid, err := tx.ValidateTxSalary(db)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, isValid)

	isSalaryTx := tx.IsSalaryTx()
	assert.Equal(t, true, isSalaryTx)
}
