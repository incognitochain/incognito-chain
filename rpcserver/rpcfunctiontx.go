package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/ninjadotorg/constant/wire"
)

/*
// handleList returns a slice of objects representing the wallet
// transactions fitting the given criteria. The confirmations will be more than
// minconf, less than maxconf and if addresses is populated only the addresses
// contained within it will be considered.  If we know nothing about a
// transaction an empty array will be returned.
// params:
Parameter #1—the minimum number of confirmations an output must have
Parameter #2—the maximum number of confirmations an output may have
Parameter #3—the list readonly which be used to view utxo
*/
func (self RpcServer) handleListOutputCoins(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	result := jsonresult.ListUnspentResult{
		ListUnspentResultItems: make(map[string][]jsonresult.ListUnspentResultItem),
	}

	// get params
	paramsArray := common.InterfaceSlice(params)
	listKeyParams := common.InterfaceSlice(paramsArray[0])
	for _, keyParam := range listKeyParams {
		keys := keyParam.(map[string]interface{})

		// get keyset only contain readonly-key by deserializing
		readonlyKeyStr := keys["ReadonlyKey"].(string)
		readonlyKey, err := wallet.Base58CheckDeserialize(readonlyKeyStr)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}

		// get keyset only contain pub-key by deserializing
		pubKeyStr := keys["PaymentAddress"].(string)
		pubKey, err := wallet.Base58CheckDeserialize(pubKeyStr)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}

		// create a key set
		keySet := cashec.KeySet{
			ReadonlyKey:    readonlyKey.KeySet.ReadonlyKey,
			PaymentAddress: pubKey.KeySet.PaymentAddress,
		}
		lastByte := keySet.PaymentAddress.Pk[len(keySet.PaymentAddress.Pk)-1]
		chainIdSender, err := common.GetTxSenderChain(lastByte)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		constantTokenID := &common.Hash{}
		constantTokenID.SetBytes(common.ConstantID[:])
		outputCoins, err := self.config.BlockChain.GetListOutputCoinsByKeyset(&keySet, chainIdSender, constantTokenID)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		listTxs := make([]jsonresult.ListUnspentResultItem, 0)
		item := jsonresult.ListUnspentResultItem{
			OutCoins: make([]jsonresult.OutCoin, 0),
		}
		for _, outCoin := range outputCoins {
			item.OutCoins = append(item.OutCoins, jsonresult.OutCoin{
				SerialNumber:   base58.Base58Check{}.Encode(outCoin.CoinDetails.SerialNumber.Compress(), byte(0x00)),
				PublicKey:      base58.Base58Check{}.Encode(outCoin.CoinDetails.PublicKey.Compress(), byte(0x00)),
				Value:          outCoin.CoinDetails.Value,
				Info:           base58.Base58Check{}.Encode(outCoin.CoinDetails.Info[:], byte(0x00)),
				CoinCommitment: base58.Base58Check{}.Encode(outCoin.CoinDetails.CoinCommitment.Compress(), byte(0x00)),
				Randomness:     *outCoin.CoinDetails.Randomness,
				SNDerivator:    *outCoin.CoinDetails.SNDerivator,
			})
			listTxs = append(listTxs, item)
			result.ListUnspentResultItems[readonlyKeyStr] = listTxs
		}
	}

	return result, nil
}

func (self RpcServer) buildRawTransaction(params interface{}) (*transaction.Tx, *RPCError) {
	Logger.log.Info(params)

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

	// param #2: list receiver
	totalAmmount := int64(0)
	receiversParam := arrayParams[1].(map[string]interface{})
	paymentInfos := make([]*privacy.PaymentInfo, 0)
	for pubKeyStr, amount := range receiversParam {
		receiverPubKey, err := wallet.Base58CheckDeserialize(pubKeyStr)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		paymentInfo := &privacy.PaymentInfo{
			Amount:         common.ConstantToMiliConstant(uint64(amount.(float64))),
			PaymentAddress: receiverPubKey.KeySet.PaymentAddress,
		}
		totalAmmount += int64(paymentInfo.Amount)
		paymentInfos = append(paymentInfos, paymentInfo)
	}

	// param #3: estimation fee nano constant per kb
	estimateFeeCoinPerKb := int64(arrayParams[2].(float64))

	// param #4: estimation fee coin per kb by numblock
	numBlock := uint64(arrayParams[3].(float64))

	// list unspent tx for estimation fee
	estimateTotalAmount := totalAmmount
	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	outCoins, err := self.config.BlockChain.GetListOutputCoinsByKeyset(&senderKey.KeySet, chainIdSender, constantTokenID)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	if len(outCoins) == 0 {
		return nil, NewRPCError(ErrUnexpected, nil)
	}
	candidateOutputCoins := make([]*privacy.OutputCoin, 0)
	for _, note := range outCoins {
		amount := note.CoinDetails.Value
		candidateOutputCoins = append(candidateOutputCoins, note)
		estimateTotalAmount -= int64(amount)
		if estimateTotalAmount <= 0 {
			break
		}
	}

	// check real fee(nano constant) per tx
	realFee := self.EstimateFee(estimateFeeCoinPerKb, candidateOutputCoins, paymentInfos, chainIdSender, numBlock)

	// list unspent tx for create tx
	totalAmmount += int64(realFee)
	estimateTotalAmount = totalAmmount
	if totalAmmount > 0 {
		candidateOutputCoins = make([]*privacy.OutputCoin, 0)
		for _, note := range outCoins {
			amount := note.CoinDetails.Value
			candidateOutputCoins = append(candidateOutputCoins, note)
			estimateTotalAmount -= int64(amount)
			if estimateTotalAmount <= 0 {
				break
			}
		}
	}

	//missing flag for privacy-protocol
	// false by default
	inputCoins := transaction.ConvertOutputCoinToInputCoin(candidateOutputCoins)
	tx := transaction.Tx{}
	err = tx.Init(
		&senderKey.KeySet.PrivateKey,
		paymentInfos,
		inputCoins,
		realFee,
		true,
		*self.config.Database,
		nil, // use for constant coin -> nil is valid
	)
	return &tx, NewRPCError(ErrUnexpected, err)
}

/*
// handleCreateTransaction handles createtransaction commands.
*/
func (self RpcServer) handleCreateRawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawTransaction(params)
	if err != nil {
		Logger.log.Critical(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	byteArrays, err := json.Marshal(tx)
	if err != nil {
		// return hex for a new tx
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

/*
// handleSendTransaction implements the sendtransaction command.
Parameter #1—a serialized transaction to broadcast
Parameter #2–whether to allow high fees
Result—a TXID or error Message
*/
func (self RpcServer) handleSendRawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckData := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	var tx transaction.Tx
	// Logger.log.Info(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	hash, txDesc, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast Message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	err = self.config.Server.PushMessageToAll(txMsg)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txID := tx.Hash().String()
	result := jsonresult.CreateTransactionResult{
		TxID: txID,
	}
	return result, nil
}

/*
handleCreateAndSendTx - RPC creates transaction and send to network
*/
func (self RpcServer) handleCreateAndSendTx(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawTransaction(params, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := self.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

/*
handleGetMempoolInfo - RPC returns information about the node's current txs memory pool
*/
func (self RpcServer) handleGetMempoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetMempoolInfo{}
	result.Size = self.config.TxMemPool.Count()
	result.Bytes = self.config.TxMemPool.Size()
	result.MempoolMaxFee = self.config.TxMemPool.MaxFee()
	result.ListTxs = self.config.TxMemPool.ListTxs()
	return result, nil
}

// Get transaction by Hash
func (self RpcServer) handleGetTransactionByHash(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	// param #1: transaction Hash
	Logger.log.Infof("Get TransactionByHash input Param %+v", arrayParams[0].(string))
	txHash, _ := common.Hash{}.NewHashFromStr(arrayParams[0].(string))
	Logger.log.Infof("Get Transaction By Hash %+v", txHash)
	chainId, blockHash, index, tx, err := self.config.BlockChain.GetTransactionByHash(txHash)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.TransactionDetail{}
	switch tx.GetType() {
	case common.TxNormalType:
		{
			tempTx := tx.(*transaction.Tx)
			result = jsonresult.TransactionDetail{
				BlockHash: blockHash.String(),
				Index:     uint64(index),
				ChainId:   chainId,
				Hash:      tx.Hash().String(),
				Version:   tempTx.Version,
				Type:      tempTx.Type,
				LockTime:  tempTx.LockTime,
				Fee:       tempTx.Fee,
				Proof:     tempTx.Proof,
				SigPubKey: tempTx.SigPubKey,
				Sig:       tempTx.Sig,
			}
		}
	case common.TxCustomTokenType:
		{
			tempTx := tx.(*transaction.TxCustomToken)
			result = jsonresult.TransactionDetail{
				BlockHash: blockHash.String(),
				Index:     uint64(index),
				ChainId:   chainId,
				Hash:      tx.Hash().String(),
				Version:   tempTx.Version,
				Type:      tempTx.Type,
				LockTime:  tempTx.LockTime,
				Fee:       tempTx.Fee,
				Proof:     tempTx.Proof,
				SigPubKey: tempTx.SigPubKey,
				Sig:       tempTx.Sig,
			}
			txCustomData, _ := json.MarshalIndent(tempTx.TxTokenData, "", "\t")
			result.MetaData = string(txCustomData)
		}
	default:
		{

		}
	}
	return result, nil
}

func (self RpcServer) handleGetCommitteeCandidateList(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// param #1: private key of sender
	cndList := self.config.BlockChain.GetCommitteeCandidateList()
	return cndList, nil
}

func (self RpcServer) handleRetrieveCommiteeCandidate(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	candidateInfo := self.config.BlockChain.GetCommitteCandidate(params.(string))
	if candidateInfo == nil {
		return nil, nil
	}
	result := jsonresult.RetrieveCommitteecCandidateResult{}
	result.Init(candidateInfo)
	return result, nil
}

func (self RpcServer) handleGetBlockProducerList(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := make(map[string]string)
	for chainID, bestState := range self.config.BlockChain.BestState {
		if bestState.BestBlock.BlockProducer != "" {
			result[strconv.Itoa(chainID)] = bestState.BestBlock.BlockProducer
		} else {
			result[strconv.Itoa(chainID)] = self.config.ChainParams.GenesisBlock.Header.Committee[chainID]
		}
	}
	return result, nil
}

// handleCreateRawCustomTokenTransaction - handle create a custom token command and return in hex string format.
func (self RpcServer) handleCreateRawCustomTokenTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawCustomTokenTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

// handleSendRawTransaction...
func (self RpcServer) handleSendRawCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckData := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := transaction.TxCustomToken{}
	//tx := transaction.TxCustomToken{}
	// Logger.log.Info(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	hash, txDesc, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

// handleCreateAndSendCustomTokenTransaction - create and send a tx which process on a custom token look like erc-20 on eth
func (self RpcServer) handleCreateAndSendCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawCustomTokenTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawCustomTokenTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) handleGetListCustomTokenBalance(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	accountParam := arrayParams[0].(string)
	account, err := wallet.Base58CheckDeserialize(accountParam)
	if err != nil {
		return nil, nil
	}
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}
	result.PaymentAddress = accountParam
	accountPaymentAddress := account.KeySet.PaymentAddress
	temps, err := self.config.BlockChain.ListCustomToken()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	for _, tx := range temps {
		item := jsonresult.CustomTokenBalance{}
		item.Name = tx.TxTokenData.PropertyName
		item.Symbol = tx.TxTokenData.PropertySymbol
		item.TokenID = tx.TxTokenData.PropertyID.String()
		tokenID := tx.TxTokenData.PropertyID
		res, err := self.config.BlockChain.GetListTokenHolders(&tokenID)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		pubkey := base58.Base58Check{}.Encode(accountPaymentAddress.Pk, 0x00)
		item.Amount = res[pubkey]
		if item.Amount == 0 {
			continue
		}
		result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)
		result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	}
	return result, nil
}

func (self RpcServer) handleGetListPrivacyCustomTokenBalance(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	privateKey := arrayParams[0].(string)
	account, err := wallet.Base58CheckDeserialize(privateKey)
	account.KeySet.ImportFromPrivateKey(&account.KeySet.PrivateKey)
	if err != nil {
		return nil, nil
	}
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}
	result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	temps, err := self.config.BlockChain.ListCustomToken()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	for _, tx := range temps {
		item := jsonresult.CustomTokenBalance{}
		item.Name = tx.TxTokenData.PropertyName
		item.Symbol = tx.TxTokenData.PropertySymbol
		item.TokenID = tx.TxTokenData.PropertyID.String()
		tokenID := tx.TxTokenData.PropertyID

		balance := uint64(0)
		// get balance for accountName in wallet
		lastByte := account.KeySet.PaymentAddress.Pk[len(account.KeySet.PaymentAddress.Pk)-1]
		chainIdSender, err := common.GetTxSenderChain(lastByte)
		constantTokenID := &common.Hash{}
		constantTokenID.SetBytes(common.ConstantID[:])
		outcoints, err := self.config.BlockChain.GetListOutputCoinsByKeyset(&account.KeySet, chainIdSender, &tokenID)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		for _, out := range outcoints {
			balance += out.CoinDetails.Value
		}

		item.Amount = balance
		if item.Amount == 0 {
			continue
		}
		result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)
		result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	}
	return result, nil
}

// handleCustomTokenDetail - return list tx which relate to custom token by token id
func (self RpcServer) handleCustomTokenDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	tokenID, err := common.Hash{}.NewHashFromStr(arrayParams[0].(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	txs, _ := self.config.BlockChain.GetCustomTokenTxsHash(tokenID)
	result := jsonresult.CustomToken{
		ListTxs: []string{},
	}
	for _, tx := range txs {
		result.ListTxs = append(result.ListTxs, tx.String())
	}
	return result, nil
}

// handlePrivacyCustomTokenDetail - return list tx which relate to privacy custom token by token id
func (self RpcServer) handlePrivacyCustomTokenDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	tokenID, err := common.Hash{}.NewHashFromStr(arrayParams[0].(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	txs, _ := self.config.BlockChain.GetPrivacyCustomTokenTxsHash(tokenID)
	result := jsonresult.CustomToken{
		ListTxs: []string{},
	}
	for _, tx := range txs {
		result.ListTxs = append(result.ListTxs, tx.String())
	}
	return result, nil
}

func (self RpcServer) handleListUnspentCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	// param #1: paymentaddress of sender
	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKeyset := senderKey.KeySet

	// param #2: tokenID
	tokenIDParam := arrayParams[1]
	tokenID, _ := common.Hash{}.NewHashFromStr(tokenIDParam.(string))
	unspentTxTokenOuts, err := self.config.BlockChain.GetUnspentTxCustomTokenVout(senderKeyset, tokenID)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return unspentTxTokenOuts, NewRPCError(ErrUnexpected, err)
}

// handleCreateSignatureOnCustomTokenTx - return a signature which is signed on raw custom token tx
func (self RpcServer) handleCreateSignatureOnCustomTokenTx(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckDate := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := transaction.TxCustomToken{}
	// Logger.log.Info(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKeyParam := arrayParams[1]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	jsSignByteArray, err := tx.GetTxCustomTokenSignature(senderKey.KeySet)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Failed to sign the custom token"))
	}
	return hex.EncodeToString(jsSignByteArray), nil
}

// handleRandomCommitments - from input of outputcoin, random to create data for create new tx
func (self RpcServer) handleRandomCommitments(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

	// #1: payment address
	paymentAddressStr := arrayParams[0].(string)
	key, err := wallet.Base58CheckDeserialize(paymentAddressStr)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	lastByte := key.KeySet.PaymentAddress.Pk[len(key.KeySet.PaymentAddress.Pk)-1]
	chainIdSender, err := common.GetTxSenderChain(lastByte)

	// #2: available inputCoin from old outputcoin
	data := jsonresult.ListUnspentResultItem{}
	data.Init(arrayParams[0])
	usableOutputCoins := []*privacy.OutputCoin{}
	for _, item := range data.OutCoins {
		i := &privacy.OutputCoin{
			CoinDetails: &privacy.Coin{
				Value:       item.Value,
				Randomness:  &item.Randomness,
				SNDerivator: &item.SNDerivator,
			},
		}
		i.CoinDetails.Info, _, _ = base58.Base58Check{}.Decode(item.Info)

		CoinCommitmentBytes, _, _ := base58.Base58Check{}.Decode(item.CoinCommitment)
		CoinCommitment := &privacy.EllipticPoint{}
		_ = CoinCommitment.Decompress(CoinCommitmentBytes)
		i.CoinDetails.CoinCommitment = CoinCommitment

		PublicKeyBytes, _, _ := base58.Base58Check{}.Decode(item.PublicKey)
		PublicKey := &privacy.EllipticPoint{}
		_ = PublicKey.Decompress(PublicKeyBytes)
		i.CoinDetails.PublicKey = PublicKey

		InfoBytes, _, _ := base58.Base58Check{}.Decode(item.Info)
		i.CoinDetails.Info = InfoBytes

		usableOutputCoins = append(usableOutputCoins, i)
	}
	usableInputCoins := transaction.ConvertOutputCoinToInputCoin(usableOutputCoins)
	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	commitmentIndexs, myCommitmentIndexs := self.config.BlockChain.RandomCommitmentsProcess(usableInputCoins, 0, chainIdSender, constantTokenID)
	result := make(map[string]interface{})
	result["CommitmentIndexs"] = commitmentIndexs
	result["MyCommitmentIndexs"] = myCommitmentIndexs

	return result, nil
}

// handleHasSerialNumbers - check list serial numbers existed in db of node
func (self RpcServer) handleHasSerialNumbers(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

	// #1: payment address
	paymentAddressStr := arrayParams[0].(string)
	key, err := wallet.Base58CheckDeserialize(paymentAddressStr)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	lastByte := key.KeySet.PaymentAddress.Pk[len(key.KeySet.PaymentAddress.Pk)-1]
	chainIdSender, err := common.GetTxSenderChain(lastByte)

	//#2: list serialnumbers in base58check encode string
	serialNumbersStr := arrayParams[1].([]interface{})

	result := make(map[byte][]string)
	result[0] = []string{}
	result[1] = []string{}
	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	for _, item := range serialNumbersStr {
		serialNumber, _, _ := base58.Base58Check{}.Decode(item.(string))
		db := *(self.config.Database)
		ok, err := db.HasSerialNumber(constantTokenID, serialNumber, chainIdSender)
		if ok && err != nil {
			result[0] = append(result[0], item.(string))
		} else {
			result[1] = append(result[1], item.(string))
		}
	}

	return result, nil
}

// handleCreateRawCustomTokenTransaction - handle create a custom token command and return in hex string format.
func (self RpcServer) handleCreateRawPrivacyCustomTokenTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	tx, err := self.buildRawPrivacyCustomTokenTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

// handleSendRawTransaction...
func (self RpcServer) handleSendRawPrivacyCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckData := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := transaction.TxCustomTokenPrivacy{}
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	hash, txDesc, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdPrivacyCustomToken)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

// handleCreateAndSendCustomTokenTransaction - create and send a tx which process on a custom token look like erc-20 on eth
func (self RpcServer) handleCreateAndSendPrivacyCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawPrivacyCustomTokenTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	return txId, err
}
