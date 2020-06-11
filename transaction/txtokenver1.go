package transaction

import (
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
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
	params, ok := paramsInterface.(*TxPrivacyTokenInitParams)
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
	normalTx, err := NewTxPrivacyFromParams(txPrivacyParams)
	if err != nil {
		Logger.Log.Errorf("Cannot create tx from params, error %v", err)
		return NewTransactionErr(PrivacyTokenInitFeeParamsError, err)
	}
	if err = normalTx.Init(txPrivacyParams); err != nil {
		return NewTransactionErr(PrivacyTokenInitPRVError, err)
	}
	// override TxCustomTokenPrivacyType type
	normalTx.SetType(common.TxCustomTokenPrivacyType)
	txToken.TxBase = NewTxBaseFromMetadataTx(normalTx)

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
	txToken.TxPrivacyTokenData.SetType(params.tokenParams.TokenTxType)
	txToken.TxPrivacyTokenData.SetPropertyName(params.tokenParams.PropertyName)
	txToken.TxPrivacyTokenData.SetPropertySymbol(params.tokenParams.PropertySymbol)

	switch params.tokenParams.TokenTxType {
		case CustomTokenInit: {
			// case init a new privacy custom token
			handled = true
			txToken.TxPrivacyTokenData.SetAmount(params.tokenParams.Amount)
			temp := TxVersion1{}
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

			sndOut := privacy.RandomScalar()
			tempOutputCoin[0].CoinDetails.SetSNDerivator(sndOut)
			// create coin commitment
			err = tempOutputCoin[0].CoinDetails.CommitAll()
			if err != nil {
				return NewTransactionErr(CommitOutputCoinError, err)
			}
			temp.Proof.SetOutputCoins(coin.ArrayCoinV1ToCoin(tempOutputCoin))

			// get last byte
			temp.PubKeyLastByteSender = params.tokenParams.Receiver[0].PaymentAddress.Pk[len(params.tokenParams.Receiver[0].PaymentAddress.Pk)-1]

			// sign Tx
			temp.SigPubKey = params.tokenParams.Receiver[0].PaymentAddress.Pk
			temp.sigPrivKey = *params.senderKey
			err = temp.sign()
			if err != nil {
				Logger.Log.Error(errors.New("can't sign this tx"))
				return NewTransactionErr(SignTxError, err)
			}

			txToken.TxPrivacyTokenData.TxNormal = NewTxBaseFromMetadataTx(&temp)
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
				txToken.TxPrivacyTokenData.Mintable = true
			} else {
				//NOTICE: @merman update PropertyID calculated from hash of tokendata and shardID
				newHashInitToken := common.HashH(append(hashInitToken.GetBytes(), params.shardID))
				Logger.Log.Debug("New Privacy Token %+v ", newHashInitToken)
				existed := statedb.PrivacyTokenIDExisted(params.transactionStateDB, newHashInitToken)
				if existed {
					Logger.Log.Error("INIT Tx Custom Token Privacy is Existed", newHashInitToken)
					return NewTransactionErr(TokenIDExistedError, errors.New("this token is existed in network"))
				}
				txToken.TxPrivacyTokenData.PropertyID = newHashInitToken
				Logger.Log.Debugf("A new token privacy wil be issued with ID: %+v", txToken.TxPrivacyTokenData.PropertyID.String())
			}
		}
		case CustomTokenTransfer: {
			handled = true
			// make a transfering for privacy custom token
			// fee always 0 and reuse function of normal tx for custom token ID
			temp := TxVersion1{}
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
			txToken.TxPrivacyTokenData.SetPropertyID(*propertyID)
			txToken.TxPrivacyTokenData.SetMintable(params.tokenParams.Mintable)

			err := temp.Init(NewTxPrivacyInitParams(params.senderKey,
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
			txToken.TxPrivacyTokenData.TxNormal = NewTxBaseFromMetadataTx(&temp)
		}
	}
	if !handled {
		return NewTransactionErr(PrivacyTokenTxTypeNotHandleError, errors.New("can't handle this TokenTxType"))
	}
	return nil
}