package pdex

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
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
		uncollectedRewards Reward
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
				uncollectedRewards: Reward{
					common.PRVCoinID: 100,
				},
			},
			args: args{
				compareOrderReward: &OrderReward{
					uncollectedRewards: Reward{
						common.PRVCoinID:  100,
						common.PDEXCoinID: 100,
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
