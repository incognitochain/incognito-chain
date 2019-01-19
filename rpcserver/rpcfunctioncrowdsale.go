package rpcserver

func (rpcServer RpcServer) handleGetListOngoingCrowdsale(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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
	if err != nil {
		for i, endBlock := range endBlocks {
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
	}
	return result, nil
}
