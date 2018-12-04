package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
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
func (self RpcServer) handleListTransactions(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	result := jsonresult.ListUnspentResult{
		ListUnspentResultItems: make(map[string]map[byte][]jsonresult.ListUnspentResultItem),
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

		txsMap, err := self.config.BlockChain.GetListTxByReadonlyKey(&keySet)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		listTxs := make([]jsonresult.ListUnspentResultItem, 0)
		for chainId, txs := range txsMap {
			for _, tx := range txs {
				item := jsonresult.ListUnspentResultItem{
					TxId:          tx.Hash().String(),
					JoinSplitDesc: make([]jsonresult.JoinSplitDesc, 0),
				}
				for _, desc := range tx.Descs {
					notes := desc.GetNote()
					amounts := make([]uint64, 0)
					for _, note := range notes {
						amounts = append(amounts, note.Value)
					}
					item.JoinSplitDesc = append(item.JoinSplitDesc, jsonresult.JoinSplitDesc{
						Anchors:     desc.Anchor,
						Commitments: desc.Commitments,
						Amounts:     amounts,
					})
				}
				listTxs = append(listTxs, item)
			}
			result.ListUnspentResultItems[readonlyKeyStr][chainId] = listTxs
		}
	}

	return result, nil
}

/*
// handleCreateTransaction handles createtransaction commands.
*/
func (self RpcServer) handleCreateRawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
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

	// param #3: estimation fee coin per kb
	estimateFeeCoinPerKb := int64(arrayParams[2].(float64))

	// param #4: estimation fee coin per kb by numblock
	numBlock := uint32(arrayParams[3].(float64))

	// list unspent tx for estimation fee
	estimateTotalAmount := totalAmmount
	usableTxsMap, _ := self.config.BlockChain.GetListUnspentTxByKeyset(&senderKey.KeySet, transaction.SortByAmount, false)
	candidateTxs := make([]*transaction.Tx, 0)
	candidateTxsMap := make(map[byte][]*transaction.Tx)
	for chainId, usableTxs := range usableTxsMap {
		for _, temp := range usableTxs {
			for _, desc := range temp.Descs {
				for _, note := range desc.GetNote() {
					amount := note.Value
					estimateTotalAmount -= int64(amount)
				}
			}
			txData := temp
			candidateTxsMap[chainId] = append(candidateTxsMap[chainId], &txData)
			candidateTxs = append(candidateTxs, &txData)
			if estimateTotalAmount <= 0 {
				break
			}
		}
	}

	// check real fee per Tx
	var realFee uint64
	if int64(estimateFeeCoinPerKb) == -1 {
		temp, _ := self.config.FeeEstimator[chainIdSender].EstimateFee(numBlock)
		estimateFeeCoinPerKb = int64(temp)
	}
	estimateFeeCoinPerKb += int64(self.config.Wallet.Config.IncrementalFee)
	estimateTxSizeInKb := transaction.EstimateTxSize(candidateTxs, paymentInfos)
	realFee = uint64(estimateFeeCoinPerKb) * uint64(estimateTxSizeInKb)

	// list unspent tx for create tx
	totalAmmount += int64(realFee)
	estimateTotalAmount = totalAmmount
	candidateTxsMap = make(map[byte][]*transaction.Tx, 0)
	for chainId, usableTxs := range usableTxsMap {
		for _, temp := range usableTxs {
			for _, desc := range temp.Descs {
				for _, note := range desc.GetNote() {
					amount := note.Value
					estimateTotalAmount -= int64(amount)
				}
			}
			txData := temp
			candidateTxsMap[chainId] = append(candidateTxsMap[chainId], &txData)
			if estimateTotalAmount <= 0 {
				break
			}
		}
	}

	// get merkleroot commitments, nullifers db, commitments db for every chain
	nullifiersDb := make(map[byte]([][]byte))
	commitmentsDb := make(map[byte]([][]byte))
	merkleRootCommitments := make(map[byte]*common.Hash)
	for chainId, _ := range candidateTxsMap {
		merkleRootCommitments[chainId] = &self.config.BlockChain.BestState[chainId].BestBlock.Header.MerkleRootCommitments
		// get tx view point
		txViewPoint, _ := self.config.BlockChain.FetchTxViewPoint(chainId)
		nullifiersDb[chainId] = txViewPoint.ListNullifiers()
		commitmentsDb[chainId] = txViewPoint.ListCommitments()
	}
	//missing flag for privacy-protocol
	// false by default
	flag := false
	tx, err := transaction.CreateTx(&senderKey.KeySet.PrivateKey, paymentInfos,
		merkleRootCommitments,
		candidateTxsMap,
		commitmentsDb,
		realFee,
		chainIdSender,
		flag)
	if err != nil {
		Logger.log.Critical(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	byteArrays, err := json.Marshal(tx)
	if err != nil {
		// return hex for a new tx
		return nil, NewRPCError(ErrUnexpected, err)
	}
	hexData := hex.EncodeToString(byteArrays)
	result := jsonresult.CreateTransactionResult{
		HexData: hexData,
	}
	return result, nil
}

/*
// handleSendTransaction implements the sendtransaction command.
Parameter #1—a serialized transaction to broadcast
Parameter #2–whether to allow high fees
Result—a TXID or error Message
*/
func (self RpcServer) handleSendRawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

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
handlCreateAndSendTx - RPC creates transaction and send to network
*/
func (self RpcServer) handlCreateAndSendTx(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawTransaction(params, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	hexStrOfTx := tx.HexData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, hexStrOfTx)
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
func (self RpcServer) handleGetMempoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonresult.GetMempoolInfo{}
	result.Size = self.config.TxMemPool.Count()
	result.Bytes = self.config.TxMemPool.Size()
	result.MempoolMaxFee = self.config.TxMemPool.MaxFee()
	result.ListTxs = self.config.TxMemPool.ListTxs()
	return result, nil
}

// Get transaction by Hash
func (self RpcServer) handleGetTransactionByHash(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	arrayParams := common.InterfaceSlice(params)
	// param #1: transaction Hash
	Logger.log.Infof("Get TransactionByHash input Param %+v", arrayParams[0].(string))
	txHash, _ := common.Hash{}.NewHashFromStr(arrayParams[0].(string))
	Logger.log.Infof("Get Transaction By Hash %+v", txHash)
	chainId, blockHash, index, tx, err := self.config.BlockChain.GetTransactionByHash(txHash)
	if err != nil {
		return nil, err
	}
	result := jsonresult.TransactionDetail{}
	switch tx.GetType() {
	case common.TxNormalType:
		{
			tempTx := tx.(*transaction.Tx)
			result = jsonresult.TransactionDetail{
				BlockHash:       blockHash.String(),
				Index:           uint64(index),
				ChainId:         chainId,
				Hash:            tx.Hash().String(),
				Version:         tempTx.Version,
				Type:            tempTx.Type,
				LockTime:        tempTx.LockTime,
				Fee:             tempTx.Fee,
				Descs:           tempTx.Descs,
				JSPubKey:        tempTx.JSPubKey,
				JSSig:           tempTx.JSSig,
				AddressLastByte: tempTx.AddressLastByte,
			}
		}
	case common.TxCustomTokenType:
		{
			tempTx := tx.(*transaction.TxCustomToken)
			result = jsonresult.TransactionDetail{
				BlockHash:       blockHash.String(),
				Index:           uint64(index),
				ChainId:         chainId,
				Hash:            tx.Hash().String(),
				Version:         tempTx.Version,
				Type:            tempTx.Type,
				LockTime:        tempTx.LockTime,
				Fee:             tempTx.Fee,
				Descs:           tempTx.Descs,
				JSPubKey:        tempTx.JSPubKey,
				JSSig:           tempTx.JSSig,
				AddressLastByte: tempTx.AddressLastByte,
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

func (self RpcServer) handleGetCommitteeCandidateList(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// param #1: private key of sender
	cndList := self.config.BlockChain.GetCommitteeCandidateList()
	return cndList, nil
}

func (self RpcServer) handleRetrieveCommiteeCandidate(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	candidateInfo := self.config.BlockChain.GetCommitteCandidate(params.(string))
	if candidateInfo == nil {
		return nil, nil
	}
	result := jsonresult.RetrieveCommitteecCandidateResult{}
	result.Init(candidateInfo)
	return result, nil
}

func (self RpcServer) handleGetBlockProducerList(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
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
) (interface{}, error) {
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
	hexData := hex.EncodeToString(byteArrays)
	result := jsonresult.CreateTransactionResult{
		HexData: hexData,
	}
	return result, nil
}

// handleSendRawTransaction...
func (self RpcServer) handleSendRawCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

	if err != nil {
		return nil, err
	}
	tx := transaction.TxCustomToken{}
	//tx := transaction.TxCustomToken{}
	// Logger.log.Info(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, err
	}

	hash, txDesc, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, err
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
	if err != nil {
		return nil, err
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

// handleCreateAndSendCustomTokenTransaction - create and send a tx which process on a custom token look like erc-20 on eth
func (self RpcServer) handleCreateAndSendCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawCustomTokenTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	hexStrOfTx := tx.HexData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, hexStrOfTx)
	txId, err := self.handleSendRawCustomTokenTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) handleGetListCustomTokenBalance(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
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
		return nil, err
	}
	for _, tx := range temps {
		item := jsonresult.CustomTokenBalance{}
		item.Name = tx.TxTokenData.PropertyName
		item.Symbol = tx.TxTokenData.PropertySymbol
		item.TokenID = tx.TxTokenData.PropertyID.String()
		tokenID := tx.TxTokenData.PropertyID
		res, err := self.config.BlockChain.GetListTokenHolders(&tokenID)
		if err != nil {
			return nil, err
		}
		item.Amount = res[hex.EncodeToString(accountPaymentAddress.Pk)]
		result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)
		result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	}
	return result, nil
}

// handleCustomTokenDetail - return list tx which relate to custom token by token id
func (self RpcServer) handleCustomTokenDetail(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	arrayParams := common.InterfaceSlice(params)
	tokenID, err := common.Hash{}.NewHashFromStr(arrayParams[0].(string))
	if err != nil {
		return nil, err
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

func (self RpcServer) handleListUnspentCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
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
	return unspentTxTokenOuts, err
}

// handleCreateSignatureOnCustomTokenTx - return a signature which is signed on raw custom token tx
func (self RpcServer) handleCreateSignatureOnCustomTokenTx(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

	if err != nil {
		return nil, err
	}
	tx := transaction.TxCustomToken{}
	// Logger.log.Info(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, err
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
		return nil, errors.New("Failed to sign the custom token")
	}
	return hex.EncodeToString(jsSignByteArray), nil
}

func (self RpcServer) handleCreateRawTxWithBuySellRequest(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	arrayParams := common.InterfaceSlice(params)
	tx, err := self.buildRawCustomTokenTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// Req param #4: buy/sell request info
	buySellReq := arrayParams[4].(map[string]interface{})
	tx.Tx.Metadata = metadata.NewBuySellRequest(buySellReq)

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	hexData := hex.EncodeToString(byteArrays)
	result := jsonresult.CreateTransactionResult{
		HexData: hexData,
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendTxWithBuySellRequest(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawTxWithBuySellRequest(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	hexStrOfTx := tx.HexData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, hexStrOfTx)
	txId, err := self.handleSendRawCustomTokenTransaction(newParam, closeChan)
	return txId, err
}
