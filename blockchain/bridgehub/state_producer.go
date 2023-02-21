package bridgehub

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridgeHub "github.com/incognitochain/incognito-chain/metadata/bridgehub"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/metadata/tss"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/schnorr"
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
		inst, _ := buildBridgeHubRegisterBridgeInst(*meta, shardID, action.TxReqID, "", common.RejectedStatusStr, InvalidNumberValidatorError)
		return [][]string{inst}, state, nil
	}

	// check all ValidatorPubKeys staked or not
	for _, validatorPubKeyStr := range meta.ValidatorPubKeys {
		if state.stakingInfos[validatorPubKeyStr] < state.params.MinStakedAmountValidator() {
			inst, _ := buildBridgeHubRegisterBridgeInst(*meta, shardID, action.TxReqID, "", common.RejectedStatusStr, InvalidStakedAmountValidatorError)
			return [][]string{inst}, state, nil
		}
	}

	// check bridgeID existed or not
	bridgeIDBytes := append([]byte(meta.ExtChainID), []byte(meta.BridgePoolPubKey)...)
	bridgeID := common.HashH(bridgeIDBytes).String()

	if state.bridgeInfos[bridgeID] != nil {
		inst, _ := buildBridgeHubRegisterBridgeInst(*meta, shardID, action.TxReqID, bridgeID, common.RejectedStatusStr, BridgeIDExistedError)
		return [][]string{inst}, state, nil

	}

	// TODO: 0xkraken: if chainID is BTC, init pToken with pBTC ID from portal v4

	// update state
	clonedState := state.Clone()
	clonedState.bridgeInfos[bridgeID] = &BridgeInfo{
		Info:          statedb.NewBridgeInfoStateWithValue(bridgeID, meta.ExtChainID, meta.ValidatorPubKeys, meta.BridgePoolPubKey, []string{}, ""),
		PTokenAmounts: map[string]*statedb.BridgeHubPTokenState{},
	}

	// build accepted instruction
	inst, _ := buildBridgeHubRegisterBridgeInst(*meta, shardID, action.TxReqID, bridgeID, common.AcceptedStatusStr, 0)
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

	var receiver string
	var receivingShardID byte
	var depositKeyBytes []byte
	otaReceiver := new(privacy.OTAReceiver)
	_ = otaReceiver.FromString(issuingBTCHubReqAction.Meta.Receiver) // error has been handle at shard side
	otaReceiverBytes, _ := otaReceiver.Bytes()
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

	depositPubKey, err := new(operation.Point).FromBytesS(depositKeyBytes)
	if err != nil {
		Logger.log.Warn("[Bridge hub] WARNING: invalid OTDepositPubKey %v", issuingBTCHubReqAction.Meta.Receiver)
		return [][]string{rejectedInst}, state, ac, err
	}
	sigPubKey := new(privacy.SchnorrPublicKey)
	sigPubKey.Set(depositPubKey)

	tmpSig := new(schnorr.SchnSignature)
	_ = tmpSig.SetBytes(issuingBTCHubReqAction.Meta.Signature) // error has been handle at shard side

	if isValid := sigPubKey.Verify(tmpSig, common.HashB(otaReceiverBytes)); !isValid {
		Logger.log.Warn("[Bridge hub] invalid signature", issuingBTCHubReqAction.Meta.Signature)
		return [][]string{rejectedInst}, state, ac, err
	}

	receiver = issuingBTCHubReqAction.Meta.Receiver
	receivingShardID = otaReceiver.GetShardID()

	md := issuingBTCHubReqAction.Meta
	Logger.log.Infof("[Bridge hub] Processing for tx: %s, tokenid: %s", issuingBTCHubReqAction.TxReqID.String(), md.IncTokenID.String())
	// todo: validate the request
	ok, err := tss.VerifyTSSSig("", "", issuingBTCHubReqAction.Meta.TSS)
	if err != nil || !ok {
		Logger.log.Warn("[Bridge hub] WARNING: an issue occurred verify signature: ", err, ok)
		if err != nil {
			err = errors.New("invalid signature")
		}
		return [][]string{rejectedInst}, state, ac, err
	}

	// check tx issued
	isIssued, err := isTxHashIssued(stateDBs[common.BeaconChainID], issuingBTCHubReqAction.Meta.BTCTxID.Bytes())
	if err != nil || !isIssued {
		Logger.log.Warn("WARNING: an issue occured while checking the bridge tx hash is issued or not: %v ", err)
		return [][]string{rejectedInst}, state, ac, err
	}

	// todo: verify token id must be btc token
	// todo: add logic update the collateral and amount shielded

	issuingAcceptedInst := metadataBridgeHub.ShieldingBTCAcceptedInst{
		ShardID:         receivingShardID,
		IssuingAmount:   issuingBTCHubReqAction.Meta.Amount,
		Receiver:        receiver,
		IncTokenID:      md.IncTokenID,
		TxReqID:         issuingBTCHubReqAction.TxReqID,
		UniqTx:          issuingBTCHubReqAction.Meta.BTCTxID.Bytes(),
		ExternalTokenID: issuingBTCHubReqAction.Meta.IncTokenID.Bytes(),
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
	ac *metadata.AccumulatedValues,
	stateDBs map[int]*statedb.StateDB,
	isTxHashIssued func(stateDB *statedb.StateDB, uniqueBtcTx []byte) (bool, error),
) ([][]string, *BridgeHubState, error) {
	Logger.log.Info("[Bridge hub] Starting...")

	var stakeReqAction metadataBridgeHub.StakeReqAction
	err := metadataBridgeHub.DecodeContent(contentStr, &stakeReqAction)
	if err != nil {
		// todo: cryptolover add logic here
	}

	return [][]string{}, state, err
}
