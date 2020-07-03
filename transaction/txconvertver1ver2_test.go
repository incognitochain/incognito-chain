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
	numInputs = 10
	numOutputs = 5
	numTests = 100
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

func createSampleInputCoin(privkey privacy.PrivateKey, pubKey *operation.Point, amount uint64, msg []byte) (*coin.PlainCoinV1, error) {
	c := new(coin.PlainCoinV1).Init()

	c.SetValue(amount)
	c.SetInfo(msg)
	c.SetPublicKey(pubKey)
	c.SetSNDerivator(operation.RandomScalar())
	c.SetRandomness(operation.RandomScalar())

	//Derive serial number from snDerivator
	c.SetKeyImage(new(operation.Point).Derive(privacy.PedCom.G[0], new(operation.Scalar).FromBytesS(privkey), c.GetSNDerivator()))

	//Create commitment
	err := c.CommitAll()

	if err != nil {
		return nil, err
	}

	return c, nil
}

func createConversionParams(numInputs, numOutputs int)(*incognitokey.KeySet,
	[]*privacy.PaymentInfo, *TxConvertVer1ToVer2InitParams, error){
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

	for i := 0; i< numInputs; i++ {
		amount := uint64(common.RandIntInterval(0, 1000))
		inputCoins[i], err = createSampleInputCoin(senderSK, pubKey, amount, nil)
		if err != nil {
			return nil, nil, nil, err
		}
		sumInput += amount
	}

	sumOutput := uint64(0)

	paymentInfo := make([]*key.PaymentInfo, numOutputs)

	for i :=0;i< numOutputs;i++{
		amount := uint64(common.RandIntInterval(0, int(sumInput-sumOutput)))
		paymentInfo[i] = key.InitPaymentInfo(keySet.PaymentAddress, amount, nil)
		sumOutput += amount
	}

	//calculate sample fee
	fee := sumInput - sumOutput

	//create conversion params
	txConvertParams := NewTxConvertVer1ToVer2InitParams(
		&keySet.PrivateKey,
		paymentInfo,
		inputCoins,
		fee,
		testDB,
		&common.PRVCoinID, // use for prv coin -> nil is valid
		nil,
		nil,
	)

	return keySet, paymentInfo, txConvertParams, nil
}

func TestInitializeTxConversion(t *testing.T) {
	for i := 0; i < numTests; i++ {
		_, _, txConvertParams, err := createConversionParams(numInputs, numOutputs)
		assert.Equal(t, nil, err, "createConversionParams returns an error: %v", err)

		//Test initializeTxConversion
		txConversionOutput := new(TxVersion2)
		err = initializeTxConversion(txConversionOutput, txConvertParams)

		assert.Equal(t, nil, err, "initializeTxConversion returns an error: %v", err)
	}
}

func TestProveVerifyTxNormalConversion(t *testing.T){
	for i:=0; i< numTests; i++ {
		txConvertOutput := new(TxVersion2)

		_, _, txConvertParams, err := createConversionParams(numInputs, numOutputs)
		assert.Equal(t, nil, err, "createConversionParams returns an error: %v", err)

		initializeTxConversion(txConvertOutput, txConvertParams)

		err = proveConversion(txConvertOutput, txConvertParams)
		assert.Equal(t, nil, err, "proveConversion returns an error: %v", err)

		res, err := validateConversionVer1ToVer2(txConvertOutput, txConvertParams.stateDB, 0, &common.PRVCoinID)
		assert.Equal(t, true, err==nil, "validateConversionVer1ToVer2 should not return any error")
		assert.Equal(t, true, res, "validateConversionVer1ToVer2 should be true")
	}
}

func TestProveVerifyTxNormalConversionTampered(t *testing.T) {
	for i := 0; i < numTests; i++ {
		txConversionOutput := new(TxVersion2)
		_, _, txConversionParams, err := createConversionParams(numInputs, numOutputs)
		assert.Equal(t, nil, err, "createConversionParams returns an error: %v", err)

		senderSk := txConversionParams.senderSK
		testDB := txConversionParams.stateDB

		//Initialize conversion and prove
		err = InitConversion(txConversionOutput, txConversionParams)
		assert.Equal(t, nil, err, "proveConversion returns an error: %v", err)

		m := common.RandInt()

		switch m % 5{
		case 0: //tamper with fee
			fmt.Println("------------------Tampering with fee-------------------")
			txConversionOutput.Fee = uint64(common.RandIntInterval(0, 1000))

			//Re-sign transaction
			txConversionOutput.Sig, txConversionOutput.SigPubKey, err = signNoPrivacy(senderSk, txConversionOutput.Hash()[:])
			assert.Equal(t, nil, err)

		case 1: //tamper with randomness => current code will fail this test
			fmt.Println("------------------Tampering with randomness-------------------")
			inputCoins := txConversionOutput.Proof.GetInputCoins()
			for j := 0; j < numInputs; j++ {
				inputCoins[j].SetRandomness(operation.RandomScalar())
			}

			//Re-sign transaction
			txConversionOutput.Sig, txConversionOutput.SigPubKey, err = signNoPrivacy(senderSk, txConversionOutput.Hash()[:])
			assert.Equal(t, nil, err)

		case 2: //attempt to convert used coins (used serial numbers)
			fmt.Println("------------------Tampering with serial numbers-------------------")
			serialToBeStored := [][]byte{}
			for _, inputCoin := range txConversionParams.inputCoins {
				serialToBeStored = append(serialToBeStored, inputCoin.GetKeyImage().ToBytesS())
			}

			err := statedb.StoreSerialNumbers(testDB, *txConversionParams.tokenID, serialToBeStored, 0)
			assert.Equal(t, true, err == nil)

		case 3: //tamper with OTAs
			fmt.Println("------------------Tampering with OTAs-------------------")
			otaToBeStored := [][]byte{}
			outputCoinToBeStored := [][]byte{}
			for _, outputCoin := range txConversionOutput.Proof.GetOutputCoins() {
				outputCoinToBeStored = append(outputCoinToBeStored, outputCoin.Bytes())
				otaToBeStored = append(otaToBeStored, outputCoin.GetPublicKey().ToBytesS())
			}

			err := statedb.StoreOTACoinsAndOnetimeAddresses(testDB, *txConversionParams.tokenID, 0, outputCoinToBeStored, otaToBeStored, 0)
			assert.Equal(t, true, err == nil)

		case 4: //tamper with commitment of output => current code will fail this test
			fmt.Println("------------------Tampering with commitment-------------------")
			//fmt.Println("Tx hash before altered", txConversionOutput.Hash().String())
			newOutputCoins := []coin.Coin{}
			for _, outputCoin := range txConversionOutput.Proof.GetOutputCoins() {
				newOutputCoin, ok := outputCoin.(*coin.CoinV2)
				assert.Equal(t, true, ok)

				r := common.RandInt()
				//Attempt to alter some output coins!
				if r % 4 < 3 {
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
		//Attemp to verify
		res, err := validateConversionVer1ToVer2(txConversionOutput, testDB, 0, &common.PRVCoinID)
		//This validation should return an error
		assert.Equal(t, true, err != nil)
		//Result should be false
		assert.Equal(t, false, res)
	}
}
