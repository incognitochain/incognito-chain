package pdex

import (
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
	nftIDs map[string]bool
}

func (txBuilder *TxBuilderV2) ClearCache() {
	txBuilder.nftIDs = make(map[string]bool)
}

func (txBuilder *TxBuilderV2) Build(
	metaType int,
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) ([]metadata.Transaction, error) {

	res := []metadata.Transaction{}
	var err error

	switch metaType {
	case metadataCommon.Pdexv3AddLiquidityRequestMeta:
		if len(inst) < 3 {
			return res, fmt.Errorf("Length of instruction is invalid expectm equal or greater than %v but get %v", 3, len(inst))
		}
		switch inst[1] {
		case common.PDEContributionRefundChainStatus:
			tx, err := buildRefundContributionTxv2(inst, producerPrivateKey, shardID, transactionStateDB)
			if err != nil {
				return res, err
			}
			if tx != nil {
				res = append(res, tx)
			}
		case common.PDEContributionMatchedChainStatus:
			tx, err := buildMatchContributionTxv2(inst, producerPrivateKey, shardID, transactionStateDB)
			if err != nil {
				return res, err
			}
			if tx != nil {
				res = append(res, tx)
			}
		case common.PDEContributionMatchedNReturnedChainStatus:
			txs, err := buildMatchAndReturnContributionTxv2(inst, producerPrivateKey, shardID, transactionStateDB, txBuilder.nftIDs)
			if err != nil {
				return res, err
			}
			if len(txs) != 0 {
				res = append(res, txs...)
			}
		}
	case metadataCommon.Pdexv3TradeRequestMeta:
		switch inst[1] {
		case strconv.Itoa(metadataPdexv3.TradeAcceptedStatus):
			action := instruction.Action{Content: &metadataPdexv3.AcceptedTrade{}}
			err := action.FromStringSlice(inst)
			if err != nil {
				return res, err
			}
			tx, err := v2.TradeAcceptTx(action, producerPrivateKey, shardID, transactionStateDB)
			if err != nil {
				return res, err
			}
			res = append(res, tx)
		case strconv.Itoa(metadataPdexv3.TradeRefundedStatus):
			action := instruction.Action{Content: &metadataPdexv3.RefundedTrade{}}
			err := action.FromStringSlice(inst)
			if err != nil {
				return nil, err
			}
			tx, err := v2.TradeRefundTx(action, producerPrivateKey, shardID, transactionStateDB)
			if err != nil {
				return res, err
			}
			res = append(res, tx)
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
			tx, err := v2.OrderRefundTx(action, producerPrivateKey, shardID, transactionStateDB)
			if err != nil {
				return res, err
			}
			res = append(res, tx)
		default:
			return nil, fmt.Errorf("Invalid status %s from instruction", inst[1])
		}
	}

	return res, err
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
	tx, err = buildMintTokenTxs(
		refundContributionValue.TokenID(), refundContributionValue.Amount(),
		otaReceiver, producerPrivateKey, transactionStateDB, metaData,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
		return tx, err
	}
	return tx, nil
}

func buildMatchContributionTxv2(
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	matchInst := instruction.NewMatchAddLiquidity()
	err := matchInst.FromStringSlice(inst)
	if err != nil {
		return tx, err
	}

	matchContribution := matchInst.Contribution()
	matchContributionValue := matchContribution.Value()
	if matchContributionValue.ShardID() != shardID || matchInst.NftID().IsZeroValue() {
		return tx, nil
	}
	metaData := metadataPdexv3.NewAddLiquidityResponseWithValue(
		common.PDEContributionMatchedChainStatus,
		matchContributionValue.TxReqID().String(),
	)
	otaReceiver := privacy.OTAReceiver{}
	err = otaReceiver.FromString(matchContributionValue.ReceiveAddress())
	if err != nil {
		return tx, err
	}
	tx, err = buildMintTokenTxs(
		matchInst.NftID(), 1,
		otaReceiver, producerPrivateKey, transactionStateDB, metaData,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
		return tx, err
	}
	return tx, nil

}

func buildMatchAndReturnContributionTxv2(
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
	nftIDs map[string]bool,
) ([]metadata.Transaction, error) {
	res := []metadata.Transaction{}
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
		common.PDEContributionMatchedChainStatus,
		matchAndReturnContributionValue.TxReqID().String(),
	)
	if !nftIDs[matchAndReturnInst.NftID().String()] || matchAndReturnInst.NftID().IsZeroValue() {
		receiveAddress := privacy.OTAReceiver{}
		err = receiveAddress.FromString(matchAndReturnContributionValue.ReceiveAddress())
		if err != nil {
			return res, err
		}
		tx0, err := buildMintTokenTxs(
			matchAndReturnInst.NftID(), 1,
			receiveAddress, producerPrivateKey, transactionStateDB, metaData,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
			return res, err
		}
		res = append(res, tx0)
		nftIDs[matchAndReturnInst.NftID().String()] = true
	}
	refundAddress := privacy.OTAReceiver{}
	err = refundAddress.FromString(matchAndReturnContributionValue.RefundAddress())
	if err != nil {
		return res, err
	}
	tx1, err := buildMintTokenTxs(
		matchAndReturnContributionValue.TokenID(), matchAndReturnInst.ReturnAmount(),
		refundAddress, producerPrivateKey, transactionStateDB, metaData,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
		return res, err
	}
	res = append(res, tx1)
	return res, nil
}

func buildMintTokenTxs(
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
