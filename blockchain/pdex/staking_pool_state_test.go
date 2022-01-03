package pdex

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

func TestStakingPoolState_updateLiquidity(t *testing.T) {
	type fields struct {
		liquidity uint64
		stakers   map[string]*Staker
	}
	type args struct {
		accessID     string
		liquidity    uint64
		beaconHeight uint64
		accessOTA    []byte
		operator     byte
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		wantErr            bool
	}{
		{
			name: "Remove liquidity from invalid staker",
			fields: fields{
				liquidity: 0,
				stakers:   map[string]*Staker{},
			},
			args: args{
				accessID:     nftID,
				liquidity:    50,
				beaconHeight: 20,
				operator:     subOperator,
			},
			wantErr: true,
		},
		{
			name: "Add new liquidity",
			fields: fields{
				liquidity: 0,
				stakers:   map[string]*Staker{},
			},
			args: args{
				accessID:     nftID,
				liquidity:    100,
				beaconHeight: 20,
				operator:     addOperator,
			},
			fieldsAfterProcess: fields{
				liquidity: 100,
				stakers: map[string]*Staker{
					nftID: &Staker{
						liquidity:           100,
						rewards:             map[common.Hash]uint64{},
						lastRewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Add available liquidity",
			fields: fields{
				liquidity: 100,
				stakers: map[string]*Staker{
					nftID1: &Staker{
						liquidity:           100,
						rewards:             map[common.Hash]uint64{},
						lastRewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
			},
			args: args{
				accessID:     nftID1,
				liquidity:    50,
				beaconHeight: 30,
				operator:     addOperator,
			},
			fieldsAfterProcess: fields{
				liquidity: 150,
				stakers: map[string]*Staker{
					nftID1: &Staker{
						liquidity:           150,
						rewards:             map[common.Hash]uint64{},
						lastRewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Add available liquidity - 2",
			fields: fields{
				liquidity: 100,
				stakers: map[string]*Staker{
					nftID1: &Staker{
						liquidity:           100,
						rewards:             map[common.Hash]uint64{},
						lastRewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
			},
			args: args{
				accessID:     nftID,
				liquidity:    50,
				beaconHeight: 30,
				operator:     addOperator,
			},
			fieldsAfterProcess: fields{
				liquidity: 150,
				stakers: map[string]*Staker{
					nftID1: &Staker{
						liquidity:           100,
						rewards:             map[common.Hash]uint64{},
						lastRewardsPerShare: map[common.Hash]*big.Int{},
					},
					nftID: &Staker{
						liquidity:           50,
						rewards:             map[common.Hash]uint64{},
						lastRewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Remove liquidity",
			fields: fields{
				liquidity: 100,
				stakers: map[string]*Staker{
					nftID1: &Staker{
						liquidity:           100,
						rewards:             map[common.Hash]uint64{},
						lastRewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
			},
			args: args{
				accessID:     nftID1,
				liquidity:    50,
				beaconHeight: 30,
				operator:     subOperator,
			},
			fieldsAfterProcess: fields{
				liquidity: 50,
				stakers: map[string]*Staker{
					nftID1: &Staker{
						liquidity:           50,
						rewards:             map[common.Hash]uint64{},
						lastRewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Remove liquidity - 2",
			fields: fields{
				liquidity: 150,
				stakers: map[string]*Staker{
					nftID1: &Staker{
						liquidity:           100,
						rewards:             map[common.Hash]uint64{},
						lastRewardsPerShare: map[common.Hash]*big.Int{},
					},
					nftID: &Staker{
						liquidity:           50,
						rewards:             map[common.Hash]uint64{},
						lastRewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
			},
			args: args{
				accessID:     nftID1,
				liquidity:    50,
				beaconHeight: 30,
				operator:     subOperator,
			},
			fieldsAfterProcess: fields{
				liquidity: 100,
				stakers: map[string]*Staker{
					nftID1: &Staker{
						liquidity:           50,
						rewards:             map[common.Hash]uint64{},
						lastRewardsPerShare: map[common.Hash]*big.Int{},
					},
					nftID: &Staker{
						liquidity:           50,
						rewards:             map[common.Hash]uint64{},
						lastRewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StakingPoolState{
				liquidity:       tt.fields.liquidity,
				stakers:         tt.fields.stakers,
				rewardsPerShare: map[common.Hash]*big.Int{},
			}
			if err := s.updateLiquidity(tt.args.accessID, tt.args.liquidity, tt.args.beaconHeight, tt.args.accessOTA, tt.args.operator); (err != nil) != tt.wantErr {
				t.Errorf("StakingPoolState.updateLiquidity() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.fieldsAfterProcess.liquidity, s.liquidity) {
				t.Errorf("liquidity got = %v, want %v", s.liquidity, tt.fieldsAfterProcess.liquidity)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.fieldsAfterProcess.stakers, s.stakers) {
				t.Errorf("stakers got = %v, want %v", s.stakers, tt.fieldsAfterProcess.stakers)
				return
			}
		})
	}
}
