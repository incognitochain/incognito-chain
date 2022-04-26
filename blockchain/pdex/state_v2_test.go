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
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataMocks "github.com/incognitochain/incognito-chain/metadata/common/mocks"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/stretchr/testify/assert"
)

func Test_stateV2_BuildInstructions(t *testing.T) {

	config.AbortParam()
	config.Param().PDexParams.Pdexv3BreakPointHeight = 1
	config.Param().PDexParams.ProtocolFundAddress = "12svfkP6w5UDJDSCwqH978PvqiqBxKmUnA9em9yAYWYJVRv7wuXY1qhhYpPAm4BDz2mLbFrRmdK3yRhnTqJCZXKHUmoi7NV83HCH2YFpctHNaDdkSiQshsjw2UFUuwdEvcidgaKmF3VJpY5f8RdN"
	config.Param().EpochParam.NumberOfBlockInEpoch = 50

	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)
	firstTxHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)
	secondTxHash, err := common.Hash{}.NewHashFromStr("111111")
	assert.Nil(t, err)
	thirdTxHash, err := common.Hash{}.NewHashFromStr("111000")
	assert.Nil(t, err)
	fourthTxHash, err := common.Hash{}.NewHashFromStr("111001")
	assert.Nil(t, err)

	nftHash, err := common.Hash{}.NewHashFromStr(nftID)
	assert.Nil(t, err)
	nftHash1, err := common.Hash{}.NewHashFromStr(nftID1)
	assert.Nil(t, err)
	newNftHash, err := common.Hash{}.NewHashFromStr(newNftID)
	assert.Nil(t, err)

	txReqID, err := common.Hash{}.NewHashFromStr("1111122222")
	assert.Nil(t, err)
	otaReceiver0 := privacy.OTAReceiver{}
	otaReceiver0.FromString(validOTAReceiver0)
	otaReceiver1 := privacy.OTAReceiver{}
	otaReceiver1.FromString(validOTAReceiver1)

	// first contribution tx
	firstContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0,
		token0ID.String(), nftID1, 100, 20000,
	)
	assert.Nil(t, err)
	contributionTx0 := &metadataMocks.Transaction{}
	contributionTx0.On("GetMetadata").Return(firstContributionMetadata)
	contributionTx0.On("GetMetadataType").Return(metadataCommon.Pdexv3AddLiquidityRequestMeta)
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, 1)
	contributionTx0.On("GetValidationEnv").Return(valEnv)
	contributionTx0.On("Hash").Return(firstTxHash)
	waitingContributionStateDB := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0,
			*token0ID, *firstTxHash, *nftHash1, 100, 20000, 1,
		),
		"pair_hash")
	waitingContributionInst := instruction.NewWaitingAddLiquidityWithValue(*waitingContributionStateDB)
	waitingContributionInstBytes, err := json.Marshal(waitingContributionInst)
	//

	// second contribution tx
	secondContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver1,
		token1ID.String(), nftID1, 400, 20000,
	)
	assert.Nil(t, err)
	contributionTx1 := &metadataMocks.Transaction{}
	contributionTx1.On("GetMetadata").Return(secondContributionMetadata)
	contributionTx1.On("GetMetadataType").Return(metadataCommon.Pdexv3AddLiquidityRequestMeta)
	valEnv1 := tx_generic.DefaultValEnv()
	valEnv1 = tx_generic.WithShardID(valEnv1, 1)
	contributionTx1.On("GetValidationEnv").Return(valEnv1)
	contributionTx1.On("Hash").Return(secondTxHash)
	matchContributionStateDB := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver1,
			*token1ID, *secondTxHash, *nftHash1, 400, 20000, 1,
		),
		"pair_hash")
	matchContributionInst := instruction.NewMatchAddLiquidityWithValue(*matchContributionStateDB, poolPairID)
	matchContributionInstBytes, err := json.Marshal(matchContributionInst)
	//

	// user mint nft
	acceptInst, err := instruction.NewAcceptUserMintNftWithValue(
		validOTAReceiver0, 100, 1, *nftHash, *txReqID,
	).StringSlice()
	assert.Nil(t, err)

	metaData := metadataPdexv3.NewUserMintNftRequestWithValue(validOTAReceiver0, 100)
	userMintNftTx := &metadataMocks.Transaction{}
	userMintNftTx.On("GetMetadata").Return(metaData)
	userMintNftTx.On("GetMetadataType").Return(metadataCommon.Pdexv3UserMintNftRequestMeta)
	valEnv = tx_generic.WithShardID(valEnv, 1)
	userMintNftTx.On("GetValidationEnv").Return(valEnv)
	userMintNftTx.On("Hash").Return(txReqID)

	// user mint nft
	acceptInst0, err := instruction.NewAcceptUserMintNftWithValue(
		validOTAReceiver0, 100, 1, *newNftHash, *txReqID,
	).StringSlice()
	assert.Nil(t, err)

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
		common.PRVIDStr:   validOTAReceiver1,
	}

	// withdrawLiquidity tx
	withdrawLiquidityMetadata := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
		poolPairID, nftID1, withdrawLiquidityOtaReceivers, 100,
	)
	withdrawLiquidityTx := &metadataMocks.Transaction{}
	withdrawLiquidityTx.On("GetMetadata").Return(withdrawLiquidityMetadata)
	withdrawLiquidityTx.On("GetMetadataType").Return(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta)
	valEnv3 := tx_generic.DefaultValEnv()
	valEnv3 = tx_generic.WithShardID(valEnv3, 1)
	withdrawLiquidityTx.On("GetValidationEnv").Return(valEnv3)
	withdrawLiquidityTx.On("Hash").Return(txReqID)
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

	// unstaking tx
	unstakingMetadata := metadataPdexv3.NewUnstakingRequestWithValue(
		common.PRVIDStr, nftID1, withdrawLiquidityOtaReceivers, 50,
	)
	unstakingTx := &metadataMocks.Transaction{}
	unstakingTx.On("GetMetadata").Return(unstakingMetadata)
	unstakingTx.On("GetMetadataType").Return(metadataCommon.Pdexv3UnstakingRequestMeta)
	valEnv4 := tx_generic.DefaultValEnv()
	valEnv4 = tx_generic.WithShardID(valEnv2, 1)
	unstakingTx.On("GetValidationEnv").Return(valEnv4)
	unstakingTx.On("Hash").Return(txReqID)
	unstakingInst, err := instruction.NewAcceptUnstakingWithValue(
		common.PRVCoinID, *nftHash1, 50, validOTAReceiver1, *txReqID, 1,
	).StringSlice()
	assert.Nil(t, err)
	unstakingMintNftInst, err := instruction.NewMintNftWithValue(*nftHash1, validOTAReceiver0, 1, *txReqID).
		StringSlice(strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta))
	assert.Nil(t, err)
	//

	// tradeTx with token to sell fee
	trade0RequestMetaData, _ := metadataPdexv3.NewTradeRequest(
		[]string{poolPairID}, *token0ID, 100, 380, 10,
		map[common.Hash]privacy.OTAReceiver{
			*token0ID: otaReceiver0,
			*token1ID: otaReceiver1,
		},
		metadataCommon.Pdexv3TradeRequestMeta,
	)
	/*validCoin := &coinMocks.Coin{}*/
	/*validCoin.On("GetValue").Return(uint64(200))*/
	trade0Tx := &metadataMocks.Transaction{}
	trade0Tx.On("GetMetadata").Return(trade0RequestMetaData)
	trade0Tx.On("GetMetadataType").Return(metadataCommon.Pdexv3TradeRequestMeta)
	valEnv5 := tx_generic.DefaultValEnv()
	valEnv5 = tx_generic.WithShardID(valEnv5, 1)
	trade0Tx.On("GetValidationEnv").Return(valEnv5)
	trade0Tx.On("Hash").Return(txReqID)
	trade0Tx.On("GetTxFullBurnData").Return(true, nil, nil, token0ID, nil)
	trade0Inst := instruction.NewAction(
		metadataPdexv3.AcceptedTrade{
			Receiver:   otaReceiver1,
			Amount:     380,
			TradePath:  []string{poolPairID},
			TokenToBuy: *token1ID,
			PairChanges: [][2]*big.Int{
				[2]*big.Int{big.NewInt(100), big.NewInt(-380)},
			},
			RewardEarned: []map[common.Hash]uint64{
				map[common.Hash]uint64{
					*token0ID: 10,
				},
			},
			OrderChanges: []map[string][2]*big.Int{
				map[string][2]*big.Int{},
			},
		},
		*txReqID,
		byte(valEnv5.ShardID()), // sender & receiver shard must be the same
	).StringSlice()

	// trade tx with prv fee
	trade1RequestMetaData, _ := metadataPdexv3.NewTradeRequest(
		[]string{poolPairID}, *token0ID, 100, 340, 10,
		map[common.Hash]privacy.OTAReceiver{
			*token0ID:        otaReceiver0,
			*token1ID:        otaReceiver1,
			common.PRVCoinID: otaReceiver1,
		},
		metadataCommon.Pdexv3TradeRequestMeta,
	)
	trade1Tx := &metadataMocks.Transaction{}
	trade1Tx.On("GetMetadata").Return(trade1RequestMetaData)
	trade1Tx.On("GetMetadataType").Return(metadataCommon.Pdexv3TradeRequestMeta)
	valEnv6 := tx_generic.DefaultValEnv()
	valEnv6 = tx_generic.WithShardID(valEnv6, 1)
	trade1Tx.On("GetValidationEnv").Return(valEnv6)
	trade1Tx.On("Hash").Return(txReqID)
	trade1Tx.On("GetTxFullBurnData").Return(true, nil, nil, token0ID, nil)
	trade1Inst := instruction.NewAction(
		metadataPdexv3.AcceptedTrade{
			Receiver:   otaReceiver1,
			Amount:     343,
			TradePath:  []string{poolPairID},
			TokenToBuy: *token1ID,
			PairChanges: [][2]*big.Int{
				[2]*big.Int{big.NewInt(100), big.NewInt(-343)},
			},
			RewardEarned: []map[common.Hash]uint64{
				map[common.Hash]uint64{
					*token0ID: 10,
				},
			},
			OrderChanges: []map[string][2]*big.Int{
				map[string][2]*big.Int{},
			},
		},
		*txReqID,
		byte(valEnv5.ShardID()), // sender & receiver shard must be the same
	).StringSlice()

	// add order trading direction 0
	addOrderRequest0Metadata, _ := metadataPdexv3.NewAddOrderRequest(
		*token0ID, poolPairID, 10, 35,
		map[common.Hash]privacy.OTAReceiver{
			*token0ID: otaReceiver0,
			*token1ID: otaReceiver1,
		},
		*nftHash1, metadataCommon.Pdexv3AddOrderRequestMeta,
	)
	addOrder0Tx := &metadataMocks.Transaction{}
	addOrder0Tx.On("GetMetadata").Return(addOrderRequest0Metadata)
	addOrder0Tx.On("GetMetadataType").Return(metadataCommon.Pdexv3AddOrderRequestMeta)
	valEnv7 := tx_generic.DefaultValEnv()
	valEnv7 = tx_generic.WithShardID(valEnv7, 1)
	addOrder0Tx.On("GetValidationEnv").Return(valEnv7)
	addOrder0Tx.On("Hash").Return(firstTxHash)
	addOrder0Inst := instruction.NewAction(
		metadataPdexv3.AcceptedAddOrder{
			PoolPairID:     poolPairID,
			OrderID:        firstTxHash.String(),
			NftID:          *nftHash1,
			Token0Rate:     10,
			Token1Rate:     35,
			Token0Balance:  10,
			Token1Balance:  0,
			TradeDirection: 0,
			Receiver:       [2]string{validOTAReceiver0, validOTAReceiver1},
		},
		*firstTxHash,
		byte(valEnv7.ShardID()),
	).StringSlice()

	// add order trading direction 1
	addOrderRequest1Metadata, _ := metadataPdexv3.NewAddOrderRequest(
		*token1ID, poolPairID, 40, 9,
		map[common.Hash]privacy.OTAReceiver{
			*token0ID: otaReceiver0,
			*token1ID: otaReceiver1,
		},
		*nftHash1, metadataCommon.Pdexv3AddOrderRequestMeta,
	)
	addOrder1Tx := &metadataMocks.Transaction{}
	addOrder1Tx.On("GetMetadata").Return(addOrderRequest1Metadata)
	addOrder1Tx.On("GetMetadataType").Return(metadataCommon.Pdexv3AddOrderRequestMeta)
	valEnv8 := tx_generic.DefaultValEnv()
	valEnv8 = tx_generic.WithShardID(valEnv8, 1)
	addOrder1Tx.On("GetValidationEnv").Return(valEnv8)
	addOrder1Tx.On("Hash").Return(secondTxHash)
	addOrder1Inst := instruction.NewAction(
		metadataPdexv3.AcceptedAddOrder{
			PoolPairID:     poolPairID,
			OrderID:        secondTxHash.String(),
			NftID:          *nftHash1,
			Token0Rate:     9,
			Token1Rate:     40,
			Token0Balance:  0,
			Token1Balance:  40,
			TradeDirection: 1,
			Receiver:       [2]string{validOTAReceiver0, validOTAReceiver1},
		},
		*secondTxHash,
		byte(valEnv8.ShardID()),
	).StringSlice()

	// withdraw staking reward
	withdrawStakingRewardMetadata, _ := metadataPdexv3.NewPdexv3WithdrawalStakingRewardRequest(
		metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta,
		common.PRVIDStr, *nftHash1,
		map[common.Hash]privacy.OTAReceiver{
			common.PRVCoinID: otaReceiver0,
			*nftHash1:        otaReceiver1,
		},
	)
	withdrawStakingRewardTx := &metadataMocks.Transaction{}
	withdrawStakingRewardTx.On("GetMetadata").Return(withdrawStakingRewardMetadata)
	withdrawStakingRewardTx.On("GetMetadataType").Return(metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta)
	valEnv9 := tx_generic.DefaultValEnv()
	valEnv9 = tx_generic.WithShardID(valEnv9, 1)
	withdrawStakingRewardTx.On("GetValidationEnv").Return(valEnv9)
	withdrawStakingRewardTx.On("Hash").Return(txReqID)
	withdrawStakingRewardInst := v2utils.BuildWithdrawStakingRewardInsts(
		common.PRVIDStr, *nftHash1,
		map[common.Hash]metadataPdexv3.ReceiverInfo{
			common.PRVCoinID: metadataPdexv3.ReceiverInfo{
				Address: otaReceiver0,
				Amount:  100,
			},
		},
		1, *txReqID, metadataPdexv3.RequestAcceptedChainStatus,
	)[0]
	mintNftWithdrawStakingRewardInst, _ := instruction.
		NewMintNftWithValue(*nftHash1, validOTAReceiver1, 1, *txReqID).
		StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta))

	// withdraw lp fee
	withdrawLpFeeMetadata, _ := metadataPdexv3.NewPdexv3WithdrawalLPFeeRequest(
		metadataCommon.Pdexv3WithdrawLPFeeRequestMeta,
		poolPairID, *nftHash1,
		map[common.Hash]privacy.OTAReceiver{
			common.PRVCoinID: otaReceiver0,
			*nftHash1:        otaReceiver1,
		},
	)
	withdrawLpFeeTx := &metadataMocks.Transaction{}
	withdrawLpFeeTx.On("GetMetadata").Return(withdrawLpFeeMetadata)
	withdrawLpFeeTx.On("GetMetadataType").Return(metadataCommon.Pdexv3WithdrawLPFeeRequestMeta)
	valEnv10 := tx_generic.DefaultValEnv()
	valEnv10 = tx_generic.WithShardID(valEnv10, 1)
	withdrawLpFeeTx.On("GetValidationEnv").Return(valEnv10)
	withdrawLpFeeTx.On("Hash").Return(txReqID)
	withdrawLpFeeInst := v2utils.BuildWithdrawLPFeeInsts(
		poolPairID, *nftHash1,
		map[common.Hash]metadataPdexv3.ReceiverInfo{
			common.PRVCoinID: metadataPdexv3.ReceiverInfo{
				Address: otaReceiver0,
				Amount:  100,
			},
		}, 1, *txReqID, metadataPdexv3.RequestAcceptedChainStatus,
	)[0]
	mintNftWithdrawLpFeeInst, _ := instruction.
		NewMintNftWithValue(*nftHash1, validOTAReceiver1, 1, *txReqID).
		StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawLPFeeRequestMeta))

	// modifyParamsTx
	modifyParamMetadata, _ := metadataPdexv3.NewPdexv3ParamsModifyingRequest(
		metadataCommon.Pdexv3ModifyParamsMeta, metadataPdexv3.Pdexv3Params{
			DefaultFeeRateBPS:               20,
			FeeRateBPS:                      map[string]uint{poolPairID: 20},
			PRVDiscountPercent:              20,
			TradingProtocolFeePercent:       5,
			TradingStakingPoolRewardPercent: 5,
			PDEXRewardPoolPairsShare:        map[string]uint{poolPairID: 20},
			StakingPoolsShare:               map[string]uint{common.PRVIDStr: 5},
			StakingRewardTokens:             []common.Hash{common.PRVCoinID},
			MintNftRequireAmount:            1000,
			MaxOrdersPerNft:                 20,
			AutoWithdrawOrderLimitAmount:    20,
			MinPRVReserveTradingRate:        10,
		},
	)
	modifyParamsTx := &metadataMocks.Transaction{}
	modifyParamsTx.On("GetMetadata").Return(modifyParamMetadata)
	modifyParamsTx.On("GetMetadataType").Return(metadataCommon.Pdexv3ModifyParamsMeta)
	valEnv11 := tx_generic.DefaultValEnv()
	valEnv11 = tx_generic.WithShardID(valEnv11, 1)
	modifyParamsTx.On("GetValidationEnv").Return(valEnv11)
	modifyParamsTx.On("Hash").Return(txReqID)
	modifyParamsInst := v2utils.BuildModifyParamsInst(
		metadataPdexv3.Pdexv3Params{
			DefaultFeeRateBPS:               20,
			FeeRateBPS:                      map[string]uint{poolPairID: 20},
			PRVDiscountPercent:              20,
			TradingProtocolFeePercent:       5,
			TradingStakingPoolRewardPercent: 5,
			PDEXRewardPoolPairsShare:        map[string]uint{poolPairID: 20},
			StakingPoolsShare:               map[string]uint{common.PRVIDStr: 5},
			StakingRewardTokens:             []common.Hash{common.PRVCoinID},
			MintNftRequireAmount:            1000,
			MaxOrdersPerNft:                 20,
			AutoWithdrawOrderLimitAmount:    20,
			MinPRVReserveTradingRate:        10,
		},
		"", 1, *txReqID, metadataPdexv3.RequestAcceptedChainStatus,
	)

	// withdraw protocol fee
	withdrawProtocolFeeMetadata, _ := metadataPdexv3.NewPdexv3WithdrawalProtocolFeeRequest(
		metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta, poolPairID,
	)
	withdrawProtocolTx := &metadataMocks.Transaction{}
	withdrawProtocolTx.On("GetMetadata").Return(withdrawProtocolFeeMetadata)
	withdrawProtocolTx.On("GetMetadataType").Return(metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta)
	valEnv12 := tx_generic.DefaultValEnv()
	valEnv12 = tx_generic.WithShardID(valEnv12, 0)
	withdrawProtocolTx.On("GetValidationEnv").Return(valEnv12)
	withdrawProtocolTx.On("Hash").Return(txReqID)
	withdrawProtocolInst := v2utils.BuildWithdrawProtocolFeeInsts(
		poolPairID, config.Param().PDexParams.ProtocolFundAddress,
		map[common.Hash]uint64{common.PRVCoinID: 100}, 0, *txReqID, metadataPdexv3.RequestAcceptedChainStatus,
	)[0]

	// third contribution tx
	thirdContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0,
		token0ID.String(), nftID1, 100, 20000,
	)
	assert.Nil(t, err)
	contributionTx3 := &metadataMocks.Transaction{}
	contributionTx3.On("GetMetadata").Return(thirdContributionMetadata)
	contributionTx3.On("GetMetadataType").Return(metadataCommon.Pdexv3AddLiquidityRequestMeta)
	valEnv13 := tx_generic.DefaultValEnv()
	valEnv13 = tx_generic.WithShardID(valEnv13, 1)
	contributionTx3.On("GetValidationEnv").Return(valEnv13)
	contributionTx3.On("Hash").Return(thirdTxHash)
	waitingContribution3StateDB := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0,
			*token0ID, *thirdTxHash, *nftHash1, 100, 20000, 1,
		),
		"pair_hash")
	waitingContribution3Inst := instruction.NewWaitingAddLiquidityWithValue(*waitingContribution3StateDB)
	waitingContribution3InstBytes, err := json.Marshal(waitingContribution3Inst)
	//

	// fourth contribution tx
	fourthContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver1,
		token1ID.String(), nftID1, 400, 20000,
	)
	assert.Nil(t, err)
	contributionTx4 := &metadataMocks.Transaction{}
	contributionTx4.On("GetMetadata").Return(fourthContributionMetadata)
	contributionTx4.On("GetMetadataType").Return(metadataCommon.Pdexv3AddLiquidityRequestMeta)
	valEnv14 := tx_generic.DefaultValEnv()
	valEnv14 = tx_generic.WithShardID(valEnv14, 1)
	contributionTx4.On("GetValidationEnv").Return(valEnv14)
	contributionTx4.On("Hash").Return(fourthTxHash)
	matchContribution4StateDB := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver1,
			*token1ID, *fourthTxHash, *nftHash1, 400, 20000, 1,
		),
		"pair_hash")
	matchContribution4Inst := instruction.NewMatchAddLiquidityWithValue(*matchContribution4StateDB, newPoolPairID)
	matchContribution4InstBytes, err := json.Marshal(matchContribution4Inst)

	// with draw order
	withdrawOrderInst := instruction.NewAction(
		&metadataPdexv3.AcceptedWithdrawOrder{
			PoolPairID: poolPairID,
			OrderID:    txReqID.String(),
			Receiver:   otaReceiver1,
			TokenID:    *token1ID,
			Amount:     40,
		},
		*txReqID,
		byte(0),
	).StringSlice()

	//

	type fields struct {
		stateBase                   stateBase
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		params                      *Params
		stakingPoolStates           map[string]*StakingPoolState
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
		want               [][]string
		wantErr            bool
	}{
		{
			name: "2 valid txs",
			fields: fields{
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				poolPairs:                   map[string]*PoolPairState{},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				params: &Params{
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PRVDiscountPercent:              25,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
					PDEXRewardPoolPairsShare:        map[string]uint{},
					StakingPoolsShare:               map[string]uint{},
					StakingRewardTokens:             []common.Hash{},
					MintNftRequireAmount:            1000000000,
					MaxOrdersPerNft:                 10,
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
							*token0ID, *token1ID,
							200, 0, 100, 400,
							big.NewInt(200), big.NewInt(800),
							20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID1: &Share{
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
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				params: &Params{
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PRVDiscountPercent:              25,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
					PDEXRewardPoolPairsShare:        map[string]uint{},
					StakingPoolsShare:               map[string]uint{},
					StakingRewardTokens:             []common.Hash{},
					MintNftRequireAmount:            1000000000,
					MaxOrdersPerNft:                 10,
				},
			},
			args: args{
				env: &stateEnvironment{
					prevBeaconHeight: 10,
					listTxs: map[byte][]metadataCommon.Transaction{
						1: []metadataCommon.Transaction{
							contributionTx0, contributionTx1,
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
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedChainStatus,
					string(matchContributionInstBytes),
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
				params: &Params{
					MintNftRequireAmount: 100,
				},
			},
			fieldsAfterProcess: fields{
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				nftIDs: map[string]uint64{
					nftID: 100,
				},
				params: &Params{
					MintNftRequireAmount: 100,
				},
			},
			args: args{
				env: &stateEnvironment{
					prevBeaconHeight: 10,
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
				params: &Params{},
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
				params: &Params{},
				stakingPoolStates: map[string]*StakingPoolState{
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
			},
			args: args{
				env: &stateEnvironment{
					prevBeaconHeight: 10,
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
							*token0ID, *token1ID, 300, 0, 150, 600,
							big.NewInt(0).SetUint64(300),
							big.NewInt(0).SetUint64(1200), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID1: &Share{
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
					nftID1: 100,
				},
				params: &Params{},
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
							*token0ID, *token1ID, 200, 0, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees: map[common.Hash]uint64{
							common.PRVCoinID: 0,
							*token0ID:        0,
							*token1ID:        0,
						},
						shares: map[string]*Share{
							nftID1: &Share{
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
				params: &Params{},
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
					prevBeaconHeight: 20,
					listTxs: map[byte][]metadataCommon.Transaction{
						1: []metadataCommon.Transaction{withdrawLiquidityTx},
					},
				},
			},
			want:    [][]string{acceptWithdrawLiquidityInst0, acceptWithdrawLiquidityInst1, withdrawLiquidityMintNftInst},
			wantErr: false,
		},
		{
			name: "Accept unstaking tx",
			fields: fields{
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				params: NewParams(),
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 150,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity:           150,
								rewards:             map[common.Hash]uint64{},
								lastRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
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
								liquidity:           100,
								rewards:             map[common.Hash]uint64{},
								lastRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						rewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
				params: NewParams(),
			},
			args: args{
				env: &stateEnvironment{
					prevBeaconHeight: 20,
					listTxs: map[byte][]metadataCommon.Transaction{
						1: []metadataCommon.Transaction{unstakingTx},
					},
				},
			},
			want:    [][]string{unstakingInst, unstakingMintNftInst},
			wantErr: false,
		},
		{
			name: "Full txs type",
			fields: fields{
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair": *rawdbv2.NewPdexv3ContributionWithValue(
						"pool_pair_id", validOTAReceiver0,
						common.PRVCoinID, common.PRVCoinID, *nftHash1, 100,
						20000, 1,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 2000, 0, 1000, 4000,
							big.NewInt(0).SetUint64(2000),
							big.NewInt(0).SetUint64(8000), 20000,
						),
						lpFeesPerShare: map[common.Hash]*big.Int{
							common.PRVCoinID: big.NewInt(100),
						},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees: map[common.Hash]uint64{
							common.PRVCoinID: 100,
						},
						stakingPoolFees: map[common.Hash]uint64{
							common.PRVCoinID: 100,
						},
						shares: map[string]*Share{
							nftID1: &Share{
								amount: 2000,
								tradingFees: map[common.Hash]uint64{
									common.PRVCoinID: 100,
								},
								lastLPFeesPerShare: map[common.Hash]*big.Int{
									common.PRVCoinID: big.NewInt(100),
								},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderRewards: map[string]*OrderReward{},
						makingVolume: map[common.Hash]*MakingVolume{},
						orderbook: Orderbook{[]*Order{
							rawdbv2.NewPdexv3OrderWithValue(
								txReqID.String(),
								*nftHash1,
								10, 40, 0, 40, 0,
								[2]string{
									validOTAReceiver0,
									validOTAReceiver1,
								},
							),
						}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
					poolPairPRV: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							common.PRVCoinID, *token0ID, 200, 0, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID1: &Share{
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
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				params: &Params{
					DefaultFeeRateBPS:               30,
					PRVDiscountPercent:              25,
					TradingProtocolFeePercent:       10,
					TradingStakingPoolRewardPercent: 10,
					FeeRateBPS:                      map[string]uint{},
					PDEXRewardPoolPairsShare:        map[string]uint{},
					StakingPoolsShare: map[string]uint{
						common.PRVIDStr: 10,
					},
					StakingRewardTokens:          []common.Hash{},
					MaxOrdersPerNft:              10,
					MinPRVReserveTradingRate:     1,
					AutoWithdrawOrderLimitAmount: 10,
					MintNftRequireAmount:         100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 150,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity: 150,
								rewards: map[common.Hash]uint64{
									common.PRVCoinID: 100,
								},
								lastRewardsPerShare: map[common.Hash]*big.Int{
									common.PRVCoinID: big.NewInt(100),
								},
							},
						},
						rewardsPerShare: map[common.Hash]*big.Int{
							common.PRVCoinID: big.NewInt(100),
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair": *rawdbv2.NewPdexv3ContributionWithValue(
						"pool_pair_id", validOTAReceiver0,
						common.PRVCoinID, common.PRVCoinID, *nftHash1, 100,
						20000, 1,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 1900, 0, 1150, 3077,
							big.NewInt(0).SetUint64(2100),
							big.NewInt(0).SetUint64(6877), 20000,
						),
						lpFeesPerShare: map[common.Hash]*big.Int{
							common.PRVCoinID: big.NewInt(100),
							*token0ID:        big.NewInt(9473684210526314),
						},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees: map[common.Hash]uint64{
							common.PRVCoinID: 0,
							*token0ID:        2,
						},
						stakingPoolFees: map[common.Hash]uint64{
							common.PRVCoinID: 0,
							*token0ID:        0,
							*token1ID:        0,
						},
						shares: map[string]*Share{
							nftID1: &Share{
								amount: 1900,
								tradingFees: map[common.Hash]uint64{
									common.PRVCoinID: 0,
								},
								lastLPFeesPerShare: map[common.Hash]*big.Int{
									common.PRVCoinID: big.NewInt(100),
								},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderRewards: map[string]*OrderReward{},
						makingVolume: map[common.Hash]*MakingVolume{},
						orderbook: Orderbook{[]*Order{
							rawdbv2.NewPdexv3OrderWithValue(
								txReqID.String(),
								*nftHash1,
								10, 40, 0, 0, 0,
								[2]string{
									validOTAReceiver0,
									validOTAReceiver1,
								},
							),
						}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
					poolPairPRV: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							common.PRVCoinID, *token0ID, 200, 0, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees: map[common.Hash]uint64{
							common.PRVCoinID: 0,
							*token0ID:        0,
						},
						shares: map[string]*Share{
							nftID1: &Share{
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
					newPoolPairID: &PoolPairState{
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
							nftID1: &Share{
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
				nftIDs: map[string]uint64{
					nftID1:   100,
					newNftID: 100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 200,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity: 200,
								rewards:   map[common.Hash]uint64{},
								lastRewardsPerShare: map[common.Hash]*big.Int{
									common.PRVCoinID: big.NewInt(100),
								},
							},
						},
						rewardsPerShare: map[common.Hash]*big.Int{
							common.PRVCoinID: big.NewInt(100),
						},
					},
				},
				params: &Params{
					DefaultFeeRateBPS:               20,
					FeeRateBPS:                      map[string]uint{poolPairID: 20},
					PRVDiscountPercent:              20,
					TradingProtocolFeePercent:       5,
					TradingStakingPoolRewardPercent: 5,
					PDEXRewardPoolPairsShare:        map[string]uint{poolPairID: 20},
					StakingPoolsShare:               map[string]uint{common.PRVIDStr: 5},
					StakingRewardTokens:             []common.Hash{common.PRVCoinID},
					MintNftRequireAmount:            1000,
					MaxOrdersPerNft:                 20,
					AutoWithdrawOrderLimitAmount:    20,
					MinPRVReserveTradingRate:        10,
				},
			},
			args: args{
				env: &stateEnvironment{
					prevBeaconHeight: 30,
					listTxs: map[byte][]metadataCommon.Transaction{
						1: []metadataCommon.Transaction{
							contributionTx3, contributionTx4,
							userMintNftTx,
							stakingTx,
							withdrawLiquidityTx,
							unstakingTx,
							modifyParamsTx,
							trade0Tx, trade1Tx,
							addOrder0Tx, addOrder1Tx,
							withdrawLpFeeTx, withdrawStakingRewardTx,
							withdrawProtocolTx,
						},
					},
				},
			},
			want: [][]string{ // (x, y, x', y')
				mintNftWithdrawLpFeeInst,
				withdrawLpFeeInst,    // (1000, 4000, 2000, 8000)
				withdrawProtocolInst, // (1000, 4000, 2000, 8000)
				acceptWithdrawLiquidityInst0,
				acceptWithdrawLiquidityInst1,
				withdrawLiquidityMintNftInst, // (950, 3800, 1900, 7600)
				unstakingInst,
				unstakingMintNftInst, // (950, 3800, 1900, 7600)
				mintNftWithdrawStakingRewardInst,
				withdrawStakingRewardInst,
				trade0Inst, trade1Inst,
				withdrawOrderInst, // (1150, 3077, 2100, 6877)
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionWaitingChainStatus,
					string(waitingContribution3InstBytes),
				},
				[]string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.PDEContributionMatchedChainStatus,
					string(matchContribution4InstBytes),
				}, // (1150, 3077, 2100, 6877)
				stakingInst,
				addOrder0Inst, addOrder1Inst, // (1150, 3077, 2100, 6877)
				acceptInst0,
				modifyParamsInst,
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
			if !reflect.DeepEqual(s.params, tt.fieldsAfterProcess.params) {
				t.Errorf("params = %v, want %v", s.params, tt.fieldsAfterProcess.params)
				return
			}
			if !reflect.DeepEqual(s.poolPairs, tt.fieldsAfterProcess.poolPairs) {
				t.Errorf("poolPairs = %v, want %v", s.poolPairs, tt.fieldsAfterProcess.poolPairs)
				return
			}
			if !reflect.DeepEqual(s.nftIDs, tt.fieldsAfterProcess.nftIDs) {
				t.Errorf("nftIDs = %v, want %v", s.nftIDs, tt.fieldsAfterProcess.nftIDs)
				return
			}
			if !reflect.DeepEqual(s.stakingPoolStates, tt.fieldsAfterProcess.stakingPoolStates) {
				t.Errorf("stakingPoolStates = %v, want %v", s.stakingPoolStates[common.PRVIDStr], tt.fieldsAfterProcess.stakingPoolStates[common.PRVIDStr])
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

	config.AbortParam()
	config.Param().PDexParams.Pdexv3BreakPointHeight = 1
	config.Param().PDexParams.ProtocolFundAddress = "12svfkP6w5UDJDSCwqH978PvqiqBxKmUnA9em9yAYWYJVRv7wuXY1qhhYpPAm4BDz2mLbFrRmdK3yRhnTqJCZXKHUmoi7NV83HCH2YFpctHNaDdkSiQshsjw2UFUuwdEvcidgaKmF3VJpY5f8RdN"
	config.Param().EpochParam.NumberOfBlockInEpoch = 50

	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)
	firstTxHash, err := common.Hash{}.NewHashFromStr("abc")
	assert.Nil(t, err)
	secondTxHash, err := common.Hash{}.NewHashFromStr("111111")
	assert.Nil(t, err)
	thirdTxHash, err := common.Hash{}.NewHashFromStr("111000")
	assert.Nil(t, err)
	fourthTxHash, err := common.Hash{}.NewHashFromStr("111001")
	assert.Nil(t, err)
	nftHash1, err := common.Hash{}.NewHashFromStr(nftID1)
	assert.Nil(t, err)
	newNftHash, err := common.Hash{}.NewHashFromStr(newNftID)
	assert.Nil(t, err)
	txReqID, err := common.Hash{}.NewHashFromStr("1111122222")
	assert.Nil(t, err)

	otaReceiver0 := privacy.OTAReceiver{}
	otaReceiver0.FromString(validOTAReceiver0)
	otaReceiver1 := privacy.OTAReceiver{}
	otaReceiver1.FromString(validOTAReceiver1)

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
			*token0ID, *firstTxHash, *nftHash1, 100, 20000, 1,
		),
		"pair_hash")
	waitingContributionInst := instruction.NewWaitingAddLiquidityWithValue(*waitingContributionStateDB)
	waitingContributionInstBytes, err := json.Marshal(waitingContributionInst)
	//

	// second contribution tx
	secondContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver1,
		token1ID.String(), nftID1, 400, 20000,
	)
	assert.Nil(t, err)
	contributionTx1 := &metadataMocks.Transaction{}
	contributionTx1.On("GetMetadata").Return(secondContributionMetadata)
	contributionTx1.On("GetMetadataType").Return(metadataCommon.Pdexv3AddLiquidityRequestMeta)
	valEnv1 := tx_generic.DefaultValEnv()
	valEnv1 = tx_generic.WithShardID(valEnv1, 1)
	contributionTx1.On("GetValidationEnv").Return(valEnv1)
	contributionTx1.On("Hash").Return(secondTxHash)
	matchContributionStateDB := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver1,
			*token1ID, *secondTxHash, *nftHash1, 400, 20000, 1,
		),
		"pair_hash")
	matchContributionInst := instruction.NewMatchAddLiquidityWithValue(*matchContributionStateDB, poolPairID)
	matchContributionInstBytes, err := json.Marshal(matchContributionInst)
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
	stakingInst, err := instruction.NewAcceptStakingWtihValue(
		*nftHash1, common.PRVCoinID, *txReqID, 1, 100,
	).StringSlice()
	assert.Nil(t, err)
	//

	//unstaking
	unstakingInst, err := instruction.NewAcceptUnstakingWithValue(
		common.PRVCoinID, *nftHash1, 50, validOTAReceiver1, *txReqID, 1,
	).StringSlice()
	assert.Nil(t, err)
	unstakingMintNftInst, err := instruction.NewMintNftWithValue(*nftHash1, validOTAReceiver0, 1, *txReqID).
		StringSlice(strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta))
	assert.Nil(t, err)
	//

	// withdraw lp fee inst
	withdrawLpFeeInst := v2utils.BuildWithdrawLPFeeInsts(
		poolPairID, *nftHash1,
		map[common.Hash]metadataPdexv3.ReceiverInfo{
			common.PRVCoinID: metadataPdexv3.ReceiverInfo{
				Address: otaReceiver0,
				Amount:  100,
			},
		}, 1, *txReqID, metadataPdexv3.RequestAcceptedChainStatus,
	)[0]
	mintNftWithdrawLpFeeInst, _ := instruction.
		NewMintNftWithValue(*nftHash1, validOTAReceiver1, 1, *txReqID).
		StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawLPFeeRequestMeta))

	// withdraw protocol inst
	withdrawProtocolInst := v2utils.BuildWithdrawProtocolFeeInsts(
		poolPairID, config.Param().PDexParams.ProtocolFundAddress,
		map[common.Hash]uint64{common.PRVCoinID: 100}, 0, *txReqID, metadataPdexv3.RequestAcceptedChainStatus,
	)[0]

	// accept instructions
	acceptWithdrawLiquidityInst0, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *nftHash1, *token0ID, 50, 100, validOTAReceiver1, *txReqID, 1,
	).StringSlice()
	assert.Nil(t, err)
	acceptWithdrawLiquidityInst1, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		poolPairID, *nftHash1, *token1ID, 200, 100, validOTAReceiver1, *txReqID, 1,
	).StringSlice()
	assert.Nil(t, err)
	withdrawLiquidityMintNftInst, err := instruction.NewMintNftWithValue(*nftHash1, validOTAReceiver0, 1, *txReqID).
		StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta))
	assert.Nil(t, err)

	// withdraw staking reward
	mintNftWithdrawStakingRewardInst, _ := instruction.
		NewMintNftWithValue(*nftHash1, validOTAReceiver1, 1, *txReqID).
		StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta))
	withdrawStakingRewardInst := v2utils.BuildWithdrawStakingRewardInsts(
		common.PRVIDStr, *nftHash1,
		map[common.Hash]metadataPdexv3.ReceiverInfo{
			common.PRVCoinID: metadataPdexv3.ReceiverInfo{
				Address: otaReceiver0,
				Amount:  100,
			},
		},
		1, *txReqID, metadataPdexv3.RequestAcceptedChainStatus,
	)[0]

	// trade Inst
	trade0Inst := instruction.NewAction(
		metadataPdexv3.AcceptedTrade{
			Receiver:   otaReceiver1,
			Amount:     380,
			TradePath:  []string{poolPairID},
			TokenToBuy: *token1ID,
			PairChanges: [][2]*big.Int{
				[2]*big.Int{big.NewInt(100), big.NewInt(-380)},
			},
			RewardEarned: []map[common.Hash]uint64{
				map[common.Hash]uint64{
					*token0ID: 10,
				},
			},
			OrderChanges: []map[string][2]*big.Int{
				map[string][2]*big.Int{},
			},
		},
		*txReqID,
		byte(1), // sender & receiver shard must be the same
	).StringSlice()

	trade1Inst := instruction.NewAction(
		metadataPdexv3.AcceptedTrade{
			Receiver:   otaReceiver1,
			Amount:     343,
			TradePath:  []string{poolPairID},
			TokenToBuy: *token1ID,
			PairChanges: [][2]*big.Int{
				[2]*big.Int{big.NewInt(100), big.NewInt(-343)},
			},
			RewardEarned: []map[common.Hash]uint64{
				map[common.Hash]uint64{
					*token0ID: 10,
				},
			},
			OrderChanges: []map[string][2]*big.Int{
				map[string][2]*big.Int{},
			},
		},
		*txReqID,
		byte(1), // sender & receiver shard must be the same
	).StringSlice()

	// third contribution tx
	thirdContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver0,
		token0ID.String(), nftID1, 100, 20000,
	)
	assert.Nil(t, err)
	contributionTx3 := &metadataMocks.Transaction{}
	contributionTx3.On("GetMetadata").Return(thirdContributionMetadata)
	contributionTx3.On("GetMetadataType").Return(metadataCommon.Pdexv3AddLiquidityRequestMeta)
	valEnv13 := tx_generic.DefaultValEnv()
	valEnv13 = tx_generic.WithShardID(valEnv13, 1)
	contributionTx3.On("GetValidationEnv").Return(valEnv13)
	contributionTx3.On("Hash").Return(thirdTxHash)
	waitingContribution3StateDB := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver0,
			*token0ID, *thirdTxHash, *nftHash1, 100, 20000, 1,
		),
		"pair_hash")
	waitingContribution3Inst := instruction.NewWaitingAddLiquidityWithValue(*waitingContribution3StateDB)
	waitingContribution3InstBytes, err := json.Marshal(waitingContribution3Inst)
	//

	// fourth contribution tx
	fourthContributionMetadata := metadataPdexv3.NewAddLiquidityRequestWithValue(
		"", "pair_hash",
		validOTAReceiver1,
		token1ID.String(), nftID1, 400, 20000,
	)
	assert.Nil(t, err)
	contributionTx4 := &metadataMocks.Transaction{}
	contributionTx4.On("GetMetadata").Return(fourthContributionMetadata)
	contributionTx4.On("GetMetadataType").Return(metadataCommon.Pdexv3AddLiquidityRequestMeta)
	valEnv14 := tx_generic.DefaultValEnv()
	valEnv14 = tx_generic.WithShardID(valEnv14, 1)
	contributionTx4.On("GetValidationEnv").Return(valEnv14)
	contributionTx4.On("Hash").Return(fourthTxHash)
	matchContribution4StateDB := statedb.NewPdexv3ContributionStateWithValue(
		*rawdbv2.NewPdexv3ContributionWithValue(
			"", validOTAReceiver1,
			*token1ID, *fourthTxHash, *nftHash1, 400, 20000, 1,
		),
		"pair_hash")
	matchContribution4Inst := instruction.NewMatchAddLiquidityWithValue(*matchContribution4StateDB, newPoolPairID)
	matchContribution4InstBytes, err := json.Marshal(matchContribution4Inst)

	// add order inst
	addOrder0Inst := instruction.NewAction(
		metadataPdexv3.AcceptedAddOrder{
			PoolPairID:     poolPairID,
			OrderID:        firstTxHash.String(),
			NftID:          *nftHash1,
			Token0Rate:     10,
			Token1Rate:     35,
			Token0Balance:  10,
			Token1Balance:  0,
			TradeDirection: 0,
			Receiver:       [2]string{validOTAReceiver0, validOTAReceiver1},
		},
		*firstTxHash,
		byte(1),
	).StringSlice()
	addOrder1Inst := instruction.NewAction(
		metadataPdexv3.AcceptedAddOrder{
			PoolPairID:     poolPairID,
			OrderID:        secondTxHash.String(),
			NftID:          *nftHash1,
			Token0Rate:     9,
			Token1Rate:     40,
			Token0Balance:  0,
			Token1Balance:  40,
			TradeDirection: 1,
			Receiver:       [2]string{validOTAReceiver0, validOTAReceiver1},
		},
		*secondTxHash,
		byte(1),
	).StringSlice()

	// user mint nft
	acceptInst0, err := instruction.NewAcceptUserMintNftWithValue(
		validOTAReceiver0, 100, 1, *newNftHash, *txReqID,
	).StringSlice()
	assert.Nil(t, err)

	// modify params
	modifyParamsInst := v2utils.BuildModifyParamsInst(
		metadataPdexv3.Pdexv3Params{
			DefaultFeeRateBPS:               20,
			FeeRateBPS:                      map[string]uint{poolPairID: 20},
			PRVDiscountPercent:              20,
			TradingProtocolFeePercent:       5,
			TradingStakingPoolRewardPercent: 5,
			PDEXRewardPoolPairsShare:        map[string]uint{poolPairID: 20},
			StakingPoolsShare:               map[string]uint{common.PRVIDStr: 5},
			StakingRewardTokens:             []common.Hash{common.PRVCoinID},
			MintNftRequireAmount:            1000,
			MaxOrdersPerNft:                 20,
			AutoWithdrawOrderLimitAmount:    20,
			MinPRVReserveTradingRate:        10,
		},
		"", 1, *txReqID, metadataPdexv3.RequestAcceptedChainStatus,
	)

	// with draw order
	withdrawOrderInst := instruction.NewAction(
		&metadataPdexv3.AcceptedWithdrawOrder{
			PoolPairID: poolPairID,
			OrderID:    txReqID.String(),
			Receiver:   otaReceiver1,
			TokenID:    *token1ID,
			Amount:     40,
		},
		*txReqID,
		byte(0),
	).StringSlice()

	type fields struct {
		stateBase                   stateBase
		waitingContributions        map[string]rawdbv2.Pdexv3Contribution
		deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution
		poolPairs                   map[string]*PoolPairState
		params                      *Params
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
			name: "[Add Liquidity] 2 insts",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				params:                      &Params{},
			},
			args: args{
				env: &stateEnvironment{
					prevBeaconHeight: 10,
					stateDB:          sDB,
					beaconInstructions: [][]string{
						[]string{
							strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
							common.PDEContributionWaitingChainStatus,
							string(waitingContributionInstBytes),
						},
						[]string{
							strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
							common.PDEContributionMatchedChainStatus,
							string(matchContributionInstBytes),
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID,
							200, 0, 100, 400,
							big.NewInt(200), big.NewInt(800),
							20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID1: &Share{
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
				params:    &Params{},
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
				params:                      &Params{},
			},
			args: args{
				env: &stateEnvironment{
					prevBeaconHeight:   10,
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
				params:    &Params{},
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
				params:                      &Params{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
			},
			args: args{
				env: &stateEnvironment{
					prevBeaconHeight:   10,
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
				params:    &Params{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			wantErr: false,
		},
		{
			name: "Accept staking inst",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				params:    &Params{},
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
					prevBeaconHeight:   10,
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
				params:    &Params{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 100,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity:           100,
								rewards:             map[common.Hash]uint64{},
								lastRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Accept unstaking inst",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				params:    &Params{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 150,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity:           150,
								rewards:             map[common.Hash]uint64{},
								lastRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
					},
				},
			},
			args: args{
				env: &stateEnvironment{
					prevBeaconHeight:   20,
					stateDB:            sDB,
					beaconInstructions: [][]string{unstakingInst},
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
				params:    &Params{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 100,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity:           100,
								rewards:             map[common.Hash]uint64{},
								lastRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Full txs type",
			fields: fields{
				stateBase: stateBase{},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair": *rawdbv2.NewPdexv3ContributionWithValue(
						"pool_pair_id", validOTAReceiver0,
						common.PRVCoinID, common.PRVCoinID, *nftHash1, 100,
						20000, 1,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 2000, 0, 1000, 4000,
							big.NewInt(0).SetUint64(2000),
							big.NewInt(0).SetUint64(8000), 20000,
						),
						lpFeesPerShare: map[common.Hash]*big.Int{
							common.PRVCoinID: big.NewInt(100),
						},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees: map[common.Hash]uint64{
							common.PRVCoinID: 100,
						},
						stakingPoolFees: map[common.Hash]uint64{
							common.PRVCoinID: 100,
						},
						shares: map[string]*Share{
							nftID1: &Share{
								amount: 2000,
								tradingFees: map[common.Hash]uint64{
									common.PRVCoinID: 100,
								},
								lastLPFeesPerShare: map[common.Hash]*big.Int{
									common.PRVCoinID: big.NewInt(100),
								},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderRewards: map[string]*OrderReward{},
						makingVolume: map[common.Hash]*MakingVolume{},
						orderbook: Orderbook{[]*Order{
							rawdbv2.NewPdexv3OrderWithValue(
								txReqID.String(),
								*nftHash1,
								10, 40, 0, 40, 0,
								[2]string{
									validOTAReceiver0,
									validOTAReceiver1,
								},
							),
						}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
					poolPairPRV: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							common.PRVCoinID, *token0ID, 200, 0, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID1: &Share{
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
				nftIDs: map[string]uint64{
					nftID1: 100,
				},
				params: &Params{
					DefaultFeeRateBPS:               30,
					PRVDiscountPercent:              25,
					TradingProtocolFeePercent:       10,
					TradingStakingPoolRewardPercent: 10,
					FeeRateBPS:                      map[string]uint{},
					PDEXRewardPoolPairsShare:        map[string]uint{},
					StakingPoolsShare: map[string]uint{
						common.PRVIDStr: 10,
					},
					StakingRewardTokens:          []common.Hash{},
					MaxOrdersPerNft:              10,
					MinPRVReserveTradingRate:     1,
					AutoWithdrawOrderLimitAmount: 10,
					MintNftRequireAmount:         100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 150,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity: 150,
								rewards: map[common.Hash]uint64{
									common.PRVCoinID: 100,
								},
								lastRewardsPerShare: map[common.Hash]*big.Int{
									common.PRVCoinID: big.NewInt(100),
								},
							},
						},
						rewardsPerShare: map[common.Hash]*big.Int{
							common.PRVCoinID: big.NewInt(100),
						},
					},
				},
			},
			args: args{
				env: &stateEnvironment{
					prevBeaconHeight: 30,
					stateDB:          sDB,
					beaconInstructions: [][]string{
						mintNftWithdrawLpFeeInst,
						withdrawLpFeeInst,
						withdrawProtocolInst,
						acceptWithdrawLiquidityInst0,
						acceptWithdrawLiquidityInst1,
						withdrawLiquidityMintNftInst,
						unstakingInst,
						unstakingMintNftInst,
						mintNftWithdrawStakingRewardInst,
						withdrawStakingRewardInst,
						trade0Inst, trade1Inst,
						withdrawOrderInst,
						[]string{
							strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
							common.PDEContributionWaitingChainStatus,
							string(waitingContribution3InstBytes),
						},
						[]string{
							strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
							common.PDEContributionMatchedChainStatus,
							string(matchContribution4InstBytes),
						},
						stakingInst,
						addOrder0Inst, addOrder1Inst,
						acceptInst0,
						modifyParamsInst,
					},
				},
			},
			fieldsAfterProcess: fields{
				stateBase: stateBase{},
				waitingContributions: map[string]rawdbv2.Pdexv3Contribution{
					"pair": *rawdbv2.NewPdexv3ContributionWithValue(
						"pool_pair_id", validOTAReceiver0,
						common.PRVCoinID, common.PRVCoinID, *nftHash1, 100,
						20000, 1,
					),
				},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				producer:                    stateProducerV2{},
				processor:                   stateProcessorV2{},
				poolPairs: map[string]*PoolPairState{
					poolPairID: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							*token0ID, *token1ID, 1900, 0, 1150, 3077,
							big.NewInt(0).SetUint64(2100),
							big.NewInt(0).SetUint64(6877), 20000,
						),
						lpFeesPerShare: map[common.Hash]*big.Int{
							common.PRVCoinID: big.NewInt(100),
							*token0ID:        big.NewInt(9473684210526314),
						},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees: map[common.Hash]uint64{
							common.PRVCoinID: 0,
							*token0ID:        2,
						},
						stakingPoolFees: map[common.Hash]uint64{
							common.PRVCoinID: 0,
							*token0ID:        0,
							*token1ID:        0,
						},
						shares: map[string]*Share{
							nftID1: &Share{
								amount: 1900,
								tradingFees: map[common.Hash]uint64{
									common.PRVCoinID: 0,
								},
								lastLPFeesPerShare: map[common.Hash]*big.Int{
									common.PRVCoinID: big.NewInt(100),
								},
								lastLmRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderRewards: map[string]*OrderReward{},
						makingVolume: map[common.Hash]*MakingVolume{},
						orderbook: Orderbook{[]*Order{
							rawdbv2.NewPdexv3OrderWithValue(
								firstTxHash.String(),
								*nftHash1,
								10, 35, 10, 0, 0,
								[2]string{
									validOTAReceiver0,
									validOTAReceiver1,
								},
							),
							rawdbv2.NewPdexv3OrderWithValue(
								secondTxHash.String(),
								*nftHash1,
								9, 40, 0, 40, 1,
								[2]string{
									validOTAReceiver0,
									validOTAReceiver1,
								},
							),
						}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
					poolPairPRV: &PoolPairState{
						state: *rawdbv2.NewPdexv3PoolPairWithValue(
							common.PRVCoinID, *token0ID, 200, 0, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees: map[common.Hash]uint64{
							common.PRVCoinID: 0,
							*token0ID:        0,
						},
						shares: map[string]*Share{
							nftID1: &Share{
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
					newPoolPairID: &PoolPairState{
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
							nftID1: &Share{
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
				nftIDs: map[string]uint64{
					nftID1:   100,
					newNftID: 100,
				},
				stakingPoolStates: map[string]*StakingPoolState{
					common.PRVIDStr: &StakingPoolState{
						liquidity: 200,
						stakers: map[string]*Staker{
							nftID1: &Staker{
								liquidity: 200,
								rewards:   map[common.Hash]uint64{},
								lastRewardsPerShare: map[common.Hash]*big.Int{
									common.PRVCoinID: big.NewInt(100),
								},
							},
						},
						rewardsPerShare: map[common.Hash]*big.Int{
							common.PRVCoinID: big.NewInt(100),
						},
					},
				},
				params: &Params{
					DefaultFeeRateBPS:               20,
					FeeRateBPS:                      map[string]uint{poolPairID: 20},
					PRVDiscountPercent:              20,
					TradingProtocolFeePercent:       5,
					TradingStakingPoolRewardPercent: 5,
					PDEXRewardPoolPairsShare:        map[string]uint{poolPairID: 20},
					StakingPoolsShare:               map[string]uint{common.PRVIDStr: 5},
					StakingRewardTokens:             []common.Hash{common.PRVCoinID},
					MintNftRequireAmount:            1000,
					MaxOrdersPerNft:                 20,
					AutoWithdrawOrderLimitAmount:    20,
					MinPRVReserveTradingRate:        10,
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
			if !reflect.DeepEqual(s.params, tt.fieldsAfterProcess.params) {
				t.Errorf("params = %v, want %v", s.params, tt.fieldsAfterProcess.params)
				return
			}
			if !reflect.DeepEqual(s.stakingPoolStates, tt.fieldsAfterProcess.stakingPoolStates) {
				t.Errorf("stakingPoolStates = %v, want %v", s.stakingPoolStates, tt.fieldsAfterProcess.stakingPoolStates)
				return
			}
			if !reflect.DeepEqual(s.poolPairs, tt.fieldsAfterProcess.poolPairs) {
				t.Errorf("poolPairs = %v, want %v", s.poolPairs, tt.fieldsAfterProcess.poolPairs)
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
		params                      *Params
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
		params                      *Params
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
								amount: 200,
								tradingFees: map[common.Hash]uint64{
									common.PRVCoinID: 10,
								},
								lastLPFeesPerShare: map[common.Hash]*big.Int{},
							},
						},
						orderRewards: map[string]*OrderReward{},
						makingVolume: map[common.Hash]*MakingVolume{},
						orderbook:    Orderbook{[]*Order{}},
					},
				},
				params: &Params{},
			},
			args: args{
				env: &stateEnvironment{
					stateDB: sDB,
				},
				stateChange: &StateChange{
					PoolPairs: map[string]*v2utils.PoolPairChange{
						poolPairID: &v2utils.PoolPairChange{
							IsChanged: true,
							Shares: map[string]*v2utils.ShareChange{
								nftID: &v2utils.ShareChange{
									IsChanged:          true,
									TradingFees:        map[string]bool{},
									LastLPFeesPerShare: map[string]bool{},
								},
							},
							LpFeesPerShare:  map[string]bool{},
							ProtocolFees:    map[string]bool{},
							StakingPoolFees: map[string]bool{},
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
		params                      *Params
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
							*token0ID, *token1ID, 200, 0, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID1: &Share{
								amount: 200,
								tradingFees: map[common.Hash]uint64{
									common.PRVCoinID: 100,
								},
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
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PDEXRewardPoolPairsShare:        map[string]uint{},
					PRVDiscountPercent:              25,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
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
					params: &Params{
						DefaultFeeRateBPS: 30,
						FeeRateBPS: map[string]uint{
							"abc": 12,
						},
						PDEXRewardPoolPairsShare:        map[string]uint{},
						PRVDiscountPercent:              25,
						TradingProtocolFeePercent:       0,
						TradingStakingPoolRewardPercent: 10,
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
					PoolPairs:    map[string]*v2utils.PoolPairChange{},
					StakingPools: map[string]*v2utils.StakingPoolChange{},
				},
			},
			want: &stateV2{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
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
							nftID1: &Share{
								amount: 200,
								tradingFees: map[common.Hash]uint64{
									common.PRVCoinID: 100,
								},
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
				params:            NewParams(),
				stakingPoolStates: map[string]*StakingPoolState{},
				nftIDs:            map[string]uint64{},
				producer:          stateProducerV2{},
				processor:         stateProcessorV2{},
			},
			want1: &StateChange{
				PoolPairs: map[string]*v2utils.PoolPairChange{
					poolPairID: &v2utils.PoolPairChange{
						IsChanged: true,
						Shares: map[string]*v2utils.ShareChange{
							nftID1: &v2utils.ShareChange{
								IsChanged: true,
								TradingFees: map[string]bool{
									common.PRVIDStr: true,
								},
								LastLPFeesPerShare:    map[string]bool{},
								LastLmRewardsPerShare: map[string]bool{},
							},
						},
						OrderIDs:          map[string]bool{},
						LpFeesPerShare:    map[string]bool{},
						LmRewardsPerShare: map[string]bool{},
						ProtocolFees:      map[string]bool{},
						StakingPoolFees:   map[string]bool{},
						MakingVolume:      map[string]*v2utils.MakingVolumeChange{},
						OrderRewards:      map[string]*v2utils.OrderRewardChange{},
						LmLockedShare:     map[string]map[uint64]bool{},
					},
				},
				StakingPools: map[string]*v2utils.StakingPoolChange{},
			},
			wantErr: false,
		},
		{
			name: "Only poolpair order rewards",
			fields: fields{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
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
						shares:            map[string]*Share{},
						orderRewards: map[string]*OrderReward{
							nftID: {
								uncollectedRewards: Reward{
									*token0ID: 100,
									*token1ID: 0,
								},
							},
							nftID1: {
								uncollectedRewards: Reward{
									common.PRVCoinID: 50,
								},
							},
						},
						makingVolume:  map[common.Hash]*MakingVolume{},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				params: &Params{
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PDEXRewardPoolPairsShare:        map[string]uint{},
					PRVDiscountPercent:              25,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
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
								*token0ID, *token1ID, 200, 0, 100, 400,
								big.NewInt(0).SetUint64(200),
								big.NewInt(0).SetUint64(800), 20000,
							),
							lpFeesPerShare:  map[common.Hash]*big.Int{},
							protocolFees:    map[common.Hash]uint64{},
							stakingPoolFees: map[common.Hash]uint64{},
							shares:          map[string]*Share{},
							orderRewards: map[string]*OrderReward{
								nftID: {
									uncollectedRewards: Reward{
										*token0ID: 100,
										*token1ID: 200,
									},
								},
							},
							makingVolume:  map[common.Hash]*MakingVolume{},
							orderbook:     Orderbook{[]*Order{}},
							lmLockedShare: map[string]map[uint64]uint64{},
						},
					},
					params: &Params{
						DefaultFeeRateBPS: 30,
						FeeRateBPS: map[string]uint{
							"abc": 12,
						},
						PDEXRewardPoolPairsShare:        map[string]uint{},
						PRVDiscountPercent:              25,
						TradingProtocolFeePercent:       0,
						TradingStakingPoolRewardPercent: 10,
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
					PoolPairs:    map[string]*v2utils.PoolPairChange{},
					StakingPools: map[string]*v2utils.StakingPoolChange{},
				},
			},
			want: &stateV2{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
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
						shares:            map[string]*Share{},
						orderRewards: map[string]*OrderReward{
							nftID: {
								uncollectedRewards: Reward{
									*token0ID: 100,
									*token1ID: 0,
								},
							},
							nftID1: {
								uncollectedRewards: Reward{
									common.PRVCoinID: 50,
								},
							},
						},
						makingVolume:  map[common.Hash]*MakingVolume{},
						orderbook:     Orderbook{[]*Order{}},
						lmLockedShare: map[string]map[uint64]uint64{},
					},
				},
				params:            NewParams(),
				stakingPoolStates: map[string]*StakingPoolState{},
				nftIDs:            map[string]uint64{},
				producer:          stateProducerV2{},
				processor:         stateProcessorV2{},
			},
			want1: &StateChange{
				PoolPairs: map[string]*v2utils.PoolPairChange{
					poolPairID: &v2utils.PoolPairChange{
						IsChanged:         false,
						Shares:            map[string]*v2utils.ShareChange{},
						OrderIDs:          map[string]bool{},
						LpFeesPerShare:    map[string]bool{},
						LmRewardsPerShare: map[string]bool{},
						ProtocolFees:      map[string]bool{},
						StakingPoolFees:   map[string]bool{},
						MakingVolume:      map[string]*v2utils.MakingVolumeChange{},
						OrderRewards: map[string]*v2utils.OrderRewardChange{
							nftID: {
								UncollectedReward: map[string]bool{
									token1ID.String(): true,
								},
							},
							nftID1: {
								UncollectedReward: map[string]bool{
									common.PRVIDStr: true,
								},
							},
						},
						LmLockedShare: map[string]map[uint64]bool{},
					},
				},
				StakingPools: map[string]*v2utils.StakingPoolChange{},
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
							*token0ID, *token1ID, 200, 0, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID1: &Share{
								amount: 200,
								tradingFees: map[common.Hash]uint64{
									common.PRVCoinID: 100,
								},
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
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PDEXRewardPoolPairsShare:        map[string]uint{},
					PRVDiscountPercent:              25,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
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
								*token0ID, *token1ID, 200, 0, 100, 400,
								big.NewInt(0).SetUint64(200),
								big.NewInt(0).SetUint64(800), 20000,
							),
							lpFeesPerShare:    map[common.Hash]*big.Int{},
							lmRewardsPerShare: map[common.Hash]*big.Int{},
							protocolFees:      map[common.Hash]uint64{},
							stakingPoolFees:   map[common.Hash]uint64{},
							shares: map[string]*Share{
								nftID1: &Share{
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
					params: &Params{
						DefaultFeeRateBPS: 30,
						FeeRateBPS: map[string]uint{
							"abc": 12,
						},
						PDEXRewardPoolPairsShare:        map[string]uint{},
						PRVDiscountPercent:              25,
						TradingProtocolFeePercent:       0,
						TradingStakingPoolRewardPercent: 10,
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
					PoolPairs:    map[string]*v2utils.PoolPairChange{},
					StakingPools: map[string]*v2utils.StakingPoolChange{},
				},
			},
			want: &stateV2{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
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
							nftID1: &Share{
								amount: 200,
								tradingFees: map[common.Hash]uint64{
									common.PRVCoinID: 100,
								},
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
				params:            NewParams(),
				stakingPoolStates: map[string]*StakingPoolState{},
				nftIDs:            map[string]uint64{},
				producer:          stateProducerV2{},
				processor:         stateProcessorV2{},
			},
			want1: &StateChange{
				PoolPairs: map[string]*v2utils.PoolPairChange{
					poolPairID: &v2utils.PoolPairChange{
						IsChanged: false,
						Shares: map[string]*v2utils.ShareChange{
							nftID1: &v2utils.ShareChange{
								IsChanged: false,
								TradingFees: map[string]bool{
									common.PRVIDStr: true,
								},
								LastLPFeesPerShare:    map[string]bool{},
								LastLmRewardsPerShare: map[string]bool{},
							},
						},
						OrderIDs:          map[string]bool{},
						LpFeesPerShare:    map[string]bool{},
						ProtocolFees:      map[string]bool{},
						StakingPoolFees:   map[string]bool{},
						MakingVolume:      map[string]*v2utils.MakingVolumeChange{},
						OrderRewards:      map[string]*v2utils.OrderRewardChange{},
						LmRewardsPerShare: map[string]bool{},
						LmLockedShare:     map[string]map[uint64]bool{},
					},
				},
				StakingPools: map[string]*v2utils.StakingPoolChange{},
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
							*token0ID, *token1ID, 200, 0, 100, 400,
							big.NewInt(0).SetUint64(200),
							big.NewInt(0).SetUint64(800), 20000,
						),
						lpFeesPerShare:    map[common.Hash]*big.Int{},
						lmRewardsPerShare: map[common.Hash]*big.Int{},
						protocolFees:      map[common.Hash]uint64{},
						stakingPoolFees:   map[common.Hash]uint64{},
						shares: map[string]*Share{
							nftID1: &Share{
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
				params: &Params{
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PDEXRewardPoolPairsShare:        map[string]uint{},
					PRVDiscountPercent:              25,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
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
					params: &Params{
						DefaultFeeRateBPS: 30,
						FeeRateBPS: map[string]uint{
							"abc": 12,
						},
						PDEXRewardPoolPairsShare:        map[string]uint{},
						PRVDiscountPercent:              25,
						TradingProtocolFeePercent:       0,
						TradingStakingPoolRewardPercent: 10,
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
					PoolPairs:    map[string]*v2utils.PoolPairChange{},
					StakingPools: map[string]*v2utils.StakingPoolChange{},
				},
			},
			want: &stateV2{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
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
							nftID1: &Share{
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
				params:            NewParams(),
				stakingPoolStates: map[string]*StakingPoolState{},
				nftIDs:            map[string]uint64{},
				producer:          stateProducerV2{},
				processor:         stateProcessorV2{},
			},
			want1: &StateChange{
				PoolPairs: map[string]*v2utils.PoolPairChange{
					poolPairID: &v2utils.PoolPairChange{
						IsChanged: true,
						Shares: map[string]*v2utils.ShareChange{
							nftID1: &v2utils.ShareChange{
								IsChanged:             true,
								TradingFees:           map[string]bool{},
								LastLPFeesPerShare:    map[string]bool{},
								LastLmRewardsPerShare: map[string]bool{},
							},
						},
						OrderIDs:          map[string]bool{},
						LpFeesPerShare:    map[string]bool{},
						LmRewardsPerShare: map[string]bool{},
						ProtocolFees:      map[string]bool{},
						StakingPoolFees:   map[string]bool{},
						MakingVolume:      map[string]*v2utils.MakingVolumeChange{},
						OrderRewards:      map[string]*v2utils.OrderRewardChange{},
						LmLockedShare:     map[string]map[uint64]bool{},
					},
				},
				StakingPools: map[string]*v2utils.StakingPoolChange{},
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
				params: &Params{
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PDEXRewardPoolPairsShare:        map[string]uint{},
					PRVDiscountPercent:              25,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
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
								liquidity:           100,
								rewards:             map[common.Hash]uint64{},
								lastRewardsPerShare: map[common.Hash]*big.Int{},
							},
						},
						rewardsPerShare: map[common.Hash]*big.Int{},
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
					params: &Params{
						DefaultFeeRateBPS: 30,
						FeeRateBPS: map[string]uint{
							"abc": 12,
						},
						PDEXRewardPoolPairsShare:        map[string]uint{},
						PRVDiscountPercent:              25,
						TradingProtocolFeePercent:       0,
						TradingStakingPoolRewardPercent: 10,
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
					PoolPairs:    map[string]*v2utils.PoolPairChange{},
					StakingPools: map[string]*v2utils.StakingPoolChange{},
				},
			},
			want: &stateV2{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				params:                      NewParams(),
				stakingPoolStates: map[string]*StakingPoolState{
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
				nftIDs:    map[string]uint64{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			want1: &StateChange{
				PoolPairs: map[string]*v2utils.PoolPairChange{},
				StakingPools: map[string]*v2utils.StakingPoolChange{
					common.PRVIDStr: &v2utils.StakingPoolChange{
						RewardsPerShare: map[string]bool{},
						Stakers: map[string]*v2utils.StakerChange{
							nftID1: &v2utils.StakerChange{
								IsChanged:           true,
								Rewards:             map[string]bool{},
								LastRewardsPerShare: map[string]bool{},
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
				params: &Params{
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PDEXRewardPoolPairsShare:        map[string]uint{},
					PRVDiscountPercent:              25,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
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
								liquidity:           100,
								rewards:             map[common.Hash]uint64{},
								lastRewardsPerShare: map[common.Hash]*big.Int{},
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
					params: &Params{
						DefaultFeeRateBPS: 30,
						FeeRateBPS: map[string]uint{
							"abc": 12,
						},
						PDEXRewardPoolPairsShare:        map[string]uint{},
						PRVDiscountPercent:              25,
						TradingProtocolFeePercent:       0,
						TradingStakingPoolRewardPercent: 10,
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
									liquidity: 100,
								},
							},
						},
					},
					nftIDs:    map[string]uint64{},
					producer:  stateProducerV2{},
					processor: stateProcessorV2{},
				},
				stateChange: &StateChange{
					PoolPairs:    map[string]*v2utils.PoolPairChange{},
					StakingPools: map[string]*v2utils.StakingPoolChange{},
				},
			},
			want: &stateV2{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				params:                      NewParams(),
				stakingPoolStates: map[string]*StakingPoolState{
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
				nftIDs:    map[string]uint64{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			want1: &StateChange{
				PoolPairs: map[string]*v2utils.PoolPairChange{},
				StakingPools: map[string]*v2utils.StakingPoolChange{
					common.PRVIDStr: &v2utils.StakingPoolChange{
						RewardsPerShare: map[string]bool{},
						Stakers: map[string]*v2utils.StakerChange{
							nftID1: &v2utils.StakerChange{
								IsChanged:           false,
								Rewards:             map[string]bool{},
								LastRewardsPerShare: map[string]bool{},
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
				params: &Params{
					DefaultFeeRateBPS: 30,
					FeeRateBPS: map[string]uint{
						"abc": 12,
					},
					PDEXRewardPoolPairsShare:        map[string]uint{},
					PRVDiscountPercent:              25,
					TradingProtocolFeePercent:       0,
					TradingStakingPoolRewardPercent: 10,
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
								liquidity:           100,
								rewards:             map[common.Hash]uint64{},
								lastRewardsPerShare: map[common.Hash]*big.Int{},
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
					params: &Params{
						DefaultFeeRateBPS: 30,
						FeeRateBPS: map[string]uint{
							"abc": 12,
						},
						PDEXRewardPoolPairsShare:        map[string]uint{},
						PRVDiscountPercent:              25,
						TradingProtocolFeePercent:       0,
						TradingStakingPoolRewardPercent: 10,
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
									liquidity: 100,
								},
							},
						},
					},
					nftIDs:    map[string]uint64{},
					producer:  stateProducerV2{},
					processor: stateProcessorV2{},
				},
				stateChange: &StateChange{
					PoolPairs:    map[string]*v2utils.PoolPairChange{},
					StakingPools: map[string]*v2utils.StakingPoolChange{},
				},
			},
			want: &stateV2{
				stateBase:                   stateBase{},
				waitingContributions:        map[string]rawdbv2.Pdexv3Contribution{},
				deletedWaitingContributions: map[string]rawdbv2.Pdexv3Contribution{},
				poolPairs:                   map[string]*PoolPairState{},
				params:                      NewParams(),
				stakingPoolStates: map[string]*StakingPoolState{
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
				nftIDs:    map[string]uint64{},
				producer:  stateProducerV2{},
				processor: stateProcessorV2{},
			},
			want1: &StateChange{
				PoolPairs: map[string]*v2utils.PoolPairChange{},
				StakingPools: map[string]*v2utils.StakingPoolChange{
					common.PRVIDStr: &v2utils.StakingPoolChange{
						RewardsPerShare: map[string]bool{},
						Stakers: map[string]*v2utils.StakerChange{
							nftID1: &v2utils.StakerChange{
								IsChanged:           false,
								Rewards:             map[string]bool{},
								LastRewardsPerShare: map[string]bool{},
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
				poolPairs1 := make(map[string]*PoolPairState)
				json.Unmarshal(got.Reader().PoolPairs(), &poolPairs1)
				poolPairs2 := make(map[string]*PoolPairState)
				json.Unmarshal(tt.want.Reader().PoolPairs(), &poolPairs2)

				pool1, isExisted1 := poolPairs1[poolPairID]
				pool2, isExisted2 := poolPairs2[poolPairID]
				if isExisted1 && isExisted2 && !reflect.DeepEqual(pool1, pool2) {
					t.Errorf("stateV2.GetDiff() got = %v, want %v",
						getStrPoolPairState(pool1),
						getStrPoolPairState(pool2),
					)
				}
				pool3, isExisted3 := got.Reader().StakingPools()[common.PRVIDStr]
				pool4, isExisted4 := tt.want.Reader().StakingPools()[common.PRVIDStr]
				if isExisted3 && isExisted4 && !reflect.DeepEqual(pool3, pool4) {
					t.Errorf("stateV2.GetDiff() got = %v, want %v",
						getStrStakingPoolState(pool3),
						getStrStakingPoolState(pool4),
					)
				}

				t.Errorf("stateV2.GetDiff() got = %+v, want %+v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("stateV2.GetDiff() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
