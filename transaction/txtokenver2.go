package transaction

import (
	"fmt"
	"errors"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"strconv"
)

type TxTokenVersion2 struct {
	TxTokenBase
}

func (txToken *TxTokenVersion2) Init(paramsInterface interface{}) error {
	params, ok := paramsInterface.(*TxPrivacyTokenInitParams)
	if !ok {
		return errors.New("Cannot init TxCustomTokenPrivacy because params is not correct")
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

			temp := TxVersion2{}
			temp.SetVersion(txVersion2Number)
			temp.SetType(common.TxNormalType)
			temp.Proof = new(privacy.ProofV2)
			temp.Proof.Init()
			if err = temp.Proof.SetOutputCoins(tempOutputCoin); err != nil {
				Logger.Log.Errorf("Init customPrivacyToken cannot set outputCoins")
				return err
			}
			if temp.PubKeyLastByteSender, err = params.inputCoin[0].GetShardID(); err != nil {
				return NewTransactionErr(GetShardIDByPublicKeyError, err)
			}
			temp.sigPrivKey = *params.senderKey
			temp.Sig, temp.SigPubKey, err = signNoPrivacy(params.senderKey, temp.Hash()[:])
			if err != nil {
				Logger.Log.Error(errors.New("can't sign this tx"))
				return NewTransactionErr(SignTxError, err)
			}
			temp.SigPubKey = params.tokenParams.Receiver[0].PaymentAddress.Pk

			fmt.Println("Checking TxNormal Type")
			fmt.Println("Checking TxNormal Type")
			fmt.Println("Checking TxNormal Type")
			fmt.Println("Checking TxNormal Type")
			fmt.Println("Temp Type =", temp.GetType())
			txToken.TxPrivacyTokenData.TxNormal = NewTxBaseFromMetadataTx(&temp)
			fmt.Println("TxNormal Type =", txToken.TxPrivacyTokenData.TxNormal.GetType())

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

			params := NewTxPrivacyInitParams(
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
			tx, err := NewTxPrivacyFromParams(params)
			if err != nil {
				Logger.Log.Errorf("Cannot init NewTxPrivacyInitParams: params has error %v", err)
				return NewTransactionErr(PrivacyTokenInitTokenDataError, err)
			}
			if err = tx.Init(params); err != nil {
				return NewTransactionErr(PrivacyTokenInitTokenDataError, err)
			}
			txToken.TxPrivacyTokenData.TxNormal = NewTxBaseFromMetadataTx(tx)
		}
	}
	if !handled {
		return NewTransactionErr(PrivacyTokenTxTypeNotHandleError, errors.New("can't handle this TokenTxType"))
	}
	return nil
}

func (txToken *TxTokenVersion2) InitTxTokenSalary(otaCoin *coin.CoinV2, privKey *privacy.PrivateKey, stateDB *statedb.StateDB, metaData metadata.Metadata, coinID *common.Hash, coinName string) error {
	// init data for tx PRV for fee
	txPrivacyParams := NewTxPrivacyInitParams(
		privKey, []*privacy.PaymentInfo{}, nil, 0, false, stateDB, nil, metaData, nil,
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
	// sign Tx
	temp.sigPrivKey = *privKey
	if temp.Sig, _, err = signNoPrivacy(privKey, temp.Hash()[:]); err != nil {
		Logger.Log.Error(errors.New("can't sign this tx"))
		return NewTransactionErr(SignTxError, err)
	}
	temp.SigPubKey = otaCoin.GetPublicKey().ToBytesS()

	txToken.TxPrivacyTokenData.TxNormal = NewTxBaseFromMetadataTx(&temp)
	return nil
}
