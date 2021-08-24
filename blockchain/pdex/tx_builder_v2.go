package pdex

import (
	"errors"
	"fmt"
	"strconv"

	v2 "github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
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
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	var err error

	switch metaType {
	case metadataCommon.Pdexv3UserMintNftRequestMeta:
		if len(inst) != 3 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 3, len(inst))
		}
		tx, err = buildPdexv3UserMintNft(inst, producerPrivateKey, shardID, transactionStateDB)
	case metadataCommon.Pdexv3MintNftRequestMeta:
		if len(inst) != 3 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 3, len(inst))
		}
		tx, err = buildPdexv3MintNft(inst, producerPrivateKey, shardID, transactionStateDB)
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
			tx, err = buildAcceptedWithdrawLiquidity(inst, producerPrivateKey, shardID, transactionStateDB)
		default:
			return tx, errors.New("Invalid withdraw liquidity status")
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
	err = otaReceiver.FromString(refundContributionValue.RefundAddress())
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

func buildPdexv3UserMintNft(
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
	var otaReceiveStr, status, txReqID string
	var amount uint64
	switch inst[1] {
	case common.Pdexv3RejectUserMintNftStatus:
		refundInst := instruction.NewRejectUserMintNft()
		err := refundInst.FromStringSlice(inst)
		if err != nil {
			return tx, err
		}
		instShardID = refundInst.ShardID()
		tokenID = common.PRVCoinID
		otaReceiveStr = refundInst.OtaReceive()
		amount = refundInst.Amount()
		txReqID = refundInst.TxReqID().String()
	case common.Pdexv3AcceptUserMintNftStatus:
		acceptInst := instruction.NewAcceptUserMintNft()
		err := acceptInst.FromStringSlice(inst)
		if err != nil {
			return tx, err
		}
		instShardID = acceptInst.ShardID()
		tokenID = acceptInst.NftID()
		otaReceiveStr = acceptInst.OtaReceive()
		amount = 1
		txReqID = acceptInst.TxReqID().String()
	default:
		return tx, errors.New("Can not recognize status")
	}
	if instShardID != shardID || tokenID.IsZeroValue() {
		return tx, nil
	}

	status = inst[1]
	otaReceive := privacy.OTAReceiver{}
	err := otaReceive.FromString(otaReceiveStr)
	if err != nil {
		return tx, err
	}
	metaData := metadataPdexv3.NewUserMintNftResponseWithValue(status, txReqID)
	tx, err = buildMintTokenTx(tokenID, amount, otaReceive, producerPrivateKey, transactionStateDB, metaData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
	}
	return tx, err

}

func buildPdexv3MintNft(
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
	err = refundAddress.FromString(matchAndReturnContributionValue.RefundAddress())
	if err != nil {
		return res, err
	}
	res, err = buildMintTokenTx(
		matchAndReturnContributionValue.TokenID(), matchAndReturnInst.ReturnAmount(),
		refundAddress, producerPrivateKey, transactionStateDB, metaData,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
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
	err = otaReceiver.FromString(withdrawLiquidityInst.OtaReceive())
	if err != nil {
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

func buildRejectedWithdrawLiquidity(
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
	err = otaReceiver.FromString(refundContributionValue.RefundAddress())
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
