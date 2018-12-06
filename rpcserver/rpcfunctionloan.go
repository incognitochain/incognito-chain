package rpcserver

import (
	"encoding/hex"
	"encoding/json"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/ninjadotorg/constant/wire"
	"github.com/pkg/errors"
)

func (self RpcServer) handleGetLoanParams(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return self.config.BlockChain.BestState[0].BestBlock.Header.DCBConstitution.DCBParams.LoanParams, nil
}

func (self RpcServer) handleCreateRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
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

	// param #2: Fee
	fee := uint64(arrayParams[1].(float64))
	if fee == 0 {
		fee = self.config.BlockChain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.BasicSalary
	}
	totalAmmount := fee

	// param #3: loan params
	loanParams := arrayParams[2].(map[string]interface{})
	loanRequest := metadata.NewLoanRequest(loanParams)
	if loanRequest == nil {
		return nil, errors.New("Miss data")
	}

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
					estimateTotalAmount -= uint64(amount)
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
	noPrivacy := false
	tx, err := transaction.CreateTx(
		&senderKey.KeySet.PrivateKey,
		nil,
		merkleRootCommitments,
		candidateTxsMap,
		commitmentsDb,
		fee,
		chainIdSender,
		noPrivacy,
	)
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

func (self RpcServer) handleSendRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

	if err != nil {
		return nil, err
	}
	tx := transaction.Tx{}
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
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCLoanRequestToken)
	if err != nil {
		return nil, err
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateAndSendLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawLoanRequest(params, closeChan)
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
	txId, err := self.handleSendRawLoanRequest(newParam, closeChan)
	return txId, err
}

func (self RpcServer) handleCreateRawLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
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

	// param #2: Fee
	fee := uint64(arrayParams[1].(float64))
	if fee == 0 {
		fee = self.config.BlockChain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.BasicSalary
	}
	totalAmmount := fee

	// param #3: loan params
	loanParams := arrayParams[2].(map[string]interface{})
	loanResponse := transaction.NewLoanResponse(loanParams)
	if loanResponse == nil {
		return nil, errors.New("Miss data")
	}

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
					estimateTotalAmount -= uint64(amount)
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
	tx, err := transaction.CreateTxLoanResponse(transaction.FeeArgs{
		Fee:           fee,
		Commitments:   commitmentsDb,
		UsableTx:      candidateTxsMap,
		PaymentInfo:   nil,
		Rts:           merkleRootCommitments,
		SenderChainID: chainIdSender,
		SenderKey:     &senderKey.KeySet.PrivateKey,
	}, loanResponse)
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

func (self RpcServer) handleSendRawLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

	if err != nil {
		return nil, err
	}
	tx := transaction.TxLoanResponse{}
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
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCLoanResponseToken)
	if err != nil {
		return nil, err
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateAndSendLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawLoanResponse(params, closeChan)
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
	txId, err := self.handleSendRawLoanResponse(newParam, closeChan)
	return txId, err
}

func (self RpcServer) handleCreateRawLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
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

	// param #2: Fee
	fee := uint64(arrayParams[1].(float64))
	if fee == 0 {
		fee = self.config.BlockChain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.BasicSalary
	}
	totalAmmount := fee

	// param #3: loan params
	loanParams := arrayParams[2].(map[string]interface{})
	loanWithdraw := transaction.NewLoanWithdraw(loanParams)
	if loanWithdraw == nil {
		return nil, errors.New("Miss data")
	}

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
					estimateTotalAmount -= uint64(amount)
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
	tx, err := transaction.CreateTxLoanWithdraw(transaction.FeeArgs{
		Fee:           fee,
		Commitments:   commitmentsDb,
		UsableTx:      candidateTxsMap,
		PaymentInfo:   nil,
		Rts:           merkleRootCommitments,
		SenderChainID: chainIdSender,
		SenderKey:     &senderKey.KeySet.PrivateKey,
	}, loanWithdraw)
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

func (self RpcServer) handleSendRawLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

	if err != nil {
		return nil, err
	}
	tx := transaction.TxLoanWithdraw{}
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
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCLoanWithdrawToken)
	if err != nil {
		return nil, err
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateAndSendLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawLoanWithdraw(params, closeChan)
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
	txId, err := self.handleSendRawLoanWithdraw(newParam, closeChan)
	return txId, err
}

func (self RpcServer) handleCreateRawLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
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

	// param #2: Fee
	fee := uint64(arrayParams[1].(float64))
	if fee == 0 {
		fee = self.config.BlockChain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.BasicSalary
	}
	totalAmmount := fee

	// param #3: loan params
	loanParams := arrayParams[2].(map[string]interface{})
	loanPayment := transaction.NewLoanPayment(loanParams)
	if loanPayment == nil {
		return nil, errors.New("Miss data")
	}

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
					estimateTotalAmount -= uint64(amount)
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
	tx, err := transaction.CreateTxLoanPayment(transaction.FeeArgs{
		Fee:           fee,
		Commitments:   commitmentsDb,
		UsableTx:      candidateTxsMap,
		PaymentInfo:   nil,
		Rts:           merkleRootCommitments,
		SenderChainID: chainIdSender,
		SenderKey:     &senderKey.KeySet.PrivateKey,
	}, loanPayment)
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

func (self RpcServer) handleSendRawLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

	if err != nil {
		return nil, err
	}
	tx := transaction.TxLoanWithdraw{}
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
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCLoanPayToken)
	if err != nil {
		return nil, err
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateAndSendLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawLoanPayment(params, closeChan)
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
	txId, err := self.handleSendRawLoanPayment(newParam, closeChan)
	return txId, err
}
