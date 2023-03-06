package types

import (
	// "encoding/hex"
	"encoding/json"
	// "fmt"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	numInputs = 5
	// must be 1
	numOutputs         = 1
	unitFeeNativeToken = 100
	//create a sample test DB

)

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
		cmtBytesToBeSaved := [][]byte{inputCoins[i].GetCommitment().ToBytesS()}
		err = statedb.StoreCommitments(dummyDB, *tokenID, cmtBytesToBeSaved, 0)
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
	txConvertParams := NewTxConvertVer1ToVer2InitParams(
		&keySet.PrivateKey,
		paymentInfo,
		inputCoins,
		fee,
		dummyDB,
		tokenID, // use for prv coin -> nil is valid
		nil,
		nil,
	)

	return keySet, paymentInfo, txConvertParams, nil
}

func createAndSaveTokens(numCoins int, tokenID common.Hash, keySets []*incognitokey.KeySet, dummyDB *statedb.StateDB, version int) ([]coin.Coin, error) {
	var err error
	if version == coin.CoinVersion1 {
		coinsToBeSaved := make([]coin.Coin, numCoins*len(keySets))
		for i, keySet := range keySets {
			pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
			if err != nil {
				return nil, err
			}
			for j := 0; j < numCoins; j++ {
				amount := uint64(common.RandIntInterval(0, 100000000))
				tmpCoin, err := createSamplePlainCoinV1(keySet.PrivateKey, pubKey, amount, nil)
				tmpCoin2 := new(coin.CoinV1)
				tmpCoin2.CoinDetails = tmpCoin
				if err != nil {
					return nil, err
				}
				coinsToBeSaved[i*numCoins+j] = tmpCoin2
			}
		}
		cmtBytesToBeSaved := make([][]byte, 0)
		for _, coin := range coinsToBeSaved {
			cmtBytesToBeSaved = append(cmtBytesToBeSaved, coin.GetCommitment().ToBytesS())
		}
		err = statedb.StoreCommitments(dummyDB, tokenID, cmtBytesToBeSaved, 0)

		return coinsToBeSaved, err
	} else {
		coinsToBeSaved := make([]coin.Coin, numCoins*len(keySets))
		for i, keySet := range keySets {
			for j := 0; j < numCoins; j++ {
				amount := uint64(common.RandIntInterval(0, 100000000))
				paymentInfo := key.InitPaymentInfo(keySet.PaymentAddress, amount, []byte("Dummy token"))

				tmpCoin, err := coin.NewCoinFromPaymentInfo(privacy.NewCoinParams().FromPaymentInfo(paymentInfo))
				if err != nil {
					return nil, err
				}

				err = tmpCoin.ConcealOutputCoin(keySet.PaymentAddress.GetPublicView())
				if err != nil {
					return nil, err
				}
				coinsToBeSaved[i*numCoins+j] = tmpCoin
			}
		}

		coinsBytesToBeSaved := make([][]byte, 0)
		otasToBeSaved := make([][]byte, 0)
		for _, c := range coinsToBeSaved {
			coinsBytesToBeSaved = append(coinsBytesToBeSaved, c.Bytes())
			otasToBeSaved = append(otasToBeSaved, c.GetPublicKey().ToBytesS())
		}
		err = statedb.StoreOTACoinsAndOnetimeAddresses(dummyDB, tokenID, 0, coinsBytesToBeSaved, otasToBeSaved, 0)
		if err != nil {
			return nil, err
		}
		return coinsToBeSaved, nil
	}

}

func prepareKeySets(numKeySets int) ([]*incognitokey.KeySet, error) {
	keySets := make([]*incognitokey.KeySet, numKeySets)
	//generate keysets: we want the public key to be in Shard 0
	for i := 0; i < numKeySets; i++ {
		for {
			//generate a private key
			privateKey := key.GeneratePrivateKey(common.RandBytes(32))

			//make keySets from privateKey
			keySet := new(incognitokey.KeySet)
			err := keySet.InitFromPrivateKey(&privateKey)

			if err != nil {
				return nil, err
			}

			//we want the public key to belong to Shard 0
			if keySet.PaymentAddress.Pk[31] == 0 {
				keySets[i] = keySet
				break
			}
		}
	}
	return keySets, nil
}

func createSamplePlainCoinsFromTotalAmount(senderSK privacy.PrivateKey, pubkey *operation.Point, totalAmount uint64, numFeeInputs, version int) ([]coin.PlainCoin, error) {
	coinList := []coin.PlainCoin{}
	tmpAmount := totalAmount / uint64(numFeeInputs)
	if version == coin.CoinVersion1 {
		for i := 0; i < numFeeInputs-1; i++ {
			amount := tmpAmount - uint64(common.RandIntInterval(0, int(tmpAmount)/2))
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
			coin, err := coin.NewCoinFromPaymentInfo(privacy.NewCoinParams().FromPaymentInfo(paymentInfo))
			if err != nil {
				return nil, err
			}
			coinList = append(coinList, coin)
			totalAmount -= amount
		}
		paymentInfo := key.InitPaymentInfo(keySet.PaymentAddress, totalAmount, []byte("Hello there"))
		coin, err := coin.NewCoinFromPaymentInfo(privacy.NewCoinParams().FromPaymentInfo(paymentInfo))
		if err != nil {
			return nil, err
		}
		coinList = append(coinList, coin)
	}
	return coinList, nil
}

func createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, feeInputVersion int, tokenID *common.Hash) (*incognitokey.KeySet, *TxTokenConvertVer1ToVer2InitParams, error) {
	keySets, err := prepareKeySets(1)
	if err != nil {
		return nil, nil, err
	}
	keySet := keySets[0]
	senderSK := keySet.PrivateKey

	//create input tokens
	inputTokens := make([]coin.PlainCoin, numTokenInputs)
	sumInput := uint64(0)

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

	//create PRV fee input coins of version 2 (bigger than realFee => over balance) and store onto the database
	feeInputs, err := createSamplePlainCoinsFromTotalAmount(senderSK, pubKey, realFee+overBalance, numFeeInputs, feeInputVersion)
	for _, feeInput := range feeInputs {
		keyImage, err := feeInput.ParseKeyImageWithPrivateKey(senderSK)
		if err != nil {
			return nil, nil, err
		}
		feeInput.SetKeyImage(keyImage)
	}
	feeInputBytes := [][]byte{}
	otas := [][]byte{}
	for _, feeInput := range feeInputs {
		feeInputBytes = append(feeInputBytes, feeInput.Bytes())
		otas = append(otas, feeInput.GetPublicKey().ToBytesS())
	}
	err = statedb.StoreOTACoinsAndOnetimeAddresses(dummyDB, common.PRVCoinID, 0, feeInputBytes, otas, 0)
	if err != nil {
		return nil, nil, err
	}

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
		dummyDB,
		bridgeDB,
		tokenID,
		nil,
		nil)

	return keySet, txTokenConversionParams, nil
}

func createInitTokenParams(theInputCoin coin.Coin, db *statedb.StateDB, tokenID, tokenName string, keySet *incognitokey.KeySet) (*tx_generic.TxTokenParams, *tx_generic.TokenParam, error) {
	msgCipherText := []byte("Testing Init Token")
	initAmount := uint64(1000000000)
	tokenPayments := []*privacy.PaymentInfo{{PaymentAddress: keySet.PaymentAddress, Amount: initAmount, Message: msgCipherText}}

	myOnlyInputCoin, err := theInputCoin.Decrypt(keySet)
	if err != nil {
		return nil, nil, err
	}
	inputCoinsPRV := []coin.PlainCoin{myOnlyInputCoin}
	paymentInfoPRV := []*privacy.PaymentInfo{key.InitPaymentInfo(keySet.PaymentAddress, uint64(100), []byte("test out"))}

	// token param for init new token
	tokenParam := &tx_generic.TokenParam{
		PropertyID:     tokenID,
		PropertyName:   tokenName,
		PropertySymbol: "DEFAULT",
		Amount:         initAmount,
		TokenTxType:    utils.CustomTokenInit,
		Receiver:       tokenPayments,
		TokenInput:     []coin.PlainCoin{},
		Mintable:       false,
		Fee:            0,
	}

	paramToCreateTx := tx_generic.NewTxTokenParams(&keySet.PrivateKey,
		paymentInfoPRV, inputCoinsPRV, 10, tokenParam, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{}, db)
	return paramToCreateTx, tokenParam, nil
}

func TestInitializeTxConversion(t *testing.T) {
	Convey("Conversion Parameter Test", t, func() {
		_, _, txConvertParams, err := createConversionParams(numInputs, numOutputs, &common.PRVCoinID)
		So(err, ShouldBeNil)
		//Test initializeTxConversion
		txConvertOutput := new(Tx)
		err = initializeTxConversion(txConvertOutput, txConvertParams)
		So(err, ShouldBeNil)
	})
}

func TestProveVerifyTxNormalConversion(t *testing.T) {
	txConvertOutput := new(Tx)
	var txConvertParams *TxConvertVer1ToVer2InitParams
	var err error

	Convey("TX Convert", t, func() {
		Convey("create TX", func() {
			_, _, txConvertParams, err = createConversionParams(numInputs, numOutputs, &common.PRVCoinID)
			So(err, ShouldBeNil)
			err = initializeTxConversion(txConvertOutput, txConvertParams)
			So(err, ShouldBeNil)

			err = proveConversion(txConvertOutput, txConvertParams)
			So(err, ShouldBeNil)
		})

		Convey("should verify TX", func() {
			res, err := tx_generic.ValidateSanity(txConvertOutput, nil, nil, nil, 0)
			So(err, ShouldBeNil)
			So(res, ShouldBeTrue)

			res, err = tx_generic.MdValidate(txConvertOutput, false, txConvertParams.stateDB, nil, 0)
			So(err, ShouldBeNil)
			So(res, ShouldBeTrue)

			boolParams := make(map[string]bool)
			boolParams["hasPrivacy"] = false
			boolParams["isNewTransaction"] = true
			res, err = txConvertOutput.ValidateTxByItself(boolParams, txConvertParams.stateDB, nil, nil, 0, nil, nil)
			So(err, ShouldBeNil)
			So(res, ShouldBeTrue)
			Println("should reject since PRV inputs are not in db")
		})
	})
}

func TestProveVerifyTxNormalConversionTampered(t *testing.T) {
	txConvertOutput := new(Tx)
	var txConvertParams *TxConvertVer1ToVer2InitParams
	var err error
	var senderSk *privacy.PrivateKey

	Convey("TX Convert - Reject Cases", t, func() {
		Convey("create TX", func() {
			_, _, txConvertParams, err = createConversionParams(numInputs, numOutputs, &common.PRVCoinID)
			So(err, ShouldBeNil)
			err = initializeTxConversion(txConvertOutput, txConvertParams)
			So(err, ShouldBeNil)
			err = proveConversion(txConvertOutput, txConvertParams)
			So(err, ShouldBeNil)
			senderSk = txConvertParams.senderSK
			// Initialize conversion and prove
			err = InitConversion(txConvertOutput, txConvertParams)
			So(err, ShouldBeNil)
		})

		Convey("change TX", func() {
			m := common.RandInt()

			switch m % 5 { //Change this if you want to test a specific case

			//tamper with fee
			case 0:
				Print("--Tampering with fee--")
				txConvertOutput.Fee = uint64(common.RandIntInterval(0, 1000))

				//Re-sign transaction
				txConvertOutput.Sig, txConvertOutput.SigPubKey, err = tx_generic.SignNoPrivacy(senderSk, txConvertOutput.Hash()[:])
				So(err, ShouldBeNil)

			//tamper with randomness
			case 1:
				Print("--Tampering with randomness--")
				inputCoins := txConvertOutput.Proof.GetInputCoins()
				for j := 0; j < numInputs; j++ {
					inputCoins[j].SetRandomness(operation.RandomScalar())
				}

				//Re-sign transaction
				txConvertOutput.Sig, txConvertOutput.SigPubKey, err = tx_generic.SignNoPrivacy(senderSk, txConvertOutput.Hash()[:])
				So(err, ShouldBeNil)

			//attempt to convert used coins (used serial numbers)
			case 2:
				Print("--Tampering with serial numbers--")
				usedIndex := common.RandInt() % len(txConvertOutput.Proof.GetInputCoins())
				inputCoin := txConvertOutput.Proof.GetInputCoins()[usedIndex]

				err := statedb.StoreSerialNumbers(dummyDB, *txConvertParams.tokenID, [][]byte{inputCoin.GetKeyImage().ToBytesS()}, 0)
				So(err, ShouldBeNil)

			//tamper with OTAs
			case 3:
				Print("--Tampering with OTAs--")
				usedIndex := common.RandInt() % len(txConvertOutput.Proof.GetOutputCoins())
				outputCoin := txConvertOutput.Proof.GetOutputCoins()[usedIndex]

				So(storeCoins(dummyDB, []coin.Coin{outputCoin}, 0, *txConvertParams.tokenID), ShouldBeNil)

			//tamper with commitment of output
			case 4:
				Print("--Tampering with commitment--")
				//Println("Tx hash before altered", txConvertOutput.Hash().String())
				newOutputCoins := []coin.Coin{}
				outputCoin := txConvertOutput.Proof.GetOutputCoins()[0]
				newOutputCoin, ok := outputCoin.(*coin.CoinV2)
				So(ok, ShouldBeTrue)

				//Attempt to alter some output coins!
				value := new(operation.Scalar).Add(newOutputCoin.GetAmount(), new(operation.Scalar).FromUint64(100))
				randomness := outputCoin.GetRandomness()
				newCommitment := operation.PedCom.CommitAtIndex(value, randomness, operation.PedersenValueIndex)
				newOutputCoin.SetCommitment(newCommitment)
				newOutputCoins = append(newOutputCoins, newOutputCoin)
				txConvertOutput.Proof.SetOutputCoins(newOutputCoins)
				//Println("Tx hash after altered", txConvertOutput.Hash().String())

				//Re-sign transaction
				txConvertOutput.Sig, txConvertOutput.SigPubKey, err = tx_generic.SignNoPrivacy(senderSk, txConvertOutput.Hash()[:])
				So(err, ShouldBeNil)
			default:

			}
			//Attempt to verify
			isValidSanity, _ := txConvertOutput.ValidateSanityData(nil, nil, nil, 0)
			var isValid bool
			if !isValidSanity {
				isValid = false
			} else {
				boolParams := make(map[string]bool)
				boolParams["hasPrivacy"] = false
				boolParams["isNewTransaction"] = true
				isValid, err = txConvertOutput.ValidateTxByItself(boolParams, txConvertParams.stateDB, nil, nil, 0, nil, nil)
				tx_generic.MdValidate(txConvertOutput, false, txConvertParams.stateDB, nil, 0)
			}
			if isValid {
				jsb, _ := json.Marshal(txConvertOutput)
				Printf("Unexpected error in case %d : %v\nTransaction : %s\n", m%5, err, string(jsb))
				// return
			}
			//This validation should return an error
			So(isValid, ShouldBeFalse)
		})
	})
}

func TestValidateTxTokenConversion(t *testing.T) {
	r := common.RandBytes(1)[0]
	tokenID := &common.Hash{r}

	Convey("Convert Token TX", t, func() {
		m := common.RandInt()
		switch m % 6 {
		//attempt to use a large number of input tokens
		case 0:
			numTokenInputs, numFeeInputs, numFeePayments := 256, 5, 10
			_, txTokenConversionParams, err := createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, 2, tokenID)
			So(err, ShouldBeNil)

			//Test validate txTokenConversionParams
			err = validateTxTokenConvertVer1ToVer2Params(txTokenConversionParams)
			So(err, ShouldNotBeNil)
		//attempt to use a large number of input PRV fee coins
		case 1:
			numTokenInputs, numFeeInputs, numFeePayments := 5, 256, 10
			_, txTokenConversionParams, err := createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, 2, tokenID)
			So(err, ShouldBeNil)

			//Test validate txTokenConversionParams
			err = validateTxTokenConvertVer1ToVer2Params(txTokenConversionParams)
			So(err, ShouldNotBeNil)
		//attempt to return a large number of output coins
		case 2:
			numTokenInputs, numFeeInputs, numFeePayments := 5, 10, 256
			_, txTokenConversionParams, err := createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, 2, tokenID)
			So(err, ShouldBeNil)

			//Test validate txTokenConversionParams
			err = validateTxTokenConvertVer1ToVer2Params(txTokenConversionParams)
			So(err, ShouldNotBeNil)
		//attempt to use PRV coins ver 1 as fee inputs
		case 3:
			// storeOTA cannot be used to store coinv1 or the db will be corrupted
			return
			numTokenInputs, numFeeInputs, numFeePayments := common.RandIntInterval(0, 255), common.RandIntInterval(0, 255), common.RandIntInterval(0, 255)
			_, txTokenConversionParams, err := createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, 1, tokenID)
			So(err, ShouldBeNil)

			//Test validate txTokenConversionParams
			err = validateTxTokenConvertVer1ToVer2Params(txTokenConversionParams)
			So(err, ShouldNotBeNil)
		//attempt to tamper with the total amount
		case 4:
			numTokenInputs, numFeeInputs, numFeePayments := common.RandIntInterval(0, 255), common.RandIntInterval(0, 255), common.RandIntInterval(0, 255)
			_, txTokenConversionParams, err := createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, 2, tokenID)
			So(err, ShouldBeNil)

			txTokenConversionParams.tokenParams.tokenPayments[0].Amount += uint64(common.RandIntInterval(1, 1000))
			//Test validate txTokenConversionParams
			err = validateTxTokenConvertVer1ToVer2Params(txTokenConversionParams)
			So(err, ShouldNotBeNil)

		// default: create valid conversion
		default:
			numTokenInputs, numFeeInputs, numFeePayments := common.RandIntInterval(0, 100), common.RandIntInterval(0, 100), common.RandIntInterval(0, 100)
			_, txTokenConversionParams, err := createTokenConversionParams(numTokenInputs, numFeeInputs, numFeePayments, 2, tokenID)
			So(err, ShouldBeNil)

			//Test validate txTokenConversionParams
			err = validateTxTokenConvertVer1ToVer2Params(txTokenConversionParams)
			So(err, ShouldBeNil)
		}
	})
}
