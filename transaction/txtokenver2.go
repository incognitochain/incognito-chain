package transaction

import (
	"errors"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
	"strconv"
)

type TxTokenVersion2 struct {
	TxTokenBase
}

func (txToken *TxTokenVersion2) initToken(params *TxPrivacyTokenInitParams) error {
	txToken.TxPrivacyTokenData.SetType(params.tokenParams.TokenTxType)
	txToken.TxPrivacyTokenData.SetPropertyName(params.tokenParams.PropertyName)
	txToken.TxPrivacyTokenData.SetPropertySymbol(params.tokenParams.PropertySymbol)
	txToken.TxPrivacyTokenData.SetMintable(params.tokenParams.Mintable)

	var handled = false
	switch params.tokenParams.TokenTxType {
	case CustomTokenInit:
		// case init a new privacy custom token
		{
			handled = true
			txToken.TxPrivacyTokenData.SetAmount(params.tokenParams.Amount)

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
			tempOutputCoin := make([]coin.Coin, 1)
			tempOutputCoin[0] = c

			temp := new(TxVersion2)
			temp.SetVersion(txVersion2Number)
			temp.SetType(common.TxNormalType)
			temp.Proof = new(privacy.ProofV2)
			temp.Proof.Init()
			if err := temp.Proof.SetOutputCoins(tempOutputCoin); err != nil {
				Logger.Log.Errorf("Init customPrivacyToken cannot set outputCoins")
				return err
			}

			var err error
			if temp.PubKeyLastByteSender, err = params.inputCoin[0].GetShardID(); err != nil {
				return NewTransactionErr(GetShardIDByPublicKeyError, err)
			}
			temp.sigPrivKey = *params.senderKey

			temp.Sig, temp.SigPubKey, err = signNoPrivacy(params.senderKey, temp.Hash()[:])
			if err != nil {
				Logger.Log.Error(errors.New("can't signOnMessage this tx"))
				return NewTransactionErr(SignTxError, err)
			}
			temp.SigPubKey = params.tokenParams.Receiver[0].PaymentAddress.Pk
			txToken.TxPrivacyTokenData.TxNormal = *NewTxBaseFromTransaction(temp)

			hashInitToken, err := txToken.TxPrivacyTokenData.Hash()
			if err != nil {
				Logger.Log.Error(errors.New("can't hash this token data"))
				return NewTransactionErr(UnexpectedError, err)
			}
			if params.tokenParams.Mintable {
				propertyID, err := common.Hash{}.NewHashFromStr(params.tokenParams.PropertyID)
				if err != nil {
					return NewTransactionErr(TokenIDInvalidError, err, propertyID.String())
				}
				txToken.TxPrivacyTokenData.PropertyID = *propertyID
			} else {
				//NOTICE: @merman update PropertyID calculated from hash of tokendata and shardID
				newHashInitToken := common.HashH(append(hashInitToken.GetBytes(), params.shardID))
				Logger.Log.Debug("New Privacy Token %+v ", newHashInitToken)
				existed := txDatabaseWrapper.privacyTokenIDExisted(params.transactionStateDB, newHashInitToken)
				if existed {
					Logger.Log.Error("INIT Tx Custom Token Privacy is Existed", newHashInitToken)
					return NewTransactionErr(TokenIDExistedError, errors.New("this token is existed in network"))
				}
				txToken.TxPrivacyTokenData.SetPropertyID(newHashInitToken)
				Logger.Log.Debugf("A new token privacy wil be issued with ID: %+v", txToken.TxPrivacyTokenData.PropertyID.String())
			}
		}
	case CustomTokenTransfer:
		{
			handled = true
			// make a transfering for privacy custom token
			// fee always 0 and reuse function of normal tx for custom token ID
			propertyID, _ := common.Hash{}.NewHashFromStr(params.tokenParams.PropertyID)
			existed := txDatabaseWrapper.privacyTokenIDExisted(params.transactionStateDB, *propertyID)
			if !existed {
				isBridgeToken := false
				allBridgeTokensBytes, err := txDatabaseWrapper.getAllBridgeTokens(params.bridgeStateDB)
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
						if propertyID.IsEqual(bridgeTokens.TokenID) {
							isBridgeToken = true
							break
						}
					}
				}
				if !isBridgeToken {
					return NewTransactionErr(TokenIDExistedError, errors.New("invalid Token ID"))
				}
			}
			Logger.Log.Debugf("Token %+v wil be transfered with", propertyID)

			txToken.TxPrivacyTokenData.SetPropertyID(*propertyID)
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
			txNormal := new(TxBase)
			if err := txNormal.Init(txParams); err != nil {
				return NewTransactionErr(PrivacyTokenInitTokenDataError, err)
			}
			txToken.TxPrivacyTokenData.TxNormal = *NewTxBaseFromTransaction(txNormal)
		}
	}
	if !handled {
		return NewTransactionErr(PrivacyTokenTxTypeNotHandleError, errors.New("can't handle this TokenTxType"))
	}
	return nil
}

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

// Special kind of function that init and signOnMessage onto the message, not on txHash
func (tx *TxVersion2) initWithMessage(paramsInterface interface{}, hashedTokenMessage []byte) error {
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

	if err := tx.proveWithMessage(params, hashedTokenMessage); err != nil {
		return err
	}
	return nil
}

func (txToken *TxTokenVersion2) initPRVFee(params *TxPrivacyTokenInitParams) error {
	// init data for tx PRV for fee
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
	feeTx := new(TxVersion2)
	txTokenDataHash, err := txToken.TxPrivacyTokenData.Hash()
	if err != nil {
		Logger.Log.Errorf("Cannot calculate txPrivacyTokenData Hash, err %v", err)
		return err
	}
	if err := feeTx.initWithMessage(txPrivacyParams, txTokenDataHash[:]); err != nil {
		return NewTransactionErr(PrivacyTokenInitPRVError, err)
	}
	// override TxCustomTokenPrivacyType type
	feeTx.SetType(common.TxCustomTokenPrivacyType)
	txToken.TxBase = *NewTxBaseFromTransaction(feeTx)

	return nil
}

func (txToken *TxTokenVersion2) Init(paramsInterface interface{}) error {
	params, ok := paramsInterface.(*TxPrivacyTokenInitParams)
	if !ok {
		return errors.New("Cannot init TxCustomTokenPrivacy because params is not correct")
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
	if err := txToken.initPRVFee(params); err != nil {
		Logger.Log.Errorf("Cannot init token ver2: err %v", err)
		return err
	}

	return nil
}

func (txToken *TxTokenVersion2) InitTxTokenSalary(otaCoin *coin.CoinV2, privKey *privacy.PrivateKey, stateDB *statedb.StateDB, metaData metadata.Metadata, coinID *common.Hash, coinName string) error {
	var err error
	// init data for tx PRV for fee
	txPrivacyParams := NewTxPrivacyInitParams(
		privKey, []*privacy.PaymentInfo{}, nil, 0, false, stateDB, nil, metaData, nil,
	)
	if err := txToken.TxBase.Init(txPrivacyParams); err != nil {
		return NewTransactionErr(PrivacyTokenInitPRVError, err)
	}
	// override TxCustomTokenPrivacyType type
	txToken.TxBase.SetType(common.TxCustomTokenPrivacyType)

	// check tx size
	publicKeyBytes := otaCoin.GetPublicKey().ToBytesS()
	if txSize := estimateTxSizeOfInitTokenSalary(publicKeyBytes, otaCoin.GetValue(), coinName, coinID); txSize > common.MaxTxSize {
		return NewTransactionErr(ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}
	// check action type and create privacy custom toke data
	var propertyID [common.HashSize]byte
	copy(propertyID[:], coinID[:])
	txToken.TxPrivacyTokenData.PropertyID = propertyID
	txToken.TxPrivacyTokenData.SetType(CustomTokenInit)
	txToken.TxPrivacyTokenData.SetPropertyName(coinName)
	txToken.TxPrivacyTokenData.SetPropertySymbol(coinName)
	txToken.TxPrivacyTokenData.SetAmount(otaCoin.GetValue())
	txToken.TxPrivacyTokenData.SetMintable( true)

	tempOutputCoin := []coin.Coin{otaCoin}
	proof := new(privacy.ProofV2)
	proof.Init()
	if err = proof.SetOutputCoins(tempOutputCoin); err != nil {
		Logger.Log.Errorf("Init customPrivacyToken cannot set outputCoins")
		return err
	}
	temp := TxVersion2{}
	temp.Version = txVersion2Number
	temp.Type = common.TxNormalType
	temp.Proof = proof
	temp.PubKeyLastByteSender = publicKeyBytes[len(publicKeyBytes)-1]
	// signOnMessage Tx
	temp.sigPrivKey = *privKey
	if temp.Sig, _, err = signNoPrivacy(privKey, temp.Hash()[:]); err != nil {
		Logger.Log.Error(errors.New("can't signOnMessage this tx"))
		return NewTransactionErr(SignTxError, err)
	}
	temp.SigPubKey = otaCoin.GetPublicKey().ToBytesS()

	txToken.TxPrivacyTokenData.TxNormal = *NewTxBaseFromTransaction(&temp)
	return nil
}

func (txToken *TxTokenVersion2) verifySig(transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash) (bool, error) {
	// check input transaction
	if txToken.TxBase.Sig == nil || txToken.TxBase.SigPubKey == nil {
		return false, NewTransactionErr(UnexpectedError, errors.New("input transaction must be a signed one"))
	}
	var err error

	// Reform Ring
	sumOutputCoinsWithFee := calculateSumOutputsWithFee(txToken.TxBase.Proof.GetOutputCoins(), txToken.TxBase.Fee)
	ring, err := getRingFromSigPubKeyAndLastColumnCommitment(
		txToken.TxBase.SigPubKey, sumOutputCoinsWithFee,
		transactionStateDB, shardID, tokenID,
	)
	if err != nil {
		Logger.Log.Errorf("Error when querying database to construct mlsag ring: %v ", err)
		return false, err
	}

	// Reform MLSAG Signature
	inputCoins := txToken.TxBase.Proof.GetInputCoins()
	keyImages := make([]*operation.Point, len(inputCoins)+1)
	for i := 0; i < len(inputCoins); i += 1 {
		keyImages[i] = inputCoins[i].GetKeyImage()
	}
	// The last column is gone, so just fill in any value
	keyImages[len(inputCoins)] = operation.RandomPoint()
	mlsagSignature, err := getMLSAGSigFromTxSigAndKeyImages(txToken.TxBase.Sig, keyImages)
	if err != nil {
		Logger.Log.Errorf("Error when reconstructing mlsagSignature: %v ", err)
		return false, err
	}

	txTokenDataHash, err := txToken.TxPrivacyTokenData.Hash()
	if err != nil {
		Logger.Log.Errorf("Error when getting txTokenData Hash: %v ", err)
		return false, err

	}
	message := common.HashH(append(txToken.TxBase.Hash()[:], txTokenDataHash[:]...))
	return mlsag.Verify(mlsagSignature, ring, message[:])
}

func (txToken TxTokenVersion2) ValidateTransaction(hasPrivacyCoin bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	var err error
	if tokenID, err = parseTokenID(tokenID); err != nil {
		return false, err
	}
	ok, err := txToken.verifySig(transactionStateDB, shardID, tokenID)
	if ok {
		// validate for pToken
		tokenID := txToken.TxPrivacyTokenData.PropertyID
		if txToken.TxPrivacyTokenData.Type == CustomTokenInit {
			if txToken.Type == common.TxRewardType && txToken.TxPrivacyTokenData.Mintable {
				isBridgeCentralizedToken, _ := txDatabaseWrapper.isBridgeTokenExistedByType(bridgeStateDB, tokenID, true)
				isBridgeDecentralizedToken, _ := txDatabaseWrapper.isBridgeTokenExistedByType(bridgeStateDB, tokenID, false)
				if isBridgeCentralizedToken || isBridgeDecentralizedToken {
					return true, nil
				}
				return false, nil
			} else {
				// check exist token
				if txDatabaseWrapper.privacyTokenIDExisted(transactionStateDB, tokenID) {
					return false, nil
				}
				return true, nil
			}
		} else {
			if err != nil {
				Logger.Log.Errorf("Cannot create txPrivacyFromVersionNumber from TxPrivacyTokenDataVersion1, err %v", err)
				return false, err
			}
			return txToken.TxPrivacyTokenData.TxNormal.ValidateTransaction(
				txToken.TxPrivacyTokenData.TxNormal.IsPrivacy(),
				transactionStateDB, bridgeStateDB, shardID, &tokenID, isBatch, isNewTransaction)
		}
	}
	return false, err
}