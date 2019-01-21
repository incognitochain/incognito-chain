package rpcserver

import (
	"errors"
	"fmt"

	chainParams "github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/wallet"
)

func (rpcServer RpcServer) handleCreateAndSendCrowdsaleRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
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
	chainID, _ := common.GetTxSenderChain(paymentAddr.KeySet.PaymentAddress.Pk[len(paymentAddr.KeySet.PaymentAddress.Pk)-1])
	height := rpcServer.config.BlockChain.GetChainHeight(chainID)

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
