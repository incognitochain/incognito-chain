package pdex

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
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

	// first contribution tx
	firstContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
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
			"", validOTAReceiver0, validOTAReceiver1,
			*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
		),
		"pair_hash")
	waitingContributionInst := instruction.NewWaitingAddLiquidityWithValue(*waitingContributionStateDB)
	waitingContributionInstBytes, err := json.Marshal(waitingContributionInst)
	//

	// return contribution tx by sanity
	refundContributionSanityMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		token0ID.String(), nftID, 200, 20000,
	)
	assert.Nil(t, err)
	refundContributionSanityTx := &metadataMocks.Transaction{}
	refundContributionSanityTx.On("GetMetadata").Return(refundContributionSanityMetaData)
	refundContributionSanityTx.On("GetValidationEnv").Return(valEnv)
	refundContributionSanityTx.On("Hash").Return(secondTxHash)
	refundContributionSanityState0 := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0, validOTAReceiver1,
			*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
		),
		"pair_hash")
	refundContributionSanityInst0 := instruction.NewRefundAddLiquidityWithValue(*refundContributionSanityState0)
	refundContributionSanityInstBytes0, err := json.Marshal(refundContributionSanityInst0)
	refundContributionSanityState1 := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0, validOTAReceiver1,
			*token0ID, *secondTxHash, *nftHash, 200, 20000, 1,
		),
		"pair_hash")
	refundContributionSanityInst1 := instruction.NewRefundAddLiquidityWithValue(*refundContributionSanityState1)
	refundContributionSanityInstBytes1, err := json.Marshal(refundContributionSanityInst1)
	//

	// match contribution
	matchContributionMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash", validOTAReceiver0, validOTAReceiver1,
		token1ID.String(), nftID, 400, 20000,
	)
	assert.Nil(t, err)
	matchContributionTx := &metadataMocks.Transaction{}
	matchContributionTx.On("GetMetadata").Return(matchContributionMetaData)
	matchContributionTx.On("GetValidationEnv").Return(valEnv)
	matchContributionTx.On("Hash").Return(secondTxHash)
	matchContributionState := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0, validOTAReceiver1,
			*token1ID, *secondTxHash, *nftHash, 400, 20000, 1,
		),
		"pair_hash")
	matchContributionInst := instruction.NewMatchAddLiquidityWithValue(*matchContributionState, poolPairID)
	matchContributionInstBytes, err := json.Marshal(matchContributionInst)
	//

	// match contribution - 2
	matchContribution2MetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		token1ID.String(), nftID, 400, 20000,
	)
	matchContribution2Tx := &metadataMocks.Transaction{}
	matchContribution2Tx.On("GetMetadata").Return(matchContribution2MetaData)
	matchContribution2Tx.On("GetValidationEnv").Return(valEnv)
	matchContribution2Tx.On("Hash").Return(secondTxHash)
	matchContribution2State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0, validOTAReceiver1,
			*token1ID, *secondTxHash, *nftHash, 400, 20000, 1,
		),
		"pair_hash")
	matchContribution2Inst := instruction.NewMatchAddLiquidityWithValue(*matchContribution2State, poolPairID)
	matchContribution2InstBytes, err := json.Marshal(matchContribution2Inst)
	//

	// refund contributions by amount
	refundContributionAmountMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		poolPairID, "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		token1ID.String(), nftID, 0, 20000,
	)
	refundContributionAmountTx := &metadataMocks.Transaction{}
	refundContributionAmountTx.On("GetMetadata").Return(refundContributionAmountMetaData)
	refundContributionAmountTx.On("GetValidationEnv").Return(valEnv)
	refundContributionAmountTx.On("Hash").Return(fourthTxHash)

	refundContributionAmount0State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0, validOTAReceiver1,
			*token0ID, *thirdTxHash, *nftHash, 50, 20000, 1,
		),
		"pair_hash")
	refundContributionAmount0Inst := instruction.NewRefundAddLiquidityWithValue(*refundContributionAmount0State)
	refundContributionAmount0InstBytes, err := json.Marshal(refundContributionAmount0Inst)
	refundContributionAmount1State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0, validOTAReceiver1,
			*token1ID, *fourthTxHash, *nftHash, 0, 20000, 1,
		),
		"pair_hash")
	refundContributionAmount1Inst := instruction.NewRefundAddLiquidityWithValue(*refundContributionAmount1State)
	refundContributionAmount1InstBytes, err := json.Marshal(refundContributionAmount1Inst)
	//

	// match and return contribution
	matchAndReturnContributionMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		poolPairID, "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		token1ID.String(), nftID, 200, 20000,
	)
	matchAndReturnContributionTx := &metadataMocks.Transaction{}
	matchAndReturnContributionTx.On("GetMetadata").Return(matchAndReturnContributionMetaData)
	matchAndReturnContributionTx.On("GetValidationEnv").Return(valEnv)
	matchAndReturnContributionTx.On("Hash").Return(fourthTxHash)

	matchAndReturnContribution0State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0, validOTAReceiver1,
			*token0ID, *thirdTxHash, *nftHash, 50, 20000, 1,
		),
		"pair_hash")
	matchAndReturnContritubtion0Inst := instruction.NewMatchAndReturnAddLiquidityWithValue(
		*matchAndReturnContribution0State, 100, 0, 200, 0, *token1ID)
	matchAndReturnContritubtion0InstBytes, err := json.Marshal(matchAndReturnContritubtion0Inst)
	matchAndReturnContribution1State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0, validOTAReceiver1,
			*token1ID, *fourthTxHash, *nftHash, 200, 20000, 1,
		),
		"pair_hash")
	matchAndReturnContritubtion1Inst := instruction.NewMatchAndReturnAddLiquidityWithValue(
		*matchAndReturnContribution1State, 100, 0, 50, 0, *token0ID)
	matchAndReturnContritubtion1InstBytes, err := json.Marshal(matchAndReturnContritubtion1Inst)
	//

	// match and return contribution - 2
	matchAndReturnContribution2MetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		poolPairID, "pair_hash", validOTAReceiver0, validOTAReceiver1,
		token1ID.String(), nftID, 200, 20000,
	)
	matchAndReturnContribution2Tx := &metadataMocks.Transaction{}
	matchAndReturnContribution2Tx.On("GetMetadata").Return(matchAndReturnContribution2MetaData)
	matchAndReturnContribution2Tx.On("GetValidationEnv").Return(valEnv)
	matchAndReturnContribution2Tx.On("Hash").Return(fourthTxHash)

	matchAndReturnContribution0State2 := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0, validOTAReceiver1,
			*token0ID, *thirdTxHash, *nftHash, 50, 20000, 1,
		),
		"pair_hash")
	matchAndReturnContritubtion0Inst2 := instruction.NewMatchAndReturnAddLiquidityWithValue(
		*matchAndReturnContribution0State2, 100, 0, 200, 0, *token1ID)
	matchAndReturnContritubtion0InstBytes2, err := json.Marshal(matchAndReturnContritubtion0Inst2)
	matchAndReturnContribution1State2 := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0, validOTAReceiver1,
			*token1ID, *fourthTxHash, *nftHash, 200, 20000, 1,
		),
		"pair_hash")
	matchAndReturnContritubtion1Inst2 := instruction.NewMatchAndReturnAddLiquidityWithValue(
		*matchAndReturnContribution1State2, 100, 0, 50, 0, *token0ID)
	matchAndReturnContritubtion1InstBytes2, err := json.Marshal(matchAndReturnContritubtion1Inst2)
	//

	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		txs                  []metadata.Transaction
		beaconHeight         uint64
		poolPairs            map[string]*PoolPairState
		waitingContributions map[string]rawdbv2.Pdexv3Contribution
		nftIDs               map[string]uint64
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
			},
			want: [][]string{
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionWaitingChainStatus,
					string(waitingContributionInstBytes),
				},
			},
			want1: map[string]*PoolPairState{},
			want2: map[string]rawdbv2.Pdexv3Contribution{
				"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
					"", validOTAReceiver0, validOTAReceiver1,
					*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
				),
			},
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
						"", validOTAReceiver0, validOTAReceiver1,
						*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
					),
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionRefundChainStatus,
					string(refundContributionSanityInstBytes0),
				},
				[]string{
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
						"", validOTAReceiver0, validOTAReceiver1,
						*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
					),
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedChainStatus,
					string(matchContributionInstBytes),
				},
			},
			want1: map[string]*PoolPairState{
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
					poolPairID + "123": &PoolPairState{
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
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						"", validOTAReceiver0, validOTAReceiver1,
						*token0ID, *firstTxHash, *nftHash, 100, 20000, 1,
					),
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedChainStatus,
					string(matchContribution2InstBytes),
				},
			},
			want1: map[string]*PoolPairState{
				poolPairID + "123": &PoolPairState{
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
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						poolPairID, validOTAReceiver0, validOTAReceiver1,
						*token0ID, *thirdTxHash, *nftHash, 50, 20000, 1,
					),
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionRefundChainStatus,
					string(refundContributionAmount0InstBytes),
				},
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionRefundChainStatus,
					string(refundContributionAmount1InstBytes),
				},
			},
			want1: map[string]*PoolPairState{
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
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						poolPairID, validOTAReceiver0, validOTAReceiver1,
						*token0ID, *thirdTxHash, *nftHash, 50, 20000, 1,
					),
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedNReturnedChainStatus,
					string(matchAndReturnContritubtion0InstBytes),
				},
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedNReturnedChainStatus,
					string(matchAndReturnContritubtion1InstBytes),
				},
			},
			want1: map[string]*PoolPairState{
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
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						poolPairID, validOTAReceiver0, validOTAReceiver1,
						*token0ID, *thirdTxHash, *nftHash, 50, 20000, 1,
					),
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedNReturnedChainStatus,
					string(matchAndReturnContritubtion0InstBytes2),
				},
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedNReturnedChainStatus,
					string(matchAndReturnContritubtion1InstBytes2),
				},
			},
			want1: map[string]*PoolPairState{
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
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV2.addLiquidity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV2.addLiquidity() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				for k, v := range got1 {
					if !reflect.DeepEqual(v.shares, tt.want1[k].shares) {
						for key, value := range v.shares {
							fmt.Println("key & value:", key, value)
							fmt.Println("want value:", tt.want1[k].shares[key])
						}
						t.Errorf("shares got1 = %v, want %v", v.shares, tt.want1[k].shares)
					}
				}
				t.Errorf("stateProducerV2.addLiquidity() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("stateProducerV2.addLiquidity() got2 = %v, want %v", got2, tt.want2)
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

	//invalidPoolPairID
	invalidPoolPairIDMetaData := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
		"123", nftID, validOTAReceiver0, validOTAReceiver1, validOTAReceiver1, 100,
	)
	invalidPoolPairIDTx := &metadataMocks.Transaction{}
	invalidPoolPairIDTx.On("GetMetadata").Return(invalidPoolPairIDMetaData)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)
	invalidPoolPairIDTx.On("GetValidationEnv").Return(valEnv)
	invalidPoolPairIDTx.On("Hash").Return(txHash)
	//

	//reject instruction
	rejectWithdrawLiquidityInst, err := instruction.NewRejectWithdrawLiquidityWithValue(*txHash, 1).StringSlice()
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
		poolPairID, common.PRVIDStr, validOTAReceiver0, validOTAReceiver1, validOTAReceiver1, 100,
	)
	invalidNftIDTx := &metadataMocks.Transaction{}
	invalidNftIDTx.On("GetMetadata").Return(invalidNftIDMetaData)
	invalidNftIDTx.On("GetValidationEnv").Return(valEnv)
	invalidNftIDTx.On("Hash").Return(txHash)
	//

	//deductShareAmountFail
	deductShareAmountFailMetaData := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
		poolPairID, nftID, validOTAReceiver0, validOTAReceiver1, validOTAReceiver1, 0,
	)
	deductShareAmountFailTx := &metadataMocks.Transaction{}
	deductShareAmountFailTx.On("GetMetadata").Return(deductShareAmountFailMetaData)
	deductShareAmountFailTx.On("GetValidationEnv").Return(valEnv)
	deductShareAmountFailTx.On("Hash").Return(txHash)
	//

	//validInput
	validInputMetaData := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
		poolPairID, nftID, validOTAReceiver0, validOTAReceiver1, validOTAReceiver1, 100,
	)
	validInputTx := &metadataMocks.Transaction{}
	validInputTx.On("GetMetadata").Return(validInputMetaData)
	validInputTx.On("GetValidationEnv").Return(valEnv)
	validInputTx.On("Hash").Return(txHash)
	//

	//accept instructions
	acceptWithdrawLiquidityInst0, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *nftHash, *token0ID, 50, 100, validOTAReceiver1,
		*txHash, 1,
	).StringSlice()
	assert.Nil(t, err)
	acceptWithdrawLiquidityInst1, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *nftHash, *token1ID, 200, 100, validOTAReceiver1,
		*txHash, 1,
	).StringSlice()
	assert.Nil(t, err)
	//

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
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{rejectWithdrawLiquidityInst, mintNftInst},
			want1: map[string]*PoolPairState{
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
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{rejectWithdrawLiquidityInst, mintPrvNftInst},
			want1: map[string]*PoolPairState{
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
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{rejectWithdrawLiquidityInst, mintNftInst},
			want1: map[string]*PoolPairState{
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
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{acceptWithdrawLiquidityInst0, acceptWithdrawLiquidityInst1, mintNftInst},
			want1: map[string]*PoolPairState{
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
			sp := &stateProducerV2{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			got, got1, err := sp.withdrawLiquidity(tt.args.txs, tt.args.poolPairs, tt.args.nftIDs)
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
		"123", *nftHash, map[common.Hash]privacy.OTAReceiver{
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
		"123", *nftHash,
		map[common.Hash]metadataPdexv3.ReceiverInfo{},
		1, *txHash, metadataPdexv3.RequestRejectedChainStatus,
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
		poolPairID, common.PRVCoinID, map[common.Hash]privacy.OTAReceiver{
			common.PRVCoinID: otaReceiver0,
		},
	)
	invalidNftIDTx := &metadataMocks.Transaction{}
	invalidNftIDTx.On("GetMetadata").Return(invalidNftIDMetaData)
	invalidNftIDTx.On("GetValidationEnv").Return(valEnv)
	invalidNftIDTx.On("Hash").Return(txHash)

	rejectNftIDInst := v2utils.BuildWithdrawLPFeeInsts(
		poolPairID, common.PRVCoinID,
		map[common.Hash]metadataPdexv3.ReceiverInfo{},
		1, *txHash, metadataPdexv3.RequestRejectedChainStatus,
	)[0]
	assert.Nil(t, err)

	// validInput
	validInputMetaData, _ := metadataPdexv3.NewPdexv3WithdrawalLPFeeRequest(
		metadataCommon.Pdexv3WithdrawLPFeeRequestMeta,
		poolPairID, *nftHash, map[common.Hash]privacy.OTAReceiver{
			*nftHash:  otaReceiver0,
			*token0ID: otaReceiver0,
			*token1ID: otaReceiver1,
		},
	)
	validInputTx := &metadataMocks.Transaction{}
	validInputTx.On("GetMetadata").Return(validInputMetaData)
	validInputTx.On("GetValidationEnv").Return(valEnv)
	validInputTx.On("Hash").Return(txHash)

	// accept instructions
	acceptWithdrawLPInsts := v2utils.BuildWithdrawLPFeeInsts(
		poolPairID, *nftHash, map[common.Hash]metadataPdexv3.ReceiverInfo{
			*token0ID: {
				Address: otaReceiver0,
				Amount:  300,
			},
			*token1ID: {
				Address: otaReceiver1,
				Amount:  1200,
			},
		},
		1, *txHash, metadataPdexv3.RequestAcceptedChainStatus,
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
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{mintNftInst, rejectPoolPairInst},
			want1: map[string]*PoolPairState{
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
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{mintPrvNftInst, rejectNftIDInst},
			want1: map[string]*PoolPairState{
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
							*token0ID, *token1ID, 300, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
							map[common.Hash]*big.Int{
								*token0ID: convertToLPFeesPerShare(300, 300),
								*token1ID: convertToLPFeesPerShare(1200, 300),
							},
							map[common.Hash]uint64{}, map[common.Hash]uint64{},
						),
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
							},
						},
						orderbook: Orderbook{[]*Order{}},
					},
				},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{mintNftInst, acceptWithdrawLPInsts[0], acceptWithdrawLPInsts[1]},
			want1: map[string]*PoolPairState{
				poolPairID: &PoolPairState{
					state: *rawdbv2.NewPdexv3PoolPairWithValue(
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
						map[common.Hash]*big.Int{
							*token0ID: convertToLPFeesPerShare(300, 300),
							*token1ID: convertToLPFeesPerShare(1200, 300),
						},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
					shares: map[string]*Share{
						nftID: &Share{
							amount:      300,
							tradingFees: map[common.Hash]uint64{},
							lastLPFeesPerShare: map[common.Hash]*big.Int{
								*token0ID: convertToLPFeesPerShare(300, 300),
								*token1ID: convertToLPFeesPerShare(1200, 300),
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

	// invalidPoolPairID
	invalidPoolPairIDMetaData, _ := metadataPdexv3.NewPdexv3WithdrawalProtocolFeeRequest(
		metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta,
		"123", map[common.Hash]privacy.OTAReceiver{},
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
		map[common.Hash]metadataPdexv3.ReceiverInfo{},
		1, *txHash, metadataPdexv3.RequestRejectedChainStatus,
	)[0]
	assert.Nil(t, err)

	// validInput
	validInputMetaData, _ := metadataPdexv3.NewPdexv3WithdrawalProtocolFeeRequest(
		metadataCommon.Pdexv3WithdrawLPFeeRequestMeta,
		poolPairID, map[common.Hash]privacy.OTAReceiver{
			*token0ID: otaReceiver0,
			*token1ID: otaReceiver1,
		},
	)
	validInputTx := &metadataMocks.Transaction{}
	validInputTx.On("GetMetadata").Return(validInputMetaData)
	validInputTx.On("GetValidationEnv").Return(valEnv)
	validInputTx.On("Hash").Return(txHash)

	// accept instructions
	acceptWithdrawLPInsts := v2utils.BuildWithdrawProtocolFeeInsts(
		poolPairID, map[common.Hash]metadataPdexv3.ReceiverInfo{
			*token0ID: {
				Address: otaReceiver0,
				Amount:  10,
			},
			*token1ID: {
				Address: otaReceiver1,
				Amount:  20,
			},
		},
		1, *txHash, metadataPdexv3.RequestAcceptedChainStatus,
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
				nftIDs: map[string]uint64{
					nftID: 100,
				},
			},
			want: [][]string{rejectPoolPairInst},
			want1: map[string]*PoolPairState{
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
							*token0ID, *token1ID, 300, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
							map[common.Hash]*big.Int{
								*token0ID: convertToLPFeesPerShare(300, 300),
								*token1ID: convertToLPFeesPerShare(1200, 300),
							},
							map[common.Hash]uint64{
								*token0ID: 10,
								*token1ID: 20,
							}, map[common.Hash]uint64{},
						),
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
							},
						},
						orderbook: Orderbook{[]*Order{}},
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
						*token0ID, *token1ID, 300, 150, 600,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
						map[common.Hash]*big.Int{
							*token0ID: convertToLPFeesPerShare(300, 300),
							*token1ID: convertToLPFeesPerShare(1200, 300),
						},
						map[common.Hash]uint64{}, map[common.Hash]uint64{},
					),
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
			got, got1, err := sp.userMintNft(tt.args.txs, tt.args.nftIDs, tt.args.beaconHeight, tt.args.mintNftRequireAmount)
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
		})
	}
}
