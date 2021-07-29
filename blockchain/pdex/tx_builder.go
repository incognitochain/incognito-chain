package pdex

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

type TxBuilder struct {
	nfctIDs map[string]bool
}

func (txBuilder *TxBuilder) ClearCache() {
	txBuilder.nfctIDs = make(map[string]bool)
}

func (txBuilder *TxBuilder) Build(
	metaType int,
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) ([]metadata.Transaction, error) {

	res := []metadata.Transaction{}

	switch metaType {
	case metadata.PDETradeRequestMeta:
		if len(inst) != 4 {
			return res, fmt.Errorf("Length of instruction is invalid expect %v but get %v", 4, len(inst))
		}
		tx, err := buildTradeIssuanceTx(
			inst[2],
			inst[3],
			producerPrivateKey,
			shardID,
			transactionStateDB,
		)
		if err != nil {
			return res, err
		}
		res = append(res, tx)
	case metadata.PDECrossPoolTradeRequestMeta:
		if len(inst) != 4 {
			return res, fmt.Errorf("Length of instruction is invalid expect %v but get %v", 4, len(inst))
		}
		tx, err := buildCrossPoolTradeIssuanceTx(
			inst[2],
			inst[3],
			producerPrivateKey,
			shardID,
			transactionStateDB,
		)
		if err != nil {
			return res, err
		}
		res = append(res, tx)
	case metadata.PDEWithdrawalRequestMeta:
		if len(inst) != 4 {
			return res, fmt.Errorf("Length of instruction is invalid expect %v but get %v", 4, len(inst))
		}
		if inst[2] == common.PDEWithdrawalAcceptedChainStatus {
			tx, err := buildWithdrawalTx(
				inst[3],
				producerPrivateKey,
				shardID,
				transactionStateDB,
			)
			if err != nil {
				return res, err
			}
			res = append(res, tx)
		}
	case metadata.PDEFeeWithdrawalRequestMeta:
		if len(inst) != 4 {
			return res, fmt.Errorf("Length of instruction is invalid expect %v but get %v", 4, len(inst))
		}
		tx, err := buildFeeWithdrawalTx(
			inst[3],
			producerPrivateKey,
			shardID,
			transactionStateDB,
		)
		if err != nil {
			return res, err
		}
		res = append(res, tx)
	case metadata.PDEContributionMeta, metadata.PDEPRVRequiredContributionRequestMeta:
		if len(inst) != 4 {
			return res, fmt.Errorf("Length of instruction is invalid expect %v but get %v", 4, len(inst))
		}
		if inst[2] == common.PDEContributionRefundChainStatus {
			tx, err := buildRefundContributionTx(
				inst[3],
				producerPrivateKey,
				shardID,
				transactionStateDB,
			)
			if err != nil {
				return res, err
			}
			res = append(res, tx)
		} else if inst[2] == common.PDEContributionMatchedNReturnedChainStatus {
			tx, err := buildMatchedAndReturnedContributionTx(
				inst[3],
				producerPrivateKey,
				shardID,
				transactionStateDB,
			)
			if err != nil {
				return res, err
			}
			res = append(res, tx)
		}
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
			txs, err := buildMatchAndReturnContributionTxv2(inst, producerPrivateKey, shardID, transactionStateDB, txBuilder.nfctIDs)
			if err != nil {
				return res, err
			}
			if len(txs) != 0 {
				res = append(res, txs...)
			}
		}
	}

	return res, nil
}

func buildTradeIssuanceTx(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	Logger.log.Info("[PDE Trade] Starting...")
	if instStatus == common.PDETradeRefundChainStatus {
		return buildTradeRefundTx(
			instStatus,
			contentStr,
			producerPrivateKey,
			shardID,
			transactionStateDB,
		)
	}
	return buildTradeAcceptedTx(
		instStatus,
		contentStr,
		producerPrivateKey,
		shardID,
		transactionStateDB,
	)
}

func buildTradeRefundTx(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	tradeRequestAction, err := parseTradeRefundContent(contentStr)
	if err != nil {
		return nil, nil
	}

	if shardID != tradeRequestAction.ShardID {
		return nil, nil
	}

	meta := metadata.NewPDETradeResponse(
		instStatus,
		tradeRequestAction.TxReqID,
		metadata.PDETradeResponseMeta,
	)

	resTx, err := buildTradeResTx(
		tradeRequestAction.Meta.TraderAddressStr,
		tradeRequestAction.Meta.TxRandomStr,
		tradeRequestAction.Meta.SellAmount+tradeRequestAction.Meta.TradingFee,
		tradeRequestAction.Meta.TokenIDToSellStr,
		producerPrivateKey,
		transactionStateDB,
		meta,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing refunded trading response tx: %+v", err)
		return nil, nil
	}
	Logger.log.Info("[PDE Trade] Create refunded tx ok.")
	return resTx, nil
}

func buildTradeAcceptedTx(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	tradeAcceptedContent, err := parseTradeAcceptedContent(contentStr)
	if err != nil {
		return nil, nil
	}

	if shardID != tradeAcceptedContent.ShardID {
		return nil, nil
	}

	meta := metadata.NewPDETradeResponse(
		instStatus,
		tradeAcceptedContent.RequestedTxID,
		metadata.PDETradeResponseMeta,
	)

	resTx, err := buildTradeResTx(
		tradeAcceptedContent.TraderAddressStr,
		tradeAcceptedContent.TxRandomStr,
		tradeAcceptedContent.ReceiveAmount,
		tradeAcceptedContent.TokenIDToBuyStr,
		producerPrivateKey,
		transactionStateDB,
		meta,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
		return nil, nil
	}
	Logger.log.Info("[PDE Trade] Create accepted tx ok.")
	return resTx, nil
}

func parseTradeRefundContent(
	contentStr string,
) (*metadata.PDETradeRequestAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of pde trade refund instruction: %+v", err)
		return nil, err
	}
	var tradeRequestAction metadata.PDETradeRequestAction
	err = json.Unmarshal(contentBytes, &tradeRequestAction)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde trade refund content: %+v", err)
		return nil, err
	}

	return &tradeRequestAction, nil
}

func parseCrossPoolTradeRefundContent(
	contentStr string,
) (*metadata.PDERefundCrossPoolTrade, error) {
	contentBytes := []byte(contentStr)
	var refundCrossPoolTrade metadata.PDERefundCrossPoolTrade
	err := json.Unmarshal(contentBytes, &refundCrossPoolTrade)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde cross pool trade refund content: %+v", err)
		return nil, err
	}
	return &refundCrossPoolTrade, nil
}

func parseTradeAcceptedContent(
	contentStr string,
) (*metadata.PDETradeAcceptedContent, error) {
	contentBytes := []byte(contentStr)
	var tradeAcceptedContent metadata.PDETradeAcceptedContent
	err := json.Unmarshal(contentBytes, &tradeAcceptedContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde trade accepted content: %+v", err)
		return nil, err
	}
	return &tradeAcceptedContent, nil
}

func parseCrossPoolTradeAcceptedContent(
	contentStr string,
) ([]metadata.PDECrossPoolTradeAcceptedContent, error) {
	contentBytes := []byte(contentStr)
	var crossPoolTradeAcceptedContent []metadata.PDECrossPoolTradeAcceptedContent
	err := json.Unmarshal(contentBytes, &crossPoolTradeAcceptedContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde cross pool trade accepted content: %+v", err)
		return nil, err
	}
	return crossPoolTradeAcceptedContent, nil
}

func buildTradeResTx(
	receiverAddressStr string,
	txRandomStr string,
	receiveAmt uint64,
	tokenIDStr string,
	producerPrivateKey *privacy.PrivateKey,
	transactionStateDB *statedb.StateDB,
	meta metadata.Metadata,
) (metadata.Transaction, error) {

	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while converting tokenid to hash: %+v", err)
		return nil, err
	}

	var txParam transaction.TxSalaryOutputParams
	if len(txRandomStr) > 0 {
		publicKey, txRandom, err := coin.ParseOTAInfoFromString(receiverAddressStr, txRandomStr)
		if err != nil {
			return nil, err
		}
		txParam = transaction.TxSalaryOutputParams{Amount: receiveAmt, ReceiverAddress: nil, PublicKey: publicKey, TxRandom: txRandom, TokenID: tokenID, Info: []byte{}}
	} else {
		paymentAddress, err := wallet.Base58CheckDeserialize(receiverAddressStr)
		if err != nil {
			return nil, err
		}
		txParam = transaction.TxSalaryOutputParams{Amount: receiveAmt, ReceiverAddress: &paymentAddress.KeySet.PaymentAddress, TokenID: tokenID, Info: []byte{}}
	}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, func(c privacy.Coin) metadata.Metadata {
		return meta
	})
}

func buildCrossPoolTradeIssuanceTx(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	Logger.log.Info("[PDE Cross Pool Trade] Starting...")
	if instStatus == common.PDECrossPoolTradeFeeRefundChainStatus || instStatus == common.PDECrossPoolTradeSellingTokenRefundChainStatus {
		return buildCrossPoolTradeRefundTx(
			instStatus,
			contentStr,
			producerPrivateKey,
			shardID,
			transactionStateDB,
		)
	}
	return buildCrossPoolTradeAcceptedTx(
		instStatus,
		contentStr,
		producerPrivateKey,
		shardID,
		transactionStateDB,
	)
}

func buildCrossPoolTradeRefundTx(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	crossPoolTradeRefundContent, err := parseCrossPoolTradeRefundContent(contentStr)
	if err != nil {
		return nil, nil
	}

	if shardID != crossPoolTradeRefundContent.ShardID {
		return nil, nil
	}
	meta := metadata.NewPDECrossPoolTradeResponse(
		instStatus,
		crossPoolTradeRefundContent.TxReqID,
		metadata.PDECrossPoolTradeResponseMeta,
	)

	resTx, err := buildTradeResTx(
		crossPoolTradeRefundContent.TraderAddressStr,
		crossPoolTradeRefundContent.TxRandomStr,
		crossPoolTradeRefundContent.Amount,
		crossPoolTradeRefundContent.TokenIDStr,
		producerPrivateKey,
		transactionStateDB,
		meta,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing refunded trading response tx: %+v", err)
		return nil, nil
	}
	Logger.log.Info("[PDE Cross Pool Trade] Create refunded tx ok.")
	return resTx, nil
}

func buildCrossPoolTradeAcceptedTx(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	crossPoolTradeAcceptedContents, err := parseCrossPoolTradeAcceptedContent(contentStr)
	if err != nil {
		return nil, nil
	}
	len := len(crossPoolTradeAcceptedContents)
	if len == 0 {
		Logger.log.Warn("WARNING: cross pool trade contents is empty.")
		return nil, nil
	}

	finalCrossPoolTradeAcceptedContent := crossPoolTradeAcceptedContents[len-1]

	if shardID != finalCrossPoolTradeAcceptedContent.ShardID {
		return nil, nil
	}

	meta := metadata.NewPDECrossPoolTradeResponse(
		instStatus,
		finalCrossPoolTradeAcceptedContent.RequestedTxID,
		metadata.PDECrossPoolTradeResponseMeta,
	)

	resTx, err := buildTradeResTx(
		finalCrossPoolTradeAcceptedContent.TraderAddressStr,
		finalCrossPoolTradeAcceptedContent.TxRandomStr,
		finalCrossPoolTradeAcceptedContent.ReceiveAmount,
		finalCrossPoolTradeAcceptedContent.TokenIDToBuyStr,
		producerPrivateKey,
		transactionStateDB,
		meta,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted cross pool trading response tx: %+v", err)
		return nil, nil
	}
	Logger.log.Info("[PDE Cross Pool Trade] Create accepted tx ok.")
	return resTx, nil
}

func buildWithdrawalTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	Logger.log.Info("[PDE Withdrawal] Starting...")
	contentBytes := []byte(contentStr)
	var wdAcceptedContent metadata.PDEWithdrawalAcceptedContent
	err := json.Unmarshal(contentBytes, &wdAcceptedContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde withdrawal content: %+v", err)
		return nil, nil
	}
	if wdAcceptedContent.ShardID != shardID {
		return nil, nil
	}

	withdrawalTokenIDStr := wdAcceptedContent.WithdrawalTokenIDStr
	meta := metadata.NewPDEWithdrawalResponse(
		withdrawalTokenIDStr,
		wdAcceptedContent.TxReqID,
		metadata.PDEWithdrawalResponseMeta,
	)
	tokenID, err := common.Hash{}.NewHashFromStr(withdrawalTokenIDStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while converting tokenid to hash: %+v", err)
		return nil, nil
	}
	keyWallet, err := wallet.Base58CheckDeserialize(wdAcceptedContent.WithdrawerAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing trader address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress

	txParam := transaction.TxSalaryOutputParams{Amount: wdAcceptedContent.DeductingPoolValue, ReceiverAddress: &receiverAddr, TokenID: tokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, makeMD)
}

func buildFeeWithdrawalTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	Logger.log.Info("[PDE Fee Withdrawal] Starting...")
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of pde withdrawal action: %+v", err)
		return nil, nil
	}
	var feeWithdrawalRequestAction metadata.PDEFeeWithdrawalRequestAction
	err = json.Unmarshal(contentBytes, &feeWithdrawalRequestAction)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde fee withdrawal request action: %+v", err)
		return nil, nil
	}

	if feeWithdrawalRequestAction.ShardID != shardID {
		return nil, nil
	}

	meta := metadata.NewPDEFeeWithdrawalResponse(
		feeWithdrawalRequestAction.TxReqID,
		metadata.PDEFeeWithdrawalResponseMeta,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(feeWithdrawalRequestAction.Meta.WithdrawerAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing withdrawer address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress

	txParam := transaction.TxSalaryOutputParams{
		Amount:          feeWithdrawalRequestAction.Meta.WithdrawalFeeAmt,
		ReceiverAddress: &receiverAddr,
		TokenID:         &common.PRVCoinID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, makeMD)
}

func buildRefundContributionTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	Logger.log.Info("[PDE Refund contribution] Starting...")
	contentBytes := []byte(contentStr)
	var refundContribution metadata.PDERefundContribution
	err := json.Unmarshal(contentBytes, &refundContribution)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde refund contribution content: %+v", err)
		return nil, nil
	}
	if refundContribution.ShardID != shardID {
		return nil, nil
	}

	meta := metadata.NewPDEContributionResponse(
		"refund",
		refundContribution.TxReqID,
		refundContribution.TokenIDStr,
		metadata.PDEContributionResponseMeta,
	)
	refundTokenIDStr := refundContribution.TokenIDStr
	tokenID, err := common.Hash{}.NewHashFromStr(refundTokenIDStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while converting tokenid to hash: %+v", err)
		return nil, nil
	}
	keyWallet, err := wallet.Base58CheckDeserialize(refundContribution.ContributorAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing contributor address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	// create ota coin
	txParam := transaction.TxSalaryOutputParams{Amount: refundContribution.ContributedAmount, ReceiverAddress: &receiverAddr, TokenID: tokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	// set shareRandom for metadata
	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, makeMD)
}

func buildMatchedAndReturnedContributionTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	Logger.log.Info("[PDE Matched and Returned contribution] Starting...")
	contentBytes := []byte(contentStr)
	var matchedNReturnedContribution metadata.PDEMatchedNReturnedContribution
	err := json.Unmarshal(contentBytes, &matchedNReturnedContribution)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde matched and  returned contribution content: %+v", err)
		return nil, nil
	}
	if matchedNReturnedContribution.ShardID != shardID {
		return nil, nil
	}
	if matchedNReturnedContribution.ReturnedContributedAmount == 0 {
		return nil, nil
	}

	meta := metadata.NewPDEContributionResponse(
		common.PDEContributionMatchedNReturnedChainStatus,
		matchedNReturnedContribution.TxReqID,
		matchedNReturnedContribution.TokenIDStr,
		metadata.PDEContributionResponseMeta,
	)
	tokenIDStr := matchedNReturnedContribution.TokenIDStr
	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while converting tokenid to hash: %+v", err)
		return nil, nil
	}
	keyWallet, err := wallet.Base58CheckDeserialize(matchedNReturnedContribution.ContributorAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing contributor address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	// create ota coin
	txParam := transaction.TxSalaryOutputParams{Amount: matchedNReturnedContribution.ReturnedContributedAmount, ReceiverAddress: &receiverAddr, TokenID: tokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	// set shareRandom for metadata
	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, makeMD)
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
	if matchContributionValue.ShardID() != shardID {
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
		matchInst.NfctID(), 1,
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
	nfctIDs map[string]bool,
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
	if !nfctIDs[matchAndReturnInst.NfctID().String()] {
		receiveAddress := privacy.OTAReceiver{}
		err = receiveAddress.FromString(matchAndReturnContributionValue.ReceiveAddress())
		if err != nil {
			return res, err
		}
		tx0, err := buildMintTokenTxs(
			matchAndReturnInst.NfctID(), 1,
			receiveAddress, producerPrivateKey, transactionStateDB, metaData,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
			return res, err
		}
		res = append(res, tx0)
		nfctIDs[matchAndReturnInst.NfctID().String()] = true
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
