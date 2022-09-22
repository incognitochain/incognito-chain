package bridgeagg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
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
			*meta, shardID, action.TxReqID, NotFoundUnifiedTokenIDError)
		return [][]string{rejectedInst}, state, nil
	}

	// check IncTokenID
	vault, ok := clonedVaults[meta.TokenID]
	if !ok {
		Logger.log.Errorf("[BridgeAgg] IncTokenID is not found: %v", meta.TokenID)
		rejectedInst := buildRejectedConvertReqInst(
			*meta, shardID, action.TxReqID, InvalidPTokenIDError)
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
			*meta, shardID, action.TxReqID, CalRewardError)
		return [][]string{rejectedInst}, state, nil
	}

	// update state
	vault, err = updateVaultForRefill(vault, pUnifiedTokenConvertAmt, reward)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Error convert to punified token amount: %v", err)
		rejectedInst := buildRejectedConvertReqInst(
			*meta, shardID, action.TxReqID, ProducerUpdateStateError)
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
			metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, NotFoundUnifiedTokenIDError, []byte{})
		return [][]string{rejectedInst}, state, ac, nil
	}

	clonedAC := ac.Clone()
	acceptedShieldData := []metadataBridge.AcceptedShieldRequestData{}
	var incAddrStr string
	var acceptedReshieldInstructions [][]string
	for i, shieldData := range meta.Data {
		// check incTokenID
		vault, ok := clonedVaults[shieldData.IncTokenID]
		if !ok || vault == nil {
			Logger.log.Errorf("[BridgeAgg] Vault IncTokenID is not found: %v", shieldData.IncTokenID)
			rejectedInst := buildRejectedInst(
				metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, InvalidPTokenIDError, []byte{})
			return [][]string{rejectedInst}, state, ac, nil
		}

		// check networkID
		if vault.NetworkID() != shieldData.NetworkID {
			Logger.log.Errorf("[BridgeAgg] Network ID is not matched: %v", shieldData.NetworkID)
			rejectedInst := buildRejectedInst(
				metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, InvalidNetworkIDError, []byte{})
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
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, UnmarshalShieldProofError, []byte{})
				return [][]string{rejectedInst}, state, ac, nil
			}

			evmInfo, err := metadataBridge.GetEVMInfoByNetworkID(shieldData.NetworkID, clonedAC)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not get evm info by network id - Error %v", err)
				rejectedInst := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, InvalidNetworkIDError, []byte{})
				return [][]string{rejectedInst}, state, ac, nil
			}

			// extract info from Receipt
			extAmount, incAddr, extTokenID, errShield := metadataBridge.ExtractIssueEVMDataFromReceipt(
				txReceipt, evmInfo.ContractAddress, evmInfo.Prefix, incAddrStr,
			)
			redepositDataLst, errReshield := metadataBridge.ExtractRedepositEVMDataFromReceipt(txReceipt, evmInfo.ContractAddress, evmInfo.Prefix)
			if errShield == nil {
				incAddrStr = incAddr
				key, _ := wallet.Base58CheckDeserialize(incAddr)
				shardID, _ := metadataBridge.GetShardIDFromPaymentAddress(key.KeySet.PaymentAddress)
				uniqTx := append(proofData.BlockHash[:], []byte(strconv.Itoa(int(proofData.TxIndex)))...)
				redepositDataLst = append(redepositDataLst, metadataBridge.DepositEventData{
					Amount:          extAmount,
					ReceiverStr:     incAddr,
					ExternalTokenID: extTokenID,
					IncTxID:         uniqTx,
					ShardID:         shardID,
					IsOneTime:       false,
				})
			} else {
				Logger.log.Errorf("[BridgeAgg] Can not extract data from Receipt - Error %v", errShield)
				if errReshield != nil {
					Logger.log.Warnf("[BridgeAgg] Extract Redeposit EVM events failed - %v", errReshield)
					rejectedInst := buildRejectedInst(
						metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, ExtractDataFromReceiptError, []byte{})
					return [][]string{rejectedInst}, state, ac, nil
				}
			}

			var validShieldProof bool = false
			for _, d := range redepositDataLst {
				vault, _ := clonedVaults[shieldData.IncTokenID]
				// check double use shielding proof in current block and previous blocks
				isValid, _, err := ValidateDoubleShieldProof(d.IncTxID, evmInfo.ListTxUsedInBlock, evmInfo.IsTxHashIssued, stateDBs[common.BeaconChainID])
				if err != nil {
					Logger.log.Errorf("[BridgeAgg] Cannot validate double shielding proof - Error %v", err)
					continue
				}
				if !isValid {
					Logger.log.Errorf("[BridgeAgg] Shielding proof was submitted - Error %v", err)
					continue
				}
				// verify token pair pTokenID - extTokenID
				err = metadataBridge.VerifyTokenPair(stateDBs, ac, shieldData.IncTokenID, d.ExternalTokenID)
				if err != nil {
					Logger.log.Errorf("[BridgeAgg] Invalid token pair pTokenID and extTokenID - Error %v", err)
					continue
				}

				// calculate shielding amount (in pDecimal)
				extDecimal := vault.ExtDecimal()
				if !bytes.Equal(append([]byte(evmInfo.Prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), d.ExternalTokenID) {
					if extDecimal > config.Param().BridgeAggParam.BaseDecimal {
						extDecimal = config.Param().BridgeAggParam.BaseDecimal
					}
				}
				incAmount, err := ConvertAmountByDecimal(d.Amount, extDecimal, true)
				if err != nil {
					Logger.log.Errorf("[BridgeAgg] Cannot convert external amount to incognito amount - Error %v", err)
					continue
				}

				// calculate shielding reward (in pDecimal)
				reward, err := CalRewardForRefillVault(vault, incAmount.Uint64())
				if err != nil {
					Logger.log.Errorf("[BridgeAgg] Cannot calculate shielding reward - Error %v", err)
					continue
				}

				// update state: list of shielding txs, vault state (waiting unshield amount, waiting reward)
				evmInfo.ListTxUsedInBlock = append(evmInfo.ListTxUsedInBlock, d.IncTxID)
				clonedAC, err := clonedAC.UpdateUniqTxsUsed(shieldData.NetworkID, evmInfo.ListTxUsedInBlock)
				if err != nil {
					Logger.log.Errorf("[BridgeAgg] Cannot update accumulate values for shield request - Error %v", err)
					continue
				}

				Logger.log.Infof("Cloned AC after update: %+v\n", clonedAC)
				clonedVaults[shieldData.IncTokenID], err = updateVaultForRefill(vault, incAmount.Uint64(), reward)
				if err != nil {
					Logger.log.Errorf("[BridgeAgg] Cannot update vault state for shield request - Error %v", err)
					continue
				}
				validShieldProof = true

				if d.IsOneTime {
					var recv privacy.OTAReceiver
					recv.FromString(d.ReceiverStr)
					c := metadataBridge.AcceptedReshieldRequest{
						UnifiedTokenID: &meta.UnifiedTokenID,
						Receiver:       recv,
						TxReqID:        action.TxReqID,
						ReshieldData: metadataBridge.AcceptedShieldRequestData{
							ShieldAmount:    incAmount.Uint64(),
							Reward:          reward,
							UniqTx:          d.IncTxID,
							ExternalTokenID: d.ExternalTokenID,
							NetworkID:       shieldData.NetworkID,
							IncTokenID:      shieldData.IncTokenID,
						},
					}
					contentBytes, _ := json.Marshal(c)
					inst := metadataCommon.NewInstructionWithValue(
						metadataCommon.IssuingReshieldResponseMeta,
						common.AcceptedStatusStr,
						d.ShardID,
						base64.StdEncoding.EncodeToString(contentBytes),
					)
					acceptedReshieldInstructions = append(acceptedReshieldInstructions, inst.StringSlice())
				} else {
					acceptedShieldData = append(acceptedShieldData, metadataBridge.AcceptedShieldRequestData{
						ShieldAmount:    incAmount.Uint64(),
						Reward:          reward,
						UniqTx:          d.IncTxID,
						ExternalTokenID: d.ExternalTokenID,
						NetworkID:       shieldData.NetworkID,
						IncTokenID:      shieldData.IncTokenID,
					})
				}
			}
			if !validShieldProof {
				Logger.log.Errorf("[BridgeAgg] no valid shield event found in proof")
				rejectedInst := buildRejectedInst(metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, NoValidShieldEventError, []byte{})
				return [][]string{rejectedInst}, state, ac, nil
			}

		default:
			Logger.log.Errorf("[BridgeAgg] Network ID is not matched: %v", shieldData.NetworkID)
			rejectedInst := buildRejectedInst(
				metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, InvalidNetworkIDError, []byte{})
			return [][]string{rejectedInst}, state, ac, nil
		}
	}

	var resInst [][]string
	if len(acceptedShieldData) > 0 {
		// build accepted instruction
		// incAddrStr was validated in func ExtractIssueEVMDataFromReceipt => don't catch error here
		key, _ := wallet.Base58CheckDeserialize(incAddrStr)
		receivingShardID, _ := metadataBridge.GetShardIDFromPaymentAddress(key.KeySet.PaymentAddress)
		content, err := buildAcceptedShieldContent(acceptedShieldData, key.KeySet.PaymentAddress, meta.UnifiedTokenID, action.TxReqID)
		if err != nil {
			Logger.log.Errorf("[BridgeAgg] Cannot build contents for shield instruction: %v", err)
			rejectedInst := buildRejectedInst(
				metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
			return [][]string{rejectedInst}, state, ac, nil
		}
		resInst = append(resInst, buildAcceptedInst(metadataCommon.IssuingUnifiedTokenRequestMeta, receivingShardID, [][]byte{content})...)
	}

	resInst = append(resInst, acceptedReshieldInstructions...)
	// update vaults state
	state.unifiedTokenVaults[meta.UnifiedTokenID] = clonedVaults
	clonedAC.DBridgeTokenPair[meta.UnifiedTokenID.String()] = GetExternalTokenIDForUnifiedToken()
	return resInst, state, clonedAC, nil
}

func (sp *stateProducer) handleWaitingUnshieldReqs(
	state *State,
	beaconHeight uint64,
	stateDB *statedb.StateDB,
) ([][]string, *State, error) {
	clonedState := state.Clone()
	insts := [][]string{}
	var err error
	var isEnoughVault bool
	lockedVaultAmts := map[common.Hash]map[common.Hash]uint64{}

	for unifiedTokenID, waitingUnshieldReqs := range state.waitingUnshieldReqs {
		lockedVaultAmts[unifiedTokenID] = map[common.Hash]uint64{}
		for _, waitingUnshieldReq := range waitingUnshieldReqs {
			vaults, ok := clonedState.unifiedTokenVaults[unifiedTokenID]
			if !ok {
				Logger.log.Errorf("[BridgeAgg] UnifiedTokenID is not found: %v", unifiedTokenID)
				continue
			}

			// check vault enough for process this waiting unshield req
			isEnoughVault, lockedVaultAmts[unifiedTokenID] = checkVaultForWaitUnshieldReq(vaults, waitingUnshieldReq.GetData(), lockedVaultAmts[unifiedTokenID])
			if !isEnoughVault {
				continue
			}

			// update state
			clonedState, err = updateStateForUnshield(clonedState, unifiedTokenID, waitingUnshieldReq, common.FilledStatusStr)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Update state error: %v", err)
				continue
			}

			// build burning confirm insts
			burningConfirmInsts := buildBurningConfirmInsts(waitingUnshieldReq, beaconHeight)
			// build unshield inst with filled status
			// unshield requests in waiting list isDepositToSC is always false
			filledInst := buildUnshieldInst(unifiedTokenID, false, waitingUnshieldReq, common.FilledStatusStr, common.BridgeShardID)
			insts = append(burningConfirmInsts, filledInst)
		}
	}

	return insts, clonedState, nil
}

func (sp *stateProducer) unshield(
	unshieldAction UnshieldActionForProducer,
	state *State,
	beaconHeightForConfirmInst uint64,
	stateDB *statedb.StateDB,
) ([][]string, *State, error) {
	meta, ok := unshieldAction.Meta.(*metadataBridge.UnshieldRequest)
	if !ok {
		Logger.log.Errorf("[BridgeAgg] Cannot parse BridgeAgg UnshieldRequest metadata")
		return [][]string{}, state, nil
	}
	txReqID := unshieldAction.TxReqID
	beaconHeightForUnshield := unshieldAction.BeaconHeight
	shardID := unshieldAction.ShardID

	var insts [][]string
	var burningConfirmInsts [][]string
	clonedState := state.Clone()

	// check UnifiedTokenID
	vaults, err := clonedState.CloneVaultsByUnifiedTokenID(meta.UnifiedTokenID)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] UnifiedTokenID is not found: %v", meta.UnifiedTokenID)
		rejectedInst := buildRejectedUnshieldReqInst(
			*meta, shardID, txReqID, NotFoundUnifiedTokenIDError)
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
			*meta, shardID, txReqID, CheckVaultUnshieldError)
		return [][]string{rejectedInst}, state, nil
	}

	waitingUnshieldReq := statedb.NewBridgeAggWaitingUnshieldReqStateWithValue(waitingUnshieldDatas, txReqID, beaconHeightForUnshield)
	statusStr := common.WaitingStatusStr

	if isEnoughVault {
		statusStr = common.AcceptedStatusStr

		// build burning confirm insts
		burningConfirmInsts = buildBurningConfirmInsts(waitingUnshieldReq, beaconHeightForConfirmInst)
		insts = append(insts, burningConfirmInsts...)
	}

	// build unshield inst with accepted/waiting status
	unshieldInst := buildUnshieldInst(meta.UnifiedTokenID, meta.IsDepositToSC, waitingUnshieldReq, statusStr, shardID)
	insts = append(insts, unshieldInst)

	// update state
	clonedState, err = updateStateForUnshield(clonedState, meta.UnifiedTokenID, waitingUnshieldReq, statusStr)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Error when updating state for unshield: %v", err)
		rejectedInst := buildRejectedUnshieldReqInst(
			*meta, shardID, txReqID, ProducerUpdateStateError)
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
			_, foundUnifiedToken := clonedUnifiedTokenInfos[unifiedTokenID]
			if isExisted && !foundUnifiedToken {
				Logger.log.Warnf("BridgeAggAddToken WARNING: unifiedTokenID %s has existed in privacy tokens", unifiedTokenID.String())
				return []string{}, state, ac, nil
			}
			unifiedTokenIDs[unifiedTokenID.String()] = true
			if !foundUnifiedToken {
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
				externalTokenID, err := GetExternalTokenIDByNetworkID(vault.ExternalTokenID, vault.NetworkID)
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

func (sp *stateProducer) burnForCall(
	action UnshieldActionForProducer,
	state *State,
	beaconHeightForConfirmInst uint64,
	stateDB *statedb.StateDB,
) ([][]string, *State, error) {
	meta, ok := action.Meta.(*metadataBridge.BurnForCallRequest)
	if !ok {
		return [][]string{}, state, fmt.Errorf("Cannot parse BurnForCallRequest metadata")
	}
	txReqID := action.TxReqID
	beaconHeightForUnshield := action.BeaconHeight
	shardID := action.ShardID

	var insts [][]string
	var burningConfirmInsts [][]string
	clonedState := state.Clone()

	useUnifiedToken := meta.BurnTokenID != meta.Data[0].IncTokenID
	vaults, errVault := clonedState.CloneVaultsByUnifiedTokenID(meta.BurnTokenID)
	unifiedVaultExists := errVault == nil
	if useUnifiedToken {
		if !unifiedVaultExists {
			Logger.log.Errorf("[BridgeAgg] UnifiedTokenID is not found: %v", meta.BurnTokenID)
			rejectedInst := buildRejectedBurnForCallReqInst(
				*meta, shardID, txReqID, NotFoundUnifiedTokenIDError)
			return [][]string{rejectedInst}, state, nil
		}

		var unshieldRequestDataForVault []metadataBridge.UnshieldRequestData
		for _, d := range meta.Data {
			unshieldRequestDataForVault = append(unshieldRequestDataForVault, metadataBridge.UnshieldRequestData{
				IncTokenID:        d.IncTokenID,
				BurningAmount:     d.BurningAmount,
				MinExpectedAmount: d.BurningAmount,
			})
		}

		// check vault balance
		isEnoughVault, waitingUnshieldDatas, err := checkVaultForNewUnshieldReq(
			vaults, unshieldRequestDataForVault,
			true, // use accept/reject flow without waiting
			clonedState.param.PercentFeeWithDec(), stateDB,
		)
		if err != nil {
			Logger.log.Errorf("[BridgeAgg] Error when checking for burnForCall: %v", err)
			rejectedInst := buildRejectedBurnForCallReqInst(
				*meta, shardID, txReqID, CheckVaultUnshieldError)
			return [][]string{rejectedInst}, state, nil
		}

		if isEnoughVault {
			rejectedInst := buildRejectedBurnForCallReqInst(*meta, shardID, txReqID, ProducerUpdateStateError)
			// create waitingUnshieldReq to update state
			waitingUnshieldReq := statedb.NewBridgeAggWaitingUnshieldReqStateWithValue(waitingUnshieldDatas, txReqID, beaconHeightForUnshield)

			for i := range meta.Data {
				netPrefix, err := getPrefixByNetworkID(meta.Data[i].ExternalNetworkID)
				if err != nil {
					Logger.log.Errorf("[BridgeAgg] BurnForCall external network error %v", err)
					return [][]string{rejectedInst}, state, nil
				}
				// build burning confirm insts
				extAddressForInst, err := metadataBridge.TrimNetworkPrefix(waitingUnshieldDatas[i].ExternalTokenID, netPrefix)
				if err != nil {
					Logger.log.Errorf("[BridgeAgg] BurnForCall external address error %v", err)
					return [][]string{rejectedInst}, state, nil
				}
				burningConfirmInsts = buildBurnForCallConfirmInsts(meta, waitingUnshieldReq.GetUnshieldID(), &waitingUnshieldDatas[i], extAddressForInst, meta.Data[i].ExternalNetworkID, i, beaconHeightForConfirmInst)
				insts = append(insts, burningConfirmInsts...)
			}
			// use unshield state update functions
			unshieldInst := buildAcceptedBurnForCallReqInst(&meta.BurnTokenID, waitingUnshieldReq, common.AcceptedStatusStr, shardID)
			insts = append(insts, unshieldInst)
			clonedState, err = updateStateForUnshield(clonedState, meta.BurnTokenID, waitingUnshieldReq, common.AcceptedStatusStr)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Error producing burnForCall: %v", err)
				return [][]string{rejectedInst}, state, nil
			}
		} else {
			rejectedInst := buildRejectedBurnForCallReqInst(
				*meta, shardID, txReqID, InsufficientFundsVaultError)
			return [][]string{rejectedInst}, state, nil
		}
	} else {
		rejectedInst := buildRejectedBurnForCallReqInst(*meta, shardID, txReqID, ProducerUpdateStateError)
		if unifiedVaultExists {
			return [][]string{rejectedInst}, state, fmt.Errorf("expect non-unified burnTokenID, got %v", meta.BurnTokenID)
		}
		var waitingUnshieldDatas []statedb.WaitingUnshieldReqData
		for _, d := range meta.Data {
			if d.IncTokenID != meta.BurnTokenID {
				return [][]string{rejectedInst}, state, fmt.Errorf("non-unified burnTokenID mismatch")
			}
			netPrefix, err := getPrefixByNetworkID(d.ExternalNetworkID)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Error when updating state for non-unified burnForCall: %v", err)
				return [][]string{rejectedInst}, state, nil
			}
			// Convert to external tokenID
			tokenID, err := metadataBridge.FindExternalTokenID(stateDB, d.IncTokenID, netPrefix)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Error when updating state for non-unified burnForCall: %v", err)
				return [][]string{rejectedInst}, state, nil
			}
			extAddressForInst, err := metadataBridge.TrimNetworkPrefix(tokenID, netPrefix)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] BurnForCall non-unified external address error %v", err)
				return [][]string{rejectedInst}, state, nil
			}

			// Convert amount to big.Int to get bytes later
			amount := big.NewInt(0).SetUint64(d.BurningAmount)
			if bytes.Equal(tokenID, append([]byte(netPrefix), rCommon.HexToAddress(common.NativeToken).Bytes()...)) {
				// Convert Gwei to Wei for Ether
				amount = amount.Mul(amount, big.NewInt(1000000000))
			}

			wdata := statedb.WaitingUnshieldReqData{
				ExternalTokenID:        tokenID,
				ExternalReceivedAmt:    amount,
				IncTokenID:             d.IncTokenID,
				BurningAmount:          d.BurningAmount,
				BurningConfirmMetaType: metadataCommon.BurnForCallConfirmMeta,
			}
			waitingUnshieldDatas = append(waitingUnshieldDatas, wdata)
			burningConfirmInsts = buildBurnForCallConfirmInsts(meta, txReqID, &wdata, extAddressForInst, d.ExternalNetworkID, 0, beaconHeightForConfirmInst)
			insts = append(insts, burningConfirmInsts...)
		}
		waitingUnshieldReq := statedb.NewBridgeAggWaitingUnshieldReqStateWithValue(waitingUnshieldDatas, txReqID, beaconHeightForUnshield)
		unshieldInst := buildAcceptedBurnForCallReqInst(nil, waitingUnshieldReq, common.AcceptedStatusStr, shardID)
		insts = append(insts, unshieldInst)
	}

	return insts, clonedState, nil
}
