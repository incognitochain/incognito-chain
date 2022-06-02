package bridgeagg

import (
	"bytes"
	"encoding/json"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type stateProducer struct{}

func (sp *stateProducer) modifyParam(
	contentStr string, state *State, sDBs map[int]*statedb.StateDB, shardID byte,
) ([][]string, *State, error) {
	Logger.log.Infof("Build instruction for modify param...")
	// decode action
	action := metadataCommon.NewAction()
	meta := &metadataBridge.ModifyBridgeAggParamReq{}
	action.Meta = meta
	err := action.FromString(contentStr)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Can not decode action modify param from shard: %v - Error: %v", contentStr, err)
		return [][]string{}, state, nil
	}

	// update state
	state = updateStateForModifyParam(state, meta.PercentFeeWithDec)

	// build accepted modify param instruction
	acceptedContent := metadataBridge.ModifyBridgeAggParamContentInst{
		TxReqID:           action.TxReqID,
		PercentFeeWithDec: meta.PercentFeeWithDec,
	}
	content, _ := json.Marshal(acceptedContent)
	insts := buildAcceptedInst(metadataCommon.BridgeAggModifyParamMeta, shardID, [][]byte{content})

	return insts, state, nil
}

func (sp *stateProducer) convert(
	contentStr string, state *State, sDBs map[int]*statedb.StateDB, shardID byte,
) ([][]string, *State, error) {
	// decode action
	action := metadataCommon.NewAction()
	meta := &metadataBridge.ConvertTokenToUnifiedTokenRequest{}
	action.Meta = meta
	err := action.FromString(contentStr)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Can not decode action convert to punified token from shard: %v - Error: %v", contentStr, err)
		return [][]string{}, state, nil
	}

	// check unifiedTokenID
	clonedVaults, err := state.CloneVaultsByUnifiedTokenID(meta.UnifiedTokenID)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] UnifiedTokenID is not found: %v", meta.UnifiedTokenID)
		rejectedInst := buildRejectedConvertReqInst(
			*meta, shardID, action.TxReqID, NotFoundTokenIDInNetworkError)
		return [][]string{rejectedInst}, state, nil
	}

	// check IncTokenID
	vault, ok := clonedVaults[meta.TokenID]
	if !ok {
		Logger.log.Errorf("[BridgeAgg] IncTokenID is not found: %v", meta.TokenID)
		rejectedInst := buildRejectedConvertReqInst(
			*meta, shardID, action.TxReqID, NotFoundTokenIDInNetworkError)
		return [][]string{rejectedInst}, state, nil
	}

	// convert pToken amount to pUnifiedToken amount
	pUnifiedTokenConvertAmt, err := convertPTokenAmtToPUnifiedTokenAmt(vault.ExtDecimal(), meta.Amount)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Error convert to punified token amount: %v", err)
		rejectedInst := buildRejectedConvertReqInst(
			*meta, shardID, action.TxReqID, InvalidConvertAmountError)
		return [][]string{rejectedInst}, state, nil
	}

	// calculate converting reward (in pDecimal)
	reward, err := CalRewardForRefillVault(vault, pUnifiedTokenConvertAmt)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Error convert to punified token amount: %v", err)
		rejectedInst := buildRejectedConvertReqInst(
			*meta, shardID, action.TxReqID, InvalidConvertAmountError)
		return [][]string{rejectedInst}, state, nil
	}

	// update state
	vault, err = updateVaultForRefill(vault, pUnifiedTokenConvertAmt, reward)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Error convert to punified token amount: %v", err)
		rejectedInst := buildRejectedConvertReqInst(
			*meta, shardID, action.TxReqID, InvalidConvertAmountError)
		return [][]string{rejectedInst}, state, nil
	}
	clonedVaults[meta.TokenID] = vault
	state.unifiedTokenVaults[meta.UnifiedTokenID] = clonedVaults

	// build accepted convert instruction
	acceptedContent := metadataBridge.AcceptedConvertTokenToUnifiedToken{
		UnifiedTokenID:        meta.UnifiedTokenID,
		TokenID:               meta.TokenID,
		Receiver:              meta.Receiver,
		TxReqID:               action.TxReqID,
		ConvertPUnifiedAmount: pUnifiedTokenConvertAmt,
		ConvertPTokenAmount:   meta.Amount,
		Reward:                reward,
	}
	content, _ := json.Marshal(acceptedContent)
	insts := buildAcceptedInst(metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta, shardID, [][]byte{content})

	return insts, state, nil

	// defer func() {
	// 	if err != nil {
	// 		rejectedConvertRequest := metadataBridge.RejectedConvertTokenToUnifiedToken{
	// 			TokenID:  meta.TokenID,
	// 			Amount:   meta.Amount,
	// 			Receiver: meta.Receiver,
	// 		}
	// 		Logger.log.Warnf("Convert token with tx %s err %v", action.TxReqID.String(), err)
	// 		content, err := json.Marshal(rejectedConvertRequest)
	// 		if err != nil {
	// 			errorType = OtherError
	// 			return
	// 		}
	// 		contents = append(contents, content)
	// 	}
	// 	resInst, err = buildInstruction(
	// 		metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta,
	// 		errorType, contents, action.TxReqID, shardID, err,
	// 	)
	// 	if err != nil {
	// 		Logger.log.Warnf("Cannot buildInstruction with tx %s err %v", action.TxReqID.String(), err)
	// 	}
	// 	err = nil
	// }()

	// if _, found := resState.unifiedTokenVaults[meta.UnifiedTokenID]; !found {
	// 	errorType = NotFoundTokenIDInNetworkError
	// 	err = errors.New("Cannot find unifiedTokenID")
	// 	return
	// }
	// if vault, found := resState.unifiedTokenVaults[meta.UnifiedTokenID][meta.TokenID]; !found {
	// 	errorType = NotFoundTokenIDInNetworkError
	// 	err = fmt.Errorf("Cannot find tokenID %s", meta.TokenID.String())
	// 	return
	// } else {
	// 	v := vault.Clone()
	// 	v, mintAmount, e := convert(v, meta.Amount)
	// 	if err != nil {
	// 		Logger.log.Warnf("Invalid convert amount error: %v tx %s", err, action.TxReqID.String())
	// 		errorType = InvalidConvertAmountError
	// 		err = e
	// 		return
	// 	}
	// 	acceptedContent := metadataBridge.AcceptedConvertTokenToUnifiedToken{
	// 		ConvertTokenToUnifiedTokenRequest: *meta,
	// 		TxReqID:                           action.TxReqID,
	// 		ConvertPUnifiedAmount:             mintAmount,
	// 	}
	// 	var content []byte
	// 	content, err = json.Marshal(acceptedContent)
	// 	if err != nil {
	// 		errorType = OtherError
	// 		return
	// 	}
	// 	resState.unifiedTokenVaults[meta.UnifiedTokenID][meta.TokenID] = v
	// 	contents = append(contents, content)
	// }
	// return
}

func (sp *stateProducer) shield(
	contentStr string,
	state *State,
	ac *metadata.AccumulatedValues,
	shardID byte,
	stateDBs map[int]*statedb.StateDB,
) ([][]string, *State, *metadata.AccumulatedValues, error) {
	// decode action from shard
	action := metadataCommon.NewAction()
	meta := &metadataBridge.ShieldRequest{}
	action.Meta = meta
	err := action.FromString(contentStr)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Can not decode action shield from shard: %v - Error: %v", contentStr, err)
		return [][]string{}, state, ac, nil
	}

	// check unifiedTokenID
	clonedVaults, err := state.CloneVaultsByUnifiedTokenID(meta.UnifiedTokenID)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] UnifiedTokenID is not found: %v", meta.UnifiedTokenID)
		rejectedInst := buildRejectedInst(
			metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, NotFoundTokenIDInNetworkError, []byte{})
		return [][]string{rejectedInst}, state, ac, nil
	}

	clonedAC := ac.Clone()
	acceptedShieldData := []metadataBridge.AcceptedShieldRequestData{}
	var incAddrStr string
	for i, shieldData := range meta.Data {
		// check incTokenID
		vault, ok := clonedVaults[shieldData.IncTokenID]
		if !ok || vault == nil {
			Logger.log.Errorf("[BridgeAgg] Vault IncTokenID is not found: %v", shieldData.IncTokenID)
			rejectedInst := buildRejectedInst(
				metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, NotFoundTokenIDInNetworkError, []byte{})
			return [][]string{rejectedInst}, state, ac, nil
		}

		// check networkID
		if vault.NetworkID() != shieldData.NetworkID {
			Logger.log.Errorf("[BridgeAgg] Network ID is not matched: %v", shieldData.NetworkID)
			rejectedInst := buildRejectedInst(
				metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, NotFoundNetworkIDError, []byte{})
			return [][]string{rejectedInst}, state, ac, nil
		}

		// validate shielding proof
		networkType, _ := metadataBridge.GetNetworkTypeByNetworkID(shieldData.NetworkID)
		switch networkType {
		case common.EVMNetworkType:
			// unmarshal proof and receipt
			proofData, txReceipt, err := UnmarshalEVMShieldProof(shieldData.Proof, action.ExtraData[i])
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not unmarshal shielding proof - Error %v", err)
				rejectedInst := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
				return [][]string{rejectedInst}, state, ac, nil
			}

			evmInfo, err := metadataBridge.GetEVMInfoByNetworkID(shieldData.NetworkID, clonedAC)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not get evm info by network id - Error %v", err)
				rejectedInst := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
				return [][]string{rejectedInst}, state, ac, nil
			}

			// check double use shielding proof in current block and previous blocks
			isValid, uniqTx, err := ValidateDoubleShieldProof(proofData, evmInfo.ListTxUsedInBlock, evmInfo.IsTxHashIssued, stateDBs[common.BeaconChainID])
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not validate double shielding proof - Error %v", err)
				rejectedInst := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
				return [][]string{rejectedInst}, state, ac, nil
			}
			if !isValid {
				Logger.log.Errorf("[BridgeAgg] Can not validate double shielding proof - Error %v", err)
				rejectedInst := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, ShieldProofIsSubmittedError, []byte{})
				return [][]string{rejectedInst}, state, ac, nil
			}

			// extract info from Receipt
			extAmount, incAddr, extTokenID, err := metadataBridge.ExtractIssueEVMDataFromReceipt(
				txReceipt, evmInfo.ContractAddress, evmInfo.Prefix, incAddrStr,
			)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not extract data from Receipt - Error %v", err)
				rejectedInst := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
				return [][]string{rejectedInst}, state, ac, nil
			}
			incAddrStr = incAddr

			// verify token pair pTokenID - extTokenID
			err = metadataBridge.VerifyTokenPair(stateDBs, ac, shieldData.IncTokenID, extTokenID)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Invalid token pair pTokenID and extTokenID - Error %v", err)
				rejectedInst := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, FailToVerifyTokenPairError, []byte{})
				return [][]string{rejectedInst}, state, ac, nil
			}

			// calculate shielding amount (in pDecimal)
			extDecimal := vault.ExtDecimal()
			if !bytes.Equal(append([]byte(evmInfo.Prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), extTokenID) {
				if extDecimal > config.Param().BridgeAggParam.BaseDecimal {
					extDecimal = config.Param().BridgeAggParam.BaseDecimal
				}
			}
			incAmount, err := ConvertAmountByDecimal(extAmount, extDecimal, true)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not convert external amount to incognito amount - Error %v", err)
				rejectedInst := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OutOfRangeUni64Error, []byte{})
				return [][]string{rejectedInst}, state, ac, nil
			}

			// calculate shielding reward (in pDecimal)
			reward, err := CalRewardForRefillVault(vault, incAmount.Uint64())
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not calcalate shielding reward - Error %v", err)
				rejectedInst := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OutOfRangeUni64Error, []byte{})
				return [][]string{rejectedInst}, state, ac, nil
			}

			// update state: list of shielding txs, vault state (waiting unshield amount, waiting reward)
			// TODO: update clonedAC?
			evmInfo.ListTxUsedInBlock = append(evmInfo.ListTxUsedInBlock, uniqTx)
			clonedAC, err := clonedAC.UpdateUniqTxsUsed(shieldData.NetworkID, evmInfo.ListTxUsedInBlock)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not update accumulate values for shield request - Error %v", err)
				rejectedInst := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
				return [][]string{rejectedInst}, state, ac, nil
			}

			Logger.log.Infof("Cloned AC after update: %+v\n", clonedAC)
			clonedVaults[shieldData.IncTokenID], err = updateVaultForRefill(vault, incAmount.Uint64(), reward)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not update vault state for shield request - Error %v", err)
				rejectedInst := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
				return [][]string{rejectedInst}, state, ac, nil
			}

			acceptedShieldData = append(acceptedShieldData, metadataBridge.AcceptedShieldRequestData{
				ShieldAmount:    incAmount.Uint64(),
				Reward:          reward,
				UniqTx:          uniqTx,
				ExternalTokenID: extTokenID,
				NetworkID:       shieldData.NetworkID,
				IncTokenID:      shieldData.IncTokenID,
			})

		default:
			Logger.log.Errorf("[BridgeAgg] Network ID is not matched: %v", shieldData.NetworkID)
			rejectedInst := buildRejectedInst(
				metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, NotFoundNetworkIDError, []byte{})
			return [][]string{rejectedInst}, state, ac, nil

		}
	}

	// build accepted instruction
	// incAddrStr was validated in func ExtractIssueEVMDataFromReceipt => don't catch error here
	key, _ := wallet.Base58CheckDeserialize(incAddrStr)
	receivingShardID, _ := metadataBridge.GetShardIDFromPaymentAddress(key.KeySet.PaymentAddress)
	content, err := buildAcceptedShieldContent(acceptedShieldData, key.KeySet.PaymentAddress, meta.UnifiedTokenID, action.TxReqID, receivingShardID)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Can not build contents for shield instruction: %v", err)
		rejectedInst := buildRejectedInst(
			metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
		return [][]string{rejectedInst}, state, ac, nil
	}

	resInst := buildAcceptedInst(metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, [][]byte{content})
	// update vaults state
	state.unifiedTokenVaults[meta.UnifiedTokenID] = clonedVaults
	// TODO: review 0xkraken
	clonedAC.DBridgeTokenPair[meta.UnifiedTokenID.String()] = GetExternalTokenIDForUnifiedToken()
	return resInst, state, clonedAC, nil
}

//TODO: 0xkraken
func (sp *stateProducer) handleWaitingUnshieldReqs(
	state *State,
	beaconHeight uint64,
	stateDB *statedb.StateDB,
) ([][]string, *State, error) {
	clonedState := state.Clone()
	insts := [][]string{}
	var err error

	for unifiedTokenID, waitingUnshieldReqs := range state.waitingUnshieldReqs {
		for i, waitingUnshieldReq := range waitingUnshieldReqs {
			vaults, ok := clonedState.unifiedTokenVaults[unifiedTokenID]
			if !ok {
				Logger.log.Errorf("[BridgeAgg] UnifiedTokenID is not found: %v", unifiedTokenID)
				continue
			}

			// check vault enough for process this waiting unshield req
			isEnoughVault := checkVaultForWaitUnshieldReq(vaults, waitingUnshieldReq.GetData())
			if !isEnoughVault {
				continue
			}

			// update state
			clonedState, err = updateStateForUnshield(clonedState, unifiedTokenID, waitingUnshieldReq, common.FilledStatusStr, i)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Update state error: %v", err)
				continue
			}

			// build burning confirm insts
			burningConfirmInsts := buildBurningConfirmInsts(waitingUnshieldReq)
			// build unshield inst with filled status
			// unshield requests in waiting list isDepositToSC is always false
			filledInst := buildUnshieldInst(unifiedTokenID, false, waitingUnshieldReq, common.FilledStatusStr, common.BridgeShardID)
			insts = append(burningConfirmInsts, filledInst)
		}
	}

	return insts, clonedState, nil
}

func (sp *stateProducer) unshield(
	contentStr string,
	state *State,
	beaconHeight uint64, shardID byte,
	stateDB *statedb.StateDB,
) ([][]string, *State, error) {
	// decode action from shard
	action := metadataCommon.NewAction()
	meta := &metadataBridge.UnshieldRequest{}
	action.Meta = meta
	err := action.FromString(contentStr)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Can not decode unshield action from shard %v", err)
		return [][]string{}, state, nil
	}

	var insts [][]string
	var burningConfirmInsts [][]string
	clonedState := state.Clone()

	// check UnifiedTokenID
	vaults, err := clonedState.CloneVaultsByUnifiedTokenID(meta.UnifiedTokenID)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] UnifiedTokenID is not found: %v", meta.UnifiedTokenID)
		rejectedInst := buildRejectedUnshieldReqInst(
			*meta, shardID, action.TxReqID, NotFoundTokenIDInNetworkError)
		return [][]string{rejectedInst}, state, nil
	}

	// check vaults
	isEnoughVault, waitingUnshieldDatas, err := checkVaultForNewUnshieldReq(
		vaults,
		meta.Data,
		meta.IsDepositToSC,
		clonedState.param.PercentFeeWithDec(),
		stateDB,
	)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Error when checking for unshield: %v", err)
		rejectedInst := buildRejectedUnshieldReqInst(
			*meta, shardID, action.TxReqID, CheckVaultUnshieldError)
		return [][]string{rejectedInst}, state, nil
	}

	waitingUnshieldReq := statedb.NewBridgeAggWaitingUnshieldReqStateWithValue(waitingUnshieldDatas, action.TxReqID, beaconHeight)
	statusStr := common.WaitingStatusStr

	if isEnoughVault {
		statusStr = common.AcceptedStatusStr

		// build burning confirm insts
		burningConfirmInsts = buildBurningConfirmInsts(waitingUnshieldReq)
		insts = append(insts, burningConfirmInsts...)
	}

	// build unshield inst with accepted/waiting status
	unshieldInst := buildUnshieldInst(meta.UnifiedTokenID, meta.IsDepositToSC, waitingUnshieldReq, statusStr, shardID)
	insts = append(insts, unshieldInst)

	// update state
	clonedState, err = updateStateForUnshield(clonedState, meta.UnifiedTokenID, waitingUnshieldReq, statusStr, -1)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Error when updating state for unshield: %v", err)
		rejectedInst := buildRejectedUnshieldReqInst(
			*meta, shardID, action.TxReqID, ProcessUnshieldError)
		return [][]string{rejectedInst}, state, nil
	}

	return insts, clonedState, nil
}

func (sp *stateProducer) addToken(
	state *State, beaconHeight uint64,
	sDBs map[int]*statedb.StateDB, ac *metadata.AccumulatedValues, checkpoint uint64,
) ([]string, *State, *metadata.AccumulatedValues, error) {
	var clonedUnifiedTokenInfos map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState
	addToken := metadataBridge.AddToken{}
	configUnifiedTokens := config.UnifiedToken()
	var clonedAC *metadata.AccumulatedValues
	var newListTokens map[common.Hash]map[common.Hash]config.Vault

	//at beacon height after trigger punified , we get param from configUnifiedTokens (retrieving by date key, ex 20220422)
	if unifiedTokens, found := configUnifiedTokens[checkpoint]; found {
		clonedUnifiedTokenInfos = state.CloneUnifiedTokenVaults()
		unifiedTokenIDs := make(map[string]bool)
		incTokenIDs := make(map[string]bool)
		newListTokens = make(map[common.Hash]map[common.Hash]config.Vault)
		clonedAC = ac.Clone()
		for unifiedTokenID, vaults := range unifiedTokens {
			if unifiedTokenID == common.PRVCoinID || unifiedTokenID == common.PDEXCoinID {
				Logger.log.Warnf("BridgeAggAddToken unifiedTokenID is prv or pdex")
				return []string{}, state, ac, nil
			}
			if unifiedTokenID.IsZeroValue() {
				Logger.log.Warnf("BridgeAggAddToken Found empty unifiedTokenID")
				return []string{}, state, ac, nil
			}
			newListTokens[unifiedTokenID] = make(map[common.Hash]config.Vault)
			if unifiedTokenIDs[unifiedTokenID.String()] {
				Logger.log.Warnf("BridgeAggAddToken Duplicate unifiedTokenID %s", unifiedTokenID.String())
				return []string{}, state, ac, nil
			}
			if incTokenIDs[unifiedTokenID.String()] {
				Logger.log.Warnf("BridgeAggAddToken Duplicate incTokenID %s", unifiedTokenID.String())
				return []string{}, state, ac, nil
			}
			isExisted, err := statedb.CheckTokenIDExisted(sDBs, unifiedTokenID)
			if err != nil {
				Logger.log.Warnf("BridgeAggAddToken WARNING: Error in finding unifiedTokenID %s", unifiedTokenID.String())
				return []string{}, state, ac, nil
			}
			if isExisted {
				Logger.log.Warnf("BridgeAggAddToken WARNING: unifiedTokenID %s has existed", unifiedTokenID.String())
				return []string{}, state, ac, nil
			}
			unifiedTokenIDs[unifiedTokenID.String()] = true
			if _, found := clonedUnifiedTokenInfos[unifiedTokenID]; !found {
				clonedUnifiedTokenInfos[unifiedTokenID] = make(map[common.Hash]*statedb.BridgeAggVaultState)
			}
			externalTokenIDIndex := make(map[string]bool)
			for tokenID, vault := range vaults {
				if tokenID.IsZeroValue() {
					Logger.log.Warnf("BridgeAggAddToken WARNING: incTokenID cannot be empty")
					return []string{}, state, ac, nil
				}
				if unifiedTokenIDs[tokenID.String()] {
					Logger.log.Warnf("BridgeAggAddToken Duplicate incTokenID with unifiedTokenID %s", tokenID.String())
					return []string{}, state, ac, nil
				}
				err := validateConfigVault(sDBs, tokenID, vault)
				if err != nil {
					Logger.log.Warnf("BridgeAggAddToken Validate config vault fail by error %v", err)
					return []string{}, state, ac, nil
				}
				if _, found := state.unifiedTokenVaults[unifiedTokenID][tokenID]; found {
					Logger.log.Warnf("BridgeAggAddToken Add an existed vault unifiedTokenID %s tokenID %s", unifiedTokenID.String(), tokenID.String())
					return []string{}, state, ac, nil
				}
				externalTokenID, err := getExternalTokenIDByNetworkID(vault.ExternalTokenID, vault.NetworkID)
				if err != nil {
					Logger.log.Warnf("BridgeAggAddToken Cannot get externalTokenID error %v", err)
					return []string{}, state, ac, nil
				}
				if externalTokenIDIndex[string(externalTokenID)] {
					Logger.log.Warnf("BridgeAggAddToken Duplicate externalTokenID")
					return []string{}, state, ac, nil
				}
				externalTokenIDIndex[string(externalTokenID)] = true
				v := statedb.NewBridgeAggVaultStateWithValue(0, 0, 0, 0, vault.ExternalDecimal, vault.NetworkID, tokenID)
				clonedUnifiedTokenInfos[unifiedTokenID][tokenID] = v
				incTokenIDs[tokenID.String()] = true
				clonedAC.DBridgeTokenPair[tokenID.String()] = externalTokenID
				clonedAC.DBridgeTokenPair[unifiedTokenID.String()] = GetExternalTokenIDForUnifiedToken()
				newListTokens[unifiedTokenID][tokenID] = vault
			}
		}
		if len(incTokenIDs) != 0 {
			addToken.NewListTokens = newListTokens
		}
	} else {
		Logger.log.Warnf("BridgeAggAddToken Notfound checkpoint %d", checkpoint)
		return []string{}, state, ac, nil
	}

	if len(addToken.NewListTokens) != 0 {
		temp, err := addToken.StringSlice()
		if err != nil {
			Logger.log.Warnf("BridgeAggAddToken Error in building instruction %v", err)
			return []string{}, state, ac, nil
		}
		state.unifiedTokenVaults = clonedUnifiedTokenInfos
		Logger.log.Info("BridgeAggAddToken instruction:", temp)
		return temp, state, clonedAC, nil
	}
	return []string{}, state, ac, nil
}
