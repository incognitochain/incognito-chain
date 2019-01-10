package rpcserver

import (
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
)

func (self RpcServer) chooseOutsCoinByKeyset(paymentInfos []*privacy.PaymentInfo, estimateFeeCoinPerKb int64, numBlock uint64, keyset *wallet.Key, chainIdSender byte) ([]*privacy.InputCoin, uint64, *RPCError) {
	if numBlock == 0 {
		numBlock = 8
	}
	// calculate total amount to send
	totalAmmount := uint64(0)
	for _, receiver := range paymentInfos {
		totalAmmount += receiver.Amount
	}

	// get list outputcoins tx
	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	outCoins, err := self.config.BlockChain.GetListOutputCoinsByKeyset(&keyset.KeySet, chainIdSender, constantTokenID)
	if err != nil {
		return nil, 0, NewRPCError(ErrUnexpected, err)
	}
	if len(outCoins) == 0 && totalAmmount > 0 {
		return nil, 0, NewRPCError(ErrUnexpected, nil)
	}
	// Use Knapsack to get candiate output coin
	candidateOutputCoins, outCoins, candidateOutputCoinAmount, err := self.chooseBestOutCoinsToSpent(outCoins, totalAmmount)
	if err != nil {
		return nil, 0, NewRPCError(ErrUnexpected, err)
	}

	// check real fee(nano constant) per tx
	realFee := self.estimateFee(estimateFeeCoinPerKb, candidateOutputCoins, paymentInfos, chainIdSender, numBlock)
	if len(outCoins) == 0 {
		realFee = 0
	}
	needToPayFee := int64((totalAmmount + realFee) - candidateOutputCoinAmount)

	// if not enough to pay fee
	if needToPayFee > 0 {
		if len(outCoins) > 0 {
			candidateOutputCoinsForFee, _, _, err1 := self.chooseBestOutCoinsToSpent(outCoins, uint64(needToPayFee))
			if err != nil {
				return nil, 0, NewRPCError(ErrUnexpected, err1)
			}
			candidateOutputCoins = append(candidateOutputCoins, candidateOutputCoinsForFee...)
		}
	}
	// convert to inputcoins
	inputCoins := transaction.ConvertOutputCoinToInputCoin(candidateOutputCoins)
	return inputCoins, realFee, nil
}

func (self RpcServer) buildRawTransaction(params interface{}, meta metadata.Metadata) (*transaction.Tx, *RPCError) {
	Logger.log.Infof("Params: \n%+v\n\n\n", params)

	/******* START Fetch all params to ******/
	// all params
	arrayParams := common.InterfaceSlice(params)

	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	lastByte := senderKey.KeySet.PaymentAddress.Pk[len(senderKey.KeySet.PaymentAddress.Pk)-1]
	chainIdSender, err := common.GetTxSenderChain(lastByte)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	fmt.Printf("Done param #1: keyset: %+v\n", senderKey.KeySet)

	// param #2: list receiver
	receiversParam := make(map[string]interface{})
	if arrayParams[1] != nil {
		receiversParam = arrayParams[1].(map[string]interface{})
	}
	paymentInfos := make([]*privacy.PaymentInfo, 0)
	for pubKeyStr, amount := range receiversParam {
		receiverPubKey, err := wallet.Base58CheckDeserialize(pubKeyStr)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		paymentInfo := &privacy.PaymentInfo{
			Amount:         uint64(amount.(float64)),
			PaymentAddress: receiverPubKey.KeySet.PaymentAddress,
		}
		paymentInfos = append(paymentInfos, paymentInfo)
	}

	// param #3: estimation fee nano constant per kb
	estimateFeeCoinPerKb := int64(arrayParams[2].(float64))

	// param #4: hasPrivacy flag: 1 or -1
	hasPrivacy := int(arrayParams[3].(float64)) > 0
	/********* END Fetch all params to *******/

	/******* START choose output coins constant, which is used to create tx *****/
	inputCoins, realFee, err := self.chooseOutsCoinByKeyset(paymentInfos, estimateFeeCoinPerKb, 0, senderKey, chainIdSender)
	if err.(*RPCError) != nil {
		return nil, err.(*RPCError)
	}
	/******* END GET output coins constant, which is used to create tx *****/

	// START create tx
	// missing flag for privacy
	// false by default
	fmt.Printf("#inputCoins: %d\n", len(inputCoins))
	tx := transaction.Tx{}
	err = tx.Init(
		&senderKey.KeySet.PrivateKey,
		paymentInfos,
		inputCoins,
		realFee,
		hasPrivacy,
		*self.config.Database,
		nil, // use for constant coin -> nil is valid
		meta,
	)
	// END create tx

	if err.(*transaction.TransactionError) != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return &tx, nil
}

// buildRawCustomTokenTransaction ...
func (self RpcServer) buildRawCustomTokenTransaction(
	params interface{},
	metaData metadata.Metadata,
) (*transaction.TxCustomToken, error) {
	// all params
	arrayParams := common.InterfaceSlice(params)

	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, err
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	lastByte := senderKey.KeySet.PaymentAddress.Pk[len(senderKey.KeySet.PaymentAddress.Pk)-1]
	chainIdSender, err := common.GetTxSenderChain(lastByte)
	if err != nil {
		return nil, err
	}

	// param #2: estimation fee coin per kb
	estimateFeeCoinPerKb := int64(arrayParams[1].(float64))

	// param #3: hasPrivacy flag
	hasPrivacy := int(arrayParams[2].(float64)) > 0

	// param #4: token params
	tokenParamsRaw := arrayParams[3].(map[string]interface{})
	tokenParams := &transaction.CustomTokenParamTx{
		PropertyID:     tokenParamsRaw["TokenID"].(string),
		PropertyName:   tokenParamsRaw["TokenName"].(string),
		PropertySymbol: tokenParamsRaw["TokenSymbol"].(string),
		TokenTxType:    int(tokenParamsRaw["TokenTxType"].(float64)),
		Amount:         uint64(tokenParamsRaw["TokenAmount"].(float64)),
	}
	voutsAmount := int64(0)
	tokenParams.Receiver, voutsAmount = transaction.CreateCustomTokenReceiverArray(tokenParamsRaw["TokenReceivers"])
	// get list custom token
	listCustomTokens, err := self.config.BlockChain.ListCustomToken()
	switch tokenParams.TokenTxType {
	case transaction.CustomTokenTransfer:
		{
			tokenID, err := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}

			if _, ok := listCustomTokens[*tokenID]; !ok {
				return nil, NewRPCError(ErrUnexpected, errors.New("Invalid Token ID"))
			}

			unspentTxTokenOuts, err := self.config.BlockChain.GetUnspentTxCustomTokenVout(senderKey.KeySet, tokenID)
			Logger.log.Info("buildRawCustomTokenTransaction ", unspentTxTokenOuts)
			if err != nil {
				return nil, err
			}
			if len(unspentTxTokenOuts) == 0 {
				return nil, errors.Wrap(errors.New("Balance of token is zero"), "")
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
				signature, err := senderKey.KeySet.Sign(out.Hash()[:])
				if err != nil {
					return nil, NewRPCError(ErrUnexpected, err)
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
				return nil, NewRPCError(ErrUnexpected, errors.New("Init with wrong max amount of property"))
			}
		}
	}

	/******* START choose output coins constant, which is used to create tx *****/
	paymentInfos := []*privacy.PaymentInfo{}
	inputCoins, realFee, err := self.chooseOutsCoinByKeyset(paymentInfos, estimateFeeCoinPerKb, 0, senderKey, chainIdSender)
	if err.(*RPCError) != nil {
		return nil, err.(*RPCError)
	}
	if len(paymentInfos) == 0 && realFee == 0 {
		hasPrivacy = false
	}
	/******* END GET output coins constant, which is used to create tx *****/

	tx := &transaction.TxCustomToken{}
	err = tx.Init(
		&senderKey.KeySet.PrivateKey,
		nil,
		inputCoins,
		realFee,
		tokenParams,
		listCustomTokens,
		metaData,
		hasPrivacy,
	)
	if err.(*transaction.TransactionError) != nil {
		return nil, err
	}
	return tx, nil
}

// buildRawCustomTokenTransaction ...
func (self RpcServer) buildRawPrivacyCustomTokenTransaction(
	params interface{},
) (*transaction.TxCustomTokenPrivacy, error) {
	// all params
	arrayParams := common.InterfaceSlice(params)

	/****** START FEtch data from params *********/
	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	lastByte := senderKey.KeySet.PaymentAddress.Pk[len(senderKey.KeySet.PaymentAddress.Pk)-1]
	chainIdSender, err := common.GetTxSenderChain(lastByte)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// param #2: list receiver
	receiversParam := make(map[string]interface{})
	if arrayParams[1] != nil {
		receiversParam = arrayParams[1].(map[string]interface{})
	}
	paymentInfos := make([]*privacy.PaymentInfo, 0)
	for pubKeyStr, amount := range receiversParam {
		receiverPubKey, err := wallet.Base58CheckDeserialize(pubKeyStr)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		paymentInfo := &privacy.PaymentInfo{
			Amount:         uint64(amount.(float64)),
			PaymentAddress: receiverPubKey.KeySet.PaymentAddress,
		}
		paymentInfos = append(paymentInfos, paymentInfo)
	}

	// param #3: estimation fee coin per kb
	estimateFeeCoinPerKb := int64(arrayParams[2].(float64))

	// param #4: hasPrivacy flag for constant
	hasPrivacyConst := int(arrayParams[3].(float64)) > 0

	// param #5: token params
	tokenParamsRaw := arrayParams[4].(map[string]interface{})
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID:     tokenParamsRaw["TokenID"].(string),
		PropertyName:   tokenParamsRaw["TokenName"].(string),
		PropertySymbol: tokenParamsRaw["TokenSymbol"].(string),
		TokenTxType:    int(tokenParamsRaw["TokenTxType"].(float64)),
		Amount:         uint64(tokenParamsRaw["TokenAmount"].(float64)),
		TokenInput:     nil,
	}
	voutsAmount := int64(0)
	tokenParams.Receiver, voutsAmount = transaction.CreateCustomTokenPrivacyReceiverArray(tokenParamsRaw["TokenReceivers"])

	// get list custom token
	listCustomTokens, err := self.config.BlockChain.ListPrivacyCustomToken()

	switch tokenParams.TokenTxType {
	case transaction.CustomTokenTransfer:
		{
			tokenID, err := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			if _, ok := listCustomTokens[*tokenID]; !ok {
				return nil, NewRPCError(ErrUnexpected, errors.New("Invalid Token ID"))
			}
			outputTokens, err := self.config.BlockChain.GetListOutputCoinsByKeyset(&senderKey.KeySet, chainIdSender, tokenID)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			candidateOutputTokens, outputTokens, _, err := self.chooseBestOutCoinsToSpent(outputTokens, uint64(voutsAmount))
			intputToken := transaction.ConvertOutputCoinToInputCoin(candidateOutputTokens)
			tokenParams.TokenInput = intputToken
		}
	case transaction.CustomTokenInit:
		{
			if tokenParams.Receiver[0].Amount != tokenParams.Amount { // Init with wrong max amount of custom token
				return nil, NewRPCError(ErrUnexpected, errors.New("Init with wrong max amount of property"))
			}
		}
	}
	/****** END FEtch data from params *********/

	/******* START choose output coins constant, which is used to create tx *****/
	inputCoins, realFee, err := self.chooseOutsCoinByKeyset(paymentInfos, estimateFeeCoinPerKb, 0, senderKey, chainIdSender)
	if err.(*RPCError) != nil {
		return nil, err.(*RPCError)
	}
	if len(paymentInfos) == 0 && realFee == 0 {
		hasPrivacyConst = false
	}
	/******* END GET output coins constant, which is used to create tx *****/

	tx := &transaction.TxCustomTokenPrivacy{}
	err = tx.Init(
		&senderKey.KeySet.PrivateKey,
		nil,
		inputCoins,
		realFee,
		tokenParams,
		listCustomTokens,
		*self.config.Database,
		hasPrivacyConst,
	)

	if err.(*transaction.TransactionError) != nil {
		return nil, err
	}

	return tx, err
}

func (self RpcServer) estimateFee(defaultFee int64, candidateOutputCoins []*privacy.OutputCoin, paymentInfos []*privacy.PaymentInfo, chainID byte, numBlock uint64) uint64 {
	if numBlock == 0 {
		numBlock = 10
	}
	// check real fee(nano constant) per tx
	var realFee uint64
	estimateFeeCoinPerKb := uint64(0)
	if defaultFee == -1 {
		temp, _ := self.config.FeeEstimator[chainID].EstimateFee(numBlock)
		estimateFeeCoinPerKb = uint64(temp)
	}
	if estimateFeeCoinPerKb == 0 {
		estimateFeeCoinPerKb = self.config.BlockChain.GetFeePerKbTx()
	}
	estimateFeeCoinPerKb += uint64(self.config.Wallet.Config.IncrementalFee)
	estimateTxSizeInKb := transaction.EstimateTxSize(candidateOutputCoins, nil)
	realFee = uint64(estimateFeeCoinPerKb) * uint64(estimateTxSizeInKb)
	return realFee
}

// chooseBestOutCoinsToSpent returns list of unspent coins for spending with amount
func (self RpcServer) chooseBestOutCoinsToSpent(outCoins []*privacy.OutputCoin, amount uint64) (resultOutputCoins []*privacy.OutputCoin, remainOutputCoins []*privacy.OutputCoin, totalResultOutputCoinAmount uint64, err error) {
	resultOutputCoins = make([]*privacy.OutputCoin, 0)
	remainOutputCoins = make([]*privacy.OutputCoin, 0)
	totalResultOutputCoinAmount = uint64(0)

	if amount == 0 || len(outCoins) == 0 {
		return resultOutputCoins, remainOutputCoins, totalResultOutputCoinAmount, nil
	}

	// Calculate sum of all output coins' value
	sumValue := uint64(0)
	values := make([]uint64, 0)
	for _, outCoin := range outCoins {
		sumValue += outCoin.CoinDetails.Value
		values = append(values, outCoin.CoinDetails.Value)
	}

	// target
	target := int64(sumValue - amount)
	if target < 0 {
		return nil, remainOutputCoins, uint64(0), errors.Wrap(errors.New("Not enough coin"), "")
	}
	choices := privacy.Knapsack(values, uint64(target))
	for i, choice := range choices {
		if !choice {
			totalResultOutputCoinAmount += outCoins[i].CoinDetails.Value
			resultOutputCoins = append(resultOutputCoins, outCoins[i])
		} else {
			remainOutputCoins = append(remainOutputCoins, outCoins[i])
		}
	}
	return resultOutputCoins, remainOutputCoins, totalResultOutputCoinAmount, nil
}
