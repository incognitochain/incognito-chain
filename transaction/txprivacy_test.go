package transaction

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"math/big"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
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
			coinBaseOutput, uint64(fee), hasPrivacy, db, nil, nil, []byte{},
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
			coinBaseOutput, 1, true, db, nil, nil, []byte{}))
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

type CoinObject struct {
	PublicKey string
	CoinCommitment string
	SNDerivator string
	SerialNumber string
	Randomness string
	Value uint64
	Info string
}

func ParseCoinObjectToStruct(coinObjects []CoinObject) ([]*privacy.InputCoin, uint64) {
	coins := make([]*privacy.InputCoin, len(coinObjects))
	sumValue := uint64(0)

	for i := 0; i<len(coins); i++{

		publicKey, _, _ := base58.Base58Check{}.Decode(coinObjects[i].PublicKey)
		publicKeyPoint := new(privacy.EllipticPoint)
		publicKeyPoint.Decompress(publicKey)

		coinCommitment, _, _ := base58.Base58Check{}.Decode(coinObjects[i].CoinCommitment)
		coinCommitmentPoint := new(privacy.EllipticPoint)
		coinCommitmentPoint.Decompress(coinCommitment)

		snd, _, _ := base58.Base58Check{}.Decode(coinObjects[i].SNDerivator)
		sndBN := new(big.Int).SetBytes(snd)

		serialNumber, _, _ := base58.Base58Check{}.Decode(coinObjects[i].SerialNumber)
		serialNumberPoint := new(privacy.EllipticPoint)
		serialNumberPoint.Decompress(serialNumber)

		randomness, _, _ := base58.Base58Check{}.Decode(coinObjects[i].Randomness)
		randomnessBN := new(big.Int).SetBytes(randomness)

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

func TestInitTx2(t *testing.T) {
	//witness := new(PaymentWitness)
	//witnessParam := new(PaymentWitnessParam)

	keyWallet, _ := wallet.Base58CheckDeserialize("112t8rnXHD9s2MXSXigMyMtKdGFtSJmhA9cCBN34Fj55ox3cJVL6Fykv8uNWkDagL56RnA4XybQKNRrNXinrDDfKZmq9Y4LR18NscSrc9inc")
	_ = keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	_ = new(big.Int).SetBytes(keyWallet.KeySet.PrivateKey)
	senderPKPoint := new(privacy.EllipticPoint)
	senderPKPoint.Decompress(keyWallet.KeySet.PaymentAddress.Pk)
	shardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk) -1])


	coinStrs := []CoinObject{
		{
			PublicKey: "183XvUp5gn7gtTWjMBGwpBSgER6zexEMAqmvvQsd9ZavsErG89y",
			CoinCommitment: "16rDxiXDg9AhyC3o3XiBQZAtg4P2x1ER9umyspRFC4AUWGj9LnK",
			SNDerivator: "12bf2zoKdYw8c8BT3YMKNaVkLppoQqEkLtSCymEa6EK65FSowV7",
			SerialNumber: "17ioQJTBFV8HGK6TYQn9mWfdT8Z7wRCMyn9GjFYhMx6dP8UrnJp",
			Randomness: "13CyLqj6BErihknHV7AWqHdAodLAwRwGuqkEdDqFb5chS5uhLN",
			Value: 13063917525,
			Info: "13PMpZ4",
		},
		{
			PublicKey: "183XvUp5gn7gtTWjMBGwpBSgER6zexEMAqmvvQsd9ZavsErG89y",
			CoinCommitment: "17pb83j2YcrB8WLr1jPNGsT6Qgo3dEan7U6NsJwR2QAY1PcmXWa",
			SNDerivator: "12M48gjxpPUkb69ieMLc9EhBDcCerTbhtHnAgdoaEToXUYhFiCb",
			SerialNumber: "18fJDPSbjLnTCxk2QUrEig4Ai5kWbPYediD1KhKinKm142smQVs",
			Randomness: "12cyHe5MyGLDGeKDZSknP2DEny48mNMC49Rd8CHhdiBCh35bnTs",
			Value: 4230769230,
			Info: "13PMpZ4",
		},
		{
			PublicKey: "183XvUp5gn7gtTWjMBGwpBSgER6zexEMAqmvvQsd9ZavsErG89y",
			CoinCommitment: "17tcbagBHAjG8fr2RGLc3FjAJ5Mkbqitdv6KCtQWLiBydHoHpRP",
			SNDerivator: "1SXpgdZKqwENjSYgLhaam6PS3u5CciYMHuwyt1ipr5SUQQMYGn",
			SerialNumber: "15Mnm5Do1Np3eoPdGRvECJb8mjhHLgvDYoWxNQgAxXTLCUh2MYa",
			Randomness: "12Br463SeHFafpPEntE1L81S6vk5HShUtgE7tiCfPzr1aWiSZMU",
			Value: 13395348837,
			Info: "13PMpZ4",
		},
		{
			PublicKey: "183XvUp5gn7gtTWjMBGwpBSgER6zexEMAqmvvQsd9ZavsErG89y",
			CoinCommitment: "16ep85MLtTigBiwPf1b6bRcKJ9NJfVazxjC1GCzEsqKK1J5927t",
			SNDerivator: "1EXomopZG5uUDbC9fRWyg4UnvboqY3PQnmjF1srRyRUDFXaUzd",
			SerialNumber: "16PpZUXgsQntxB8Js6yoPRzyZEiQyQTGtSXUEYnvV1uEbv7wPdw",
			Randomness: "12uWji6kLpUo8Xg5AJx1odDkeP2ZQ7g9p2tnSMwLJfvgDDmkWVS",
			Value: 466999090249,
			Info: "13PMpZ4",
		},
		{
			PublicKey: "183XvUp5gn7gtTWjMBGwpBSgER6zexEMAqmvvQsd9ZavsErG89y",
			CoinCommitment: "17Pw2SmoW4zXojM8HHHMEpX5k3SjKL8UAeGXBDKjqpJBJKtHSkf",
			SNDerivator: "122qVAS24X5AjWdWsiX54npCN7WDrAyDk4VmGSbFNexWcofzNXa",
			SerialNumber: "18LrSQofiFy9HuiCbdPJZp7nFKg9z6xNiN1EoeRVWdCiMf6Yyrm",
			Randomness: "12XdvDLJ2UKASYX2wCSEKvda3xYrJKeUaP4XXmQ3f6f5hA399pg",
			Value: 13423728813,
			Info: "13PMpZ4",
		},
		{
			PublicKey: "183XvUp5gn7gtTWjMBGwpBSgER6zexEMAqmvvQsd9ZavsErG89y",
			CoinCommitment: "188C79Y2jJmKNxxuGN56S5rSXDYAZqP7erMEebmui74DaS7qf4V",
			SNDerivator: "1xD4hPppKFkwTkUK2GkR6VVEszhF94KZEFqpqvSynqUePGKnrh",
			SerialNumber: "16TKsDv351rbn64bw4CTnwfSd626oJ6bYRYjUQqYP2dmyRqYpXn",
			Randomness: "1WPLdUVWt6566hpjENoNkSmukPSyYBjbWwrv2nyQx49DByPR36",
			Value: 6285714285,
			Info: "13PMpZ4",
		},
		{
			PublicKey: "183XvUp5gn7gtTWjMBGwpBSgER6zexEMAqmvvQsd9ZavsErG89y",
			CoinCommitment: "15pYpezPQtG4hfJXXNTKcUw1jirBxBdfA6VujmHjjF4kN275KBQ",
			SNDerivator: "14Bx5hYq5JEdXyjUXJELJgQzL2PpvGv3fgJZBKMPNTwwjEh1T8",
			SerialNumber: "18hRCi5DiC3i3RYi7WAdLTYJyhyidPi7dfEJP1Xe9PGt3PfY4FU",
			Randomness: "1DjLCFnnfnuPmNU1ZXxJ5wjums7CxVE9uvFGcPBHcX6PBwqoVH",
			Value: 6226415094,
			Info: "13PMpZ4",
		},
	}

	inputCoins, _ := ParseCoinObjectToStruct(coinStrs)
	fmt.Printf("Parse input done!!!!\n")

	keyWalletReceiver, _ := wallet.Base58CheckDeserialize("112t8rnXHD9s2MXSXigMyMtKdGFtSJmhA9cCBN34Fj55ox3cJVL6Fykv8uNWkDagL56RnA4XybQKNRrNXinrDDfKZmq9Y4LR18NscSrc9inc")
	_ = keyWalletReceiver.KeySet.InitFromPrivateKey(&keyWalletReceiver.KeySet.PrivateKey)
	//receiverKeyBN := new(big.Int).SetBytes(keyWalletReceiver.KeySet.PrivateKey)
	receiverPublicKey := keyWalletReceiver.KeySet.PaymentAddress.Pk
	receiverPublicKeyPoint := new(privacy.EllipticPoint)
	receiverPublicKeyPoint.Decompress(receiverPublicKey)

	amountTransfer := uint64(1000000000)

	//outputCoins := make([]*privacy.OutputCoin, 2)
	//outputCoins[0].Init()
	//outputCoins[0].CoinDetails.SetValue(uint64(amountTransfer))
	//outputCoins[0].CoinDetails.SetPublicKey(receiverPublicKeyPoint)
	//outputCoins[0].CoinDetails.SetSNDerivator(privacy.RandScalar(rand.Reader))
	//
	//changeAmount :=sumValue - amountTransfer
	//
	//outputCoins[1].Init()
	//outputCoins[1].CoinDetails.SetValue(changeAmount)
	//outputCoins[1].CoinDetails.SetPublicKey(senderPKPoint)
	//outputCoins[1].CoinDetails.SetSNDerivator(privacy.RandScalar(rand.Reader))

	// using default PRV
	tokenID := &common.Hash{}
	_ = tokenID.SetBytes(common.PRVCoinID[:])

	// store output coin's coin commitments in coin base tx
	for i:= 0; i<len(inputCoins); i++{
		db.StoreCommitments(
			common.PRVCoinID,
			keyWallet.KeySet.PaymentAddress.Pk,
			[][]byte{inputCoins[i].CoinDetails.GetCoinCommitment().Compress()},
			shardID)
	}

	for i:=0; i<len(inputCoins); i++{
		commitmentBytes, _ := db.GetCommitmentByIndex(*tokenID, uint64(i), shardID)
		fmt.Printf("index %v : commitmentBytes %v\n", i, commitmentBytes)
		fmt.Printf("index %v : serial number %v\n", i, inputCoins[i].CoinDetails.GetSerialNumber().Compress())
	}



	commitment, _ := db.ListCommitmentIndices(*tokenID, shardID)
	for i, cm := range commitment{
		fmt.Printf("cm %v: %v\n", i, cm)
	}

	fmt.Printf("Set db done!!!!\n")

	// init new tx without privacy
	tx1 := Tx{}
	hasPrivacy := true
	fee := 0
	err := tx1.Init(
		NewTxPrivacyInitParams(
			&keyWallet.KeySet.PrivateKey,
			[]*privacy.PaymentInfo{{PaymentAddress: keyWalletReceiver.KeySet.PaymentAddress, Amount: uint64(amountTransfer)}},
			inputCoins[:len(inputCoins) - 1], uint64(fee), hasPrivacy, db, nil, nil, []byte{},
		),
	)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("Init tx done!!!!")

	isValidSanity, err := tx1.ValidateSanityData(nil)
	assert.Equal(t, true, isValidSanity)
	assert.Equal(t, nil, err)

	proof := tx1.Proof
	fmt.Printf("Proof: %v\n", proof)
	proofBytes := proof.Bytes()
	fmt.Printf("proofBytes: %v\n", proofBytes)
	fmt.Printf("\n\n")

	for i := 0; i< len(proof.GetOneOfManyProof()); i++{
		//p := proof.GetOneOfManyProof()[i]
		for j :=0; j<8; j++{
			fmt.Printf("proof.GetOneOfManyProof()[%v].Statement.Commitments[%v].Compress(): %v\n",i, j, proof.GetOneOfManyProof()[i].Statement.Commitments[j].Compress())
		}

	}

	proof2 := new(zkp.PaymentProof)
	proof2.SetBytes(proofBytes)
	fmt.Printf("Proof: %v\n", proof2)

	//p1 := proof.GetOneOfManyProof()
	//p2 := proof2.GetOneOfManyProof()

	//for i := 0; i< len(p1); i++{
	//	for j := 0; j < 8; j++{
	//		if !p1[i].Statement.Commitments[j].IsEqual(p2[i].Statement.Commitments[j]){
	//			fmt.Printf("p1[%v].Statement.Commitments[%v]: %v\n ", i, j, p1[i].Statement.Commitments[i])
	//			fmt.Printf("p2[%v].Statement.Commitments[%v]: %v\n ", i, j, p2[i].Statement.Commitments[i])
	//		}
	//	}
	//
	//
	//	//if !p1[i].[i].IsEqual(p1[i].Statement.Commitments[i]){
	//	//	fmt.Printf("p1[i].Statement.Commitments[%v]: %v\n ", i, p1[i].Statement.Commitments[i])
	//	//}
	//}

	proofBytes2 := proof2.Bytes()
	fmt.Printf("proofBytes2: %v\n", proofBytes2)



	res, err2 := proof.Verify(hasPrivacy, keyWallet.KeySet.PaymentAddress.Pk, uint64(fee), db, shardID, tokenID)
	assert.Equal(t, true, res)
	assert.Equal(t, nil, err)
	fmt.Printf("err2 %v\n", err2)

	//isValid, err := tx1.ValidateTransaction(hasPrivacy, db, shardID, tokenID)
	//assert.Equal(t, true, isValid)
	//assert.Equal(t, nil, err)
	////fmt.Printf("err %v\n", err)
	//
	//
	//isValid, err = tx1.ValidateTxByItself(hasPrivacy, db, nil, shardID)
	//assert.Equal(t, nil, err)
	//assert.Equal(t, true, isValid)



}

