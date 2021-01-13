package committeestate

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate/mocks"
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBeaconCommitteeEngineV2_SplitReward(t *testing.T) {

	hash, _ := common.Hash{}.NewHashFromStr("123")

	initLog()
	initTestParams()

	type fields struct {
		beaconCommitteeEngineSlashingBase
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
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{
										*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
									},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
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
									mu: &sync.RWMutex{},
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
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
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{
										*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
									},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
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
									mu: &sync.RWMutex{},
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
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
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{
										*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
									},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
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
									mu: &sync.RWMutex{},
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
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
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{
										*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
									},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
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
									mu: &sync.RWMutex{},
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
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
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{
										*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
									},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
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
									mu: &sync.RWMutex{},
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
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
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{
										*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
									},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
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
									mu: &sync.RWMutex{},
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
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
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{
										*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
									},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
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
									mu: &sync.RWMutex{},
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
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
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{
										*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
									},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
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
									mu: &sync.RWMutex{},
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
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
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{
										*incKey0, *incKey, *incKey2, *incKey3, *incKey, *incKey2, *incKey3,
									},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
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
									mu: &sync.RWMutex{},
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
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
			b := &BeaconCommitteeEngineV2{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHeight:     tt.fields.beaconHeight,
						beaconHash:       tt.fields.beaconHash,
						finalState:       tt.fields.finalState,
						uncommittedState: tt.fields.uncommittedState,
					},
				},
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

func TestBeaconCommitteeEngineV2_UpdateCommitteeState(t *testing.T) {
	hash, _ := common.Hash{}.NewHashFromStr("123")
	tempHash, _ := common.Hash{}.NewHashFromStr("456")
	initTestParams()
	initLog()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	paymentAddress0, err := wallet.Base58CheckDeserialize(paymentAddreessKey0)
	assert.Nil(t, err)
	rewardReceiverkey0 := incKey0.GetIncKeyBase58()
	rewardReceiverkey4 := incKey4.GetIncKeyBase58()
	rewardReceiverKey := incKey.GetIncKeyBase58()
	paymentAddress, err := wallet.Base58CheckDeserialize(paymentAddreessKey0)
	assert.Nil(t, err)

	committeeChangeProcessStakeInstruction := NewCommitteeChange()
	committeeChangeProcessStakeInstruction.NextEpochShardCandidateAdded = []incognitokey.CommitteePublicKey{
		*incKey0,
	}

	committeeChangeProcessRandomInstruction := NewCommitteeChange()
	committeeChangeProcessRandomInstruction.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey6,
	}
	committeeChangeProcessRandomInstruction.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey6,
	}

	committeeChangeProcessStopAutoStakeInstruction := NewCommitteeChange()
	committeeChangeProcessStopAutoStakeInstruction.StopAutoStake = []string{key5}

	committeeChangeProcessSwapShardInstruction := NewCommitteeChange()
	committeeChangeProcessSwapShardInstruction.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey4,
	}
	committeeChangeProcessSwapShardInstruction.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey0,
	}
	committeeChangeProcessSwapShardInstruction.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey4,
	}
	committeeChangeProcessSwapShardInstruction.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey0,
	}

	committeeChangeProcessSwapShardInstruction2KeysIn := NewCommitteeChange()
	committeeChangeProcessSwapShardInstruction2KeysIn.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey0, *incKey,
	}
	committeeChangeProcessSwapShardInstruction2KeysIn.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey0, *incKey,
	}

	committeeChangeProcessSwapShardInstruction2KeysOut := NewCommitteeChange()
	committeeChangeProcessSwapShardInstruction2KeysOut.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey, *incKey4,
	}
	committeeChangeProcessSwapShardInstruction2KeysOut.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey, *incKey4,
	}

	committeeChangeProcessSwapShardInstruction2Keys := NewCommitteeChange()
	committeeChangeProcessSwapShardInstruction2Keys.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey2, *incKey3,
	}
	committeeChangeProcessSwapShardInstruction2Keys.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey2, *incKey3,
	}

	committeeChangeProcessSwapShardInstruction2Keys.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey0, *incKey,
	}
	committeeChangeProcessSwapShardInstruction2Keys.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey0, *incKey,
	}

	committeeChangeProcessUnstakeInstruction := NewCommitteeChange()
	committeeChangeProcessUnstakeInstruction.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{*incKey0}
	committeeChangeProcessUnstakeInstruction.RemovedStaker = []string{key0}
	committeeChangeProcessUnstakeInstruction.TermsRemoved = []string{}

	committeeChangeSwapRuleV3 := NewCommitteeChange()
	committeeChangeSwapRuleV3.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey13, *incKey14,
	}
	committeeChangeSwapRuleV3.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey13, *incKey14,
	}
	committeeChangeSwapRuleV3.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey11,
	}
	committeeChangeSwapRuleV3.RemovedStaker = []string{key11}
	committeeChangeSwapRuleV3.SlashingCommittee[0] = []string{key11}

	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey4, *incKey11},
		map[string]privacy.PaymentAddress{
			rewardReceiverkey0:         paymentAddress0.KeySet.PaymentAddress,
			rewardReceiverKey:          paymentAddress.KeySet.PaymentAddress,
			rewardReceiverkey4:         paymentAddress0.KeySet.PaymentAddress,
			incKey11.GetIncKeyBase58(): paymentAddress.KeySet.PaymentAddress,
		},
		map[string]bool{
			key0:  true,
			key:   true,
			key4:  true,
			key11: true,
		},
		map[string]common.Hash{
			key0:  *hash,
			key:   *tempHash,
			key4:  *tempHash,
			key11: *hash,
		},
	)

	finalMu := &sync.RWMutex{}
	unCommitteedMu := &sync.RWMutex{}

	//define swapRule interface
	swapRuleSingleInstructionOut := &mocks.SwapRule{}
	swapRuleSingleInstructionOut.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey0,
			},
			OutPublicKeys: []string{key0},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey4,
			},
			InPublicKeys: []string{key4},
			ChainID:      0,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key, key2, key3, key4}, []string{}, []string{}, []string{key0})
	swapRuleSingleInstructionOut.On("Version").Return(swapRuleTestVersion)

	swapRuleIn2Keys := &mocks.SwapRule{}
	swapRuleIn2Keys.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey0, *incKey,
			},
			InPublicKeys:        []string{key0, key},
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{},
			OutPublicKeys:       []string{},
			ChainID:             0,
			Type:                instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key2, key3, key4, key5, key6, key7, key0, key}, []string{}, []string{}, []string{})
	swapRuleIn2Keys.On("Version").Return(swapRuleTestVersion)

	swapRuleOut2Keys := &mocks.SwapRule{}
	swapRuleOut2Keys.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey, *incKey4,
			},
			OutPublicKeys:      []string{key, key4},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{},
			InPublicKeys:       []string{},
			ChainID:            0,
			Type:               instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key0, key5, key2, key3, key6}, []string{}, []string{}, []string{key, key4})
	swapRuleOut2Keys.On("Version").Return(swapRuleTestVersion)

	swapRuleInAndOut2Keys := &mocks.SwapRule{}
	swapRuleInAndOut2Keys.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey0, *incKey,
			},
			OutPublicKeys: []string{key0, key},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey2, *incKey3,
			},
			InPublicKeys: []string{key2, key3},
			ChainID:      0,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key4, key5, key6, key7, key2, key3}, []string{}, []string{}, []string{key0, key})
	swapRuleInAndOut2Keys.On("Version").Return(swapRuleTestVersion)

	swapRuleV3 := &mocks.SwapRule{}
	swapRuleV3.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey11,
			},
			OutPublicKeys: []string{key11},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey13, *incKey14,
			},
			InPublicKeys: []string{key13, key14},
			ChainID:      0,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{
			key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key12, key13, key14,
		}, []string{}, []string{key11}, []string{})
	swapRuleV3.On("Version").Return(swapRuleTestVersion)

	type fields struct {
		beaconCommitteeEngineSlashingBase beaconCommitteeEngineSlashingBase
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *BeaconCommitteeStateHash
		want1              *CommitteeChange
		want2              [][]string
		wantErr            bool
	}{
		{
			name: "Process Stake Instruction",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey, *incKey2, *incKey3, *incKey4,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey5,
										},
									},
									autoStake:      map[string]bool{},
									rewardReceiver: map[string]privacy.PaymentAddress{},
									stakingTx:      map[string]common.Hash{},
									mu:             finalMu,
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: unCommitteedMu,
								},
							},
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState:   &BeaconCommitteeStateV2{},
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey, *incKey2, *incKey3, *incKey4,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey5,
										},
									},
									autoStake: map[string]bool{
										key0: true,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{
										incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
									},
									stakingTx: map[string]common.Hash{
										key0: *hash,
									},
									mu: unCommitteedMu,
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{
									*incKey0,
								},
							},
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.STAKE_ACTION,
							key0,
							instruction.SHARD_INST,
							hash.String(),
							paymentAddreessKey0,
							"true",
						},
					},
					ConsensusStateDB:      sDB,
					MaxShardCommitteeSize: 4,
					ActiveShards:          1,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessStakeInstruction,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Random Instruction",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey, *incKey2, *incKey3, *incKey4,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{},
									},
									mu:             finalMu,
									autoStake:      map[string]bool{},
									rewardReceiver: map[string]privacy.PaymentAddress{},
									stakingTx:      map[string]common.Hash{},
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{
									*incKey6,
								},
								numberOfAssignedCandidates: 1,
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: unCommitteedMu,
								},
							},
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHeight: 10,
						beaconHash:   *hash,
						finalState:   &BeaconCommitteeStateV2{},
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey, *incKey2, *incKey3, *incKey4,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey6,
										},
									},
									mu:             unCommitteedMu,
									autoStake:      map[string]bool{},
									rewardReceiver: map[string]privacy.PaymentAddress{},
									stakingTx:      map[string]common.Hash{},
								},
								shardCommonPool:            []incognitokey.CommitteePublicKey{},
								numberOfAssignedCandidates: 0,
							},
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.RANDOM_ACTION,
							"800000",
							"120000",
							"350000",
							"190000",
						},
					},
					ActiveShards:          1,
					BeaconHeight:          100,
					MaxShardCommitteeSize: 5,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessRandomInstruction,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Stop Auto Stake Instruction",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey, *incKey2, *incKey3, *incKey4,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey5,
										},
									},
									mu: finalMu,
									autoStake: map[string]bool{
										key5: true,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{
										key5: paymentAddress0.KeySet.PaymentAddress,
									},
									stakingTx: map[string]common.Hash{
										key5: *hash,
									},
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: unCommitteedMu,
								},
							},
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey, *incKey2, *incKey3, *incKey4,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey5,
										},
									},
									mu: unCommitteedMu,
									autoStake: map[string]bool{
										key5: false,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{
										key5: paymentAddress0.KeySet.PaymentAddress,
									},
									stakingTx: map[string]common.Hash{
										key5: *hash,
									},
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{},
							},
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.STOP_AUTO_STAKE_ACTION,
							key5,
						},
					},
					MaxShardCommitteeSize: 4,
					ActiveShards:          1,
					ConsensusStateDB:      sDB,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessStopAutoStakeInstruction,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Swap Shard Instructions",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHeight: 5,
						beaconHash:   *hash,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey0, *incKey, *incKey2, *incKey3,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey4,
										},
									},
									mu: finalMu,
									autoStake: map[string]bool{
										key0: true,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{
										key0: paymentAddress0.KeySet.PaymentAddress,
									},
									stakingTx: map[string]common.Hash{
										key0: *hash,
									},
								},
								swapRule:        swapRuleSingleInstructionOut,
								shardCommonPool: []incognitokey.CommitteePublicKey{},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: unCommitteedMu,
								},
							},
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey, *incKey2, *incKey3, *incKey4,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey0,
										},
									},
									autoStake: map[string]bool{
										key0: true,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{
										key0: paymentAddress0.KeySet.PaymentAddress,
									},
									stakingTx: map[string]common.Hash{
										key0: *hash,
									},
									mu: unCommitteedMu,
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{},
								swapRule:        swapRuleSingleInstructionOut,
							},
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key4,
							key0,
							"0",
							"1",
						},
					},
					MaxShardCommitteeSize: 4,
					ActiveShards:          1,
					ConsensusStateDB:      sDB,
					RandomNumber:          5000,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessSwapShardInstruction,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Unstake Instruction",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHash:   *hash,
						beaconHeight: 10,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey, *incKey2, *incKey3, *incKey4,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey5,
										},
									},
									autoStake:      map[string]bool{},
									rewardReceiver: map[string]privacy.PaymentAddress{},
									stakingTx:      map[string]common.Hash{},
									mu:             finalMu,
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{
									*incKey0,
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey, *incKey2, *incKey3, *incKey4,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey5,
										},
									},
									autoStake:      map[string]bool{},
									rewardReceiver: map[string]privacy.PaymentAddress{},
									stakingTx:      map[string]common.Hash{},
									mu:             unCommitteedMu,
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{
									*incKey0,
								},
							},
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey, *incKey2, *incKey3, *incKey4,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey5,
										},
									},
									autoStake:      map[string]bool{},
									rewardReceiver: map[string]privacy.PaymentAddress{},
									stakingTx:      map[string]common.Hash{},
									mu:             unCommitteedMu,
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{},
							},
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					newUnassignedCommonPool: []string{key0},
					ConsensusStateDB:        sDB,
					MaxShardCommitteeSize:   4,
					ActiveShards:            1,
				},
			},
			want:  &BeaconCommitteeStateHash{},
			want1: committeeChangeProcessUnstakeInstruction,
			want2: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key0,
					hash.String(),
					"100",
				},
			},
			wantErr: false,
		},
		{
			name: "Swap in 2 keys",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHeight: 5,
						beaconHash:   *hash,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey2, *incKey3, *incKey4, *incKey5, *incKey6, *incKey7,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey0, *incKey,
										},
									},
									mu: finalMu,
									autoStake: map[string]bool{
										key0: true,
										key:  true,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{
										incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
										incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
									},
									stakingTx: map[string]common.Hash{
										key0: *hash,
										key:  *tempHash,
									},
								},
								swapRule:        swapRuleIn2Keys,
								shardCommonPool: []incognitokey.CommitteePublicKey{},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: unCommitteedMu,
								},
							},
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey2, *incKey3, *incKey4, *incKey5, *incKey6, *incKey7, *incKey0, *incKey,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{},
									},
									autoStake: map[string]bool{
										key0: true,
										key:  true,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{
										incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
										incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
									},
									stakingTx: map[string]common.Hash{
										key0: *hash,
										key:  *tempHash,
									},
									mu: unCommitteedMu,
								},
								swapRule:        swapRuleIn2Keys,
								shardCommonPool: []incognitokey.CommitteePublicKey{},
							},
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							strings.Join([]string{key0, key}, ","),
							"",
							"0",
							"0",
						},
					},
					MaxShardCommitteeSize:            8,
					ActiveShards:                     1,
					ConsensusStateDB:                 sDB,
					NumberOfFixedShardBlockValidator: 6,
					MinShardCommitteeSize:            6,
					RandomNumber:                     5000,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessSwapShardInstruction2KeysIn,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap out 2 keys",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHeight: 5,
						beaconHash:   *hash,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey0, *incKey, *incKey4, *incKey5, *incKey2, *incKey3, *incKey6,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{},
									},
									mu: finalMu,
									autoStake: map[string]bool{
										key:  true,
										key4: true,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{
										incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
										incKey4.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
									},
									stakingTx: map[string]common.Hash{
										key:  *tempHash,
										key4: *tempHash,
									},
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{},
								swapRule:        swapRuleOut2Keys,
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: unCommitteedMu,
								},
							},
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey0, *incKey5, *incKey2, *incKey3, *incKey6,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey, *incKey4,
										},
									},
									autoStake: map[string]bool{
										key:  true,
										key4: true,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{
										incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
										incKey4.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
									},
									stakingTx: map[string]common.Hash{
										key:  *tempHash,
										key4: *tempHash,
									},
									mu: unCommitteedMu,
								},
								swapRule:        swapRuleOut2Keys,
								shardCommonPool: []incognitokey.CommitteePublicKey{},
							},
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							"",
							strings.Join([]string{key, key4}, ","),
							"0",
							"0",
						},
					},
					MaxShardCommitteeSize:            7,
					ActiveShards:                     1,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					NumberOfFixedShardBlockValidator: 1,
					MinShardCommitteeSize:            1,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessSwapShardInstruction2KeysOut,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap in and out 2 keys",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHeight: 5,
						beaconHash:   *hash,
						finalState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey0, *incKey, *incKey4, *incKey5, *incKey6, *incKey7,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey2, *incKey3,
										},
									},
									mu: finalMu,
									autoStake: map[string]bool{
										key0: true,
										key:  true,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{
										incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
										incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
									},
									stakingTx: map[string]common.Hash{
										key0: *hash,
										key:  *tempHash,
									},
								},
								swapRule:        swapRuleInAndOut2Keys,
								shardCommonPool: []incognitokey.CommitteePublicKey{},
							},
						},
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: unCommitteedMu,
								},
							},
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						uncommittedState: &BeaconCommitteeStateV2{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey4, *incKey5, *incKey6, *incKey7, *incKey2, *incKey3,
										},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{
											*incKey0, *incKey,
										},
									},
									autoStake: map[string]bool{
										key0: true,
										key:  true,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{
										incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
										incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
									},
									stakingTx: map[string]common.Hash{
										key0: *hash,
										key:  *tempHash,
									},
									mu: unCommitteedMu,
								},
								swapRule:        swapRuleInAndOut2Keys,
								shardCommonPool: []incognitokey.CommitteePublicKey{},
							},
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							strings.Join([]string{key2, key3}, ","),
							strings.Join([]string{key0, key}, ","),
							"0",
							"0",
						},
					},
					MaxShardCommitteeSize: 6,
					ActiveShards:          1,
					ConsensusStateDB:      sDB,
					RandomNumber:          5000,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessSwapShardInstruction2Keys,
			want2:   [][]string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &BeaconCommitteeEngineV2{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHeight:     tt.fields.beaconCommitteeEngineSlashingBase.beaconHeight,
						beaconHash:       tt.fields.beaconCommitteeEngineSlashingBase.beaconHash,
						finalState:       tt.fields.beaconCommitteeEngineSlashingBase.finalState,
						uncommittedState: tt.fields.beaconCommitteeEngineSlashingBase.uncommittedState,
					},
				},
			}
			_, got1, got2, err := engine.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BeaconCommitteeEngineV2.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Fatalf("BeaconCommitteeEngineV2.UpdateCommitteeState() got1 = %v, want1 = %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Fatalf("BeaconCommitteeEngineV2.UpdateCommitteeState() got2 = %v, want2 = %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(tt.fields.beaconCommitteeEngineSlashingBase.uncommittedState,
				tt.fieldsAfterProcess.beaconCommitteeEngineSlashingBase.uncommittedState) {
				t.Fatalf(`BeaconCommitteeEngineV2.UpdateCommitteeState() tt.fields.uncommittedState = %v, 
					tt.fieldsAfterProcess.uncommittedState = %v`,
					tt.fields.beaconCommitteeEngineSlashingBase.uncommittedState, tt.fieldsAfterProcess.beaconCommitteeEngineSlashingBase.uncommittedState)
			}
		})
	}
}

func TestBeaconCommitteeEngineV2_UpdateCommitteeState_MultipleInstructions(t *testing.T) {
	hash, _ := common.Hash{}.NewHashFromStr("123")
	tempHash, _ := common.Hash{}.NewHashFromStr("456")
	initTestParams()
	initLog()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	paymentAddress0, err := wallet.Base58CheckDeserialize(paymentAddreessKey0)
	assert.Nil(t, err)
	rewardReceiverKey := incKey.GetIncKeyBase58()
	rewardReceiverkey0 := incKey0.GetIncKeyBase58()
	rewardReceiverkey4 := incKey4.GetIncKeyBase58()
	rewardReceiverkey10 := incKey10.GetIncKeyBase58()
	rewardReceiverkey7 := incKey7.GetIncKeyBase58()
	paymentAddress, err := wallet.Base58CheckDeserialize(paymentAddreessKey0)
	assert.Nil(t, err)

	committeeChangeStakeAndAssginResult := NewCommitteeChange()
	committeeChangeStakeAndAssginResult.NextEpochShardCandidateAdded = []incognitokey.CommitteePublicKey{
		*incKey0,
	}

	committeeChangeStakeAndAssginResult.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey5,
	}

	committeeChangeStakeAndAssginResult.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}

	committeeChangeUnstakeAssign := NewCommitteeChange()
	committeeChangeUnstakeAssign.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeUnstakeAssign.StopAutoStake = []string{
		key5,
	}
	committeeChangeUnstakeAssign.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey5,
	}

	committeeChangeUnstakeAssign2 := NewCommitteeChange()
	committeeChangeUnstakeAssign2.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey6,
	}
	committeeChangeUnstakeAssign2.RemovedStaker = []string{
		key0,
	}
	committeeChangeUnstakeAssign2.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey0,
		*incKey6,
	}
	committeeChangeUnstakeAssign2.TermsRemoved = []string{}

	committeeChangeUnstakeAssign3 := NewCommitteeChange()
	committeeChangeUnstakeAssign3.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey6,
	}
	committeeChangeUnstakeAssign3.RemovedStaker = []string{
		key0,
	}
	committeeChangeUnstakeAssign3.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey6,
		*incKey0,
	}
	committeeChangeUnstakeAssign3.TermsRemoved = []string{}

	committeeChangeUnstakeSwap := NewCommitteeChange()
	committeeChangeUnstakeSwap.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeUnstakeSwap.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey0,
	}
	committeeChangeUnstakeSwap.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}

	committeeChangeUnstakeSwap.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey0,
	}
	committeeChangeUnstakeSwap.StopAutoStake = []string{
		key0,
	}

	committeeChangeUnstakeSwapOut := NewCommitteeChange()
	committeeChangeUnstakeSwapOut.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey0,
	}
	committeeChangeUnstakeSwapOut.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeUnstakeSwapOut.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey0,
	}
	committeeChangeUnstakeSwapOut.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}

	committeeChangeUnstakeSwapOut.StopAutoStake = []string{
		key0,
	}

	committeeChangeUnstakeAndRandomTime := NewCommitteeChange()
	committeeChangeUnstakeAndRandomTime.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey0,
	}
	committeeChangeUnstakeAndRandomTime.TermsRemoved = []string{}

	committeeChangeUnstakeAndRandomTime.RemovedStaker = []string{key0}
	committeeChangeUnstakeAndRandomTime2 := NewCommitteeChange()
	committeeChangeUnstakeAndRandomTime2.StopAutoStake = []string{key0}

	committeeChangeStopAutoStakeAndRandomTime := NewCommitteeChange()
	committeeChangeStopAutoStakeAndRandomTime.StopAutoStake = []string{key0}

	committeeChangeTwoSwapOut := NewCommitteeChange()
	committeeChangeTwoSwapOut.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeTwoSwapOut.ShardCommitteeRemoved[1] = []incognitokey.CommitteePublicKey{
		*incKey7,
	}
	committeeChangeTwoSwapOut.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeTwoSwapOut.ShardCommitteeAdded[1] = []incognitokey.CommitteePublicKey{
		*incKey11,
	}
	committeeChangeTwoSwapOut.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey7,
	}
	committeeChangeTwoSwapOut.ShardSubstituteAdded[1] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeTwoSwapOut.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeTwoSwapOut.ShardSubstituteRemoved[1] = []incognitokey.CommitteePublicKey{
		*incKey11,
	}

	committeeChangeTwoSlashing := NewCommitteeChange()
	committeeChangeTwoSlashing.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey4,
	}
	committeeChangeTwoSlashing.ShardCommitteeRemoved[1] = []incognitokey.CommitteePublicKey{
		*incKey10,
	}
	committeeChangeTwoSlashing.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeTwoSlashing.ShardCommitteeAdded[1] = []incognitokey.CommitteePublicKey{
		*incKey11,
	}
	committeeChangeTwoSlashing.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeTwoSlashing.ShardSubstituteRemoved[1] = []incognitokey.CommitteePublicKey{
		*incKey11,
	}
	committeeChangeTwoSlashing.RemovedStaker = []string{key4, key10}
	committeeChangeTwoSlashing.SlashingCommittee[0] = []string{key4}
	committeeChangeTwoSlashing.SlashingCommittee[1] = []string{key10}
	committeeChangeTwoSlashing.TermsRemoved = []string{}
	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{*incKey, *incKey0, *incKey4, *incKey10, *incKey7},
		map[string]privacy.PaymentAddress{
			rewardReceiverKey:   paymentAddress.KeySet.PaymentAddress,
			rewardReceiverkey0:  paymentAddress0.KeySet.PaymentAddress,
			rewardReceiverkey4:  paymentAddress0.KeySet.PaymentAddress,
			rewardReceiverkey10: paymentAddress0.KeySet.PaymentAddress,
			rewardReceiverkey7:  paymentAddress0.KeySet.PaymentAddress,
		},
		map[string]bool{
			key:   true,
			key0:  true,
			key4:  true,
			key10: true,
			key7:  true,
		},
		map[string]common.Hash{
			key:   *tempHash,
			key0:  *hash,
			key4:  *hash,
			key10: *hash,
			key7:  *hash,
		},
	)

	finalMu := &sync.RWMutex{}
	unCommitteedMu := &sync.RWMutex{}

	//Declare swaprule
	swapRule1 := &mocks.SwapRule{}
	swapRule1.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey,
			},
			OutPublicKeys: []string{key},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey0,
			},
			InPublicKeys: []string{key0},
			ChainID:      0,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key2, key3, key4, key0}, []string{}, []string{}, []string{key})
	swapRule1.On("Version").Return(swapRuleTestVersion)
	swapRule1.On("AssignOffset", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(1)

	swapRule2 := &mocks.SwapRule{}
	swapRule2.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey0,
			},
			OutPublicKeys: []string{key0},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey,
			},
			InPublicKeys: []string{key},
			ChainID:      0,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key2, key3, key4, key}, []string{}, []string{}, []string{key0})
	swapRule2.On("Version").Return(swapRuleTestVersion)
	swapRule2.On("AssignOffset", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(1)

	swapRule3 := &mocks.SwapRule{}
	swapRule3.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey,
			},
			OutPublicKeys: []string{key},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey5,
			},
			InPublicKeys: []string{key5},
			ChainID:      0,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key2, key3, key4, key5}, []string{}, []string{}, []string{key})

	swapRule3.On("GenInstructions", uint8(1), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey7,
			},
			OutPublicKeys: []string{key7},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey11,
			},
			InPublicKeys: []string{key11},
			ChainID:      1,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key8, key9, key10, key11}, []string{}, []string{}, []string{key7})

	swapRule3.On("Version").Return(swapRuleTestVersion)
	swapRule3.On("AssignOffset", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(1)

	swapRule4 := &mocks.SwapRule{}
	swapRule4.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey4,
			},
			OutPublicKeys: []string{key4},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey5,
			},
			InPublicKeys: []string{key5},
			ChainID:      0,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key, key2, key3, key5}, []string{}, []string{key4}, []string{})

	swapRule4.On("GenInstructions", uint8(1), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey10,
			},
			OutPublicKeys: []string{key10},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey11,
			},
			InPublicKeys: []string{key11},
			ChainID:      1,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key7, key8, key9, key11}, []string{}, []string{key10}, []string{})

	swapRule4.On("Version").Return(swapRuleTestVersion)
	swapRule4.On("AssignOffset", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(1)

	type fields struct {
		beaconHeight                      uint64
		beaconHash                        common.Hash
		finalBeaconCommitteeStateV2       *BeaconCommitteeStateV2
		uncommittedBeaconCommitteeStateV2 *BeaconCommitteeStateV2
		version                           uint
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *BeaconCommitteeStateHash
		want1              *CommitteeChange
		want2              [][]string
		wantErr            bool
	}{
		{
			name: "Stake Then Assign",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
							mu:             finalMu,
						},
						numberOfAssignedCandidates: 1,
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey5,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconHash:                  *hash,
				version:                     SLASHING_VERSION,
				beaconHeight:                10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey5,
								},
							},
							autoStake: map[string]bool{
								key0: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
							mu: unCommitteedMu,
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey0,
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.STAKE_ACTION,
							key0,
							instruction.SHARD_INST,
							hash.String(),
							paymentAddreessKey0,
							"false",
						},
						[]string{
							instruction.RANDOM_ACTION,
							"3157440766",
							"637918",
							"3157440766",
							"3157440766",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeStakeAndAssginResult,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Assign Then Stake",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
							mu:             finalMu,
						},
						numberOfAssignedCandidates: 1,
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey5,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconHash:                  *hash,
				version:                     SLASHING_VERSION,
				beaconHeight:                10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey5,
								},
							},
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
							mu: unCommitteedMu,
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey0,
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.RANDOM_ACTION,
							"3157440766",
							"637918",
							"3157440766",
							"3157440766",
						},
						[]string{
							instruction.STAKE_ACTION,
							key0,
							instruction.SHARD_INST,
							hash.String(),
							paymentAddreessKey0,
							"true",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeStakeAndAssginResult,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Unstake And Assign 1, Fail to Unstake because Key in Current Epoch Candidate, only turn off auto stake flag",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key5: true,
								key6: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								key5: paymentAddress0.KeySet.PaymentAddress,
								key6: paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key5: *hash,
								key6: *hash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey5,
							*incKey6,
						},
						numberOfAssignedCandidates: 1,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey5,
								},
							},
							mu: unCommitteedMu,
							autoStake: map[string]bool{
								key5: false,
								key6: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								key5: paymentAddress0.KeySet.PaymentAddress,
								key6: paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key5: *hash,
								key6: *hash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey6,
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key5,
						},
						[]string{
							instruction.RANDOM_ACTION,
							"3157440766",
							"637918",
							"3157440766",
							"3157440766",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeAssign,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Assign Then Unstake 2, Fail to Unstake because Key in Current Epoch Candidate, only turn off auto stake flag",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key5: true,
								key6: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								key5: paymentAddress0.KeySet.PaymentAddress,
								key6: paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key5: *hash,
								key6: *hash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey5,
							*incKey6,
						},
						numberOfAssignedCandidates: 1,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey5,
								},
							},
							mu: unCommitteedMu,
							autoStake: map[string]bool{
								key5: false,
								key6: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								key5: paymentAddress0.KeySet.PaymentAddress,
								key6: paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key5: *hash,
								key6: *hash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey6,
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.RANDOM_ACTION,
							"3157440766",
							"637918",
							"3157440766",
							"3157440766",
						},
						[]string{
							instruction.UNSTAKE_ACTION,
							key5,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeAssign,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Unstake And Assign 3, Success to Unstake because Key in Next Epoch Candidate",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey6,
							*incKey0,
						},
						numberOfAssignedCandidates: 1,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey6,
								},
							},
							mu:             unCommitteedMu,
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.RANDOM_ACTION,
							"3157440766",
							"637918",
							"3157440766",
							"3157440766",
						},
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:  &BeaconCommitteeStateHash{},
			want1: committeeChangeUnstakeAssign3,
			want2: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key0,
					hash.String(),
					"100",
				},
			},
			wantErr: false,
		},
		{
			name: "Unstake And Assign 4, Success to Unstake because Key in Next Epoch Candidate",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						numberOfAssignedCandidates: 1,
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey6,
							*incKey0,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey6,
								},
							},
							mu:             unCommitteedMu,
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
						[]string{
							instruction.RANDOM_ACTION,
							"3157440766",
							"637918",
							"3157440766",
							"3157440766",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:  &BeaconCommitteeStateHash{},
			want1: committeeChangeUnstakeAssign2,
			want2: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key0,
					hash.String(),
					"100",
				},
			},
			wantErr: false,
		},
		{
			name: "Unstake Then Swap 1, Failed to Unstake Swap Out key, Only turn off auto stake",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key:  false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key:  *tempHash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{},
						swapRule:        swapRule1,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey2, *incKey3, *incKey4, *incKey0,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey,
								},
							},
							mu: unCommitteedMu,
							autoStake: map[string]bool{
								key0: false,
								key:  false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key:  *tempHash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{},
						swapRule:        swapRule1,
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key0,
							key,
							"0",
							"0",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeSwap,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap Then Unstake 2, Failed to Unstake Swap Out key, Only turn off auto stake",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key:  *tempHash,
							},
						},
						swapRule:        swapRule1,
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey2, *incKey3, *incKey4, *incKey0,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey,
								},
							},
							mu: unCommitteedMu,
							autoStake: map[string]bool{
								key0: false,
								key:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key:  *tempHash,
							},
						},
						swapRule:        swapRule1,
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key0,
							key,
							"0",
							"0",
						},
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeSwap,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Is Beacon Random Time == False And Unstake",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey0,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							mu:             unCommitteedMu,
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:  &BeaconCommitteeStateHash{},
			want1: committeeChangeUnstakeAndRandomTime,
			want2: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key0,
					hash.String(),
					"100",
				},
			},
			wantErr: false,
		},
		{
			name: "Is Beacon Random Time == True And Unstake",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey0,
						},
						swapRule:                   swapRule1,
						numberOfAssignedCandidates: 0,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							mu: unCommitteedMu,
							autoStake: map[string]bool{
								key0: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey0,
						},
						swapRule:                   swapRule1,
						numberOfAssignedCandidates: 1,
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
					IsBeaconRandomTime:    true,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeAndRandomTime2,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Is Beacon Random Time And Stop Auto Stake",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey0,
						},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							mu: unCommitteedMu,
							autoStake: map[string]bool{
								key0: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey0,
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.STOP_AUTO_STAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeStopAutoStakeAndRandomTime,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap Out And Unstake",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key:  *tempHash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{},
						swapRule:        swapRule2,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey2, *incKey3, *incKey4, *incKey,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0,
								},
							},
							mu: unCommitteedMu,
							autoStake: map[string]bool{
								key0: false,
								key:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key:  *tempHash,
							},
						},
						swapRule:        swapRule2,
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key,
							key0,
							"0",
							"0",
						},
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeSwapOut,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Unstake And Swap Out",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key:  *tempHash,
							},
						},
						swapRule:        swapRule2,
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey2, *incKey3, *incKey4, *incKey,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0,
								},
							},
							mu: unCommitteedMu,
							autoStake: map[string]bool{
								key0: false,
								key:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key:  *tempHash,
							},
						},
						swapRule:        swapRule2,
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key,
							key0,
							"0",
							"0",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeSwapOut,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Stop Auto Stake And Swap Out",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key:  *tempHash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{},
						swapRule:        swapRule2,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey2, *incKey3, *incKey4, *incKey,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0,
								},
							},
							mu: unCommitteedMu,
							autoStake: map[string]bool{
								key0: false,
								key:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key:  *tempHash,
							},
						},
						swapRule:        swapRule2,
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key,
							key0,
							"0",
							"0",
						},
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeSwapOut,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap Out And Stop Auto Stake",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key:  *tempHash,
							},
						},
						swapRule:        swapRule2,
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey2, *incKey3, *incKey4, *incKey,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0,
								},
							},
							mu: unCommitteedMu,
							autoStake: map[string]bool{
								key0: false,
								key:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key:  *tempHash,
							},
						},
						swapRule:        swapRule2,
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.STOP_AUTO_STAKE_ACTION,
							key0,
						},
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key,
							key0,
							"0",
							"0",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeSwapOut,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Double Swap Instruction for 2 shard, Back to substitutes",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey7, *incKey8, *incKey9, *incKey10,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey5, *incKey13,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey11, *incKey12,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key7: true,
								key:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey7.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key7: *hash,
								key:  *tempHash,
							},
						},
						swapRule:        swapRule3,
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey2, *incKey3, *incKey4, *incKey5,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey8, *incKey9, *incKey10, *incKey11,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey13, *incKey7,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey12, *incKey,
								},
							},
							mu: unCommitteedMu,
							autoStake: map[string]bool{
								key0: true,
								key7: true,
								key:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey7.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key7: *hash,
								key:  *tempHash,
							},
						},
						swapRule:        swapRule3,
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key5,
							key,
							"0",
							"0",
						},
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key11,
							key7,
							"1",
							"0",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          2,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeTwoSwapOut,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Double Swap Instruction for 2 shard, Slashing",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey7, *incKey8, *incKey9, *incKey10,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey5, *incKey13,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey11, *incKey12,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0:  true,
								key4:  true,
								key10: true,
								key7:  true,
								key:   true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
								incKey7.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
								incKey4.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
								incKey10.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():   paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0:  *hash,
								key7:  *hash,
								key4:  *hash,
								key10: *hash,
								key:   *tempHash,
							},
						},
						swapRule:        swapRule4,
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu: unCommitteedMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey5,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey7, *incKey8, *incKey9, *incKey11,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey13,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey12,
								},
							},
							mu: unCommitteedMu,
							autoStake: map[string]bool{
								key0: true,
								key7: true,
								key:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey7.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key7: *hash,
								key:  *tempHash,
							},
						},
						swapRule:        swapRule4,
						shardCommonPool: []incognitokey.CommitteePublicKey{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key5,
							key4,
							"0",
							"0",
						},
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key11,
							key10,
							"1",
							"0",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          2,
					MaxShardCommitteeSize: 4,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key4:  samplePenalty,
						key10: samplePenalty,
					},
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeTwoSlashing,
			want2:   [][]string{instruction.NewReturnStakeInsWithValue([]string{key4, key10}, []string{hash.String(), hash.String()}).ToString()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &BeaconCommitteeEngineV2{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHeight:     tt.fields.beaconHeight,
						beaconHash:       tt.fields.beaconHash,
						finalState:       tt.fields.finalBeaconCommitteeStateV2,
						uncommittedState: tt.fields.uncommittedBeaconCommitteeStateV2,
					},
				},
			}
			_, got1, got2, err := engine.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV2.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeEngineV2.UpdateCommitteeState() got1 = %v, want1 = %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeEngineV2.UpdateCommitteeState() got2 = %v, want2 = %v", got2, tt.want2)
			}

			if !reflect.DeepEqual(tt.fields.uncommittedBeaconCommitteeStateV2,
				tt.fieldsAfterProcess.uncommittedBeaconCommitteeStateV2) {
				t.Errorf(`BeaconCommitteeEngineV2.UpdateCommitteeState() tt.fields.uncommittedBeaconCommitteeStateV2 = %v,
			tt.fieldsAfterProcess.uncommittedBeaconCommitteeStateV2 = %v`,
					tt.fields.uncommittedBeaconCommitteeStateV2, tt.fieldsAfterProcess.uncommittedBeaconCommitteeStateV2)
			}
		})
	}
}
