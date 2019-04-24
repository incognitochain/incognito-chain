package rpcserver

import (
	"sort"
	
	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/pkg/errors"
)

func (rpcServer RpcServer) chooseOutsCoinByKeyset(paymentInfos []*privacy.PaymentInfo,
	estimateFeeCoinPerKb int64, numBlock uint64, keyset *cashec.KeySet, shardIDSender byte,
	hasPrivacy bool,
	metadata metadata.Metadata,
	customTokenParams *transaction.CustomTokenParamTx,
	privacyCustomTokenParams *transaction.CustomTokenPrivacyParamTx,
) ([]*privacy.InputCoin, uint64, *RPCError) {
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
	outCoins, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(keyset, shardIDSender, constantTokenID)
	if err != nil {
		return nil, 0, NewRPCError(ErrGetOutputCoin, err)
	}
	// remove out coin in mem pool
	outCoins, err = rpcServer.filterMemPoolOutCoinsToSpent(outCoins)
	if err != nil {
		return nil, 0, NewRPCError(ErrGetOutputCoin, err)
	}
	if len(outCoins) == 0 && totalAmmount > 0 {
		return nil, 0, NewRPCError(ErrGetOutputCoin, errors.New("not enough output coin"))
	}
	// Use Knapsack to get candiate output coin
	candidateOutputCoins, outCoins, candidateOutputCoinAmount, err := rpcServer.chooseBestOutCoinsToSpent(outCoins, totalAmmount)
	if err != nil {
		return nil, 0, NewRPCError(ErrGetOutputCoin, err)
	}
	// refund out put for sender
	overBalanceAmount := candidateOutputCoinAmount - totalAmmount
	if overBalanceAmount > 0 {
		// add more into output for estimate fee
		paymentInfos = append(paymentInfos, &privacy.PaymentInfo{
			PaymentAddress: keyset.PaymentAddress,
			Amount:         overBalanceAmount,
		})
	}

	// check real fee(nano constant) per tx
	realFee, _, _ := rpcServer.estimateFee(estimateFeeCoinPerKb, candidateOutputCoins,
		paymentInfos, shardIDSender, numBlock, hasPrivacy,
		metadata, customTokenParams,
		privacyCustomTokenParams)
	needToPayFee := int64((totalAmmount + realFee) - candidateOutputCoinAmount)
	// if not enough to pay fee
	if needToPayFee > 0 {
		if len(outCoins) > 0 {
			candidateOutputCoinsForFee, _, _, err1 := rpcServer.chooseBestOutCoinsToSpent(outCoins, uint64(needToPayFee))
			if err != nil {
				return nil, 0, NewRPCError(ErrGetOutputCoin, err1)
			}
			candidateOutputCoins = append(candidateOutputCoins, candidateOutputCoinsForFee...)
		}
	}
	// convert to inputcoins
	inputCoins := transaction.ConvertOutputCoinToInputCoin(candidateOutputCoins)
	return inputCoins, realFee, nil
}

func (rpcServer RpcServer) makeArrayInputCoinHashHs(inputCoins []*privacy.InputCoin) []common.Hash {
	inCoinHs := make([]common.Hash, 0)
	for _, inCoin := range inputCoins {
		hash := inCoin.CoinDetails.HashH()
		if hash != nil {
			inCoinHs = append(inCoinHs, *hash)
		}
	}
	return inCoinHs
}

func (rpcServer RpcServer) buildRawTransaction(params interface{}, meta metadata.Metadata) (*transaction.Tx, *RPCError) {
	Logger.log.Infof("Params: \n%+v\n\n\n", params)

	/******* START Fetch all component to ******/
	// all component
	arrayParams := common.InterfaceSlice(params)

	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKeySet, err := rpcServer.GetKeySetFromPrivateKeyParams(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrInvalidSenderPrivateKey, err)
	}
	lastByte := senderKeySet.PaymentAddress.Pk[len(senderKeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	//fmt.Printf("Done param #1: keyset: %+v\n", senderKeySet)

	// param #2: list receiver
	receiversPaymentAddressStrParam := make(map[string]interface{})
	if arrayParams[1] != nil {
		receiversPaymentAddressStrParam = arrayParams[1].(map[string]interface{})
	}
	paymentInfos := make([]*privacy.PaymentInfo, 0)
	for paymentAddressStr, amount := range receiversPaymentAddressStrParam {
		keyWalletReceiver, err := wallet.Base58CheckDeserialize(paymentAddressStr)
		if err != nil {
			return nil, NewRPCError(ErrInvalidReceiverPaymentAddress, err)
		}
		paymentInfo := &privacy.PaymentInfo{
			Amount:         uint64(amount.(float64)),
			PaymentAddress: keyWalletReceiver.KeySet.PaymentAddress,
		}
		paymentInfos = append(paymentInfos, paymentInfo)
	}

	// param #3: estimation fee nano constant per kb
	estimateFeeCoinPerKb := int64(arrayParams[2].(float64))

	// param #4: hasPrivacy flag: 1 or -1
	hasPrivacy := int(arrayParams[3].(float64)) > 0
	/********* END Fetch all component to *******/

	/******* START choose output coins constant, which is used to create tx *****/
	inputCoins, realFee, err1 := rpcServer.chooseOutsCoinByKeyset(paymentInfos, estimateFeeCoinPerKb, 0, senderKeySet, shardIDSender, hasPrivacy, meta, nil, nil)
	if err1 != nil {
		return nil, err1
	}

	// build hash array for input coin
	inputCoinHs := rpcServer.makeArrayInputCoinHashHs(inputCoins)

	/******* END GET output coins constant, which is used to create tx *****/

	// START create tx
	// missing flag for privacy
	// false by default
	//fmt.Printf("#inputCoins: %d\n", len(inputCoins))
	tx := transaction.Tx{}
	err = tx.Init(
		&senderKeySet.PrivateKey,
		paymentInfos,
		inputCoins,
		realFee,
		hasPrivacy,
		*rpcServer.config.Database,
		nil, // use for constant coin -> nil is valid
		meta,
	)
	// END create tx

	if err.(*transaction.TransactionError) != nil {
		return nil, NewRPCError(ErrCreateTxData, err)
	}

	// pool inCoinsH
	txHash := tx.Hash()
	if txHash != nil {
		rpcServer.config.TxMemPool.PrePoolTxCoinHashH(*txHash, inputCoinHs)
	}

	return &tx, nil
}

func (rpcServer RpcServer) buildCustomTokenParam(tokenParamsRaw map[string]interface{}, senderKeySet *cashec.KeySet) (*transaction.CustomTokenParamTx, map[common.Hash]transaction.TxCustomToken, *RPCError) {
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
	//listCustomTokens, err := rpcServer.config.BlockChain.ListCustomToken()
	//if err != nil {
	//	return nil, nil, NewRPCError(ErrListCustomTokenNotFound, err)
	//}
	switch tokenParams.TokenTxType {
	case transaction.CustomTokenTransfer:
		{
			tokenID, err := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
			if err != nil {
				return nil, nil, NewRPCError(ErrRPCInvalidParams, errors.Wrap(err, "Token ID is invalid"))
			}

			//if _, ok := listCustomTokens[*tokenID]; !ok {
			//	return nil, nil, NewRPCError(ErrRPCInvalidParams, errors.New("Invalid Token ID"))
			//}

			existed := rpcServer.config.BlockChain.CustomTokenIDExisted(tokenID)
			if !existed {
				return nil, nil, NewRPCError(ErrRPCInvalidParams, errors.New("Invalid Token ID"))
			}

			unspentTxTokenOuts, err := rpcServer.config.BlockChain.GetUnspentTxCustomTokenVout(*senderKeySet, tokenID)
			Logger.log.Info("buildRawCustomTokenTransaction ", unspentTxTokenOuts)
			if err != nil {
				return nil, nil, NewRPCError(ErrGetOutputCoin, errors.New("Token out invalid"))
			}
			if len(unspentTxTokenOuts) == 0 {
				return nil, nil, NewRPCError(ErrGetOutputCoin, errors.New("Token out invalid"))
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
					return nil, nil, NewRPCError(ErrCanNotSign, err)
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
				return nil, nil, NewRPCError(ErrUnexpected, errors.New("Init with wrong max amount of property"))
			}
		}
	}
	//return tokenParams, listCustomTokens, nil
	return tokenParams, nil, nil
}

// buildRawCustomTokenTransaction ...
func (rpcServer RpcServer) buildRawCustomTokenTransaction(
	params interface{},
	metaData metadata.Metadata,
) (*transaction.TxCustomToken, *RPCError) {
	// all params
	arrayParams := common.InterfaceSlice(params)

	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKeySet, err := rpcServer.GetKeySetFromPrivateKeyParams(senderKeyParam.(string))
	if err != nil {
		return nil, err.(*RPCError)
	}
	lastByte := senderKeySet.PaymentAddress.Pk[len(senderKeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)

	// param #2: list receiver
	receiversPaymentAddressParam := make(map[string]interface{})
	if arrayParams[1] != nil {
		receiversPaymentAddressParam = arrayParams[1].(map[string]interface{})
	}
	paymentInfos := make([]*privacy.PaymentInfo, 0)
	for paymentAddressStr, amount := range receiversPaymentAddressParam {
		keyWalletReceiver, err := wallet.Base58CheckDeserialize(paymentAddressStr)
		if err != nil {
			return nil, NewRPCError(ErrInvalidReceiverPaymentAddress, err)
		}
		paymentInfo := &privacy.PaymentInfo{
			Amount:         uint64(amount.(float64)),
			PaymentAddress: keyWalletReceiver.KeySet.PaymentAddress,
		}
		paymentInfos = append(paymentInfos, paymentInfo)
	}

	// param #3: estimation fee coin per kb
	estimateFeeCoinPerKb := int64(arrayParams[2].(float64))

	// param #4: hasPrivacy flag
	hasPrivacy := int(arrayParams[3].(float64)) > 0

	// param #5: token params
	tokenParamsRaw := arrayParams[4].(map[string]interface{})
	tokenParams, listCustomTokens, err := rpcServer.buildCustomTokenParam(tokenParamsRaw, senderKeySet)
	_ = listCustomTokens
	if err.(*RPCError) != nil {
		return nil, err.(*RPCError)
	}
	/******* START choose output coins constant, which is used to create tx *****/
	inputCoins, realFee, err := rpcServer.chooseOutsCoinByKeyset(paymentInfos, estimateFeeCoinPerKb, 0,
		senderKeySet, shardIDSender, hasPrivacy,
		metaData, tokenParams, nil)
	if err.(*RPCError) != nil {
		return nil, err.(*RPCError)
	}
	if len(paymentInfos) == 0 && realFee == 0 {
		hasPrivacy = false
	}
	/******* END GET output coins constant, which is used to create tx *****/

	// build hash array for input coin
	inputCoinHs := rpcServer.makeArrayInputCoinHashHs(inputCoins)

	tx := &transaction.TxCustomToken{}
	err = tx.Init(
		&senderKeySet.PrivateKey,
		nil,
		inputCoins,
		realFee,
		tokenParams,
		//listCustomTokens,
		*rpcServer.config.Database,
		metaData,
		hasPrivacy,
		shardIDSender,
	)
	if err.(*transaction.TransactionError) != nil {
		return nil, NewRPCError(ErrCreateTxData, err)
	}

	// pool inCoinsH
	txHash := tx.Hash()
	if txHash != nil {
		rpcServer.config.TxMemPool.PrePoolTxCoinHashH(*txHash, inputCoinHs)
	}

	return tx, nil
}

func (rpcServer RpcServer) buildPrivacyCustomTokenParam(tokenParamsRaw map[string]interface{}, senderKeySet *cashec.KeySet, shardIDSender byte) (*transaction.CustomTokenPrivacyParamTx, map[common.Hash]transaction.TxCustomTokenPrivacy, map[common.Hash]blockchain.CrossShardTokenPrivacyMetaData, *RPCError) {
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
	/*listCustomTokens, listCustomTokenCrossShard, err := rpcServer.config.BlockChain.ListPrivacyCustomToken()
	if err != nil {
		return nil, nil, nil, NewRPCError(ErrListCustomTokenNotFound, err)
	}*/
	switch tokenParams.TokenTxType {
	case transaction.CustomTokenTransfer:
		{
			tokenID, err := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
			if err != nil {
				return nil, nil, nil, NewRPCError(ErrRPCInvalidParams, err)
			}
			//if _, ok := listCustomTokens[*tokenID]; !ok {
			//	if _, ok := listCustomTokenCrossShard[*tokenID]; !ok {
			//		return nil, nil, nil, NewRPCError(ErrRPCInvalidParams, errors.New("Invalid Token ID"))
			//	}
			//}
			existed := rpcServer.config.BlockChain.PrivacyCustomTokenIDExisted(tokenID)
			existedCrossShard := rpcServer.config.BlockChain.PrivacyCustomTokenIDCrossShardExisted(tokenID)
			if !existed && !existedCrossShard {
				return nil, nil, nil, NewRPCError(ErrRPCInvalidParams, errors.New("Invalid Token ID"))
			}
			outputTokens, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(senderKeySet, shardIDSender, tokenID)
			if err != nil {
				return nil, nil, nil, NewRPCError(ErrGetOutputCoin, err)
			}
			candidateOutputTokens, _, _, err := rpcServer.chooseBestOutCoinsToSpent(outputTokens, uint64(voutsAmount))
			if err != nil {
				return nil, nil, nil, NewRPCError(ErrGetOutputCoin, err)
			}
			intputToken := transaction.ConvertOutputCoinToInputCoin(candidateOutputTokens)
			tokenParams.TokenInput = intputToken
		}
	case transaction.CustomTokenInit:
		{
			if tokenParams.Receiver[0].Amount != tokenParams.Amount { // Init with wrong max amount of custom token
				return nil, nil, nil, NewRPCError(ErrRPCInvalidParams, errors.New("Init with wrong max amount of property"))
			}
		}
	}
	//return tokenParams, listCustomTokens, listCustomTokenCrossShard, nil
	return tokenParams, nil, nil, nil
}

// buildRawCustomTokenTransaction ...
func (rpcServer RpcServer) buildRawPrivacyCustomTokenTransaction(
	params interface{},
) (*transaction.TxCustomTokenPrivacy, *RPCError) {
	// all component
	arrayParams := common.InterfaceSlice(params)

	/****** START FEtch data from component *********/
	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKeySet, err := rpcServer.GetKeySetFromPrivateKeyParams(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrInvalidSenderPrivateKey, err)
	}
	lastByte := senderKeySet.PaymentAddress.Pk[len(senderKeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)

	// param #2: list receiver
	receiversPaymentAddressStrParam := make(map[string]interface{})
	if arrayParams[1] != nil {
		receiversPaymentAddressStrParam = arrayParams[1].(map[string]interface{})
	}
	paymentInfos := make([]*privacy.PaymentInfo, 0)
	for paymentAddressStr, amount := range receiversPaymentAddressStrParam {
		keyWalletReceiver, err := wallet.Base58CheckDeserialize(paymentAddressStr)
		if err != nil {
			return nil, NewRPCError(ErrInvalidReceiverPaymentAddress, err)
		}
		paymentInfo := &privacy.PaymentInfo{
			Amount:         uint64(amount.(float64)),
			PaymentAddress: keyWalletReceiver.KeySet.PaymentAddress,
		}
		paymentInfos = append(paymentInfos, paymentInfo)
	}

	// param #3: estimation fee coin per kb
	estimateFeeCoinPerKb := int64(arrayParams[2].(float64))

	// param #4: hasPrivacy flag for constant
	hasPrivacyConst := int(arrayParams[3].(float64)) > 0

	// param #5: token component
	tokenParamsRaw := arrayParams[4].(map[string]interface{})
	tokenParams, listCustomTokens, listCustomTokenCrossShard, err := rpcServer.buildPrivacyCustomTokenParam(tokenParamsRaw, senderKeySet, shardIDSender)
	/*mapCustomTokenCrossShard := make(map[common.Hash]bool)
	for _, customTokenCrossShard := range listCustomTokenCrossShard {
		mapCustomTokenCrossShard[customTokenCrossShard.TokenID] = true
	}*/
	_ = listCustomTokenCrossShard
	_ = listCustomTokens
	if err.(*RPCError) != nil {
		return nil, err.(*RPCError)
	}
	/****** END FEtch data from params *********/

	/******* START choose output coins constant, which is used to create tx *****/
	inputCoins, realFee, err := rpcServer.chooseOutsCoinByKeyset(paymentInfos,
		estimateFeeCoinPerKb, 0, senderKeySet,
		shardIDSender, hasPrivacyConst, nil,
		nil, tokenParams)
	if err.(*RPCError) != nil {
		return nil, err.(*RPCError)
	}
	if len(paymentInfos) == 0 && realFee == 0 {
		hasPrivacyConst = false
	}
	/******* END GET output coins constant, which is used to create tx *****/

	// build hash array for input coin
	inputCoinHs := rpcServer.makeArrayInputCoinHashHs(inputCoins)

	tx := &transaction.TxCustomTokenPrivacy{}
	err = tx.Init(
		&senderKeySet.PrivateKey,
		nil,
		inputCoins,
		realFee,
		tokenParams,
		//listCustomTokens,
		*rpcServer.config.Database,
		hasPrivacyConst,
		shardIDSender,
		//mapCustomTokenCrossShard,
	)

	if err.(*transaction.TransactionError) != nil {
		return nil, NewRPCError(ErrCreateTxData, err)
	}

	// pool inCoinsH
	txHash := tx.Hash()
	if txHash != nil {
		rpcServer.config.TxMemPool.PrePoolTxCoinHashH(*txHash, inputCoinHs)
	}

	return tx, nil
}

func (rpcServer RpcServer) estimateFeeWithEstimator(defaultFee int64, shardID byte, numBlock uint64) uint64 {
	estimateFeeCoinPerKb := uint64(0)
	if defaultFee == -1 {
		if _, ok := rpcServer.config.FeeEstimator[shardID]; ok {
			temp, _ := rpcServer.config.FeeEstimator[shardID].EstimateFee(numBlock)
			estimateFeeCoinPerKb = uint64(temp)
		}
		if estimateFeeCoinPerKb == 0 {
			estimateFeeCoinPerKb = rpcServer.config.BlockChain.GetFeePerKbTx()
		}
	} else {
		estimateFeeCoinPerKb = uint64(defaultFee)
	}
	return estimateFeeCoinPerKb
}

func (rpcServer RpcServer) estimateFee(defaultFee int64, candidateOutputCoins []*privacy.OutputCoin,
	paymentInfos []*privacy.PaymentInfo, shardID byte,
	numBlock uint64, hasPrivacy bool,
	metadata metadata.Metadata,
	customTokenParams *transaction.CustomTokenParamTx,
	privacyCustomTokenParams *transaction.CustomTokenPrivacyParamTx) (uint64, uint64, uint64) {
	if numBlock == 0 {
		numBlock = 10
	}
	// check real fee(nano constant) per tx
	var realFee uint64
	estimateFeeCoinPerKb := uint64(0)
	estimateTxSizeInKb := uint64(0)

	estimateFeeCoinPerKb = rpcServer.estimateFeeWithEstimator(defaultFee, shardID, numBlock)

	if rpcServer.config.Wallet != nil {
		estimateFeeCoinPerKb += uint64(rpcServer.config.Wallet.GetConfig().IncrementalFee)
	}
	estimateTxSizeInKb = transaction.EstimateTxSize(candidateOutputCoins, paymentInfos, hasPrivacy, metadata, customTokenParams, privacyCustomTokenParams)
	realFee = uint64(estimateFeeCoinPerKb) * uint64(estimateTxSizeInKb)
	return realFee, estimateFeeCoinPerKb, estimateTxSizeInKb
}

func (rpcServer RpcServer) filterMemPoolOutCoinsToSpent(outCoins []*privacy.OutputCoin) (remainOutputCoins []*privacy.OutputCoin, err error) {
	remainOutputCoins = make([]*privacy.OutputCoin, 0)
	for _, outCoin := range outCoins {
		hash := outCoin.CoinDetails.HashH()
		if hash != nil {
			if rpcServer.config.TxMemPool.ValidateCoinHashH(*hash) == nil {
				remainOutputCoins = append(remainOutputCoins, outCoin)
			}
		}
	}
	return remainOutputCoins, nil
}

// chooseBestOutCoinsToSpent returns list of unspent coins for spending with amount
func (rpcServer RpcServer) chooseBestOutCoinsToSpent(outCoins []*privacy.OutputCoin, amount uint64) (resultOutputCoins []*privacy.OutputCoin, remainOutputCoins []*privacy.OutputCoin, totalResultOutputCoinAmount uint64, err error) {
	resultOutputCoins = make([]*privacy.OutputCoin, 0)
	remainOutputCoins = make([]*privacy.OutputCoin, 0)
	totalResultOutputCoinAmount = uint64(0)

	// just choose output coins have value less than amount for Knapsack algorithm
	sumValueKnapsack := uint64(0)
	valuesKnapsack := make([]uint64, 0)
	outCoinKnapsack := make([]*privacy.OutputCoin, 0)
	outCoinUnknapsack := make([]*privacy.OutputCoin, 0)

	for _, outCoin := range outCoins {
		if outCoin.CoinDetails.Value > amount {
			outCoinUnknapsack = append(outCoinUnknapsack, outCoin)
		} else {
			sumValueKnapsack += outCoin.CoinDetails.Value
			valuesKnapsack = append(valuesKnapsack, outCoin.CoinDetails.Value)
			outCoinKnapsack = append(outCoinKnapsack, outCoin)
		}
	}

	// target
	target := int64(sumValueKnapsack - amount)

	// if target > 1000, using Greedy algorithm
	// if target > 0, using Knapsack algorithm to choose coins
	// if target == 0, coins need to be spent is coins for Knapsack, we don't need to run Knapsack to find solution
	// if target < 0, instead of using Knapsack, we get the coin that has value is minimum in list unKnapsack coins
	if target > 1000 {
		choices := privacy.Greedy(outCoins, amount)
		for i, choice := range choices {
			if choice {
				totalResultOutputCoinAmount += outCoins[i].CoinDetails.Value
				resultOutputCoins = append(resultOutputCoins, outCoins[i])
			} else {
				remainOutputCoins = append(remainOutputCoins, outCoins[i])
			}
		}
	} else if target > 0 {
		choices := privacy.Knapsack(valuesKnapsack, uint64(target))
		for i, choice := range choices {
			if !choice {
				totalResultOutputCoinAmount += outCoinKnapsack[i].CoinDetails.Value
				resultOutputCoins = append(resultOutputCoins, outCoinKnapsack[i])
			} else {
				remainOutputCoins = append(remainOutputCoins, outCoinKnapsack[i])
			}
		}
		remainOutputCoins = append(remainOutputCoins, outCoinUnknapsack...)
	} else if target == 0 {
		totalResultOutputCoinAmount = sumValueKnapsack
		resultOutputCoins = outCoinKnapsack
		remainOutputCoins = outCoinUnknapsack
	} else {
		if len(outCoinUnknapsack) == 0 {
			return resultOutputCoins, remainOutputCoins, totalResultOutputCoinAmount, errors.New("Not enough coin")
		} else {
			sort.Slice(outCoinUnknapsack, func(i, j int) bool {
				return outCoinUnknapsack[i].CoinDetails.Value < outCoinUnknapsack[j].CoinDetails.Value
			})
			resultOutputCoins = append(resultOutputCoins, outCoinUnknapsack[0])
			totalResultOutputCoinAmount = outCoinUnknapsack[0].CoinDetails.Value
			for i := 1; i < len(outCoinUnknapsack); i++ {
				remainOutputCoins = append(remainOutputCoins, outCoinUnknapsack[i])
			}
			remainOutputCoins = append(remainOutputCoins, outCoinKnapsack...)
		}
	}

	return resultOutputCoins, remainOutputCoins, totalResultOutputCoinAmount, nil
}

// GetPaymentAddressFromPrivateKeyParams- deserialize a private key string
// and return paymentaddress object which relate to private key exactly
func (rpcServer RpcServer) GetPaymentAddressFromPrivateKeyParams(senderKeyParam string) (*privacy.PaymentAddress, error) {
	keySet, err := rpcServer.GetKeySetFromPrivateKeyParams(senderKeyParam)
	if err != nil {
		return nil, err
	}
	return &keySet.PaymentAddress, err
}

// GetKeySetFromKeyParams - deserialize a key string(wallet serialized)
// into keyWallet - this keywallet may contain
func (rpcServer RpcServer) GetKeySetFromKeyParams(keyParam string) (*cashec.KeySet, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(keyParam)
	if err != nil {
		return nil, err
	}
	return &keyWallet.KeySet, nil
}

// GetKeySetFromPrivateKeyParams - deserialize a private key string
// into keyWallet object and fill all keyset in keywallet with private key
func (rpcServer RpcServer) GetKeySetFromPrivateKeyParams(privateKeyWalletStr string) (*cashec.KeySet, error) {
	// deserialize to crate keywallet object which contain private key
	keyWallet, err := wallet.Base58CheckDeserialize(privateKeyWalletStr)
	if err != nil {
		return nil, err
	}
	// fill paymentaddress and readonly key with privatekey
	keyWallet.KeySet.ImportFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	return &keyWallet.KeySet, nil
}
