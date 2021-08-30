package pdex

import (
	"encoding/json"
	"math/big"
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
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)
	firstTxHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)
	nftHash, err := common.Hash{}.NewHashFromStr(nftID)
	assert.Nil(t, err)
	nftHash1, err := common.Hash{}.NewHashFromStr(nftID1)
	assert.Nil(t, err)
	txReqID, err := common.Hash{}.NewHashFromStr("1111122222")
	assert.Nil(t, err)

	// first contribution tx
	firstContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0,
		token0ID.String(), nftID, 100, 20000,
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
			"", validOTAReceiver0,
			*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
		),
		"pair_hash")
	waitingContributionInst := instruction.NewWaitingAddLiquidityWithValue(*waitingContributionStateDB)
	waitingContributionInstBytes, err := json.Marshal(waitingContributionInst)
	//

	// user mint nft
	acceptInst, err := instruction.NewAcceptUserMintNftWithValue(
		validOTAReceiver0, 100, 1, *nftHash1, *txReqID,
	).StringSlice()
	assert.Nil(t, err)

	metaData := metadataPdexv3.NewUserMintNftRequestWithValue(validOTAReceiver0, 100)
	userMintNftTx := &metadataMocks.Transaction{}
	userMintNftTx.On("GetMetadata").Return(metaData)
	userMintNftTx.On("GetMetadataType").Return(metadataCommon.Pdexv3UserMintNftRequestMeta)
	valEnv = tx_generic.WithShardID(valEnv, 1)
	userMintNftTx.On("GetValidationEnv").Return(valEnv)
	userMintNftTx.On("Hash").Return(txReqID)
	//

	// staking
	stakingMetadata := metadataPdexv3.NewStakingRequestWithValue(
		common.PRVIDStr, nftID1, validOTAReceiver0, 100,
	)
	stakingTx := &metadataMocks.Transaction{}
	stakingTx.On("GetMetadata").Return(stakingMetadata)
	stakingTx.On("GetMetadataType").Return(metadataCommon.Pdexv3StakingRequestMeta)
	valEnv2 := tx_generic.DefaultValEnv()
	valEnv2 = tx_generic.WithShardID(valEnv2, 1)
	stakingTx.On("GetValidationEnv").Return(valEnv2)
	stakingTx.On("Hash").Return(txReqID)
	stakingInst, err := instruction.NewAcceptStakingWtihValue(
		*nftHash1, common.PRVCoinID, *txReqID, 1, 100,
	).StringSlice()
	assert.Nil(t, err)
	//

	withdrawLiquidityOtaReceivers := map[string]string{
		token0ID.String(): validOTAReceiver1,
		token1ID.String(): validOTAReceiver1,
		nftID1:            validOTAReceiver0,
	}

	// withdraw tx
	withdrawMetadata := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
		poolPairID, nftID1, withdrawLiquidityOtaReceivers, 100,
	)
	withdrawTx := &metadataMocks.Transaction{}
	withdrawTx.On("GetMetadata").Return(withdrawMetadata)
	withdrawTx.On("GetMetadataType").Return(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta)
	valEnv3 := tx_generic.DefaultValEnv()
	valEnv3 = tx_generic.WithShardID(valEnv3, 1)
	withdrawTx.On("GetValidationEnv").Return(valEnv3)
	withdrawTx.On("Hash").Return(txReqID)
	withdrawLiquidityMintNftInst, err := instruction.NewMintNftWithValue(*nftHash1, validOTAReceiver0, 1, *txReqID).
		StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta))
	assert.Nil(t, err)

	//

	//accept instructions
	acceptWithdrawLiquidityInst0, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *nftHash1, *token0ID, 50, 100, validOTAReceiver1, *txReqID, 1,
	).StringSlice()
	assert.Nil(t, err)
	acceptWithdrawLiquidityInst1, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *nftHash1, *token1ID, 200, 100, validOTAReceiver1, *txReqID, 1,
	).StringSlice()
	assert.Nil(t, err)
	//

	type fields struct {
		stateBase                   stateBase
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		params                      Params
		stakingPoolStates           map[string]*StakingPoolState
		nftIDs                      map[string]uint64
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
			name: "WaitingContributions",
			fields: fields{
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			fieldsAfterProcess: fields{
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						"", validOTAReceiver0,
						*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
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
		{
			name: "User mint nft",
			fields: fields{
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				nftIDs:                      map[string]uint64{},
				params: Params{
					MintNftRequireAmount: 100,
				},
			},
			fieldsAfterProcess: fields{
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
			},
			args: args{
				env: &stateEnvironment{
					beaconHeight: 10,
					listTxs: map[byte][]metadataCommon.Transaction{
						1: []metadataCommon.Transaction{userMintNftTx},
					},
				},
			},
			want:    [][]string{acceptInst},
			wantErr: false,
		},
		{
			name: "Accept staking tx",
			fields: fields{
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				params: Params{},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						stakers: map[string]*Staker{},
					},
				},
			},
			fieldsAfterProcess: fields{
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 100,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity:               100,
								lastUpdatedBeaconHeight: 10,
								rewards:                 map[string]uint64{},
							},
						},
					},
				},
			},
			args: args{
				env: &stateEnvironment{
					beaconHeight: 10,
					listTxs: map[byte][]metadataCommon.Transaction{
						1: []metadataCommon.Transaction{stakingTx},
					},
				},
			},
			want:    [][]string{stakingInst},
			wantErr: false,
		},
		{
			name: "Accept withdraw liquidity tx",
			fields: fields{
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						shares: map[string]*Share{
							nftID1: &Share{
								amount:                  300,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				params: Params{},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						stakers: map[string]*Staker{},
					},
				},
			},
			fieldsAfterProcess: fields{
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 200, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						shares: map[string]*Share{
							nftID1: &Share{
								amount:                  200,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 20,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						stakers: map[string]*Staker{},
					},
				},
			},
			args: args{
				env: &stateEnvironment{
					beaconHeight: 20,
					listTxs: map[byte][]metadataCommon.Transaction{
						1: []metadataCommon.Transaction{withdrawTx},
					},
				},
			},
			want:    [][]string{acceptWithdrawLiquidityInst0, acceptWithdrawLiquidityInst1, withdrawLiquidityMintNftInst},
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
				stakingPoolStates:           tt.fields.stakingPoolStates,
				nftIDs:                      tt.fields.nftIDs,
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
				t.Errorf("waitingContributions = %v, want %v", s.waitingContributions, tt.fieldsAfterProcess.waitingContributions)
				return
			}
			if !reflect.DeepEqual(s.poolPairs, tt.fieldsAfterProcess.poolPairs) {
				t.Errorf("waitingContributions = %v, want %v", s.waitingContributions, tt.fieldsAfterProcess.waitingContributions)
				return
			}
			if !reflect.DeepEqual(s.nftIDs, tt.fieldsAfterProcess.nftIDs) {
				t.Errorf("nftIDs = %v, want %v", s.nftIDs, tt.fieldsAfterProcess.nftIDs)
				return
			}
			if !reflect.DeepEqual(s.stakingPoolStates, tt.fieldsAfterProcess.stakingPoolStates) {
				t.Errorf("stakingPoolStates = %v, want %v", s.stakingPoolStates, tt.fieldsAfterProcess.stakingPoolStates)
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
	nftHash1, err := common.Hash{}.NewHashFromStr(nftID1)
	assert.Nil(t, err)
	txReqID, err := common.Hash{}.NewHashFromStr("1111122222")
	assert.Nil(t, err)

	// first contribution tx
	firstContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0,
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
			"", validOTAReceiver0,
			*token0ID, *firstTxHash, common.Hash{}, 100, 20000, 1,
		),
		"pair_hash")
	waitingContributionInst := instruction.NewWaitingAddLiquidityWithValue(*waitingContributionStateDB)
	waitingContributionInstBytes, err := json.Marshal(waitingContributionInst)
	//

	// user mint nft
	acceptInst, err := instruction.NewAcceptUserMintNftWithValue(
		validOTAReceiver0, 100, 1, *nftHash1, *txReqID,
	).StringSlice()
	assert.Nil(t, err)

	metaData := metadataPdexv3.NewUserMintNftRequestWithValue(validOTAReceiver0, 100)
	userMintNftTx := &metadataMocks.Transaction{}
	userMintNftTx.On("GetMetadata").Return(metaData)
	userMintNftTx.On("GetMetadataType").Return(metadataCommon.Pdexv3UserMintNftRequestMeta)
	valEnv = tx_generic.WithShardID(valEnv, 1)
	userMintNftTx.On("GetValidationEnv").Return(valEnv)
	userMintNftTx.On("Hash").Return(txReqID)
	//

	// staking
	stakingMetadata := metadataPdexv3.NewStakingRequestWithValue(
		common.PRVIDStr, nftID1, validOTAReceiver0, 100,
	)
	stakingTx := &metadataMocks.Transaction{}
	stakingTx.On("GetMetadata").Return(stakingMetadata)
	stakingTx.On("GetMetadataType").Return(metadataCommon.Pdexv3StakingRequestMeta)
	valEnv2 := tx_generic.DefaultValEnv()
	valEnv2 = tx_generic.WithShardID(valEnv2, 1)
	stakingTx.On("GetValidationEnv").Return(valEnv2)
	stakingTx.On("Hash").Return(txReqID)
	stakingInst, err := instruction.NewAcceptStakingWtihValue(
		*nftHash1, common.PRVCoinID, *txReqID, 1, 100,
	).StringSlice()
	assert.Nil(t, err)
	//

	type fields struct {
		stateBase                   stateBase
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		params                      Params
		stakingPoolStates           map[string]*StakingPoolState
		orders                      map[int64][]Order
		nftIDs                      map[string]uint64
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
						"", validOTAReceiver0,
						*token0ID, *firstTxHash, common.Hash{}, 100, 20000, 1,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
			},
			wantErr: false,
		},
		{
			name: "User mint Nft",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				nftIDs:                      map[string]uint64{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
			},
			args: args{
				env: &stateEnvironment{
					beaconHeight:       10,
					stateDB:            sDB,
					beaconInstructions: [][]string{acceptInst},
				},
			},
			fieldsAfterProcess: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			wantErr: false,
		},
		{
			name: "User mint Nft",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				nftIDs:                      map[string]uint64{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
			},
			args: args{
				env: &stateEnvironment{
					beaconHeight:       10,
					stateDB:            sDB,
					beaconInstructions: [][]string{acceptInst},
				},
			},
			fieldsAfterProcess: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			wantErr: false,
		},
		{
			name: "Accept staking tx",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						stakers: map[string]*Staker{},
					},
				},
			},
			args: args{
				env: &stateEnvironment{
					beaconHeight:       10,
					stateDB:            sDB,
					beaconInstructions: [][]string{stakingInst},
				},
			},
			fieldsAfterProcess: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 100,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity:               100,
								lastUpdatedBeaconHeight: 10,
								rewards:                 map[string]uint64{},
							},
						},
					},
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
				stakingPoolStates:           tt.fields.stakingPoolStates,
				nftIDs:                      tt.fields.nftIDs,
				producer:                    tt.fields.producer,
				processor:                   tt.fields.processor,
			}
			if err := s.Process(tt.args.env); (err != nil) != tt.wantErr {
				t.Errorf("stateV2.Process() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s.waitingContributions, tt.fieldsAfterProcess.waitingContributions) {
				t.Errorf("waitingContributions = %v, want %v", s.waitingContributions, tt.fieldsAfterProcess.waitingContributions)
				return
			}
			if !reflect.DeepEqual(s.nftIDs, tt.fieldsAfterProcess.nftIDs) {
				t.Errorf("nftIDs = %v, want %v", s.nftIDs, tt.fieldsAfterProcess.nftIDs)
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
		stakingPoolStates           map[string]*StakingPoolState
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
				stakingPoolStates:           tt.fields.stakingPoolStates,
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
	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	type fields struct {
		stateBase                   stateBase
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		params                      Params
		stakingPoolStates           map[string]*StakingPoolState
		nftIDs                      map[string]uint64
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
		{
			name: "Store pool pair + trading fees",
			fields: fields{
				stateBase: stateBase{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 200, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount: 200,
								tradingFees: map[string]uint64{
									common.PRVIDStr: 10,
								},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
			},
			args: args{
				env: &stateEnvironment{
					stateDB: sDB,
				},
				stateChange: &StateChange{
					poolPairIDs: map[string]bool{
						poolPairID: true,
					},
					shares: map[string]*ShareChange{
						nftID: &ShareChange{
							isChanged: true,
							tokenIDs: map[string]bool{
								common.PRVIDStr: true,
							},
						},
					},
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
				stakingPoolStates:           tt.fields.stakingPoolStates,
				nftIDs:                      tt.fields.nftIDs,
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
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	type fields struct {
		stateBase                   stateBase
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		params                      Params
		stakingPoolStates           map[string]*StakingPoolState
		nftIDs                      map[string]uint64
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
		{
			name: "full pool pair",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 200, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						shares: map[string]*Share{
							nftID1: &Share{
								amount: 200,
								tradingFees: map[string]uint64{
									common.PRVIDStr: 100,
								},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				params: Params{
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PRVDiscountPercent:              25,
					LimitProtocolFeePercent:         0,
					LimitStakingPoolRewardPercent:   10,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
					DefaultStakingPoolsShare:        0,
					StakingPoolsShare: map[string]uint{
						common.PRVIDStr: 10,
					},
					MintNftRequireAmount: 1000000000,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{},
				},
				nftIDs:    map[string]uint64{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			args: args{
				compareState: &stateV2{
					stateBase:                   stateBase{},
					waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
					deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
					poolPairs:                   map[string]*PoolPairState{},
					params: Params{
						DefaultFeeRateBPS: 30,
						FeeRateBPS: map[string]uint{
							"abc": 12,
						},
						PRVDiscountPercent:              25,
						LimitProtocolFeePercent:         0,
						LimitStakingPoolRewardPercent:   10,
						TradingProtocolFeePercent:       0,
						TradingStakingPoolRewardPercent: 10,
						DefaultStakingPoolsShare:        0,
						StakingPoolsShare: map[string]uint{
							common.PRVIDStr: 10,
						},
						MintNftRequireAmount: 1000000000,
					},
					stakingPoolStates: map[string]*StakingPoolState{
						common.PRVIDStr: &StakingPoolState{},
					},
					nftIDs:    map[string]uint64{},
					producer:  stateProducerV2{},
					processor: stateProcessorV2{},
				},
				stateChange: &StateChange{
					poolPairIDs: map[string]bool{},
					shares:      map[string]*ShareChange{},
					orderIDs:    map[string]bool{},
					stakingPool: map[string]map[string]*StakingChange{},
				},
			},
			want: &stateV2{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 200, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						shares: map[string]*Share{
							nftID1: &Share{
								amount: 200,
								tradingFees: map[string]uint64{
									common.PRVIDStr: 100,
								},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				params:            Params{},
				stakingPoolStates: map[string]*StakingPoolState{},
				nftIDs:            map[string]uint64{},
				producer:          stateProducerV2{},
				processor:         stateProcessorV2{},
			},
			want1: &StateChange{
				poolPairIDs: map[string]bool{
					poolPairID: true,
				},
				shares: map[string]*ShareChange{
					nftID1: &ShareChange{
						isChanged: true,
						tokenIDs: map[string]bool{
							common.PRVIDStr: true,
						},
					},
				},
				orderIDs:    map[string]bool{},
				stakingPool: map[string]map[string]*StakingChange{},
			},
			wantErr: false,
		},
		{
			name: "Only poolpair trading fees",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 200, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						shares: map[string]*Share{
							nftID1: &Share{
								amount: 200,
								tradingFees: map[string]uint64{
									common.PRVIDStr: 100,
								},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				params: Params{
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PRVDiscountPercent:              25,
					LimitProtocolFeePercent:         0,
					LimitStakingPoolRewardPercent:   10,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
					DefaultStakingPoolsShare:        0,
					StakingPoolsShare: map[string]uint{
						common.PRVIDStr: 10,
					},
					MintNftRequireAmount: 1000000000,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{},
				},
				nftIDs:    map[string]uint64{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			args: args{
				compareState: &stateV2{
					stateBase:                   stateBase{},
					waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
					deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
					poolPairs: map[string]*PoolPairState{
						poolPairID: &PoolPairState{
							state: *rawdbv2.NewPdexv3PoolPairWithValue(
								*token0ID, *token1ID, 200, 100, 400,
								big.NewInt(0).SetUint64(200),
								big.NewInt(0).SetUint64(800), 20000,
							),
							shares: map[string]*Share{
								nftID1: &Share{
									amount:                  200,
									tradingFees:             map[string]uint64{},
									lastUpdatedBeaconHeight: 11,
								},
							},
							orderbook: Orderbook{[]*Order{}},
						},
					},
					params: Params{
						DefaultFeeRateBPS: 30,
						FeeRateBPS: map[string]uint{
							"abc": 12,
						},
						PRVDiscountPercent:              25,
						LimitProtocolFeePercent:         0,
						LimitStakingPoolRewardPercent:   10,
						TradingProtocolFeePercent:       0,
						TradingStakingPoolRewardPercent: 10,
						DefaultStakingPoolsShare:        0,
						StakingPoolsShare: map[string]uint{
							common.PRVIDStr: 10,
						},
						MintNftRequireAmount: 1000000000,
					},
					stakingPoolStates: map[string]*StakingPoolState{
						common.PRVIDStr: &StakingPoolState{},
					},
					nftIDs:    map[string]uint64{},
					producer:  stateProducerV2{},
					processor: stateProcessorV2{},
				},
				stateChange: &StateChange{
					poolPairIDs: map[string]bool{},
					shares:      map[string]*ShareChange{},
					orderIDs:    map[string]bool{},
					stakingPool: map[string]map[string]*StakingChange{},
				},
			},
			want: &stateV2{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 200, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						shares: map[string]*Share{
							nftID1: &Share{
								amount: 200,
								tradingFees: map[string]uint64{
									common.PRVIDStr: 100,
								},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				params:            Params{},
				stakingPoolStates: map[string]*StakingPoolState{},
				nftIDs:            map[string]uint64{},
				producer:          stateProducerV2{},
				processor:         stateProcessorV2{},
			},
			want1: &StateChange{
				poolPairIDs: map[string]bool{},
				shares: map[string]*ShareChange{
					nftID1: &ShareChange{
						tokenIDs: map[string]bool{
							common.PRVIDStr: true,
						},
					},
				},
				orderIDs:    map[string]bool{},
				stakingPool: map[string]map[string]*StakingChange{},
			},
			wantErr: false,
		},
		{
			name: "Only poolpair + share",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 200, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						shares: map[string]*Share{
							nftID1: &Share{
								amount:                  200,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				params: Params{
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PRVDiscountPercent:              25,
					LimitProtocolFeePercent:         0,
					LimitStakingPoolRewardPercent:   10,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
					DefaultStakingPoolsShare:        0,
					StakingPoolsShare: map[string]uint{
						common.PRVIDStr: 10,
					},
					MintNftRequireAmount: 1000000000,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{},
				},
				nftIDs:    map[string]uint64{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			args: args{
				compareState: &stateV2{
					stateBase:                   stateBase{},
					waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
					deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
					poolPairs:                   map[string]*PoolPairState{},
					params: Params{
						DefaultFeeRateBPS: 30,
						FeeRateBPS: map[string]uint{
							"abc": 12,
						},
						PRVDiscountPercent:              25,
						LimitProtocolFeePercent:         0,
						LimitStakingPoolRewardPercent:   10,
						TradingProtocolFeePercent:       0,
						TradingStakingPoolRewardPercent: 10,
						DefaultStakingPoolsShare:        0,
						StakingPoolsShare: map[string]uint{
							common.PRVIDStr: 10,
						},
						MintNftRequireAmount: 1000000000,
					},
					stakingPoolStates: map[string]*StakingPoolState{
						common.PRVIDStr: &StakingPoolState{},
					},
					nftIDs:    map[string]uint64{},
					producer:  stateProducerV2{},
					processor: stateProcessorV2{},
				},
				stateChange: &StateChange{
					poolPairIDs: map[string]bool{},
					shares:      map[string]*ShareChange{},
					orderIDs:    map[string]bool{},
					stakingPool: map[string]map[string]*StakingChange{},
				},
			},
			want: &stateV2{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 200, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						shares: map[string]*Share{
							nftID1: &Share{
								amount:                  200,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				params:            Params{},
				stakingPoolStates: map[string]*StakingPoolState{},
				nftIDs:            map[string]uint64{},
				producer:          stateProducerV2{},
				processor:         stateProcessorV2{},
			},
			want1: &StateChange{
				poolPairIDs: map[string]bool{
					poolPairID: true,
				},
				shares: map[string]*ShareChange{
					nftID1: &ShareChange{
						isChanged: true,
					},
				},
				orderIDs:    map[string]bool{},
				stakingPool: map[string]map[string]*StakingChange{},
			},
			wantErr: false,
		},
		{
			name: "Full stakingPoolStates",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				params: Params{
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PRVDiscountPercent:              25,
					LimitProtocolFeePercent:         0,
					LimitStakingPoolRewardPercent:   10,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
					DefaultStakingPoolsShare:        0,
					StakingPoolsShare: map[string]uint{
						common.PRVIDStr: 10,
					},
					MintNftRequireAmount: 1000000000,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 100,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity:               100,
								lastUpdatedBeaconHeight: 20,
								rewards: map[string]uint64{
									common.PRVIDStr: 100,
								},
							},
						},
					},
				},
				nftIDs:    map[string]uint64{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			args: args{
				compareState: &stateV2{
					stateBase:                   stateBase{},
					waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
					deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
					poolPairs:                   map[string]*PoolPairState{},
					params: Params{
						DefaultFeeRateBPS: 30,
						FeeRateBPS: map[string]uint{
							"abc": 12,
						},
						PRVDiscountPercent:              25,
						LimitProtocolFeePercent:         0,
						LimitStakingPoolRewardPercent:   10,
						TradingProtocolFeePercent:       0,
						TradingStakingPoolRewardPercent: 10,
						DefaultStakingPoolsShare:        0,
						StakingPoolsShare: map[string]uint{
							common.PRVIDStr: 10,
						},
						MintNftRequireAmount: 1000000000,
					},
					stakingPoolStates: map[string]*StakingPoolState{
						common.PRVIDStr: &StakingPoolState{},
					},
					nftIDs:    map[string]uint64{},
					producer:  stateProducerV2{},
					processor: stateProcessorV2{},
				},
				stateChange: &StateChange{
					poolPairIDs: map[string]bool{},
					shares:      map[string]*ShareChange{},
					orderIDs:    map[string]bool{},
					stakingPool: map[string]map[string]*StakingChange{},
				},
			},
			want: &stateV2{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				params:                      Params{},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 100,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity:               100,
								lastUpdatedBeaconHeight: 20,
								rewards: map[string]uint64{
									common.PRVIDStr: 100,
								},
							},
						},
					},
				},
				nftIDs:    map[string]uint64{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			want1: &StateChange{
				poolPairIDs: map[string]bool{},
				shares:      map[string]*ShareChange{},
				orderIDs:    map[string]bool{},
				stakingPool: map[string]map[string]*StakingChange{
					common.PRVIDStr: map[string]*StakingChange{
						nftID1: &StakingChange{
							isChanged: true,
							tokenIDs: map[string]bool{
								common.PRVIDStr: true,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Only stakingPoolStates rewards",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				params: Params{
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PRVDiscountPercent:              25,
					LimitProtocolFeePercent:         0,
					LimitStakingPoolRewardPercent:   10,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
					DefaultStakingPoolsShare:        0,
					StakingPoolsShare: map[string]uint{
						common.PRVIDStr: 10,
					},
					MintNftRequireAmount: 1000000000,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 100,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity:               100,
								lastUpdatedBeaconHeight: 20,
								rewards: map[string]uint64{
									common.PRVIDStr: 100,
								},
							},
						},
					},
				},
				nftIDs:    map[string]uint64{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			args: args{
				compareState: &stateV2{
					stateBase:                   stateBase{},
					waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
					deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
					poolPairs:                   map[string]*PoolPairState{},
					params: Params{
						DefaultFeeRateBPS: 30,
						FeeRateBPS: map[string]uint{
							"abc": 12,
						},
						PRVDiscountPercent:              25,
						LimitProtocolFeePercent:         0,
						LimitStakingPoolRewardPercent:   10,
						TradingProtocolFeePercent:       0,
						TradingStakingPoolRewardPercent: 10,
						DefaultStakingPoolsShare:        0,
						StakingPoolsShare: map[string]uint{
							common.PRVIDStr: 10,
						},
						MintNftRequireAmount: 1000000000,
					},
					stakingPoolStates: map[string]*StakingPoolState{
						common.PRVIDStr: &StakingPoolState{
							liquidity: 100,
							stakers: map[string]*Staker{
								nftID1: &Staker{
									liquidity:               100,
									lastUpdatedBeaconHeight: 20,
									rewards:                 map[string]uint64{},
								},
							},
						},
					},
					nftIDs:    map[string]uint64{},
					producer:  stateProducerV2{},
					processor: stateProcessorV2{},
				},
				stateChange: &StateChange{
					poolPairIDs: map[string]bool{},
					shares:      map[string]*ShareChange{},
					orderIDs:    map[string]bool{},
					stakingPool: map[string]map[string]*StakingChange{},
				},
			},
			want: &stateV2{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				params:                      Params{},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 100,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity:               100,
								lastUpdatedBeaconHeight: 20,
								rewards: map[string]uint64{
									common.PRVIDStr: 100,
								},
							},
						},
					},
				},
				nftIDs:    map[string]uint64{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			want1: &StateChange{
				poolPairIDs: map[string]bool{},
				shares:      map[string]*ShareChange{},
				orderIDs:    map[string]bool{},
				stakingPool: map[string]map[string]*StakingChange{
					common.PRVIDStr: map[string]*StakingChange{
						nftID1: &StakingChange{
							tokenIDs: map[string]bool{
								common.PRVIDStr: true,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Only stakingPoolStates + stakers",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				params: Params{
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PRVDiscountPercent:              25,
					LimitProtocolFeePercent:         0,
					LimitStakingPoolRewardPercent:   10,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
					DefaultStakingPoolsShare:        0,
					StakingPoolsShare: map[string]uint{
						common.PRVIDStr: 10,
					},
					MintNftRequireAmount: 1000000000,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 100,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity:               100,
								lastUpdatedBeaconHeight: 20,
								rewards: map[string]uint64{
									common.PRVIDStr: 100,
								},
							},
						},
					},
				},
				nftIDs:    map[string]uint64{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			args: args{
				compareState: &stateV2{
					stateBase:                   stateBase{},
					waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
					deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
					poolPairs:                   map[string]*PoolPairState{},
					params: Params{
						DefaultFeeRateBPS: 30,
						FeeRateBPS: map[string]uint{
							"abc": 12,
						},
						PRVDiscountPercent:              25,
						LimitProtocolFeePercent:         0,
						LimitStakingPoolRewardPercent:   10,
						TradingProtocolFeePercent:       0,
						TradingStakingPoolRewardPercent: 10,
						DefaultStakingPoolsShare:        0,
						StakingPoolsShare: map[string]uint{
							common.PRVIDStr: 10,
						},
						MintNftRequireAmount: 1000000000,
					},
					stakingPoolStates: map[string]*StakingPoolState{
						common.PRVIDStr: &StakingPoolState{
							liquidity: 100,
							stakers: map[string]*Staker{
								nftID1: &Staker{
									liquidity:               100,
									lastUpdatedBeaconHeight: 20,
									rewards:                 map[string]uint64{},
								},
							},
						},
					},
					nftIDs:    map[string]uint64{},
					producer:  stateProducerV2{},
					processor: stateProcessorV2{},
				},
				stateChange: &StateChange{
					poolPairIDs: map[string]bool{},
					shares:      map[string]*ShareChange{},
					orderIDs:    map[string]bool{},
					stakingPool: map[string]map[string]*StakingChange{},
				},
			},
			want: &stateV2{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				params:                      Params{},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 100,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity:               100,
								lastUpdatedBeaconHeight: 20,
								rewards: map[string]uint64{
									common.PRVIDStr: 100,
								},
							},
						},
					},
				},
				nftIDs:    map[string]uint64{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			want1: &StateChange{
				poolPairIDs: map[string]bool{},
				shares:      map[string]*ShareChange{},
				orderIDs:    map[string]bool{},
				stakingPool: map[string]map[string]*StakingChange{
					common.PRVIDStr: map[string]*StakingChange{
						nftID1: &StakingChange{
							tokenIDs: map[string]bool{
								common.PRVIDStr: true,
							},
						},
					},
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
				stakingPoolStates:           tt.fields.stakingPoolStates,
				nftIDs:                      tt.fields.nftIDs,
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
