package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy/coin"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

func parseTradeRefundContent(
	contentStr string,
) (*metadata.PDETradeRequestAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of pde trade refund instruction: %+v", err)
		return nil, err
	}
	var pdeTradeRequestAction metadata.PDETradeRequestAction
	err = json.Unmarshal(contentBytes, &pdeTradeRequestAction)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde trade refund content: %+v", err)
		return nil, err
	}
	fmt.Println("Trade Request TxRandom ", pdeTradeRequestAction.Meta.TxRandomStr)
	fmt.Println("Trade Request TradeAddress ", pdeTradeRequestAction.Meta.TraderAddressStr)
	fmt.Println("Trade Request TradeAmount ", pdeTradeRequestAction.Meta.MinAcceptableAmount)
	fmt.Println("Trade Request Request Tx ", pdeTradeRequestAction.TxReqID)

	return &pdeTradeRequestAction, nil
}

func parseTradeAcceptedContent(
	contentStr string,
) (*metadata.PDETradeAcceptedContent, error) {
	contentBytes := []byte(contentStr)
	var pdeTradeAcceptedContent metadata.PDETradeAcceptedContent
	err := json.Unmarshal(contentBytes, &pdeTradeAcceptedContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde trade accepted content: %+v", err)
		return nil, err
	}
	fmt.Println("Trade Request TxRandom ", pdeTradeAcceptedContent.TxRandomStr)
	fmt.Println("Trade Request TradeAddress ", pdeTradeAcceptedContent.TraderAddressStr)
	fmt.Println("Trade Request TradeAmount ", pdeTradeAcceptedContent.ReceiveAmount)
	fmt.Println("Trade Request RequestID ", pdeTradeAcceptedContent.ReceiveAmount)
	return &pdeTradeAcceptedContent, nil
}

func buildTradeResTx(
	instStatus string,
	receiverAddressStr string,
	txRandomStr string,
	receiveAmt uint64,
	tokenIDStr string,
	requestedTxID common.Hash,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
	bridgeStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	meta := metadata.NewPDETradeResponse(
		instStatus,
		requestedTxID,
		metadata.PDETradeResponseMeta,
	)
	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while converting tokenid to hash: %+v", err)
		return nil, err
	}

	var otaCoin *privacy.CoinV2
	if len(txRandomStr) > 0 {
		publickey, txRandom, err := coin.ParseOTAInfoFromString(receiverAddressStr, txRandomStr)
		if err != nil {
			return nil, err
		}
		otaCoin = coin.NewCoinFromAmountAndTxRandomBytes(receiveAmt, publickey, txRandom, []byte{})

	} else {
		return nil, errors.New("Cannnot create trade response without txRandom info")
	}
	if tokenIDStr == common.PRVCoinID.String() {
		return BuildInitTxSalaryTx(otaCoin, producerPrivateKey, transactionStateDB, meta)
	} else {
		return BuildInitTxTokenSalaryTx(otaCoin, producerPrivateKey, transactionStateDB, meta, tokenID)
	}
}

func (blockGenerator *BlockGenerator) buildPDETradeRefundTx(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	shardView *ShardBestState,
	beaconView *BeaconBestState,
) (metadata.Transaction, error) {
	pdeTradeRequestAction, err := parseTradeRefundContent(contentStr)
	if err != nil {
		return nil, nil
	}
	if shardID != pdeTradeRequestAction.ShardID {
		return nil, nil
	}

	resTx, err := buildTradeResTx(
		instStatus,
		pdeTradeRequestAction.Meta.TraderAddressStr,
		pdeTradeRequestAction.Meta.TxRandomStr,
		pdeTradeRequestAction.Meta.SellAmount+pdeTradeRequestAction.Meta.TradingFee,
		pdeTradeRequestAction.Meta.TokenIDToSellStr,
		pdeTradeRequestAction.TxReqID,
		producerPrivateKey,
		shardID,
		shardView.GetCopiedTransactionStateDB(),
		beaconView.GetBeaconFeatureStateDB(),
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing refunded trading response tx: %+v", err)
		return nil, nil
	}
	Logger.log.Info("[PDE Trade] Create refunded tx ok.")
	return resTx, nil
}

func (blockGenerator *BlockGenerator) buildPDETradeAcceptedTx(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	shardView *ShardBestState,
	beaconView *BeaconBestState,
) (metadata.Transaction, error) {
	pdeTradeAcceptedContent, err := parseTradeAcceptedContent(contentStr)
	if err != nil {
		return nil, nil
	}
	if shardID != pdeTradeAcceptedContent.ShardID {
		return nil, nil
	}
	resTx, err := buildTradeResTx(
		instStatus,
		pdeTradeAcceptedContent.TraderAddressStr,
		pdeTradeAcceptedContent.TxRandomStr,
		pdeTradeAcceptedContent.ReceiveAmount,
		pdeTradeAcceptedContent.TokenIDToBuyStr,
		pdeTradeAcceptedContent.RequestedTxID,
		producerPrivateKey,
		shardID,
		shardView.GetCopiedTransactionStateDB(),
		beaconView.GetBeaconFeatureStateDB(),
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing accepted trading response tx: %+v", err)
		return nil, nil
	}
	Logger.log.Info("[PDE Trade] Create accepted tx ok.")
	return resTx, nil
}

func (blockGenerator *BlockGenerator) buildPDETradeIssuanceTx(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	shardView *ShardBestState,
	beaconView *BeaconBestState,
) (metadata.Transaction, error) {
	Logger.log.Info("[PDE Trade] Starting...")
	if instStatus == common.PDETradeRefundChainStatus {
		return blockGenerator.buildPDETradeRefundTx(
			instStatus,
			contentStr,
			producerPrivateKey,
			shardID,
			shardView,
			beaconView,
		)
	}
	return blockGenerator.buildPDETradeAcceptedTx(
		instStatus,
		contentStr,
		producerPrivateKey,
		shardID,
		shardView,
		beaconView,
	)
}

func (blockGenerator *BlockGenerator) buildPDEWithdrawalTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	shardView *ShardBestState,
	beaconView *BeaconBestState,
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

	otaCoin, err := coin.NewCoinFromAmountAndReceiver(wdAcceptedContent.DeductingPoolValue, receiverAddr)
	if err != nil {
		Logger.log.Errorf("Cannot get new coin from amount and payment address")
		return nil, err
	}
	meta.SetSharedRandom(otaCoin.GetSharedRandom().ToBytesS())

	if withdrawalTokenIDStr == common.PRVCoinID.String() {
		return BuildInitTxSalaryTx(otaCoin, producerPrivateKey, shardView.GetCopiedTransactionStateDB(), meta)
	} else {
		return BuildInitTxTokenSalaryTx(otaCoin, producerPrivateKey, shardView.GetCopiedTransactionStateDB(), meta, tokenID)
	}
}

func (blockGenerator *BlockGenerator) buildPDERefundContributionTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	shardView *ShardBestState,
	beaconView *BeaconBestState,
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
	otaCoin, err := coin.NewCoinFromAmountAndReceiver(refundContribution.ContributedAmount, receiverAddr)
	if err != nil {
		Logger.log.Errorf("Cannot get new coin from amount and receiver")
		return nil, err
	}
	// set shareRandom for metadata
	meta.SetSharedRandom(otaCoin.GetSharedRandom().ToBytesS())

	if refundTokenIDStr == common.PRVCoinID.String() {
		return BuildInitTxSalaryTx(otaCoin, producerPrivateKey, shardView.GetCopiedTransactionStateDB(), meta)
	} else {
		return BuildInitTxTokenSalaryTx(otaCoin, producerPrivateKey, shardView.GetCopiedTransactionStateDB(), meta, tokenID)
	}
}

func (blockGenerator *BlockGenerator) buildPDEMatchedNReturnedContributionTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	shardView *ShardBestState,
	beaconView *BeaconBestState,
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
	otaCoin, err := coin.NewCoinFromAmountAndReceiver(matchedNReturnedContribution.ReturnedContributedAmount, receiverAddr)
	if err != nil {
		Logger.log.Errorf("Cannot get new coin from amount and receiver")
		return nil, err
	}
	// set shareRandom for metadata
	meta.SetSharedRandom(otaCoin.GetSharedRandom().ToBytesS())

	if tokenIDStr == common.PRVCoinID.String() {
		return BuildInitTxSalaryTx(otaCoin, producerPrivateKey, shardView.GetCopiedTransactionStateDB(), meta)
	} else {
		return BuildInitTxTokenSalaryTx(otaCoin, producerPrivateKey, shardView.GetCopiedTransactionStateDB(), meta, tokenID)
	}
}
