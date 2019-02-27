package rpcserver

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/wallet"
)

func (rpcServer RpcServer) handleCreateCrowdsaleRequestToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := metaConstructors[CreateAndSendCrowdsaleRequestToken]
	return rpcServer.createRawCustomTokenTxWithMetadata(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendCrowdsaleRequestToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawCustomTokenTxWithMetadata(params, closeChan)
}

// handleCreateAndSendCrowdsaleRequestToken for user to sell bonds to DCB
func (rpcServer RpcServer) handleCreateAndSendCrowdsaleRequestToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateCrowdsaleRequestToken,
		RpcServer.handleSendCrowdsaleRequestToken,
	)
}

func (rpcServer RpcServer) handleCreateCrowdsaleRequestConstant(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := metaConstructors[CreateAndSendCrowdsaleRequestConstant]
	return rpcServer.createRawTxWithMetadata(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendCrowdsaleRequestConstant(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawTxWithMetadata(params, closeChan)
}

// handleCreateAndSendCrowdsaleRequestToken for user to buy bonds from DCB
func (rpcServer RpcServer) handleCreateAndSendCrowdsaleRequestConstant(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateCrowdsaleRequestConstant,
		RpcServer.handleSendCrowdsaleRequestConstant,
	)
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
		SaleID           string
		EndBlock         uint64
		BuyingAsset      string
		BuyingAmount     uint64
		DefaultBuyPrice  uint64
		SellingAsset     string
		SellingAmount    uint64
		DefaultSellPrice uint64
		Type             string
	}
	result := []CrowdsaleInfo{}
	saleDataList, err := rpcServer.config.BlockChain.GetAllCrowdsales()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Error querying crowdsales"))
	}
	for _, saleData := range saleDataList {
		if height >= saleData.EndBlock {
			continue
		}

		// Add type for better ux, not blockchain-related
		crowdsaleType := "buyable"
		if saleData.SellingAsset.IsEqual(&common.ConstantID) {
			crowdsaleType = "sellable"
		}

		info := CrowdsaleInfo{
			SaleID:           hex.EncodeToString(saleData.SaleID),
			EndBlock:         saleData.EndBlock,
			BuyingAsset:      saleData.BuyingAsset.String(),
			BuyingAmount:     saleData.BuyingAmount,
			DefaultBuyPrice:  saleData.DefaultBuyPrice,
			SellingAsset:     saleData.SellingAsset.String(),
			SellingAmount:    saleData.SellingAmount,
			DefaultSellPrice: saleData.DefaultSellPrice,
			Type:             crowdsaleType,
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
		saleData := param.(map[string]interface{})
		saleID, _ := hex.DecodeString(saleData["SaleID"].(string))
		proposalTxHash, _ := common.NewHashFromStr(saleData["ProposalTxHash"].(string))
		fmt.Printf("[db] proposalTxHash: %+v\n", proposalTxHash)
		buyingAmount := uint64(saleData["BuyingAmount"].(float64))
		sellingAmount := uint64(saleData["SellingAmount"].(float64))
		if _, _, _, err := (*rpcServer.config.Database).GetCrowdsaleData(saleID); err == nil {
			fmt.Printf("[db] cs existed\n")
			continue
		}
		if err := (*rpcServer.config.Database).StoreCrowdsaleData(
			saleID,
			*proposalTxHash,
			buyingAmount,
			sellingAmount,
		); err != nil {
			fmt.Printf("[db] fail store crowdsale data %+v\n", err)
			return nil, NewRPCError(ErrUnexpected, err)
		}
	}
	return true, nil
}

func (rpcServer RpcServer) handleGetListDCBProposalBuyingAssets(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// TODO(@0xankylosaurus): call get list bonds
	buyingAssets := map[string]string{
		"Bond 1":   "4c420b974449ac188c155a7029706b8419a591ee398977d00000000000000000",
		"Constant": common.ConstantID.String(),
	} // From asset name to asset id
	return buyingAssets, nil
}

func (rpcServer RpcServer) handleGetListDCBProposalSellingAssets(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// TODO(@0xankylosaurus): call get list bonds
	sellingAssets := map[string]string{
		"Constant": common.ConstantID.String(),
		"Bond 2":   "4c420b974449ac188c155a7029706b8419a591ee398977d00000000000000000",
	} // From asset name to asset id
	return sellingAssets, nil
}
