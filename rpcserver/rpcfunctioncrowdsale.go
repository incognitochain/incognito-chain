package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
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

// handleGetListOngoingCrowdsale finds all ongoing crowdsales
func (rpcServer RpcServer) handleGetListOngoingCrowdsale(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// Get all ongoing crowdsales for that chain
	type CrowdsaleInfo struct {
		SaleID   string
		EndBlock uint64
		BondID   string
		Amount   uint64
		Price    uint64
		Buy      bool
		Type     string
	}
	result := []CrowdsaleInfo{}
	saleDataList, err := rpcServer.config.BlockChain.GetAllSaleData()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Error querying crowdsales"))
	}
	beaconHeight := rpcServer.config.BlockChain.BestState.Beacon.BeaconHeight
	for _, saleData := range saleDataList {
		if beaconHeight >= saleData.EndBlock {
			continue
		}

		// Add type for better ux, not blockchain-related
		crowdsaleType := "buyable"
		if saleData.Buy {
			crowdsaleType = "sellable"
		}

		info := CrowdsaleInfo{
			SaleID:   hex.EncodeToString(saleData.SaleID),
			EndBlock: saleData.EndBlock,
			BondID:   saleData.BondID.String(),
			Amount:   saleData.Amount,
			Price:    saleData.Price,
			Buy:      saleData.Buy,
			Type:     crowdsaleType,
		}
		result = append(result, info)
	}
	return result, nil
}

func (rpcServer RpcServer) handleGetDCBBondInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	type dcbBondInfo struct {
		AmountAvailable   uint64
		TotalConstantPaid uint64
	}
	infos := map[string]dcbBondInfo{}
	db := *rpcServer.config.Database
	soldBondTypesBytesArr, err := db.GetSoldBondTypes()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	for _, soldBondTypesBytes := range soldBondTypesBytesArr {
		var bondInfo component.SellingBonds
		err = json.Unmarshal(soldBondTypesBytes, &bondInfo)
		if err != nil {
			continue
		}
		bondIDStr := bondInfo.GetID().String()
		amountAvail, cstPaid := db.GetDCBBondInfo(bondInfo.GetID())
		if amountAvail > 0 {
			infos[bondIDStr] = dcbBondInfo{
				AmountAvailable:   amountAvail,
				TotalConstantPaid: cstPaid,
			}
		}
	}
	return infos, nil
}

func (rpcServer RpcServer) handleGetListDCBProposalBuyingAssets(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// TODO(@0xbunyip): call get list bonds
	buyingAssets := map[string]string{
		"Bond 1":   "4c420b974449ac188c155a7029706b8419a591ee398977d00000000000000000",
		"Constant": common.ConstantID.String(),
	} // From asset name to asset id
	return buyingAssets, nil
}

func (rpcServer RpcServer) handleGetListDCBProposalSellingAssets(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// TODO(@0xbunyip): call get list bonds
	sellingAssets := map[string]string{
		"Constant": common.ConstantID.String(),
		"Bond 2":   "4c420b974449ac188c155a7029706b8419a591ee398977d00000000000000000",
	} // From asset name to asset id
	return sellingAssets, nil
}
