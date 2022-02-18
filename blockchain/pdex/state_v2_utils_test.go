package pdex

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
)

func TestMakingVolume_getDiff(t *testing.T) {
	type fields struct {
		volume map[string]*big.Int
	}
	type args struct {
		compareMakingVolume *MakingVolume
		makingVolumeChange  *v2utils.MakingVolumeChange
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *v2utils.MakingVolumeChange
	}{
		{
			name: "Delete making volume",
			fields: fields{
				volume: map[string]*big.Int{
					common.PRVIDStr: big.NewInt(200),
				},
			},
			args: args{
				compareMakingVolume: &MakingVolume{
					volume: map[string]*big.Int{
						common.PRVIDStr:  big.NewInt(200),
						common.PDEXIDStr: big.NewInt(200),
					},
				},
				makingVolumeChange: &v2utils.MakingVolumeChange{},
			},
			want: &v2utils.MakingVolumeChange{
				Volume: map[string]bool{
					common.PDEXIDStr: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			makingVolume := &MakingVolume{
				volume: tt.fields.volume,
			}
			if got := makingVolume.getDiff(tt.args.compareMakingVolume, tt.args.makingVolumeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakingVolume.getDiff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrderReward_getDiff(t *testing.T) {
	type fields struct {
		uncollectedRewards map[common.Hash]*OrderRewardDetail
	}
	type args struct {
		compareOrderReward *OrderReward
		orderRewardChange  *v2utils.OrderRewardChange
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *v2utils.OrderRewardChange
	}{
		{
			name: "Valid input",
			fields: fields{
				uncollectedRewards: map[common.Hash]*OrderRewardDetail{
					common.PRVCoinID: {
						amount: 100,
					},
				},
			},
			args: args{
				compareOrderReward: &OrderReward{
					uncollectedRewards: map[common.Hash]*OrderRewardDetail{
						common.PDEXCoinID: {
							amount: 100,
						},
						common.PRVCoinID: {
							amount: 100,
						},
					},
				},
				orderRewardChange: &v2utils.OrderRewardChange{},
			},
			want: &v2utils.OrderRewardChange{
				UncollectedReward: map[string]bool{
					common.PDEXIDStr: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderReward := &OrderReward{
				uncollectedRewards: tt.fields.uncollectedRewards,
			}
			if got := orderReward.getDiff(tt.args.compareOrderReward, tt.args.orderRewardChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OrderReward.getDiff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShare_getDiff(t *testing.T) {
	tokenID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)

	type fields struct {
		amount                uint64
		lmLockedAmount        uint64
		tradingFees           map[common.Hash]uint64
		lastLPFeesPerShare    map[common.Hash]*big.Int
		lastLMRewardsPerShare map[common.Hash]*big.Int
	}
	type args struct {
		compareShare *Share
		shareChange  *v2utils.ShareChange
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *v2utils.ShareChange
	}{
		{
			name: "Store, update and delete fees",
			fields: fields{
				amount:         100,
				lmLockedAmount: 0,
				tradingFees: map[common.Hash]uint64{
					common.PRVCoinID: 200,
					*tokenID:         300,
				},
				lastLPFeesPerShare: map[common.Hash]*big.Int{
					common.PRVCoinID: big.NewInt(200),
					*tokenID:         big.NewInt(300),
				},
				lastLMRewardsPerShare: map[common.Hash]*big.Int{
					common.PRVCoinID: big.NewInt(400),
				},
			},
			args: args{
				compareShare: &Share{
					amount:         100,
					lmLockedAmount: 0,
					tradingFees: map[common.Hash]uint64{
						common.PRVCoinID:  100,
						common.PDEXCoinID: 200,
					},
					lastLPFeesPerShare: map[common.Hash]*big.Int{
						common.PRVCoinID:  big.NewInt(100),
						common.PDEXCoinID: big.NewInt(200),
					},
					lastLmRewardsPerShare: map[common.Hash]*big.Int{
						common.PRVCoinID: big.NewInt(400),
					},
				},
				shareChange: &v2utils.ShareChange{},
			},
			want: &v2utils.ShareChange{
				IsChanged: false,
				TradingFees: map[string]bool{
					common.PRVIDStr:  true,
					common.PDEXIDStr: true,
					tokenID.String(): true,
				},
				LastLPFeesPerShare: map[string]bool{
					common.PRVIDStr:  true,
					common.PDEXIDStr: true,
					tokenID.String(): true,
				},
				LastLmRewardsPerShare: map[string]bool{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			share := &Share{
				amount:                tt.fields.amount,
				tradingFees:           tt.fields.tradingFees,
				lastLPFeesPerShare:    tt.fields.lastLPFeesPerShare,
				lastLmRewardsPerShare: tt.fields.lastLMRewardsPerShare,
			}
			if got := share.getDiff(tt.args.compareShare, tt.args.shareChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Share.getDiff() = %v, want %v", got, tt.want)
			}
		})
	}
}
