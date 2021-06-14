package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

// NOTE: for whole bridge's deposit process, anytime an error occurs it will be logged for debugging and the request will be skipped for retry later. No error will be returned so that the network can still continue to process others.

func buildInstruction(metaType int, shardID byte, instStatus string, contentStr string) []string {
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		instStatus,
		contentStr,
	}
}

func getShardIDFromPaymentAddress(addressStr string) (byte, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(addressStr)
	if err != nil {
		return byte(0), err
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return byte(0), errors.New("Payment address' public key must not be empty")
	}
	// calculate shard ID
	lastByte := keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1]
	shardID := common.GetShardIDFromLastByte(lastByte)
	return shardID, nil
}

func (blockchain *BlockChain) buildInstructionsForContractingReq(
	contentStr string,
	shardID byte,
	metaType int,
) ([][]string, error) {
	inst := buildInstruction(metaType, shardID, "accepted", contentStr)
	return [][]string{inst}, nil
}

func (blockchain *BlockChain) buildInstructionsForIssuingReq(
	beaconBestState *BeaconBestState,
	stateDB *statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	ac *metadata.AccumulatedValues,
) ([][]string, error) {
	Logger.log.Info("[Centralized bridge token issuance] Starting...")
	instructions := [][]string{}
	issuingReqAction, err := metadata.ParseIssuingInstContent(contentStr)
	if err != nil {
		Logger.log.Info("WARNING: an issue occured while parsing issuing action content: ", err)
		return nil, nil
	}

	Logger.log.Infof("[Centralized bridge token issuance] Processing for tx: %s, tokenid: %s", issuingReqAction.TxReqID.String(), issuingReqAction.Meta.TokenID.String())
	issuingReq := issuingReqAction.Meta
	issuingTokenID := issuingReq.TokenID
	issuingTokenName := issuingReq.TokenName
	rejectedInst := buildInstruction(metaType, shardID, "rejected", issuingReqAction.TxReqID.String())

	if !ac.CanProcessCIncToken(issuingTokenID) {
		Logger.log.Warnf("WARNING: The issuing token (%s) was already used in the current block.", issuingTokenID.String())
		return append(instructions, rejectedInst), nil
	}

	privacyTokenExisted, err := blockchain.PrivacyTokenIDExistedInAllShards(beaconBestState, issuingTokenID)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while checking it can process for the incognito token or not: ", err)
		return append(instructions, rejectedInst), nil
	}
	ok, err := statedb.CanProcessCIncToken(stateDB, issuingTokenID, privacyTokenExisted)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while checking it can process for the incognito token or not: ", err)
		return append(instructions, rejectedInst), nil
	}
	if !ok {
		Logger.log.Warnf("WARNING: The issuing token (%s) was already used in the previous blocks.", issuingTokenID.String())
		return append(instructions, rejectedInst), nil
	}

	if len(issuingReq.ReceiverAddress.Pk) == 0 {
		Logger.log.Info("WARNING: invalid receiver address")
		return append(instructions, rejectedInst), nil
	}
	lastByte := issuingReq.ReceiverAddress.Pk[len(issuingReq.ReceiverAddress.Pk)-1]
	receivingShardID := common.GetShardIDFromLastByte(lastByte)

	issuingAcceptedInst := metadata.IssuingAcceptedInst{
		ShardID:         receivingShardID,
		DepositedAmount: issuingReq.DepositedAmount,
		ReceiverAddr:    issuingReq.ReceiverAddress,
		IncTokenID:      issuingTokenID,
		IncTokenName:    issuingTokenName,
		TxReqID:         issuingReqAction.TxReqID,
	}
	issuingAcceptedInstBytes, err := json.Marshal(issuingAcceptedInst)
	if err != nil {
		Logger.log.Info("WARNING: an error occured while marshaling issuingAccepted instruction: ", err)
		return append(instructions, rejectedInst), nil
	}

	ac.CBridgeTokens = append(ac.CBridgeTokens, &issuingTokenID)
	returnedInst := buildInstruction(metaType, shardID, "accepted", base64.StdEncoding.EncodeToString(issuingAcceptedInstBytes))
	Logger.log.Info("[Centralized bridge token issuance] Process finished without error...")
	return append(instructions, returnedInst), nil
}

type BridgeReqActionInfo struct {
	BlockHash  rCommon.Hash
	TxIndex    uint
	ProofStrs  []string
	IncTokenID common.Hash
	TxReqID    common.Hash
	TxReceipt  *types.Receipt
}

func (blockchain *BlockChain) buildInstructionsForIssuingBridgeReq(
	beaconBestState *BeaconBestState,
	stateDB *statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	ac *metadata.AccumulatedValues,
) ([][]string, error) {
	Logger.log.Info("[Decentralized bridge token issuance] Starting...")
	instructions := [][]string{}
	var issuingBridgeReqAction *BridgeReqActionInfo
	var listTxUsed [][]byte
	var contractAddress, prefix string
	if metaType == metadata.IssuingETHRequestMeta {
		issuingETHBridgeReqAction, err := metadata.ParseETHIssuingInstContent(contentStr)
		if err != nil {
			Logger.log.Warn("WARNING: an issue occured while parsing issuing action content: ", err)
			return nil, nil
		}
		issuingBridgeReqAction = &BridgeReqActionInfo{
			BlockHash:  issuingETHBridgeReqAction.Meta.BlockHash,
			TxIndex:    issuingETHBridgeReqAction.Meta.TxIndex,
			ProofStrs:  issuingETHBridgeReqAction.Meta.ProofStrs,
			IncTokenID: issuingETHBridgeReqAction.Meta.IncTokenID,
			TxReqID:    issuingETHBridgeReqAction.TxReqID,
			TxReceipt:  issuingETHBridgeReqAction.ETHReceipt,
		}
		listTxUsed = ac.UniqETHTxsUsed
		contractAddress = config.Param().EthContractAddressStr
		prefix = ""
	} else if metaType == metadata.IssuingBSCRequestMeta {
		issuingBSCBridgeReqAction, err := metadata.ParseBSCIssuingInstContent(contentStr)
		if err != nil {
			Logger.log.Warn("WARNING: an issue occured while parsing issuing action content: ", err)
			return nil, nil
		}
		issuingBridgeReqAction = &BridgeReqActionInfo{
			BlockHash:  issuingBSCBridgeReqAction.Meta.BlockHash,
			TxIndex:    issuingBSCBridgeReqAction.Meta.TxIndex,
			ProofStrs:  issuingBSCBridgeReqAction.Meta.ProofStrs,
			IncTokenID: issuingBSCBridgeReqAction.Meta.IncTokenID,
			TxReqID:    issuingBSCBridgeReqAction.TxReqID,
			TxReceipt:  issuingBSCBridgeReqAction.BSCReceipt,
		}
		listTxUsed = ac.UniqBSCTxsUsed
		contractAddress = config.Param().BscContractAddressStr
		prefix = common.BSCPrefix
	} else {
		return nil, nil
	}

	Logger.log.Infof("[Decentralized bridge token issuance] Processing for tx: %s, tokenid: %s", issuingBridgeReqAction.TxReqID.String(), issuingBridgeReqAction.IncTokenID.String())

	rejectedInst := buildInstruction(metaType, shardID, "rejected", issuingBridgeReqAction.TxReqID.String())

	txReceipt := issuingBridgeReqAction.TxReceipt
	if txReceipt == nil {
		Logger.log.Warn("WARNING: bridge tx receipt is null.")
		return append(instructions, rejectedInst), nil
	}

	// NOTE: since TxHash from constructedReceipt is always '0x0000000000000000000000000000000000000000000000000000000000000000'
	// so must build unique eth tx as combination of block hash and tx index.
	uniqTx := append(issuingBridgeReqAction.BlockHash[:], []byte(strconv.Itoa(int(issuingBridgeReqAction.TxIndex)))...)
	isUsedInBlock := IsBridgeTxHashUsedInBlock(uniqTx, listTxUsed)
	if isUsedInBlock {
		Logger.log.Warn("WARNING: already issued for the hash in current block: ", uniqTx)
		return append(instructions, rejectedInst), nil
	}
	var isIssued bool
	var err error
	if metaType == metadata.IssuingETHRequestMeta {
		isIssued, err = statedb.IsETHTxHashIssued(stateDB, uniqTx)
	} else {
		isIssued, err = statedb.IsBSCTxHashIssued(stateDB, uniqTx)
	}
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while checking the bridge tx hash is issued or not: ", err)
		return append(instructions, rejectedInst), nil
	}
	if isIssued {
		Logger.log.Warn("WARNING: already issued for the hash in previous blocks: ", uniqTx)
		return append(instructions, rejectedInst), nil
	}

	logMap, err := metadata.PickAndParseLogMapFromReceipt(txReceipt, contractAddress)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while parsing log map from receipt: ", err)
		return append(instructions, rejectedInst), nil
	}
	if logMap == nil {
		Logger.log.Warn("WARNING: could not find log map out from receipt")
		return append(instructions, rejectedInst), nil
	}

	logMapBytes, _ := json.Marshal(logMap)
	Logger.log.Warn("INFO: eth logMap json - ", string(logMapBytes))

	// the token might be ETH/ERC20 BNB/BEP20
	tokenAddr, ok := logMap["token"].(rCommon.Address)
	if !ok {
		Logger.log.Warn("WARNING: could not parse eth token id from log map.")
		return append(instructions, rejectedInst), nil
	}
	token := append([]byte(prefix), tokenAddr.Bytes()...)
	canProcess, err := ac.CanProcessTokenPair(token, issuingBridgeReqAction.IncTokenID)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while checking it can process for token pair on the current block or not: ", err)
		return append(instructions, rejectedInst), nil
	}
	if !canProcess {
		Logger.log.Warn("WARNING: pair of incognito token id & bridge's id is invalid in current block")
		return append(instructions, rejectedInst), nil
	}
	privacyTokenExisted, err := blockchain.PrivacyTokenIDExistedInAllShards(beaconBestState, issuingBridgeReqAction.IncTokenID)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while checking it can process for the incognito token or not: ", err)
		return append(instructions, rejectedInst), nil
	}
	isValid, err := statedb.CanProcessTokenPair(stateDB, token, issuingBridgeReqAction.IncTokenID, privacyTokenExisted)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while checking it can process for token pair on the previous blocks or not: ", err)
		return append(instructions, rejectedInst), nil
	}
	if !isValid {
		Logger.log.Warn("WARNING: pair of incognito token id & bridge's id is invalid with previous blocks")
		return append(instructions, rejectedInst), nil
	}

	addressStr, ok := logMap["incognitoAddress"].(string)
	if !ok {
		Logger.log.Warn("WARNING: could not parse incognito address from bridge log map.")
		return append(instructions, rejectedInst), nil
	}
	amt, ok := logMap["amount"].(*big.Int)
	if !ok {
		Logger.log.Warn("WARNING: could not parse amount from bridge log map.")
		return append(instructions, rejectedInst), nil
	}
	amount := uint64(0)
	if bytes.Equal(append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), token) {
		// convert amt from wei (10^18) to nano eth (10^9)
		amount = big.NewInt(0).Div(amt, big.NewInt(1000000000)).Uint64()
	} else { // ERC20 / BEP20
		amount = amt.Uint64()
	}

	receivingShardID, err := getShardIDFromPaymentAddress(addressStr)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while getting shard id from payment address: ", err)
		return append(instructions, rejectedInst), nil
	}

	var issuingAcceptedInstBytes []byte
	if metaType == metadata.IssuingETHRequestMeta {
		issuingETHAcceptedInst := metadata.IssuingETHAcceptedInst{
			ShardID:         receivingShardID,
			IssuingAmount:   amount,
			ReceiverAddrStr: addressStr,
			IncTokenID:      issuingBridgeReqAction.IncTokenID,
			TxReqID:         issuingBridgeReqAction.TxReqID,
			UniqETHTx:       uniqTx,
			ExternalTokenID: token,
		}
		issuingAcceptedInstBytes, err = json.Marshal(issuingETHAcceptedInst)
		if err != nil {
			Logger.log.Warn("WARNING: an error occured while marshaling issuingBridgeAccepted instruction: ", err)
			return append(instructions, rejectedInst), nil
		}
		ac.UniqETHTxsUsed = append(ac.UniqETHTxsUsed, uniqTx)
	} else {
		issuingBSCAcceptedInst := metadata.IssuingBSCAcceptedInst{
			ShardID:         receivingShardID,
			IssuingAmount:   amount,
			ReceiverAddrStr: addressStr,
			IncTokenID:      issuingBridgeReqAction.IncTokenID,
			TxReqID:         issuingBridgeReqAction.TxReqID,
			UniqBSCTx:       uniqTx,
			ExternalTokenID: token,
		}
		issuingAcceptedInstBytes, err = json.Marshal(issuingBSCAcceptedInst)
		if err != nil {
			Logger.log.Warn("WARNING: an error occured while marshaling issuingBridgeAccepted instruction: ", err)
			return append(instructions, rejectedInst), nil
		}
		ac.UniqBSCTxsUsed = append(ac.UniqBSCTxsUsed, uniqTx)
	}
	ac.DBridgeTokenPair[issuingBridgeReqAction.IncTokenID.String()] = token

	acceptedInst := buildInstruction(metaType, shardID, "accepted", base64.StdEncoding.EncodeToString(issuingAcceptedInstBytes))
	Logger.log.Info("[Decentralized bridge token issuance] Process finished without error...")
	return append(instructions, acceptedInst), nil
}

func (blockGenerator *BlockGenerator) buildIssuanceTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	shardView *ShardBestState,
	featureStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	Logger.log.Info("[Centralized bridge token issuance] Starting...")
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while decoding content string of accepted issuance instruction: ", err)
		return nil, nil
	}
	var issuingAcceptedInst metadata.IssuingAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while unmarshaling accepted issuance instruction: ", err)
		return nil, nil
	}

	Logger.log.Infof("[Centralized bridge token issuance] Processing for tx: %s", issuingAcceptedInst.TxReqID.String())

	if shardID != issuingAcceptedInst.ShardID {
		Logger.log.Infof("Ignore due to shardid difference, current shardid %d, receiver's shardid %d", shardID, issuingAcceptedInst.ShardID)
		return nil, nil
	}
	issuingRes := metadata.NewIssuingResponse(
		issuingAcceptedInst.TxReqID,
		metadata.IssuingResponseMeta,
	)
	receiver := &privacy.PaymentInfo{
		Amount:         issuingAcceptedInst.DepositedAmount,
		PaymentAddress: issuingAcceptedInst.ReceiverAddr,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], issuingAcceptedInst.IncTokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID:     propID.String(),
		PropertyName:   issuingAcceptedInst.IncTokenName,
		PropertySymbol: issuingAcceptedInst.IncTokenName,
		Amount:         issuingAcceptedInst.DepositedAmount,
		TokenTxType:    transaction.CustomTokenInit,
		Receiver:       []*privacy.PaymentInfo{receiver},
		TokenInput:     []*privacy.InputCoin{},
		Mintable:       true,
	}
	resTx := &transaction.TxCustomTokenPrivacy{}
	initErr := resTx.Init(
		transaction.NewTxPrivacyTokenInitParams(producerPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			tokenParams,
			shardView.GetCopiedTransactionStateDB(),
			issuingRes,
			false,
			false,
			shardID,
			nil,
			featureStateDB))

	if initErr != nil {
		Logger.log.Warn("WARNING: an error occured while initializing response tx: ", initErr)
		return nil, nil
	}
	Logger.log.Infof("[Centralized token issuance] Create tx ok: %s", resTx.Hash().String())
	return resTx, nil
}

type BridgeIssuanceTxInfo struct {
	ShardID         byte
	IssuingAmount   uint64
	ReceiverAddrStr string
	IncTokenID      common.Hash
	TxReqID         common.Hash
	UniqTx          []byte
	ExternalTokenID []byte
}

func (blockGenerator *BlockGenerator) buildBridgeIssuanceTx(
	instruction []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	shardView *ShardBestState,
	featureStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	Logger.log.Info("[Decentralized bridge token issuance] Starting...")
	if len(instruction) < 4 {
		return nil, nil // skip the instruction
	}
	contentBytes, err := base64.StdEncoding.DecodeString(instruction[3])
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while decoding content string of ETH accepted issuance instruction: ", err)
		return nil, nil
	}
	var bridgeIssuanceTxInfo *BridgeIssuanceTxInfo
	var issuingRes metadata.Metadata
	if instruction[0] == strconv.Itoa(metadata.IssuingETHRequestMeta) {
		var issuingETHAcceptedInst metadata.IssuingETHAcceptedInst
		err = json.Unmarshal(contentBytes, &issuingETHAcceptedInst)
		if err != nil {
			Logger.log.Warn("WARNING: an error occured while unmarshaling accepted issuance instruction: ", err)
			return nil, nil
		}
		bridgeIssuanceTxInfo = &BridgeIssuanceTxInfo{
			ShardID:         issuingETHAcceptedInst.ShardID,
			ReceiverAddrStr: issuingETHAcceptedInst.ReceiverAddrStr,
			IncTokenID:      issuingETHAcceptedInst.IncTokenID,
			TxReqID:         issuingETHAcceptedInst.TxReqID,
			ExternalTokenID: issuingETHAcceptedInst.ExternalTokenID,
			UniqTx:          issuingETHAcceptedInst.UniqETHTx,
			IssuingAmount:   issuingETHAcceptedInst.IssuingAmount,
		}
		issuingRes = metadata.NewIssuingETHResponse(
			bridgeIssuanceTxInfo.TxReqID,
			bridgeIssuanceTxInfo.UniqTx,
			bridgeIssuanceTxInfo.ExternalTokenID,
			metadata.IssuingETHResponseMeta,
		)
	} else if instruction[0] == strconv.Itoa(metadata.IssuingBSCRequestMeta) {
		var issuingBSCAcceptedInst metadata.IssuingBSCAcceptedInst
		err = json.Unmarshal(contentBytes, &issuingBSCAcceptedInst)
		if err != nil {
			Logger.log.Warn("WARNING: an error occured while unmarshaling accepted issuance instruction: ", err)
			return nil, nil
		}
		bridgeIssuanceTxInfo = &BridgeIssuanceTxInfo{
			ShardID:         issuingBSCAcceptedInst.ShardID,
			ReceiverAddrStr: issuingBSCAcceptedInst.ReceiverAddrStr,
			IncTokenID:      issuingBSCAcceptedInst.IncTokenID,
			TxReqID:         issuingBSCAcceptedInst.TxReqID,
			ExternalTokenID: issuingBSCAcceptedInst.ExternalTokenID,
			UniqTx:          issuingBSCAcceptedInst.UniqBSCTx,
			IssuingAmount:   issuingBSCAcceptedInst.IssuingAmount,
		}
		issuingRes = metadata.NewIssuingBSCResponse(
			bridgeIssuanceTxInfo.TxReqID,
			bridgeIssuanceTxInfo.UniqTx,
			bridgeIssuanceTxInfo.ExternalTokenID,
			metadata.IssuingBSCResponseMeta,
		)
	} else {
		return nil, nil
	}

	if shardID != bridgeIssuanceTxInfo.ShardID {
		Logger.log.Infof("Ignore due to shardid difference, current shardid %d, receiver's shardid %d", shardID, bridgeIssuanceTxInfo.ShardID)
		return nil, nil
	}
	key, err := wallet.Base58CheckDeserialize(bridgeIssuanceTxInfo.ReceiverAddrStr)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while deserializing receiver address string: ", err)
		return nil, nil
	}
	receiver := &privacy.PaymentInfo{
		Amount:         bridgeIssuanceTxInfo.IssuingAmount,
		PaymentAddress: key.KeySet.PaymentAddress,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], bridgeIssuanceTxInfo.IncTokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID:  propID.String(),
		Amount:      bridgeIssuanceTxInfo.IssuingAmount,
		TokenTxType: transaction.CustomTokenInit,
		Receiver:    []*privacy.PaymentInfo{receiver},
		TokenInput:  []*privacy.InputCoin{},
		Mintable:    true,
	}

	resTx := &transaction.TxCustomTokenPrivacy{}
	initErr := resTx.Init(
		transaction.NewTxPrivacyTokenInitParams(producerPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			tokenParams,
			shardView.GetCopiedTransactionStateDB(),
			issuingRes,
			false,
			false,
			shardID, nil,
			featureStateDB))

	if initErr != nil {
		Logger.log.Warn("WARNING: an error occured while initializing response tx: ", initErr)
		return nil, nil
	}
	Logger.log.Infof("[Decentralized bridge token issuance] Create tx ok: %s", resTx.Hash().String())
	return resTx, nil
}
