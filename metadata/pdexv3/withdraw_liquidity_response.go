package pdexv3

import metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"

type WithdrawLiquidityResponse struct {
	metadataCommon.MetadataBase
	status  string
	txReqID string
}

func NewWithdrawLiquidityResponse() *WithdrawLiquidityResponse {
	return &WithdrawLiquidityResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.WithDrawRewardResponseMeta,
		},
	}
}

func NewWithdrawLiquidityResponseWithValue(status, txReqID string) *WithdrawLiquidityResponse {
	return &WithdrawLiquidityResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.WithDrawRewardResponseMeta,
		},
		status:  status,
		txReqID: txReqID,
	}
}
