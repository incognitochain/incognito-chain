package pdexv3

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

func TestModifyParams_Hash(t *testing.T) {
	metadata, err := NewPdexv3ParamsModifyingRequest(
		metadataCommon.Pdexv3ModifyParamsMeta,
		Pdexv3Params{
			DefaultFeeRateBPS:               30,
			FeeRateBPS:                      map[string]uint{},
			PRVDiscountPercent:              25,
			TradingProtocolFeePercent:       0,
			TradingStakingPoolRewardPercent: 10,
			PDEXRewardPoolPairsShare:        map[string]uint{},
			StakingPoolsShare: map[string]uint{
				"0000000000000000000000000000000000000000000000000000000000000004": 10,
				"0000000000000000000000000000000000000000000000000000000000000006": 20,
			},
			StakingRewardTokens:          []common.Hash{},
			MintNftRequireAmount:         1000000000,
			MaxOrdersPerNft:              10,
			AutoWithdrawOrderLimitAmount: 100,
			MinPRVReserveTradingRate:     100000,
		},
	)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	t.Logf("Hash: %v", metadata.Hash().String())
}
