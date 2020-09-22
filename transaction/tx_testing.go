package transaction

import (
	"fmt"
	"errors"
	"strconv"

	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy"
)

func TestResignTxV1(txv1_generic metadata.Transaction){
	txv1, ok := txv1_generic.(*TxVersion1)
	if !ok{
		panic("Error when casting")
	}
	txv1.cachedHash = nil
	txv1.SetSig(nil)
	txv1.SetSigPubKey(nil)
	err := txv1.sign()
	if err!=nil{
		// if it fails, something's wrong
		fmt.Printf("Error when resigning : %v\n",err)
	}
}

func TestResignTxV1WithKey(txv1_generic metadata.Transaction, sk []byte){
	txv1, ok := txv1_generic.(*TxVersion1)
	if !ok{
		panic("Error when casting")
	}
	txv1.sigPrivKey = sk
	txv1.cachedHash = nil
	txv1.SetSig(nil)
	txv1.SetSigPubKey(nil)
	err := txv1.sign()
	if err!=nil{
		// if it fails, something's wrong
		fmt.Printf("Error when resigning : %v\n",err)
	}
}

func (tx *TxVersion2) InitTestOldOTA(paramsInterface interface{}) error {
	params, ok := paramsInterface.(*TxPrivacyInitParams)
	if !ok {
		return errors.New("params of tx Init is not TxPrivacyInitParam")
	}

	Logger.Log.Debugf("CREATING TX........\n")
	if err := validateTxParams(params); err != nil {
		return err
	}

	// Init tx and params (tx and params will be changed)
	if err := tx.initializeTxAndParams(params); err != nil {
		return err
	}

	// Check if this tx is nonPrivacyNonInput
	// Case 1: tx ptoken transfer with ptoken fee
	// Case 2: tx Reward
	// If it is non privacy non input then return
	if check, err := tx.isNonPrivacyNonInput(params); check {
		return err
	}

	outputCoins, err := newCoinV2ArrayFromPaymentInfoArray(params.paymentInfo, params.tokenID, params.stateDB)
	if err != nil {
		Logger.Log.Errorf("Cannot parse outputCoinV2 to outputCoins, error %v ", err)
		return err
	}

	// inputCoins is plainCoin because it may have coinV1 with coinV2
	inputCoins := params.inputCoins
	clonedPK := &operation.Point{}
	clonedPK.FromBytes(inputCoins[0].GetPublicKey().ToBytes())
	outputCoins[0].SetPublicKey(clonedPK)

	tx.Proof, err = privacy_v2.Prove(inputCoins, outputCoins, nil, params.hasPrivacy, params.paymentInfo)
	if err != nil {
		Logger.Log.Errorf("Error in privacy_v2.Prove, error %v ", err)
		return err
	}

	if tx.ShouldSignMetaData() {
		if err := tx.signMetadata(params.senderSK); err != nil {
			Logger.Log.Error("Cannot signOnMessage txMetadata in shouldSignMetadata")
			return err
		}
	}
	err = tx.signOnMessage(inputCoins, outputCoins, params, tx.Hash()[:])
	return err
	// return nil
}

func (txToken *TxTokenVersion2) InitTestOldOTAToken(paramsInterface interface{}) error {
	params, ok := paramsInterface.(*TxTokenParams)
	if !ok {
		return errors.New("Cannot init TxCustomTokenPrivacy because params is not correct")
	}

	// Check validate params first, before creating tx token
	// Because there are some validation must be made first
	// Please dont change their order when you dont really understand
	txPrivacyParams := NewTxPrivacyInitParams(
		params.senderKey,
		params.paymentInfo,
		params.inputCoin,
		params.feeNativeCoin,
		params.hasPrivacyCoin,
		params.transactionStateDB,
		nil,
		params.metaData,
		params.info,
	)
	if err := validateTxParams(txPrivacyParams); err != nil {
		return err
	}
	// Init tx and params (tx and params will be changed)
	tx := new(TxVersion2)
	if err := tx.initializeTxAndParams(txPrivacyParams); err != nil {
		return err
	}

	// Check if this tx is nonPrivacyNonInput
	// Case 1: tx ptoken transfer with ptoken fee
	// Case 2: tx Reward
	// If it is non privacy non input then return
	if check, err := tx.isNonPrivacyNonInput(txPrivacyParams); check {
		return err
	}

	// check tx size
	limitFee := uint64(0)
	estimateTxSizeParam := NewEstimateTxSizeParam(len(params.inputCoin), len(params.paymentInfo),
		params.hasPrivacyCoin, nil, params.tokenParams, limitFee)
	if txSize := EstimateTxSize(estimateTxSizeParam); txSize > common.MaxTxSize {
		return NewTransactionErr(ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}

	// Init Token first
	if err := txToken.initTokenOldOTA(params); err != nil {
		Logger.Log.Errorf("Cannot init token ver2: err %v", err)
		return err
	}

	// Init PRV Fee on the whole transaction
	if err := txToken.initPRVFee(tx, txPrivacyParams); err != nil {
		Logger.Log.Errorf("Cannot init token ver2: err %v", err)
		return err
	}

	return nil
}

func (txToken *TxTokenVersion2) initTokenOldOTA(params *TxTokenParams) error {
	txToken.TxTokenData.SetType(params.tokenParams.TokenTxType)
	txToken.TxTokenData.SetPropertyName(params.tokenParams.PropertyName)
	txToken.TxTokenData.SetPropertySymbol(params.tokenParams.PropertySymbol)
	txToken.TxTokenData.SetMintable(params.tokenParams.Mintable)

	var handled = false
	switch params.tokenParams.TokenTxType {
	case CustomTokenInit:
		// case init a new privacy custom token
		{
			handled = true
			txToken.TxTokenData.SetAmount(params.tokenParams.Amount)

			message := []byte{}
			if len(params.tokenParams.Receiver[0].Message) > 0 {
				if len(params.tokenParams.Receiver[0].Message) > coin.MaxSizeInfoCoin {
					return NewTransactionErr(ExceedSizeInfoOutCoinError, nil)
				}
				message = params.tokenParams.Receiver[0].Message
			}
			tempPaymentInfo := key.InitPaymentInfo(params.tokenParams.Receiver[0].PaymentAddress, params.tokenParams.Amount, message)
			c, errCoin := coin.NewCoinFromPaymentInfo(tempPaymentInfo)
			if errCoin != nil {
				Logger.Log.Errorf("Cannot create new coin based on payment info err %v", errCoin)
				return errCoin
			}

			temp := new(TxVersion2)
			temp.SetVersion(TxVersion2Number)
			temp.SetType(common.TxNormalType)
			temp.Proof = new(privacy.ProofV2)
			temp.Proof.Init()
			if err := temp.Proof.SetOutputCoins([]coin.Coin{c}); err != nil {
				Logger.Log.Errorf("Init customPrivacyToken cannot set outputCoins")
				return err
			}
			var err error
			if temp.PubKeyLastByteSender, err = params.inputCoin[0].GetShardID(); err != nil {
				return NewTransactionErr(GetShardIDByPublicKeyError, err)
			}
			//temp.sigPrivKey = *params.senderKey

			temp.Sig, _, err = signNoPrivacy(params.senderKey, temp.Hash()[:])
			if err != nil {
				Logger.Log.Error(errors.New("can't signOnMessage this tx"))
				return NewTransactionErr(SignTxError, err)
			}
			temp.SigPubKey = params.tokenParams.Receiver[0].PaymentAddress.Pk
			txToken.TxTokenData.TxNormal = temp

			hashInitToken, err := txToken.TxTokenData.Hash()
			if err != nil {
				Logger.Log.Error(errors.New("can't hash this token data"))
				return NewTransactionErr(UnexpectedError, err)
			}
			if params.tokenParams.Mintable {
				propertyID, err := common.Hash{}.NewHashFromStr(params.tokenParams.PropertyID)
				if err != nil {
					return NewTransactionErr(TokenIDInvalidError, err, propertyID.String())
				}
				txToken.TxTokenData.PropertyID = *propertyID
			} else {
				//NOTICE: @merman update PropertyID calculated from hash of tokendata and shardID
				newHashInitToken := common.HashH(append(hashInitToken.GetBytes(), params.shardID))
				Logger.Log.Debug("New Privacy Token %+v ", newHashInitToken)
				existed := txDatabaseWrapper.privacyTokenIDExisted(params.transactionStateDB, newHashInitToken)
				if existed {
					Logger.Log.Error("INIT Tx Custom Token Privacy is Existed", newHashInitToken)
					return NewTransactionErr(TokenIDExistedError, errors.New("this token is existed in network"))
				}
				txToken.TxTokenData.SetPropertyID(newHashInitToken)
				Logger.Log.Debugf("A new token privacy wil be issued with ID: %+v", txToken.TxTokenData.PropertyID.String())
			}
		}
	case CustomTokenTransfer:
		{
			handled = true
			// make a transfering for privacy custom token
			// fee always 0 and reuse function of normal tx for custom token ID
			propertyID, _ := common.TokenStringToHash(params.tokenParams.PropertyID)
			existed := txDatabaseWrapper.privacyTokenIDExisted(params.transactionStateDB, *propertyID)
			if !existed {
				if err := checkIsBridgeTokenID(params.bridgeStateDB, propertyID); err != nil {
					return err
				}
			}
			Logger.Log.Debugf("Token %+v wil be transfered with", propertyID)

			txToken.TxTokenData.SetPropertyID(*propertyID)
			txParams := NewTxPrivacyInitParams(
				params.senderKey,
				params.tokenParams.Receiver,
				params.tokenParams.TokenInput,
				params.tokenParams.Fee,
				params.hasPrivacyToken,
				params.transactionStateDB,
				propertyID,
				nil,
				nil,
			)
			txNormal := new(TxVersion2)
			if err := txNormal.InitTestOldOTA(txParams); err != nil {
				return NewTransactionErr(PrivacyTokenInitTokenDataError, err)
			}
			txNormal.SetType(common.TxCustomTokenPrivacyType)
			txToken.TxTokenData.TxNormal = txNormal
		}
	}
	if !handled {
		return NewTransactionErr(PrivacyTokenTxTypeNotHandleError, errors.New("can't handle this TokenTxType"))
	}
	return nil
}