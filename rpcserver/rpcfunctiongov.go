package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/rpcserver/jsonresult"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
)

func (rpcServer RpcServer) getBondTypes() (*jsonresult.GetBondTypeResult, error) {
	db := *rpcServer.config.Database
	soldBondTypesBytesArr, err := db.GetSoldBondTypes()
	if err != nil {
		return nil, err
	}

	result := &jsonresult.GetBondTypeResult{
		BondTypes: make(map[string]jsonresult.GetBondTypeResultItem),
	}
	for _, soldBondTypesBytes := range soldBondTypesBytesArr {
		var bondInfo component.SellingBonds
		err = json.Unmarshal(soldBondTypesBytes, &bondInfo)
		if err != nil {
			return nil, err
		}
		bondIDStr := bondInfo.GetID().String()
		bondType := jsonresult.GetBondTypeResultItem{
			BondName:       bondInfo.BondName,
			BondSymbol:     bondInfo.BondSymbol,
			BondID:         bondIDStr,
			StartSellingAt: bondInfo.StartSellingAt,
			EndSellingAt:   bondInfo.StartSellingAt + bondInfo.SellingWithin,
			Maturity:       bondInfo.Maturity,
			BuyBackPrice:   bondInfo.BuyBackPrice,
			BuyPrice:       bondInfo.BondPrice,
			TotalIssue:     bondInfo.TotalIssue,
			Available:      bondInfo.BondsToSell,
		}
		result.BondTypes[bondIDStr] = bondType
	}
	return result, nil
}

func (rpcServer RpcServer) handleGetOracleTokenIDs(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	bondTypesRes, err := rpcServer.getBondTypes()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	oracleBonds := make([]*jsonresult.OracleToken, len(bondTypesRes.BondTypes))
	i := 0
	for _, bondType := range bondTypesRes.BondTypes {
		oracleBonds[i] = &jsonresult.OracleToken{
			TokenID:   bondType.BondID,
			TokenName: bondType.BondName,
		}
		i += 1
	}
	oracleTokens := []*jsonresult.OracleToken{
		&jsonresult.OracleToken{
			TokenID:   common.USDAssetID.String(),
			TokenName: "USD",
		},
		&jsonresult.OracleToken{
			TokenID:   common.ETHAssetID.String(),
			TokenName: "ETH",
		},
		&jsonresult.OracleToken{
			TokenID:   common.BTCAssetID.String(),
			TokenName: "BTC",
		},
		&jsonresult.OracleToken{
			TokenID:   common.ConstantID.String(),
			TokenName: "Constant",
		},
		&jsonresult.OracleToken{
			TokenID:   common.GOVTokenID.String(),
			TokenName: "GOV",
		},
		&jsonresult.OracleToken{
			TokenID:   common.DCBTokenID.String(),
			TokenName: "DCB",
		},
	}
	oracleTokens = append(oracleTokens, oracleBonds...)
	result := &jsonresult.GetOracleTokensResult{
		OracleTokens: oracleTokens,
	}
	return result, nil
}

func (rpcServer RpcServer) handleGetBondTypes(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result, err := rpcServer.getBondTypes()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return result, nil
}

func (rpcServer RpcServer) handleGetCurrentSellingBondTypes(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	stabilityInfo := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo
	sellingBondsParam := stabilityInfo.GOVConstitution.GOVParams.SellingBonds
	if sellingBondsParam == nil {
		return nil, nil
	}
	buyPrice := uint64(0)
	bondID := sellingBondsParam.GetID()
	bondPriceFromOracle := stabilityInfo.Oracle.Bonds[bondID.String()]
	if bondPriceFromOracle == 0 {
		buyPrice = sellingBondsParam.BondPrice
	} else {
		buyPrice = bondPriceFromOracle
	}

	bondTypeRes := jsonresult.GetBondTypeResultItem{
		BondName:       sellingBondsParam.BondName,
		BondSymbol:     sellingBondsParam.BondSymbol,
		BondID:         bondID.String(),
		StartSellingAt: sellingBondsParam.StartSellingAt,
		EndSellingAt:   sellingBondsParam.StartSellingAt + sellingBondsParam.SellingWithin,
		Maturity:       sellingBondsParam.Maturity,
		BuyBackPrice:   sellingBondsParam.BuyBackPrice, // in constant
		BuyPrice:       buyPrice,                       // in constant
		TotalIssue:     sellingBondsParam.TotalIssue,
		Available:      sellingBondsParam.BondsToSell,
	}
	result := jsonresult.GetBondTypeResult{
		BondTypes: make(map[string]jsonresult.GetBondTypeResultItem),
	}
	result.BondTypes[bondID.String()] = bondTypeRes
	return result, nil
}

func (rpcServer RpcServer) handleGetGOVParams(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constitution := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.GOVConstitution
	govParams := constitution.GOVParams
	results := make(map[string]interface{})
	results["GOVParams"] = govParams
	results["ExecuteDuration"] = constitution.ExecuteDuration
	results["Explanation"] = constitution.Explanation
	return results, nil
}

func (rpcServer RpcServer) handleGetGOVConstitution(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constitution := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.GOVConstitution
	return constitution, nil
}

func (rpcServer RpcServer) handleGetListGOVBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	res := ListPaymentAddressToListString(rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.GOVGovernor.BoardPaymentAddress)
	return res, nil
}

func (rpcServer RpcServer) handleGetListGOVBoardPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	res := []string{}
	listPayment := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.GOVGovernor.BoardPaymentAddress
	for _, i := range listPayment {
		wtf := wallet.KeyWallet{}
		wtf.KeySet.PaymentAddress = i
		res = append(res, wtf.Base58CheckSerialize(wallet.PaymentAddressType))
	}
	return res, nil
}

func (rpcServer RpcServer) handleAppendListGOVBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	senderKey := arrayParams[0].(string)
	paymentAddress, _ := metadata.GetPaymentAddressFromSenderKeyParams(senderKey)
	rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress = append(rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress, *paymentAddress)
	res := ListPaymentAddressToListString(rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.GOVGovernor.BoardPaymentAddress)
	return res, nil
}

func (rpcServer RpcServer) handleCreateRawTxWithBuyBackRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	paymentAddr := senderKey.KeySet.PaymentAddress
	tokenParamsRaw := arrayParams[4].(map[string]interface{})
	_, voutsAmount := transaction.CreateCustomTokenReceiverArray(tokenParamsRaw["TokenReceivers"])

	meta := metadata.NewBuyBackRequest(
		paymentAddr,
		uint64(voutsAmount),
		metadata.BuyBackRequestMeta,
	)
	customTokenTx, rpcErr := rpcServer.buildRawCustomTokenTransaction(params, meta)
	// rpcErr := err1.(*RPCError)
	if rpcErr != nil {
		Logger.log.Error(rpcErr)
		return nil, rpcErr
	}

	byteArrays, err := json.Marshal(customTokenTx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            customTokenTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateAndSendTxWithBuyBackRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawTxWithBuyBackRequest(params, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := rpcServer.handleSendRawCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	txID := sendResult.(*common.Hash)
	result := jsonresult.CreateTransactionResult{
		// TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
		TxID: txID.String(),
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateRawTxWithBuySellRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

	// Req param #5: buy/sell request info
	buySellReq := arrayParams[4].(map[string]interface{})

	paymentAddressP := buySellReq["PaymentAddress"].(string)
	key, _ := wallet.Base58CheckDeserialize(paymentAddressP)
	tokenIDStr := buySellReq["TokenID"].(string)
	tokenID, _ := common.Hash{}.NewHashFromStr(tokenIDStr)
	amount := uint64(buySellReq["Amount"].(float64))
	buyPrice := uint64(buySellReq["BuyPrice"].(float64))
	metaType := metadata.BuyFromGOVRequestMeta
	meta := metadata.NewBuySellRequest(
		key.KeySet.PaymentAddress,
		*tokenID,
		amount,
		buyPrice,
		metaType,
	)
	normalTx, err := rpcServer.buildRawTransaction(params, meta)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err1 := json.Marshal(normalTx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateAndSendTxWithBuySellRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawTxWithBuySellRequest(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateRawVoteGOVBoardTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	// params = setBuildRawBurnTransactionParams(params, FeeVote)
	arrayParams := common.InterfaceSlice(params)
	arrayParams[1] = nil
	return rpcServer.createRawCustomTokenTxWithMetadata(arrayParams, closeChan, metadata.NewVoteGOVBoardMetadataFromRPC)
}

func (rpcServer RpcServer) handleCreateAndSendVoteGOVBoardTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawVoteGOVBoardTransaction,
		RpcServer.handleSendRawCustomTokenTransaction,
	)
}

func (rpcServer RpcServer) handleCreateRawSubmitGOVProposalTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	params, err := rpcServer.buildParamsSubmitGOVProposal(params)
	if err != nil {
		return nil, err
	}
	return rpcServer.createRawTxWithMetadata(
		params,
		closeChan,
		metadata.NewSubmitGOVProposalMetadataFromRPC,
	)
}

func (rpcServer RpcServer) handleCreateAndSendSubmitGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawSubmitGOVProposalTransaction,
		RpcServer.handleSendRawTransaction,
	)
}

func (rpcServer RpcServer) handleCreateRawTxWithOracleFeed(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	feederAddr := senderKey.KeySet.PaymentAddress

	// Req param #4: oracle feed
	oracleFeed := arrayParams[4].(map[string]interface{})

	assetTypeStr := oracleFeed["AssetType"].(string)
	assetType, _ := common.Hash{}.NewHashFromStr(assetTypeStr)

	price := uint64(oracleFeed["Price"].(float64))
	metaType := metadata.OracleFeedMeta

	meta := metadata.NewOracleFeed(
		*assetType,
		price,
		metaType,
		feederAddr,
	)

	normalTx, err := rpcServer.buildRawTransaction(params, meta)
	rpcErr := err.(*RPCError)
	if rpcErr != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err1 := json.Marshal(normalTx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateAndSendTxWithOracleFeed(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawTxWithOracleFeed(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateRawTxWithUpdatingOracleBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	// Req param #4: updating oracle board info
	updatingOracleBoard := arrayParams[4].(map[string]interface{})
	action := int8(updatingOracleBoard["Action"].(float64))
	oraclePubKeys := updatingOracleBoard["OraclePubKeys"].([]interface{})
	assertedOraclePKs := []string{}
	for _, pk := range oraclePubKeys {
		hexStrPk := pk.(string)
		// pkBytes, _ := hex.DecodeString(hexStrPk)
		assertedOraclePKs = append(assertedOraclePKs, hexStrPk)
	}
	signs := updatingOracleBoard["Signs"].(map[string]interface{})
	assertedSigns := map[string][]byte{}
	for k, s := range signs {
		hexStrSign := s.(string)
		signBytes, _ := hex.DecodeString(hexStrSign)
		assertedSigns[k] = signBytes
	}
	metaType := metadata.UpdatingOracleBoardMeta
	meta := metadata.NewUpdatingOracleBoard(
		action,
		assertedOraclePKs,
		assertedSigns,
		metaType,
	)

	normalTx, err := rpcServer.buildRawTransaction(params, meta)
	// rpcErr := err.(*RPCError)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err1 := json.Marshal(normalTx)
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateAndSendTxWithUpdatingOracleBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawTxWithUpdatingOracleBoard(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

func (self RpcServer) handleCreateRawTxWithSenderAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

	hasPrivacy := int(arrayParams[3].(float64)) > 0
	if hasPrivacy {
		return nil, NewRPCError(ErrUnexpected, errors.New("Could not stick sender address to metadata when enabling privacy feature."))
	}

	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	senderAddr := senderKey.KeySet.PaymentAddress
	metaType := metadata.WithSenderAddressMeta

	meta := metadata.NewWithSenderAddress(senderAddr, metaType)
	normalTx, err := self.buildRawTransaction(params, meta)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err1 := json.Marshal(normalTx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendTxWithSenderAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawTxWithSenderAddress(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := self.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

func (rpcServer RpcServer) handleGetCurrentSellingGOVTokens(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	stabilityInfo := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo
	sellingGOVTokensParam := stabilityInfo.GOVConstitution.GOVParams.SellingGOVTokens
	if sellingGOVTokensParam == nil {
		return nil, nil
	}

	buyPrice := uint64(0)
	govTokenPriceFromOracle := stabilityInfo.Oracle.GOVToken
	if govTokenPriceFromOracle == 0 {
		buyPrice = sellingGOVTokensParam.GOVTokenPrice
	} else {
		buyPrice = govTokenPriceFromOracle
	}

	result := jsonresult.GetCurrentSellingGOVTokens{
		GOVTokenID:     common.GOVTokenID.String(),
		StartSellingAt: sellingGOVTokensParam.StartSellingAt,
		EndSellingAt:   sellingGOVTokensParam.StartSellingAt + sellingGOVTokensParam.SellingWithin,
		BuyPrice:       buyPrice, // in constant
		TotalIssue:     sellingGOVTokensParam.TotalIssue,
		Available:      sellingGOVTokensParam.GOVTokensToSell,
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateRawTxWithBuyGOVTokensRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

	// Req param #5: buy gov tokens request info
	buyGOVTokensReq := arrayParams[4].(map[string]interface{})

	paymentAddressP := buyGOVTokensReq["PaymentAddress"].(string)
	key, _ := wallet.Base58CheckDeserialize(paymentAddressP)
	tokenIDStr := buyGOVTokensReq["TokenID"].(string)
	tokenID, _ := common.Hash{}.NewHashFromStr(tokenIDStr)
	amount := uint64(buyGOVTokensReq["Amount"].(float64))
	buyPrice := uint64(buyGOVTokensReq["BuyPrice"].(float64))
	metaType := metadata.BuyGOVTokenRequestMeta
	meta := metadata.NewBuyGOVTokenRequest(
		key.KeySet.PaymentAddress,
		*tokenID,
		amount,
		buyPrice,
		metaType,
	)
	normalTx, err := rpcServer.buildRawTransaction(params, meta)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err1 := json.Marshal(normalTx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateAndSendTxWithBuyGOVTokensRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawTxWithBuyGOVTokensRequest(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

func (rpcServer RpcServer) handleGetCurrentOracleNetworkParams(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	stabilityInfo := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo
	oracleNetwork := stabilityInfo.GOVConstitution.GOVParams.OracleNetwork
	oracleNetworkResult := jsonresult.OracleNetworkResult{
		WrongTimesAllowed:      oracleNetwork.WrongTimesAllowed,
		Quorum:                 oracleNetwork.Quorum,
		AcceptableErrorMargin:  oracleNetwork.AcceptableErrorMargin,
		UpdateFrequency:        oracleNetwork.UpdateFrequency,
		OracleRewardMultiplier: oracleNetwork.OracleRewardMultiplier,
		OraclePubKeys:          oracleNetwork.OraclePubKeys,
	}
	// if oracleNetwork != nil {
	// 	oraclePubKeys := oracleNetwork.OraclePubKeys
	// 	oracleNetworkResult.OraclePubKeys = make([]string, len(oraclePubKeys))
	// 	for idx, pkBytes := range oraclePubKeys {
	// 		oracleNetworkResult.OraclePubKeys[idx] = hex.EncodeToString(pkBytes)
	// 	}
	// }
	return oracleNetworkResult, nil
}

func (rpcServer RpcServer) handleGetCurrentStabilityInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	stabilityInfo := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo
	return stabilityInfo, nil
}

func (rpcServer RpcServer) handleSignUpdatingOracleBoardContent(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	action := int8(arrayParams[1].(float64))        // action
	oraclePubKeys := arrayParams[2].([]interface{}) // OraclePubKeys
	assertedOraclePKs := []string{}
	for _, pk := range oraclePubKeys {
		hexStrPk := pk.(string)
		// pkBytes, _ := hex.DecodeString(hexStrPk)
		assertedOraclePKs = append(assertedOraclePKs, hexStrPk)
	}
	record := string(action)
	for _, pk := range assertedOraclePKs {
		record += pk
	}
	record += common.HashH([]byte(strconv.Itoa(metadata.UpdatingOracleBoardMeta))).String()
	hash := common.HashH([]byte(record))
	hashContent := hash[:]

	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	privKey := senderKey.KeySet.PrivateKey

	sk := new(big.Int).SetBytes(privKey[:privacy.BigIntSize])
	r := new(big.Int).SetBytes(privKey[privacy.BigIntSize:])
	sigKey := new(privacy.SchnPrivKey)
	sigKey.Set(sk, r)

	// signing
	signature, _ := sigKey.Sign(hashContent)

	// convert signature to byte array
	signatureBytes := signature.Bytes()
	signStr := hex.EncodeToString(signatureBytes)
	return signStr, nil
}

func (rpcServer RpcServer) handleGetAssetPrice(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	assetIDRaw := arrayParams[0].(string)
	assetID, err := common.NewHashFromStr(assetIDRaw)
	if err != nil {
		return uint64(0), nil
	}
	return rpcServer.config.BlockChain.BestState.Beacon.GetAssetPrice(*assetID), nil
}
