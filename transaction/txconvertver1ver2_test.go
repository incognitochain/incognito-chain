package transaction

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/trie"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

var (
	numInputs          = 5
	numOutputs         = 5
	numTests           = 10
	unitFeeNativeToken = 100
)

var _ = func() (_ struct{}) {
	// initialize a `test` db in the OS's tempdir
	// and with it, a db access wrapper that reads/writes our transactions
	fmt.Println("This runs before init()!")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func createSamplePlainCoinV1(privKey privacy.PrivateKey, pubKey *operation.Point, amount uint64, msg []byte) (*coin.PlainCoinV1, error) {
	c := new(coin.PlainCoinV1).Init()

	c.SetValue(amount)
	c.SetInfo(msg)
	c.SetPublicKey(pubKey)
	c.SetSNDerivator(operation.RandomScalar())
	c.SetRandomness(operation.RandomScalar())

	//Derive serial number from snDerivator
	c.SetKeyImage(new(operation.Point).Derive(privacy.PedCom.G[0], new(operation.Scalar).FromBytesS(privKey), c.GetSNDerivator()))

	//Create commitment
	err := c.CommitAll()

	if err != nil {
		return nil, err
	}

	return c, nil
}

func createConversionParams(numInputs, numOutputs int, tokenID *common.Hash) (*incognitokey.KeySet,
	[]*privacy.PaymentInfo, *TxConvertVer1ToVer2InitParams, error) {
	var senderSK privacy.PrivateKey
	var keySet *incognitokey.KeySet

	//create a sample test DB
	testDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)

	//generate keyset: we want the public key to be in Shard 0
	for {
		//generate a private key
		senderSK = key.GeneratePrivateKey(common.RandBytes(32))

		//make keySets from privateKey
		keySet = new(incognitokey.KeySet)
		err := keySet.InitFromPrivateKey(&senderSK)

		if err != nil {
			return nil, nil, nil, err
		}

		//we want the public key to belong to Shard 0
		if keySet.PaymentAddress.Pk[31] == 0 {
			break
		}
	}

	//create input coins
	inputCoins := make([]coin.PlainCoin, numInputs)
	sumInput := uint64(0)
	var err error

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	if err != nil {
		return nil, nil, nil, err
	}

	for i := 0; i < numInputs; i++ {
		amount := uint64(common.RandIntInterval(0, 1000))
		inputCoins[i], err = createSamplePlainCoinV1(senderSK, pubKey, amount, nil)
		if err != nil {
			return nil, nil, nil, err
		}
		sumInput += amount
	}

	sumOutput := uint64(0)

	paymentInfo := make([]*key.PaymentInfo, numOutputs)

	for i := 0; i < numOutputs; i++ {
		amount := uint64(common.RandIntInterval(0, int(sumInput-sumOutput)))
		paymentInfo[i] = key.InitPaymentInfo(keySet.PaymentAddress, amount, nil)
		sumOutput += amount
	}

	//calculate sample fee
	fee := sumInput - sumOutput

	//create conversion params
	txConversionParams := NewTxConvertVer1ToVer2InitParams(
		&keySet.PrivateKey,
		paymentInfo,
		inputCoins,
		fee,
		testDB,
		tokenID, // use for prv coin -> nil is valid
		nil,
		nil,
	)

	return keySet, paymentInfo, txConversionParams, nil
}

func createSampleCoinsFromTotalAmount(senderSK privacy.PrivateKey, pubkey *operation.Point, totalAmount uint64, numFeeInputs, version int) ([]coin.PlainCoin, error) {
	coinList := []coin.PlainCoin{}
	if version == coin.CoinVersion1 {
		for i := 0; i < numFeeInputs-1; i++ {
			amount := uint64(common.RandIntInterval(0, int(totalAmount)-1))
			coin, err := createSamplePlainCoinV1(senderSK, pubkey, amount, nil)
			if err != nil {
				return nil, err
			}
			coinList = append(coinList, coin)
			totalAmount -= amount
		}
		coin, err := createSamplePlainCoinV1(senderSK, pubkey, totalAmount, nil)
		if err != nil {
			return nil, err
		}
		coinList = append(coinList, coin)
	} else {
		keySet := new(incognitokey.KeySet)
		err := keySet.InitFromPrivateKey(&senderSK)
		if err != nil {
			return nil, err
		}
		for i := 0; i < numFeeInputs-1; i++ {
			amount := uint64(common.RandIntInterval(0, int(totalAmount)-1))
			paymentInfo := key.InitPaymentInfo(keySet.PaymentAddress, amount, []byte("Hello there"))
			coin, err := coin.NewCoinFromPaymentInfo(paymentInfo)
			if err != nil {
				return nil, err
			}
			coinList = append(coinList, coin)
			totalAmount -= amount
		}
		paymentInfo := key.InitPaymentInfo(keySet.PaymentAddress, totalAmount, []byte("Hello there"))
		coin, err := coin.NewCoinFromPaymentInfo(paymentInfo)
		if err != nil {
			return nil, err
		}
		coinList = append(coinList, coin)
	}
	return coinList, nil
}

func createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, feeInputVersion int, tokenID *common.Hash) (*incognitokey.KeySet, *TxTokenConvertVer1ToVer2InitParams, error) {
	var senderSK privacy.PrivateKey
	var keySet *incognitokey.KeySet

	//create a sample test DB
	testDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	bridgeDB := testDB.Copy()

	//generate keyset: we want the public key to be in Shard 0
	for {
		//generate a private key
		senderSK = key.GeneratePrivateKey(common.RandBytes(32))

		//make keySets from privateKey
		keySet = new(incognitokey.KeySet)
		err := keySet.InitFromPrivateKey(&senderSK)

		if err != nil {
			return nil, nil, err
		}

		//we want the public key to belong to Shard 0
		if keySet.PaymentAddress.Pk[31] == 0 {
			break
		}
	}

	//create input tokens
	inputTokens := make([]coin.PlainCoin, numTokenInputs)
	sumInput := uint64(0)
	var err error

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	if err != nil {
		return nil, nil, err
	}

	for i := 0; i < numTokenInputs; i++ {
		amount := uint64(common.RandIntInterval(0, 1000))
		inputTokens[i], err = createSamplePlainCoinV1(senderSK, pubKey, amount, nil)
		if err != nil {
			return nil, nil, err
		}
		sumInput += amount
	}

	//initialize payment info of input token coins
	paymentInfos := coin.CreatePaymentInfosFromPlainCoinsAndAddress(inputTokens, keySet.PaymentAddress, nil)

	//create some fake fee and PRV coins for paying this fee
	realFee := uint64(common.RandIntInterval(0, 1000))

	//we want the sum of these PRV coins to be large than the real fee
	overBalance := uint64(common.RandIntInterval(0, 1000))

	//create PRV fee input coins of version 2 (bigger than realFee => over balance)
	feeInputs, err := createSampleCoinsFromTotalAmount(senderSK, pubKey, realFee+overBalance, numFeeInputs, feeInputVersion)

	if err != nil {
		return nil, nil, err
	}

	//create a return payment for the sender if overbalance > 0
	feePayments := []*privacy.PaymentInfo{}
	if overBalance > 0 {
		for i := 0; i < numFeePayments-1; i++ {
			amount := uint64(common.RandIntInterval(0, int(overBalance)-1))
			feePayment := key.InitPaymentInfo(keySet.PaymentAddress, amount, []byte("Returning over balance to the sender"))
			feePayments = append(feePayments, feePayment)
			overBalance -= amount
		}
		feePayment := key.InitPaymentInfo(keySet.PaymentAddress, overBalance, []byte("Returning over balance to the sender"))
		feePayments = append(feePayments, feePayment)
	}

	//create conversion params
	txTokenConversionParams := NewTxTokenConvertVer1ToVer2InitParams(&keySet.PrivateKey,
		feeInputs,
		feePayments,
		inputTokens,
		paymentInfos,
		realFee,
		testDB,
		bridgeDB,
		tokenID,
		nil,
		nil)

	return keySet, txTokenConversionParams, nil
}

func TestInitializeTxConversion(t *testing.T) {
	for i := 0; i < numTests; i++ {
		_, _, txConversionParams, err := createConversionParams(numInputs, numOutputs, &common.PRVCoinID)
		assert.Equal(t, nil, err, "createConversionParams returns an error: %v", err)

		//Test initializeTxConversion
		txConversionOutput := new(TxVersion2)
		err = initializeTxConversion(txConversionOutput, txConversionParams)

		assert.Equal(t, nil, err, "initializeTxConversion returns an error: %v", err)
	}
}

func TestProveVerifyTxNormalConversion(t *testing.T) {
	for i := 0; i < numTests; i++ {
		txConvertOutput := new(TxVersion2)

		_, _, txConvertParams, err := createConversionParams(numInputs, numOutputs, &common.PRVCoinID)
		assert.Equal(t, nil, err, "createConversionParams returns an error: %v", err)

		initializeTxConversion(txConvertOutput, txConvertParams)

		err = proveConversion(txConvertOutput, txConvertParams)
		assert.Equal(t, nil, err, "proveConversion returns an error: %v", err)

		res, err := validateConversionVer1ToVer2(txConvertOutput, txConvertParams.stateDB, 0, &common.PRVCoinID)
		assert.Equal(t, true, err == nil, "validateConversionVer1ToVer2 should not return any error")
		assert.Equal(t, true, res, "validateConversionVer1ToVer2 should be true")
	}
}

func TestProveVerifyTxNormalConversionTampered(t *testing.T) {
	for i := 0; i < numTests; i++ {
		txConversionOutput := new(TxVersion2)
		_, _, txConversionParams, err := createConversionParams(numInputs, numOutputs, &common.PRVCoinID)
		assert.Equal(t, nil, err, "createConversionParams returns an error: %v", err)

		senderSk := txConversionParams.senderSK
		testDB := txConversionParams.stateDB

		//Initialize conversion and prove
		err = InitConversion(txConversionOutput, txConversionParams)
		assert.Equal(t, nil, err, "proveConversion returns an error: %v", err)

		m := common.RandInt()

		switch m % 5 { //Change this if you want to test a specific case

		//tamper with fee
		case 0:
			fmt.Println("------------------Tampering with fee-------------------")
			txConversionOutput.Fee = uint64(common.RandIntInterval(0, 1000))

			//Re-sign transaction
			txConversionOutput.Sig, txConversionOutput.SigPubKey, err = signNoPrivacy(senderSk, txConversionOutput.Hash()[:])
			assert.Equal(t, nil, err)

		//tamper with randomness
		case 1:
			fmt.Println("------------------Tampering with randomness-------------------")
			inputCoins := txConversionOutput.Proof.GetInputCoins()
			for j := 0; j < numInputs; j++ {
				inputCoins[j].SetRandomness(operation.RandomScalar())
			}

			//Re-sign transaction
			txConversionOutput.Sig, txConversionOutput.SigPubKey, err = signNoPrivacy(senderSk, txConversionOutput.Hash()[:])
			assert.Equal(t, nil, err)

		//attempt to convert used coins (used serial numbers)
		case 2:
			fmt.Println("------------------Tampering with serial numbers-------------------")
			serialToBeStored := [][]byte{}
			for _, inputCoin := range txConversionParams.inputCoins {
				serialToBeStored = append(serialToBeStored, inputCoin.GetKeyImage().ToBytesS())
			}

			err := statedb.StoreSerialNumbers(testDB, *txConversionParams.tokenID, serialToBeStored, 0)
			assert.Equal(t, true, err == nil)

		//tamper with OTAs
		case 3:
			fmt.Println("------------------Tampering with OTAs-------------------")
			otaToBeStored := [][]byte{}
			outputCoinToBeStored := [][]byte{}
			for _, outputCoin := range txConversionOutput.Proof.GetOutputCoins() {
				outputCoinToBeStored = append(outputCoinToBeStored, outputCoin.Bytes())
				otaToBeStored = append(otaToBeStored, outputCoin.GetPublicKey().ToBytesS())
			}

			err := statedb.StoreOTACoinsAndOnetimeAddresses(testDB, *txConversionParams.tokenID, 0, outputCoinToBeStored, otaToBeStored, 0)
			assert.Equal(t, true, err == nil)

		//tamper with commitment of output
		case 4:
			fmt.Println("------------------Tampering with commitment-------------------")
			//fmt.Println("Tx hash before altered", txConversionOutput.Hash().String())
			newOutputCoins := []coin.Coin{}
			for _, outputCoin := range txConversionOutput.Proof.GetOutputCoins() {
				newOutputCoin, ok := outputCoin.(*coin.CoinV2)
				assert.Equal(t, true, ok)

				r := common.RandInt()
				//Attempt to alter some output coins!
				if r%4 < 3 {
					value := new(operation.Scalar).Add(newOutputCoin.GetAmount(), new(operation.Scalar).FromUint64(100))
					randomness := outputCoin.GetRandomness()
					newCommitment := operation.PedCom.CommitAtIndex(value, randomness, operation.PedersenValueIndex)
					newOutputCoin.SetCommitment(newCommitment)
				}
				newOutputCoins = append(newOutputCoins, newOutputCoin)
			}
			txConversionOutput.Proof.SetOutputCoins(newOutputCoins)
			//fmt.Println("Tx hash after altered", txConversionOutput.Hash().String())

			//Re-sign transaction
			txConversionOutput.Sig, txConversionOutput.SigPubKey, err = signNoPrivacy(senderSk, txConversionOutput.Hash()[:])
			assert.Equal(t, nil, err)

		default:

		}
		//Attempt to verify
		res, err := validateConversionVer1ToVer2(txConversionOutput, testDB, 0, &common.PRVCoinID)
		//This validation should return an error
		assert.Equal(t, true, err != nil)
		fmt.Println(err)
		//Result should be false
		assert.Equal(t, false, res)
	}
}

func TestValidateTxTokenConversion(t *testing.T) {
	for i := 0; i < 100; i++ {
		r := common.RandBytes(1)[0]
		tokenID := &common.Hash{r}

		m := common.RandInt()
		switch m%6 {
		//attempt to use a large number of input tokens
		case 0:
			numTokenInputs, numFeeInputs, numFeePayments := 256, 5, 10
			_, txTokenConversionParams, err := createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, 2, tokenID)
			assert.Equal(t, nil, err, "createTokenConversionParams returns an error: %v", err)

			//Test validate txTokenConversionParams
			err = validateTxTokenConvertVer1ToVer2Params(txTokenConversionParams)
			assert.Equal(t, true, err != nil)
		//attempt to use a large number of input PRV fee coins
		case 1:
			numTokenInputs, numFeeInputs, numFeePayments := 5, 256, 10
			_, txTokenConversionParams, err := createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, 2, tokenID)
			assert.Equal(t, nil, err, "createTokenConversionParams returns an error: %v", err)

			//Test validate txTokenConversionParams
			err = validateTxTokenConvertVer1ToVer2Params(txTokenConversionParams)
			assert.Equal(t, true, err != nil)
		//attempt to return a large number of output coins
		case 2:
			numTokenInputs, numFeeInputs, numFeePayments := 5, 10, 256
			_, txTokenConversionParams, err := createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, 2, tokenID)
			assert.Equal(t, nil, err, "createTokenConversionParams returns an error: %v", err)

			//Test validate txTokenConversionParams
			err = validateTxTokenConvertVer1ToVer2Params(txTokenConversionParams)
			assert.Equal(t, true, err != nil)
		//attempt to use PRV coins ver 1 as fee inputs
		case 3:
			numTokenInputs, numFeeInputs, numFeePayments := common.RandIntInterval(0, 255), common.RandIntInterval(0, 255), common.RandIntInterval(0, 255)
			_, txTokenConversionParams, err := createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, 1, tokenID)
			assert.Equal(t, nil, err, "createTokenConversionParams returns an error: %v", err)

			//Test validate txTokenConversionParams
			err = validateTxTokenConvertVer1ToVer2Params(txTokenConversionParams)
			assert.Equal(t, true, err != nil)
		//attempt to tamper with the total amount
		case 4:
			numTokenInputs, numFeeInputs, numFeePayments := common.RandIntInterval(0, 255), common.RandIntInterval(0, 255), common.RandIntInterval(0, 255)
			_, txTokenConversionParams, err := createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, 2, tokenID)
			assert.Equal(t, nil, err, "createTokenConversionParams returns an error: %v", err)

			txTokenConversionParams.tokenParams.tokenPayments[0].Amount += uint64(common.RandIntInterval(1, 1000))
			//Test validate txTokenConversionParams
			err = validateTxTokenConvertVer1ToVer2Params(txTokenConversionParams)
			assert.Equal(t, true, err != nil)
			fmt.Println(err)
		//testing as usual
		default:
			numTokenInputs, numFeeInputs, numFeePayments := common.RandIntInterval(0, 255), common.RandIntInterval(0, 255), common.RandIntInterval(0, 255)
			_, txTokenConversionParams, err := createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, 2, tokenID)
			assert.Equal(t, nil, err, "createTokenConversionParams returns an error: %v", err)

			//Test validate txTokenConversionParams
			err = validateTxTokenConvertVer1ToVer2Params(txTokenConversionParams)
			assert.Equal(t, nil, err, "validateTxTokenConvertVer1ToVer2Params returns an error: %v", err)
		}

	}
}


