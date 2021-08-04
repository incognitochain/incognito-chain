package pdex

import (
	"encoding/json"
	"fmt"
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
	initNfctHash, err := common.Hash{}.NewHashFromStr(initNfctID)
	assert.Nil(t, err)
	nfctHash, err := common.Hash{}.NewHashFromStr(nfctID)
	assert.Nil(t, err)
	thirdTxHash, err := common.Hash{}.NewHashFromStr("bbb")
	assert.Nil(t, err)
	fourthTxHash, err := common.Hash{}.NewHashFromStr("ccc")
	assert.Nil(t, err)

	// first contribution tx
	firstContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		token0ID.String(), 100, 20000,
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
			*token0ID, *firstTxHash, 100, 20000, 1,
		),
		"pair_hash")
	waitingContributionInst := instruction.NewWaitingAddLiquidityWithValue(*waitingContributionStateDB)
	waitingContributionInstBytes, err := json.Marshal(waitingContributionInst)
	//

	// return contribution tx by sanity
	refundContributionSanityMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		token0ID.String(), 200, 20000,
	)
	assert.Nil(t, err)
	refundContributionSanityTx := &metadataMocks.Transaction{}
	refundContributionSanityTx.On("GetMetadata").Return(refundContributionSanityMetaData)
	refundContributionSanityTx.On("GetValidationEnv").Return(valEnv)
	refundContributionSanityTx.On("Hash").Return(secondTxHash)
	refundContributionSanityState0 := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0, validOTAReceiver1,
			*token0ID, *firstTxHash, 100, 20000, 1,
		),
		"pair_hash")
	refundContributionSanityInst0 := instruction.NewRefundAddLiquidityWithValue(*refundContributionSanityState0)
	refundContributionSanityInstBytes0, err := json.Marshal(refundContributionSanityInst0)
	refundContributionSanityState1 := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0, validOTAReceiver1,
			*token0ID, *secondTxHash, 200, 20000, 1,
		),
		"pair_hash")
	refundContributionSanityInst1 := instruction.NewRefundAddLiquidityWithValue(*refundContributionSanityState1)
	refundContributionSanityInstBytes1, err := json.Marshal(refundContributionSanityInst1)
	//

	// match contribution
	matchContributionMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		token1ID.String(), 400, 20000,
	)
	assert.Nil(t, err)
	matchContributionTx := &metadataMocks.Transaction{}
	matchContributionTx.On("GetMetadata").Return(matchContributionMetaData)
	matchContributionTx.On("GetValidationEnv").Return(valEnv)
	matchContributionTx.On("Hash").Return(secondTxHash)
	matchContributionState := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0, validOTAReceiver1,
			*token1ID, *secondTxHash, 400, 20000, 1,
		),
		"pair_hash")
	matchContributionInst := instruction.NewMatchAddLiquidityWithValue(*matchContributionState, poolPairID, *initNfctHash)
	matchContributionInstBytes, err := json.Marshal(matchContributionInst)
	//

	// refund contributions by amount
	refundContributionAmountMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		poolPairID, "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		token1ID.String(), 0, 20000,
	)
	refundContributionAmountTx := &metadataMocks.Transaction{}
	refundContributionAmountTx.On("GetMetadata").Return(refundContributionAmountMetaData)
	refundContributionAmountTx.On("GetValidationEnv").Return(valEnv)
	refundContributionAmountTx.On("Hash").Return(fourthTxHash)

	refundContributionAmount0State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0, validOTAReceiver1,
			*token0ID, *thirdTxHash, 50, 20000, 1,
		),
		"pair_hash")
	refundContributionAmount0Inst := instruction.NewRefundAddLiquidityWithValue(*refundContributionAmount0State)
	refundContributionAmount0InstBytes, err := json.Marshal(refundContributionAmount0Inst)
	refundContributionAmount1State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0, validOTAReceiver1,
			*token1ID, *fourthTxHash, 0, 20000, 1,
		),
		"pair_hash")
	refundContributionAmount1Inst := instruction.NewRefundAddLiquidityWithValue(*refundContributionAmount1State)
	refundContributionAmount1InstBytes, err := json.Marshal(refundContributionAmount1Inst)
	//

	// match and return contribution
	matchAndReturnContributionMetaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		poolPairID, "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		token1ID.String(), 200, 20000,
	)
	matchAndReturnContributionTx := &metadataMocks.Transaction{}
	matchAndReturnContributionTx.On("GetMetadata").Return(matchAndReturnContributionMetaData)
	matchAndReturnContributionTx.On("GetValidationEnv").Return(valEnv)
	matchAndReturnContributionTx.On("Hash").Return(fourthTxHash)

	matchAndReturnContribution0State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0, validOTAReceiver1,
			*token0ID, *thirdTxHash, 50, 20000, 1,
		),
		"pair_hash")
	matchAndReturnContritubtion0Inst := instruction.NewMatchAndReturnAddLiquidityWithValue(
		*matchAndReturnContribution0State, 100, 0, 200, 0, *token1ID, *nfctHash)
	matchAndReturnContritubtion0InstBytes, err := json.Marshal(matchAndReturnContritubtion0Inst)
	matchAndReturnContribution1State := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			poolPairID, validOTAReceiver0, validOTAReceiver1,
			*token1ID, *fourthTxHash, 200, 20000, 1,
		),
		"pair_hash")
	matchAndReturnContritubtion1Inst := instruction.NewMatchAndReturnAddLiquidityWithValue(
		*matchAndReturnContribution1State, 100, 0, 50, 0, *token0ID, *nfctHash)
	matchAndReturnContritubtion1InstBytes, err := json.Marshal(matchAndReturnContritubtion1Inst)
	//

	type fields struct {
		stateProducerBase stateProducerBase
	}
	type args struct {
		txs                  []metadata.Transaction
		beaconHeight         uint64
		poolPairs            map[string]*PoolPairState
		waitingContributions map[string]rawdbv2.Pdexv3Contribution
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		want1   map[string]*PoolPairState
		want2   map[string]rawdbv2.Pdexv3Contribution
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
					*token0ID, *firstTxHash, 100, 20000, 1,
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
						*token0ID, *firstTxHash, 100, 20000, 1,
					),
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
						*token0ID, *firstTxHash, 100, 20000, 1,
					),
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
						1,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(800), 20000,
					),
					shares: map[string]*Share{
						initNfctID: &Share{
							amount:                  200,
							tradingFees:             map[string]uint64{},
							lastUpdatedBeaconHeight: 11,
						},
					},
					orderbook: Orderbook{[]*Order{}},
				},
			},
			want2:   map[string]rawdbv2.Pdexv3Contribution{},
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
							1,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						shares: map[string]*Share{
							initNfctID: &Share{
								amount:                  200,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 11,
							},
						},
					},
				},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						poolPairID, validOTAReceiver0, validOTAReceiver1,
						*token0ID, *thirdTxHash, 50, 20000, 1,
					),
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
						1,
						big.NewInt(0).SetUint64(200),
						big.NewInt(0).SetUint64(800), 20000,
					),
					shares: map[string]*Share{
						initNfctID: &Share{
							amount:                  200,
							tradingFees:             map[string]uint64{},
							lastUpdatedBeaconHeight: 11,
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
							1,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						shares: map[string]*Share{
							initNfctID: &Share{
								amount:                  200,
								tradingFees:             map[string]uint64{},
								lastUpdatedBeaconHeight: 11,
							},
						},
					},
				},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair_hash": *rawdbv2.NewPdexv3ContributionWithValue(
						poolPairID, validOTAReceiver0, validOTAReceiver1,
						*token0ID, *thirdTxHash, 50, 20000, 1,
					),
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
						2,
						big.NewInt(0).SetUint64(300),
						big.NewInt(0).SetUint64(1200), 20000,
					),
					shares: map[string]*Share{
						initNfctID: &Share{
							amount:                  200,
							tradingFees:             map[string]uint64{},
							lastUpdatedBeaconHeight: 11,
						},
						nfctHash.String(): &Share{
							amount:                  100,
							tradingFees:             map[string]uint64{},
							lastUpdatedBeaconHeight: 11,
						},
					},
				},
			},
			want2:   map[string]rawdbv2.Pdexv3Contribution{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &stateProducerV2{
				stateProducerBase: tt.fields.stateProducerBase,
			}
			got, got1, got2, err := sp.addLiquidity(tt.args.txs, tt.args.beaconHeight, tt.args.poolPairs, tt.args.waitingContributions)
			if (err != nil) != tt.wantErr {
				t.Errorf("stateProducerV2.addLiquidity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateProducerV2.addLiquidity() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				for k, v := range got1 {
					fmt.Println("v.orderbook", v.orderbook)
					fmt.Println("tt.want1[k].orderbook", tt.want1[k].orderbook)
				}
				t.Errorf("stateProducerV2.addLiquidity() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("stateProducerV2.addLiquidity() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
