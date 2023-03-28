package bridgehub

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridgeHub "github.com/incognitochain/incognito-chain/metadata/bridgehub"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type stateProducer struct{}

func (sp *stateProducer) registerBridge(
	contentStr string, state *BridgeHubState, sDBs map[int]*statedb.StateDB, shardID byte,
) ([][]string, *BridgeHubState, error) {
	Logger.log.Infof("[BriHub] Beacon producer - Handle register bridge request")

	// decode action
	action := metadataCommon.NewAction()
	meta := &metadataBridgeHub.RegisterBridgeRequest{}
	action.Meta = meta
	err := action.FromString(contentStr)
	if err != nil {
		Logger.log.Errorf("[BriHub] Beacon producer - Can not decode action register bridge from shard: %v - Error: %v", contentStr, err)
		return [][]string{}, state, nil
	}

	// don't need to verify the signature because it was verified in func ValidateSanityData

	// check number of validators
	if uint(len(meta.ValidatorPubKeys)) < state.params.MinNumberValidators() {
		inst, _ := buildBridgeHubRegisterBridgeInst(*meta, shardID, action.TxReqID, common.RejectedStatusStr, InvalidNumberValidatorError)
		return [][]string{inst}, state, nil
	}

	// check bridgeID existed or not
	bridgeID := meta.BridgePoolPubKey

	if state.bridgeInfos[bridgeID] != nil {
		inst, _ := buildBridgeHubRegisterBridgeInst(*meta, shardID, action.TxReqID, common.RejectedStatusStr, BridgeIDExistedError)
		return [][]string{inst}, state, nil
	}

	// TODO: 0xkraken: if chainID is BTC, init pToken with pBTC ID from portal v4

	// update state
	clonedState := state.Clone()
	clonedState.bridgeInfos[bridgeID] = &BridgeInfo{
		Info:        statedb.NewBridgeInfoStateWithValue(meta.ValidatorPubKeys, meta.BridgePoolPubKey, []string{}, ""),
		NetworkInfo: newBridgeHubNetworkInfo(meta.VaultAddress),
	}

	// build accepted instruction
	inst, _ := buildBridgeHubRegisterBridgeInst(*meta, shardID, action.TxReqID, common.AcceptedStatusStr, 0)
	return [][]string{inst}, clonedState, nil
}

func (sp *stateProducer) shield(
	contentStr string,
	state *BridgeHubState,
	ac *metadata.AccumulatedValues,
	stateDBs map[int]*statedb.StateDB,
	isTxHashIssued func(stateDB *statedb.StateDB, uniqueBtcTx []byte) (bool, error),
) ([][]string, *BridgeHubState, *metadata.AccumulatedValues, error) {
	Logger.log.Info("[Bridge hub] Starting...")

	issuingBTCHubReqAction, err := metadataBridgeHub.ParseBTCIssuingInstContent(contentStr)
	if err != nil {
		return [][]string{}, state, ac, err
	}

	var receivingShardID byte
	otaReceiver := issuingBTCHubReqAction.Meta.Receiver
	pkBytes := otaReceiver.PublicKey.ToBytesS()
	shardID := common.GetShardIDFromLastByte(pkBytes[len(pkBytes)-1])
	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.ShieldingBTCRequestMeta,
		common.RejectedStatusStr,
		shardID,
		issuingBTCHubReqAction.TxReqID.String(),
	)
	rejectedInst := inst.StringSlice()
	if err != nil {
		Logger.log.Warn("[Bridge hub] WARNING: an issue occurred while parsing issuing action content: ", err)
		return [][]string{rejectedInst}, state, ac, err
	}

	receivingShardID = otaReceiver.GetShardID()

	md := issuingBTCHubReqAction.Meta
	Logger.log.Infof("[Bridge hub] Processing for tx: %s, tokenid: %s", issuingBTCHubReqAction.TxReqID.String(), md.IncTokenID.String())
	// todo: validate the request
	//ok, err := tss.VerifyTSSSig("", "", issuingBTCHubReqAction.Meta.Signature)
	//if err != nil || !ok {
	//	Logger.log.Warn("[Bridge hub] WARNING: an issue occurred verify signature: ", err, ok)
	//	if err != nil {
	//		err = errors.New("invalid signature")
	//	}
	//	return [][]string{rejectedInst}, state, ac, err
	//}
	// todo: verify validators has enough collateral to mint more btc

	// check tx issued
	isIssued, err := isTxHashIssued(stateDBs[common.BeaconChainID], issuingBTCHubReqAction.Meta.BTCTxID.Bytes())
	if err != nil || !isIssued {
		Logger.log.Warn("WARNING: an issue occured while checking the bridge tx hash is issued or not: %v ", err)
		return [][]string{rejectedInst}, state, ac, err
	}

	// todo: verify token id must be btc token
	// todo: add logic update the collateral and amount shielded

	// update state
	clonedState := state.Clone()
	if clonedState.bridgeInfos[md.BridgePoolPubKey] == nil || clonedState.bridgeInfos[md.BridgePoolPubKey].NetworkInfo[md.ExtChainID] == nil {
		return [][]string{rejectedInst}, state, ac, errors.New("[Bridge Hub] The bridge pool pub key, external chain id is non-existing")
	}
	clonedState.bridgeInfos[md.BridgePoolPubKey].NetworkInfo[md.ExtChainID].PTokens[md.BTCTxID] += md.Amount

	issuingAcceptedInst := metadataBridgeHub.ShieldingBTCAcceptedInst{
		ShardID:          receivingShardID,
		IssuingAmount:    issuingBTCHubReqAction.Meta.Amount,
		Receiver:         issuingBTCHubReqAction.Meta.Receiver,
		IncTokenID:       md.IncTokenID,
		TxReqID:          issuingBTCHubReqAction.TxReqID,
		UniqTx:           issuingBTCHubReqAction.Meta.BTCTxID.Bytes(),
		ExtChainID:       md.ExtChainID,
		BridgePoolPubKey: md.BridgePoolPubKey,
	}
	issuingAcceptedInstBytes, err := json.Marshal(issuingAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while marshaling issuingBridgeAccepted instruction: ", err)
		return [][]string{rejectedInst}, state, ac, err
	}
	inst.Status = common.AcceptedStatusStr
	inst.Content = base64.StdEncoding.EncodeToString(issuingAcceptedInstBytes)
	Logger.log.Info("[Decentralized bridge token issuance] Process finished without error...")
	return [][]string{inst.StringSlice()}, state, ac, err
}

func (sp *stateProducer) stake(
	contentStr string,
	state *BridgeHubState,
	stateDBs map[int]*statedb.StateDB,
	shardID byte,
) ([][]string, *BridgeHubState, error) {
	Logger.log.Info("[Bridge hub] Starting...")

	// decode action
	action := metadataCommon.NewAction()
	meta := &metadataBridgeHub.StakePRVRequest{}
	action.Meta = meta
	err := action.FromString(contentStr)
	if err != nil {
		Logger.log.Warn("[Bridge hub] decode request stake got error: ", err)
		return [][]string{}, state, err
	}
	// todo: cryptolover add more validation
	if state.bridgeInfos[meta.BridgePoolPubKey] == nil {
		inst, _ := buildBridgeHubStakeInst(*meta, shardID, action.TxReqID, common.RejectedStatusStr, BridgeIDNotExistedError)
		return [][]string{inst}, state, nil
	}

	// check bridgeID existed or not
	isBridgeKeyExist := false
	for _, k := range state.bridgeInfos[meta.BridgePoolPubKey].Info.BriValidators() {
		if k == meta.BridgePubKey {
			isBridgeKeyExist = true
			break
		}
	}
	if !isBridgeKeyExist {
		inst, _ := buildBridgeHubStakeInst(*meta, shardID, action.TxReqID, common.RejectedStatusStr, BridgeKeyNotMatchInValidatorList)
		return [][]string{inst}, state, nil
	}

	// update state
	clonedState := state.Clone()
	_, found := clonedState.stakingInfos[meta.BridgePubKey]
	if !found {
		clonedState.stakingInfos[meta.BridgePubKey] = &statedb.BridgeStakingInfoState{}
	}
	clonedState.stakingInfos[meta.BridgePubKey].SetStakingAmount(clonedState.stakingInfos[meta.BridgePubKey].StakingAmount() + meta.StakeAmount)
	// build accepted instruction
	inst, _ := buildBridgeHubStakeInst(*meta, shardID, action.TxReqID, common.AcceptedStatusStr, 0)
	return [][]string{inst}, clonedState, nil
}

// create burn token for bridge hub instruction
func (sp *stateProducer) unshield(
	contentStr string,
	state *BridgeHubState,
	height uint64,
	stateDBs *statedb.StateDB,
) ([][]string, *BridgeHubState, error) {
	// decode action
	action := metadataCommon.NewAction()
	meta := &metadataBridgeHub.BridgeHubUnshieldRequest{}
	action.Meta = meta
	err := action.FromString(contentStr)
	if err != nil {
		Logger.log.Warn("[Bridge hub] decode request stake got error: ", err)
		return [][]string{}, state, err
	}

	txID := action.TxReqID // to prevent double-release token

	// Convert amount to big.Int to get bytes later
	amount := big.NewInt(0).SetUint64(meta.BurningAmount)
	// Convert height to big.Int to get bytes later
	h := big.NewInt(0).SetUint64(height)

	results := []string{
		strconv.Itoa(metadataCommon.BridgeHubUnshieldConfirm),
		base58.Base58Check{}.Encode(meta.TokenID[:], 0x00),
		meta.RemoteAddress,
		base58.Base58Check{}.Encode(amount.Bytes(), 0x00),
		txID.String(),
		meta.ExtChainID,
	}
	// todo: update bridge hub state

	results = append(results, base58.Base58Check{}.Encode(h.Bytes(), 0x00))
	return [][]string{results}, state, nil
}
