package pdex

import (
	"encoding/json"
	"math/big"
	"reflect"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	v2 "github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
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

func Test_stateProcessorV2_waitingContribution(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	firstTxHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)
	nftHash, err := common.Hash{}.NewHashFromStr(nftID)
	assert.Nil(t, err)

	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	// first contribution tx
	firstContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0,
		token0ID.String(), 100, 20000,
		&metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	)
	assert.Nil(t, err)
	contributionTx := &metadataMocks.Transaction{}
	contributionTx.On("GetMetadata").Return(firstContributionMetadata)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)
	contributionTx.On("GetValidationEnv").Return(valEnv)
	contributionTx.On("Hash").Return(firstTxHash)
	waitingContributionStateDB := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0,
			*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	waitingContributionInst := instruction.NewWaitingAddLiquidityWithValue(*waitingContributionStateDB)
	waitingContributionInstBytes, err := json.Marshal(waitingContributionInst)
	//

	type fields struct {
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB              *statedb.StateDB
		inst                 []string
		waitingContributions map[string]rawdbv2.Pdexv3Contribution
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]rawdbv2.Pdexv3Contribution
		want1   *v2utils.ContributionStatus
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				stateProcessorBase: stateProcessorBase{},
			},
			args: args{
				stateDB: sDB,
				inst: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionWaitingChainStatus,
					string(waitingContributionInstBytes),
				},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
			},
			want: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					"", validOTAReceiver0,
					*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
					nil, nil,
				),
			},
			want1: &v2utils.ContributionStatus{
				Token0ID:                token0ID.String(),
				Token0ContributedAmount: 100,
				Status:                  common.PDEContributionWaitingStatus,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, err := sp.waitingContribution(tt.args.stateDB, tt.args.inst, tt.args.waitingContributions)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.waitingContribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("got1 = %v, want %v", got1, tt.want1)
				return
			}
		})
	}
}

func Test_stateProcessorV2_refundContribution(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	firstTxHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)
	nftHash, err := common.Hash{}.NewHashFromStr(nftID)
	assert.Nil(t, err)

	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	// return contribution tx by sanity
	refundContributionSanityMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0,
		token0ID.String(), 200, 20000,
		&metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	)
	assert.Nil(t, err)
	refundContributionSanityTx := &metadataMocks.Transaction{}
	refundContributionSanityTx.On("GetMetadata").Return(refundContributionSanityMetaData)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)
	refundContributionSanityTx.On("GetValidationEnv").Return(valEnv)
	refundContributionSanityTx.On("Hash").Return(firstTxHash)
	refundContributionSanityState := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0,
			*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	refundContributionSanityInst := instruction.NewRefundAddLiquidityWithValue(*refundContributionSanityState)
	refundContributionSanityInstBytes, err := json.Marshal(refundContributionSanityInst)
	//

	type fields struct {
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB                     *statedb.StateDB
		inst                        []string
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]rawdbv2.Pdexv3Contribution
		want1   map[string]rawdbv2.Pdexv3Contribution
		want2   *v2utils.ContributionStatus
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				stateProcessorBase: stateProcessorBase{},
			},
			args: args{
				stateDB: sDB,
				inst: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionRefundChainStatus,
					string(refundContributionSanityInstBytes),
				},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						"", validOTAReceiver0,
						*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
						nil, nil,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
			},
			want: map[string]rawdbv2.Pdexv3Contribution{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					"", validOTAReceiver0,
					*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
					nil, nil,
				),
			},
			want2: &v2utils.ContributionStatus{
				Status: common.PDEContributionRefundStatus,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, got2, err := sp.refundContribution(tt.args.stateDB, tt.args.inst, tt.args.waitingContributions, tt.args.deletedWaitingContributions)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.refundContribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.refundContribution() got = %v, want %v", got, tt.want)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProcessorV2.refundContribution() got1 = %v, want %v", got1, tt.want1)
				return
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("stateProcessorV2.refundContribution() got2 = %v, want %v", got2, tt.want2)
				return
			}
		})
	}
}

func Test_stateProcessorV2_matchContribution(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)
	firstTxHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)
	secondTxHash, err := common.Hash{}.NewHashFromStr("aaa")
	assert.Nil(t, err)
	nftHash, err := common.Hash{}.NewHashFromStr(nftID)
	assert.Nil(t, err)

	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	// match contribution
	matchContributionMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0,
		token1ID.String(), 400, 20000,
		&metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	)
	assert.Nil(t, err)
	matchContributionTx := &metadataMocks.Transaction{}
	matchContributionTx.On("GetMetadata").Return(matchContributionMetaData)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)
	matchContributionTx.On("GetValidationEnv").Return(valEnv)
	matchContributionTx.On("Hash").Return(secondTxHash)
	matchContributionState := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0,
			*token1ID, *secondTxHash, *nftHash, 400, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	matchContributionInst := instruction.NewMatchAddLiquidityWithValue(*matchContributionState, poolPairID)
	matchContributionInstBytes, err := json.Marshal(matchContributionInst)
	//

	type fields struct {
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB                     *statedb.StateDB
		inst                        []string
		beaconHeight                uint64
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		nftIDs                      map[string]uint64
		params                      *Params
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]rawdbv2.Pdexv3Contribution
		want1   map[string]rawdbv2.Pdexv3Contribution
		want2   map[string]*PoolPairState
		want3   *v2utils.ContributionStatus
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				stateProcessorBase: stateProcessorBase{},
			},
			args: args{
				beaconHeight: 11,
				stateDB:      sDB,
				inst: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedChainStatus,
					string(matchContributionInstBytes),
				},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						"", validOTAReceiver0,
						*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
						nil, nil,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				nftIDs:                      map[string]uint64{},
				params:                      NewParams(),
			},
			want: map[string]rawdbv2.Pdexv3Contribution{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					"", validOTAReceiver0,
					*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
					nil, nil,
				),
			},
			want2: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 200, 0, 100, 400,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(800), 20000,
					),
					lpFeesPerShare:    map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount:                200,
							tradingFees:           map[common.Hash]uint64{},
							lastLPFeesPerShare:    map[common.Hash]*big.Int{},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderRewards:  map[string]*OrderReward{},
					makingVolume:  map[common.Hash]*MakingVolume{},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			want3: &v2utils.ContributionStatus{
				Status:   common.PDEContributionAcceptedStatus,
				AccessID: nftHash,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, got2, got3, err := sp.matchContribution(
				tt.args.stateDB, tt.args.inst,
				tt.args.beaconHeight, tt.args.waitingContributions,
				tt.args.deletedWaitingContributions, tt.args.poolPairs,
				tt.args.params,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.matchContribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.matchContribution() got = %v, want %v", got, tt.want)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProcessorV2.matchContribution() got1 = %v, want %v", got1, tt.want1)
				return
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("stateProcessorV2.matchContribution() got2 = %v, want %v", got2, tt.want2)
				return
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("stateProcessorV2.matchContribution() got3 = %v, want %v", got3, tt.want3)
				return
			}
		})
	}
}

func Test_stateProcessorV2_matchAndReturnContribution(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)
	secondTxHash, err := common.Hash{}.NewHashFromStr("aaa")
	assert.Nil(t, err)
	thirdTxHash, err := common.Hash{}.NewHashFromStr("abb")
	assert.Nil(t, err)
	nftHash, err := common.Hash{}.NewHashFromStr(nftID)
	assert.Nil(t, err)

	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	// match and return contribution
	matchAndReturnContributionMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		poolPairID, "pair_hash",
		validOTAReceiver0,
		token1ID.String(), 200, 20000,
		&metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)

	matchAndReturnContributionTx := &metadataMocks.Transaction{}
	matchAndReturnContributionTx.On("GetMetadata").Return(matchAndReturnContributionMetaData)
	matchAndReturnContributionTx.On("GetValidationEnv").Return(valEnv)
	matchAndReturnContributionTx.On("Hash").Return(thirdTxHash)

	matchAndReturnContribution0State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0,
			*token0ID, *thirdTxHash, *nftHash, 50, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	matchAndReturnContritubtion0Inst := instruction.NewMatchAndReturnAddLiquidityWithValue(
		*matchAndReturnContribution0State, 100, 0, 200, 0, *token1ID, nil)
	matchAndReturnContritubtion0InstBytes, err := json.Marshal(matchAndReturnContritubtion0Inst)
	//

	type fields struct {
		pairHashCache      map[string]common.Hash
		withdrawTxCache    map[string]uint64
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB                     *statedb.StateDB
		inst                        []string
		beaconHeight                uint64
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		nftIDs                      map[string]uint64
		params                      *Params
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]rawdbv2.Pdexv3Contribution
		want1   map[string]rawdbv2.Pdexv3Contribution
		want2   map[string]*PoolPairState
		want3   *v2utils.ContributionStatus
		wantErr bool
	}{
		{
			name: "First Instruction",
			fields: fields{
				pairHashCache:      map[string]common.Hash{},
				stateProcessorBase: stateProcessorBase{},
			},
			args: args{
				beaconHeight: 11,
				stateDB:      sDB,
				inst: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedNReturnedChainStatus,
					string(matchAndReturnContritubtion0InstBytes),
				},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						poolPairID, validOTAReceiver0,
						*token0ID, *secondTxHash, *nftHash, 100, 20000, 1,
						nil, nil,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 200, 0, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount:                200,
								tradingFees:           map[common.Hash]uint64{},
								lastLPFeesPerShare:    map[common.Hash]*big.Int{},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				nftIDs: map[string]uint64{},
				params: NewParams(),
			},
			want: map[string]rawdbv2.Pdexv3Contribution{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					poolPairID, validOTAReceiver0,
					*token0ID, *secondTxHash, *nftHash, 100, 20000, 1,
					nil, nil,
				),
			},
			want2: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 0, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					lpFeesPerShare:    map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount:                300,
							tradingFees:           map[common.Hash]uint64{},
							lastLPFeesPerShare:    map[common.Hash]*big.Int{},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			want3:   &v2utils.ContributionStatus{},
			wantErr: false,
		},
		{
			name: "Second Instruction",
			fields: fields{
				pairHashCache: map[string]common.Hash{
					"pair_hash": *thirdTxHash,
				},
				stateProcessorBase: stateProcessorBase{},
			},
			args: args{
				beaconHeight: 11,
				stateDB:      sDB,
				inst: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedNReturnedChainStatus,
					string(matchAndReturnContritubtion0InstBytes),
				},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						poolPairID, validOTAReceiver0,
						*token0ID, *secondTxHash, *nftHash, 100, 20000, 1,
						nil, nil,
					),
				},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 0, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount:                300,
								tradingFees:           map[common.Hash]uint64{},
								lastLPFeesPerShare:    map[common.Hash]*big.Int{},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				params: NewParams(),
			},
			want: map[string]rawdbv2.Pdexv3Contribution{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					poolPairID, validOTAReceiver0,
					*token0ID, *secondTxHash, *nftHash, 100, 20000, 1,
					nil, nil,
				),
			},
			want2: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 0, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					lpFeesPerShare:    map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount:                300,
							tradingFees:           map[common.Hash]uint64{},
							lastLPFeesPerShare:    map[common.Hash]*big.Int{},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			want3: &v2utils.ContributionStatus{
				Token0ID:                token0ID.String(),
				Token0ContributedAmount: 50,
				Status:                  common.PDEContributionMatchedNReturnedStatus,
				Token1ID:                token1ID.String(),
				Token1ContributedAmount: 200,
				PoolPairID:              poolPairID,
				AccessID:                nftHash,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				pairHashCache:      tt.fields.pairHashCache,
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, got2, got3, err := sp.matchAndReturnContribution(
				tt.args.stateDB, tt.args.inst,
				tt.args.beaconHeight, tt.args.waitingContributions,
				tt.args.deletedWaitingContributions, tt.args.poolPairs,
				tt.args.params,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.matchAndReturnContribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.matchAndReturnContribution() got = %v, want %v", got, tt.want)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProcessorV2.matchAndReturnContribution() got1 = %v, want %v", got1, tt.want1)
				return
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("stateProcessorV2.matchAndReturnContribution() got2 = %v, want %v", got2, tt.want2)
				return
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("stateProcessorV2.matchAndReturnContribution() got3 = %v, want %v", got3, tt.want3)
				return
			}
		})
	}
}

func Test_stateProcessorV2_acceptWithdrawLiquidity(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)
	txHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)
	nftHash, err := common.Hash{}.NewHashFromStr(nftID)
	assert.Nil(t, err)

	// invalid insturction
	invalidInst, err := instruction.NewRejectWithdrawLiquidityWithValue(
		*txHash, 1, "", nil, nil,
	).StringSlice()
	assert.Nil(t, err)
	//

	// invalidPoolPairID insturction
	invalidPoolPairIDInst, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		"123", *token0ID, 50, 100, validOTAReceiver0,
		*txHash, 1, metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	).StringSlice()
	assert.Nil(t, err)
	//

	// valid insturction
	acceptWithdrawLiquidityInst0, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *token0ID, 50, 100, validOTAReceiver0,
		*txHash, 1, metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	).StringSlice()
	assert.Nil(t, err)

	// valid insturction
	acceptWithdrawLiquidityInst1, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *token1ID, 200, 100, validOTAReceiver1,
		*txHash, 1, metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	).StringSlice()
	assert.Nil(t, err)

	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	type fields struct {
		withdrawTxCache    map[string]uint64
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB        *statedb.StateDB
		inst           []string
		poolPairs      map[string]*PoolPairState
		beaconHeight   uint64
		lmLockedBlocks uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]*PoolPairState
		want1   *v2utils.WithdrawStatus
		wantErr bool
	}{
		{
			name:   "Invalid instruction",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    invalidInst,
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 0, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount:                300,
								tradingFees:           map[common.Hash]uint64{},
								lastLPFeesPerShare:    map[common.Hash]*big.Int{},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 0, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					lpFeesPerShare:    map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount:                300,
							tradingFees:           map[common.Hash]uint64{},
							lastLPFeesPerShare:    map[common.Hash]*big.Int{},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid pool pair id",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    invalidPoolPairIDInst,
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 0, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount:                300,
								tradingFees:           map[common.Hash]uint64{},
								lastLPFeesPerShare:    map[common.Hash]*big.Int{},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 0, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					lpFeesPerShare:    map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount:                300,
							tradingFees:           map[common.Hash]uint64{},
							lastLPFeesPerShare:    map[common.Hash]*big.Int{},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid Input - Token 0",
			fields: fields{
				withdrawTxCache:    map[string]uint64{},
				stateProcessorBase: stateProcessorBase{},
			},
			args: args{
				stateDB: sDB,
				inst:    acceptWithdrawLiquidityInst0,
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 0, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:                300,
								tradingFees:           map[common.Hash]uint64{},
								lastLPFeesPerShare:    map[common.Hash]*big.Int{},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook:         Orderbook{[]*Order{}},
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						lmLockedShare:     map[string]map[uint64]uint64{},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 0, 100, 600,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:                300,
							tradingFees:           map[common.Hash]uint64{},
							lastLPFeesPerShare:    map[common.Hash]*big.Int{},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderbook:         Orderbook{[]*Order{}},
					lpFeesPerShare:    map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					lmLockedShare:     map[string]map[uint64]uint64{},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid Input - Token 1",
			fields: fields{
				withdrawTxCache: map[string]uint64{
					txHash.String(): 50,
				},
				stateProcessorBase: stateProcessorBase{},
			},
			args: args{
				stateDB: sDB,
				inst:    acceptWithdrawLiquidityInst1,
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 0, 100, 600,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount:                300,
								tradingFees:           map[common.Hash]uint64{},
								lastLPFeesPerShare:    map[common.Hash]*big.Int{},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 200, 0, 100, 400,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(800), 20000,
					),
					lpFeesPerShare:    map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount:                200,
							tradingFees:           map[common.Hash]uint64{},
							lastLPFeesPerShare:    map[common.Hash]*big.Int{},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			want1: &v2utils.WithdrawStatus{
				Status:       common.Pdexv3AcceptStatus,
				Token0ID:     token0ID.String(),
				Token0Amount: 50,
				Token1ID:     token1ID.String(),
				Token1Amount: 200,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				withdrawTxCache:    tt.fields.withdrawTxCache,
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, err := sp.acceptWithdrawLiquidity(tt.args.stateDB, tt.args.inst, tt.args.poolPairs, tt.args.beaconHeight, tt.args.lmLockedBlocks)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.acceptWithdrawLiquidity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("got1 = %v, want1 %v", got1, tt.want1)
			}
		})
	}
}

func Test_stateProcessorV2_rejectWithdrawLiquidity(t *testing.T) {
	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	nftHash, err := common.Hash{}.NewHashFromStr(nftID)
	assert.Nil(t, err)

	// valid insturction
	inst, err := instruction.NewRejectWithdrawLiquidityWithValue(*nftHash, 1, "", nil, nil).StringSlice()
	assert.Nil(t, err)

	type fields struct {
		pairHashCache      map[string]common.Hash
		withdrawTxCache    map[string]uint64
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB   *statedb.StateDB
		inst      []string
		poolPairs map[string]*PoolPairState
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v2utils.WithdrawStatus
		wantErr bool
	}{
		{
			name:   "Valid input",
			fields: fields{},
			args: args{
				stateDB:   sDB,
				inst:      inst,
				poolPairs: map[string]*PoolPairState{},
			},
			want: &v2utils.WithdrawStatus{
				Status: common.Pdexv3RejectStatus,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				pairHashCache:      tt.fields.pairHashCache,
				withdrawTxCache:    tt.fields.withdrawTxCache,
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, err := sp.rejectWithdrawLiquidity(tt.args.stateDB, tt.args.inst, tt.args.poolPairs)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.rejectWithdrawLiquidity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.rejectWithdrawLiquidity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stateProcessorV2_userMintNft(t *testing.T) {
	txReqID, err := common.Hash{}.NewHashFromStr("1111122222")
	assert.Nil(t, err)
	nftHash1, err := common.Hash{}.NewHashFromStr(nftID1)
	assert.Nil(t, err)

	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	rejectInst, err := instruction.NewRejectUserMintNftWithValue(validOTAReceiver0, 100, 1, *txReqID).StringSlice()
	assert.Nil(t, err)
	acceptInst, err := instruction.NewAcceptUserMintNftWithValue(
		validOTAReceiver0, 100, 1, *nftHash1, *txReqID,
	).StringSlice()
	assert.Nil(t, err)

	type fields struct {
		pairHashCache      map[string]common.Hash
		withdrawTxCache    map[string]uint64
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB *statedb.StateDB
		inst    []string
		nftIDs  map[string]uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]uint64
		want1   *v2utils.MintNftStatus
		wantErr bool
	}{
		{
			name:   "Reject mint nft",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    rejectInst,
				nftIDs:  map[string]uint64{},
			},
			want: map[string]uint64{},
			want1: &v2utils.MintNftStatus{
				Status:      common.Pdexv3RejectStatus,
				BurntAmount: 100,
			},
			wantErr: false,
		},
		{
			name: "Accept mint nft",
			args: args{
				stateDB: sDB,
				inst:    acceptInst,
				nftIDs:  map[string]uint64{},
			},
			want: map[string]uint64{
				nftID1: 100,
			},
			want1: &v2utils.MintNftStatus{
				Status:      common.Pdexv3AcceptStatus,
				BurntAmount: 100,
				NftID:       nftID1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				pairHashCache:      tt.fields.pairHashCache,
				withdrawTxCache:    tt.fields.withdrawTxCache,
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			atc, err := (&v2utils.NFTAssetTagsCache{}).FromIDs(tt.args.nftIDs)
			assert.Nil(t, err)
			got, got1, err := sp.userMintNft(tt.args.stateDB, tt.args.inst, tt.args.nftIDs, atc)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.userMintNft() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.userMintNft() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProcessorV2.userMintNft() got1 = %v, want %v", got1, tt.want1)
			}

			// check assetTagsCache consistency
			expectedAtc, err := (&v2utils.NFTAssetTagsCache{}).FromIDs(tt.want)
			assert.Nil(t, err)
			if !reflect.DeepEqual(atc, expectedAtc) {
				t.Errorf("stateProcessorV2.userMintNft() got1 = %v, want %v", atc, expectedAtc)
			}
		})
	}
}

func Test_stateProcessorV2_staking(t *testing.T) {
	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	txReqID, err := common.Hash{}.NewHashFromStr("1111122222")
	assert.Nil(t, err)
	nftHash1, err := common.Hash{}.NewHashFromStr(nftID1)
	assert.Nil(t, err)

	rejectInst, err := instruction.NewRejectStakingWithValue(
		validOTAReceiver0, common.PRVCoinID, *txReqID, 1, 100,
	).StringSlice()
	assert.Nil(t, err)
	acceptInst, err := instruction.NewAcceptStakingWithValue(
		common.PRVCoinID, *txReqID, 1, 100, nil, metadataPdexv3.AccessOption{
			NftID: nftHash1,
		},
	).StringSlice()
	assert.Nil(t, err)

	type fields struct {
		pairHashCache      map[string]common.Hash
		withdrawTxCache    map[string]uint64
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB           *statedb.StateDB
		inst              []string
		nftIDs            map[string]uint64
		stakingPoolStates map[string]*StakingPoolState
		beaconHeight      uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]*StakingPoolState
		want1   *v2.StakingStatus
		wantErr bool
	}{
		{
			name:   "Reject inst",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    rejectInst,
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						stakers: map[string]*Staker{},
					},
				},
			},
			want: map[string]*StakingPoolState{
				common.PRVIDStr: &StakingPoolState{
					stakers: map[string]*Staker{},
				},
			},
			want1: &v2.StakingStatus{
				Status:        common.Pdexv3RejectStatus,
				StakingPoolID: common.PRVIDStr,
				Liquidity:     100,
			},
			wantErr: false,
		},
		{
			name:   "Accept inst",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    acceptInst,
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity:       0,
						stakers:         map[string]*Staker{},
						rewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*StakingPoolState{
				common.PRVIDStr: &StakingPoolState{
					liquidity: 100,
					stakers: map[string]*Staker{
						nftID1: &Staker{
							liquidity:           100,
							rewards:             map[common.Hash]uint64{},
							lastRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					rewardsPerShare: map[common.Hash]*big.Int{},
				},
			},
			want1: &v2.StakingStatus{
				Status:        common.Pdexv3AcceptStatus,
				StakingPoolID: common.PRVIDStr,
				Liquidity:     100,
				NftID:         nftID1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				pairHashCache:      tt.fields.pairHashCache,
				withdrawTxCache:    tt.fields.withdrawTxCache,
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, err := sp.staking(tt.args.stateDB, tt.args.inst, tt.args.nftIDs, tt.args.stakingPoolStates, tt.args.beaconHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.staking() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.staking() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProcessorV2.staking() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_stateProcessorV2_feeTrackingAddLiquidity(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)
	firstTxHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)
	secondTxHash, err := common.Hash{}.NewHashFromStr("aaa")
	assert.Nil(t, err)
	nftHash, err := common.Hash{}.NewHashFromStr(nftID)
	assert.Nil(t, err)

	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	// match contribution
	matchContributionMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0,
		token1ID.String(), 400, 20000,
		metadataPdexv3.NewAccessOptionWithValue(nil, nftHash, nil),
		nil,
	)
	assert.Nil(t, err)
	matchContributionTx := &metadataMocks.Transaction{}
	matchContributionTx.On("GetMetadata").Return(matchContributionMetaData)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)
	matchContributionTx.On("GetValidationEnv").Return(valEnv)
	matchContributionTx.On("Hash").Return(secondTxHash)
	matchContributionState := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0,
			*token1ID, *secondTxHash, *nftHash, 400, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	matchContributionInst := instruction.NewMatchAddLiquidityWithValue(*matchContributionState, poolPairID)
	matchContributionInstBytes, err := json.Marshal(matchContributionInst)
	//

	type fields struct {
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB                     *statedb.StateDB
		inst                        []string
		beaconHeight                uint64
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		nftIDs                      map[string]uint64
		params                      *Params
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]rawdbv2.Pdexv3Contribution
		want1   map[string]rawdbv2.Pdexv3Contribution
		want2   map[string]*PoolPairState
		want3   *v2utils.ContributionStatus
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				stateProcessorBase: stateProcessorBase{},
			},
			args: args{
				beaconHeight: 11,
				stateDB:      sDB,
				inst: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedChainStatus,
					string(matchContributionInstBytes),
				},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						"", validOTAReceiver0,
						*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
						nil, nil,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				nftIDs:                      map[string]uint64{},
				params: &Params{
					PDEXRewardPoolPairsShare:  map[string]uint{poolPairID: 1},
					MiningRewardPendingBlocks: 50,
				},
			},
			want: map[string]rawdbv2.Pdexv3Contribution{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					"", validOTAReceiver0,
					*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
					nil, nil,
				),
			},
			want2: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 200, 200, 100, 400,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(800), 20000,
					),
					lpFeesPerShare:    map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount:                200,
							lmLockedAmount:        200,
							tradingFees:           map[common.Hash]uint64{},
							lastLPFeesPerShare:    map[common.Hash]*big.Int{},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderRewards: map[string]*OrderReward{},
					makingVolume: map[common.Hash]*MakingVolume{},
					orderbook:    Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{
						nftID: {
							11: 200,
						},
					},
				},
			},
			want3: &v2utils.ContributionStatus{
				Status:   common.PDEContributionAcceptedStatus,
				AccessID: nftHash,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, got2, got3, err := sp.matchContribution(
				tt.args.stateDB, tt.args.inst,
				tt.args.beaconHeight, tt.args.waitingContributions,
				tt.args.deletedWaitingContributions, tt.args.poolPairs,
				tt.args.params,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.matchContribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.matchContribution() got = %v, want %v", got, tt.want)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProcessorV2.matchContribution() got1 = %v, want %v", got1, tt.want1)
				return
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("stateProcessorV2.matchContribution() got2 = %v, want %v", got2, tt.want2)
				return
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("stateProcessorV2.matchContribution() got3 = %v, want %v", got3, tt.want3)
				return
			}
		})
	}
}

func Test_stateProcessorV2_feeTrackingWithdrawLiquidity(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)
	txHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)
	nftHash, err := common.Hash{}.NewHashFromStr(nftID)
	assert.Nil(t, err)
	nft1Hash, err := common.Hash{}.NewHashFromStr(nftID1)
	assert.Nil(t, err)

	// valid instruction
	acceptWithdrawLiquidityInst1, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *token1ID, 200, 100, validOTAReceiver1,
		*txHash, 1, metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	).StringSlice()
	assert.Nil(t, err)

	// valid instruction
	acceptWithdrawLiquidityInst2, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *token1ID, 100, 50, validOTAReceiver1,
		*txHash, 1, metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	).StringSlice()
	acceptWithdrawLiquidityInst3, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *token1ID, 200, 100, validOTAReceiver1,
		*txHash, 1, metadataPdexv3.AccessOption{
			NftID: nft1Hash,
		}, nil,
	).StringSlice()

	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	type fields struct {
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB        *statedb.StateDB
		inst           [][]string
		poolPairs      map[string]*PoolPairState
		beaconHeight   uint64
		lmLockedBlocks uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]*PoolPairState
		wantErr bool
	}{
		{
			name:   "Single withdrawal",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    [][]string{acceptWithdrawLiquidityInst1},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 150, 100, 600,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare: map[common.Hash]*big.Int{
							*token0ID:         convertToLPFeesPerShare(100, 300),
							*token1ID:         convertToLPFeesPerShare(200, 300),
							common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
							common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
						},
						lmRewardsPerShare: map[common.Hash]*big.Int{
							common.PRVCoinID: convertToLPFeesPerShare(10, 150),
						},
						protocolFees:    map[common.Hash]uint64{},
						stakingPoolFees: map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount:                300,
								lmLockedAmount:        150,
								tradingFees:           map[common.Hash]uint64{},
								lastLPFeesPerShare:    map[common.Hash]*big.Int{},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 200, 50, 100, 400,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(800), 20000,
					),
					lpFeesPerShare: map[common.Hash]*big.Int{
						*token0ID:         convertToLPFeesPerShare(100, 300),
						*token1ID:         convertToLPFeesPerShare(200, 300),
						common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
						common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
					},
					lmRewardsPerShare: map[common.Hash]*big.Int{
						common.PRVCoinID: convertToLPFeesPerShare(10, 150),
					},
					protocolFees:    map[common.Hash]uint64{},
					stakingPoolFees: map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount:         200,
							lmLockedAmount: 50,
							tradingFees: map[common.Hash]uint64{
								*token0ID:         99,
								*token1ID:         199,
								common.PRVCoinID:  18,
								common.PDEXCoinID: 19,
							},
							lastLPFeesPerShare: map[common.Hash]*big.Int{
								*token0ID:         convertToLPFeesPerShare(100, 300),
								*token1ID:         convertToLPFeesPerShare(200, 300),
								common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
								common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
							},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{
								common.PRVCoinID: convertToLPFeesPerShare(10, 150),
							},
						},
					},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			wantErr: false,
		},
		{
			name:   "Double withdrawal",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    [][]string{acceptWithdrawLiquidityInst1, acceptWithdrawLiquidityInst1},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 0, 100, 600,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare: map[common.Hash]*big.Int{
							*token0ID:         convertToLPFeesPerShare(100, 300),
							*token1ID:         convertToLPFeesPerShare(200, 300),
							common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
							common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
						},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount:                300,
								tradingFees:           map[common.Hash]uint64{},
								lastLPFeesPerShare:    map[common.Hash]*big.Int{},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 100, 0, 100, 200,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(400), 20000,
					),
					lpFeesPerShare: map[common.Hash]*big.Int{
						*token0ID:         convertToLPFeesPerShare(100, 300),
						*token1ID:         convertToLPFeesPerShare(200, 300),
						common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
						common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
					},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount: 100,
							tradingFees: map[common.Hash]uint64{
								*token0ID:         99,
								*token1ID:         199,
								common.PRVCoinID:  9,
								common.PDEXCoinID: 19,
							},
							lastLPFeesPerShare: map[common.Hash]*big.Int{
								*token0ID:         convertToLPFeesPerShare(100, 300),
								*token1ID:         convertToLPFeesPerShare(200, 300),
								common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
								common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
							},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			wantErr: false,
		},
		{
			name:   "Multiple LPs same block",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    [][]string{acceptWithdrawLiquidityInst2, acceptWithdrawLiquidityInst3},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 0, 100, 600,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare: map[common.Hash]*big.Int{
							*token0ID:         convertToLPFeesPerShare(100, 300),
							*token1ID:         convertToLPFeesPerShare(200, 300),
							common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
							common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
						},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount:                100,
								tradingFees:           map[common.Hash]uint64{},
								lastLPFeesPerShare:    map[common.Hash]*big.Int{},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
							nftID1: &Share{
								amount:                200,
								tradingFees:           map[common.Hash]uint64{},
								lastLPFeesPerShare:    map[common.Hash]*big.Int{},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						lmLockedShare: map[string]map[uint64]uint64{},
						orderbook:     Orderbook{[]*Order{}},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 150, 0, 100, 300,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(600), 20000,
					),
					lpFeesPerShare: map[common.Hash]*big.Int{
						*token0ID:         convertToLPFeesPerShare(100, 300),
						*token1ID:         convertToLPFeesPerShare(200, 300),
						common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
						common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
					},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount: 50,
							tradingFees: map[common.Hash]uint64{
								*token0ID:         33,
								*token1ID:         66,
								common.PRVCoinID:  3,
								common.PDEXCoinID: 6,
							},
							lastLPFeesPerShare: map[common.Hash]*big.Int{
								*token0ID:         convertToLPFeesPerShare(100, 300),
								*token1ID:         convertToLPFeesPerShare(200, 300),
								common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
								common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
							},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
						nftID1: &Share{
							amount: 100,
							tradingFees: map[common.Hash]uint64{
								*token0ID:         66,
								*token1ID:         133,
								common.PRVCoinID:  6,
								common.PDEXCoinID: 13,
							},
							lastLPFeesPerShare: map[common.Hash]*big.Int{
								*token0ID:         convertToLPFeesPerShare(100, 300),
								*token1ID:         convertToLPFeesPerShare(200, 300),
								common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
								common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
							},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					lmLockedShare: map[string]map[uint64]uint64{},
					orderbook:     Orderbook{[]*Order{}},
				},
			},
			wantErr: false,
		},
		{
			name:   "Multiple LPs different blocks",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    [][]string{acceptWithdrawLiquidityInst2, acceptWithdrawLiquidityInst3},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 0, 100, 600,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare: map[common.Hash]*big.Int{
							*token0ID:         new(big.Int).Add(convertToLPFeesPerShare(50, 100), convertToLPFeesPerShare(50, 300)),
							*token1ID:         new(big.Int).Add(convertToLPFeesPerShare(100, 100), convertToLPFeesPerShare(100, 300)),
							common.PRVCoinID:  new(big.Int).Add(convertToLPFeesPerShare(5, 100), convertToLPFeesPerShare(5, 300)),
							common.PDEXCoinID: new(big.Int).Add(convertToLPFeesPerShare(10, 100), convertToLPFeesPerShare(10, 300)),
						},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						// share_0 (share = 100): earn 50% fee, share_1 (share = 200) add liquidity later
						// share_0 earn ~ 66% fee, share_1 earn ~ 33% fee
						shares: map[string]*Share{
							nftID: &Share{
								amount:                100,
								tradingFees:           map[common.Hash]uint64{},
								lastLPFeesPerShare:    map[common.Hash]*big.Int{},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
							nftID1: &Share{
								amount:      200,
								tradingFees: map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{
									*token0ID:         convertToLPFeesPerShare(50, 100),
									*token1ID:         convertToLPFeesPerShare(100, 100),
									common.PRVCoinID:  convertToLPFeesPerShare(5, 100),
									common.PDEXCoinID: convertToLPFeesPerShare(10, 100),
								},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						lmLockedShare: map[string]map[uint64]uint64{},
						orderbook:     Orderbook{[]*Order{}},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 150, 0, 100, 300,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(600), 20000,
					),
					lpFeesPerShare: map[common.Hash]*big.Int{
						*token0ID:         new(big.Int).Add(convertToLPFeesPerShare(50, 100), convertToLPFeesPerShare(50, 300)),
						*token1ID:         new(big.Int).Add(convertToLPFeesPerShare(100, 100), convertToLPFeesPerShare(100, 300)),
						common.PRVCoinID:  new(big.Int).Add(convertToLPFeesPerShare(5, 100), convertToLPFeesPerShare(5, 300)),
						common.PDEXCoinID: new(big.Int).Add(convertToLPFeesPerShare(10, 100), convertToLPFeesPerShare(10, 300)),
					},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount: 50,
							tradingFees: map[common.Hash]uint64{
								*token0ID:         66,
								*token1ID:         133,
								common.PRVCoinID:  6,
								common.PDEXCoinID: 13,
							},
							lastLPFeesPerShare: map[common.Hash]*big.Int{
								*token0ID:         new(big.Int).Add(convertToLPFeesPerShare(50, 100), convertToLPFeesPerShare(50, 300)),
								*token1ID:         new(big.Int).Add(convertToLPFeesPerShare(100, 100), convertToLPFeesPerShare(100, 300)),
								common.PRVCoinID:  new(big.Int).Add(convertToLPFeesPerShare(5, 100), convertToLPFeesPerShare(5, 300)),
								common.PDEXCoinID: new(big.Int).Add(convertToLPFeesPerShare(10, 100), convertToLPFeesPerShare(10, 300)),
							},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
						nftID1: &Share{
							amount: 100,
							tradingFees: map[common.Hash]uint64{
								*token0ID:         33,
								*token1ID:         66,
								common.PRVCoinID:  3,
								common.PDEXCoinID: 6,
							},
							lastLPFeesPerShare: map[common.Hash]*big.Int{
								*token0ID:         new(big.Int).Add(convertToLPFeesPerShare(50, 100), convertToLPFeesPerShare(50, 300)),
								*token1ID:         new(big.Int).Add(convertToLPFeesPerShare(100, 100), convertToLPFeesPerShare(100, 300)),
								common.PRVCoinID:  new(big.Int).Add(convertToLPFeesPerShare(5, 100), convertToLPFeesPerShare(5, 300)),
								common.PDEXCoinID: new(big.Int).Add(convertToLPFeesPerShare(10, 100), convertToLPFeesPerShare(10, 300)),
							},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					lmLockedShare: map[string]map[uint64]uint64{},
					orderbook:     Orderbook{[]*Order{}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			poolPairs := tt.args.poolPairs
			for _, inst := range tt.args.inst {
				sp.withdrawTxCache = map[string]uint64{}
				poolPairs, _, err = sp.acceptWithdrawLiquidity(tt.args.stateDB, inst, poolPairs, tt.args.beaconHeight, tt.args.lmLockedBlocks)
				if err != nil {
					if !tt.wantErr {
						t.Errorf("stateProcessorV2.acceptWithdrawLiquidity() error = %v", err)
					}
					return
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.acceptWithdrawLiquidity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(poolPairs, tt.want) {
				t.Errorf("stateProcessorV2.acceptWithdrawLiquidity() = %v, want %v",
					getStrPoolPairState(poolPairs[poolPairID]),
					getStrPoolPairState(tt.want[poolPairID]),
				)
			}
		})
	}
}

func Test_stateProcessorV2_unstaking(t *testing.T) {
	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	txReqID, err := common.Hash{}.NewHashFromStr("1111122222")
	assert.Nil(t, err)
	nftHash1, err := common.Hash{}.NewHashFromStr(nftID1)
	assert.Nil(t, err)

	rejectInst, err := instruction.NewRejectUnstakingWithValue(*txReqID, 1, "", nil, nil).StringSlice()
	assert.Nil(t, err)
	acceptInst, err := instruction.NewAcceptUnstakingWithValue(
		common.PRVCoinID, 50, utils.EmptyString, *txReqID, 1, metadataPdexv3.AccessOption{
			NftID: nftHash1,
		}, nil,
	).StringSlice()
	assert.Nil(t, err)

	type fields struct {
		pairHashCache      map[string]common.Hash
		withdrawTxCache    map[string]uint64
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB           *statedb.StateDB
		inst              []string
		nftIDs            map[string]uint64
		stakingPoolStates map[string]*StakingPoolState
		beaconHeight      uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]*StakingPoolState
		want1   *v2.UnstakingStatus
		wantErr bool
	}{
		{
			name: "Valid reject inst",
			fields: fields{
				stateProcessorBase: stateProcessorBase{},
			},
			args: args{
				stateDB: sDB,
				inst:    rejectInst,
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 150,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity: 150,
							},
						},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*StakingPoolState{
				common.PRVIDStr: &StakingPoolState{
					liquidity: 150,
					stakers: map[string]*Staker{
						nftID1: &Staker{
							liquidity: 150,
						},
					},
				},
			},
			want1: &v2.UnstakingStatus{
				Status: common.Pdexv3RejectStatus,
			},
			wantErr: false,
		},
		{
			name: "Valid accept inst",
			fields: fields{
				stateProcessorBase: stateProcessorBase{},
			},
			args: args{
				stateDB: sDB,
				inst:    acceptInst,
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 150,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity: 150,
							},
						},
						rewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*StakingPoolState{
				common.PRVIDStr: &StakingPoolState{
					liquidity: 100,
					stakers: map[string]*Staker{
						nftID1: &Staker{
							liquidity:           100,
							rewards:             map[common.Hash]uint64{},
							lastRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					rewardsPerShare: map[common.Hash]*big.Int{},
				},
			},
			want1: &v2.UnstakingStatus{
				Status:        common.Pdexv3AcceptStatus,
				NftID:         nftID1,
				StakingPoolID: common.PRVIDStr,
				Liquidity:     50,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				pairHashCache:      tt.fields.pairHashCache,
				withdrawTxCache:    tt.fields.withdrawTxCache,
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, err := sp.unstaking(tt.args.stateDB, tt.args.inst, tt.args.nftIDs, tt.args.stakingPoolStates, tt.args.beaconHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.unstaking() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.unstaking() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProcessorV2.unstaking() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_stateProcessorV2_distributeMiningOrderReward(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	mintLOPReward1 := v2utils.BuildDistributeMiningOrderRewardInsts(
		poolPairID, *token0ID, 150000, common.PRVCoinID,
	)
	mintLOPReward2 := v2utils.BuildDistributeMiningOrderRewardInsts(
		poolPairID, *token1ID, 150000, common.PRVCoinID,
	)

	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	type fields struct {
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB   *statedb.StateDB
		inst      [][]string
		poolPairs map[string]*PoolPairState
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]*PoolPairState
		wantErr bool
	}{
		{
			name:   "Single making token",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    mintLOPReward1,
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 0, 100, 600,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare:  map[common.Hash]*big.Int{},
						protocolFees:    map[common.Hash]uint64{},
						stakingPoolFees: map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount:             300,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook:    Orderbook{[]*Order{}},
						orderRewards: map[string]*OrderReward{},
						makingVolume: map[common.Hash]*MakingVolume{
							*token0ID: &MakingVolume{
								volume: map[string]*big.Int{
									nftID:  big.NewInt(0).SetUint64(60),
									nftID1: big.NewInt(0).SetUint64(20),
								},
							},
						},
					},
				},
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 0, 100, 600,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					lpFeesPerShare:  map[common.Hash]*big.Int{},
					protocolFees:    map[common.Hash]uint64{},
					stakingPoolFees: map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount:             300,
							tradingFees:        map[common.Hash]uint64{},
							lastLPFeesPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderbook: Orderbook{[]*Order{}},
					orderRewards: map[string]*OrderReward{
						nftID: {
							uncollectedRewards: map[common.Hash]*OrderRewardDetail{
								common.PRVCoinID: {
									amount: 112500,
								},
							},
						},
						nftID1: {
							uncollectedRewards: map[common.Hash]*OrderRewardDetail{
								common.PRVCoinID: {
									amount: 37500,
								},
							},
						},
					},
					makingVolume: map[common.Hash]*MakingVolume{},
				},
			},
			wantErr: false,
		},
		{
			name:   "Double making tokens",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    [][]string{mintLOPReward1[0], mintLOPReward2[0]},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 0, 100, 600,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare:  map[common.Hash]*big.Int{},
						protocolFees:    map[common.Hash]uint64{},
						stakingPoolFees: map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount:             300,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook:    Orderbook{[]*Order{}},
						orderRewards: map[string]*OrderReward{},
						makingVolume: map[common.Hash]*MakingVolume{
							*token0ID: &MakingVolume{
								volume: map[string]*big.Int{
									nftID:  big.NewInt(0).SetUint64(60),
									nftID1: big.NewInt(0).SetUint64(20),
								},
							},
							*token1ID: &MakingVolume{
								volume: map[string]*big.Int{
									nftID:  big.NewInt(0).SetUint64(50),
									nftID1: big.NewInt(0).SetUint64(100),
								},
							},
						},
					},
				},
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 0, 100, 600,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					lpFeesPerShare:  map[common.Hash]*big.Int{},
					protocolFees:    map[common.Hash]uint64{},
					stakingPoolFees: map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount:             300,
							tradingFees:        map[common.Hash]uint64{},
							lastLPFeesPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderbook: Orderbook{[]*Order{}},
					orderRewards: map[string]*OrderReward{
						nftID: {
							uncollectedRewards: map[common.Hash]*OrderRewardDetail{
								common.PRVCoinID: {
									amount: 162500,
								},
							},
						},
						nftID1: {
							uncollectedRewards: map[common.Hash]*OrderRewardDetail{
								common.PRVCoinID: {
									amount: 137500,
								},
							},
						},
					},
					makingVolume: map[common.Hash]*MakingVolume{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			poolPairs := tt.args.poolPairs
			for _, inst := range tt.args.inst {
				sp.withdrawTxCache = map[string]uint64{}
				poolPairs, err = sp.distributeMiningOrderReward(tt.args.stateDB, inst, poolPairs)
				if err != nil {
					if !tt.wantErr {
						t.Errorf("stateProcessorV2.distributeMiningOrderReward() error = %v", err)
					}
					return
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.distributeMiningOrderReward() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(poolPairs, tt.want) {
				t.Errorf("stateProcessorV2.distributeMiningOrderReward() = %v, want %v",
					getStrPoolPairState(poolPairs[poolPairID]),
					getStrPoolPairState(tt.want[poolPairID]),
				)
			}
		})
	}
}
