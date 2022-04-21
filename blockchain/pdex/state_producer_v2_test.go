package pdex

import (
	"encoding/json"
	"math/big"
	"reflect"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataMocks "github.com/incognitochain/incognito-chain/metadata/common/mocks"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/stretchr/testify/assert"
)

func Test_stateProducerV2_addLiquidity(t *testing.T) {
	initTestParams(t)
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
	thirdTxHash, err := common.Hash{}.NewHashFromStr("bbb")
	assert.Nil(t, err)
	fourthTxHash, err := common.Hash{}.NewHashFromStr("ccc")
	assert.Nil(t, err)
	accessID, err := common.Hash{}.NewHashFromStr("a1cb3f20e8c21c5b0dee89db55395efe54a7935f575f58f95f2d5506c5b45654")
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
	refundContributionSanityTx.On("GetValidationEnv").Return(valEnv)
	refundContributionSanityTx.On("Hash").Return(secondTxHash)
	refundContributionSanityState0 := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0,
			*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	refundContributionSanityInst0 := instruction.NewRefundAddLiquidityWithValue(*refundContributionSanityState0)
	refundContributionSanityInstBytes0, err := json.Marshal(refundContributionSanityInst0)
	refundContributionSanityState1 := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0,
			*token0ID, *secondTxHash, *nftHash, 200, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	refundContributionSanityInst1 := instruction.NewRefundAddLiquidityWithValue(*refundContributionSanityState1)
	refundContributionSanityInstBytes1, err := json.Marshal(refundContributionSanityInst1)
	//

	// match contribution
	matchContributionMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash", validOTAReceiver0,
		token1ID.String(), 400, 20000,
		&metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	)
	assert.Nil(t, err)
	matchContributionTx := &metadataMocks.Transaction{}
	matchContributionTx.On("GetMetadata").Return(matchContributionMetaData)
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

	// match contribution - 2
	matchContribution2MetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0,
		token1ID.String(), 400, 20000,
		&metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	)
	matchContribution2Tx := &metadataMocks.Transaction{}
	matchContribution2Tx.On("GetMetadata").Return(matchContribution2MetaData)
	matchContribution2Tx.On("GetValidationEnv").Return(valEnv)
	matchContribution2Tx.On("Hash").Return(secondTxHash)
	matchContribution2State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0,
			*token1ID, *secondTxHash, *nftHash, 400, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	matchContribution2Inst := instruction.NewMatchAddLiquidityWithValue(*matchContribution2State, poolPairID)
	matchContribution2InstBytes, err := json.Marshal(matchContribution2Inst)
	//

	// refund contributions by amount
	refundContributionAmountMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		poolPairID, "pair_hash",
		validOTAReceiver0,
		token1ID.String(), 0, 20000,
		&metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	)
	refundContributionAmountTx := &metadataMocks.Transaction{}
	refundContributionAmountTx.On("GetMetadata").Return(refundContributionAmountMetaData)
	refundContributionAmountTx.On("GetValidationEnv").Return(valEnv)
	refundContributionAmountTx.On("Hash").Return(fourthTxHash)

	refundContributionAmount0State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0,
			*token0ID, *thirdTxHash, *nftHash, 50, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	refundContributionAmount0Inst := instruction.NewRefundAddLiquidityWithValue(*refundContributionAmount0State)
	refundContributionAmount0InstBytes, err := json.Marshal(refundContributionAmount0Inst)
	refundContributionAmount1State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0,
			*token1ID, *fourthTxHash, *nftHash, 0, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	refundContributionAmount1Inst := instruction.NewRefundAddLiquidityWithValue(*refundContributionAmount1State)
	refundContributionAmount1InstBytes, err := json.Marshal(refundContributionAmount1Inst)
	//

	// match and return contribution
	matchAndReturnContributionMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		poolPairID, "pair_hash",
		validOTAReceiver0,
		token1ID.String(), 200, 20000,
		&metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	)
	matchAndReturnContributionTx := &metadataMocks.Transaction{}
	matchAndReturnContributionTx.On("GetMetadata").Return(matchAndReturnContributionMetaData)
	matchAndReturnContributionTx.On("GetValidationEnv").Return(valEnv)
	matchAndReturnContributionTx.On("Hash").Return(fourthTxHash)

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
	matchAndReturnContribution1State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0,
			*token1ID, *fourthTxHash, *nftHash, 200, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	matchAndReturnContritubtion1Inst := instruction.NewMatchAndReturnAddLiquidityWithValue(
		*matchAndReturnContribution1State, 100, 0, 50, 0, *token0ID, nil)
	matchAndReturnContritubtion1InstBytes, err := json.Marshal(matchAndReturnContritubtion1Inst)
	//

	// match and return contribution - 2
	matchAndReturnContribution2MetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		poolPairID, "pair_hash", validOTAReceiver0,
		token1ID.String(), 200, 20000,
		&metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	)
	matchAndReturnContribution2Tx := &metadataMocks.Transaction{}
	matchAndReturnContribution2Tx.On("GetMetadata").Return(matchAndReturnContribution2MetaData)
	matchAndReturnContribution2Tx.On("GetValidationEnv").Return(valEnv)
	matchAndReturnContribution2Tx.On("Hash").Return(fourthTxHash)

	matchAndReturnContribution0State2 := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0,
			*token0ID, *thirdTxHash, *nftHash, 50, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	matchAndReturnContritubtion0Inst2 := instruction.NewMatchAndReturnAddLiquidityWithValue(
		*matchAndReturnContribution0State2, 100, 0, 200, 0, *token1ID, nil)
	matchAndReturnContritubtion0InstBytes2, err := json.Marshal(matchAndReturnContritubtion0Inst2)
	matchAndReturnContribution1State2 := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0,
			*token1ID, *fourthTxHash, *nftHash, 200, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	matchAndReturnContritubtion1Inst2 := instruction.NewMatchAndReturnAddLiquidityWithValue(
		*matchAndReturnContribution1State2, 100, 0, 50, 0, *token0ID, nil)
	matchAndReturnContritubtion1InstBytes2, err := json.Marshal(matchAndReturnContritubtion1Inst2)
	//

	// out of range materials
	outOfRangeMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		poolPairID, "pair_hash", validOTAReceiver0,
		token1ID.String(), 10000000000000000000, 20000,
		&metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	)
	outOfRangeTx := &metadataMocks.Transaction{}
	outOfRangeTx.On("GetMetadata").Return(outOfRangeMetaData)
	outOfRangeTx.On("GetValidationEnv").Return(valEnv)
	outOfRangeTx.On("Hash").Return(fourthTxHash)

	outOfRangeContribution0State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0,
			*token0ID, *thirdTxHash, *nftHash, 10000000000000000000, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	outOfRangeInst0 := instruction.NewRefundAddLiquidityWithValue(*outOfRangeContribution0State)
	outOfRangeInst0Bytes, err := json.Marshal(outOfRangeInst0)
	outOfRangeContribution1State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0,
			*token1ID, *fourthTxHash, *nftHash, 10000000000000000000, 20000, 1,
			nil, nil,
		),
		"pair_hash")
	outOfRangeInst1 := instruction.NewRefundAddLiquidityWithValue(*outOfRangeContribution1State)
	outOfRangeInst1Bytes, err := json.Marshal(outOfRangeInst1)

	// accessOTAContributions
	nullAccessOTAContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0,
		token1ID.String(), 400, 20000,
		&metadataPdexv3.AccessOption{}, map[common.Hash]privacy.OTAReceiver{*token1ID: otaReceiver1},
	)
	assert.Nil(t, err)
	nullAccessOTAContributionTx := &metadataMocks.Transaction{}
	nullAccessOTAContributionTx.On("GetMetadata").Return(nullAccessOTAContributionMetadata)
	nullAccessOTAContributionTx.On("GetValidationEnv").Return(valEnv)
	nullAccessOTAContributionTx.On("Hash").Return(secondTxHash)
	nullAccessOTAContributionStateDB := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0,
			*token1ID, *secondTxHash, common.Hash{}, 400, 20000, 1,
			nil, map[common.Hash]string{
				*token1ID: validOTAReceiver1,
			},
		),
		"pair_hash")
	waitingNullAccessOTAContributionInst := instruction.NewWaitingAddLiquidityWithValue(*nullAccessOTAContributionStateDB)
	waitingNullAccessOTAContributionInstBytes, err := json.Marshal(waitingNullAccessOTAContributionInst)

	// accessOTAContributions
	accessOTAContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0,
		token0ID.String(), 100, 20000,
		&metadataPdexv3.AccessOption{}, map[common.Hash]privacy.OTAReceiver{*token0ID: otaReceiver0, common.PdexAccessCoinID: otaReceiver0},
	)
	assert.Nil(t, err)
	accessOTAContributionTx := &metadataMocks.Transaction{}
	accessOTAContributionTx.On("GetMetadata").Return(accessOTAContributionMetadata)
	accessOTAContributionTx.On("GetValidationEnv").Return(valEnv)
	accessOTAContributionTx.On("Hash").Return(firstTxHash)
	accessOTAContributionStateDB := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0,
			*token0ID, *firstTxHash, *accessID, 100, 20000, 1,
			otaReceiver0.PublicKey.ToBytesS(), map[common.Hash]string{
				*token0ID:               validOTAReceiver0,
				common.PdexAccessCoinID: validOTAReceiver0,
			},
		),
		"pair_hash")

	matchAccessOTAContributionInst := instruction.NewMatchAddLiquidityWithValue(*accessOTAContributionStateDB, accessOTANewPoolPairID)
	matchAccessOTAContributionInstBytes, err := json.Marshal(matchAccessOTAContributionInst)

	mintAccessCoinInst := instruction.NewMintAccessTokenWithValue(validOTAReceiver0, 1, *secondTxHash)
	mintAccessCoinInstStr, err := mintAccessCoinInst.StringSlice(strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta))
	assert.Nil(t, err)

	/*matchedNReturnedAccessOTAContributionInst := instruction.NewMatchAndReturnAddLiquidityWithValue(*accessOTAContributionStateDB)*/
	/*matchedNReturnedAccessOTAContributionInstBytes, err := json.Marshal(matchedNReturnedAccessOTAContributionInst)*/

	// end of data

	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		txs                  []metadata.Transaction
		beaconHeight         uint64
		poolPairs            map[string]*PoolPairState
		waitingContributions map[string]rawdbv2.Pdexv3Contribution
		nftIDs               map[string]uint64
		params               *Params
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		want1   map[string]*PoolPairState
		want2   map[string]rawdbv2.Pdexv3Contribution
		want3   map[string]bool
		wantErr bool
	}{
		{
			name: "Add to waitingContributions list",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{
					contributionTx,
				},
				beaconHeight:         10,
				poolPairs:            map[string]*PoolPairState{},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				params: NewParams(),
			},
			want: [][]string{
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionWaitingChainStatus,
					string(waitingContributionInstBytes),
				},
			},
			want1: map[string]*PoolPairState{},
			want2: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					"", validOTAReceiver0,
					*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
					nil, nil,
				),
			},
			wantErr: false,
		},
		{
			name: "Add to waitingContributions list - accessOTA",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{
					nullAccessOTAContributionTx,
				},
				beaconHeight:         10,
				poolPairs:            map[string]*PoolPairState{},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				nftIDs:               map[string]uint64{},
				params:               NewParams(),
			},
			want: [][]string{
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionWaitingChainStatus,
					string(waitingNullAccessOTAContributionInstBytes),
				},
			},
			want1: map[string]*PoolPairState{},
			want2: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					"", validOTAReceiver0,
					*token1ID, *secondTxHash, common.Hash{}, 400, 20000, 1,
					nil, map[common.Hash]string{
						*token1ID: validOTAReceiver1,
					},
				),
			},
			wantErr: false,
		},
		{
			name: "Matched contributions - accessOTA",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{
					accessOTAContributionTx,
				},
				beaconHeight: 10,
				poolPairs:    map[string]*PoolPairState{},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						"", validOTAReceiver0,
						*token1ID, *secondTxHash, common.Hash{}, 400, 20000, 1,
						nil, map[common.Hash]string{
							*token1ID: validOTAReceiver1,
						},
					),
				},
				nftIDs: map[string]uint64{},
				params: NewParams(),
			},
			want: [][]string{
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedChainStatus,
					string(matchAccessOTAContributionInstBytes),
				},
				mintAccessCoinInstStr,
			},
			want1: map[string]*PoolPairState{
				accessOTANewPoolPairID: {
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
						accessID.String(): {
							amount:                200,
							tradingFees:           map[common.Hash]uint64{},
							accessOTA:             otaReceiver0.PublicKey.ToBytesS(),
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
			want2:   map[string]rawdbv2.Pdexv3Contribution{},
			wantErr: false,
		},
		{
			name: "refund by invalid sanity data contribution",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{
					refundContributionSanityTx,
				},
				beaconHeight: 11,
				poolPairs:    map[string]*PoolPairState{},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						"", validOTAReceiver0,
						*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
						nil, nil,
					),
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				params: NewParams(),
			},
			want: [][]string{
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionRefundChainStatus,
					string(refundContributionSanityInstBytes0),
				},
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionRefundChainStatus,
					string(refundContributionSanityInstBytes1),
				},
			},
			want1:   map[string]*PoolPairState{},
			want2:   map[string]rawdbv2.Pdexv3Contribution{},
			wantErr: false,
		},
		{
			name: "matched contribution",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{
					matchContributionTx,
				},
				beaconHeight: 11,
				poolPairs:    map[string]*PoolPairState{},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						"", validOTAReceiver0,
						*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
						nil, nil,
					),
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				params: NewParams(),
			},
			want: [][]string{
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedChainStatus,
					string(matchContributionInstBytes),
				},
			},
			want1: map[string]*PoolPairState{
				poolPairID: {
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
						nftID: {
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
			want2: map[string]rawdbv2.Pdexv3Contribution{},
			want3: map[string]bool{
				nftID: true,
			},
			wantErr: false,
		},
		{
			name: "matched contribution - 2",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{
					matchContribution2Tx,
				},
				beaconHeight: 11,
				poolPairs: map[string]*PoolPairState{
					poolPairID + "123": {
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
							nftID: {
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
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						"", validOTAReceiver0,
						*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
						nil, nil,
					),
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				params: NewParams(),
			},
			want: [][]string{
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedChainStatus,
					string(matchContribution2InstBytes),
				},
			},
			want1: map[string]*PoolPairState{
				poolPairID + "123": {
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
						nftID: {
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
				poolPairID: {
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
						nftID: {
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
			want2: map[string]rawdbv2.Pdexv3Contribution{},
			want3: map[string]bool{
				nftID: true,
			},
			wantErr: false,
		},
		{
			name: "refund by contributions amount",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{
					refundContributionAmountTx,
				},
				beaconHeight: 11,
				poolPairs: map[string]*PoolPairState{
					poolPairID: {
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
							nftID: {
								amount:                200,
								tradingFees:           map[common.Hash]uint64{},
								lastLPFeesPerShare:    map[common.Hash]*big.Int{},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						poolPairID, validOTAReceiver0,
						*token0ID, *thirdTxHash, *nftHash, 50, 20000, 1,
						nil, nil,
					),
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				params: NewParams(),
			},
			want: [][]string{
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionRefundChainStatus,
					string(refundContributionAmount0InstBytes),
				},
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionRefundChainStatus,
					string(refundContributionAmount1InstBytes),
				},
			},
			want1: map[string]*PoolPairState{
				poolPairID: {
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
						nftID: {
							amount:                200,
							tradingFees:           map[common.Hash]uint64{},
							lastLPFeesPerShare:    map[common.Hash]*big.Int{},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			want2:   map[string]rawdbv2.Pdexv3Contribution{},
			wantErr: false,
		},
		{
			name: "matched and return contribution",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{
					matchAndReturnContributionTx,
				},
				beaconHeight: 11,
				poolPairs: map[string]*PoolPairState{
					poolPairID: {
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
							nftID: {
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
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						poolPairID, validOTAReceiver0,
						*token0ID, *thirdTxHash, *nftHash, 50, 20000, 1,
						nil, nil,
					),
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				params: NewParams(),
			},
			want: [][]string{
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedNReturnedChainStatus,
					string(matchAndReturnContritubtion0InstBytes),
				},
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedNReturnedChainStatus,
					string(matchAndReturnContritubtion1InstBytes),
				},
			},
			want1: map[string]*PoolPairState{
				poolPairID: {
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
						nftID: {
							amount:                300,
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
			want2: map[string]rawdbv2.Pdexv3Contribution{},
			want3: map[string]bool{
				nftID: true,
			},
			wantErr: false,
		},
		{
			name: "matched and return contribution - 2",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{
					matchAndReturnContribution2Tx,
				},
				beaconHeight: 12,
				poolPairs: map[string]*PoolPairState{
					poolPairID: {
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
							nftID: {
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
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						poolPairID, validOTAReceiver0,
						*token0ID, *thirdTxHash, *nftHash, 50, 20000, 1,
						nil, nil,
					),
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				params: NewParams(),
			},
			want: [][]string{
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedNReturnedChainStatus,
					string(matchAndReturnContritubtion0InstBytes2),
				},
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedNReturnedChainStatus,
					string(matchAndReturnContritubtion1InstBytes2),
				},
			},
			want1: map[string]*PoolPairState{
				poolPairID: {
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
						nftID: {
							amount:                300,
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
			want2: map[string]rawdbv2.Pdexv3Contribution{},
			want3: map[string]bool{
				nftID: true,
			},
			wantErr: false,
		},
		{
			name: "Out of range uint64",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{
					outOfRangeTx,
				},
				beaconHeight: 12,
				poolPairs: map[string]*PoolPairState{
					poolPairID: {
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 10000000000000000000, 0,
							10000000000000000000, 200,
							big.NewInt(0).SetUint64(10000000000000000000),
							big.NewInt(0).SetUint64(10000000000000000000), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: {
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
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						poolPairID, validOTAReceiver0,
						*token0ID, *thirdTxHash, *nftHash, 10000000000000000000, 20000, 1,
						nil, nil,
					),
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				params: NewParams(),
			},
			want: [][]string{
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionRefundChainStatus,
					string(outOfRangeInst0Bytes),
				},
				{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionRefundChainStatus,
					string(outOfRangeInst1Bytes),
				},
			},
			want1: map[string]*PoolPairState{
				poolPairID: {
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 10000000000000000000, 0,
						10000000000000000000, 200,
						big.NewInt(0).SetUint64(10000000000000000000),
						big.NewInt(0).SetUint64(10000000000000000000), 20000,
					),
					lpFeesPerShare:    map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: {
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
			want2: map[string]rawdbv2.Pdexv3Contribution{},
			want3: map[string]bool{
				nftID: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerV2{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			got, got1, got2, err := sp.addLiquidity(
				tt.args.txs, tt.args.beaconHeight,
				tt.args.poolPairs, tt.args.waitingContributions,
				tt.args.nftIDs,
				tt.args.params,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV2.addLiquidity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV2.addLiquidity() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProducerV2.addLiquidity() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("stateProducerV2.addLiquidity() got2 = %v, want %v", got2["pair_hash"], tt.want2["pair_hash"])
			}
		})
	}
}

func Test_stateProducerV2_withdrawLiquidity(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)
	txHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)
	nftHash, err := common.Hash{}.NewHashFromStr(nftID)
	assert.Nil(t, err)

	otaReceivers := map[string]string{
		nftID:                     validOTAReceiver0,
		token0ID.String():         validOTAReceiver1,
		token1ID.String():         validOTAReceiver1,
		common.PRVCoinID.String(): validOTAReceiver0,
	}

	//invalidPoolPairID
	invalidPoolPairIDMetaData := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
		"123", otaReceivers, 100, &metadataPdexv3.AccessOption{
			NftID: nftHash,
		},
	)
	invalidPoolPairIDTx := &metadataMocks.Transaction{}
	invalidPoolPairIDTx.On("GetMetadata").Return(invalidPoolPairIDMetaData)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)
	invalidPoolPairIDTx.On("GetValidationEnv").Return(valEnv)
	invalidPoolPairIDTx.On("Hash").Return(txHash)
	//

	//reject instruction
	tempRejectWithdrawLiquidityInst, err := instruction.NewRejectWithdrawLiquidityWithValue(*txHash, 1, "123", nil, nil).StringSlice()
	assert.Nil(t, err)

	//reject instruction
	rejectWithdrawLiquidityInst, err := instruction.NewRejectWithdrawLiquidityWithValue(*txHash, 1, poolPairID, nil, nil).StringSlice()
	assert.Nil(t, err)
	//

	//mint nft
	mintNftInst, err := instruction.NewMintNftWithValue(*nftHash, validOTAReceiver0, 1, *txHash).
		StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta))
	assert.Nil(t, err)
	//

	//mint prv nft
	mintPrvNftInst, err := instruction.NewMintNftWithValue(common.PRVCoinID, validOTAReceiver0, 1, *txHash).
		StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta))
	assert.Nil(t, err)
	//

	//invalidNftID
	invalidNftIDMetaData := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
		poolPairID, otaReceivers, 100, &metadataPdexv3.AccessOption{
			NftID: &common.PRVCoinID,
		},
	)
	invalidNftIDTx := &metadataMocks.Transaction{}
	invalidNftIDTx.On("GetMetadata").Return(invalidNftIDMetaData)
	invalidNftIDTx.On("GetValidationEnv").Return(valEnv)
	invalidNftIDTx.On("Hash").Return(txHash)
	//

	//deductShareAmountFail
	deductShareAmountFailMetaData := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
		poolPairID, otaReceivers, 0, &metadataPdexv3.AccessOption{
			NftID: nftHash,
		},
	)
	deductShareAmountFailTx := &metadataMocks.Transaction{}
	deductShareAmountFailTx.On("GetMetadata").Return(deductShareAmountFailMetaData)
	deductShareAmountFailTx.On("GetValidationEnv").Return(valEnv)
	deductShareAmountFailTx.On("Hash").Return(txHash)
	//

	//validInput
	validInputMetaData := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
		poolPairID, otaReceivers, 100, &metadataPdexv3.AccessOption{
			NftID: nftHash,
		},
	)
	validInputTx := &metadataMocks.Transaction{}
	validInputTx.On("GetMetadata").Return(validInputMetaData)
	validInputTx.On("GetValidationEnv").Return(valEnv)
	validInputTx.On("Hash").Return(txHash)
	//

	//accept instructions
	acceptWithdrawLiquidityInst0, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *token0ID, 50, 100, validOTAReceiver1,
		*txHash, 1, metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	).StringSlice()
	assert.Nil(t, err)
	acceptWithdrawLiquidityInst1, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *token1ID, 200, 100, validOTAReceiver1,
		*txHash, 1, metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	).StringSlice()
	assert.Nil(t, err)
	//

	temp, ok := big.NewInt(0).SetString("18446744073709551616", 10)
	assert.Equal(t, ok, true)
	assert.NotNil(t, temp)

	//out of range tx
	outOfRangeMetaData := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
		poolPairID, otaReceivers, 48019194174972302, &metadataPdexv3.AccessOption{
			NftID: nftHash,
		},
	)
	outOfRangeTx := &metadataMocks.Transaction{}
	outOfRangeTx.On("GetMetadata").Return(outOfRangeMetaData)
	outOfRangeTx.On("GetValidationEnv").Return(valEnv)
	outOfRangeTx.On("Hash").Return(txHash)
	//

	//accept instructions
	outOfRangeInst0, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *token0ID, 250000000000000, 48019194174972302, validOTAReceiver1,
		*txHash, 1, metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	).StringSlice()
	assert.Nil(t, err)
	outOfRangeInst1, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *token1ID, 9223372036854775808, 48019194174972302, validOTAReceiver1,
		*txHash, 1, metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, nil,
	).StringSlice()
	assert.Nil(t, err)
	//

	invalidOtaReceivers := map[string]string{
		nftID:             validOTAReceiver0,
		token0ID.String(): validOTAReceiver1,
	}

	// invalid otaReceivers tx
	invalidOtaReceiversMetadata := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
		poolPairID, invalidOtaReceivers, 100, &metadataPdexv3.AccessOption{
			NftID: nftHash,
		},
	)
	invalidOtaReceiversTx := &metadataMocks.Transaction{}
	invalidOtaReceiversTx.On("GetMetadata").Return(invalidOtaReceiversMetadata)
	invalidOtaReceiversTx.On("GetValidationEnv").Return(valEnv)
	invalidOtaReceiversTx.On("Hash").Return(txHash)
	//

	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		txs            []metadata.Transaction
		poolPairs      map[string]*PoolPairState
		nftIDs         map[string]uint64
		beaconHeight   uint64
		lmLockedBlocks uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		want1   map[string]*PoolPairState
		wantErr bool
	}{
		{
			name:   "Invalid pool pair id",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{invalidPoolPairIDTx},
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
						orderRewards:  map[string]*OrderReward{},
						makingVolume:  map[common.Hash]*MakingVolume{},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				beaconHeight: 20,
			},
			want: [][]string{tempRejectWithdrawLiquidityInst, mintNftInst},
			want1: map[string]*PoolPairState{
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
					orderRewards:  map[string]*OrderReward{},
					makingVolume:  map[common.Hash]*MakingVolume{},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			wantErr: false,
		},
		{
			name:   "Invalid nftID",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{invalidNftIDTx},
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
						orderRewards:  map[string]*OrderReward{},
						makingVolume:  map[common.Hash]*MakingVolume{},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				beaconHeight: 20,
			},
			want: [][]string{rejectWithdrawLiquidityInst, mintPrvNftInst},
			want1: map[string]*PoolPairState{
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
					orderRewards:  map[string]*OrderReward{},
					makingVolume:  map[common.Hash]*MakingVolume{},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			wantErr: false,
		},
		{
			name:   "Deduct share amount fail",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{deductShareAmountFailTx},
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
						orderRewards:  map[string]*OrderReward{},
						makingVolume:  map[common.Hash]*MakingVolume{},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				beaconHeight: 20,
			},
			want: [][]string{rejectWithdrawLiquidityInst, mintNftInst},
			want1: map[string]*PoolPairState{
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
					orderRewards:  map[string]*OrderReward{},
					makingVolume:  map[common.Hash]*MakingVolume{},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			wantErr: false,
		},
		{
			name:   "Valid Input",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{validInputTx},
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
						orderRewards:  map[string]*OrderReward{},
						makingVolume:  map[common.Hash]*MakingVolume{},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				beaconHeight: 20,
			},
			want: [][]string{acceptWithdrawLiquidityInst0, acceptWithdrawLiquidityInst1, mintNftInst},
			want1: map[string]*PoolPairState{
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
			wantErr: false,
		},
		{
			name:   "Out of range uint64 virtual amount - Valid",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{outOfRangeTx},
				poolPairs: map[string]*PoolPairState{
					poolPairID: {
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 48019194174972302, 0,
							250000000000000, 9223372036854775808,
							big.NewInt(0).SetUint64(500000000000000),
							temp, 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: {
								amount:                48019194174972302,
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
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				beaconHeight: 20,
			},
			want: [][]string{outOfRangeInst0, outOfRangeInst1, mintNftInst},
			want1: map[string]*PoolPairState{
				poolPairID: {
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 0, 0,
						0, 0,
						big.NewInt(0).SetUint64(0),
						big.NewInt(0).SetUint64(0), 20000,
					),
					lpFeesPerShare:    map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares:            map[string]*Share{},
					orderRewards:      map[string]*OrderReward{},
					makingVolume:      map[common.Hash]*MakingVolume{},
					orderbook:         Orderbook{[]*Order{}},
					lmLockedShare:     map[string]map[uint64]uint64{},
				},
			},
			wantErr: false,
		},
		{
			name:   "Invalid ota receivers",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{invalidOtaReceiversTx},
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
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				beaconHeight: 20,
			},
			want: [][]string{rejectWithdrawLiquidityInst, mintNftInst},
			want1: map[string]*PoolPairState{
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerV2{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			got, got1, err := sp.withdrawLiquidity(tt.args.txs, tt.args.poolPairs, tt.args.nftIDs, tt.args.beaconHeight, tt.args.lmLockedBlocks)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV2.withdrawLiquidity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV2.withdrawLiquidity() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProducerV2.withdrawLiquidity() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_stateProducerV2_withdrawLPFee(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)
	txHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)
	nftHash, err := common.Hash{}.NewHashFromStr(nftID)
	assert.Nil(t, err)

	otaReceiver0 := privacy.OTAReceiver{}
	err = otaReceiver0.FromString(validOTAReceiver0)
	assert.Nil(t, err)
	otaReceiver1 := privacy.OTAReceiver{}
	err = otaReceiver1.FromString(validOTAReceiver1)
	assert.Nil(t, err)

	// invalidPoolPairID
	invalidPoolPairIDMetaData, _ := metadataPdexv3.NewPdexv3WithdrawalLPFeeRequest(
		metadataCommon.Pdexv3WithdrawLPFeeRequestMeta,
		"123", metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, map[common.Hash]privacy.OTAReceiver{
			*nftHash: otaReceiver0,
		},
	)

	invalidPoolPairIDTx := &metadataMocks.Transaction{}
	invalidPoolPairIDTx.On("GetMetadata").Return(invalidPoolPairIDMetaData)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)
	invalidPoolPairIDTx.On("GetValidationEnv").Return(valEnv)
	invalidPoolPairIDTx.On("Hash").Return(txHash)

	// invalid pool pair
	rejectPoolPairInst := v2utils.BuildWithdrawLPFeeInsts(
		"123", metadataPdexv3.AccessOption{
			NftID: nftHash,
		},
		map[common.Hash]metadataPdexv3.ReceiverInfo{},
		1, *txHash, metadataPdexv3.RequestRejectedChainStatus, nil,
	)[0]
	assert.Nil(t, err)

	//mint nft
	mintNftInst, err := instruction.NewMintNftWithValue(*nftHash, validOTAReceiver0, 1, *txHash).
		StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawLPFeeRequestMeta))
	assert.Nil(t, err)

	// invalid nftID (PRV) mint PRV inst
	mintPrvNftInst, err := instruction.NewMintNftWithValue(common.PRVCoinID, validOTAReceiver0, 1, *txHash).
		StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawLPFeeRequestMeta))
	assert.Nil(t, err)

	// invalidNftID
	invalidNftIDMetaData, _ := metadataPdexv3.NewPdexv3WithdrawalLPFeeRequest(
		metadataCommon.Pdexv3WithdrawLPFeeRequestMeta,
		poolPairID, metadataPdexv3.AccessOption{
			NftID: &common.PRVCoinID,
		}, map[common.Hash]privacy.OTAReceiver{
			common.PRVCoinID: otaReceiver0,
		},
	)
	invalidNftIDTx := &metadataMocks.Transaction{}
	invalidNftIDTx.On("GetMetadata").Return(invalidNftIDMetaData)
	invalidNftIDTx.On("GetValidationEnv").Return(valEnv)
	invalidNftIDTx.On("Hash").Return(txHash)

	rejectNftIDInst := v2utils.BuildWithdrawLPFeeInsts(
		poolPairID, metadataPdexv3.AccessOption{
			NftID: &common.PRVCoinID,
		},
		map[common.Hash]metadataPdexv3.ReceiverInfo{},
		1, *txHash, metadataPdexv3.RequestRejectedChainStatus, nil,
	)[0]
	assert.Nil(t, err)

	// validInput
	validInputMetaData, _ := metadataPdexv3.NewPdexv3WithdrawalLPFeeRequest(
		metadataCommon.Pdexv3WithdrawLPFeeRequestMeta,
		poolPairID, metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, map[common.Hash]privacy.OTAReceiver{
			*nftHash:         otaReceiver0,
			*token0ID:        otaReceiver0,
			*token1ID:        otaReceiver1,
			common.PRVCoinID: otaReceiver1,
		},
	)
	validInputTx := &metadataMocks.Transaction{}
	validInputTx.On("GetMetadata").Return(validInputMetaData)
	validInputTx.On("GetValidationEnv").Return(valEnv)
	validInputTx.On("Hash").Return(txHash)

	// accept instructions
	acceptWithdrawLPInstsOnlyLP := v2utils.BuildWithdrawLPFeeInsts(
		poolPairID, metadataPdexv3.AccessOption{
			NftID: nftHash,
		},
		map[common.Hash]metadataPdexv3.ReceiverInfo{
			*token0ID: {
				Address: otaReceiver0,
				Amount:  300,
			},
			*token1ID: {
				Address: otaReceiver1,
				Amount:  1200,
			},
			common.PRVCoinID: {
				Address: otaReceiver1,
				Amount:  600,
			},
		},
		1, *txHash, metadataPdexv3.RequestAcceptedChainStatus, nil,
	)
	acceptWithdrawLPInstsOnlyOrderReward := v2utils.BuildWithdrawLPFeeInsts(
		poolPairID, metadataPdexv3.AccessOption{
			NftID: nftHash,
		},
		map[common.Hash]metadataPdexv3.ReceiverInfo{
			*token0ID: {
				Address: otaReceiver0,
				Amount:  150,
			},
			*token1ID: {
				Address: otaReceiver1,
				Amount:  250,
			},
		},
		1, *txHash, metadataPdexv3.RequestAcceptedChainStatus, nil,
	)
	acceptWithdrawLPInstsBothReward := v2utils.BuildWithdrawLPFeeInsts(
		poolPairID, metadataPdexv3.AccessOption{
			NftID: nftHash,
		}, map[common.Hash]metadataPdexv3.ReceiverInfo{
			*token0ID: {
				Address: otaReceiver0,
				Amount:  450,
			},
			*token1ID: {
				Address: otaReceiver1,
				Amount:  1450,
			},
		},
		1, *txHash, metadataPdexv3.RequestAcceptedChainStatus, nil,
	)

	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		txs       []metadata.Transaction
		poolPairs map[string]*PoolPairState
		nftIDs    map[string]uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		want1   map[string]*PoolPairState
		wantErr bool
	}{
		{
			name:   "Invalid pool pair id",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{invalidPoolPairIDTx},
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
						orderRewards:  map[string]*OrderReward{},
						makingVolume:  map[common.Hash]*MakingVolume{},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{mintNftInst, rejectPoolPairInst},
			want1: map[string]*PoolPairState{
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
					orderRewards:  map[string]*OrderReward{},
					makingVolume:  map[common.Hash]*MakingVolume{},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			wantErr: false,
		},
		{
			name:   "Invalid nftID",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{invalidNftIDTx},
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
						orderRewards:  map[string]*OrderReward{},
						makingVolume:  map[common.Hash]*MakingVolume{},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{mintPrvNftInst, rejectNftIDInst},
			want1: map[string]*PoolPairState{
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
					orderRewards:  map[string]*OrderReward{},
					makingVolume:  map[common.Hash]*MakingVolume{},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			wantErr: false,
		},
		{
			name:   "Valid LP withdrawal",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{validInputTx},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 100, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare: map[common.Hash]*big.Int{
							*token0ID: convertToLPFeesPerShare(300, 300),
							*token1ID: convertToLPFeesPerShare(1200, 300),
						},
						lmRewardsPerShare: map[common.Hash]*big.Int{
							common.PRVCoinID: convertToLPFeesPerShare(600, 200),
						},
						protocolFees:    map[common.Hash]uint64{},
						stakingPoolFees: map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount:         300,
								lmLockedAmount: 100,
								tradingFees: map[common.Hash]uint64{
									*token0ID:        100,
									*token1ID:        200,
									common.PRVCoinID: 100,
								},
								lastLPFeesPerShare: map[common.Hash]*big.Int{
									*token0ID: convertToLPFeesPerShare(100, 300),
									*token1ID: convertToLPFeesPerShare(200, 300),
								},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{
									common.PRVCoinID: convertToLPFeesPerShare(100, 200),
								},
							},
						},
						orderRewards: map[string]*OrderReward{},
						makingVolume: map[common.Hash]*MakingVolume{},
						orderbook:    Orderbook{[]*Order{}},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{mintNftInst, acceptWithdrawLPInstsOnlyLP[0], acceptWithdrawLPInstsOnlyLP[1], acceptWithdrawLPInstsOnlyLP[2]},
			want1: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 100, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					lpFeesPerShare: map[common.Hash]*big.Int{
						*token0ID: convertToLPFeesPerShare(300, 300),
						*token1ID: convertToLPFeesPerShare(1200, 300),
					},
					lmRewardsPerShare: map[common.Hash]*big.Int{
						common.PRVCoinID: convertToLPFeesPerShare(600, 200),
					},
					protocolFees:    map[common.Hash]uint64{},
					stakingPoolFees: map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount:         300,
							lmLockedAmount: 100,
							tradingFees: map[common.Hash]uint64{
								*token0ID:        0,
								*token1ID:        0,
								common.PRVCoinID: 0,
							},
							lastLPFeesPerShare: map[common.Hash]*big.Int{
								*token0ID: convertToLPFeesPerShare(300, 300),
								*token1ID: convertToLPFeesPerShare(1200, 300),
							},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{
								common.PRVCoinID: convertToLPFeesPerShare(600, 200),
							},
						},
					},
					orderRewards: map[string]*OrderReward{},
					makingVolume: map[common.Hash]*MakingVolume{},
					orderbook:    Orderbook{[]*Order{}},
				},
			},
			wantErr: false,
		},
		{
			name:   "Valid order reward withdrawal",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{validInputTx},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 0, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare: map[common.Hash]*big.Int{
							*token0ID: convertToLPFeesPerShare(300, 300),
							*token1ID: convertToLPFeesPerShare(1200, 300),
						},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares:            map[string]*Share{},
						orderRewards: map[string]*OrderReward{
							nftID: &OrderReward{
								uncollectedRewards: map[common.Hash]*OrderRewardDetail{
									*token0ID: {
										amount: 150,
									},
									*token1ID: {
										amount: 250,
									},
								},
								withdrawnStatus: DefaultWithdrawnOrderReward,
							},
						},
						makingVolume: map[common.Hash]*MakingVolume{},
						orderbook:    Orderbook{[]*Order{}},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{mintNftInst, acceptWithdrawLPInstsOnlyOrderReward[0], acceptWithdrawLPInstsOnlyOrderReward[1]},
			want1: map[string]*PoolPairState{
				poolPairID: {
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 0, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					lpFeesPerShare: map[common.Hash]*big.Int{
						*token0ID: convertToLPFeesPerShare(300, 300),
						*token1ID: convertToLPFeesPerShare(1200, 300),
					},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares:            map[string]*Share{},
					orderRewards:      map[string]*OrderReward{},
					makingVolume:      map[common.Hash]*MakingVolume{},
					orderbook:         Orderbook{[]*Order{}},
				},
			},
			wantErr: false,
		},
		{
			name:   "Valid both rewards withdrawal",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{validInputTx},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 0, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare: map[common.Hash]*big.Int{
							*token0ID: convertToLPFeesPerShare(300, 300),
							*token1ID: convertToLPFeesPerShare(1200, 300),
						},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount: 300,
								tradingFees: map[common.Hash]uint64{
									*token0ID: 100,
									*token1ID: 200,
								},
								lastLPFeesPerShare: map[common.Hash]*big.Int{
									*token0ID: convertToLPFeesPerShare(100, 300),
									*token1ID: convertToLPFeesPerShare(200, 300),
								},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderRewards: map[string]*OrderReward{
							nftID: &OrderReward{
								uncollectedRewards: map[common.Hash]*OrderRewardDetail{
									*token0ID: {
										amount: 150,
									},
									*token1ID: {
										amount: 250,
									},
								},
								withdrawnStatus: DefaultWithdrawnOrderReward,
							},
						},
						makingVolume: map[common.Hash]*MakingVolume{},
						orderbook:    Orderbook{[]*Order{}},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{mintNftInst, acceptWithdrawLPInstsBothReward[0], acceptWithdrawLPInstsBothReward[1]},
			want1: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 0, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					lpFeesPerShare: map[common.Hash]*big.Int{
						*token0ID: convertToLPFeesPerShare(300, 300),
						*token1ID: convertToLPFeesPerShare(1200, 300),
					},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount: 300,
							tradingFees: map[common.Hash]uint64{
								*token0ID: 0,
								*token1ID: 0,
							},
							lastLPFeesPerShare: map[common.Hash]*big.Int{
								*token0ID: convertToLPFeesPerShare(300, 300),
								*token1ID: convertToLPFeesPerShare(1200, 300),
							},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderRewards: map[string]*OrderReward{},
					makingVolume: map[common.Hash]*MakingVolume{},
					orderbook:    Orderbook{[]*Order{}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerV2{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			got, got1, err := sp.withdrawLPFee(tt.args.txs, tt.args.poolPairs)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV2.withdrawLPFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV2.withdrawLPFee() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProducerV2.withdrawLPFee() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_stateProducerV2_withdrawProtocolFee(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)
	txHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)

	otaReceiver0 := privacy.OTAReceiver{}
	err = otaReceiver0.FromString(validOTAReceiver0)
	assert.Nil(t, err)
	otaReceiver1 := privacy.OTAReceiver{}
	err = otaReceiver1.FromString(validOTAReceiver1)
	assert.Nil(t, err)

	config.AbortParam()
	config.Param().PDexParams.ProtocolFundAddress = "12svfkP6w5UDJDSCwqH978PvqiqBxKmUnA9em9yAYWYJVRv7wuXY1qhhYpPAm4BDz2mLbFrRmdK3yRhnTqJCZXKHUmoi7NV83HCH2YFpctHNaDdkSiQshsjw2UFUuwdEvcidgaKmF3VJpY5f8RdN"

	// invalidPoolPairID
	invalidPoolPairIDMetaData, _ := metadataPdexv3.NewPdexv3WithdrawalProtocolFeeRequest(
		metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta,
		"123",
	)

	invalidPoolPairIDTx := &metadataMocks.Transaction{}
	invalidPoolPairIDTx.On("GetMetadata").Return(invalidPoolPairIDMetaData)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)
	invalidPoolPairIDTx.On("GetValidationEnv").Return(valEnv)
	invalidPoolPairIDTx.On("Hash").Return(txHash)

	// invalid pool pair
	rejectPoolPairInst := v2utils.BuildWithdrawProtocolFeeInsts(
		"123",
		config.Param().PDexParams.ProtocolFundAddress,
		map[common.Hash]uint64{},
		0, *txHash, metadataPdexv3.RequestRejectedChainStatus,
	)[0]
	assert.Nil(t, err)

	// validInput
	validInputMetaData, _ := metadataPdexv3.NewPdexv3WithdrawalProtocolFeeRequest(
		metadataCommon.Pdexv3WithdrawLPFeeRequestMeta,
		poolPairID,
	)
	validInputTx := &metadataMocks.Transaction{}
	validInputTx.On("GetMetadata").Return(validInputMetaData)
	validInputTx.On("GetValidationEnv").Return(valEnv)
	validInputTx.On("Hash").Return(txHash)

	// accept instructions
	acceptWithdrawLPInsts := v2utils.BuildWithdrawProtocolFeeInsts(
		poolPairID,
		config.Param().PDexParams.ProtocolFundAddress,
		map[common.Hash]uint64{
			*token0ID: 10,
			*token1ID: 20,
		},
		0, *txHash, metadataPdexv3.RequestAcceptedChainStatus,
	)

	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		txs       []metadata.Transaction
		poolPairs map[string]*PoolPairState
		nftIDs    map[string]uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		want1   map[string]*PoolPairState
		wantErr bool
	}{
		{
			name:   "Invalid pool pair id",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{invalidPoolPairIDTx},
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
						orderRewards:  map[string]*OrderReward{},
						makingVolume:  map[common.Hash]*MakingVolume{},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{rejectPoolPairInst},
			want1: map[string]*PoolPairState{
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
					orderRewards:  map[string]*OrderReward{},
					makingVolume:  map[common.Hash]*MakingVolume{},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			wantErr: false,
		},
		{
			name:   "Valid Input",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{validInputTx},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 300, 0, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare: map[common.Hash]*big.Int{
							*token0ID: convertToLPFeesPerShare(300, 300),
							*token1ID: convertToLPFeesPerShare(1200, 300),
						},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees: map[common.Hash]uint64{
							*token0ID: 10,
							*token1ID: 20,
						},
						stakingPoolFees: map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount: 300,
								tradingFees: map[common.Hash]uint64{
									*token0ID: 100,
									*token1ID: 200,
								},
								lastLPFeesPerShare: map[common.Hash]*big.Int{
									*token0ID: convertToLPFeesPerShare(100, 300),
									*token1ID: convertToLPFeesPerShare(200, 300),
								},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderRewards:  map[string]*OrderReward{},
						makingVolume:  map[common.Hash]*MakingVolume{},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{acceptWithdrawLPInsts[0], acceptWithdrawLPInsts[1]},
			want1: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 0, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					lpFeesPerShare: map[common.Hash]*big.Int{
						*token0ID: convertToLPFeesPerShare(300, 300),
						*token1ID: convertToLPFeesPerShare(1200, 300),
					},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees: map[common.Hash]uint64{
						*token0ID: 0,
						*token1ID: 0,
					},
					stakingPoolFees: map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: &Share{
							amount: 300,
							tradingFees: map[common.Hash]uint64{
								*token0ID: 100,
								*token1ID: 200,
							},
							lastLPFeesPerShare: map[common.Hash]*big.Int{
								*token0ID: convertToLPFeesPerShare(100, 300),
								*token1ID: convertToLPFeesPerShare(200, 300),
							},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderRewards:  map[string]*OrderReward{},
					makingVolume:  map[common.Hash]*MakingVolume{},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerV2{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			got, got1, err := sp.withdrawProtocolFee(tt.args.txs, tt.args.poolPairs)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV2.withdrawProtocolFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV2.withdrawProtocolFee() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProducerV2.withdrawProtocolFee() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_stateProducerV2_userMintNft(t *testing.T) {
	txReqID, err := common.Hash{}.NewHashFromStr("1111122222")
	assert.Nil(t, err)
	nftHash1, err := common.Hash{}.NewHashFromStr(nftID1)
	assert.Nil(t, err)

	rejectInst, err := instruction.NewRejectUserMintNftWithValue(validOTAReceiver0, 100, 1, *txReqID).StringSlice()
	assert.Nil(t, err)
	acceptInst, err := instruction.NewAcceptUserMintNftWithValue(
		validOTAReceiver0, 100, 1, *nftHash1, *txReqID,
	).StringSlice()
	assert.Nil(t, err)

	metaData := metadataPdexv3.NewUserMintNftRequestWithValue(validOTAReceiver0, 100)
	tx := &metadataMocks.Transaction{}
	tx.On("GetMetadata").Return(metaData)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)
	tx.On("GetValidationEnv").Return(valEnv)
	tx.On("Hash").Return(txReqID)

	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		txs                  []metadata.Transaction
		nftIDs               map[string]uint64
		beaconHeight         uint64
		mintNftRequireAmount uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		want1   map[string]uint64
		wantErr bool
	}{
		{
			name:   "Reject mint nft",
			fields: fields{},
			args: args{
				txs:                  []metadata.Transaction{tx},
				beaconHeight:         10,
				mintNftRequireAmount: 1000,
				nftIDs:               map[string]uint64{},
			},
			want:    [][]string{rejectInst},
			want1:   map[string]uint64{},
			wantErr: false,
		},
		{
			name:   "Valid mint nft",
			fields: fields{},
			args: args{
				txs:                  []metadata.Transaction{tx},
				beaconHeight:         10,
				mintNftRequireAmount: 100,
				nftIDs:               map[string]uint64{},
			},
			want: [][]string{acceptInst},
			want1: map[string]uint64{
				nftID1: 100,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerV2{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			atc, err := (&v2utils.NFTAssetTagsCache{}).FromIDs(tt.args.nftIDs)
			assert.Nil(t, err)
			got, got1, _, err := sp.userMintNft(tt.args.txs, tt.args.nftIDs, atc, tt.args.beaconHeight, tt.args.mintNftRequireAmount)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV2.userMintNft() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV2.userMintNft() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProducerV2.userMintNft() got1 = %v, want %v", got1, tt.want1)
			}

			// check assetTagsCache consistency
			expectedAtc, err := (&v2utils.NFTAssetTagsCache{}).FromIDs(tt.want1)
			assert.Nil(t, err)
			if !reflect.DeepEqual(atc, expectedAtc) {
				t.Errorf("stateProcessorV2.userMintNft() got1 = %v, want %v", atc, expectedAtc)
			}
		})
	}
}

func Test_stateProducerV2_staking(t *testing.T) {
	initTestParams(t)
	txReqID, err := common.Hash{}.NewHashFromStr("1111122222")
	assert.Nil(t, err)
	nftHash1, err := common.Hash{}.NewHashFromStr(nftID1)
	assert.Nil(t, err)

	acceptInst, err := instruction.NewAcceptStakingWithValue(
		common.PRVCoinID, *txReqID, 1, 100, nil, metadataPdexv3.AccessOption{
			NftID: nftHash1,
		},
	).StringSlice()
	assert.Nil(t, err)

	//invalidStakingPoolID
	invalidStakingPoolIDMetadata := metadataPdexv3.NewStakingRequestWithValue(
		"abcd", validOTAReceiver0, 100, metadataPdexv3.AccessOption{NftID: nftHash1}, nil,
	)
	invalidStakingPoolIDTx := &metadataMocks.Transaction{}
	invalidStakingPoolIDTx.On("GetMetadata").Return(invalidStakingPoolIDMetadata)
	valEnv0 := tx_generic.DefaultValEnv()
	valEnv0 = tx_generic.WithShardID(valEnv0, 1)
	invalidStakingPoolIDTx.On("GetValidationEnv").Return(valEnv0)
	invalidStakingPoolIDTx.On("Hash").Return(txReqID)
	tempHash, _ := common.Hash{}.NewHashFromStr("abcd")
	invalidStakingPoolIDRejectInst, err := instruction.NewRejectStakingWithValue(
		validOTAReceiver0, *tempHash, *txReqID, 1, 100,
	).StringSlice()
	assert.Nil(t, err)
	//

	//not found stakingPoolID
	notFoundStakingPoolIDMetadata := metadataPdexv3.NewStakingRequestWithValue(
		txReqID.String(), validOTAReceiver0, 100, metadataPdexv3.AccessOption{NftID: nftHash1}, nil,
	)
	notFoundStakingPoolIDTx := &metadataMocks.Transaction{}
	notFoundStakingPoolIDTx.On("GetMetadata").Return(notFoundStakingPoolIDMetadata)
	valEnv2 := tx_generic.DefaultValEnv()
	valEnv2 = tx_generic.WithShardID(valEnv2, 1)
	notFoundStakingPoolIDTx.On("GetValidationEnv").Return(valEnv2)
	notFoundStakingPoolIDTx.On("Hash").Return(txReqID)
	notFoundStakingPoolIDRejectInst, err := instruction.NewRejectStakingWithValue(
		validOTAReceiver0, *txReqID, *txReqID, 1, 100,
	).StringSlice()
	//

	//not found nftID
	notFoundNftIDMetadata := metadataPdexv3.NewStakingRequestWithValue(
		common.PRVIDStr, validOTAReceiver0, 100, metadataPdexv3.AccessOption{NftID: txReqID}, nil,
	)
	notFoundNftIDTx := &metadataMocks.Transaction{}
	notFoundNftIDTx.On("GetMetadata").Return(notFoundNftIDMetadata)
	valEnv3 := tx_generic.DefaultValEnv()
	valEnv3 = tx_generic.WithShardID(valEnv3, 1)
	notFoundNftIDTx.On("GetValidationEnv").Return(valEnv3)
	notFoundNftIDTx.On("Hash").Return(txReqID)
	notFoundNftIDRejectInst, err := instruction.NewRejectStakingWithValue(
		validOTAReceiver0, common.PRVCoinID, *txReqID, 1, 100,
	).StringSlice()

	//

	//validTx
	validMetadata := metadataPdexv3.NewStakingRequestWithValue(
		common.PRVIDStr, validOTAReceiver0, 100, metadataPdexv3.AccessOption{NftID: nftHash1}, nil,
	)
	validTx := &metadataMocks.Transaction{}
	validTx.On("GetMetadata").Return(validMetadata)
	valEnv4 := tx_generic.DefaultValEnv()
	valEnv4 = tx_generic.WithShardID(valEnv4, 1)
	validTx.On("GetValidationEnv").Return(valEnv4)
	validTx.On("Hash").Return(txReqID)
	//

	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		txs               []metadata.Transaction
		nftIDs            map[string]uint64
		stakingPoolStates map[string]*StakingPoolState
		beaconHeight      uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		want1   map[string]*StakingPoolState
		wantErr bool
	}{
		{
			name:   "Invalid StakingPoolID",
			fields: fields{},
			args: args{
				txs: []metadata.Transaction{invalidStakingPoolIDTx},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						stakers: map[string]*Staker{},
					},
				},
				beaconHeight: 10,
			},
			want: [][]string{invalidStakingPoolIDRejectInst},
			want1: map[string]*StakingPoolState{
				common.PRVIDStr: &StakingPoolState{
					stakers: map[string]*Staker{},
				},
			},
			wantErr: false,
		},
		{
			name: "Not found stakingPoolID",
			args: args{
				txs: []metadata.Transaction{notFoundStakingPoolIDTx},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						stakers: map[string]*Staker{},
					},
				},
				beaconHeight: 10,
			},
			want: [][]string{notFoundStakingPoolIDRejectInst},
			want1: map[string]*StakingPoolState{
				common.PRVIDStr: &StakingPoolState{
					stakers: map[string]*Staker{},
				},
			},
			wantErr: false,
		},
		{
			name: "Not found nftID",
			args: args{
				txs: []metadata.Transaction{notFoundNftIDTx},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{

						stakers: map[string]*Staker{},
					},
				},
				beaconHeight: 10,
			},
			want: [][]string{notFoundNftIDRejectInst},
			want1: map[string]*StakingPoolState{
				common.PRVIDStr: &StakingPoolState{

					stakers: map[string]*Staker{},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid input",
			args: args{
				txs: []metadata.Transaction{validTx},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						stakers: map[string]*Staker{},
					},
				},
				beaconHeight: 10,
			},
			want: [][]string{acceptInst},
			want1: map[string]*StakingPoolState{
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerV2{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			got, got1, err := sp.staking(tt.args.txs, tt.args.nftIDs, tt.args.stakingPoolStates, tt.args.beaconHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV2.staking() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV2.staking() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProducerV2.staking() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_stateProducerV2_unstaking(t *testing.T) {
	txReqID, err := common.Hash{}.NewHashFromStr("1111122222")
	assert.Nil(t, err)
	nftHash1, err := common.Hash{}.NewHashFromStr(nftID1)
	assert.Nil(t, err)
	nftHash, err := common.Hash{}.NewHashFromStr(nftID)
	assert.Nil(t, err)

	mintNft1Inst, err := instruction.NewMintNftWithValue(
		*nftHash1, validOTAReceiver0, 1, *txReqID,
	).StringSlice(
		strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta),
	)

	mintNftInst, err := instruction.NewMintNftWithValue(
		*nftHash, validOTAReceiver0, 1, *txReqID,
	).StringSlice(
		strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta),
	)

	otaReceivers := map[string]string{
		common.PRVIDStr: validOTAReceiver1,
		nftID1:          validOTAReceiver0,
		nftID:           validOTAReceiver0,
	}

	//invalidStakingPoolID
	invalidStakingPoolIDMetadata := metadataPdexv3.NewUnstakingRequestWithValue(
		"abcd", otaReceivers, 50, metadataPdexv3.AccessOption{NftID: nftHash1},
	)
	invalidStakingPoolIDTx := &metadataMocks.Transaction{}
	invalidStakingPoolIDTx.On("GetMetadata").Return(invalidStakingPoolIDMetadata)
	valEnv0 := tx_generic.DefaultValEnv()
	valEnv0 = tx_generic.WithShardID(valEnv0, 1)
	invalidStakingPoolIDTx.On("GetValidationEnv").Return(valEnv0)
	invalidStakingPoolIDTx.On("Hash").Return(txReqID)
	invalidStakingPoolIDRejectInst, err := instruction.NewRejectUnstakingWithValue(*txReqID, 1, "abcd", nil, nil).StringSlice()
	assert.Nil(t, err)
	//

	//Not found nftID
	invalidNftIDMetadata := metadataPdexv3.NewUnstakingRequestWithValue(
		common.PRVIDStr, otaReceivers, 50, metadataPdexv3.AccessOption{NftID: nftHash},
	)
	invalidNftIDTx := &metadataMocks.Transaction{}
	invalidNftIDTx.On("GetMetadata").Return(invalidNftIDMetadata)
	valEnv1 := tx_generic.DefaultValEnv()
	valEnv1 = tx_generic.WithShardID(valEnv1, 1)
	invalidNftIDTx.On("GetValidationEnv").Return(valEnv1)
	invalidNftIDTx.On("Hash").Return(txReqID)
	invalidNftIDRejectInst, err := instruction.NewRejectUnstakingWithValue(*txReqID, 1, common.PRVIDStr, nil, nil).StringSlice()
	//

	//invalidAmount
	invalidAmountMetadata := metadataPdexv3.NewUnstakingRequestWithValue(
		common.PRVIDStr, otaReceivers, 300, metadataPdexv3.AccessOption{NftID: nftHash1},
	)
	invalidAmountTx := &metadataMocks.Transaction{}
	invalidAmountTx.On("GetMetadata").Return(invalidAmountMetadata)
	valEnv3 := tx_generic.DefaultValEnv()
	valEnv3 = tx_generic.WithShardID(valEnv3, 1)
	invalidAmountTx.On("GetValidationEnv").Return(valEnv3)
	invalidAmountTx.On("Hash").Return(txReqID)
	invalidAmountInst, err := instruction.NewRejectUnstakingWithValue(*txReqID, 1, common.PRVIDStr, nil, nil).StringSlice()
	//

	//validTx
	validMetadata := metadataPdexv3.NewUnstakingRequestWithValue(
		common.PRVIDStr, otaReceivers, 50, metadataPdexv3.AccessOption{NftID: nftHash1},
	)
	validTx := &metadataMocks.Transaction{}
	validTx.On("GetMetadata").Return(validMetadata)
	valEnv4 := tx_generic.DefaultValEnv()
	valEnv4 = tx_generic.WithShardID(valEnv4, 1)
	validTx.On("GetValidationEnv").Return(valEnv4)
	validTx.On("Hash").Return(txReqID)
	//

	acceptInst, err := instruction.NewAcceptUnstakingWithValue(
		common.PRVCoinID, 50, validOTAReceiver1, *txReqID, 1, metadataPdexv3.AccessOption{NftID: nftHash1}, nil,
	).StringSlice()
	assert.Nil(t, err)

	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		txs               []metadata.Transaction
		nftIDs            map[string]uint64
		stakingPoolStates map[string]*StakingPoolState
		beaconHeight      uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		want1   map[string]*StakingPoolState
		wantErr bool
	}{
		{
			name: "Not found staking pool",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{invalidStakingPoolIDTx},
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
			want: [][]string{invalidStakingPoolIDRejectInst, mintNft1Inst},
			want1: map[string]*StakingPoolState{
				common.PRVIDStr: &StakingPoolState{
					liquidity: 150,
					stakers: map[string]*Staker{
						nftID1: &Staker{
							liquidity: 150,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Not found nftID in nftIDs list",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{invalidNftIDTx},
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
			want: [][]string{invalidNftIDRejectInst, mintNftInst},
			want1: map[string]*StakingPoolState{
				common.PRVIDStr: &StakingPoolState{
					liquidity: 150,
					stakers: map[string]*Staker{
						nftID1: &Staker{
							liquidity: 150,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Not found nftID in list staker",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{invalidNftIDTx},
				nftIDs: map[string]uint64{
					nftID: 100,
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
			want: [][]string{invalidNftIDRejectInst, mintNftInst},
			want1: map[string]*StakingPoolState{
				common.PRVIDStr: &StakingPoolState{
					liquidity: 150,
					stakers: map[string]*Staker{
						nftID1: &Staker{
							liquidity: 150,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Can not update liquidity",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{invalidAmountTx},
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
			want: [][]string{invalidAmountInst, mintNft1Inst},
			want1: map[string]*StakingPoolState{
				common.PRVIDStr: &StakingPoolState{
					liquidity: 150,
					stakers: map[string]*Staker{
						nftID1: &Staker{
							liquidity: 150,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid input",
			fields: fields{
				stateProducerBase: stateProducerBase{},
			},
			args: args{
				txs: []metadata.Transaction{validTx},
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
			want: [][]string{acceptInst, mintNft1Inst},
			want1: map[string]*StakingPoolState{
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerV2{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			got, got1, err := sp.unstaking(tt.args.txs, tt.args.nftIDs, tt.args.stakingPoolStates, tt.args.beaconHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV2.unstaking() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV2.unstaking() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProducerV2.unstaking() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_stateProducerV2_liquidityMining(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	config.AbortParam()
	config.Param().PDexParams.ProtocolFundAddress = "12svfkP6w5UDJDSCwqH978PvqiqBxKmUnA9em9yAYWYJVRv7wuXY1qhhYpPAm4BDz2mLbFrRmdK3yRhnTqJCZXKHUmoi7NV83HCH2YFpctHNaDdkSiQshsjw2UFUuwdEvcidgaKmF3VJpY5f8RdN"

	// update LP reward instructions
	mintLPReward1 := v2utils.BuildMintBlockRewardInst(poolPairID, 1000000, common.PRVCoinID)
	mintLPReward2 := v2utils.BuildMintBlockRewardInst(poolPairID, 700000, common.PRVCoinID)
	mintLPReward3 := v2utils.BuildMintBlockRewardInst(poolPairID, 850000, common.PRVCoinID)

	// update LOP reward instructiosn
	mintLOPReward1 := v2utils.BuildDistributeMiningOrderRewardInsts(
		poolPairID, *token0ID, 150000, common.PRVCoinID,
	)
	mintLOPReward2 := v2utils.BuildDistributeMiningOrderRewardInsts(
		poolPairID, *token1ID, 150000, common.PRVCoinID,
	)

	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		poolPairs map[string]*PoolPairState
		reward    uint64
		params    *Params
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		want1   map[string]*PoolPairState
		wantErr bool
	}{
		{
			name:   "Only mint for LPs",
			fields: fields{},
			args: args{
				reward: 1000000,
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 500, 200, 250, 1000,
							big.NewInt(0).SetUint64(500),
							big.NewInt(0).SetUint64(2000), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID: &Share{
								amount:                300,
								lmLockedAmount:        150,
								tradingFees:           map[common.Hash]uint64{},
								lastLPFeesPerShare:    map[common.Hash]*big.Int{},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
							nftID1: &Share{
								amount:                200,
								lmLockedAmount:        50,
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
				params: &Params{
					PDEXRewardPoolPairsShare: map[string]uint{
						poolPairID: 100,
					},
					OrderLiquidityMiningBPS: map[string]uint{},
				},
			},
			want: mintLPReward1,
			want1: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 500, 200, 250, 1000,
						big.NewInt(0).SetUint64(500),
						big.NewInt(0).SetUint64(2000), 20000,
					),
					lpFeesPerShare: map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{
						common.PRVCoinID: convertToLPFeesPerShare(1000000, 300),
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
						nftID1: &Share{
							amount:                200,
							lmLockedAmount:        50,
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
			wantErr: false,
		},
		{
			name:   "Split reward for LPs and LOPs",
			fields: fields{},
			args: args{
				reward: 1000000,
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
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				params: &Params{
					PDEXRewardPoolPairsShare: map[string]uint{
						poolPairID: 100,
					},
					OrderLiquidityMiningBPS: map[string]uint{
						poolPairID: 1500,
					},
				},
			},
			want: [][]string{mintLOPReward1[0], mintLOPReward2[0], mintLPReward2[0]},
			want1: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 0, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					lpFeesPerShare: map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{
						common.PRVCoinID: convertToLPFeesPerShare(700000, 300),
					},
					protocolFees:    map[common.Hash]uint64{},
					stakingPoolFees: map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: {
							amount:                300,
							tradingFees:           map[common.Hash]uint64{},
							lastLPFeesPerShare:    map[common.Hash]*big.Int{},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
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
					makingVolume:  map[common.Hash]*MakingVolume{},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			wantErr: false,
		},
		{
			name:   "One of making token volumes is zero",
			fields: fields{},
			args: args{
				reward: 1000000,
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
						makingVolume: map[common.Hash]*MakingVolume{
							*token1ID: &MakingVolume{
								volume: map[string]*big.Int{
									nftID:  big.NewInt(0).SetUint64(50),
									nftID1: big.NewInt(0).SetUint64(100),
								},
							},
						},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				params: &Params{
					PDEXRewardPoolPairsShare: map[string]uint{
						poolPairID: 100,
					},
					OrderLiquidityMiningBPS: map[string]uint{
						poolPairID: 1500,
					},
				},
			},
			want: [][]string{mintLOPReward2[0], mintLPReward3[0]},
			want1: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 0, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					lpFeesPerShare: map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{
						common.PRVCoinID: convertToLPFeesPerShare(850000, 300),
					},
					protocolFees:    map[common.Hash]uint64{},
					stakingPoolFees: map[common.Hash]uint64{},
					shares: map[string]*Share{
						nftID: {
							amount:                300,
							tradingFees:           map[common.Hash]uint64{},
							lastLPFeesPerShare:    map[common.Hash]*big.Int{},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderRewards: map[string]*OrderReward{
						nftID: {
							uncollectedRewards: map[common.Hash]*OrderRewardDetail{
								common.PRVCoinID: {
									amount: 212500, // 162500 + 50000
								},
							},
						},
						nftID1: {
							uncollectedRewards: map[common.Hash]*OrderRewardDetail{
								common.PRVCoinID: {
									amount: 237500, // 137500 + 100000
								},
							},
						},
					},
					makingVolume:  map[common.Hash]*MakingVolume{},
					orderbook:     Orderbook{[]*Order{}},
					lmLockedShare: map[string]map[uint64]uint64{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerV2{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			got, got1, err := sp.mintReward(
				common.PRVCoinID,
				tt.args.reward,
				tt.args.params,
				tt.args.poolPairs,
				true,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV2.mintReward() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV2.mintReward() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateProducerV2.mintReward() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
