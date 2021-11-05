package v2utils

import (
	"fmt"
	"math"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
)

// common
const (
	BPS = 10000
)

// params
const (
	MaxFeeRateBPS         = 200
	MaxPRVDiscountPercent = 75
)

// PDEX token
const (
	GenesisMintingAmount = 5000000                  // without mulitply with denominating rate
	MintingBlockReward   = 45000000                 // without multiply with denominating rate
	MintingBlocks        = 3600 * 24 * 30 * 60 / 40 // 60 months
	DecayIntervals       = 30
	DecayRateBPS         = 500 // 5%
)

// nft hash prefix
var (
	TotalPDEXReward         = MintingBlockReward * math.Pow(10, common.PDEXDenominatingDecimal)
	DecayRate               = float64(DecayRateBPS) / float64(BPS)
	PDEXRewardFirstInterval = uint64(TotalPDEXReward * DecayRate / (1 - math.Pow(1-DecayRate, DecayIntervals)))
)

func TestMintPDEX(t *testing.T) {
	config.AbortParam()
	config.Param().PDexParams.Pdexv3BreakPointHeight = 1
	config.Param().EpochParam.NumberOfBlockInEpoch = 5

	// lastReward := uint64(0)
	sum := uint64(0)
	mintedEpochs := uint64(0)
	for beaconHeight := uint64(0); beaconHeight < 2*MintingBlocks*config.Param().EpochParam.NumberOfBlockInEpoch; beaconHeight++ {
		reward := GetPDEXRewardsForBlock(
			beaconHeight, mintedEpochs, true, 2,
			MintingBlocks, DecayIntervals, PDEXRewardFirstInterval,
			DecayRateBPS, BPS,
		)
		if reward > 0 {
			mintedEpochs++
		}

		// if reward != lastReward {
		//  t.Logf("beaconHeight: %v, reward: %v\n", beaconHeight, reward)
		// }
		sum += reward
		// lastReward = reward
	}
	fmt.Printf("sum: %v\n", sum)
}
