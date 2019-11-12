package rpcservice

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/incognitochain/incognito-chain/wire"

	error2 "github.com/pkg/errors"
)

type TxService struct {
	DB           *database.DatabaseInterface
	BlockChain   *blockchain.BlockChain
	Wallet       *wallet.Wallet
	FeeEstimator map[byte]*mempool.FeeEstimator
	TxMemPool    *mempool.TxPool
}

// chooseBestOutCoinsToSpent returns list of unspent coins for spending with amount
func (txService TxService) chooseBestOutCoinsToSpent(outCoins []*privacy.OutputCoin, amount uint64) (resultOutputCoins []*privacy.OutputCoin, remainOutputCoins []*privacy.OutputCoin, totalResultOutputCoinAmount uint64, err error) {
	resultOutputCoins = make([]*privacy.OutputCoin, 0)
	remainOutputCoins = make([]*privacy.OutputCoin, 0)
	totalResultOutputCoinAmount = uint64(0)

	// either take the smallest coins, or a single largest one
	var outCoinOverLimit *privacy.OutputCoin
	outCoinsUnderLimit := make([]*privacy.OutputCoin, 0)

	for _, outCoin := range outCoins {
		if outCoin.CoinDetails.GetValue() < amount {
			outCoinsUnderLimit = append(outCoinsUnderLimit, outCoin)
		} else if outCoinOverLimit == nil {
			outCoinOverLimit = outCoin
		} else if outCoinOverLimit.CoinDetails.GetValue() > outCoin.CoinDetails.GetValue() {
			remainOutputCoins = append(remainOutputCoins, outCoin)
		} else {
			remainOutputCoins = append(remainOutputCoins, outCoinOverLimit)
			outCoinOverLimit = outCoin
		}
	}

	sort.Slice(outCoinsUnderLimit, func(i, j int) bool {
		return outCoinsUnderLimit[i].CoinDetails.GetValue() < outCoinsUnderLimit[j].CoinDetails.GetValue()
	})

	for _, outCoin := range outCoinsUnderLimit {
		if totalResultOutputCoinAmount < amount {
			totalResultOutputCoinAmount += outCoin.CoinDetails.GetValue()
			resultOutputCoins = append(resultOutputCoins, outCoin)
		} else {
			remainOutputCoins = append(remainOutputCoins, outCoin)
		}
	}

	if outCoinOverLimit != nil && (outCoinOverLimit.CoinDetails.GetValue() > 2*amount || totalResultOutputCoinAmount < amount) {
		remainOutputCoins = append(remainOutputCoins, resultOutputCoins...)
		resultOutputCoins = []*privacy.OutputCoin{outCoinOverLimit}
		totalResultOutputCoinAmount = outCoinOverLimit.CoinDetails.GetValue()
	} else if outCoinOverLimit != nil {
		remainOutputCoins = append(remainOutputCoins, outCoinOverLimit)
	}

	if totalResultOutputCoinAmount < amount {
		return resultOutputCoins, remainOutputCoins, totalResultOutputCoinAmount, errors.New("Not enough coin")
	} else {
		return resultOutputCoins, remainOutputCoins, totalResultOutputCoinAmount, nil
	}
}

func (txService TxService) filterMemPoolOutcoinsToSpent(outCoins []*privacy.OutputCoin) ([]*privacy.OutputCoin, error) {
	remainOutputCoins := make([]*privacy.OutputCoin, 0)

	for _, outCoin := range outCoins {
		if txService.TxMemPool.ValidateSerialNumberHashH(outCoin.CoinDetails.GetSerialNumber().ToBytesS()) == nil {
			remainOutputCoins = append(remainOutputCoins, outCoin)
		}
	}
	return remainOutputCoins, nil
}

// chooseOutsCoinByKeyset returns list of input coins native token to spent
func (txService TxService) chooseOutsCoinByKeyset(
	paymentInfos []*privacy.PaymentInfo,
	unitFeeNativeToken int64, numBlock uint64, keySet *incognitokey.KeySet, shardIDSender byte,
	hasPrivacy bool,
	metadataParam metadata.Metadata,
	customTokenParams *transaction.CustomTokenParamTx,
	privacyCustomTokenParams *transaction.CustomTokenPrivacyParamTx,
	isGetFeePToken bool,
	unitFeePToken int64,
	db database.DatabaseInterface,
) ([]*privacy.InputCoin, uint64, *RPCError) {
	// estimate fee according to 8 recent block
	if numBlock == 0 {
		numBlock = 1000
	}
	// calculate total amount to send
	totalAmmount := uint64(0)
	for _, receiver := range paymentInfos {
		totalAmmount += receiver.Amount
	}

	// get list outputcoins tx
	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	outCoins, err := txService.BlockChain.GetListOutputCoinsByKeyset(keySet, shardIDSender, prvCoinID)
	if err != nil {
		return nil, 0, NewRPCError(GetOutputCoinError, err)
	}
	// remove out coin in mem pool
	outCoins, err = txService.filterMemPoolOutcoinsToSpent(outCoins)
	if err != nil {
		return nil, 0, NewRPCError(GetOutputCoinError, err)
	}
	if len(outCoins) == 0 && totalAmmount > 0 {
		return nil, 0, NewRPCError(GetOutputCoinError, errors.New("not enough output coin"))
	}
	// Use Knapsack to get candiate output coin
	candidateOutputCoins, outCoins, candidateOutputCoinAmount, err := txService.chooseBestOutCoinsToSpent(outCoins, totalAmmount)
	if err != nil {
		return nil, 0, NewRPCError(GetOutputCoinError, err)
	}
	// refund out put for sender
	overBalanceAmount := candidateOutputCoinAmount - totalAmmount
	if overBalanceAmount > 0 {
		// add more into output for estimate fee
		paymentInfos = append(paymentInfos, &privacy.PaymentInfo{
			PaymentAddress: keySet.PaymentAddress,
			Amount:         overBalanceAmount,
		})
	}

	// check real fee(nano PRV) per tx

	beaconState, err := txService.BlockChain.BestState.GetClonedBeaconBestState()
	beaconHeight := beaconState.BeaconHeight
	realFee, _, _, err := txService.EstimateFee(unitFeeNativeToken, false, candidateOutputCoins,
		paymentInfos, shardIDSender, numBlock, hasPrivacy,
		metadataParam, customTokenParams,
		privacyCustomTokenParams, db, int64(beaconHeight))
	if err != nil {
		return nil, 0, NewRPCError(RejectInvalidFeeError, err)
	}

	if totalAmmount == 0 && realFee == 0 {
		if metadataParam != nil {
			metadataType := metadataParam.GetType()
			switch metadataType {
			case metadata.WithDrawRewardRequestMeta:
				{
					return nil, realFee, nil
				}
			}
			return nil, realFee, NewRPCError(RejectInvalidFeeError, errors.New(fmt.Sprintf("totalAmmount: %+v, realFee: %+v", totalAmmount, realFee)))
		}
		if privacyCustomTokenParams != nil {
			// for privacy token
			return nil, 0, nil
		}
	}

	needToPayFee := int64((totalAmmount + realFee) - candidateOutputCoinAmount)
	// if not enough to pay fee
	if needToPayFee > 0 {
		if len(outCoins) > 0 {
			candidateOutputCoinsForFee, _, _, err1 := txService.chooseBestOutCoinsToSpent(outCoins, uint64(needToPayFee))
			if err != nil {
				return nil, 0, NewRPCError(GetOutputCoinError, err1)
			}
			candidateOutputCoins = append(candidateOutputCoins, candidateOutputCoinsForFee...)
		}
	}
	// convert to inputcoins
	inputCoins := transaction.ConvertOutputCoinToInputCoin(candidateOutputCoins)
	return inputCoins, realFee, nil
}

// EstimateFee - estimate fee from tx data and return real full fee, fee per kb and real tx size
// if isGetPTokenFee == true: return fee for ptoken
// if isGetPTokenFee == false: return fee for native token
func (txService TxService) EstimateFee(
	defaultFee int64,
	isGetPTokenFee bool,
	candidateOutputCoins []*privacy.OutputCoin,
	paymentInfos []*privacy.PaymentInfo, shardID byte,
	numBlock uint64, hasPrivacy bool,
	metadata metadata.Metadata,
	customTokenParams *transaction.CustomTokenParamTx,
	privacyCustomTokenParams *transaction.CustomTokenPrivacyParamTx,
	db database.DatabaseInterface,
	beaconHeight int64) (uint64, uint64, uint64, error) {
	if numBlock == 0 {
		numBlock = 1000
	}
	// check real fee(nano PRV) per tx
	var realFee uint64
	estimateFeeCoinPerKb := uint64(0)
	estimateTxSizeInKb := uint64(0)

	tokenId := &common.Hash{}
	if isGetPTokenFee {
		if privacyCustomTokenParams != nil {
			tokenId, _ = common.Hash{}.NewHashFromStr(privacyCustomTokenParams.PropertyID)
		}
	} else {
		tokenId = nil
	}

	estimateFeeCoinPerKb, err := txService.EstimateFeeWithEstimator(defaultFee, shardID, numBlock, tokenId, beaconHeight, db)
	if err != nil {
		return 0, 0, 0, err
	}

	if txService.Wallet != nil {
		estimateFeeCoinPerKb += uint64(txService.Wallet.GetConfig().IncrementalFee)
	}

	limitFee := uint64(0)
	if feeEstimator, ok := txService.FeeEstimator[shardID]; ok {
		limitFee = feeEstimator.GetLimitFeeForNativeToken()
	}
	estimateTxSizeInKb = transaction.EstimateTxSize(transaction.NewEstimateTxSizeParam(candidateOutputCoins, paymentInfos, hasPrivacy, metadata, customTokenParams, privacyCustomTokenParams, limitFee))

	realFee = uint64(estimateFeeCoinPerKb) * uint64(estimateTxSizeInKb)
	return realFee, estimateFeeCoinPerKb, estimateTxSizeInKb, nil
}

// EstimateFeeWithEstimator - only estimate fee by estimator and return fee per kb
// if tokenID != nil: return fee per kb for pToken (return error if there is no exchange rate between pToken and native token)
// if tokenID == nil: return fee per kb for native token
func (txService TxService) EstimateFeeWithEstimator(defaultFee int64, shardID byte, numBlock uint64, tokenId *common.Hash, beaconHeight int64, db database.DatabaseInterface) (uint64, error) {
	if defaultFee == 0 {
		return uint64(defaultFee), nil
	}

	unitFee := uint64(0)
	if defaultFee == -1 {
		// estimate fee on the blocks before (in native token or in pToken)
		if _, ok := txService.FeeEstimator[shardID]; ok {
			temp, _ := txService.FeeEstimator[shardID].EstimateFee(numBlock, tokenId)
			unitFee = uint64(temp)
		}
	} else {
		// get default fee (in native token or in ptoken)
		unitFee = uint64(defaultFee)
	}

	// get limit fee for native token
	limitFee := uint64(0)
	if feeEstimator, ok := txService.FeeEstimator[shardID]; ok {
		limitFee = feeEstimator.GetLimitFeeForNativeToken()
	}

	if tokenId == nil {
		// check with limit fee
		if unitFee < limitFee {
			unitFee = limitFee
		}

		return unitFee, nil
	} else {
		// convert limit fee native token to limit fee ptoken
		limitFeePTokenTmp, err := metadata.ConvertNativeTokenToPrivacyToken(limitFee, tokenId, beaconHeight, db)
		limitFeePToken := uint64(math.Ceil(limitFeePTokenTmp))
		if err != nil {
			return uint64(0), err
		}

		// check with limit fee ptoken
		if unitFee < limitFeePToken {
			unitFee = limitFeePToken
		}

		// add extra fee to make sure tx is confirmed
		// extra fee = unitFee * 10%
		unitFee += uint64(math.Ceil(float64(unitFee) * float64(0.1)))
		return unitFee, nil
	}
}

func (txService TxService) BuildRawTransaction(params *bean.CreateRawTxParam, meta metadata.Metadata, db database.DatabaseInterface) (*transaction.Tx, *RPCError) {
	Logger.log.Infof("Params: \n%+v\n\n\n", params)

	// get output coins to spend and real fee
	inputCoins, realFee, err1 := txService.chooseOutsCoinByKeyset(
		params.PaymentInfos, params.EstimateFeeCoinPerKb, 0,
		params.SenderKeySet, params.ShardIDSender, params.HasPrivacyCoin,
		meta, nil, nil, false, int64(0), db)
	if err1 != nil {
		return nil, err1
	}

	// init tx
	tx := transaction.Tx{}
	err := tx.Init(
		transaction.NewTxPrivacyInitParams(
			&params.SenderKeySet.PrivateKey,
			params.PaymentInfos,
			inputCoins,
			realFee,
			params.HasPrivacyCoin,
			*txService.DB,
			nil, // use for prv coin -> nil is valid
			meta,
			params.Info,
		))
	if err != nil {
		return nil, NewRPCError(CreateTxDataError, err)
	}

	return &tx, nil
}

func (txService TxService) CreateRawTransaction(params *bean.CreateRawTxParam, meta metadata.Metadata, db database.DatabaseInterface) (*common.Hash, []byte, byte, *RPCError) {
	var err error
	tx, err := txService.BuildRawTransaction(params, meta, db)
	if err.(*RPCError) != nil {
		Logger.log.Critical(err)
		return nil, nil, byte(0), NewRPCError(CreateTxDataError, err)
	}

	txBytes, err := json.Marshal(tx)
	if err != nil {
		// return hex for a new tx
		return nil, nil, byte(0), NewRPCError(CreateTxDataError, err)
	}

	txShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())

	return tx.Hash(), txBytes, txShardID, nil
}

func (txService TxService) SendRawTransaction(txB58Check string) (wire.Message, *common.Hash, byte, *RPCError) {
	rawTxBytes, _, err := base58.Base58Check{}.Decode(txB58Check)
	if err != nil {
		Logger.log.Errorf("handleSendRawTransaction result: %+v, err: %+v", nil, err)
		return nil, nil, byte(0), NewRPCError(SendTxDataError, err)
	}
	var tx transaction.Tx
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		Logger.log.Errorf("handleSendRawTransaction result: %+v, err: %+v", nil, err)
		return nil, nil, byte(0), NewRPCError(SendTxDataError, err)
	}

	hash, _, err := txService.TxMemPool.MaybeAcceptTransaction(&tx, -1)
	//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
	if err != nil {
		mempoolErr, ok := err.(*mempool.MempoolTxError)
		if ok {
			if mempoolErr.Code == mempool.ErrCodeMessage[mempool.RejectInvalidFee].Code {
				Logger.log.Errorf("handleSendRawTransaction result: s%+v, err: %+v", nil, err)
				return nil, nil, byte(0), NewRPCError(RejectInvalidFeeError, mempoolErr)
			}
		}
		Logger.log.Errorf("handleSendRawTransaction result: %+v, err: %+v", nil, err)
		return nil, nil, byte(0), NewRPCError(SendTxDataError, err)
	}

	Logger.log.Debugf("New transaction hash: %+v \n", *hash)

	// broadcast Message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
	if err != nil {
		Logger.log.Errorf("handleSendRawTransaction result: %+v, err: %+v", nil, err)
		return nil, nil, byte(0), NewRPCError(SendTxDataError, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx

	return txMsg, tx.Hash(), tx.PubKeyLastByteSender, nil
}

func (txService TxService) BuildTokenParam(tokenParamsRaw map[string]interface{}, senderKeySet *incognitokey.KeySet, shardIDSender byte) (
	*transaction.CustomTokenParamTx, *transaction.CustomTokenPrivacyParamTx, *RPCError) {

	var customTokenParams *transaction.CustomTokenParamTx
	var privacyTokenParam *transaction.CustomTokenPrivacyParamTx
	var err *RPCError

	isPrivacy, ok := tokenParamsRaw["Privacy"].(bool)
	if !ok {
		return nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("is privacy param is invalid"))
	}

	if !isPrivacy {
		// Check normal custom token param
		customTokenParams, _, err = txService.BuildCustomTokenParam(tokenParamsRaw, senderKeySet)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// Check privacy custom token param
		privacyTokenParam, _, _, err = txService.BuildPrivacyCustomTokenParam(tokenParamsRaw, senderKeySet, shardIDSender)
		if err != nil {
			return nil, nil, err
		}
	}

	return customTokenParams, privacyTokenParam, nil

}

func (txService TxService) BuildCustomTokenParam(tokenParamsRaw map[string]interface{}, senderKeySet *incognitokey.KeySet) (*transaction.CustomTokenParamTx, map[common.Hash]transaction.TxNormalToken, *RPCError) {
	tokenIDParam, ok := tokenParamsRaw["TokenID"].(string)
	if !ok {
		return nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("token id param is invalid"))
	}

	tokenNameParam, ok := tokenParamsRaw["TokenName"].(string)
	if !ok {
		return nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("token name param is invalid"))
	}

	tokenSymbolParam, ok := tokenParamsRaw["TokenSymbol"].(string)
	if !ok {
		return nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("token symbol param is invalid"))
	}

	tokenTxTypeParam, ok := tokenParamsRaw["TokenTxType"].(float64)
	if !ok {
		return nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("token tx type param is invalid"))
	}

	tokenAmountParam, ok := tokenParamsRaw["TokenAmount"].(float64)
	if !ok {
		return nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("token amount param is invalid"))
	}

	tokenReceivers, ok := tokenParamsRaw["TokenReceivers"].(interface{})
	if !ok {
		return nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("token receivers param is invalid"))
	}

	tokenParams := &transaction.CustomTokenParamTx{
		PropertyID:     tokenIDParam,
		PropertyName:   tokenNameParam,
		PropertySymbol: tokenSymbolParam,
		TokenTxType:    int(tokenTxTypeParam),
		Amount:         uint64(tokenAmountParam),
	}
	voutsAmount := int64(0)
	var err error
	tokenParams.Receiver, voutsAmount, err = transaction.CreateCustomTokenReceiverArray(tokenReceivers)
	if err != nil {
		return nil, nil, NewRPCError(RPCInvalidParamsError, err)
	}

	switch tokenParams.TokenTxType {
	case transaction.CustomTokenTransfer:
		{
			tokenID, err := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
			if err != nil {
				return nil, nil, NewRPCError(RPCInvalidParamsError, error2.Wrap(err, "Token ID is invalid"))
			}

			//if _, ok := listCustomTokens[*tokenID]; !ok {
			//	return nil, nil, NewRPCError(ErrRPCInvalidParams, errors.New("Invalid Token ID"))
			//}

			existed := txService.BlockChain.CustomTokenIDExisted(tokenID)
			if !existed {
				return nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("Invalid Token ID"))
			}

			unspentTxTokenOuts, err := txService.BlockChain.GetUnspentTxCustomTokenVout(*senderKeySet, tokenID)
			Logger.log.Info("BuildRawCustomTokenTransaction ", unspentTxTokenOuts)
			if err != nil {
				return nil, nil, NewRPCError(GetOutputCoinError, errors.New("Token out invalid"))
			}
			if len(unspentTxTokenOuts) == 0 {
				return nil, nil, NewRPCError(GetOutputCoinError, errors.New("Token out invalid"))
			}
			txTokenIns := []transaction.TxTokenVin{}
			txTokenInsAmount := uint64(0)
			for _, out := range unspentTxTokenOuts {
				item := transaction.TxTokenVin{
					PaymentAddress:  out.PaymentAddress,
					TxCustomTokenID: out.GetTxCustomTokenID(),
					VoutIndex:       out.GetIndex(),
				}
				// create signature by keyset -> base58check.encode of txtokenout double hash
				signature, err := senderKeySet.Sign(out.Hash()[:])
				if err != nil {
					return nil, nil, NewRPCError(CanNotSignError, err)
				}
				// add signature to TxTokenVin to use token utxo
				item.Signature = base58.Base58Check{}.Encode(signature, 0)
				txTokenIns = append(txTokenIns, item)
				txTokenInsAmount += out.Value
				voutsAmount -= int64(out.Value)
				if voutsAmount <= 0 {
					break
				}
			}
			tokenParams.SetVins(txTokenIns)
			tokenParams.SetVinsAmount(txTokenInsAmount)
		}
	case transaction.CustomTokenInit:
		{
			if tokenParams.Receiver[0].Value != tokenParams.Amount { // Init with wrong max amount of custom token
				return nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("Init with wrong max amount of property"))
			}
		}
	}
	//return tokenParams, listCustomTokens, nil
	return tokenParams, nil, nil
}

// BuildRawCustomTokenTransaction ...
func (txService TxService) BuildRawCustomTokenTransaction(
	params interface{},
	metaData metadata.Metadata,
	db database.DatabaseInterface,
) (*transaction.TxNormalToken, *RPCError) {
	// all params
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 5 {
		Logger.log.Debugf("BuildRawCustomTokenTransaction result: %+v", nil)
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("param must be an array at least 5 elements"))
	}

	// create basic param tx
	txparam, err := bean.NewCreateRawTxParam(params)
	if err != nil {
		return nil, NewRPCError(RPCInvalidParamsError, err)
	}

	// param #5: token params
	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("token param is invalid"))
	}
	tokenParams, _, err := txService.BuildTokenParam(tokenParamsRaw, txparam.SenderKeySet, txparam.ShardIDSender)

	if err.(*RPCError) != nil {
		return nil, err.(*RPCError)
	}
	/******* START choose output coins native coins(PRV), which is used to create tx *****/
	inputCoins, realFee, err := txService.chooseOutsCoinByKeyset(txparam.PaymentInfos, txparam.EstimateFeeCoinPerKb, 0,
		txparam.SenderKeySet, txparam.ShardIDSender, txparam.HasPrivacyCoin,
		metaData, tokenParams, nil, false, int64(0), db)
	if err.(*RPCError) != nil {
		return nil, err.(*RPCError)
	}
	if len(txparam.PaymentInfos) == 0 && realFee == 0 {
		txparam.HasPrivacyCoin = false
	}
	/******* END GET output coins native coins(PRV), which is used to create tx *****/

	tx := &transaction.TxNormalToken{}
	err = tx.Init(
		transaction.NewTxNormalTokenInitParam(&txparam.SenderKeySet.PrivateKey,
			nil,
			inputCoins,
			realFee,
			tokenParams,
			//listCustomTokens,
			*txService.DB,
			metaData,
			txparam.HasPrivacyCoin,
			txparam.ShardIDSender))
	if err != nil {
		return nil, NewRPCError(CreateTxDataError, err)
	}

	return tx, nil
}

func (txService TxService) BuildPrivacyCustomTokenParam(tokenParamsRaw map[string]interface{}, senderKeySet *incognitokey.KeySet, shardIDSender byte) (*transaction.CustomTokenPrivacyParamTx, map[common.Hash]transaction.TxCustomTokenPrivacy, map[common.Hash]blockchain.CrossShardTokenPrivacyMetaData, *RPCError) {
	property, ok := tokenParamsRaw["TokenID"].(string)
	if !ok {
		return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("Token ID is invalid"))
	}
	tokenName, ok := tokenParamsRaw["TokenName"].(string)
	if !ok {
		return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("Token Name is invalid"))
	}
	tokenSymbol, ok := tokenParamsRaw["TokenSymbol"].(string)
	if !ok {
		return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("Token Symbol is invalid"))
	}
	tokenTxType, ok := tokenParamsRaw["TokenTxType"].(float64)
	if !ok {
		return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("Token tx type is invalid"))
	}
	tokenAmount, ok := tokenParamsRaw["TokenAmount"].(float64)
	if !ok {
		return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("Token amount is invalid"))
	}
	tokenFee, ok := tokenParamsRaw["TokenFee"].(float64)
	if !ok {
		return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("Token fee is invalid"))
	}
	if tokenTxType == transaction.CustomTokenInit {
		tokenFee = 0
	}
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID:     property,
		PropertyName:   tokenName,
		PropertySymbol: tokenSymbol,
		TokenTxType:    int(tokenTxType),
		Amount:         uint64(tokenAmount),
		TokenInput:     nil,
		Fee:            uint64(tokenFee),
	}
	voutsAmount := int64(0)
	tokenParams.Receiver, voutsAmount = transaction.CreateCustomTokenPrivacyReceiverArray(tokenParamsRaw["TokenReceivers"])
	voutsAmount += int64(tokenFee)

	// get list custom token
	switch tokenParams.TokenTxType {
	case transaction.CustomTokenTransfer:
		{
			tokenID, err := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
			if err != nil {
				return nil, nil, nil, NewRPCError(RPCInvalidParamsError, err)
			}
			existed := txService.BlockChain.PrivacyCustomTokenIDExisted(tokenID)
			existedCrossShard := txService.BlockChain.PrivacyCustomTokenIDCrossShardExisted(tokenID)
			if !existed && !existedCrossShard {
				return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("Invalid Token ID"))
			}
			outputTokens, err := txService.BlockChain.GetListOutputCoinsByKeyset(senderKeySet, shardIDSender, tokenID)
			if err != nil {
				return nil, nil, nil, NewRPCError(GetOutputCoinError, err)
			}
			outputTokens, err = txService.filterMemPoolOutcoinsToSpent(outputTokens)
			if err != nil {
				return nil, nil, nil, NewRPCError(GetOutputCoinError, err)
			}
			candidateOutputTokens, _, _, err := txService.chooseBestOutCoinsToSpent(outputTokens, uint64(voutsAmount))
			if err != nil {
				return nil, nil, nil, NewRPCError(GetOutputCoinError, err)
			}
			intputToken := transaction.ConvertOutputCoinToInputCoin(candidateOutputTokens)
			tokenParams.TokenInput = intputToken
		}
	case transaction.CustomTokenInit:
		{
			if tokenParams.Receiver[0].Amount != tokenParams.Amount { // Init with wrong max amount of custom token
				return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("Init with wrong max amount of property"))
			}
			if tokenParams.PropertyName == "" {
				return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("Init with wrong name of property"))
			}
			if tokenParams.PropertySymbol == "" {
				return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New("Init with wrong symbol of property"))
			}
		}
	}
	return tokenParams, nil, nil, nil
}

// BuildRawCustomTokenTransaction ...
func (txService TxService) BuildRawPrivacyCustomTokenTransaction(
	params interface{},
	metaData metadata.Metadata,
	db database.DatabaseInterface,
) (*transaction.TxCustomTokenPrivacy, *RPCError) {
	txParam, errParam := bean.NewCreateRawPrivacyTokenTxParam(params)
	if errParam != nil {
		return nil, NewRPCError(RPCInvalidParamsError, errParam)
	}
	tokenParamsRaw := txParam.TokenParamsRaw
	var err error
	_, tokenParams, err := txService.BuildTokenParam(tokenParamsRaw, txParam.SenderKeySet, txParam.ShardIDSender)

	if err.(*RPCError) != nil {
		return nil, err.(*RPCError)
	}

	/******* START choose output native coins(PRV), which is used to create tx *****/
	var inputCoins []*privacy.InputCoin
	realFeePRV := uint64(0)
	inputCoins, realFeePRV, err = txService.chooseOutsCoinByKeyset(txParam.PaymentInfos,
		txParam.EstimateFeeCoinPerKb, 0, txParam.SenderKeySet,
		txParam.ShardIDSender, txParam.HasPrivacyCoin, nil,
		nil, tokenParams, txParam.IsGetPTokenFee, txParam.UnitPTokenFee, db)
	if err.(*RPCError) != nil {
		return nil, err.(*RPCError)
	}

	if len(txParam.PaymentInfos) == 0 && realFeePRV == 0 {
		txParam.HasPrivacyCoin = false
	}
	/******* END GET output coins native coins(PRV), which is used to create tx *****/

	tx := &transaction.TxCustomTokenPrivacy{}
	err = tx.Init(
		transaction.NewTxPrivacyTokenInitParams(&txParam.SenderKeySet.PrivateKey,
			txParam.PaymentInfos,
			inputCoins,
			realFeePRV,
			tokenParams,
			*txService.DB,
			metaData,
			txParam.HasPrivacyCoin,
			txParam.HasPrivacyToken,
			txParam.ShardIDSender, txParam.Info))

	if err != nil {
		return nil, NewRPCError(CreateTxDataError, err)
	}

	return tx, nil
}

func (txService TxService) GetTransactionHashByReceiver(paymentAddressParam string) (map[byte][]common.Hash, error) {
	var keySet *incognitokey.KeySet

	if paymentAddressParam != "" {
		senderKey, err := wallet.Base58CheckDeserialize(paymentAddressParam)
		if err != nil {
			return nil, errors.New("payment address is invalid")
		}

		keySet = &senderKey.KeySet
	} else {
		return nil, errors.New("payment address is invalid")
	}

	return txService.BlockChain.GetTransactionHashByReceiver(keySet)
}

func (txService TxService) GetTransactionByHash(txHashStr string) (*jsonresult.TransactionDetail, *RPCError) {
	txHash, err := common.Hash{}.NewHashFromStr(txHashStr)
	if err != nil {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("tx hash is invalid"))
	}
	Logger.log.Infof("Get Transaction By Hash %+v", *txHash)

	shardID, blockHash, index, tx, err := txService.BlockChain.GetTransactionByHash(*txHash)
	if err != nil {
		// maybe tx is still in tx mempool -> check mempool
		tx, errM := txService.TxMemPool.GetTx(txHash)
		if errM != nil {
			return nil, NewRPCError(TxNotExistedInMemAndBLockError, errors.New("Tx is not existed in block or mempool"))
		}
		shardIDTemp := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		result, errM := jsonresult.NewTransactionDetail(tx, nil, 0, 0, shardIDTemp)
		if errM != nil {
			return nil, NewRPCError(UnexpectedError, errM)
		}
		result.IsInMempool = true
		return result, nil
	}

	blockHeight, _, err := (*txService.DB).GetIndexOfBlock(blockHash)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	result, err := jsonresult.NewTransactionDetail(tx, &blockHash, blockHeight, index, shardID)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	result.IsInBlock = true
	Logger.log.Debugf("handleGetTransactionByHash result: %+v", result)
	return result, nil
}

func (txService TxService) SendRawCustomTokenTransaction(base58CheckData string) (wire.Message, *transaction.TxNormalToken, *RPCError) {
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)
	if err != nil {
		Logger.log.Debugf("handleSendRawCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, nil, NewRPCError(SendTxDataError, err)
	}

	tx := transaction.TxNormalToken{}
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		Logger.log.Debugf("handleSendRawCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, nil, NewRPCError(SendTxDataError, err)
	}

	hash, _, err := txService.TxMemPool.MaybeAcceptTransaction(&tx, -1)
	//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
	if err != nil {
		Logger.log.Debugf("handleSendRawCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, nil, NewRPCError(SendTxDataError, err)
	}

	Logger.log.Debugf("New Custom Token Transaction: %s\n", hash.String())

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
	if err != nil {
		Logger.log.Debugf("handleSendRawCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, nil, NewRPCError(SendTxDataError, err)
	}

	txMsg.(*wire.MessageTxToken).Transaction = &tx

	return txMsg, &tx, nil
}

func (txService TxService) GetListCustomTokenHolders(tokenIDString string) (map[string]uint64, *RPCError) {
	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDString)
	if err != nil {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("TokenID is invalid"))
	}
	result, err := txService.BlockChain.GetListTokenHolders(tokenID)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}

	return result, nil
}

func (txService TxService) GetListCustomTokenBalance(accountParam string) (jsonresult.ListCustomTokenBalance, error) {
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}
	account, err := wallet.Base58CheckDeserialize(accountParam)
	if err != nil {
		Logger.log.Debugf("handleGetListCustomTokenBalance result: %+v, err: %+v", nil, err)
		return jsonresult.ListCustomTokenBalance{}, nil
	}
	result.PaymentAddress = accountParam
	accountPaymentAddress := account.KeySet.PaymentAddress
	temps, err := txService.BlockChain.ListCustomToken()
	if err != nil {
		Logger.log.Debugf("handleGetListCustomTokenBalance result: %+v, err: %+v", nil, err)
		return jsonresult.ListCustomTokenBalance{}, err
	}
	for _, tx := range temps {
		item := jsonresult.CustomTokenBalance{}
		item.Name = tx.TxTokenData.PropertyName
		item.Symbol = tx.TxTokenData.PropertySymbol
		item.TokenID = tx.TxTokenData.PropertyID.String()
		item.TokenImage = common.Render([]byte(item.TokenID))
		tokenID := tx.TxTokenData.PropertyID
		res, err := txService.BlockChain.GetListTokenHolders(&tokenID)
		if err != nil {
			return jsonresult.ListCustomTokenBalance{}, err
		}
		paymentAddressInStr := base58.Base58Check{}.Encode(accountPaymentAddress.Bytes(), 0x00)
		item.Amount = res[paymentAddressInStr]
		if item.Amount == 0 {
			continue
		}
		result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)
		result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	}
	Logger.log.Debugf("handleGetListCustomTokenBalance result: %+v", result)

	return result, nil
}

func (txService TxService) GetListPrivacyCustomTokenBalance(privateKey string) (jsonresult.ListCustomTokenBalance, *RPCError) {
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}
	account, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		Logger.log.Debugf("handleGetListPrivacyCustomTokenBalance result: %+v, err: %+v", nil, err)
		return jsonresult.ListCustomTokenBalance{}, NewRPCError(UnexpectedError, err)
	}
	err = account.KeySet.InitFromPrivateKey(&account.KeySet.PrivateKey)
	if err != nil {
		Logger.log.Debugf("handleGetListPrivacyCustomTokenBalance result: %+v, err: %+v", nil, err)
		return jsonresult.ListCustomTokenBalance{}, NewRPCError(UnexpectedError, err)
	}

	result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	temps, listCustomTokenCrossShard, err := txService.BlockChain.ListPrivacyCustomToken()
	if err != nil {
		Logger.log.Debugf("handleGetListPrivacyCustomTokenBalance result: %+v, err: %+v", nil, err)
		return jsonresult.ListCustomTokenBalance{}, NewRPCError(UnexpectedError, err)
	}
	tokenIDs := make(map[common.Hash]interface{})
	for tokenID, tx := range temps {
		tokenIDs[tokenID] = 0
		item := jsonresult.CustomTokenBalance{}
		item.Name = tx.TxPrivacyTokenData.PropertyName
		item.Symbol = tx.TxPrivacyTokenData.PropertySymbol
		item.TokenID = tx.TxPrivacyTokenData.PropertyID.String()
		item.TokenImage = common.Render([]byte(item.TokenID))
		tokenID := tx.TxPrivacyTokenData.PropertyID

		balance := uint64(0)
		// get balance for accountName in wallet
		lastByte := account.KeySet.PaymentAddress.Pk[len(account.KeySet.PaymentAddress.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		prvCoinID := &common.Hash{}
		err := prvCoinID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			return jsonresult.ListCustomTokenBalance{}, NewRPCError(TokenIsInvalidError, err)
		}
		outcoints, err := txService.BlockChain.GetListOutputCoinsByKeyset(&account.KeySet, shardIDSender, &tokenID)
		if err != nil {
			Logger.log.Debugf("handleGetListPrivacyCustomTokenBalance result: %+v, err: %+v", nil, err)
			return jsonresult.ListCustomTokenBalance{}, NewRPCError(UnexpectedError, err)
		}
		for _, out := range outcoints {
			balance += out.CoinDetails.GetValue()
		}

		item.Amount = balance
		if item.Amount == 0 {
			continue
		}
		item.IsPrivacy = true
		result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)
		result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	}
	for tokenID, customTokenCrossShard := range listCustomTokenCrossShard {
		if _, ok := tokenIDs[tokenID]; ok {
			continue
		}
		item := jsonresult.CustomTokenBalance{}
		item.Name = customTokenCrossShard.PropertyName
		item.Symbol = customTokenCrossShard.PropertySymbol
		item.TokenID = customTokenCrossShard.TokenID.String()
		item.TokenImage = common.Render([]byte(item.TokenID))
		tokenID := customTokenCrossShard.TokenID

		balance := uint64(0)
		// get balance for accountName in wallet
		lastByte := account.KeySet.PaymentAddress.Pk[len(account.KeySet.PaymentAddress.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		prvCoinID := &common.Hash{}
		err := prvCoinID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			return jsonresult.ListCustomTokenBalance{}, NewRPCError(TokenIsInvalidError, err)
		}
		outcoints, err := txService.BlockChain.GetListOutputCoinsByKeyset(&account.KeySet, shardIDSender, &tokenID)
		if err != nil {
			return jsonresult.ListCustomTokenBalance{}, NewRPCError(UnexpectedError, err)
		}
		for _, out := range outcoints {
			balance += out.CoinDetails.GetValue()
		}

		item.Amount = balance
		if item.Amount == 0 {
			continue
		}
		item.IsPrivacy = true
		result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)
		result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	}

	return result, nil
}

func (txService TxService) GetBalancePrivacyCustomToken(privateKey string, tokenID string) (uint64, *RPCError) {
	account, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		Logger.log.Debugf("handleGetBalancePrivacyCustomToken result: %+v, err: %+v", nil, err)
		return uint64(0), NewRPCError(UnexpectedError, err)
	}
	err = account.KeySet.InitFromPrivateKey(&account.KeySet.PrivateKey)
	if err != nil {
		Logger.log.Debugf("handleGetBalancePrivacyCustomToken result: %+v, err: %+v", nil, err)
		return uint64(0), NewRPCError(UnexpectedError, err)
	}

	temps, listCustomTokenCrossShard, err := txService.BlockChain.ListPrivacyCustomToken()
	if err != nil {
		Logger.log.Debugf("handleGetListPrivacyCustomTokenBalance result: %+v, err: %+v", nil, err)
		return uint64(0), NewRPCError(UnexpectedError, err)
	}
	totalValue := uint64(0)
	for tempTokenID := range temps {
		if tokenID == tempTokenID.String() {
			lastByte := account.KeySet.PaymentAddress.Pk[len(account.KeySet.PaymentAddress.Pk)-1]
			shardIDSender := common.GetShardIDFromLastByte(lastByte)
			outcoints, err := txService.BlockChain.GetListOutputCoinsByKeyset(&account.KeySet, shardIDSender, &tempTokenID)
			if err != nil {
				Logger.log.Debugf("handleGetBalancePrivacyCustomToken result: %+v, err: %+v", nil, err)
				return uint64(0), NewRPCError(UnexpectedError, err)
			}
			for _, out := range outcoints {
				totalValue += out.CoinDetails.GetValue()
			}
		}
	}
	for tempTokenID := range listCustomTokenCrossShard {
		if tokenID == tempTokenID.String() {
			lastByte := account.KeySet.PaymentAddress.Pk[len(account.KeySet.PaymentAddress.Pk)-1]
			shardIDSender := common.GetShardIDFromLastByte(lastByte)
			outcoints, err := txService.BlockChain.GetListOutputCoinsByKeyset(&account.KeySet, shardIDSender, &tempTokenID)
			if err != nil {
				Logger.log.Debugf("handleGetBalancePrivacyCustomToken result: %+v, err: %+v", nil, err)
				return uint64(0), NewRPCError(UnexpectedError, err)
			}
			for _, out := range outcoints {
				totalValue += out.CoinDetails.GetValue()
			}
		}
	}

	return totalValue, nil
}

func (txService TxService) CustomTokenDetail(tokenIDStr string) ([]common.Hash, error) {
	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDStr)
	if err != nil {
		Logger.log.Debugf("handleCustomTokenDetail result: %+v, err: %+v", nil, err)
		return nil, err
	}
	txs, _ := txService.BlockChain.GetCustomTokenTxsHash(tokenID)
	return txs, nil
}

func (txService TxService) PrivacyCustomTokenDetail(tokenIDStr string) ([]common.Hash, *transaction.TxPrivacyTokenData, error) {
	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDStr)
	if err != nil {
		Logger.log.Debugf("handlePrivacyCustomTokenDetail result: %+v, err: %+v", nil, err)
		return nil, nil, err
	}

	_, _, _ = txService.BlockChain.ListPrivacyCustomToken()
	/*tokenData := transaction.TxPrivacyTokenData{
		listTxInitPrivacyToken[tokenID].
	}*/

	txs, _ := txService.BlockChain.GetPrivacyCustomTokenTxsHash(tokenID)
	return txs, nil, nil
}

func (txService TxService) ListUnspentCustomToken(senderKeyParam string, tokenIDParam string) ([]transaction.TxTokenVout, error) {
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam)
	if err != nil {
		Logger.log.Debugf("handleListUnspentCustomToken result: %+v, err: %+v", nil, err)
		return nil, err
	}
	senderKeyset := senderKey.KeySet

	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDParam)
	if err != nil {
		Logger.log.Debugf("handleListUnspentCustomToken result: %+v, err: %+v", nil, err)
		return nil, err
	}

	unspentTxTokenOuts, err := txService.BlockChain.GetUnspentTxCustomTokenVout(senderKeyset, tokenID)
	if err != nil {
		Logger.log.Debugf("handleListUnspentCustomToken result: %+v, err: %+v", nil, err)
		return nil, err
	}

	return unspentTxTokenOuts, nil
}

func (txService TxService) GetBalanceCustomToken(senderKeyParam string, tokenIDParam string) (uint64, error) {
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam)
	if err != nil {
		Logger.log.Debugf("handleGetBalanceCustomToken result: %+v, err: %+v", nil, err)
		return uint64(0), err
	}
	senderKeyset := senderKey.KeySet

	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDParam)
	if err != nil {
		Logger.log.Debugf("handleGetBalanceCustomToken result: %+v, err: %+v", nil, err)
		return uint64(0), err
	}
	unspentTxTokenOuts, err := txService.BlockChain.GetUnspentTxCustomTokenVout(senderKeyset, tokenID)

	if err != nil {
		Logger.log.Debugf("handleGetBalanceCustomToken result: %+v, err: %+v", nil, err)
		return uint64(0), NewRPCError(UnexpectedError, err)
	}
	totalValue := uint64(0)
	for _, temp := range unspentTxTokenOuts {
		totalValue += temp.Value
	}

	return totalValue, nil
}

func (txService TxService) CreateSignatureOnCustomTokenTx(base58CheckDate string, senderKeyParam string) (string, error) {
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)
	if err != nil {
		return "", err
	}

	tx := transaction.TxNormalToken{}
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return "", err
	}

	keySet, _, err := GetKeySetFromPrivateKeyParams(senderKeyParam)
	if err != nil {
		Logger.log.Debugf("handleCreateSignatureOnCustomTokenTx result: %+v, err: %+v", nil, err)
		return "", err
	}

	jsSignByteArray, err := tx.GetTxCustomTokenSignature(*keySet)
	if err != nil {
		Logger.log.Debugf("handleCreateSignatureOnCustomTokenTx result: %+v, err: %+v", nil, err)
		return "", errors.New("failed to sign the custom token")
	}
	result := hex.EncodeToString(jsSignByteArray)

	return result, nil
}

func (txService TxService) RandomCommitments(paymentAddressStr string, outputs []interface{}, tokenID *common.Hash) ([]uint64, []uint64, [][]byte, *RPCError) {
	_, shardIDSender, err := GetKeySetFromPaymentAddressParam(paymentAddressStr)
	if err != nil {
		Logger.log.Debugf("handleRandomCommitments result: %+v, err: %+v", nil, err)
		return nil, nil, nil, NewRPCError(UnexpectedError, err)
	}

	usableOutputCoins := []*privacy.OutputCoin{}
	for _, item := range outputs {
		out, err1 := jsonresult.NewOutcoinFromInterface(item)
		if err1 != nil {
			return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New(fmt.Sprint("outputs is invalid", out)))
		}
		valueBNTmp := big.Int{}
		valueBNTmp.SetString(out.Value, 10)

		outputCoin := new(privacy.OutputCoin).Init()
		outputCoin.CoinDetails.SetValue(valueBNTmp.Uint64())

		RandomnessInBytes, _, err := base58.Base58Check{}.Decode(out.Randomness)
		if err != nil {
			return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New(fmt.Sprint("randomness output is invalid", out.Randomness)))
		}
		outputCoin.CoinDetails.SetRandomness(new(privacy.Scalar).FromBytesS(RandomnessInBytes))

		SNDerivatorInBytes, _, err := base58.Base58Check{}.Decode(out.SNDerivator)
		if err != nil {
			return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New(fmt.Sprint("snderivator output is invalid", out.SNDerivator)))
		}
		outputCoin.CoinDetails.SetSNDerivator(new(privacy.Scalar).FromBytesS(SNDerivatorInBytes))

		CoinCommitmentBytes, _, err := base58.Base58Check{}.Decode(out.CoinCommitment)
		if err != nil {
			return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New(fmt.Sprint("coin commitment output is invalid", out.CoinCommitment)))
		}
		CoinCommitment, err := new(privacy.Point).FromBytesS(CoinCommitmentBytes)
		if err != nil {
			return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New(fmt.Sprint("coin commitment output is invalid", CoinCommitmentBytes)))
		}
		outputCoin.CoinDetails.SetCoinCommitment(CoinCommitment)

		PublicKeyBytes, _, err := base58.Base58Check{}.Decode(out.PublicKey)
		if err != nil {
			return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New(fmt.Sprint("public key output is invalid", out.PublicKey)))
		}
		PublicKey, err := new(privacy.Point).FromBytesS(PublicKeyBytes)
		if err != nil {
			return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New(fmt.Sprint("public key output is invalid", PublicKeyBytes)))
		}
		outputCoin.CoinDetails.SetPublicKey(PublicKey)

		InfoBytes, _, err := base58.Base58Check{}.Decode(out.Info)
		if err != nil {
			return nil, nil, nil, NewRPCError(RPCInvalidParamsError, errors.New(fmt.Sprint("Info output is invalid", out.Info)))
		}
		outputCoin.CoinDetails.SetInfo(InfoBytes)

		usableOutputCoins = append(usableOutputCoins, outputCoin)
	}
	usableInputCoins := transaction.ConvertOutputCoinToInputCoin(usableOutputCoins)

	commitmentIndexs, myCommitmentIndexs, commitments := txService.BlockChain.RandomCommitmentsProcess(usableInputCoins, 0, shardIDSender, tokenID)
	return commitmentIndexs, myCommitmentIndexs, commitments, nil
}

func (txService TxService) SendRawPrivacyCustomTokenTransaction(base58CheckData string) (wire.Message, *transaction.TxCustomTokenPrivacy, error) {
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)
	if err != nil {
		Logger.log.Debugf("handleSendRawPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, nil, err
	}

	tx := transaction.TxCustomTokenPrivacy{}
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		Logger.log.Debugf("handleSendRawPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, nil, err
	}

	hash, _, err := txService.TxMemPool.MaybeAcceptTransaction(&tx, -1)
	//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
	if err != nil {
		Logger.log.Debugf("handleSendRawPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, nil, err
	}

	Logger.log.Debugf("there is hash of transaction: %s\n", hash.String())

	txMsg, err := wire.MakeEmptyMessage(wire.CmdPrivacyCustomToken)
	if err != nil {
		Logger.log.Debugf("handleSendRawPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, nil, err
	}

	txMsg.(*wire.MessageTxPrivacyToken).Transaction = &tx

	return txMsg, &tx, nil
}

func (txService TxService) BuildRawDefragmentAccountTransaction(params interface{}, meta metadata.Metadata, db database.DatabaseInterface) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 4 {
		return nil, NewRPCError(RPCInvalidParamsError, nil)
	}

	senderKeyParam, ok := arrayParams[0].(string)
	if !ok {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("senderKeyParam is invalid"))
	}

	maxValTemp, ok := arrayParams[1].(float64)
	if !ok {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("maxVal is invalid"))
	}
	maxVal := uint64(maxValTemp)

	estimateFeeCoinPerKbtemp, ok := arrayParams[2].(float64)
	if !ok {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("estimateFeeCoinPerKb is invalid"))
	}
	estimateFeeCoinPerKb := int64(estimateFeeCoinPerKbtemp)

	// param #4: hasPrivacyCoin flag: 1 or -1
	hasPrivacyCoinParam := arrayParams[3].(float64)
	hasPrivacyCoin := int(hasPrivacyCoinParam) > 0
	/********* END Fetch all component to *******/

	// param #1: private key of sender
	senderKeySet, shardIDSender, err := GetKeySetFromPrivateKeyParams(senderKeyParam)
	if err != nil {
		return nil, NewRPCError(InvalidSenderPrivateKeyError, err)
	}

	prvCoinID := &common.Hash{}
	err1 := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err1 != nil {
		return nil, NewRPCError(TokenIsInvalidError, err1)
	}
	outCoins, err := txService.BlockChain.GetListOutputCoinsByKeyset(senderKeySet, shardIDSender, prvCoinID)
	if err != nil {
		return nil, NewRPCError(GetOutputCoinError, err)
	}
	// remove out coin in mem pool
	outCoins, err = txService.filterMemPoolOutcoinsToSpent(outCoins)
	if err != nil {
		return nil, NewRPCError(GetOutputCoinError, err)
	}
	outCoins, amount := txService.calculateOutputCoinsByMinValue(outCoins, maxVal)
	if len(outCoins) == 0 {
		return nil, NewRPCError(GetOutputCoinError, nil)
	}
	paymentInfo := &privacy.PaymentInfo{
		Amount:         uint64(amount),
		PaymentAddress: senderKeySet.PaymentAddress,
		Message:        []byte{},
	}
	paymentInfos := []*privacy.PaymentInfo{paymentInfo}
	// check real fee(nano PRV) per tx
	isGetPTokenFee := false

	beaconState, err := txService.BlockChain.BestState.GetClonedBeaconBestState()
	beaconHeight := beaconState.BeaconHeight
	realFee, _, _, _ := txService.EstimateFee(
		estimateFeeCoinPerKb, isGetPTokenFee, outCoins, paymentInfos, shardIDSender, 8, hasPrivacyCoin,
		nil, nil, nil, db, int64(beaconHeight))
	if len(outCoins) == 0 {
		realFee = 0
	}

	if uint64(amount) < realFee {
		return nil, NewRPCError(GetOutputCoinError, err)
	}
	paymentInfo.Amount = uint64(amount) - realFee

	inputCoins := transaction.ConvertOutputCoinToInputCoin(outCoins)

	/******* END GET output native coins(PRV), which is used to create tx *****/
	// START create tx
	// missing flag for privacy
	// false by default
	tx := transaction.Tx{}
	err = tx.Init(
		transaction.NewTxPrivacyInitParams(&senderKeySet.PrivateKey,
			paymentInfos,
			inputCoins,
			realFee,
			hasPrivacyCoin,
			*txService.DB,
			nil, // use for prv coin -> nil is valid
			meta, nil))
	// END create tx

	if err != nil {
		return nil, NewRPCError(CreateTxDataError, err)
	}

	return &tx, nil
}

//calculateOutputCoinsByMinValue
func (txService TxService) calculateOutputCoinsByMinValue(outCoins []*privacy.OutputCoin, maxVal uint64) ([]*privacy.OutputCoin, uint64) {
	outCoinsTmp := make([]*privacy.OutputCoin, 0)
	amount := uint64(0)
	for _, outCoin := range outCoins {
		if outCoin.CoinDetails.GetValue() <= maxVal {
			outCoinsTmp = append(outCoinsTmp, outCoin)
			amount += outCoin.CoinDetails.GetValue()
		}
	}
	return outCoinsTmp, amount
}

func (txService TxService) SendRawTxWithMetadata(base58CheckDate string) (wire.Message, *common.Hash, *RPCError) {
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)
	if err != nil {
		return nil, nil, NewRPCError(RPCInvalidParamsError, err)
	}

	tx := transaction.Tx{}
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, nil, NewRPCError(JsonError, err)
	}

	hash, _, err := txService.TxMemPool.MaybeAcceptTransaction(&tx, -1)
	if err != nil {
		return nil, nil, NewRPCError(TxPoolRejectTxError, err)
	}

	Logger.log.Debugf("there is hash of transaction: %s\n", hash.String())

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
	if err != nil {
		return nil, nil, NewRPCError(UnexpectedError, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx

	return txMsg, hash, nil
}

func (txService TxService) SendRawCustomTokenTxWithMetadata(base58CheckDate string) (wire.Message, *common.Hash, *RPCError) {
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)
	if err != nil {
		return nil, nil, NewRPCError(RPCInvalidParamsError, err)
	}

	tx := transaction.TxNormalToken{}
	err = json.Unmarshal(rawTxBytes, &tx)
	fmt.Printf("%+v\n", tx)
	if err != nil {
		return nil, nil, NewRPCError(JsonError, err)
	}

	hash, _, err := txService.TxMemPool.MaybeAcceptTransaction(&tx, -1)
	if err != nil {
		return nil, nil, NewRPCError(TxPoolRejectTxError, err)
	}

	Logger.log.Debugf("there is hash of transaction: %s\n", hash.String())

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
	if err != nil {
		return nil, nil, NewRPCError(UnexpectedError, err)
	}

	txMsg.(*wire.MessageTxToken).Transaction = &tx

	return txMsg, hash, nil
}

// GetTransactionByReceiver - from keyset of receiver, we can get list tx hash which be sent to receiver
// if this keyset contain payment-addr, we can detect tx hash
// if this keyset contain viewing key, we can detect amount in tx, but can not know output in tx is spent
// because this is monitoring output to get received tx -> can not know this is a returned amount tx
func (txService TxService) GetTransactionByReceiver(keySet incognitokey.KeySet) (*jsonresult.ListReceivedTransaction, *RPCError) {
	if len(keySet.PaymentAddress.Pk) == 0 {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("Missing payment address"))
	}
	listTxsHash, err := txService.BlockChain.GetTransactionHashByReceiver(&keySet)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, errors.New("Can not find any tx"))
	}

	result := jsonresult.ListReceivedTransaction{
		ReceivedTransactions: []jsonresult.ReceivedTransaction{},
	}
	for shardID, txHashs := range listTxsHash {
		for _, txHash := range txHashs {
			item := jsonresult.ReceivedTransaction{
				FromShardID:     shardID,
				ReceivedAmounts: make(map[common.Hash]jsonresult.ReceivedInfo),
			}
			if len(keySet.ReadonlyKey.Rk) != 0 {
				_, blockHash, _, txDetail, _ := txService.BlockChain.GetTransactionByHash(txHash)
				item.LockTime = time.Unix(txDetail.GetLockTime(), 0).Format(common.DateOutputFormat)
				item.Info = base58.Base58Check{}.Encode(txDetail.GetInfo(), common.ZeroByte)
				item.BlockHash = blockHash.String()
				item.Hash = txDetail.Hash().String()

				txType := txDetail.GetType()
				item.Type = txType
				switch item.Type {
				case common.TxNormalType, common.TxRewardType, common.TxReturnStakingType:
					{
						normalTx := txDetail.(*transaction.Tx)
						item.Version = normalTx.Version
						item.IsPrivacy = normalTx.IsPrivacy()
						item.Fee = normalTx.Fee

						proof := normalTx.GetProof()
						if proof != nil {
							outputs := proof.GetOutputCoins()
							for _, output := range outputs {
								if bytes.Equal(output.CoinDetails.GetPublicKey().ToBytesS(), keySet.PaymentAddress.Pk) {
									temp := &privacy.OutputCoin{
										CoinDetails:          output.CoinDetails,
										CoinDetailsEncrypted: output.CoinDetailsEncrypted,
									}
									if temp.CoinDetailsEncrypted != nil && !temp.CoinDetailsEncrypted.IsNil() {
										// try to decrypt to get more data
										err := temp.Decrypt(keySet.ReadonlyKey)
										if err != nil {
											Logger.log.Error(err)
											continue
										}
									}
									info := jsonresult.ReceivedInfo{
										CoinDetails: jsonresult.ReceivedCoin{
											Info:      base58.Base58Check{}.Encode(temp.CoinDetails.GetInfo(), common.ZeroByte),
											PublicKey: base58.Base58Check{}.Encode(temp.CoinDetails.GetPublicKey().ToBytesS(), common.ZeroByte),
											Value:     temp.CoinDetails.GetValue(),
										},
									}
									if temp.CoinDetailsEncrypted != nil {
										info.CoinDetailsEncrypted = base58.Base58Check{}.Encode(temp.CoinDetailsEncrypted.Bytes(), common.ZeroByte)
									}
									item.ReceivedAmounts[common.PRVCoinID] = info
								}
							}
						}
					}
				case common.TxCustomTokenPrivacyType:
					{
						privacyTokenTx := txDetail.(*transaction.TxCustomTokenPrivacy)
						item.Version = privacyTokenTx.Version
						item.IsPrivacy = privacyTokenTx.IsPrivacy()
						item.PrivacyCustomTokenIsPrivacy = privacyTokenTx.TxPrivacyTokenData.TxNormal.IsPrivacy()
						item.Fee = privacyTokenTx.Fee
						item.PrivacyCustomTokenFee = privacyTokenTx.TxPrivacyTokenData.TxNormal.Fee
						item.PrivacyCustomTokenID = privacyTokenTx.TxPrivacyTokenData.PropertyID.String()
						item.PrivacyCustomTokenName = privacyTokenTx.TxPrivacyTokenData.PropertyName
						item.PrivacyCustomTokenSymbol = privacyTokenTx.TxPrivacyTokenData.PropertySymbol

						// prv proof
						proof := privacyTokenTx.GetProof()
						if proof != nil {
							outputs := proof.GetOutputCoins()
							for _, output := range outputs {
								if bytes.Equal(output.CoinDetails.GetPublicKey().ToBytesS(), keySet.PaymentAddress.Pk) {
									temp := &privacy.OutputCoin{
										CoinDetails:          output.CoinDetails,
										CoinDetailsEncrypted: output.CoinDetailsEncrypted,
									}
									if temp.CoinDetailsEncrypted != nil && !temp.CoinDetailsEncrypted.IsNil() {
										// try to decrypt to get more data
										err := temp.Decrypt(keySet.ReadonlyKey)
										if err != nil {
											Logger.log.Error(err)
											continue
										}
									}
									info := jsonresult.ReceivedInfo{
										CoinDetails: jsonresult.ReceivedCoin{
											Info:      base58.Base58Check{}.Encode(temp.CoinDetails.GetInfo(), common.ZeroByte),
											PublicKey: base58.Base58Check{}.Encode(temp.CoinDetails.GetPublicKey().ToBytesS(), common.ZeroByte),
											Value:     temp.CoinDetails.GetValue(),
										},
									}
									if temp.CoinDetailsEncrypted != nil {
										info.CoinDetailsEncrypted = base58.Base58Check{}.Encode(temp.CoinDetailsEncrypted.Bytes(), common.ZeroByte)
									}
									item.ReceivedAmounts[common.PRVCoinID] = info
								}
							}
						}

						// token proof
						proof = privacyTokenTx.TxPrivacyTokenData.TxNormal.GetProof()
						if proof != nil {
							outputs := proof.GetOutputCoins()
							for _, output := range outputs {
								if bytes.Equal(output.CoinDetails.GetPublicKey().ToBytesS(), keySet.PaymentAddress.Pk) {
									temp := &privacy.OutputCoin{
										CoinDetails:          output.CoinDetails,
										CoinDetailsEncrypted: output.CoinDetailsEncrypted,
									}
									if temp.CoinDetailsEncrypted != nil && !temp.CoinDetailsEncrypted.IsNil() {
										// try to decrypt to get more data
										err := temp.Decrypt(keySet.ReadonlyKey)
										if err != nil {
											Logger.log.Error(err)
											continue
										}
									}
									info := jsonresult.ReceivedInfo{
										CoinDetails: jsonresult.ReceivedCoin{
											Info:      base58.Base58Check{}.Encode(temp.CoinDetails.GetInfo(), common.ZeroByte),
											PublicKey: base58.Base58Check{}.Encode(temp.CoinDetails.GetPublicKey().ToBytesS(), common.ZeroByte),
											Value:     temp.CoinDetails.GetValue(),
										},
									}
									if temp.CoinDetailsEncrypted != nil {
										info.CoinDetailsEncrypted = base58.Base58Check{}.Encode(temp.CoinDetailsEncrypted.Bytes(), common.ZeroByte)
									}
									item.ReceivedAmounts[privacyTokenTx.TxPrivacyTokenData.PropertyID] = info
								}
							}
						}
					}
				}
			}
			result.ReceivedTransactions = append(result.ReceivedTransactions, item)
			sort.Slice(result.ReceivedTransactions, func(i, j int) bool {
				return result.ReceivedTransactions[i].LockTime > result.ReceivedTransactions[j].LockTime
			})
		}
	}
	return &result, nil
}
