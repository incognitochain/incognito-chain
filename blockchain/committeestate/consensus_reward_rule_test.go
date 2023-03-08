package committeestate

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func TestRewardSplitRuleV1_SplitReward(t *testing.T) {

	type args struct {
		env *SplitRewardEnvironment
	}
	totalRewardYear1 := make(map[common.Hash]uint64)
	totalRewardYear1[common.PRVCoinID] = 8751970
	beaconRewardYear1 := make(map[common.Hash]uint64)
	beaconRewardYear1[common.PRVCoinID] = 1575354
	daoRewardYear1 := make(map[common.Hash]uint64)
	daoRewardYear1[common.PRVCoinID] = 875197
	custodianRewardYear1 := make(map[common.Hash]uint64)
	shardRewardYear1 := make(map[common.Hash]uint64)
	shardRewardYear1[common.PRVCoinID] = 6301419

	totalRewardYear2 := make(map[common.Hash]uint64)
	totalRewardYear2[common.PRVCoinID] = 7964293
	beaconRewardYear2 := make(map[common.Hash]uint64)
	beaconRewardYear2[common.PRVCoinID] = 1449501
	daoRewardYear2 := make(map[common.Hash]uint64)
	daoRewardYear2[common.PRVCoinID] = 716786
	custodianRewardYear2 := make(map[common.Hash]uint64)
	shardRewardYear2 := make(map[common.Hash]uint64)
	shardRewardYear2[common.PRVCoinID] = 5798006
	tests := []struct {
		name    string
		args    args
		want    map[common.Hash]uint64
		want1   map[common.Hash]uint64
		want2   map[common.Hash]uint64
		want3   map[common.Hash]uint64
		wantErr bool
	}{
		{
			name: "year 1",
			args: args{
				env: &SplitRewardEnvironment{
					DAOPercent:                10,
					ActiveShards:              8,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward:               totalRewardYear1,
				},
			},
			want:  beaconRewardYear1,
			want1: shardRewardYear1,
			want2: daoRewardYear1,
			want3: custodianRewardYear1,
		},
		{
			name: "year 2",
			args: args{
				env: &SplitRewardEnvironment{
					DAOPercent:                9,
					ActiveShards:              8,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward:               totalRewardYear2,
				},
			},
			want:  beaconRewardYear2,
			want1: shardRewardYear2,
			want2: daoRewardYear2,
			want3: custodianRewardYear2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RewardSplitRuleV1{}
			got, got1, got2, got3, err := r.SplitReward(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitReward() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitReward() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("SplitReward() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("SplitReward() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("SplitReward() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}

func TestRewardSplitRuleV2_SplitReward(t *testing.T) {

	initTestParams()

	type args struct {
		env *SplitRewardEnvironment
	}
	tests := []struct {
		name    string
		args    args
		want    map[common.Hash]uint64
		want1   map[common.Hash]uint64
		want2   map[common.Hash]uint64
		want3   map[common.Hash]uint64
		wantErr bool
	}{
		{
			name: "Year 1",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						4: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						5: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						6: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						7: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
					},

					DAOPercent:                10,
					ActiveShards:              8,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 1093996,
					},
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 51054,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 933543,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 109399,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 2",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						4: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						5: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						6: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						7: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
					},
					DAOPercent:                9,
					ActiveShards:              8,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 995536,
					},
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 46975,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 858963,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 89598,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 3",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						4: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						5: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						6: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						7: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
					},
					DAOPercent:                8,
					ActiveShards:              8,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 905938,
					},
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 43217,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 790246,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 72475,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 4",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						4: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						5: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						6: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						7: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
					},

					DAOPercent:                7,
					ActiveShards:              8,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 824403,
					},
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 39755,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 726940,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 57708,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 5",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						4: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						5: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						6: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						7: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
					},

					DAOPercent:                6,
					ActiveShards:              8,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 750207,
					},
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 36566,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 668629,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 45012,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 6",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						4: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						5: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						6: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						7: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
					},
					DAOPercent:                5,
					ActiveShards:              8,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 682688,
					},
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 33629,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 614925,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 34134,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 7",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						4: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						5: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						6: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						7: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
					},
					DAOPercent:                4,
					ActiveShards:              8,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 621246,
					},
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 30925,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 565472,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 24849,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 8",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						4: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						5: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						6: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						7: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
					},

					DAOPercent:                3,
					ActiveShards:              8,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 565334,
					},
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 28435,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 519939,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 16960,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 9",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						4: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						5: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						6: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						7: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
					},
					DAOPercent:                3,
					ActiveShards:              8,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 514454,
					},
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 25876,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 473145,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 15433,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RewardSplitRuleV2{}
			got, got1, got2, got3, err := r.SplitReward(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV1.Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitReward() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("splitReward() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("splitReward() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("splitReward() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}

func TestRewardSplitRuleV3_SplitReward(t *testing.T) {

	initTestParams()

	type args struct {
		env *SplitRewardEnvironment
	}
	tests := []struct {
		name    string
		args    args
		want    map[common.Hash]uint64
		want1   map[common.Hash]uint64
		want2   map[common.Hash]uint64
		want3   map[common.Hash]uint64
		wantErr bool
	}{
		{
			name: "Year 1",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3,
						},
						2: []incognitokey.CommitteePublicKey{},
						3: []incognitokey.CommitteePublicKey{},
						4: []incognitokey.CommitteePublicKey{},
						5: []incognitokey.CommitteePublicKey{},
						6: []incognitokey.CommitteePublicKey{},
						7: []incognitokey.CommitteePublicKey{},
					},

					DAOPercent:                10,
					ShardID:                   1,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 546998,
					},
					MaxSubsetCommittees: 2,
					SubsetID:            1,
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 13104,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 479195,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 54699,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 2, Subset 0",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3,
						},
						2: []incognitokey.CommitteePublicKey{},
						3: []incognitokey.CommitteePublicKey{},
						4: []incognitokey.CommitteePublicKey{},
						5: []incognitokey.CommitteePublicKey{},
						6: []incognitokey.CommitteePublicKey{},
						7: []incognitokey.CommitteePublicKey{},
					},

					DAOPercent:                9,
					ShardID:                   1,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 497768,
					},
					MaxSubsetCommittees: 2,
					SubsetID:            0,
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 12057,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 440912,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 44799,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 2, Subset 1",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3,
						},
						2: []incognitokey.CommitteePublicKey{},
						3: []incognitokey.CommitteePublicKey{},
						4: []incognitokey.CommitteePublicKey{},
						5: []incognitokey.CommitteePublicKey{},
						6: []incognitokey.CommitteePublicKey{},
						7: []incognitokey.CommitteePublicKey{},
					},

					DAOPercent:                9,
					ShardID:                   1,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 497768,
					},
					MaxSubsetCommittees: 2,
					SubsetID:            1,
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 12057,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 440912,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 44799,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 3, Subset 0, shard committee size = 61",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0,
						},
						2: []incognitokey.CommitteePublicKey{},
						3: []incognitokey.CommitteePublicKey{},
						4: []incognitokey.CommitteePublicKey{},
						5: []incognitokey.CommitteePublicKey{},
						6: []incognitokey.CommitteePublicKey{},
						7: []incognitokey.CommitteePublicKey{},
					},

					DAOPercent:                8,
					ShardID:                   1,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 452696,
					},
					MaxSubsetCommittees: 2,
					SubsetID:            0,
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 11433,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 405048,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 36215,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 3, Subset 1, shard committee size = 61",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0,
						},
						2: []incognitokey.CommitteePublicKey{},
						3: []incognitokey.CommitteePublicKey{},
						4: []incognitokey.CommitteePublicKey{},
						5: []incognitokey.CommitteePublicKey{},
						6: []incognitokey.CommitteePublicKey{},
						7: []incognitokey.CommitteePublicKey{},
					},
					DAOPercent:                8,
					ShardID:                   1,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 452696,
					},
					MaxSubsetCommittees: 2,
					SubsetID:            1,
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 11804,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 404677,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 36215,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RewardSplitRuleV3{}
			got, got1, got2, got3, err := r.SplitReward(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV3.SplitReward() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.SplitReward() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeStateV3.SplitReward() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeStateV3.SplitReward() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("BeaconCommitteeStateV3.SplitReward() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}

func TestRewardSplitRuleV4_SplitReward(t *testing.T) {
	initTestParams()
	basicReward := uint64(1000000000)
	type args struct {
		env *SplitRewardEnvironment
	}
	tests := []struct {
		name    string
		args    args
		want    map[common.Hash]uint64
		want1   map[common.Hash]uint64
		want2   map[common.Hash]uint64
		want3   map[common.Hash]uint64
		wantErr bool
	}{
		{
			name: "Year 1",
			args: args{
				env: &SplitRewardEnvironment{
					BeaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
					ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						4: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						5: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						6: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
						7: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
							*incKey4, *incKey5,
						},
					},

					DAOPercent:                10,
					CommitteePercent:          30,
					TotalCreditSize:           3000,
					BeaconCreditSize:          200,
					ActiveShards:              8,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID:           basicReward * 350 * 8,
						common.ConfidentialAssetID: basicReward * 100, //just random amount
					},
				},
			},
			want: map[common.Hash]uint64{ //for all beacon
				common.PRVCoinID:           basicReward * 350 * 8 * 90 / 100 * 200 / 3000,
				common.ConfidentialAssetID: basicReward * 100 * 90 / 100 * 200 / 3000,
			},
			want1: map[common.Hash]uint64{ //for all shard, included delegation
				common.PRVCoinID:           basicReward * 350 * 8 * 90 / 100 * 2800 / 3000,
				common.ConfidentialAssetID: basicReward * 100 * 90 / 100 * 2800 / 3000,
			},
			want2: map[common.Hash]uint64{ //for DAO
				common.PRVCoinID:           basicReward * 350 * 8 * 10 / 100,
				common.ConfidentialAssetID: basicReward * 100 * 10 / 100,
			},
			want3:   map[common.Hash]uint64{}, // for Custodians
			wantErr: false,
		},
		// {
		// 	name: "Year 2",
		// 	args: args{
		// 		env: &SplitRewardEnvironment{
		// 			BeaconCommittee: []incognitokey.CommitteePublicKey{
		// 				*incKey0, *incKey, *incKey2, *incKey3,
		// 			},
		// 			ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
		// 				0: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				1: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				2: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				3: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				4: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				5: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				6: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				7: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 			},
		// 			DAOPercent:                9,
		// 			CommitteePercent:          30,
		// 			TotalCreditSize:           3000,
		// 			BeaconCreditSize:          200,
		// 			ActiveShards:              8,
		// 			IsSplitRewardForCustodian: false,
		// 			PercentCustodianReward:    0,
		// 			TotalReward: map[common.Hash]uint64{
		// 				common.PRVCoinID: 995536,
		// 			},
		// 		},
		// 	},
		// 	want: map[common.Hash]uint64{
		// 		common.PRVCoinID: 46975,
		// 	},
		// 	want1: map[common.Hash]uint64{
		// 		common.PRVCoinID: 858963,
		// 	},
		// 	want2: map[common.Hash]uint64{
		// 		common.PRVCoinID: 89598,
		// 	},
		// 	want3:   map[common.Hash]uint64{},
		// 	wantErr: false,
		// },
		// {
		// 	name: "Year 3",
		// 	args: args{
		// 		env: &SplitRewardEnvironment{
		// 			BeaconCommittee: []incognitokey.CommitteePublicKey{
		// 				*incKey0, *incKey, *incKey2, *incKey3,
		// 			},
		// 			ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
		// 				0: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				1: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				2: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				3: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				4: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				5: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				6: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				7: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 			},
		// 			DAOPercent:                8,
		// 			CommitteePercent:          30,
		// 			TotalCreditSize:           3000,
		// 			BeaconCreditSize:          200,
		// 			ActiveShards:              8,
		// 			IsSplitRewardForCustodian: false,
		// 			PercentCustodianReward:    0,
		// 			TotalReward: map[common.Hash]uint64{
		// 				common.PRVCoinID: 905938,
		// 			},
		// 		},
		// 	},
		// 	want: map[common.Hash]uint64{
		// 		common.PRVCoinID: 43217,
		// 	},
		// 	want1: map[common.Hash]uint64{
		// 		common.PRVCoinID: 790246,
		// 	},
		// 	want2: map[common.Hash]uint64{
		// 		common.PRVCoinID: 72475,
		// 	},
		// 	want3:   map[common.Hash]uint64{},
		// 	wantErr: false,
		// },
		// {
		// 	name: "Year 4",
		// 	args: args{
		// 		env: &SplitRewardEnvironment{
		// 			BeaconCommittee: []incognitokey.CommitteePublicKey{
		// 				*incKey0, *incKey, *incKey2, *incKey3,
		// 			},
		// 			ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
		// 				0: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				1: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				2: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				3: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				4: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				5: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				6: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				7: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 			},

		// 			DAOPercent:                7,
		// 			CommitteePercent:          30,
		// 			TotalCreditSize:           3000,
		// 			BeaconCreditSize:          200,
		// 			ActiveShards:              8,
		// 			IsSplitRewardForCustodian: false,
		// 			PercentCustodianReward:    0,
		// 			TotalReward: map[common.Hash]uint64{
		// 				common.PRVCoinID: 824403,
		// 			},
		// 		},
		// 	},
		// 	want: map[common.Hash]uint64{
		// 		common.PRVCoinID: 39755,
		// 	},
		// 	want1: map[common.Hash]uint64{
		// 		common.PRVCoinID: 726940,
		// 	},
		// 	want2: map[common.Hash]uint64{
		// 		common.PRVCoinID: 57708,
		// 	},
		// 	want3:   map[common.Hash]uint64{},
		// 	wantErr: false,
		// },
		// {
		// 	name: "Year 5",
		// 	args: args{
		// 		env: &SplitRewardEnvironment{
		// 			BeaconCommittee: []incognitokey.CommitteePublicKey{
		// 				*incKey0, *incKey, *incKey2, *incKey3,
		// 			},
		// 			ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
		// 				0: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				1: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				2: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				3: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				4: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				5: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				6: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				7: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 			},

		// 			DAOPercent:                6,
		// 			CommitteePercent:          30,
		// 			TotalCreditSize:           3000,
		// 			BeaconCreditSize:          200,
		// 			ActiveShards:              8,
		// 			IsSplitRewardForCustodian: false,
		// 			PercentCustodianReward:    0,
		// 			TotalReward: map[common.Hash]uint64{
		// 				common.PRVCoinID: 750207,
		// 			},
		// 		},
		// 	},
		// 	want: map[common.Hash]uint64{
		// 		common.PRVCoinID: 36566,
		// 	},
		// 	want1: map[common.Hash]uint64{
		// 		common.PRVCoinID: 668629,
		// 	},
		// 	want2: map[common.Hash]uint64{
		// 		common.PRVCoinID: 45012,
		// 	},
		// 	want3:   map[common.Hash]uint64{},
		// 	wantErr: false,
		// },
		// {
		// 	name: "Year 6",
		// 	args: args{
		// 		env: &SplitRewardEnvironment{
		// 			BeaconCommittee: []incognitokey.CommitteePublicKey{
		// 				*incKey0, *incKey, *incKey2, *incKey3,
		// 			},
		// 			ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
		// 				0: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				1: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				2: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				3: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				4: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				5: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				6: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				7: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 			},
		// 			DAOPercent:                5,
		// 			CommitteePercent:          30,
		// 			TotalCreditSize:           3000,
		// 			BeaconCreditSize:          200,
		// 			ActiveShards:              8,
		// 			IsSplitRewardForCustodian: false,
		// 			PercentCustodianReward:    0,
		// 			TotalReward: map[common.Hash]uint64{
		// 				common.PRVCoinID: 682688,
		// 			},
		// 		},
		// 	},
		// 	want: map[common.Hash]uint64{
		// 		common.PRVCoinID: 33629,
		// 	},
		// 	want1: map[common.Hash]uint64{
		// 		common.PRVCoinID: 614925,
		// 	},
		// 	want2: map[common.Hash]uint64{
		// 		common.PRVCoinID: 34134,
		// 	},
		// 	want3:   map[common.Hash]uint64{},
		// 	wantErr: false,
		// },
		// {
		// 	name: "Year 7",
		// 	args: args{
		// 		env: &SplitRewardEnvironment{
		// 			BeaconCommittee: []incognitokey.CommitteePublicKey{
		// 				*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
		// 			},
		// 			ShardCommittee: map[byte][]incognitokey.CommitteePublicKey{
		// 				0: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				1: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				2: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				3: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				4: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				5: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				6: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 				7: []incognitokey.CommitteePublicKey{
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
		// 					*incKey4, *incKey5,
		// 				},
		// 			},
		// 			DAOPercent:                4,
		// 			CommitteePercent:          30,
		// 			TotalCreditSize:           3000,
		// 			BeaconCreditSize:          200,
		// 			ActiveShards:              8,
		// 			IsSplitRewardForCustodian: false,
		// 			PercentCustodianReward:    0,
		// 			TotalReward: map[common.Hash]uint64{
		// 				common.PRVCoinID: 621246,
		// 			},
		// 		},
		// 	},
		// 	want: map[common.Hash]uint64{
		// 		common.PRVCoinID: 30925,
		// 	},
		// 	want1: map[common.Hash]uint64{
		// 		common.PRVCoinID: 565472,
		// 	},
		// 	want2: map[common.Hash]uint64{
		// 		common.PRVCoinID: 24849,
		// 	},
		// 	want3:   map[common.Hash]uint64{},
		// 	wantErr: false,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rV4 := &RewardSplitRuleV4{}
			rV2 := &RewardSplitRuleV2{}
			gotV4, got1V4, got2V4, got3V4, err := rV4.SplitReward(tt.args.env)
			gotV2, got1V2, _, _, err := rV2.SplitReward(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV1.Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotV4, tt.want) {
				t.Errorf("splitReward() got = %v, want %v", gotV4, tt.want)
			}
			if !reflect.DeepEqual(got1V4, tt.want1) {
				t.Errorf("splitReward() got1 = %v, want %v", got1V4, tt.want1)
			}
			if !reflect.DeepEqual(got2V4, tt.want2) {
				t.Errorf("splitReward() got2 = %v, want %v", got2V4, tt.want2)
			}
			if !reflect.DeepEqual(got3V4, tt.want3) {
				t.Errorf("splitReward() got3 = %v, want %v", got3V4, tt.want3)
			}
			v4BeaconPRV := gotV4[common.PRVCoinID] / 4
			v2BeaconPRV := gotV2[common.PRVCoinID] / 4
			v4ShardPRV := got1V4[common.PRVCoinID] / (32 * 8)
			v2ShardPRV := got1V2[common.PRVCoinID] / (32 * 8)
			fmt.Printf("V4 PRV for each Beacon %+v; V2: %+v \n", v4BeaconPRV, v2BeaconPRV)
			fmt.Printf("V4 PRV for each Shard %+v; V2: %+v \n", v4ShardPRV, v2ShardPRV)
		})
	}
}
