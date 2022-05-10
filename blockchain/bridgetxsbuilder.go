package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
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

	privacyTokenExisted, err := blockchain.PrivacyTokenIDExistedInNetwork(beaconBestState, issuingTokenID)
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

func (blockchain *BlockChain) buildInstructionsForIssuingBridgeReq(
	stateDBs map[int]*statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	ac *metadata.AccumulatedValues,
	listTxUsed [][]byte,
	contractAddress string,
	prefix string,
	isTxHashIssued func(stateDB *statedb.StateDB, uniqueEthTx []byte) (bool, error),
	isPRV bool,
) ([][]string, []byte, error) {
	Logger.log.Info("[Decentralized bridge token issuance] Starting...")
	issuingEVMBridgeReqAction, err := metadataBridge.ParseEVMIssuingInstContent(contentStr)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while parsing issuing action content: ", err)
		return nil, nil, nil
	}
	md := issuingEVMBridgeReqAction.Meta
	Logger.log.Infof("[Decentralized bridge token issuance] Processing for tx: %s, tokenid: %s", issuingEVMBridgeReqAction.TxReqID.String(), md.IncTokenID.String())

	inst := metadataCommon.NewInstructionWithValue(
		metaType,
		common.RejectedStatusStr,
		shardID,
		issuingEVMBridgeReqAction.TxReqID.String(),
	)
	rejectedInst := inst.StringSlice()

	amt, receivingShardID, addressStr, token, uniqTx, err := metadataBridge.ExtractIssueEVMData(
		stateDBs[common.BeaconChainID], shardID, listTxUsed,
		contractAddress, prefix, isTxHashIssued,
		issuingEVMBridgeReqAction.EVMReceipt,
		issuingEVMBridgeReqAction.Meta.BlockHash,
		issuingEVMBridgeReqAction.Meta.TxIndex,
	)
	if err != nil {
		Logger.log.Warnf(err.Error())
		return [][]string{rejectedInst}, nil, nil
	}

	amount := uint64(0)
	if bytes.Equal(append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), token) {
		// convert amt from wei (10^18) to nano eth (10^9)
		amount = big.NewInt(0).Div(amt, big.NewInt(1000000000)).Uint64()
	} else { // ERC20 / BEP20
		amount = amt.Uint64()
	}
	if !isPRV {
		err := metadataBridge.VerifyTokenPair(stateDBs, ac, md.IncTokenID, token)
		if err != nil {
			Logger.log.Warnf(err.Error())
			return [][]string{rejectedInst}, nil, nil
		}
	}

	issuingAcceptedInst := metadataBridge.IssuingEVMAcceptedInst{
		ShardID:         receivingShardID,
		IssuingAmount:   amount,
		ReceiverAddrStr: addressStr,
		IncTokenID:      md.IncTokenID,
		TxReqID:         issuingEVMBridgeReqAction.TxReqID,
		UniqTx:          uniqTx,
		ExternalTokenID: token,
	}
	issuingAcceptedInstBytes, err := json.Marshal(issuingAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while marshaling issuingBridgeAccepted instruction: ", err)
		return [][]string{rejectedInst}, nil, nil
	}
	ac.DBridgeTokenPair[md.IncTokenID.String()] = token
	inst.Status = common.AcceptedStatusStr
	inst.Content = base64.StdEncoding.EncodeToString(issuingAcceptedInstBytes)
	Logger.log.Info("[Decentralized bridge token issuance] Process finished without error...")
	return [][]string{inst.StringSlice()}, uniqTx, nil
}

func (blockchain *BlockChain) buildInstructionsForIssuingTerraBridgeReq(
	stateDBs map[int]*statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	ac *metadata.AccumulatedValues,
	listTxUsed [][]byte,
	contractAddress string,
	prefix string,
	isTxHashIssued func(stateDB *statedb.StateDB, uniqueEthTx []byte) (bool, error),
) ([][]string, []byte, error) {
	Logger.log.Info("[Decentralized bridge wasm token issuance] Starting...")
	issuingWasmBridgeReqAction, err := metadataBridge.ParseWasmIssuingInstContent(contentStr)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while parsing issuing action content: ", err)
		return nil, nil, nil
	}
	md := issuingWasmBridgeReqAction.Meta
	Logger.log.Infof("[Decentralized bridge wasm token issuance] Processing for tx: %s, tokenid: %s", issuingWasmBridgeReqAction.TxReqID.String(), md.IncTokenID.String())

	inst := metadataCommon.NewInstructionWithValue(
		metaType,
		common.RejectedStatusStr,
		shardID,
		issuingWasmBridgeReqAction.TxReqID.String(),
	)
	rejectedInst := inst.StringSlice()
	uniqTx := issuingWasmBridgeReqAction.TxReqID.Bytes()
	isUsedInBlock := metadataBridge.IsBridgeTxHashUsedInBlock(uniqTx, listTxUsed)
	if isUsedInBlock {
		Logger.log.Warn("WARNING: already issued for the hash in current block: ", uniqTx)
		return [][]string{rejectedInst}, nil, nil
	}
	isIssued, err := isTxHashIssued(stateDBs[common.BeaconChainID], uniqTx)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while checking the bridge tx hash is issued or not: ", err)
		return [][]string{rejectedInst}, nil, nil
	}
	if isIssued {
		Logger.log.Warn("WARNING: already issued for the hash in previous blocks: ", uniqTx)
		return [][]string{rejectedInst}, nil, nil
	}

	if contractAddress != issuingWasmBridgeReqAction.ContractId {
		Logger.log.Warn("WARNING: send to wrong vault smart contract ", issuingWasmBridgeReqAction.ContractId)
		return [][]string{rejectedInst}, nil, nil
	}

	receivingShardID, err := getShardIDFromPaymentAddress(issuingWasmBridgeReqAction.IncognitoAddrStr)
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while getting shard id from payment address: ", err)
		return [][]string{rejectedInst}, nil, nil
	}

	tokenBytes := append([]byte(prefix), issuingWasmBridgeReqAction.Token...)
	err = metadataBridge.VerifyTokenPair(stateDBs, ac, md.IncTokenID, tokenBytes)
	if err != nil {
		Logger.log.Warnf(err.Error())
		return [][]string{rejectedInst}, nil, nil
	}

	issuingAcceptedInst := metadataBridge.IssuingEVMAcceptedInst{
		ShardID:         receivingShardID,
		IssuingAmount:   issuingWasmBridgeReqAction.Amount,
		ReceiverAddrStr: issuingWasmBridgeReqAction.IncognitoAddrStr,
		IncTokenID:      md.IncTokenID,
		TxReqID:         issuingWasmBridgeReqAction.TxReqID,
		UniqTx:          uniqTx,
		ExternalTokenID: tokenBytes,
	}
	issuingAcceptedInstBytes, err := json.Marshal(issuingAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while marshaling issuingBridgeAccepted instruction: ", err)
		return [][]string{rejectedInst}, nil, nil
	}
	ac.DBridgeTokenPair[md.IncTokenID.String()] = tokenBytes
	inst.Status = common.AcceptedStatusStr
	inst.Content = base64.StdEncoding.EncodeToString(issuingAcceptedInstBytes)
	Logger.log.Info("[Decentralized bridge token issuance] Process finished without error...")
	return [][]string{inst.StringSlice()}, uniqTx, nil
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
		Logger.log.Warnf("WARNING: an error occurs while decode content string of accepted issuance instruction: ", err)
		return nil, nil
	}
	var issuingAcceptedInst metadata.IssuingAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingAcceptedInst)
	if err != nil {
		Logger.log.Warnf("WARNING: an error occurs while unmarshal accepted issuance instruction: ", err)
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

	tokenID := issuingAcceptedInst.IncTokenID
	if tokenID == common.PRVCoinID {
		Logger.log.Errorf("cannot issue prv in bridge")
		return nil, errors.New("cannot issue prv in bridge")
	}
	txParam := transaction.TxSalaryOutputParams{Amount: receiver.Amount, ReceiverAddress: &receiver.PaymentAddress, TokenID: &tokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			issuingRes.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return issuingRes
	}
	return txParam.BuildTxSalary(producerPrivateKey, shardView.GetCopiedTransactionStateDB(), makeMD)
}

func (blockGenerator *BlockGenerator) buildBridgeIssuanceTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	shardView *ShardBestState,
	featureStateDB *statedb.StateDB,
	metatype int,
	isPeggedPRV bool,
) (metadata.Transaction, error) {
	Logger.log.Info("[Decentralized bridge token issuance] Starting...")
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while decoding content string of EVM accepted issuance instruction: ", err)
		return nil, nil
	}
	var issuingEVMAcceptedInst metadataBridge.IssuingEVMAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingEVMAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while unmarshaling EVM accepted issuance instruction: ", err)
		return nil, nil
	}

	if shardID != issuingEVMAcceptedInst.ShardID {
		Logger.log.Warnf("Ignore due to shardid difference, current shardid %d, receiver's shardid %d", shardID, issuingEVMAcceptedInst.ShardID)
		return nil, nil
	}
	key, err := wallet.Base58CheckDeserialize(issuingEVMAcceptedInst.ReceiverAddrStr)
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while deserializing receiver address string: ", err)
		return nil, nil
	}
	receiver := &privacy.PaymentInfo{
		Amount:         issuingEVMAcceptedInst.IssuingAmount,
		PaymentAddress: key.KeySet.PaymentAddress,
	}

	issuingEVMRes := metadataBridge.NewIssuingEVMResponse(
		issuingEVMAcceptedInst.TxReqID,
		issuingEVMAcceptedInst.UniqTx,
		issuingEVMAcceptedInst.ExternalTokenID,
		metatype,
	)
	tokenID := issuingEVMAcceptedInst.IncTokenID
	if !isPeggedPRV && tokenID == common.PRVCoinID {
		Logger.log.Errorf("cannot issue prv in bridge")
		return nil, errors.New("cannot issue prv in bridge")
	}

	txParam := transaction.TxSalaryOutputParams{Amount: receiver.Amount, ReceiverAddress: &receiver.PaymentAddress, TokenID: &tokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			issuingEVMRes.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return issuingEVMRes
	}

	return txParam.BuildTxSalary(producerPrivateKey, shardView.GetCopiedTransactionStateDB(), makeMD)
}

func (blockchain *BlockChain) getStateDBsForVerifyTokenID(curView *BeaconBestState) (map[int]*statedb.StateDB, error) {
	res := make(map[int]*statedb.StateDB)
	res[common.BeaconChainID] = curView.featureStateDB

	for shardID, shardHash := range curView.BestShardHash {
		db := blockchain.GetShardChainDatabase(shardID)
		shardRootHash, err := GetShardRootsHashByBlockHash(db, shardID, shardHash)
		if err != nil {
			return res, err
		}
		stateDB, err := statedb.NewWithPrefixTrie(shardRootHash.TransactionStateDBRootHash,
			statedb.NewDatabaseAccessWarper(db))
		if err != nil {
			return res, err
		}
		res[int(shardID)] = stateDB
	}
	return res, nil
}
