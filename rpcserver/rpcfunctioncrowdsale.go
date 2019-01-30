package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"

	chainParams "github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/ninjadotorg/constant/wire"
)

func (rpcServer RpcServer) sendRawCrowdsaleTx(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckDate := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := transaction.Tx{}
	err = json.Unmarshal(rawTxBytes, &tx)
	fmt.Printf("%+v\n", tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	hash, txDesc, err := rpcServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCLoanRequestToken)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	rpcServer.config.Server.PushMessageToAll(txMsg)

	result := jsonresult.CreateTransactionResult{
		TxID: tx.Hash().String(),
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateAndSendCrowdsaleRequestToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	crowdsaleDataRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	meta, err := metadata.NewCrowdsaleRequest(crowdsaleDataRaw)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx, err := rpcServer.buildRawCustomTokenTransaction(params, meta)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	byteArrays, errMarshal := json.Marshal(tx)
	if errMarshal != nil {
		// return hex for a new tx
		return nil, NewRPCError(ErrUnexpected, errMarshal)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	//rpcServer.buildRawCustomTokenTransaction(params, metadata)
	return result, nil
}

func (rpcServer RpcServer) handleCreateAndSendCrowdsaleRequestConstant(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	//rpcServer.buildRawTransaction(params, metadata)
	_ = arrayParams
	return nil, nil
}

// handleGetListOngoingCrowdsale receives a payment address and find all ongoing crowdsales on the chain that handles that address
func (rpcServer RpcServer) handleGetListOngoingCrowdsale(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// Get height of the chain containing the payment address
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) <= 0 {
		return nil, NewRPCError(ErrUnexpected, errors.New("Must provider at least 1 payment address"))
	}
	paymentAddrStr, _ := arrayParams[0].(string)
	paymentAddr, err := wallet.Base58CheckDeserialize(paymentAddrStr)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Payment address invalid"))
	}
	shardID := common.GetShardIDFromLastByte(paymentAddr.KeySet.PaymentAddress.Pk[len(paymentAddr.KeySet.PaymentAddress.Pk)-1])
	height := rpcServer.config.BlockChain.GetChainHeight(shardID)

	// Get all ongoing crowdsales for that chain
	type CrowdsaleInfo struct {
		SaleID        string
		EndBlock      uint64
		BuyingAsset   string
		BuyingAmount  uint64
		SellingAsset  string
		SellingAmount uint64
	}
	result := []CrowdsaleInfo{}
	endBlocks, buyingAssets, buyingAmounts, sellingAssets, sellingAmounts, err := (*rpcServer.config.Database).GetAllCrowdsales()
	fmt.Println("[db] endBlocks:", endBlocks)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Error querying crowdsales"))
	}
	for i, endBlock := range endBlocks {
		if height >= endBlock {
			continue
		}
		info := CrowdsaleInfo{
			SaleID:        "",
			EndBlock:      endBlock,
			BuyingAsset:   buyingAssets[i].String(),
			BuyingAmount:  buyingAmounts[i],
			SellingAsset:  sellingAssets[i].String(),
			SellingAmount: sellingAmounts[i],
		}
		result = append(result, info)
	}
	return result, nil
}

// handleTESTStoreCrowdsale receives a crowdsale and store to database for testing
func (rpcServer RpcServer) handleTESTStoreCrowdsale(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

	// Store saledata in db
	for _, param := range arrayParams {
		data := chainParams.NewSaleDataFromJson(param)
		if _, _, _, _, _, err := (*rpcServer.config.Database).GetCrowdsaleData(data.SaleID); err == nil {
			continue
		}
		if err := (*rpcServer.config.Database).StoreCrowdsaleData(
			data.SaleID,
			data.EndBlock,
			data.BuyingAsset,
			data.BuyingAmount,
			data.SellingAsset,
			data.SellingAmount,
		); err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
	}
	return true, nil
}
