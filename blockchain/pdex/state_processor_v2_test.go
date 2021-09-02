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

func Test_stateProcessorV2_addLiquidity(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	firstTxHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)

	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
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

	type fields struct {
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB                     *statedb.StateDB
		inst                        []string
		beaconHeight                uint64
		poolPairs                   map[string]*PoolPairState
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		nftIDs                      map[string]bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]*PoolPairState
		want1   map[string]rawdbv2.Pdexv3Contribution
		want2   map[string]rawdbv2.Pdexv3Contribution
		want3   map[string]bool
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
				poolPairs:                   map[string]*PoolPairState{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				nftIDs:                      map[string]bool{},
			},
			want: map[string]*PoolPairState{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					"", validOTAReceiver0,
					*token0ID, *firstTxHash, common.Hash{}, 100, 20000, 1,
				),
			},
			want2:   map[string]rawdbv2.Pdexv3Contribution{},
			want3:   map[string]bool{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, got2, err := sp.addLiquidity(
				tt.args.stateDB, tt.args.inst,
				tt.args.beaconHeight, tt.args.poolPairs,
				tt.args.waitingContributions, tt.args.deletedWaitingContributions,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.addLiquidity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.addLiquidity() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProcessorV2.addLiquidity() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("stateProcessorV2.addLiquidity() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

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
		token0ID.String(), nftID, 100, 20000,
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
		),
		"pair_hash")
	waitingContributionInst := instruction.NewWaitingAddLiquidityWithValue(*waitingContributionStateDB)
	waitingContributionInstBytes, err := json.Marshal(waitingContributionInst)
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
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
			},
			want: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					"", validOTAReceiver0,
					*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
				),
			},
			want1: &v2utils.ContributionStatus{
				Token0ID:                token0ID.String(),
				Token0ContributedAmount: 100,
				Status:                  common.PDEContributionWaitingChainStatus,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, err := sp.waitingContribution(tt.args.stateDB, tt.args.inst, tt.args.waitingContributions, tt.args.deletedWaitingContributions)
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
		token0ID.String(), nftID, 200, 20000,
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
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
			},
			want: map[string]rawdbv2.Pdexv3Contribution{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					"", validOTAReceiver0,
					*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
				),
			},
			want2: &v2utils.ContributionStatus{
				Status: common.PDEContributionRefundChainStatus,
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
		token1ID.String(), nftID, 400, 20000,
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
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				nftIDs:                      map[string]uint64{},
			},
			want: map[string]rawdbv2.Pdexv3Contribution{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					"", validOTAReceiver0,
					*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
				),
			},
			want2: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 200, 100, 400,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(800), 20000,
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:                  200,
							tradingFees:             map[string]uint64{},
							lastUpdatedBeaconHeight: 11,
						},
					},
					orderbook: Orderbook{[]*Order{}},
				},
			},
			want3: &v2utils.ContributionStatus{
				Status: common.PDEContributionMatchedChainStatus,
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
		token1ID.String(), nftID, 200, 20000,
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
		),
		"pair_hash")
	matchAndReturnContritubtion0Inst := instruction.NewMatchAndReturnAddLiquidityWithValue(
		*matchAndReturnContribution0State, 100, 0, 200, 0, *token1ID)
	matchAndReturnContritubtion0InstBytes, err := json.Marshal(matchAndReturnContritubtion0Inst)
	//

	type fields struct {
		pairHashCache      map[string]string
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
				pairHashCache:      map[string]string{},
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
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 200, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:                  200,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 10,
							},
						},
					},
				},
				nftIDs: map[string]uint64{},
			},
			want: map[string]rawdbv2.Pdexv3Contribution{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					poolPairID, validOTAReceiver0,
					*token0ID, *secondTxHash, *nftHash, 100, 20000, 1,
				),
			},
			want2: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:                  300,
							tradingFees:             map[string]uint64{},
							lastUpdatedBeaconHeight: 11,
						},
					},
				},
			},
			want3:   &v2utils.ContributionStatus{},
			wantErr: false,
		},
		{
			name: "Second Instruction",
			fields: fields{
				pairHashCache: map[string]string{
					"pair_hash": thirdTxHash.String(),
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
					),
				},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:                  300,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 11,
							},
						},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: map[string]rawdbv2.Pdexv3Contribution{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					poolPairID, validOTAReceiver0,
					*token0ID, *secondTxHash, *nftHash, 100, 20000, 1,
				),
			},
			want2: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:                  300,
							tradingFees:             map[string]uint64{},
							lastUpdatedBeaconHeight: 11,
						},
					},
				},
			},
			want3: &v2utils.ContributionStatus{
				Token0ID:                token0ID.String(),
				Token0ContributedAmount: 50,
				Status:                  common.PDEContributionMatchedNReturnedChainStatus,
				Token1ID:                token1ID.String(),
				Token1ContributedAmount: 200,
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
		*txHash, 1,
	).StringSlice()
	assert.Nil(t, err)
	//

	// invalidPoolPairID insturction
	invalidPoolPairIDInst, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		"123", *nftHash, *token0ID, 50, 100, validOTAReceiver0,
		*txHash, 1,
	).StringSlice()
	assert.Nil(t, err)
	//

	// invalidNftID insturction
	invalidNftIDInst, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, common.PRVCoinID, *token0ID, 50, 100, validOTAReceiver0,
		*txHash, 1,
	).StringSlice()
	assert.Nil(t, err)
	//

	// invalidNftID insturction
	failToDeductShare, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, common.PRVCoinID, *token0ID, 0, 0, validOTAReceiver0,
		*txHash, 1,
	).StringSlice()
	assert.Nil(t, err)
	//

	// valid insturction
	acceptWithdrawLiquidityInst0, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *nftHash, *token0ID, 50, 100, validOTAReceiver0,
		*txHash, 1,
	).StringSlice()
	assert.Nil(t, err)

	// valid insturction
	acceptWithdrawLiquidityInst1, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *nftHash, *token1ID, 200, 100, validOTAReceiver1,
		*txHash, 1,
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
		stateDB      *statedb.StateDB
		inst         []string
		poolPairs    map[string]*PoolPairState
		beaconHeight uint64
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
							*token0ID, *token1ID, 300, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:                  300,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:                  300,
							tradingFees:             map[string]uint64{},
							lastUpdatedBeaconHeight: 11,
						},
					},
					orderbook: Orderbook{[]*Order{}},
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
							*token0ID, *token1ID, 300, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:                  300,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:                  300,
							tradingFees:             map[string]uint64{},
							lastUpdatedBeaconHeight: 11,
						},
					},
					orderbook: Orderbook{[]*Order{}},
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid nftID",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    invalidNftIDInst,
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:                  300,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:                  300,
							tradingFees:             map[string]uint64{},
							lastUpdatedBeaconHeight: 11,
						},
					},
					orderbook: Orderbook{[]*Order{}},
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid nftID",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    failToDeductShare,
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:                  300,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:                  300,
							tradingFees:             map[string]uint64{},
							lastUpdatedBeaconHeight: 11,
						},
					},
					orderbook: Orderbook{[]*Order{}},
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
							*token0ID, *token1ID, 300, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:                  300,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 100, 600,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:                  300,
							tradingFees:             map[string]uint64{},
							lastUpdatedBeaconHeight: 11,
						},
					},
					orderbook: Orderbook{[]*Order{}},
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
							*token0ID, *token1ID, 300, 100, 600,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:                  300,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 11,
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 200, 100, 400,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(800), 20000,
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:                  200,
							tradingFees:             map[string]uint64{},
							lastUpdatedBeaconHeight: 20,
						},
					},
					orderbook: Orderbook{[]*Order{}},
				},
			},
			want1: &v2utils.WithdrawStatus{
				Status:       common.PDEWithdrawalAcceptedChainStatus,
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
			got, got1, err := sp.acceptWithdrawLiquidity(tt.args.stateDB, tt.args.inst, tt.args.poolPairs, tt.args.beaconHeight)
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
	inst, err := instruction.NewRejectWithdrawLiquidityWithValue(*nftHash, 1).StringSlice()
	assert.Nil(t, err)

	type fields struct {
		pairHashCache      map[string]string
		withdrawTxCache    map[string]uint64
		stateProcessorBase stateProcessorBase
	}
	type args struct {
		stateDB *statedb.StateDB
		inst    []string
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
				stateDB: sDB,
				inst:    inst,
			},
			want: &v2utils.WithdrawStatus{
				Status: common.PDEWithdrawalRejectedChainStatus,
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
			got, err := sp.rejectWithdrawLiquidity(tt.args.stateDB, tt.args.inst)
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
		pairHashCache      map[string]string
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
				Status:      common.Pdexv3RejectUserMintNftStatus,
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
				Status:      common.Pdexv3AcceptUserMintNftStatus,
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
			got, got1, err := sp.userMintNft(tt.args.stateDB, tt.args.inst, tt.args.nftIDs)
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
	acceptInst, err := instruction.NewAcceptStakingWtihValue(
		*nftHash1, common.PRVCoinID, *txReqID, 1, 100,
	).StringSlice()
	assert.Nil(t, err)

	type fields struct {
		pairHashCache      map[string]string
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
				Status:        common.Pdexv3RejectStakingStatus,
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
						liquidity: 0,
						stakers:   map[string]*Staker{},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*StakingPoolState{
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
			want1: &v2.StakingStatus{
				Status:        common.Pdexv3AcceptStakingStatus,
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

func Test_stateProcessorV2_unstaking(t *testing.T) {
	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	txReqID, err := common.Hash{}.NewHashFromStr("1111122222")
	assert.Nil(t, err)
	nftHash1, err := common.Hash{}.NewHashFromStr(nftID1)
	assert.Nil(t, err)

	rejectInst, err := instruction.NewRejectUnstakingWithValue(*txReqID, 1).StringSlice()
	assert.Nil(t, err)
	acceptInst, err := instruction.NewAcceptUnstakingWithValue(
		common.PRVCoinID, *nftHash1, 50, validOTAReceiver0, *txReqID, 1,
	).StringSlice()
	assert.Nil(t, err)

	type fields struct {
		pairHashCache      map[string]string
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
								liquidity:               150,
								lastUpdatedBeaconHeight: 15,
								rewards:                 map[string]uint64{},
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
							liquidity:               150,
							lastUpdatedBeaconHeight: 15,
							rewards:                 map[string]uint64{},
						},
					},
				},
			},
			want1: &v2.UnstakingStatus{
				Status: common.Pdexv3RejectUnstakingStatus,
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
								liquidity:               150,
								lastUpdatedBeaconHeight: 15,
								rewards:                 map[string]uint64{},
							},
						},
					},
				},
				beaconHeight: 20,
			},
			want: map[string]*StakingPoolState{
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
			want1: &v2.UnstakingStatus{
				Status:        common.Pdexv3AcceptUnstakingStatus,
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
