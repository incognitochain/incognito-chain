package rpcserver

import (
	"encoding/hex"
	"encoding/json"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/voting"
	"github.com/ninjadotorg/constant/wire"
)

func (self RpcServer) buildRawSubmitDCBProposalTransaction(
	params interface{},
) (*transaction.Tx, error) {
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)
	DCBProposalRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	tx.Metadata = &metadata.SubmitDCBProposalMetadata{
		DCBVotingParams: voting.DCBVotingParams{
			SaleData: &voting.SaleData{
				SaleID:       []byte(DCBProposalRaw["SaleID"].(string)),
				BuyingAsset:  []byte(DCBProposalRaw["BuyingAsset"].(string)),
				SellingAsset: []byte(DCBProposalRaw["SellingAsset"].(string)),
				EndBlock:     int32(DCBProposalRaw["EndBlock"].(float64)),
			},
		},
		ExecuteDuration: int32(DCBProposalRaw["ExecuteDuration"].(float64)),
		Explanation:     DCBProposalRaw["Explanation"].(string),
	}
	return tx, err
}

func (self RpcServer) buildRawSubmitGOVProposalTransaction(
	params interface{},
) (*transaction.Tx, error) {
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)
	GOVProposalRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	tx.Metadata = &metadata.SubmitGOVProposalMetadata{
		GOVVotingParams: voting.GOVVotingParams{
			SalaryPerTx: uint64(GOVProposalRaw["SalaryPerTx"].(float64)),
			BasicSalary: uint64(GOVProposalRaw["BasicSalary"].(float64)),
			TxFee:       uint64(GOVProposalRaw["TxFee"].(float64)),
			SellingBonds: &voting.SellingBonds{
				BondsToSell:    uint64(GOVProposalRaw["BondsToSell"].(float64)),
				BondPrice:      uint64(GOVProposalRaw["BondPrice"].(float64)),
				Maturity:       uint32(GOVProposalRaw["Maturity"].(float64)),
				BuyBackPrice:   uint64(GOVProposalRaw["BuyBackPrice"].(float64)),
				StartSellingAt: uint32(GOVProposalRaw["StartSellingAt"].(float64)),
				SellingWithin:  uint32(GOVProposalRaw["SellingWithin"].(float64)),
			},
			RefundInfo: &voting.RefundInfo{
				ThresholdToLargeTx: uint64(GOVProposalRaw["ThresholdToLargeTx"].(float64)),
				RefundAmount:       uint64(GOVProposalRaw["RefundAmount"].(float64)),
			},
		},
		ExecuteDuration: int32(GOVProposalRaw["ExecuteDuration"].(float64)),
		Explaination:    string(GOVProposalRaw["Explaination"].(string)),
	}
	tx.Metadata = arrayParams[len(arrayParams)-1].(*metadata.SubmitGOVProposalMetadata)
	return tx, err
}

func (self RpcServer) handleCreateRawSubmitDCBProposalTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, error) {
	tx, err := self.buildRawSubmitDCBProposalTransaction(params)
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

func (self RpcServer) handleCreateRawSubmitGOVProposalTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, error) {
	tx, err := self.buildRawSubmitGOVProposalTransaction(params)
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

func (self RpcServer) handleSendRawSubmitDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

	if err != nil {
		return nil, err
	}
	tx := transaction.Tx{}
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
	txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
	if err != nil {
		return nil, err
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleSendRawSubmitGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

	if err != nil {
		return nil, err
	}
	tx := transaction.Tx{}
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
	txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
	if err != nil {
		return nil, err
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateAndSendSubmitDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawSubmitDCBProposalTransaction(params, closeChan)
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
	txId, err := self.handleSendRawSubmitDCBProposalTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) handleCreateAndSendSubmitGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawSubmitGOVProposalTransaction(params, closeChan)
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
	txId, err := self.handleSendRawSubmitGOVProposalTransaction(newParam, closeChan)
	return txId, err
}
