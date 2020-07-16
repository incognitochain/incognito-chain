package transaction

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
	"math"
	"strconv"
)

type TxTokenVersion2 struct {
	TxTokenBase
}

func checkIsBridgeTokenID(bridgeStateDB *statedb.StateDB, tokenID *common.Hash) error {
	isBridgeToken := false
	allBridgeTokensBytes, err := txDatabaseWrapper.getAllBridgeTokens(bridgeStateDB)
	if err != nil {
		return NewTransactionErr(TokenIDExistedError, err)
	}
	if len(allBridgeTokensBytes) > 0 {
		var allBridgeTokens []*rawdbv2.BridgeTokenInfo
		err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
		if err != nil {
			return NewTransactionErr(TokenIDExistedError, err)
		}
		for _, bridgeTokens := range allBridgeTokens {
			if tokenID.IsEqual(bridgeTokens.TokenID) {
				isBridgeToken = true
				break
			}
		}
	}
	if !isBridgeToken {
		return NewTransactionErr(TokenIDExistedError, errors.New("invalid Token ID"))
	}
	return nil
}

func (txToken *TxTokenVersion2) initToken(params *TxTokenParams) error {
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
			if err := txNormal.Init(txParams); err != nil {
				return NewTransactionErr(PrivacyTokenInitTokenDataError, err)
			}
			txToken.TxTokenData.TxNormal = txNormal
		}
	}
	if !handled {
		return NewTransactionErr(PrivacyTokenTxTypeNotHandleError, errors.New("can't handle this TokenTxType"))
	}
	return nil
}

// Prove
func (tx *TxVersion2) proveWithMessage(params *TxPrivacyInitParams, hashedTokenMessage []byte) error {
	outputCoins, err := newCoinV2ArrayFromPaymentInfoArray(params.paymentInfo, params.tokenID, params.stateDB)
	if err != nil {
		Logger.Log.Errorf("Cannot parse outputCoinV2 to outputCoins, error %v ", err)
		return err
	}

	// inputCoins is plainCoin because it may have coinV1 with coinV2
	inputCoins := params.inputCoins

	tx.Proof, err = privacy_v2.Prove(inputCoins, outputCoins, params.hasPrivacy, params.paymentInfo)
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

	// Get Hash of the whole txToken then sign on it
	message := common.HashH(append(tx.Hash()[:], hashedTokenMessage...))
	err = tx.signOnMessage(inputCoins, outputCoins, params, message[:])
	return err
}

func (txToken *TxTokenVersion2) initPRVFee(feeTx * TxVersion2, params *TxPrivacyInitParams) error {
	txTokenDataHash, err := txToken.TxTokenData.Hash()
	if err != nil {
		Logger.Log.Errorf("Cannot calculate txPrivacyTokenData Hash, err %v", err)
		return err
	}
	if err := feeTx.proveWithMessage(params, txTokenDataHash[:]); err != nil {
		return NewTransactionErr(PrivacyTokenInitPRVError, err)
	}
	// override TxCustomTokenPrivacyType type
	feeTx.SetType(common.TxCustomTokenPrivacyType)
	txToken.Tx = feeTx

	return nil
}

func (txToken *TxTokenVersion2) Init(paramsInterface interface{}) error {
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
	if err := txToken.initToken(params); err != nil {
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

func (txToken *TxTokenVersion2) InitTxTokenSalary(otaCoin *coin.CoinV2, privKey *privacy.PrivateKey, stateDB *statedb.StateDB, metaData metadata.Metadata, coinID *common.Hash, coinName string) error {
	var err error
	// Check validate params
	txPrivacyParams := NewTxPrivacyInitParams(
		privKey, []*privacy.PaymentInfo{}, nil, 0, false, stateDB, nil, metaData, nil,
	)
	if err := validateTxParams(txPrivacyParams); err != nil {
		return err
	}
	// check tx size
	publicKeyBytes := otaCoin.GetPublicKey().ToBytesS()
	if txSize := estimateTxSizeOfInitTokenSalary(publicKeyBytes, otaCoin.GetValue(), coinName, coinID); txSize > common.MaxTxSize {
		return NewTransactionErr(ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}

	// Create TxToken
	var propertyID [common.HashSize]byte
	copy(propertyID[:], coinID[:])
	txToken.TxTokenData.PropertyID = propertyID
	txToken.TxTokenData.SetType(CustomTokenInit)
	txToken.TxTokenData.SetPropertyName(coinName)
	txToken.TxTokenData.SetPropertySymbol(coinName)
	txToken.TxTokenData.SetAmount(otaCoin.GetValue())
	txToken.TxTokenData.SetMintable( true)

	tempOutputCoin := []coin.Coin{otaCoin}
	proof := new(privacy.ProofV2)
	proof.Init()
	if err = proof.SetOutputCoins(tempOutputCoin); err != nil {
		Logger.Log.Errorf("Init customPrivacyToken cannot set outputCoins")
		return err
	}
	temp := new(TxVersion2)
	temp.Version = TxVersion2Number
	temp.Type = common.TxNormalType
	temp.Proof = proof
	temp.PubKeyLastByteSender = publicKeyBytes[len(publicKeyBytes)-1]
	// signOnMessage Tx
	//temp.sigPrivKey = *privKey
	if temp.Sig, temp.SigPubKey, err = signNoPrivacy(privKey, temp.Hash()[:]); err != nil {
		Logger.Log.Error(errors.New("can't signOnMessage this tx"))
		return NewTransactionErr(SignTxError, err)
	}
	temp.SigPubKey = otaCoin.GetPublicKey().ToBytesS()
	txToken.TxTokenData.TxNormal = temp

	// Init tx fee params
	tx := new(TxVersion2)
	if err := tx.initializeTxAndParams(txPrivacyParams); err != nil {
		return err
	}
	tx.SetType(common.TxCustomTokenPrivacyType)
	tx.sigPrivKey = *txPrivacyParams.senderSK

	hashedTokenMessage := txToken.TxTokenData.TxNormal.Hash()

	message := common.HashH(append(tx.Hash()[:], hashedTokenMessage[:]...))
	if tx.Sig, tx.SigPubKey, err = signNoPrivacy(privKey, message[:]); err != nil {
		Logger.Log.Error(errors.New(fmt.Sprintf("Cannot signOnMessage tx %v\n", err)))
		return NewTransactionErr(SignTxError, err)
	}

	txToken.SetTxBase(tx)
	return nil
}

func (txToken *TxTokenVersion2) verifySig(transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash) (bool, error) {
	// check input transaction
	txFee := txToken.Tx
	if txFee.GetSig() == nil || txFee.GetSigPubKey() == nil {
		return false, NewTransactionErr(UnexpectedError, errors.New("input transaction must be a signed one"))
	}
	var err error

	// Verify TxToken Salary: NonPrivacyNonInput
	if txFee.GetProof() == nil {
		hashedTokenMessage := txToken.TxTokenData.TxNormal.Hash()
		message := common.HashH(append(txFee.Hash()[:], hashedTokenMessage[:]...))
		if valid, err := verifySigNoPrivacy(txFee.GetSig(), txFee.GetSigPubKey(), message[:]); !valid {
			if err != nil {
				Logger.Log.Debugf("Error verifying signature of tx: %+v", err)
				return false, NewTransactionErr(VerifyTxSigFailError, err)
			}
			return false, nil
		}
		return true, nil
	}

	// Reform Ring
	sumOutputCoinsWithFee := calculateSumOutputsWithFee(txFee.GetProof().GetOutputCoins(), txFee.GetTxFee())
	ring, err := getRingFromSigPubKeyAndLastColumnCommitment(
		txFee.GetSigPubKey(), sumOutputCoinsWithFee,
		transactionStateDB, shardID, tokenID,
	)
	if err != nil {
		Logger.Log.Errorf("Error when querying database to construct mlsag ring: %v ", err)
		return false, err
	}

	// Reform MLSAG Signature
	inputCoins := txFee.GetProof().GetInputCoins()
	keyImages := make([]*operation.Point, len(inputCoins)+1)
	for i := 0; i < len(inputCoins); i += 1 {
		keyImages[i] = inputCoins[i].GetKeyImage()
	}
	// The last column is gone, so just fill in any value
	keyImages[len(inputCoins)] = operation.RandomPoint()
	mlsagSignature, err := getMLSAGSigFromTxSigAndKeyImages(txFee.GetSig(), keyImages)
	if err != nil {
		Logger.Log.Errorf("Error when reconstructing mlsagSignature: %v ", err)
		return false, err
	}

	txTokenDataHash, err := txToken.TxTokenData.Hash()
	if err != nil {
		Logger.Log.Errorf("Error when getting txTokenData Hash: %v ", err)
		return false, err

	}
	message := common.HashH(append(txFee.Hash()[:], txTokenDataHash[:]...))
	return mlsag.Verify(mlsagSignature, ring, message[:])
}

func (txToken TxTokenVersion2) ValidateTxByItself(hasPrivacyCoin bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, chainRetriever metadata.ChainRetriever, shardID byte, isNewTransaction bool, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever) (bool, error) {
	// check for proof, signature ...
	if ok, err := txToken.ValidateTransaction(hasPrivacyCoin, transactionStateDB, bridgeStateDB, shardID, nil, false, isNewTransaction); !ok {
		return false, err
	}
	meta := txToken.GetMetadata()
	if meta != nil {
		validateMetadata := meta.ValidateMetadataByItself()
		if !validateMetadata {
			return validateMetadata, NewTransactionErr(UnexpectedError, errors.New("Metadata is invalid"))
		}
		return validateMetadata, nil
	}
	return true, nil
}

func (txToken TxTokenVersion2) ValidateTransaction(hasPrivacyCoin bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	var err error
	if tokenID, err = parseTokenID(tokenID); err != nil {
		return false, err
	}
	ok, err := txToken.verifySig(transactionStateDB, shardID, tokenID)
	if !ok {
		Logger.Log.Errorf("FAILED VERIFICATION SIGNATURE ver2 (token) with tx hash %s: %+v \n", txToken.Hash().String(), err)
		return false, NewTransactionErr(VerifyTxSigFailError, err)
	}else {
		// validate for pToken
		tokenID := txToken.TxTokenData.PropertyID
		switch txToken.TxTokenData.Type {
		case CustomTokenInit:
			if txToken.TxTokenData.Mintable {
				return true, nil
			} else {
				// check exist token
				if txDatabaseWrapper.privacyTokenIDExisted(transactionStateDB, tokenID) {
					return false, errors.New("Cannot validate Tx Init Token. It is tx mint from User")
				}
				return true, nil
			}
		case CustomTokenTransfer:
			if txToken.GetType() == common.TxTokenConversionType {
				return validateConversionVer1ToVer2(txToken.TxTokenData.TxNormal, transactionStateDB, shardID, &tokenID)
			} else {
				return txToken.TxTokenData.TxNormal.ValidateTransaction(
					txToken.TxTokenData.TxNormal.IsPrivacy(),
					transactionStateDB, bridgeStateDB, shardID, &tokenID, isBatch, isNewTransaction)
			}
		default:
			return false, errors.New("Cannot validate Tx Token. Unavailable type")
		}
	}
}

func (txToken TxTokenVersion2) ValidateSanityData(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) (bool, error) {
	if txToken.GetTxBase().GetProof() == nil && txToken.TxTokenData.TxNormal.GetProof() == nil {
		return false, errors.New("Tx Privacy Ver 2 must have a proof")
	}
	// validate metadata
	check, err := validateSanityMetadata(&txToken, chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if !check || err != nil {
		return false, NewTransactionErr(InvalidSanityDataPrivacyTokenError, err)
	}
	// validate sanity for tx pToken + metadata
	check, err = validateSanityTxWithoutMetadata(txToken.TxTokenData.TxNormal, chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if !check || err != nil {
		return false, NewTransactionErr(InvalidSanityDataPrivacyTokenError, err)
	}
	// validate sanity for tx pToken + without metadata
	check1, err1 := validateSanityTxWithoutMetadata(txToken.Tx, chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if !check1 || err1 != nil {
		return false, NewTransactionErr(InvalidSanityDataPrivacyTokenError, err1)
	}
	return true, nil
}

func (txToken TxTokenVersion2) GetTxActualSize() uint64 {
	sizeTx := getTxActualSizeInBytes(txToken.Tx)

	if &txToken.TxTokenData != nil {
		sizeTx += getTxActualSizeInBytes(txToken.TxTokenData.TxNormal)
		sizeTx += uint64(len(txToken.TxTokenData.PropertyName))
		sizeTx += uint64(len(txToken.TxTokenData.PropertySymbol))
		sizeTx += uint64(len(txToken.TxTokenData.PropertyID))
		sizeTx += 4 // Type
		sizeTx += 1 // Mintable
		sizeTx += 8 // Amount
	}
	meta := txToken.GetMetadata()
	fmt.Println(meta)
	if meta != nil {
		sizeTx += meta.CalculateSize()
	}

	result := uint64(math.Ceil(float64(sizeTx) / 1024))
	return result
}