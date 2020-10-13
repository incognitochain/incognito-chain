package committeestate

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/incognitokey"

	"github.com/incognitochain/incognito-chain/common"
)

func TestBeaconCommitteeEngineV2_SplitReward(t *testing.T) {

	hash, _ := common.Hash{}.NewHashFromStr("123")

	initLog()
	initPublicKey()

	type fields struct {
		beaconHeight                      uint64
		beaconHash                        common.Hash
		finalBeaconCommitteeStateV2       *BeaconCommitteeStateV2
		uncommittedBeaconCommitteeStateV2 *BeaconCommitteeStateV2
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[common.Hash]uint64
		want1   map[common.Hash]uint64
		want2   map[common.Hash]uint64
		want3   map[common.Hash]uint64
		wantErr bool
	}{
		{
			name: "Year 1",
			fields: fields{
				beaconHash:   *hash,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					DAOPercent:                10,
					ActiveShards:              4,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalRewardForShard: map[common.Hash]uint64{
						*hash: 3000,
					},
				},
			},
			want: map[common.Hash]uint64{
				*hash: 675,
			},
			want1: map[common.Hash]uint64{
				*hash: 2025,
			},
			want2: map[common.Hash]uint64{
				*hash: 300,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 2",
			fields: fields{
				beaconHash:   *hash,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					DAOPercent:                9,
					ActiveShards:              4,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalRewardForShard: map[common.Hash]uint64{
						*hash: 3000,
					},
				},
			},
			want: map[common.Hash]uint64{
				*hash: 683,
			},
			want1: map[common.Hash]uint64{
				*hash: 2047,
			},
			want2: map[common.Hash]uint64{
				*hash: 270,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 3",
			fields: fields{
				beaconHash:   *hash,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					DAOPercent:                8,
					ActiveShards:              4,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalRewardForShard: map[common.Hash]uint64{
						*hash: 3000,
					},
				},
			},
			want: map[common.Hash]uint64{
				*hash: 690,
			},
			want1: map[common.Hash]uint64{
				*hash: 2070,
			},
			want2: map[common.Hash]uint64{
				*hash: 240,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 4",
			fields: fields{
				beaconHash:   *hash,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					DAOPercent:                7,
					ActiveShards:              4,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalRewardForShard: map[common.Hash]uint64{
						*hash: 3000,
					},
				},
			},
			want: map[common.Hash]uint64{
				*hash: 698,
			},
			want1: map[common.Hash]uint64{
				*hash: 2092,
			},
			want2: map[common.Hash]uint64{
				*hash: 210,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 5",
			fields: fields{
				beaconHash:   *hash,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					DAOPercent:                6,
					ActiveShards:              4,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalRewardForShard: map[common.Hash]uint64{
						*hash: 3000,
					},
				},
			},
			want: map[common.Hash]uint64{
				*hash: 705,
			},
			want1: map[common.Hash]uint64{
				*hash: 2115,
			},
			want2: map[common.Hash]uint64{
				*hash: 180,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 6",
			fields: fields{
				beaconHash:   *hash,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					DAOPercent:                5,
					ActiveShards:              4,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalRewardForShard: map[common.Hash]uint64{
						*hash: 3000,
					},
				},
			},
			want: map[common.Hash]uint64{
				*hash: 713,
			},
			want1: map[common.Hash]uint64{
				*hash: 2137,
			},
			want2: map[common.Hash]uint64{
				*hash: 150,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 7",
			fields: fields{
				beaconHash:   *hash,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					DAOPercent:                4,
					ActiveShards:              4,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalRewardForShard: map[common.Hash]uint64{
						*hash: 3000,
					},
				},
			},
			want: map[common.Hash]uint64{
				*hash: 720,
			},
			want1: map[common.Hash]uint64{
				*hash: 2160,
			},
			want2: map[common.Hash]uint64{
				*hash: 120,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 8",
			fields: fields{
				beaconHash:   *hash,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					DAOPercent:                3,
					ActiveShards:              4,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalRewardForShard: map[common.Hash]uint64{
						*hash: 3000,
					},
				},
			},
			want: map[common.Hash]uint64{
				*hash: 728,
			},
			want1: map[common.Hash]uint64{
				*hash: 2182,
			},
			want2: map[common.Hash]uint64{
				*hash: 90,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
		{
			name: "Year 9",
			fields: fields{
				beaconHash:   *hash,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						2: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
						3: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					DAOPercent:                3,
					ActiveShards:              4,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalRewardForShard: map[common.Hash]uint64{
						*hash: 3000,
					},
				},
			},
			want: map[common.Hash]uint64{
				*hash: 728,
			},
			want1: map[common.Hash]uint64{
				*hash: 2182,
			},
			want2: map[common.Hash]uint64{
				*hash: 90,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeEngineV2{
				beaconHeight:                      tt.fields.beaconHeight,
				beaconHash:                        tt.fields.beaconHash,
				finalBeaconCommitteeStateV2:       tt.fields.finalBeaconCommitteeStateV2,
				uncommittedBeaconCommitteeStateV2: tt.fields.uncommittedBeaconCommitteeStateV2,
			}
			got, got1, got2, got3, err := b.SplitReward(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV2.SplitReward() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeEngineV2.SplitReward() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeEngineV2.SplitReward() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeEngineV2.SplitReward() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("BeaconCommitteeEngineV2.SplitReward() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}
