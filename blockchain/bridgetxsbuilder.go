package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eteu-technologies/near-api-go/pkg/types/hash"
	"github.com/incognitochain/incognito-chain/metadata/bridgehub"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
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
) ([][]string, [][]byte, error) {
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

	evmProof := metadataBridge.EVMProof{
		BlockHash: md.BlockHash,
		TxIndex:   md.TxIndex,
		Proof:     md.ProofStrs,
	}

	amt, addressStr, token, err := metadataBridge.ExtractIssueEVMDataFromReceipt(
		issuingEVMBridgeReqAction.EVMReceipt,
		contractAddress,
		prefix,
		"",
	)
	var tmpData *metadataBridge.DepositEventData
	if err != nil {
		Logger.log.Warnf(err.Error())
	} else {
		receivingShardID, err := metadataBridge.GetShardIDFromPaymentAddressStr(addressStr)
		if err != nil {
			Logger.log.Warnf(err.Error())
		} else {
			tmpData = &metadataBridge.DepositEventData{
				Amount:          amt,
				ReceiverStr:     addressStr,
				ExternalTokenID: token,
				IncTxID:         append(evmProof.BlockHash[:], []byte(strconv.Itoa(int(evmProof.TxIndex)))...),
				ShardID:         receivingShardID,
				IsOneTime:       false,
			}
		}
	}

	redepositDataLst, err := metadataBridge.ExtractRedepositEVMDataFromReceipt(issuingEVMBridgeReqAction.EVMReceipt, contractAddress, prefix)
	if err != nil {
		Logger.log.Warnf("[BridgeAgg] Extract Redeposit EVM events failed - %v", err)
	}
	if tmpData != nil {
		redepositDataLst = append(redepositDataLst, *tmpData)
	}

	var result [][]string
	var uniqTxs [][]byte
	for _, d := range redepositDataLst {
		// check double use shielding proof in current block and previous blocks
		isValid, uniqTx, err := bridgeagg.ValidateDoubleShieldProof(d.IncTxID, listTxUsed, isTxHashIssued, stateDBs[common.BeaconChainID])
		if err != nil || !isValid {
			Logger.log.Errorf("[BridgeAgg] Can not validate double shielding proof - Error %v", err)
			continue
		}
		listTxUsed = append(listTxUsed, uniqTx)

		if !isPRV {
			err := metadataBridge.VerifyTokenPair(stateDBs, ac, md.IncTokenID, d.ExternalTokenID)
			if err != nil {
				Logger.log.Warnf(err.Error())
				continue
			}
		}
		var amount uint64
		if bytes.Equal(append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), d.ExternalTokenID) {
			// convert amt from wei (10^18) to nano eth (10^9)
			amount = big.NewInt(0).Div(d.Amount, big.NewInt(1000000000)).Uint64()
		} else { // ERC20 / BEP20
			amount = d.Amount.Uint64()
		}

		issuingAcceptedInst := metadataBridge.IssuingEVMAcceptedInst{
			ShardID:         d.ShardID,
			IssuingAmount:   amount,
			ReceiverAddrStr: d.ReceiverStr,
			IncTokenID:      md.IncTokenID,
			TxReqID:         issuingEVMBridgeReqAction.TxReqID,
			UniqTx:          uniqTx,
			ExternalTokenID: d.ExternalTokenID,
		}
		issuingAcceptedInstBytes, err := json.Marshal(issuingAcceptedInst)
		if err != nil {
			Logger.log.Warn("WARNING: an error occurred while marshaling issuingBridgeAccepted instruction: ", err)
			continue
		}
		ac.DBridgeTokenPair[md.IncTokenID.String()] = d.ExternalTokenID
		var currentInst []string
		if d.IsOneTime {
			var recv privacy.OTAReceiver
			recv.FromString(d.ReceiverStr)
			networkID, _ := bridgeagg.GetNetworkIDByPrefix(prefix)
			c := metadataBridge.AcceptedReshieldRequest{
				UnifiedTokenID: nil,
				Receiver:       recv,
				TxReqID:        issuingEVMBridgeReqAction.TxReqID,
				ReshieldData: metadataBridge.AcceptedShieldRequestData{
					ShieldAmount:    amount,
					Reward:          0,
					UniqTx:          d.IncTxID,
					ExternalTokenID: d.ExternalTokenID,
					NetworkID:       networkID,
					IncTokenID:      md.IncTokenID,
				},
			}
			contentBytes, _ := json.Marshal(c)
			currentInst = metadataCommon.NewInstructionWithValue(
				metadataCommon.IssuingReshieldResponseMeta,
				common.AcceptedStatusStr,
				d.ShardID,
				base64.StdEncoding.EncodeToString(contentBytes),
			).StringSlice()
		} else {
			currentInst = metadataCommon.NewInstructionWithValue(
				metaType,
				common.AcceptedStatusStr,
				shardID,
				base64.StdEncoding.EncodeToString(issuingAcceptedInstBytes),
			).StringSlice()
		}
		result = append(result, currentInst)
		uniqTxs = append(uniqTxs, uniqTx)
		Logger.log.Info("[Decentralized bridge token issuance] Process finished without error...")
	}

	if len(result) == 0 {
		return [][]string{rejectedInst}, nil, nil
	}

	return result, uniqTxs, nil
}

func (blockchain *BlockChain) buildInstructionsForIssuingWasmBridgeReq(
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
	Logger.log.Info("[Decentralized Wasm bridge token issuance] Starting...")
	issuingWasmBridgeReqAction, err := metadataBridge.ParseWasmIssuingInstContent(contentStr)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occurred while parsing issuing action content: ", err)
		return nil, nil, nil
	}
	md := issuingWasmBridgeReqAction.Meta
	Logger.log.Infof("[Decentralized Wasm bridge token issuance] Processing for tx: %s, tokenid: %s", issuingWasmBridgeReqAction.TxReqID.String(), md.IncTokenID.String())

	inst := metadataCommon.NewInstructionWithValue(
		metaType,
		common.RejectedStatusStr,
		shardID,
		issuingWasmBridgeReqAction.TxReqID.String(),
	)
	rejectedInst := inst.StringSlice()

	if contractAddress != issuingWasmBridgeReqAction.ContractId {
		Logger.log.Warn("WARNING: contract id not match incognito contract: ", issuingWasmBridgeReqAction.ContractId)
		return [][]string{rejectedInst}, nil, nil
	}

	receivingShardID, err := metadataBridge.VerifyWasmData(
		stateDBs[common.BeaconChainID], listTxUsed, isTxHashIssued,
		md.TxHash,
		issuingWasmBridgeReqAction.IncognitoAddr)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occurred while execute VerifyWasmData: ", err)
		return [][]string{rejectedInst}, nil, nil
	}
	uniqTxCryptoHash, err := hash.NewCryptoHashFromBase58(md.TxHash)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occurred while decode wasm shielding tx hash: ", err)
		return [][]string{rejectedInst}, nil, nil
	}
	uniqTxTemp := [32]byte(uniqTxCryptoHash)
	uniqTx := uniqTxTemp[:]
	token := append([]byte(prefix), []byte(issuingWasmBridgeReqAction.TokenId)...)
	err = metadataBridge.VerifyTokenPair(stateDBs, ac, md.IncTokenID, token)
	if err != nil {
		Logger.log.Warnf(err.Error())
		return [][]string{rejectedInst}, nil, nil
	}

	issuingAcceptedInst := metadataBridge.IssuingEVMAcceptedInst{
		ShardID:         receivingShardID,
		IssuingAmount:   issuingWasmBridgeReqAction.Amount,
		ReceiverAddrStr: issuingWasmBridgeReqAction.IncognitoAddr,
		IncTokenID:      md.IncTokenID,
		TxReqID:         issuingWasmBridgeReqAction.TxReqID,
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

func (blockGenerator *BlockGenerator) buildBridgeHubIssuanceTx(
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
		Logger.log.Warn("WARNING: an error occurred while decoding content string of Bridge Hub accepted issuance instruction: ", err)
		return nil, nil
	}
	var issuingBridgeHubAcceptedInst bridgehub.ShieldingBTCAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingBridgeHubAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while unmarshaling Bridge Hub accepted issuance instruction: ", err)
		return nil, nil
	}

	if shardID != issuingBridgeHubAcceptedInst.ShardID {
		Logger.log.Warnf("Ignore due to shardid difference, current shardid %d, receiver's shardid %d", shardID, issuingBridgeHubAcceptedInst.ShardID)
		return nil, nil
	}

	issuingEVMRes := bridgehub.NewIssuingBTCResponse(
		issuingBridgeHubAcceptedInst.TxReqID,
		issuingBridgeHubAcceptedInst.UniqTx,
		issuingBridgeHubAcceptedInst.ExternalTokenID,
		metatype,
	)
	tokenID := issuingBridgeHubAcceptedInst.IncTokenID
	if !isPeggedPRV && tokenID == common.PRVCoinID {
		Logger.log.Errorf("cannot issue prv in bridge")
		return nil, errors.New("cannot issue prv in bridge")
	}

	txParam := transaction.TxSalaryOutputParams{
		Amount:  issuingBridgeHubAcceptedInst.IssuingAmount,
		TokenID: &tokenID,
	}

	otaReceiver := new(privacy.OTAReceiver)
	err = otaReceiver.FromString(issuingBridgeHubAcceptedInst.Receiver)
	if err != nil {
		return nil, fmt.Errorf("parseOTA receiver from %v error: %v", issuingBridgeHubAcceptedInst.Receiver, err)
	}
	txParam.TxRandom = &otaReceiver.TxRandom
	txParam.PublicKey = otaReceiver.PublicKey

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
