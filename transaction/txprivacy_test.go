package transaction

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUnmarshalJSON(t *testing.T) {
	key, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = key.KeySet.InitFromPrivateKey(&key.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	paymentAddress := key.KeySet.PaymentAddress
	responseMeta, err := metadata.NewWithDrawRewardResponse(&metadata.WithDrawRewardRequest{}, &common.Hash{})
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
	for i := 0; i < 1; i++ {
		//Generate sender private key & receiver payment address
		seed := privacy.RandomScalar().ToBytesS()
		masterKey, _ := wallet.NewMasterKey(seed)
		childSender, _ := masterKey.NewChildKey(uint32(1))
		privKeyB58 := childSender.Base58CheckSerialize(wallet.PriKeyType)
		childReceiver, _ := masterKey.NewChildKey(uint32(2))
		paymentAddressB58 := childReceiver.Base58CheckSerialize(wallet.PaymentAddressType)

		senderKey, err := wallet.Base58CheckDeserialize(privKeyB58)
		assert.Equal(t, nil, err)

		err = senderKey.KeySet.InitFromPrivateKey(&senderKey.KeySet.PrivateKey)
		assert.Equal(t, nil, err)

		senderPaymentAddress := senderKey.KeySet.PaymentAddress
		senderPublicKey := senderPaymentAddress.Pk

		shardID := common.GetShardIDFromLastByte(senderKey.KeySet.PaymentAddress.Pk[len(senderKey.KeySet.PaymentAddress.Pk)-1])

		// coin base tx to mint PRV
		mintedAmount := 1000
		coinBaseTx, err := BuildCoinBaseTxByCoinID(NewBuildCoinBaseTxByCoinIDParams(&senderPaymentAddress, uint64(mintedAmount), &senderKey.KeySet.PrivateKey, db, nil, common.Hash{}, NormalCoinType, "PRV", 0))

		isValidSanity, err := coinBaseTx.ValidateSanityData(nil, nil, nil)
		assert.Equal(t, nil, err)
		assert.Equal(t, true, isValidSanity)

		// store output coin's coin commitments in coin base tx
		db.StoreCommitments(
			common.PRVCoinID,
			senderPaymentAddress.Pk,
			[][]byte{coinBaseTx.(*Tx).Proof.GetOutputCoins()[0].CoinDetails.GetCoinCommitment().ToBytesS()},
			shardID)

		// get output coins from coin base tx to create new tx
		coinBaseOutput := ConvertOutputCoinToInputCoin(coinBaseTx.(*Tx).Proof.GetOutputCoins())

		// init new tx without privacy
		tx1 := Tx{}
		// calculate serial number for input coins
		serialNumber := new(privacy.Point).Derive(privacy.PedCom.G[privacy.PedersenPrivateKeyIndex],
			new(privacy.Scalar).FromBytesS(senderKey.KeySet.PrivateKey),
			coinBaseOutput[0].CoinDetails.GetSNDerivator())

		coinBaseOutput[0].CoinDetails.SetSerialNumber(serialNumber)

		receiverPaymentAddress, _ := wallet.Base58CheckDeserialize(paymentAddressB58)

		// transfer amount
		transferAmount := 5
		hasPrivacy := false
		fee := 1

		// message to receiver
		msg := "Incognito-chain"
		receiverTK, _ := new(privacy.Point).FromBytesS(senderKey.KeySet.PaymentAddress.Tk)
		msgCipherText, _ := privacy.HybridEncrypt([]byte(msg), receiverTK)

		fmt.Printf("msgCipherText: %v - len : %v\n", msgCipherText.Bytes(), len(msgCipherText.Bytes()))
		err = tx1.Init(
			NewTxPrivacyInitParams(
				&senderKey.KeySet.PrivateKey,
				[]*privacy.PaymentInfo{{PaymentAddress: receiverPaymentAddress.KeySet.PaymentAddress, Amount: uint64(transferAmount), Message: msgCipherText.Bytes()}},
				coinBaseOutput, uint64(fee), hasPrivacy, db, nil, nil, []byte{},
			),
		)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, len(msgCipherText.Bytes()), len(tx1.Proof.GetOutputCoins()[0].CoinDetails.GetInfo()))

		actualSize := tx1.GetTxActualSize()
		fmt.Printf("actualSize: %v\n", actualSize)

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
		assert.Equal(t, common.HashH(coinBaseOutput[0].CoinDetails.GetSerialNumber().ToBytesS()), listInputSerialNumber[0])

		isValidSanity, err = tx1.ValidateSanityData(nil, nil, nil)
		assert.Equal(t, true, isValidSanity)
		assert.Equal(t, nil, err)

		isValid, err := tx1.ValidateTransaction(hasPrivacy, db, shardID, nil, false)

		fmt.Printf("Error: %v\n", err)
		assert.Equal(t, true, isValid)
		assert.Equal(t, nil, err)

		isValidTxVersion := tx1.CheckTxVersion(1)
		assert.Equal(t, true, isValidTxVersion)

		//isValidTxFee := tx1.CheckTransactionFee(0)
		//assert.Equal(t, true, isValidTxFee)

		isSalaryTx := tx1.IsSalaryTx()
		assert.Equal(t, false, isSalaryTx)

		actualSenderPublicKey := tx1.GetSender()
		expectedSenderPublicKey := make([]byte, common.PublicKeySize)
		copy(expectedSenderPublicKey, senderPublicKey[:])
		assert.Equal(t, expectedSenderPublicKey, actualSenderPublicKey[:])

		//err = tx1.ValidateTxWithCurrentMempool(nil)
		//	assert.Equal(t, nil, err)

		err = tx1.ValidateDoubleSpendWithBlockchain(nil, shardID, db, nil)
		assert.Equal(t, nil, err)

		err = tx1.ValidateTxWithBlockChain(nil, shardID, nil, nil, db)
		assert.Equal(t, nil, err)

		isValid, err = tx1.ValidateTxByItself(hasPrivacy, db, nil, shardID, nil, nil)
		assert.Equal(t, nil, err)
		assert.Equal(t, true, isValid)

		metaDataType := tx1.GetMetadataType()
		assert.Equal(t, basemeta.InvalidMeta, metaDataType)

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

		//TODO: Fix IsCoinsBurning
		//isCoinsBurningTx := tx1.IsCoinsBurning()
		//assert.Equal(t, false, isCoinsBurningTx)

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
		//tx2 := Tx{}
		//
		//err = tx2.Init(
		//	NewTxPrivacyInitParams(
		//		&senderKey.KeySet.PrivateKey,
		//		[]*privacy.PaymentInfo{{PaymentAddress: senderPaymentAddress, Amount: uint64(transferAmount)}},
		//		coinBaseOutput, 1, true, db, nil, nil, []byte{}))
		//if err != nil {
		//	t.Error(err)
		//}
		//
		//isValidSanity, err = tx2.ValidateSanityData(nil)
		//assert.Equal(t, nil, err)
		//assert.Equal(t, true, isValidSanity)
		//
		//isValidTx, err := tx2.ValidateTransaction(true, db, shardID, &common.PRVCoinID)
		//assert.Equal(t, true, isValidTx)

	}
}

func TestInitTxWithMultiScenario(t *testing.T) {
	for i := 0; i < 50; i++ {
		//Generate sender private key & receiver payment address
		seed := privacy.RandomScalar().ToBytesS()
		masterKey, _ := wallet.NewMasterKey(seed)
		childSender, _ := masterKey.NewChildKey(uint32(1))
		privKeyB58 := childSender.Base58CheckSerialize(wallet.PriKeyType)
		childReceiver, _ := masterKey.NewChildKey(uint32(2))
		paymentAddressB58 := childReceiver.Base58CheckSerialize(wallet.PaymentAddressType)

		// sender key
		senderKey, err := wallet.Base58CheckDeserialize(privKeyB58)
		assert.Equal(t, nil, err)

		err = senderKey.KeySet.InitFromPrivateKey(&senderKey.KeySet.PrivateKey)
		assert.Equal(t, nil, err)

		senderPaymentAddress := senderKey.KeySet.PaymentAddress

		//receiver key
		receiverKey, _ := wallet.Base58CheckDeserialize(paymentAddressB58)
		receiverPaymentAddress := receiverKey.KeySet.PaymentAddress

		// shard ID of sender
		shardID := common.GetShardIDFromLastByte(senderKey.KeySet.PaymentAddress.Pk[len(senderKey.KeySet.PaymentAddress.Pk)-1])

		// create coin base tx to mint PRV
		mintedAmount := 1000
		coinBaseTx, err := BuildCoinBaseTxByCoinID(NewBuildCoinBaseTxByCoinIDParams(&senderPaymentAddress, uint64(mintedAmount), &senderKey.KeySet.PrivateKey, db, nil, common.Hash{}, NormalCoinType, "PRV", 0))

		isValidSanity, err := coinBaseTx.ValidateSanityData(nil, nil, nil)
		assert.Equal(t, nil, err)
		assert.Equal(t, true, isValidSanity)

		// store output coin's coin commitments in coin base tx
		db.StoreCommitments(
			common.PRVCoinID,
			senderPaymentAddress.Pk,
			[][]byte{coinBaseTx.(*Tx).Proof.GetOutputCoins()[0].CoinDetails.GetCoinCommitment().ToBytesS()},
			shardID)

		// get output coins from coin base tx to create new tx
		coinBaseOutput := ConvertOutputCoinToInputCoin(coinBaseTx.(*Tx).Proof.GetOutputCoins())

		// init new tx with privacy
		tx1 := Tx{}
		// calculate serial number for input coins
		serialNumber := new(privacy.Point).Derive(privacy.PedCom.G[privacy.PedersenPrivateKeyIndex],
			new(privacy.Scalar).FromBytesS(senderKey.KeySet.PrivateKey),
			coinBaseOutput[0].CoinDetails.GetSNDerivator())

		coinBaseOutput[0].CoinDetails.SetSerialNumber(serialNumber)

		// transfer amount
		transferAmount := 5
		hasPrivacy := true
		fee := 1
		err = tx1.Init(
			NewTxPrivacyInitParams(
				&senderKey.KeySet.PrivateKey,
				[]*privacy.PaymentInfo{{PaymentAddress: receiverPaymentAddress, Amount: uint64(transferAmount)}},
				coinBaseOutput, uint64(fee), hasPrivacy, db, nil, nil, []byte{},
			),
		)
		assert.Equal(t, nil, err)

		isValidSanity, err = tx1.ValidateSanityData(nil, nil, nil)
		assert.Equal(t, true, isValidSanity)
		assert.Equal(t, nil, err)
		fmt.Println("Hello")
		isValid, err := tx1.ValidateTransaction(hasPrivacy, db, shardID, nil, false)
		assert.Equal(t, true, isValid)
		assert.Equal(t, nil, err)
		fmt.Println("Hello")
		err = tx1.ValidateDoubleSpendWithBlockchain(nil, shardID, db, nil)
		assert.Equal(t, nil, err)

		err = tx1.ValidateTxWithBlockChain(nil, shardID, nil, nil, db)
		assert.Equal(t, nil, err)

		isValid, err = tx1.ValidateTxByItself(hasPrivacy, db, nil, shardID, nil, nil)
		assert.Equal(t, nil, err)
		assert.Equal(t, true, isValid)

		// modify Sig
		tx1.Sig[len(tx1.Sig)-1] = tx1.Sig[len(tx1.Sig)-1] ^ tx1.Sig[0]
		tx1.Sig[len(tx1.Sig)-2] = tx1.Sig[len(tx1.Sig)-2] ^ tx1.Sig[1]
		isValid, err = tx1.ValidateTransaction(hasPrivacy, db, shardID, nil, false)
		assert.Equal(t, false, isValid)
		assert.NotEqual(t, nil, err)
		tx1.Sig[len(tx1.Sig)-1] = tx1.Sig[len(tx1.Sig)-1] ^ tx1.Sig[0]
		tx1.Sig[len(tx1.Sig)-2] = tx1.Sig[len(tx1.Sig)-2] ^ tx1.Sig[1]

		// modify verification key
		tx1.SigPubKey[len(tx1.SigPubKey)-1] = tx1.SigPubKey[len(tx1.SigPubKey)-1] ^ tx1.SigPubKey[0]
		tx1.SigPubKey[len(tx1.SigPubKey)-2] = tx1.SigPubKey[len(tx1.SigPubKey)-2] ^ tx1.SigPubKey[1]

		isValid, err = tx1.ValidateTransaction(hasPrivacy, db, shardID, nil, false)
		assert.Equal(t, false, isValid)
		assert.NotEqual(t, nil, err)

		tx1.SigPubKey[len(tx1.SigPubKey)-1] = tx1.SigPubKey[len(tx1.SigPubKey)-1] ^ tx1.SigPubKey[0]
		tx1.SigPubKey[len(tx1.SigPubKey)-2] = tx1.SigPubKey[len(tx1.SigPubKey)-2] ^ tx1.SigPubKey[1]

		// modify proof
		originProof := tx1.Proof.Bytes()

		//var modifiedProof = make ([]byte, len(originProof))
		//copy(modifiedProof, originProof)
		//modifiedProof[7] = modifiedProof[8]
		//modifiedProof[5] = modifiedProof[6]
		//modifiedProof[6] = modifiedProof[16]
		//
		//tx1.Proof.SetBytes(modifiedProof)
		//
		//isValid, err = tx1.ValidateTransaction(hasPrivacy, db, shardID, nil)
		//assert.Equal(t, false, isValid)
		//assert.NotEqual(t, nil, err)

		tx1.Proof.SetBytes(originProof)

		// back to correct case
		isValid, err = tx1.ValidateTxByItself(hasPrivacy, db, nil, shardID, nil, nil)
		assert.Equal(t, nil, err)
		assert.Equal(t, true, isValid)
	}
}

func TestInitSalaryTx(t *testing.T) {
	salary := uint64(1000)

	privateKey := privacy.GeneratePrivateKey([]byte{123})
	senderKey := new(wallet.KeyWallet)
	err := senderKey.KeySet.InitFromPrivateKey(&privateKey)
	assert.Equal(t, nil, err)

	senderPaymentAddress := senderKey.KeySet.PaymentAddress
	receiverAddr := senderPaymentAddress

	tx := new(Tx)
	err = tx.InitTxSalary(salary, &receiverAddr, &senderKey.KeySet.PrivateKey, db, nil)
	assert.Equal(t, nil, err)

	isValid, err := tx.ValidateTxSalary(db)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, isValid)

	isSalaryTx := tx.IsSalaryTx()
	assert.Equal(t, true, isSalaryTx)
}

type CoinObject struct {
	PublicKey      string
	CoinCommitment string
	SNDerivator    string
	SerialNumber   string
	Randomness     string
	Value          uint64
	Info           string
}

func ParseCoinObjectToStruct(coinObjects []CoinObject) ([]*privacy.InputCoin, uint64) {
	coins := make([]*privacy.InputCoin, len(coinObjects))
	sumValue := uint64(0)

	for i := 0; i < len(coins); i++ {

		publicKey, _, _ := base58.Base58Check{}.Decode(coinObjects[i].PublicKey)
		publicKeyPoint := new(privacy.Point)
		publicKeyPoint.FromBytesS(publicKey)

		coinCommitment, _, _ := base58.Base58Check{}.Decode(coinObjects[i].CoinCommitment)
		coinCommitmentPoint := new(privacy.Point)
		coinCommitmentPoint.FromBytesS(coinCommitment)

		snd, _, _ := base58.Base58Check{}.Decode(coinObjects[i].SNDerivator)
		sndBN := new(privacy.Scalar).FromBytesS(snd)

		serialNumber, _, _ := base58.Base58Check{}.Decode(coinObjects[i].SerialNumber)
		serialNumberPoint := new(privacy.Point)
		serialNumberPoint.FromBytesS(serialNumber)

		randomness, _, _ := base58.Base58Check{}.Decode(coinObjects[i].Randomness)
		randomnessBN := new(privacy.Scalar).FromBytesS(randomness)

		coins[i] = new(privacy.InputCoin).Init()
		coins[i].CoinDetails.SetPublicKey(publicKeyPoint)
		coins[i].CoinDetails.SetCoinCommitment(coinCommitmentPoint)
		coins[i].CoinDetails.SetSNDerivator(sndBN)
		coins[i].CoinDetails.SetSerialNumber(serialNumberPoint)
		coins[i].CoinDetails.SetRandomness(randomnessBN)
		coins[i].CoinDetails.SetValue(coinObjects[i].Value)

		sumValue += coinObjects[i].Value

	}

	return coins, sumValue
}
