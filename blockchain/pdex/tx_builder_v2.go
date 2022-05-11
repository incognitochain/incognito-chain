package pdex

import (
	"errors"
	"fmt"
	"strconv"

	v2 "github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
)

type TxBuilderV2 struct {
}

func (txBuilder *TxBuilderV2) Build(
	metaType int,
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
	beaconHeight uint64,
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	var err error

	switch metaType {
	case metadataCommon.Pdexv3UnstakingRequestMeta:
		if len(inst) != 3 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 3, len(inst))
		}
		tx, err = buildUnstakingTx(inst, producerPrivateKey, shardID, transactionStateDB)
	case metadataCommon.Pdexv3StakingRequestMeta:
		if len(inst) != 3 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 3, len(inst))
		}
		tx, err = buildStakingTx(inst, producerPrivateKey, shardID, transactionStateDB)
	case metadataCommon.Pdexv3UserMintNftRequestMeta:
		if len(inst) != 3 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 3, len(inst))
		}
		tx, err = buildUserMintNftTx(inst, producerPrivateKey, shardID, transactionStateDB)
	case metadataCommon.Pdexv3MintNftRequestMeta:
		if len(inst) != 3 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 3, len(inst))
		}
		tx, err = buildMintNftTx(inst, producerPrivateKey, shardID, transactionStateDB)
	case metadataCommon.Pdexv3AddLiquidityRequestMeta:
		if len(inst) != 3 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 3, len(inst))
		}
		switch inst[1] {
		case common.PDEContributionRefundChainStatus:
			tx, err = buildRefundContributionTxv2(inst, producerPrivateKey, shardID, transactionStateDB)
		case common.PDEContributionMatchedNReturnedChainStatus:
			tx, err = buildMatchAndReturnContributionTxv2(inst, producerPrivateKey, shardID, transactionStateDB)
		}
	case metadataCommon.Pdexv3WithdrawLiquidityRequestMeta:
		if len(inst) != 3 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 3, len(inst))
		}
		switch inst[1] {
		case common.PDEWithdrawalAcceptedChainStatus:
			tx, err = buildAcceptedWithdrawLiquidity(inst, producerPrivateKey, shardID, transactionStateDB, beaconHeight)
		}
	case metadataCommon.Pdexv3TradeRequestMeta:
		switch inst[1] {
		case strconv.Itoa(metadataPdexv3.TradeAcceptedStatus):
			action := instruction.Action{Content: &metadataPdexv3.AcceptedTrade{}}
			err := action.FromStringSlice(inst)
			if err != nil {
				return tx, err
			}
			tx, err = v2.TradeAcceptTx(action, producerPrivateKey, shardID, transactionStateDB)
		case strconv.Itoa(metadataPdexv3.TradeRefundedStatus):
			action := instruction.Action{Content: &metadataPdexv3.RefundedTrade{}}
			err := action.FromStringSlice(inst)
			if err != nil {
				return tx, err
			}
			tx, err = v2.TradeRefundTx(action, producerPrivateKey, shardID, transactionStateDB)
		default:
			return nil, fmt.Errorf("Invalid status %s from instruction", inst[1])
		}
	case metadataCommon.Pdexv3AddOrderRequestMeta:
		switch inst[1] {
		case strconv.Itoa(metadataPdexv3.OrderRefundedStatus):
			action := instruction.Action{Content: &metadataPdexv3.RefundedAddOrder{}}
			err := action.FromStringSlice(inst)
			if err != nil {
				return nil, err
			}
			tx, err = v2.OrderRefundTx(action, producerPrivateKey, shardID, transactionStateDB)
			if err != nil {
				return nil, err
			}
		case strconv.Itoa(metadataPdexv3.OrderAcceptedStatus):
			return nil, nil
		default:
			return nil, fmt.Errorf("Invalid status %s from instruction", inst[1])
		}
	case metadataCommon.Pdexv3WithdrawOrderRequestMeta:
		switch inst[1] {
		case strconv.Itoa(metadataPdexv3.WithdrawOrderAcceptedStatus):
			action := instruction.Action{Content: &metadataPdexv3.AcceptedWithdrawOrder{}}
			err := action.FromStringSlice(inst)
			if err != nil {
				return nil, err
			}
			tx, err = v2.WithdrawOrderAcceptTx(action, producerPrivateKey, shardID, transactionStateDB)
			if err != nil {
				return nil, err
			}
		case strconv.Itoa(metadataPdexv3.WithdrawOrderRejectedStatus):
			return nil, nil
		default:
			return nil, fmt.Errorf("Invalid status %s from instruction", inst[1])
		}
	case metadataCommon.Pdexv3WithdrawLPFeeRequestMeta:
		if len(inst) == 4 {
			tx, err = v2.WithdrawLPFee(
				inst[2],
				inst[3],
				producerPrivateKey,
				shardID,
				transactionStateDB,
			)
		} else {
			return tx, fmt.Errorf("Length of instruction is invalid expect %v but get %v", 4, len(inst))
		}
	case metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta:
		if len(inst) == 4 {
			tx, err = v2.WithdrawProtocolFee(
				inst[2],
				inst[3],
				producerPrivateKey,
				shardID,
				transactionStateDB,
			)
		} else {
			return tx, fmt.Errorf("Length of instruction is invalid expect %v but get %v", 4, len(inst))
		}
	case metadataCommon.Pdexv3MintPDEXGenesisMeta:
		if len(inst) == 4 {
			tx, err = v2.MintPDEXGenesis(
				inst[2],
				inst[3],
				producerPrivateKey,
				shardID,
				transactionStateDB,
			)
		} else {
			return tx, fmt.Errorf("Length of instruction is invalid expect %v but get %v", 4, len(inst))
		}
	case metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta:
		if len(inst) == 4 {
			tx, err = v2.WithdrawStakingReward(
				inst[2],
				inst[3],
				producerPrivateKey,
				shardID,
				transactionStateDB,
			)
		} else {
			return tx, fmt.Errorf("Length of instruction is invalid expect %v but get %v", 4, len(inst))
		}
	case metadataCommon.Pdexv3MintAccessTokenMeta:
		if len(inst) != 3 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 3, len(inst))
		}
		tx, err = buildMintAccessTokenTx(inst, producerPrivateKey, shardID, transactionStateDB)
	}

	return tx, err
}

func buildRefundContributionTxv2(
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	refundInst := instruction.NewRefundAddLiquidity()
	err := refundInst.FromStringSlice(inst)
	if err != nil {
		return tx, err
	}
	refundContribution := refundInst.Contribution()
	refundContributionValue := refundContribution.Value()

	if refundContributionValue.ShardID() != shardID {
		return tx, nil
	}
	metaData := metadataPdexv3.NewAddLiquidityResponseWithValue(
		common.PDEContributionRefundChainStatus,
		refundContributionValue.TxReqID().String(),
	)

	otaReceiver := privacy.OTAReceiver{}
	otaReceiverStr := refundContributionValue.OtaReceiver()
	if len(refundContributionValue.OtaReceivers()) != 0 {
		otaReceiverStr = refundContributionValue.OtaReceivers()[refundContributionValue.TokenID()]
	}
	err = otaReceiver.FromString(otaReceiverStr)
	if err != nil {
		return tx, err
	}
	tx, err = buildMintTokenTx(
		refundContributionValue.TokenID(), refundContributionValue.Amount(),
		otaReceiver, producerPrivateKey, transactionStateDB, metaData,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
	}
	return tx, err
}

func buildUserMintNftTx(
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	if len(inst) != 3 {
		return tx, fmt.Errorf("Expect inst length to be %v but get %v", 3, len(inst))
	}
	if inst[0] != strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta) {
		return tx, fmt.Errorf("Expect inst metaType to be %v but get %s", metadataCommon.Pdexv3UserMintNftRequestMeta, inst[0])
	}

	var instShardID byte
	var tokenID common.Hash
	var otaReceiverStr, status, txReqID string
	var amount uint64
	switch inst[1] {
	case common.Pdexv3RejectStringStatus:
		refundInst := instruction.NewRejectUserMintNft()
		err := refundInst.FromStringSlice(inst)
		if err != nil {
			return tx, err
		}
		instShardID = refundInst.ShardID()
		tokenID = common.PRVCoinID
		otaReceiverStr = refundInst.OtaReceiver()
		amount = refundInst.Amount()
		txReqID = refundInst.TxReqID().String()
	case common.Pdexv3AcceptStringStatus:
		acceptInst := instruction.NewAcceptUserMintNft()
		err := acceptInst.FromStringSlice(inst)
		if err != nil {
			return tx, err
		}
		instShardID = acceptInst.ShardID()
		tokenID = acceptInst.NftID()
		otaReceiverStr = acceptInst.OtaReceiver()
		amount = 1
		txReqID = acceptInst.TxReqID().String()
	default:
		return tx, errors.New("Can not recognize status")
	}
	if instShardID != shardID || tokenID.IsZeroValue() {
		return tx, nil
	}

	status = inst[1]
	otaReceiver := privacy.OTAReceiver{}
	err := otaReceiver.FromString(otaReceiverStr)
	if err != nil {
		return tx, err
	}
	metaData := metadataPdexv3.NewUserMintNftResponseWithValue(status, txReqID)
	tx, err = buildMintTokenTx(tokenID, amount, otaReceiver, producerPrivateKey, transactionStateDB, metaData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
	}
	return tx, err
}

func buildStakingTx(
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	if len(inst) != 3 {
		return tx, fmt.Errorf("Expect inst length to be %v but get %v", 3, len(inst))
	}
	if inst[0] != strconv.Itoa(metadataCommon.Pdexv3StakingRequestMeta) {
		return tx, fmt.Errorf("Expect inst metaType to be %v but get %s", metadataCommon.Pdexv3StakingRequestMeta, inst[0])
	}
	if inst[1] != common.Pdexv3RejectStringStatus {
		return tx, nil
	}

	rejectInst := instruction.RejectStaking{}
	err := rejectInst.FromStringSlice(inst)
	if err != nil {
		return tx, err
	}

	if rejectInst.ShardID() != shardID || rejectInst.TokenID().IsZeroValue() {
		return tx, nil
	}
	otaReceiver := privacy.OTAReceiver{}
	err = otaReceiver.FromString(rejectInst.OtaReceiver())
	if err != nil {
		return tx, err
	}
	metaData := metadataPdexv3.NewStakingResponseWithValue(common.Pdexv3RejectStringStatus, rejectInst.TxReqID().String())
	tx, err = buildMintTokenTx(
		rejectInst.TokenID(), rejectInst.Amount(),
		otaReceiver, producerPrivateKey, transactionStateDB, metaData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
	}
	return tx, err
}

func buildMintNftTx(
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	mintNftInst := instruction.NewMintNft()
	err := mintNftInst.FromStringSlice(inst)
	if err != nil {
		return tx, err
	}

	if mintNftInst.ShardID() != shardID || mintNftInst.NftID().IsZeroValue() {
		return tx, nil
	}

	otaReceiver := privacy.OTAReceiver{}
	err = otaReceiver.FromString(mintNftInst.OtaReceiver())
	if err != nil {
		return tx, err
	}
	metaData := metadataPdexv3.NewMintNftResponseWithValue(mintNftInst.NftID().String(), mintNftInst.OtaReceiver())
	tx, err = buildMintTokenTx(
		mintNftInst.NftID(), 1,
		otaReceiver, producerPrivateKey, transactionStateDB, metaData,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
	}
	return tx, err
}

func buildMatchAndReturnContributionTxv2(
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	var res metadata.Transaction
	matchAndReturnInst := instruction.NewMatchAndReturnAddLiquidity()
	err := matchAndReturnInst.FromStringSlice(inst)
	if err != nil {
		return res, err
	}
	matchAndReturnContribution := matchAndReturnInst.Contribution()
	matchAndReturnContributionValue := matchAndReturnContribution.Value()
	if matchAndReturnContributionValue.ShardID() != shardID {
		return res, nil
	}

	metaData := metadataPdexv3.NewAddLiquidityResponseWithValue(
		common.PDEContributionMatchedNReturnedChainStatus,
		matchAndReturnContributionValue.TxReqID().String(),
	)
	refundAddress := privacy.OTAReceiver{}
	refundAddressStr := matchAndReturnContributionValue.OtaReceiver()
	if len(matchAndReturnContributionValue.OtaReceivers()) != 0 {
		refundAddressStr = matchAndReturnContributionValue.OtaReceivers()[matchAndReturnContributionValue.TokenID()]
	}
	err = refundAddress.FromString(refundAddressStr)
	if err != nil {
		return res, err
	}
	res, err = buildMintTokenTx(
		matchAndReturnContributionValue.TokenID(), matchAndReturnInst.ReturnAmount(),
		refundAddress, producerPrivateKey, transactionStateDB, metaData,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted add liquidity response tx: %+v", err)
	}
	return res, err
}

func buildMintTokenTx(
	tokenID common.Hash, tokenAmount uint64,
	otaReceiver privacy.OTAReceiver,
	producerPrivateKey *privacy.PrivateKey,
	transactionStateDB *statedb.StateDB,
	meta metadata.Metadata,
) (metadata.Transaction, error) {
	var txParam transaction.TxSalaryOutputParams
	txParam = transaction.TxSalaryOutputParams{
		Amount:          tokenAmount,
		ReceiverAddress: nil,
		PublicKey:       &otaReceiver.PublicKey,
		TxRandom:        &otaReceiver.TxRandom,
		TokenID:         &tokenID,
		Info:            []byte{},
	}
	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, func(c privacy.Coin) metadata.Metadata {
		return meta
	})
}

func buildAcceptedWithdrawLiquidity(
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
	beaconHeight uint64,
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	withdrawLiquidityInst := instruction.NewAcceptWithdrawLiquidity()
	err := withdrawLiquidityInst.FromStringSlice(inst)
	if err != nil {
		return tx, err
	}

	if withdrawLiquidityInst.ShardID() != shardID {
		return tx, nil
	}
	metaData := metadataPdexv3.NewWithdrawLiquidityResponseWithValue(
		common.PDEWithdrawalAcceptedChainStatus,
		withdrawLiquidityInst.TxReqID().String(),
	)
	otaReceiver := privacy.OTAReceiver{}
	err = otaReceiver.FromString(withdrawLiquidityInst.OtaReceiver())
	if err != nil {
		if config.Param().Net == config.Testnet2Net && beaconHeight < 3790600 {
			return tx, nil
		}
		return tx, err
	}
	tx, err = buildMintTokenTx(
		withdrawLiquidityInst.TokenID(), withdrawLiquidityInst.TokenAmount(),
		otaReceiver, producerPrivateKey, transactionStateDB, metaData,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
	}
	return tx, err
}

func buildUnstakingTx(
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	acceptInst := instruction.AcceptUnstaking{}
	err := acceptInst.FromStringSlice(inst)
	if err != nil {
		return tx, nil
	}

	if acceptInst.ShardID() != shardID || acceptInst.StakingPoolID().IsZeroValue() {
		return tx, nil
	}
	otaReceiver := privacy.OTAReceiver{}
	err = otaReceiver.FromString(acceptInst.OtaReceiver())
	if err != nil {
		return tx, err
	}
	metaData := metadataPdexv3.NewUnstakingResponseWithValue(
		common.Pdexv3AcceptStringStatus, acceptInst.TxReqID().String(),
	)
	tx, err = buildMintTokenTx(
		acceptInst.StakingPoolID(), acceptInst.Amount(),
		otaReceiver, producerPrivateKey, transactionStateDB, metaData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
	}
	return tx, err
}

func buildMintAccessTokenTx(
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	if len(inst) != 3 {
		return tx, fmt.Errorf("Expect inst length to be %v but get %v", 3, len(inst))
	}
	if inst[0] != strconv.Itoa(metadataCommon.Pdexv3MintAccessTokenMeta) {
		return tx, fmt.Errorf("Expect inst metaType to be %v but get %s", metadataCommon.Pdexv3MintAccessTokenMeta, inst[0])
	}

	var instShardID byte
	var tokenID common.Hash
	var otaReceiverStr, txReqID string
	var amount uint64
	mintAccessTokenInst := instruction.NewMintAccessToken()
	err := mintAccessTokenInst.FromStringSlice(inst)
	if err != nil {
		return tx, err
	}
	instShardID = mintAccessTokenInst.ShardID()
	tokenID = common.PdexAccessCoinID
	otaReceiverStr = mintAccessTokenInst.OtaReceiver()
	amount = 1
	txReqID = mintAccessTokenInst.TxReqID().String()
	if instShardID != shardID {
		return tx, nil
	}

	otaReceiver := privacy.OTAReceiver{}
	err = otaReceiver.FromString(otaReceiverStr)
	if err != nil {
		return tx, err
	}
	metaData := metadataPdexv3.NewMintAccessTokenWithValue(txReqID)
	tx, err = buildMintTokenTx(tokenID, amount, otaReceiver, producerPrivateKey, transactionStateDB, metaData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
	}
	return tx, err
}
