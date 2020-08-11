package transaction

import (
	"encoding/json"
	"errors"
	"math"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	zkp "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge"
	"strconv"
)

type TxTokenVersion1 struct {
	TxTokenBase
}

func (txToken *TxTokenVersion1) Init(paramsInterface interface{}) error {
	params, ok := paramsInterface.(*TxTokenParams)
	if !ok {
		return errors.New("Cannot init TxTokenBase because params is not correct")
	}
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
	txToken.Tx = new(TxVersion1)
	if err := txToken.Tx.Init(txPrivacyParams); err != nil {
		return NewTransactionErr(PrivacyTokenInitPRVError, err)
	}
	// override TxCustomTokenPrivacyType type
	txToken.Tx.SetType(common.TxCustomTokenPrivacyType)

	// check tx size
	limitFee := uint64(0)
	estimateTxSizeParam := NewEstimateTxSizeParam(len(params.inputCoin), len(params.paymentInfo),
		params.hasPrivacyCoin, nil, params.tokenParams, limitFee)
	if txSize := EstimateTxSize(estimateTxSizeParam); txSize > common.MaxTxSize {
		return NewTransactionErr(ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}

	// check action type and create privacy custom toke data
	var handled = false
	// Add token data component
	txToken.TxTokenData.SetType(params.tokenParams.TokenTxType)
	txToken.TxTokenData.SetPropertyName(params.tokenParams.PropertyName)
	txToken.TxTokenData.SetPropertySymbol(params.tokenParams.PropertySymbol)

	switch params.tokenParams.TokenTxType {
		case CustomTokenInit: {
			// case init a new privacy custom token
			handled = true
			txToken.TxTokenData.SetAmount(params.tokenParams.Amount)

			temp := new(TxVersion1)
			temp.SetVersion(TxVersion1Number)
			temp.Type = common.TxNormalType
			temp.Proof = new(zkp.PaymentProof)
			tempOutputCoin := make([]*coin.CoinV1, 1)
			tempOutputCoin[0] = new(coin.CoinV1)
			tempOutputCoin[0].CoinDetails = new(coin.PlainCoinV1)
			tempOutputCoin[0].CoinDetails.SetValue(params.tokenParams.Amount)
			PK, err := new(operation.Point).FromBytesS(params.tokenParams.Receiver[0].PaymentAddress.Pk)
			if err != nil {
				return NewTransactionErr(DecompressPaymentAddressError, err)
			}
			tempOutputCoin[0].CoinDetails.SetPublicKey(PK)
			tempOutputCoin[0].CoinDetails.SetRandomness(operation.RandomScalar())

			// set info coin for output coin
			if len(params.tokenParams.Receiver[0].Message) > 0 {
				if len(params.tokenParams.Receiver[0].Message) > privacy.MaxSizeInfoCoin {
					return NewTransactionErr(ExceedSizeInfoOutCoinError, nil)
				}
				tempOutputCoin[0].CoinDetails.SetInfo(params.tokenParams.Receiver[0].Message)
			}
			tempOutputCoin[0].CoinDetails.SetSNDerivator(privacy.RandomScalar())
			err = tempOutputCoin[0].CoinDetails.CommitAll()
			if err != nil {
				return NewTransactionErr(CommitOutputCoinError, err)
			}
			temp.Proof.SetOutputCoins(coin.ArrayCoinV1ToCoin(tempOutputCoin))

			// get last byte
			temp.PubKeyLastByteSender = params.tokenParams.Receiver[0].PaymentAddress.Pk[len(params.tokenParams.Receiver[0].PaymentAddress.Pk)-1]

			// signOnMessage Tx
			temp.SigPubKey = params.tokenParams.Receiver[0].PaymentAddress.Pk
			temp.sigPrivKey = *params.senderKey
			err = temp.sign()
			if err != nil {
				Logger.Log.Error(errors.New("can't signOnMessage this tx"))
				return NewTransactionErr(SignTxError, err)
			}
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
				txToken.TxTokenData.Mintable = true
			} else {
				//NOTICE: @merman update PropertyID calculated from hash of tokendata and shardID
				newHashInitToken := common.HashH(append(hashInitToken.GetBytes(), params.shardID))
				Logger.Log.Debug("New Privacy Token %+v ", newHashInitToken)
				existed := statedb.PrivacyTokenIDExisted(params.transactionStateDB, newHashInitToken)
				if existed {
					Logger.Log.Error("INIT Tx Custom Token Privacy is Existed", newHashInitToken)
					return NewTransactionErr(TokenIDExistedError, errors.New("this token is existed in network"))
				}
				txToken.TxTokenData.PropertyID = newHashInitToken
				Logger.Log.Debugf("A new token privacy wil be issued with ID: %+v", txToken.TxTokenData.PropertyID.String())
			}
		}
		case CustomTokenTransfer: {
			handled = true
			// make a transfering for privacy custom token
			// fee always 0 and reuse function of normal tx for custom token ID
			propertyID, _ := common.Hash{}.NewHashFromStr(params.tokenParams.PropertyID)
			existed := statedb.PrivacyTokenIDExisted(params.transactionStateDB, *propertyID)
			if !existed {
				isBridgeToken := false
				allBridgeTokensBytes, err := statedb.GetAllBridgeTokens(params.bridgeStateDB)
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
			txToken.TxTokenData.SetPropertyID(*propertyID)
			txToken.TxTokenData.SetMintable(params.tokenParams.Mintable)

			txToken.TxTokenData.TxNormal = new(TxVersion1)
			err := txToken.TxTokenData.TxNormal.Init(NewTxPrivacyInitParams(params.senderKey,
				params.tokenParams.Receiver,
				params.tokenParams.TokenInput,
				params.tokenParams.Fee,
				params.hasPrivacyToken,
				params.transactionStateDB,
				propertyID,
				nil,
				nil))
			if err != nil {
				return NewTransactionErr(PrivacyTokenInitTokenDataError, err)
			}
		}
	}
	if !handled {
		return NewTransactionErr(PrivacyTokenTxTypeNotHandleError, errors.New("can't handle this TokenTxType"))
	}
	return nil
}

func (txToken TxTokenVersion1) ValidateTxByItself(hasPrivacyCoin bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, chainRetriever metadata.ChainRetriever, shardID byte, isNewTransaction bool, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever) (bool, error) {
	// check for proof, signature ...
	if ok, err := txToken.ValidateTransaction(hasPrivacyCoin, transactionStateDB, bridgeStateDB, shardID, nil, false, isNewTransaction); !ok {
		return false, err
	}
	// check for metadata
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

func (txToken TxTokenVersion1) ValidateTransaction(hasPrivacyCoin bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	// validate for PRV
	ok, err := txToken.Tx.ValidateTransaction(hasPrivacyCoin, transactionStateDB, bridgeStateDB, shardID, nil, isBatch, isNewTransaction)
	if ok {
		// validate for pToken
		tokenID := txToken.TxTokenData.PropertyID
		if txToken.TxTokenData.Type == CustomTokenInit {
			if txToken.TxTokenData.Mintable {
				// mintable type will be handled elsewhere, here we return true
				return true, nil
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
			return txToken.TxTokenData.TxNormal.ValidateTransaction(
				txToken.TxTokenData.TxNormal.IsPrivacy(),
				transactionStateDB, bridgeStateDB, shardID, &tokenID, isBatch, isNewTransaction)
		}
	}
	return false, err
}

func (txToken TxTokenVersion1) ValidateSanityData(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) (bool, error) {
	// validate metadata
	check, err := validateSanityMetadata(&txToken, chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if !check || err != nil {
		return false, NewTransactionErr(InvalidSanityDataPrivacyTokenError, err)
	}
	if txToken.GetTokenID().String() == common.PRVCoinID.String(){
		return false, NewTransactionErr(InvalidSanityDataPrivacyTokenError, errors.New("cannot transfer PRV via txtoken"))
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

func (txToken TxTokenVersion1) GetTxActualSize() uint64 {
	normalTxSize := txToken.Tx.GetTxActualSize()
	tokenDataSize := uint64(0)
	tokenDataSize += txToken.TxTokenData.TxNormal.GetTxActualSize()
	tokenDataSize += uint64(len(txToken.TxTokenData.PropertyName))
	tokenDataSize += uint64(len(txToken.TxTokenData.PropertySymbol))
	tokenDataSize += uint64(len(txToken.TxTokenData.PropertyID))
	tokenDataSize += 4 // for TxPrivacyTokenDataVersion1.Type
	tokenDataSize += 8 // for TxPrivacyTokenDataVersion1.Amount
	meta := txToken.GetMetadata()
	if meta != nil {
		tokenDataSize += meta.CalculateSize()
	}
	return normalTxSize + uint64(math.Ceil(float64(tokenDataSize)/1024))
}