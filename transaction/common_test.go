package transaction

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
)

func mockHasSnD(stateDB *statedb.StateDB, tokenID common.Hash, snd []byte) (bool, error) { return false, nil }
func resetTxDbWrapper() { txDatabaseWrapper = NewTxDbWrapper() }

func createTokenParams(inputPRVCoin coin.Coin, inputTokens []coin.PlainCoin, db *statedb.StateDB, tokenID, tokenName string, keySet *incognitokey.KeySet, txTokenType, version int)(*TxTokenParams, *TokenParam, error){
	if version == coin.CoinVersion1 {
		switch txTokenType {
		case CustomTokenInit:
			initAmount := uint64(1000000000)
			tokenPayments := []*privacy.PaymentInfo{{PaymentAddress: keySet.PaymentAddress, Amount: initAmount, Message: nil}}

			feeInputCoin, err := inputPRVCoin.Decrypt(keySet)
			if err != nil {
				return nil, nil, err
			}
			inputPRVCoins := []coin.PlainCoin{feeInputCoin}
			paymentInfoPRV := []*privacy.PaymentInfo{key.InitPaymentInfo(keySet.PaymentAddress, uint64(100), []byte("test out"))}

			// token param for init new token
			tokenParam := &TokenParam{
				PropertyID:     tokenID,
				PropertyName:   tokenName,
				PropertySymbol: "DEFAULT",
				Amount:         initAmount,
				TokenTxType:    CustomTokenInit,
				Receiver:       tokenPayments,
				TokenInput:     []coin.PlainCoin{},
				Mintable:       false,
				Fee:            0,
			}

			txTokenParams := NewTxTokenParams(&keySet.PrivateKey,
				paymentInfoPRV, inputPRVCoins, 0, tokenParam, db, nil,
				hasPrivacyForPRV, false, shardID, []byte{}, db)

			return txTokenParams, tokenParam, nil
		}
	}else{
		switch txTokenType {
		case CustomTokenInit:
			initAmount := uint64(1000000000)
			tokenPayments := []*privacy.PaymentInfo{{PaymentAddress: keySet.PaymentAddress, Amount: initAmount, Message: nil}}

			feeInputCoin, err := inputPRVCoin.Decrypt(keySet)
			if err != nil {
				return nil, nil, err
			}
			inputPRVCoins := []coin.PlainCoin{feeInputCoin}
			paymentInfoPRV := []*privacy.PaymentInfo{key.InitPaymentInfo(keySet.PaymentAddress, uint64(100), []byte("test out"))}

			// token param for init new token
			tokenParam := &TokenParam{
				PropertyID:     tokenID,
				PropertyName:   tokenName,
				PropertySymbol: "DEFAULT",
				Amount:         initAmount,
				TokenTxType:    CustomTokenInit,
				Receiver:       tokenPayments,
				TokenInput:     []coin.PlainCoin{},
				Mintable:       false,
				Fee:            0,
			}

			txTokenParams := NewTxTokenParams(&keySet.PrivateKey,
				paymentInfoPRV, inputPRVCoins, 0, tokenParam, db, nil,
				hasPrivacyForPRV, false, shardID, []byte{}, db)

			return txTokenParams, tokenParam, nil
		}
	}
	return nil, nil, nil
}

func createSampleTx(numInputs, numOutputs int, txType string, txTokenType int, hasPrivacy bool, keySet *incognitokey.KeySet, version int) (metadata.Transaction, error){
	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	if err != nil {
		return nil, err
	}
	if version == coin.CoinVersion1 {
		if txType == common.TxCustomTokenPrivacyType{
			return createSampleTxTokenVer1(numInputs, numOutputs, txTokenType, hasPrivacy, keySet)
		}
		return createSampleTxVer1(numInputs, numOutputs, txType, hasPrivacy, keySet, pubKey)
	} else{
		if txType == common.TxCustomTokenPrivacyType{
			return createSampleTxTokenVer2(numInputs, numOutputs, txTokenType, hasPrivacy, keySet)
		}
		return createSampleTxVer2(numInputs, numOutputs, txType, hasPrivacy, keySet)
	}
}

func createSampleTxVer1(numInputs, numOutputs int, txType string, hasPrivacy bool, keySet *incognitokey.KeySet, pubKey *operation.Point) (metadata.Transaction, error) {
	coins, err := createAndSaveCoinV1s(100, 0, keySet.PrivateKey, pubKey, testDB)
	if err != nil {
		return nil, err
	}
	switch txType{
	case common.TxNormalType:
		inputCoins := coins[:numInputs]
		if err != nil {
			return nil, err
		}
		_, txPrivacyInitParam, err := createTxPrivacyInitParams(keySet, inputCoins, hasPrivacy, numOutputs)
		if err != nil {
			return nil, err
		}
		tx := new(TxVersion1)
		err = tx.Init(txPrivacyInitParam)
		return tx, err
	case common.TxRewardType:
		tx := new(TxVersion1)
		err = tx.InitTxSalary(uint64(1000000000000), &keySet.PaymentAddress, &keySet.PrivateKey, testDB, nil)
		return tx, err

	}

	return nil, nil
}

func createSampleTxVer2(numInputs, numOutputs int, txType string, hasPrivacy bool, keySet *incognitokey.KeySet) (metadata.Transaction, error) {
	coins, err := createAndSaveTokens(100, common.PRVCoinID, []*incognitokey.KeySet{keySet}, testDB, 2)
	if err != nil {
		return nil, err
	}
	switch txType{
	case common.TxNormalType:
		tmpCoins := coins[:numInputs]
		inputCoins := make([]coin.PlainCoin, len(tmpCoins))
		for i, tmpCoin := range tmpCoins {
			tmpCoin2, ok := tmpCoin.(*coin.CoinV2)
			if !ok {
				return nil, errors.New("Cannot parse coin")
			}
			inputCoins[i], err = tmpCoin2.Decrypt(keySet)
			if err != nil {
				return nil, err
			}
		}
		if err != nil {
			return nil, err
		}
		_, txPrivacyInitParam, err := createTxPrivacyInitParams(keySet, inputCoins, hasPrivacy, numOutputs)
		if err != nil {
			return nil, err
		}
		tx := new(TxVersion2)
		err = tx.Init(txPrivacyInitParam)
		return tx, err
	case common.TxRewardType:
		tx := new(TxVersion2)

		amount := uint64(common.RandIntInterval(0, 100000000))
		paymentInfo := key.InitPaymentInfo(keySet.PaymentAddress, amount, nil)
		inputCoin, err := coin.NewCoinFromPaymentInfo(paymentInfo)

		err = tx.InitTxSalary(inputCoin, &keySet.PrivateKey, testDB, nil)
		return tx, err
	}
	return nil, nil
}

func createSampleTxTokenVer1(numInputs, numOutputs int, txTokenType int, hasPrivacy bool, keySet *incognitokey.KeySet) (metadata.Transaction, error){
	tokenID := common.Hash{10}
	//create some PRVCoin
	coins, err := createAndSaveTokens(100, common.PRVCoinID, []*incognitokey.KeySet{keySet}, testDB, 1)
	if err != nil {
		return nil, err
	}
	theInputCoin, ok := coins[0].(*coin.CoinV1)
	if !ok {
		return nil, errors.New("Cannot parse coin")
	}
	paramToCreateTx, _, err := createTokenParams(theInputCoin, nil, testDB, tokenID.String(), "TOKENNAME", keySet, txTokenType, 1)

	tx := new(TxTokenVersion1)
	err = tx.Init(paramToCreateTx)
	return tx, err
}

func createSampleTxTokenVer2(numInputs, numOutputs int, txTokenType int, hasPrivacy bool, keySet *incognitokey.KeySet) (metadata.Transaction, error){
	tokenID := common.Hash{10}
	//create some PRVCoin
	coins, err := createAndSaveTokens(100, common.PRVCoinID, []*incognitokey.KeySet{keySet}, testDB, 2)
	if err != nil {
		return nil, err
	}
	theInputCoin, ok := coins[0].(*coin.CoinV2)
	if !ok {
		return nil, errors.New("Cannot parse coin")
	}
	paramToCreateTx, _, err := createTokenParams(theInputCoin, nil, testDB, tokenID.String(), "TOKENNAME", keySet, txTokenType, 2)

	tx := new(TxTokenVersion2)
	err = tx.Init(paramToCreateTx)
	return tx, err
}

func TestEstimateTxSize(t *testing.T) {
	txDatabaseWrapper.hasSNDerivator = mockHasSnD
	defer resetTxDbWrapper()

	key, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = key.KeySet.InitFromPrivateKey(&key.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	paymentAddress := key.KeySet.PaymentAddress
	tx := &TxVersion1{}
	err = tx.InitTxSalary(10, &paymentAddress, &key.KeySet.PrivateKey, nil, nil)
	if err != nil {
		t.Error(err)
	}

	payments := []*privacy.PaymentInfo{&privacy.PaymentInfo{
		PaymentAddress: paymentAddress,
		Amount:         5,
	}}

	outputCoins := tx.Proof.GetOutputCoins()
	size := EstimateTxSize(NewEstimateTxSizeParam(len(outputCoins), len(payments), true, nil, nil, 1))
	fmt.Println(size)
	assert.Greater(t, size, uint64(0))

	privacyCustomTokenParams := TokenParam{
		Receiver: []*privacy.PaymentInfo{{
			PaymentAddress: paymentAddress, Amount: 5,
		}},
	}
	size2 := EstimateTxSize(NewEstimateTxSizeParam(len(outputCoins), len(payments), true, nil, &privacyCustomTokenParams, 1))
	fmt.Println(size2)
	assert.Greater(t, size2, uint64(0))
}

func TestGetTxActualSize(t *testing.T) {
	keySets, err := prepareKeySets(1)
	assert.Equal(t, nil, err, "prepareKeySets returns an error: %v", err)
	keySet := keySets[0]

	for i:=0;i<10;i++ {
		numInputs := common.RandIntInterval(1, 10)
		numOutputs := common.RandIntInterval(2, 2)
		m := common.RandInt()

		version := common.RandInt() % 2 + 1

		var txType string
		var hasPrivacy bool
		var txTokenType int = -1
		//choose transaction type
		switch m % 5{
		case 0:
			txType = "s"
			if version == coin.CoinVersion1 {
				numInputs = 0
			}else{
				numInputs = 1
			}
			numOutputs = 1
			hasPrivacy = false
		case 1:
			txType = "tp"
			hasPrivacy = true
			txTokenType = CustomTokenInit
		case 2: //PRV conversion transaction
			tx := new(TxVersion2)

			_, _, txConvertParams, err := createConversionParams(numInputs, 1, &common.PRVCoinID)
			assert.Equal(t, nil, err, "createConversionParams returns an error: %v", err)

			err = initializeTxConversion(tx, txConvertParams)
			assert.Equal(t, nil, err, "initializeTxConversion returns an error: %v", err)

			err = proveConversion(tx, txConvertParams)
			assert.Equal(t, nil, err, "proveConversion returns an error: %v", err)

			res, err := tx.ValidateSanityData( nil, nil, nil, 0)
			assert.Equal(t, nil, err, "ValidateSanityData returns an error: %v", err)
			assert.Equal(t, true, res)

			res, err = validateTransaction(tx,false, testDB, testDB, 0, &common.PRVCoinID, false, true)
			// res, err := validateConversionVer1ToVer2(txConvertOutput, txConvertParams.stateDB, 0, &common.PRVCoinID)
			assert.Equal(t, true, err == nil, "validateTransaction returns an error: %v", err)
			assert.Equal(t, true, res)

			actualTxSize := getTxActualSizeInBytes(tx)
			fmt.Printf("TxSize: version=%d, type=%s, numInputs=%d, numOutputs=%d, hasPrivacy=%v, txSizeInBytes=%d\n", tx.GetVersion(), tx.GetType(), len(tx.GetProof().GetInputCoins()), len(tx.GetProof().GetOutputCoins()), tx.IsPrivacy(), actualTxSize)
			continue
		case 3: //token conversion transaction
			keySets, err := prepareKeySets(1)
			assert.Equal(t, nil, err, "prepareKeySets returns an errors: %v", err)

			//create and save some PRV coins
			coins, err := createAndSaveTokens(100, common.PRVCoinID, keySets, testDB, 2)
			assert.Equal(t, nil, err, "createAndSaveTokens returns an error: %v", err)

			tokenName := "Token" + string(i)
			theInputCoin, ok := coins[i].(coin.Coin)
			assert.Equal(t, true, ok, "Cannot parse coin")
			paramToCreateTx, tokenParam, err := createInitTokenParams(theInputCoin, testDB, "", tokenName, keySets[0])
			tx := &TxTokenVersion2{}

			err = tx.Init(paramToCreateTx)
			if err != nil {
				jsb, _ := json.Marshal(coins[i])
				fmt.Printf("Loop %d : Init returns an error: %v\nThat coin is %s\n", i, err, string(jsb))
				return
			}

			tokenOutputs := tx.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
			feeOutputs := tx.GetTxBase().GetProof().GetOutputCoins()
			forceSaveCoins(testDB, feeOutputs, 0, common.PRVCoinID, t)
			statedb.StorePrivacyToken(testDB, *tx.GetTokenID(), tokenParam.PropertyName, tokenParam.PropertySymbol, statedb.InitToken, tokenParam.Mintable, tokenParam.Amount, []byte{}, *tx.Hash())

			tokenExisted := statedb.PrivacyTokenIDExisted(testDB, *tx.GetTokenID())
			assert.Equal(t, true, tokenExisted)

			statedb.StoreCommitments(testDB, *tx.GetTokenID(), [][]byte{tokenOutputs[0].GetCommitment().ToBytesS()}, shardID)
			tokenID := tx.GetTokenID()

			numTokenInputs := common.RandIntInterval(1, 10)
			_, txTokenConversionParams, err := createTokenConversionParams(numTokenInputs, 1, 1, 2, tokenID)
			assert.Equal(t, nil, err, "createTokenConversionParams returns an error: %v", err)

			tx = new(TxTokenVersion2)
			err = InitTokenConversion(tx, txTokenConversionParams)
			assert.Equal(t, nil, err, "initTokenConversion returns an error: %v", err)

			isValidSanity, err := tx.ValidateSanityData(nil, nil, nil, 0)
			assert.Equal(t, true, isValidSanity)
			assert.Equal(t, nil, err)
			// validate signatures, proofs, etc. Only do after sanity checks are passed
			isValidTxItself, err := tx.ValidateTxByItself(false, testDB, nil, nil, shardID, false, nil, nil)
			assert.Equal(t, true, isValidTxItself)
			assert.Equal(t, nil, err)

			actualTxSize := getTxActualSizeInBytes(tx)
			fmt.Printf("TxSize: version=%d, type=%s, numInputs=%d, numOutputs=%d, hasPrivacy=%v, txSizeInBytes=%d\n", tx.GetVersion(), tx.GetType(), len(tx.GetProof().GetInputCoins()), len(tx.GetProof().GetOutputCoins()), tx.IsPrivacy(), actualTxSize)
			continue
			default://normal transaction
			txType = "n"
			hasPrivacy = (common.RandInt() % 2) == 0
		}

		tx, err := createSampleTx(numInputs, numOutputs, txType, txTokenType, hasPrivacy, keySet, version)
		assert.Equal(t, nil, err, "createSampleTx returns an error: %v", err)

		res, err := tx.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, nil, err, "ValidateSanityData returns an error: %v", err)
		assert.Equal(t, true, res)

		res, err = validateTransaction(tx, hasPrivacy, testDB, testDB, 0, &common.PRVCoinID, false, true)
		assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
		assert.Equal(t, true, res)

		actualTxSize := getTxActualSizeInBytes(tx)
		fmt.Printf("TxSize: version=%d, type=%s, numInputs=%d, numOutputs=%d, hasPrivacy=%v, txSizeInBytes=%d\n", tx.GetVersion(), tx.GetType(), len(tx.GetProof().GetInputCoins()), len(tx.GetProof().GetOutputCoins()), tx.IsPrivacy(), actualTxSize)
	}
}

