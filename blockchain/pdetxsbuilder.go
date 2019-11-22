package blockchain

import (
	"encoding/base64"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
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
	return &pdeTradeAcceptedContent, nil
}

func buildTradeResTx(
	instStatus string,
	receiverAddressStr string,
	receiveAmt uint64,
	tokenIDStr string,
	requestedTxID common.Hash,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	db database.DatabaseInterface,
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
	keyWallet, err := wallet.Base58CheckDeserialize(receiverAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing trader address string: %+v", err)
		return nil, err
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	// the returned currency is PRV
	if tokenIDStr == common.PRVCoinID.String() {
		resTx := new(transaction.Tx)
		err = resTx.InitTxSalary(
			receiveAmt,
			&receiverAddr,
			producerPrivateKey,
			db,
			meta,
		)
		if err != nil {
			return nil, NewBlockChainError(InitPDETradeResponseTransactionError, err)
		}
		//modify the type of the salary transaction
		// resTx.Type = common.TxBlockProducerCreatedType
		return resTx, nil
	}

	// in case the returned currency is privacy custom token
	receiver := &privacy.PaymentInfo{
		Amount:         receiveAmt,
		PaymentAddress: receiverAddr,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], tokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID: propID.String(),
		// PropertyName:   issuingAcceptedInst.IncTokenName,
		// PropertySymbol: issuingAcceptedInst.IncTokenName,
		Amount:      receiveAmt,
		TokenTxType: transaction.CustomTokenInit,
		Receiver:    []*privacy.PaymentInfo{receiver},
		TokenInput:  []*privacy.InputCoin{},
		Mintable:    true,
	}
	resTx := &transaction.TxCustomTokenPrivacy{}
	initErr := resTx.Init(
		transaction.NewTxPrivacyTokenInitParams(
			producerPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			tokenParams,
			db,
			meta,
			false,
			false,
			shardID,
			nil,
		),
	)
	if initErr != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing trade response tx: %+v", initErr)
		return nil, initErr
	}
	return resTx, nil
}

func (blockGenerator *BlockGenerator) buildPDETradeRefundTx(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
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
		pdeTradeRequestAction.Meta.SellAmount+pdeTradeRequestAction.Meta.TradingFee,
		pdeTradeRequestAction.Meta.TokenIDToSellStr,
		pdeTradeRequestAction.TxReqID,
		producerPrivateKey,
		shardID,
		blockGenerator.chain.config.DataBase,
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
		pdeTradeAcceptedContent.ReceiveAmount,
		pdeTradeAcceptedContent.TokenIDToBuyStr,
		pdeTradeAcceptedContent.RequestedTxID,
		producerPrivateKey,
		shardID,
		blockGenerator.chain.config.DataBase,
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
) (metadata.Transaction, error) {
	Logger.log.Info("[PDE Trade] Starting...")
	if instStatus == "refund" {
		return blockGenerator.buildPDETradeRefundTx(
			instStatus,
			contentStr,
			producerPrivateKey,
			shardID,
		)
	}
	return blockGenerator.buildPDETradeAcceptedTx(
		instStatus,
		contentStr,
		producerPrivateKey,
		shardID,
	)
}

func (blockGenerator *BlockGenerator) buildPDEWithdrawalTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
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
	// the returned currency is PRV
	if withdrawalTokenIDStr == common.PRVCoinID.String() {
		resTx := new(transaction.Tx)
		err = resTx.InitTxSalary(
			wdAcceptedContent.DeductingPoolValue,
			&receiverAddr,
			producerPrivateKey,
			blockGenerator.chain.config.DataBase,
			meta,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while initializing withdrawal (normal) tx: %+v", err)
			return nil, nil
		}
		//modify the type of the salary transaction
		// resTx.Type = common.TxBlockProducerCreatedType
		return resTx, nil
	}

	// in case the returned currency is privacy custom token
	receiver := &privacy.PaymentInfo{
		Amount:         wdAcceptedContent.DeductingPoolValue,
		PaymentAddress: receiverAddr,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], tokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID: propID.String(),
		// PropertyName:   tokeName,
		// PropertySymbol: tokenSymbol,
		Amount:      wdAcceptedContent.DeductingPoolValue,
		TokenTxType: transaction.CustomTokenInit,
		Receiver:    []*privacy.PaymentInfo{receiver},
		TokenInput:  []*privacy.InputCoin{},
		Mintable:    true,
	}
	resTx := &transaction.TxCustomTokenPrivacy{}
	initErr := resTx.Init(
		transaction.NewTxPrivacyTokenInitParams(
			producerPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			tokenParams,
			blockGenerator.chain.config.DataBase,
			meta,
			false,
			false,
			shardID,
			nil,
		),
	)
	if initErr != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing withdrawal response (privacy custom token) tx: %+v", initErr)
		return nil, nil
	}
	return resTx, nil
}

func (blockGenerator *BlockGenerator) buildPDERefundContributionTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
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
	// the returned currency is PRV
	if refundTokenIDStr == common.PRVCoinID.String() {
		resTx := new(transaction.Tx)
		err = resTx.InitTxSalary(
			refundContribution.ContributedAmount,
			&receiverAddr,
			producerPrivateKey,
			blockGenerator.chain.config.DataBase,
			meta,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while initializing refund contribution (normal) tx: %+v", err)
			return nil, nil
		}
		//modify the type of the salary transaction
		// resTx.Type = common.TxBlockProducerCreatedType
		return resTx, nil
	}

	// in case the returned currency is privacy custom token
	receiver := &privacy.PaymentInfo{
		Amount:         refundContribution.ContributedAmount,
		PaymentAddress: receiverAddr,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], tokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID: propID.String(),
		// PropertyName:   issuingAcceptedInst.IncTokenName,
		// PropertySymbol: issuingAcceptedInst.IncTokenName,
		Amount:      refundContribution.ContributedAmount,
		TokenTxType: transaction.CustomTokenInit,
		Receiver:    []*privacy.PaymentInfo{receiver},
		TokenInput:  []*privacy.InputCoin{},
		Mintable:    true,
	}
	resTx := &transaction.TxCustomTokenPrivacy{}
	initErr := resTx.Init(
		transaction.NewTxPrivacyTokenInitParams(
			producerPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			tokenParams,
			blockGenerator.chain.config.DataBase,
			meta,
			false,
			false,
			shardID,
			nil,
		),
	)
	if initErr != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing refund contribution response (privacy custom token) tx: %+v", initErr)
		return nil, nil
	}
	return resTx, nil
}
