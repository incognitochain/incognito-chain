package pdex

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataMocks "github.com/incognitochain/incognito-chain/metadata/common/mocks"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/stretchr/testify/assert"
)

func Test_stateV2_BuildInstructions(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	firstTxHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)

	// first contribution tx
	firstContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		token0ID.String(), utils.EmptyString, 100, 20000,
	)
	assert.Nil(t, err)
	contributionTx := &metadataMocks.Transaction{}
	contributionTx.On("GetMetadata").Return(firstContributionMetadata)
	contributionTx.On("GetMetadataType").Return(metadataCommon.Pdexv3AddLiquidityRequestMeta)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)
	contributionTx.On("GetValidationEnv").Return(valEnv)
	contributionTx.On("Hash").Return(firstTxHash)
	waitingContributionStateDB := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0, validOTAReceiver1,
			*token0ID, *firstTxHash, nil, 100, 20000, 1,
		),
		"pair_hash")
	waitingContributionInst := instruction.NewWaitingAddLiquidityWithValue(*waitingContributionStateDB)
	waitingContributionInstBytes, err := json.Marshal(waitingContributionInst)
	//

	type fields struct {
		stateBase                   stateBase
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		params                      Params
		stakingPoolsState           map[string]*StakingPoolState
		orders                      map[int64][]Order
		producer                    stateProducerV2
		processor                   stateProcessorV2
	}
	type args struct {
		env StateEnvironment
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               [][]string
		wantErr            bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
			},
			fieldsAfterProcess: fields{
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						"", validOTAReceiver0, validOTAReceiver1,
						*token0ID, *firstTxHash, nil, 100, 20000, 1,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
			},
			args: args{
				env: &stateEnvironment{
					beaconHeight: 10,
					listTxs: map[byte][]metadataCommon.Transaction{
						1: []metadataCommon.Transaction{
							contributionTx,
						},
					},
				},
			},
			want: [][]string{
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionWaitingChainStatus,
					string(waitingContributionInstBytes),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stateV2{
				stateBase:                   tt.fields.stateBase,
				waitingContributions:        tt.fields.waitingContributions,
				deletedWaitingContributions: tt.fields.deletedWaitingContributions,
				poolPairs:                   tt.fields.poolPairs,
				params:                      tt.fields.params,
				stakingPoolsState:           tt.fields.stakingPoolsState,
				producer:                    tt.fields.producer,
				processor:                   tt.fields.processor,
			}
			got, err := s.BuildInstructions(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateV2.BuildInstructions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateV2.BuildInstructions() = %v, want %v", got, tt.want)
				return
			}
			if !reflect.DeepEqual(s.waitingContributions, tt.fieldsAfterProcess.waitingContributions) {
				t.Errorf("fieldsAfterProcess = %v, want %v", s, tt.fieldsAfterProcess)
				return
			}
		})
	}
}

func Test_stateV2_Process(t *testing.T) {
	initDB()
	initLog()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	firstTxHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)

	// first contribution tx
	firstContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		token0ID.String(), utils.EmptyString, 100, 20000,
	)
	assert.Nil(t, err)
	contributionTx := &metadataMocks.Transaction{}
	contributionTx.On("GetMetadata").Return(firstContributionMetadata)
	contributionTx.On("GetMetadataType").Return(metadataCommon.Pdexv3AddLiquidityRequestMeta)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)
	contributionTx.On("GetValidationEnv").Return(valEnv)
	contributionTx.On("Hash").Return(firstTxHash)
	waitingContributionStateDB := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0, validOTAReceiver1,
			*token0ID, *firstTxHash, nil, 100, 20000, 1,
		),
		"pair_hash")
	waitingContributionInst := instruction.NewWaitingAddLiquidityWithValue(*waitingContributionStateDB)
	waitingContributionInstBytes, err := json.Marshal(waitingContributionInst)

	type fields struct {
		stateBase                   stateBase
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		params                      Params
		stakingPoolsState           map[string]*StakingPoolState
		orders                      map[int64][]Order
		producer                    stateProducerV2
		processor                   stateProcessorV2
	}
	type args struct {
		env StateEnvironment
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		wantErr            bool
	}{
		{
			name: "Add Liquidity",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
			},
			args: args{
				env: &stateEnvironment{
					beaconHeight: 10,
					stateDB:      sDB,
					beaconInstructions: [][]string{
						[]string{
							strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
							common.PDEContributionWaitingChainStatus,
							string(waitingContributionInstBytes),
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				stateBase: stateBase{},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						"", validOTAReceiver0, validOTAReceiver1,
						*token0ID, *firstTxHash, nil, 100, 20000, 1,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stateV2{
				stateBase:                   tt.fields.stateBase,
				waitingContributions:        tt.fields.waitingContributions,
				deletedWaitingContributions: tt.fields.deletedWaitingContributions,
				poolPairs:                   tt.fields.poolPairs,
				params:                      tt.fields.params,
				stakingPoolsState:           tt.fields.stakingPoolsState,
				producer:                    tt.fields.producer,
				processor:                   tt.fields.processor,
			}
			if err := s.Process(tt.args.env); (err != nil) != tt.wantErr {
				t.Errorf("stateV2.Process() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s.waitingContributions, tt.fieldsAfterProcess.waitingContributions) {
				t.Errorf("fieldsAfterProcess = %v, want %v", *s, tt.fieldsAfterProcess)
				return
			}
		})
	}
}

func Test_stateV2_Clone(t *testing.T) {

	type fields struct {
		stateBase                   stateBase
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		params                      Params
		stakingPoolsState           map[string]*StakingPoolState
		orders                      map[int64][]Order
		producer                    stateProducerV2
		processor                   stateProcessorV2
	}
	tests := []struct {
		name   string
		fields fields
		want   State
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stateV2{
				stateBase:                   tt.fields.stateBase,
				waitingContributions:        tt.fields.waitingContributions,
				deletedWaitingContributions: tt.fields.deletedWaitingContributions,
				poolPairs:                   tt.fields.poolPairs,
				params:                      tt.fields.params,
				stakingPoolsState:           tt.fields.stakingPoolsState,
				producer:                    tt.fields.producer,
				processor:                   tt.fields.processor,
			}
			if got := s.Clone(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateV2.Clone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stateV2_StoreToDB(t *testing.T) {
	type fields struct {
		stateBase                   stateBase
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		params                      Params
		stakingPoolsState           map[string]*StakingPoolState
		orders                      map[int64][]Order
		producer                    stateProducerV2
		processor                   stateProcessorV2
	}
	type args struct {
		env         StateEnvironment
		stateChange *StateChange
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stateV2{
				stateBase:                   tt.fields.stateBase,
				waitingContributions:        tt.fields.waitingContributions,
				deletedWaitingContributions: tt.fields.deletedWaitingContributions,
				poolPairs:                   tt.fields.poolPairs,
				params:                      tt.fields.params,
				stakingPoolsState:           tt.fields.stakingPoolsState,
				producer:                    tt.fields.producer,
				processor:                   tt.fields.processor,
			}
			if err := s.StoreToDB(tt.args.env, tt.args.stateChange); (err != nil) != tt.wantErr {
				t.Errorf("stateV2.StoreToDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_stateV2_GetDiff(t *testing.T) {
	type fields struct {
		stateBase                   stateBase
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		params                      Params
		stakingPoolsState           map[string]*StakingPoolState
		orders                      map[int64][]Order
		producer                    stateProducerV2
		processor                   stateProcessorV2
	}
	type args struct {
		compareState State
		stateChange  *StateChange
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    State
		want1   *StateChange
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stateV2{
				stateBase:                   tt.fields.stateBase,
				waitingContributions:        tt.fields.waitingContributions,
				deletedWaitingContributions: tt.fields.deletedWaitingContributions,
				poolPairs:                   tt.fields.poolPairs,
				params:                      tt.fields.params,
				stakingPoolsState:           tt.fields.stakingPoolsState,
				producer:                    tt.fields.producer,
				processor:                   tt.fields.processor,
			}
			got, got1, err := s.GetDiff(tt.args.compareState, tt.args.stateChange)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateV2.GetDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateV2.GetDiff() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateV2.GetDiff() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
