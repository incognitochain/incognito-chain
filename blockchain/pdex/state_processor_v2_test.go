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
	"github.com/incognitochain/incognito-chain/metadata"
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
		validOTAReceiver0, validOTAReceiver1,
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
			"", validOTAReceiver0, validOTAReceiver1,
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
					"", validOTAReceiver0, validOTAReceiver1,
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
			got, got1, got2, got3, err := sp.addLiquidity(
				tt.args.stateDB, tt.args.inst,
				tt.args.beaconHeight, tt.args.poolPairs,
				tt.args.waitingContributions, tt.args.deletedWaitingContributions,
				tt.args.nftIDs,
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
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("stateProcessorV2.addLiquidity() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}

func Test_stateProcessorV2_waitingContribution(t *testing.T) {
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
		validOTAReceiver0, validOTAReceiver1,
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
			"", validOTAReceiver0, validOTAReceiver1,
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
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]rawdbv2.Pdexv3Contribution
		want1   *metadata.PDEContributionStatus
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
					"", validOTAReceiver0, validOTAReceiver1,
					*token0ID, *firstTxHash, common.Hash{}, 100, 20000, 1,
				),
			},
			want1: &metadata.PDEContributionStatus{
				TokenID1Str:        token0ID.String(),
				Contributed1Amount: 100,
				Status:             byte(common.PDEContributionWaitingStatus),
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

	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	// return contribution tx by sanity
	refundContributionSanityMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		token0ID.String(), utils.EmptyString, 200, 20000,
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
			"", validOTAReceiver0, validOTAReceiver1,
			*token0ID, *firstTxHash, common.Hash{}, 100, 20000, 1,
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
		want2   *metadata.PDEContributionStatus
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
						"", validOTAReceiver0, validOTAReceiver1,
						*token0ID, *firstTxHash, common.Hash{}, 100, 20000, 1,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
			},
			want: map[string]rawdbv2.Pdexv3Contribution{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					"", validOTAReceiver0, validOTAReceiver1,
					*token0ID, *firstTxHash, common.Hash{}, 100, 20000, 1,
				),
			},
			want2: &metadata.PDEContributionStatus{
				Status: byte(common.PDEContributionRefundStatus),
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
		validOTAReceiver0, validOTAReceiver1,
		token1ID.String(), utils.EmptyString, 400, 20000,
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
			"", validOTAReceiver0, validOTAReceiver1,
			*token1ID, *secondTxHash, common.Hash{}, 400, 20000, 1,
		),
		"pair_hash")
	matchContributionInst := instruction.NewMatchAddLiquidityWithValue(*matchContributionState, poolPairID, *nftHash)
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
		nftIDs                      map[string]bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]rawdbv2.Pdexv3Contribution
		want1   map[string]rawdbv2.Pdexv3Contribution
		want2   map[string]*PoolPairState
		want3   map[string]bool
		want4   *metadata.PDEContributionStatus
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
						"", validOTAReceiver0, validOTAReceiver1,
						*token0ID, *firstTxHash, common.Hash{}, 100, 20000, 1,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				nftIDs:                      map[string]bool{},
			},
			want: map[string]rawdbv2.Pdexv3Contribution{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					"", validOTAReceiver0, validOTAReceiver1,
					*token0ID, *firstTxHash, common.Hash{}, 100, 20000, 1,
				),
			},
			want2: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 200, 100, 400,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(800), 20000,
						map[common.Hash]*big.Int{},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:             200,
							tradingFees:        map[common.Hash]uint64{},
							lastLPFeesPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderbook: Orderbook{[]*Order{}},
				},
			},
			want3: map[string]bool{
				nftID: true,
			},
			want4: &metadata.PDEContributionStatus{
				Status: byte(common.PDEContributionAcceptedStatus),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, got2, got3, got4, err := sp.matchContribution(
				tt.args.stateDB, tt.args.inst,
				tt.args.beaconHeight, tt.args.waitingContributions,
				tt.args.deletedWaitingContributions, tt.args.poolPairs,
				tt.args.nftIDs,
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
			if !reflect.DeepEqual(got4, tt.want4) {
				t.Errorf("stateProcessorV2.matchContribution() got4 = %v, want %v", got4, tt.want4)
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
		validOTAReceiver0, validOTAReceiver1,
		token1ID.String(), utils.EmptyString, 200, 20000,
	)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)

	matchAndReturnContributionTx := &metadataMocks.Transaction{}
	matchAndReturnContributionTx.On("GetMetadata").Return(matchAndReturnContributionMetaData)
	matchAndReturnContributionTx.On("GetValidationEnv").Return(valEnv)
	matchAndReturnContributionTx.On("Hash").Return(thirdTxHash)

	matchAndReturnContribution0State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0, validOTAReceiver1,
			*token0ID, *thirdTxHash, common.Hash{}, 50, 20000, 1,
		),
		"pair_hash")
	matchAndReturnContritubtion0Inst := instruction.NewMatchAndReturnAddLiquidityWithValue(
		*matchAndReturnContribution0State, 100, 0, 200, 0, *token1ID, *nftHash)
	matchAndReturnContritubtion0InstBytes, err := json.Marshal(matchAndReturnContritubtion0Inst)
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
		nftIDs                      map[string]bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]rawdbv2.Pdexv3Contribution
		want1   map[string]rawdbv2.Pdexv3Contribution
		want2   map[string]*PoolPairState
		want3   map[string]bool
		want4   *metadata.PDEContributionStatus
		wantErr bool
	}{
		{
			name: "First Instruction",
			fields: fields{
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
						poolPairID, validOTAReceiver0, validOTAReceiver1,
						*token0ID, *secondTxHash, common.Hash{}, 100, 20000, 1,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 200, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
							map[common.Hash]*big.Int{},
							map[common.Hash]uint64{}, map[common.Hash]uint64{},
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:             200,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
					},
				},
				nftIDs: map[string]bool{},
			},
			want: map[string]rawdbv2.Pdexv3Contribution{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					poolPairID, validOTAReceiver0, validOTAReceiver1,
					*token0ID, *secondTxHash, common.Hash{}, 100, 20000, 1,
				),
			},
			want2: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
						map[common.Hash]*big.Int{},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:             300,
							tradingFees:        map[common.Hash]uint64{},
							lastLPFeesPerShare: map[common.Hash]*big.Int{},
						},
					},
				},
			},
			want3: map[string]bool{
				nftID: true,
			},
			want4:   &metadata.PDEContributionStatus{},
			wantErr: false,
		},
		{
			name: "Second Instruction",
			fields: fields{
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
						poolPairID, validOTAReceiver0, validOTAReceiver1,
						*token0ID, *secondTxHash, common.Hash{}, 100, 20000, 1,
					),
				},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
							map[common.Hash]*big.Int{},
							map[common.Hash]uint64{}, map[common.Hash]uint64{},
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:             300,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
					},
				},
				nftIDs: map[string]bool{
					nftID: true,
				},
			},
			want: map[string]rawdbv2.Pdexv3Contribution{},
			want1: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					poolPairID, validOTAReceiver0, validOTAReceiver1,
					*token0ID, *secondTxHash, common.Hash{}, 100, 20000, 1,
				),
			},
			want2: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
						map[common.Hash]*big.Int{},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:             300,
							tradingFees:        map[common.Hash]uint64{},
							lastLPFeesPerShare: map[common.Hash]*big.Int{},
						},
					},
				},
			},
			want3: map[string]bool{
				nftID: true,
			},
			want4: &metadata.PDEContributionStatus{
				TokenID1Str:        token0ID.String(),
				Contributed1Amount: 50,
				Status:             byte(common.PDEContributionMatchedNReturnedStatus),
				TokenID2Str:        token1ID.String(),
				Contributed2Amount: 200,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProcessorV2{
				stateProcessorBase: tt.fields.stateProcessorBase,
			}
			got, got1, got2, got3, got4, err := sp.matchAndReturnContribution(
				tt.args.stateDB, tt.args.inst,
				tt.args.beaconHeight, tt.args.waitingContributions,
				tt.args.deletedWaitingContributions, tt.args.poolPairs,
				tt.args.nftIDs,
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
			if !reflect.DeepEqual(got4, tt.want4) {
				t.Errorf("stateProcessorV2.matchAndReturnContribution() got4 = %v, want %v", got3, tt.want4)
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
		want    map[string]*PoolPairState
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
							map[common.Hash]*big.Int{},
							map[common.Hash]uint64{}, map[common.Hash]uint64{},
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:             300,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
						map[common.Hash]*big.Int{},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:             300,
							tradingFees:        map[common.Hash]uint64{},
							lastLPFeesPerShare: map[common.Hash]*big.Int{},
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
							map[common.Hash]*big.Int{},
							map[common.Hash]uint64{}, map[common.Hash]uint64{},
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:             300,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
						map[common.Hash]*big.Int{},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:             300,
							tradingFees:        map[common.Hash]uint64{},
							lastLPFeesPerShare: map[common.Hash]*big.Int{},
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
							map[common.Hash]*big.Int{},
							map[common.Hash]uint64{}, map[common.Hash]uint64{},
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:             300,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
						map[common.Hash]*big.Int{},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:             300,
							tradingFees:        map[common.Hash]uint64{},
							lastLPFeesPerShare: map[common.Hash]*big.Int{},
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
							map[common.Hash]*big.Int{},
							map[common.Hash]uint64{}, map[common.Hash]uint64{},
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:             300,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
						map[common.Hash]*big.Int{},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:             300,
							tradingFees:        map[common.Hash]uint64{},
							lastLPFeesPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderbook: Orderbook{[]*Order{}},
				},
			},
			wantErr: true,
		},
		{
			name:   "Valid Input - Token 0",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    acceptWithdrawLiquidityInst0,
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
							map[common.Hash]*big.Int{},
							map[common.Hash]uint64{}, map[common.Hash]uint64{},
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:             300,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 100, 600,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(1200), 20000,
						map[common.Hash]*big.Int{},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:             300,
							tradingFees:        map[common.Hash]uint64{},
							lastLPFeesPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderbook: Orderbook{[]*Order{}},
				},
			},
			wantErr: false,
		},
		{
			name:   "Valid Input - Token 1",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    acceptWithdrawLiquidityInst1,
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 100, 600,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(1200), 20000,
							map[common.Hash]*big.Int{},
							map[common.Hash]uint64{}, map[common.Hash]uint64{},
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:             300,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 200, 100, 400,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(800), 20000,
						map[common.Hash]*big.Int{},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:             200,
							tradingFees:        map[common.Hash]uint64{},
							lastLPFeesPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderbook: Orderbook{[]*Order{}},
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
			got, err := sp.acceptWithdrawLiquidity(tt.args.stateDB, tt.args.inst, tt.args.poolPairs)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProcessorV2.acceptWithdrawLiquidity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProcessorV2.acceptWithdrawLiquidity() = %v, want %v", got, tt.want)
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
		poolPairID, *nftHash, *token1ID, 200, 100, validOTAReceiver1,
		*txHash, 1,
	).StringSlice()
	assert.Nil(t, err)

	// valid instruction
	acceptWithdrawLiquidityInst2, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *nftHash, *token1ID, 100, 50, validOTAReceiver1,
		*txHash, 1,
	).StringSlice()
	acceptWithdrawLiquidityInst3, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *nft1Hash, *token1ID, 200, 100, validOTAReceiver1,
		*txHash, 1,
	).StringSlice()

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
			name:   "Single withdrawal",
			fields: fields{},
			args: args{
				stateDB: sDB,
				inst:    [][]string{acceptWithdrawLiquidityInst1},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 100, 600,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(1200), 20000,
							map[common.Hash]*big.Int{
								*token0ID:         convertToLPFeesPerShare(100, 300),
								*token1ID:         convertToLPFeesPerShare(200, 300),
								common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
								common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
							},
							map[common.Hash]uint64{}, map[common.Hash]uint64{},
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:             300,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 200, 100, 400,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(800), 20000,
						map[common.Hash]*big.Int{
							*token0ID:         convertToLPFeesPerShare(100, 300),
							*token1ID:         convertToLPFeesPerShare(200, 300),
							common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
							common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
						},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount: 200,
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
						},
					},
					orderbook: Orderbook{[]*Order{}},
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
							*token0ID, *token1ID, 300, 100, 600,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(1200), 20000,
							map[common.Hash]*big.Int{
								*token0ID:         convertToLPFeesPerShare(100, 300),
								*token1ID:         convertToLPFeesPerShare(200, 300),
								common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
								common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
							},
							map[common.Hash]uint64{}, map[common.Hash]uint64{},
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:             300,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 100, 100, 200,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(400), 20000,
						map[common.Hash]*big.Int{
							*token0ID:         convertToLPFeesPerShare(100, 300),
							*token1ID:         convertToLPFeesPerShare(200, 300),
							common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
							common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
						},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
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
						},
					},
					orderbook: Orderbook{[]*Order{}},
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
							*token0ID, *token1ID, 300, 100, 600,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(1200), 20000,
							map[common.Hash]*big.Int{
								*token0ID:         convertToLPFeesPerShare(100, 300),
								*token1ID:         convertToLPFeesPerShare(200, 300),
								common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
								common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
							},
							map[common.Hash]uint64{}, map[common.Hash]uint64{},
						),
						shares: map[string]*Share{
							nftID: &Share{
								amount:             100,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
							nftID1: &Share{
								amount:             200,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 150, 100, 300,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(600), 20000,
						map[common.Hash]*big.Int{
							*token0ID:         convertToLPFeesPerShare(100, 300),
							*token1ID:         convertToLPFeesPerShare(200, 300),
							common.PRVCoinID:  convertToLPFeesPerShare(10, 300),
							common.PDEXCoinID: convertToLPFeesPerShare(20, 300),
						},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
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
						},
					},
					orderbook: Orderbook{[]*Order{}},
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
							*token0ID, *token1ID, 300, 100, 600,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(1200), 20000,
							map[common.Hash]*big.Int{
								*token0ID:         new(big.Int).Add(convertToLPFeesPerShare(50, 100), convertToLPFeesPerShare(50, 300)),
								*token1ID:         new(big.Int).Add(convertToLPFeesPerShare(100, 100), convertToLPFeesPerShare(100, 300)),
								common.PRVCoinID:  new(big.Int).Add(convertToLPFeesPerShare(5, 100), convertToLPFeesPerShare(5, 300)),
								common.PDEXCoinID: new(big.Int).Add(convertToLPFeesPerShare(10, 100), convertToLPFeesPerShare(10, 300)),
							},
							map[common.Hash]uint64{}, map[common.Hash]uint64{},
						),
						// share_0 (share = 100): earn 50% fee, share_1 (share = 200) add liquidity later
						// share_0 earn ~ 66% fee, share_1 earn ~ 33% fee
						shares: map[string]*Share{
							nftID: &Share{
								amount:             100,
								tradingFees:        map[common.Hash]uint64{},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
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
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
			},
			want: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 150, 100, 300,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(600), 20000,
						map[common.Hash]*big.Int{
							*token0ID:         new(big.Int).Add(convertToLPFeesPerShare(50, 100), convertToLPFeesPerShare(50, 300)),
							*token1ID:         new(big.Int).Add(convertToLPFeesPerShare(100, 100), convertToLPFeesPerShare(100, 300)),
							common.PRVCoinID:  new(big.Int).Add(convertToLPFeesPerShare(5, 100), convertToLPFeesPerShare(5, 300)),
							common.PDEXCoinID: new(big.Int).Add(convertToLPFeesPerShare(10, 100), convertToLPFeesPerShare(10, 300)),
						},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
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
						},
					},
					orderbook: Orderbook{[]*Order{}},
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
				poolPairs, err = sp.acceptWithdrawLiquidity(tt.args.stateDB, inst, poolPairs)
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
