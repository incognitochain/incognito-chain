package bridgeagg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type stateProducer struct{}

func (sp *stateProducer) convert(
	contentStr string, state *State, sDBs map[int]*statedb.StateDB, shardID byte,
) (resInst [][]string, resState *State, err error) {
	var errorType int
	var contents [][]byte
	resState = state
	action := metadataCommon.NewAction()
	md := &metadataBridge.ConvertTokenToUnifiedTokenRequest{}
	action.Meta = md
	err = action.FromString(contentStr)
	if err != nil {
		err = NewBridgeAggErrorWithValue(OtherError, err)
		return
	}

	defer func() {
		if err != nil {
			rejectedConvertRequest := metadataBridge.RejectedConvertTokenToUnifiedToken{
				TokenID:  md.TokenID,
				Amount:   md.Amount,
				Receiver: md.Receiver,
			}
			Logger.log.Warnf("Convert token with tx %s err %v", action.TxReqID.String(), err)
			content, err := json.Marshal(rejectedConvertRequest)
			if err != nil {
				errorType = OtherError
				return
			}
			contents = append(contents, content)
		}
		resInst, err = buildInstruction(
			metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta,
			errorType, contents, action.TxReqID, shardID, err,
		)
		if err != nil {
			Logger.log.Warnf("Cannot buildInstruction with tx %s err %v", action.TxReqID.String(), err)
		}
		err = nil
	}()

	if _, found := resState.unifiedTokenVaults[md.UnifiedTokenID]; !found {
		errorType = NotFoundTokenIDInNetworkError
		err = errors.New("Cannot find unifiedTokenID")
		return
	}
	if vault, found := resState.unifiedTokenVaults[md.UnifiedTokenID][md.TokenID]; !found {
		errorType = NotFoundTokenIDInNetworkError
		err = fmt.Errorf("Cannot find tokenID %s", md.TokenID.String())
		return
	} else {
		v := vault.Clone()
		v, mintAmount, e := convert(v, md.Amount)
		if err != nil {
			Logger.log.Warnf("Invalid convert amount error: %v tx %s", err, action.TxReqID.String())
			errorType = InvalidConvertAmountError
			err = e
			return
		}
		acceptedContent := metadataBridge.AcceptedConvertTokenToUnifiedToken{
			ConvertTokenToUnifiedTokenRequest: *md,
			TxReqID:                           action.TxReqID,
			MintAmount:                        mintAmount,
		}
		var content []byte
		content, err = json.Marshal(acceptedContent)
		if err != nil {
			errorType = OtherError
			return
		}
		resState.unifiedTokenVaults[md.UnifiedTokenID][md.TokenID] = v
		contents = append(contents, content)
	}
	return
}

func (sp *stateProducer) shield(
	contentStr string,
	state *State,
	ac *metadata.AccumulatedValues,
	shardID byte,
	stateDBs map[int]*statedb.StateDB,
) ([][]string, *State, *metadata.AccumulatedValues, error) {
	// resInst [][]string,
	// resState *State,
	// resAC *metadata.AccumulatedValues,
	// err error,

	// decode action from shard
	action := metadataCommon.NewAction()
	meta := &metadataBridge.ShieldRequest{}
	action.Meta = meta
	err := action.FromString(contentStr)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Can not decode action shield from shard: %v - Error: %v", contentStr, err)
		return [][]string{}, state, ac, err
	}

	// check unifiedTokenID
	clonedVaults, err := state.CloneVaultsByUnifiedTokenID(meta.UnifiedTokenID)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] UnifiedTokenID is not found: %v", meta.UnifiedTokenID)
		rejectedInst, err := buildRejectedInst(
			metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, NotFoundTokenIDInNetworkError, []byte{})
		return [][]string{rejectedInst}, state, ac, err
	}

	clonedAC := ac.Clone()
	acceptedShieldData := []metadataBridge.AcceptedShieldRequestData{}
	var incAddrStr string
	for i, shieldData := range meta.Data {
		// check incTokenID
		vault, ok := clonedVaults[shieldData.IncTokenID]
		if !ok || vault == nil {
			Logger.log.Errorf("[BridgeAgg] UnifiedTokenID is not found: %v", meta.UnifiedTokenID)
			rejectedInst, err := buildRejectedInst(
				metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, NotFoundTokenIDInNetworkError, []byte{})
			return [][]string{rejectedInst}, state, ac, err
		}

		// check networkID
		if vault.NetworkID() != shieldData.NetworkID {
			Logger.log.Errorf("[BridgeAgg] Network ID is not matched: %v", shieldData.NetworkID)
			rejectedInst, err := buildRejectedInst(
				metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, NotFoundNetworkIDError, []byte{})
			return [][]string{rejectedInst}, state, ac, err
		}

		// validate shielding proof
		networkType, _ := metadataBridge.GetNetworkTypeByNetworkID(shieldData.NetworkID)
		switch networkType {
		case common.EVMNetworkType:
			// unmarshal proof and receipt
			proofData, txReceipt, err := UnmarshalEVMShieldProof(shieldData.Proof, action.ExtraData[i])
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not unmarshal shielding proof - Error %v", err)
				rejectedInst, err := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
				return [][]string{rejectedInst}, state, ac, err
			}

			evmInfo, err := metadataBridge.GetEVMInfoByNetworkID(shieldData.NetworkID, clonedAC)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not get evm info by network id - Error %v", err)
				rejectedInst, err := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
				return [][]string{rejectedInst}, state, ac, err
			}

			// check double use shielding proof in current block and previous blocks
			isValid, uniqTx, err := ValidateDoubleShieldProof(proofData, evmInfo.ListTxUsedInBlock, evmInfo.IsTxHashIssued, stateDBs[common.BeaconChainID])
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not validate double shielding proof - Error %v", err)
				rejectedInst, err := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
				return [][]string{rejectedInst}, state, ac, err
			}
			if !isValid {
				Logger.log.Errorf("[BridgeAgg] Can not validate double shielding proof - Error %v", err)
				rejectedInst, err := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, ShieldProofIsSubmittedError, []byte{})
				return [][]string{rejectedInst}, state, ac, err
			}

			// extract info from Receipt
			extAmount, incAddr, extTokenID, err := metadataBridge.ExtractIssueEVMDataFromReceipt(
				txReceipt, evmInfo.ContractAddress, evmInfo.Prefix, incAddrStr,
			)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not extract data from Receipt - Error %v", err)
				rejectedInst, err := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
				return [][]string{rejectedInst}, state, ac, err
			}
			incAddrStr = incAddr

			// verify token pair pTokenID - extTokenID
			err = metadataBridge.VerifyTokenPair(stateDBs, ac, shieldData.IncTokenID, extTokenID)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Invalid token pair pTokenID and extTokenID - Error %v", err)
				rejectedInst, err := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, FailToVerifyTokenPairError, []byte{})
				return [][]string{rejectedInst}, state, ac, NewBridgeAggErrorWithValue(FailToVerifyTokenPairError, err)
			}

			// calculate shielding amount (in pDecimal)
			extDecimal := vault.ExtDecimal()
			if !bytes.Equal(append([]byte(evmInfo.Prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), extTokenID) {
				if extDecimal > config.Param().BridgeAggParam.BaseDecimal {
					extDecimal = config.Param().BridgeAggParam.BaseDecimal
				}
			}
			incAmount, err := ConvertAmountByDecimal(*extAmount, extDecimal, true)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not convert external amount to incognito amount - Error %v", err)
				rejectedInst, err := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OutOfRangeUni64Error, []byte{})
				return [][]string{rejectedInst}, state, ac, NewBridgeAggErrorWithValue(OutOfRangeUni64Error, err)
			}

			// calculate shielding reward (in pDecimal)
			reward, err := calShieldReward(vault, incAmount.Uint64())
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not calcalate shielding reward - Error %v", err)
				rejectedInst, err := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OutOfRangeUni64Error, []byte{})
				return [][]string{rejectedInst}, state, ac, NewBridgeAggErrorWithValue(OutOfRangeUni64Error, err)
			}

			// update state: list of shielding txs, vault state (waiting unshield amount, waiting reward)
			// TODO: update clonedAC?
			evmInfo.ListTxUsedInBlock = append(evmInfo.ListTxUsedInBlock, uniqTx)
			Logger.log.Errorf("Cloned AC after update: %+v\n", clonedAC)
			clonedVaults[shieldData.IncTokenID], err = updateVaultForShielding(vault, incAmount.Uint64(), reward)
			if err != nil {
				Logger.log.Errorf("[BridgeAgg] Can not update vault state for shield request - Error %v", err)
				rejectedInst, err := buildRejectedInst(
					metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
				return [][]string{rejectedInst}, state, ac, NewBridgeAggErrorWithValue(OtherError, err)
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
			rejectedInst, err := buildRejectedInst(
				metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, NotFoundNetworkIDError, []byte{})
			return [][]string{rejectedInst}, state, ac, err

		}
	}

	// build accepted instruction
	// incAddrStr was validated in func ExtractIssueEVMDataFromReceipt => don't catch error here
	key, _ := wallet.Base58CheckDeserialize(incAddrStr)
	receivingShardID, _ := metadataBridge.GetShardIDFromPaymentAddress(key.KeySet.PaymentAddress)
	content, err := buildAcceptedShieldContent(acceptedShieldData, key.KeySet.PaymentAddress, meta.UnifiedTokenID, action.TxReqID, receivingShardID)
	if err != nil {
		Logger.log.Errorf("[BridgeAgg] Can not build contents for shield instruction: %v", err)
		rejectedInst, err := buildRejectedInst(
			metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, action.TxReqID, OtherError, []byte{})
		return [][]string{rejectedInst}, state, ac, err
	}

	resInst := buildAcceptedInst(metadataCommon.IssuingUnifiedTokenRequestMeta, shardID, [][]byte{content})
	// update vaults state
	state.unifiedTokenVaults[meta.UnifiedTokenID] = clonedVaults
	// TODO: review 0xkraken
	clonedAC.DBridgeTokenPair[meta.UnifiedTokenID.String()] = GetExternalTokenIDForUnifiedToken()
	return resInst, state, clonedAC, nil

	// var errorType int
	// var contents [][]byte
	// resState = state.Clone()
	// resAC = ac
	// action := metadataCommon.NewAction()
	// md := &metadataBridge.ShieldRequest{}
	// action.Meta = md
	// err = action.FromString(contentStr)
	// if err != nil {
	// 	err = NewBridgeAggErrorWithValue(OtherError, err)
	// 	return
	// }
	// defer func() {
	// 	if err != nil {
	// 		Logger.log.Warnf("tx %s shield with err %v", action.TxReqID.String(), err)
	// 		contents = append(contents, action.TxReqID.Bytes())
	// 	}
	// 	resInst, err = buildInstruction(
	// 		metadataCommon.IssuingUnifiedTokenRequestMeta,
	// 		errorType, contents, action.TxReqID, shardID, err,
	// 	)
	// 	if err != nil {
	// 		Logger.log.Warnf("Cannot build tx %s with err %v", action.TxReqID.String(), err)
	// 	}
	// 	err = nil
	// }()

	// vaults, e := resState.CloneVaultsByUnifiedTokenID(md.UnifiedTokenID)
	// if e != nil {
	// 	err = e
	// 	errorType = NotFoundTokenIDInNetworkError
	// 	return
	// }
	// clonedAC := ac.Clone()
	// var receiveShardID byte
	// acceptedShieldRequestData := make([]metadataBridge.AcceptedShieldRequestData, len(md.Data))
	// acceptedShieldRequestRewardData := make([]metadataBridge.AcceptedShieldRequestData, len(md.Data))
	// var rewardAmount uint64
	// paymentAddress := ""
	// for index, data := range md.Data {
	// 	vault, found := vaults[data.IncTokenID]
	// 	if !found {
	// 		err = fmt.Errorf("Not found tokenID %s", data.IncTokenID.String())
	// 		errorType = NotFoundTokenIDInNetworkError
	// 		return
	// 	}
	// 	if vault.NetworkID() != data.NetworkID {
	// 		err = fmt.Errorf("Expect networkID %v but get %v", vault.NetworkID(), data.NetworkID)
	// 		errorType = NotFoundNetworkIDError
	// 		return
	// 	}
	// 	networkType, e := metadataBridge.GetNetworkTypeByNetworkID(data.NetworkID)
	// 	if e != nil {
	// 		err = e
	// 		errorType = OtherError
	// 		return
	// 	}
	// 	switch networkType {
	// 	case common.EVMNetworkType:
	// 		proofData := metadataBridge.EVMProof{}
	// 		err = json.Unmarshal(data.Proof, &proofData)
	// 		if err != nil {
	// 			errorType = OtherError
	// 			return
	// 		}
	// 		tempVault, actualAmount, reward, receivingShardID, token, uniqTX, addressStr, tempAC, et, e := shieldEVM(
	// 			md.UnifiedTokenID, data.IncTokenID, data.NetworkID, clonedAC, shardID,
	// 			action.TxReqID, vault, stateDBs,
	// 			action.ExtraData[index], proofData.BlockHash, proofData.TxIndex, paymentAddress,
	// 		)
	// 		if e != nil {
	// 			errorType = et
	// 			err = e
	// 			return
	// 		}
	// 		if actualAmount == 0 {
	// 			errorType = CalculateShieldAmountError
	// 			err = errors.New("Actual receive amount from shield cannot be 0")
	// 			return
	// 		}
	// 		paymentAddress = addressStr
	// 		acceptedShieldRequestData[index].IssuingAmount = actualAmount - reward
	// 		acceptedShieldRequestRewardData[index].IssuingAmount = reward
	// 		acceptedShieldRequestData[index].ExternalTokenID = token
	// 		acceptedShieldRequestData[index].UniqTx = uniqTX
	// 		acceptedShieldRequestData[index].NetworkID = data.NetworkID
	// 		acceptedShieldRequestRewardData[index].NetworkID = data.NetworkID
	// 		acceptedShieldRequestData[index].IncTokenID = data.IncTokenID
	// 		acceptedShieldRequestRewardData[index].IncTokenID = data.IncTokenID
	// 		receiveShardID = receivingShardID
	// 		rewardAmount += reward
	// 		clonedAC = tempAC
	// 		vault = tempVault
	// 	default:
	// 		errorType = OtherError
	// 		err = errors.New("Not found networkType")
	// 		return
	// 	}
	// }
	// key, e := wallet.Base58CheckDeserialize(paymentAddress)
	// if e != nil {
	// 	err = e
	// 	errorType = OtherError
	// 	return
	// }
	// contents, err = buildAcceptedShieldContents(
	// 	acceptedShieldRequestData, acceptedShieldRequestRewardData,
	// 	key.KeySet.PaymentAddress, md.UnifiedTokenID, action.TxReqID, receiveShardID, rewardAmount != 0,
	// )
	// if err != nil {
	// 	errorType = OtherError
	// 	return
	// }
	// resState.unifiedTokenVaults[md.UnifiedTokenID] = vaults
	// resAC = clonedAC
	// return
}

func (sp *stateProducer) unshield(
	contentStr string,
	state *State,
	beaconHeight uint64, shardID byte,
	stateDB *statedb.StateDB,
) (resInsts [][]string, resState *State, err error) {
	// decode action from shard
	action := metadataCommon.NewAction()
	md := &metadataBridge.UnshieldRequest{}
	action.Meta = md
	err = action.FromString(contentStr)
	if err != nil {
		err = NewBridgeAggErrorWithValue(OtherError, err)
		return [][]string{}, nil, err
	}

	var errorType int
	var contents [][]byte
	var burningInsts [][]string
	resState = state.Clone()

	defer func() {
		if err != nil {
			// create rejected instruction when ocurring error
			var burningAmount uint64
			for _, data := range md.Data {
				burningAmount += data.BurningAmount
			}
			rejectedUnshieldRequest := metadataBridge.RejectedUnshieldRequest{
				UnifiedTokenID: md.UnifiedTokenID,
				Amount:         burningAmount,
				Receiver:       md.Receiver,
			}
			Logger.log.Warnf("Unshield with tx %s err %v", action.TxReqID.String(), err)
			content, err := json.Marshal(rejectedUnshieldRequest)
			if err != nil {
				errorType = OtherError
				return
			}
			contents = append(contents, content)
		}
		insts, e := buildInstruction(
			metadataCommon.BurningUnifiedTokenRequestMeta,
			errorType, contents, action.TxReqID, shardID, err,
		)
		if e != nil {
			Logger.log.Warnf("Cannot buildInstruction with tx %s err %v", action.TxReqID.String(), e)
		}
		err = nil
		resInsts = append(resInsts, insts...)
		if len(burningInsts) != 0 {
			resInsts = append(resInsts, burningInsts...)
		}
	}()

	// check UnifiedTokenID
	vaults, err := resState.CloneVaultsByUnifiedTokenID(md.UnifiedTokenID)
	if err != nil {
		errorType = NotFoundTokenIDInNetworkError
		err = errors.New("Invalid unified token ID not found")
		return
	}

	var acceptedUnshieldRequestDatas []metadataBridge.AcceptedUnshieldRequestData
	isAddedWaitingList := false
	//TODO: 0xkraken params
	percentFee := float64(0.1)

	for index, data := range md.Data {
		// check IncTokenID
		vault, found := vaults[data.IncTokenID]
		if !found {
			errorType = NotFoundTokenIDInNetworkError
			err = fmt.Errorf("Not found tokenID %s in unified tokenID %s", data.IncTokenID.String(), md.UnifiedTokenID)
			burningInsts = [][]string{}
			return
		}
		// get networkType by NetworkID
		networkType, _ := metadataBridge.GetNetworkTypeByNetworkID(vault.NetworkID())
		switch networkType {
		case common.EVMNetworkType:
			// calculate unshield fee
			receivedAmount, fee, e := calUnshieldFee(vault, data.BurningAmount, data.MinExpectedAmount, percentFee)
			if e != nil {
				errorType = NotFoundTokenIDInNetworkError
				err = e
				burningInsts = [][]string{}
				return
			}

			// add to waiting list
			if fee > 0 {
				isAddedWaitingList = true
			}
			if isAddedWaitingList {
				acceptedUnshieldRequestData := metadataBridge.AcceptedUnshieldRequestData{
					BurningAmount:  data.BurningAmount,
					ReceivedAmount: receivedAmount,
					IncTokenID:     data.IncTokenID,
				}
				acceptedUnshieldRequestDatas = append(acceptedUnshieldRequestDatas, acceptedUnshieldRequestData)
				continue
			}

			// create burning confirm instruction if don't add to waiting list

			tempVault, externalTokenID, unshieldAmount, amount, fee, burningMetaType, et, e := unshieldEVM(data, stateDB, vault, action.TxReqID, md.IsDepositToSC)
			if e != nil {
				errorType = et
				err = e
				burningInsts = [][]string{}
				return
			}
			if amount == 0 {
				errorType = CalculateUnshieldAmountError
				err = errors.New("Actual receive amount is 0 with unshielding request")
				burningInsts = [][]string{}
				return
			}
			vault = tempVault
			h := big.NewInt(0).SetUint64(beaconHeight)
			newTxReqID := common.HashH(append(action.TxReqID.Bytes(), common.IntToBytes(index)...))
			burningInst := []string{
				strconv.Itoa(burningMetaType),
				strconv.Itoa(int(common.BridgeShardID)),
				base58.Base58Check{}.Encode(externalTokenID, 0x00),
				data.RemoteAddress,
				base58.Base58Check{}.Encode(unshieldAmount.Bytes(), 0x00),
				newTxReqID.String(),
				base58.Base58Check{}.Encode(data.IncTokenID[:], 0x00),
				base58.Base58Check{}.Encode(h.Bytes(), 0x00),
			}
			burningInsts = append(burningInsts, burningInst)
			//TODO: 0xkraken review
			acceptedUnshieldRequestData := metadataBridge.AcceptedUnshieldRequestData{
				BurningAmount:  data.BurningAmount,
				ReceivedAmount: receivedAmount,
				IncTokenID:     data.IncTokenID,
			}
			acceptedUnshieldRequestDatas = append(acceptedUnshieldRequestDatas, acceptedUnshieldRequestData)
		default:
			errorType = NotFoundNetworkIDError
			err = errors.New("Not found networkID")
			burningInsts = [][]string{}
			return
		}
	}
	resState.unifiedTokenVaults[md.UnifiedTokenID] = vaults

	acceptedContent := metadataBridge.AcceptedInstUnshieldRequest{
		UnifiedTokenID: md.UnifiedTokenID,
		TxReqID:        action.TxReqID,
		Data:           acceptedUnshieldRequestDatas,
	}
	content, e := json.Marshal(acceptedContent)
	if e != nil {
		err = e
		errorType = OtherError
		burningInsts = [][]string{}
		return
	}
	contents = append(contents, content)

	return
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
					continue
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
