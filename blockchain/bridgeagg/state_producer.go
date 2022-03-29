package bridgeagg

import (
	"encoding/base64"
	"encoding/json"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/utils"
)

type stateProducer struct {
}

func (sp *stateProducer) modifyListTokens(
	contentStr string,
	unifiedTokenInfos map[common.Hash]map[uint]*Vault,
	sDBs map[int]*statedb.StateDB,
	shardID byte,
) ([]string, map[common.Hash]map[uint]*Vault, error) {
	action := metadataCommon.NewAction()
	md := &metadataBridge.ModifyListToken{}
	action.Meta = md
	err := action.FromString(contentStr)
	if err != nil {
		return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.BridgeAggModifyListTokenMeta,
		common.AcceptedStatusStr,
		shardID,
		utils.EmptyString,
	)
	rejectContent := metadataCommon.NewRejectContentWithValue(action.TxReqID, 0, nil)

	for unifiedTokenID, vaults := range md.NewListTokens {
		_, found := unifiedTokenInfos[unifiedTokenID]
		if !found {
			unifiedTokenInfos[unifiedTokenID] = make(map[uint]*Vault)
		}
		for _, vault := range vaults {
			if _, found := unifiedTokenInfos[unifiedTokenID][vault.NetworkID()]; !found {
				var prefix string
				switch vault.NetworkID() {
				case common.ETHNetworkID:
					prefix = ""
				case common.BSCNetworkID:
					prefix = common.BSCPrefix
				case common.PLGNetworkID:
					prefix = common.PLGPrefix
				}
				if _, err := metadataBridge.FindExternalTokenID(sDBs[common.BeaconChainID], vault.TokenID(), prefix, metadataCommon.ShieldUnifiedTokenRequestMeta); err != nil {
					Logger.log.Warnf("tx %s Cannot find externalTokenID with tokenID %s err %v", action.TxReqID.String(), vault.TokenID().String(), err)
					rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetworkError].Code
					temp, err := inst.StringSliceWithRejectContent(rejectContent)
					if err != nil {
						return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
					}
					return temp, unifiedTokenInfos, nil
				}
				unifiedTokenInfos[unifiedTokenID][vault.NetworkID()] = NewVaultWithValue(
					*statedb.NewBridgeAggVaultStateWithValue(
						0, vault.RewardReserve, vault.RewardReserve, vault.Decimal,
					), vault.TokenID(),
				)
			} else {
				newRewardReserve := vault.RewardReserve
				v := unifiedTokenInfos[unifiedTokenID][vault.NetworkID()]
				lastUpdatedRewardReserve := v.LastUpdatedRewardReserve()
				currentRewardReserve := v.CurrentRewardReserve()
				if newRewardReserve < lastUpdatedRewardReserve-currentRewardReserve {
					Logger.log.Warnf("tx %s UpdatedRewardReserve is invalid", action.TxReqID.String())
					rejectContent.ErrorCode = ErrCodeMessage[InvalidRewardReserveError].Code
					temp, err := inst.StringSliceWithRejectContent(rejectContent)
					if err != nil {
						return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(InvalidRewardReserveError, err)
					}
					return temp, unifiedTokenInfos, nil
				}
				v.SetLastUpdatedRewardReserve(vault.RewardReserve)
				v.SetCurrentRewardReserve(vault.RewardReserve)
				v.tokenID = vault.TokenID()
				v.SetDecimal(vault.Decimal)
				unifiedTokenInfos[unifiedTokenID][vault.NetworkID()] = v
			}
		}
	}
	acceptedContent := metadataBridge.AcceptedModifyListToken{
		ModifyListToken: *md,
		TxReqID:         action.TxReqID,
	}
	contentBytes, err := json.Marshal(acceptedContent)
	if err != nil {
		return []string{}, unifiedTokenInfos, err
	}
	inst.Content = base64.StdEncoding.EncodeToString(contentBytes)
	return inst.StringSlice(), unifiedTokenInfos, nil
}

func (sp *stateProducer) convert(
	contentStr string, unifiedTokenInfos map[common.Hash]map[uint]*Vault, sDBs map[int]*statedb.StateDB, shardID byte,
) (resInst [][]string, resUnifiedTokenInfos map[common.Hash]map[uint]*Vault, err error) {
	var errorType int
	var contents [][]byte
	shouldContinue := true
	resUnifiedTokenInfos = unifiedTokenInfos
	action := metadataCommon.NewAction()
	md := &metadataBridge.ConvertTokenToUnifiedTokenRequest{}
	action.Meta = md
	err = action.FromString(contentStr)
	if err != nil {
		err = NewBridgeAggErrorWithValue(OtherError, err)
		return
	}
	rejectedConvertRequest := metadataBridge.RejectedConvertTokenToUnifiedToken{
		TokenID:  md.TokenID,
		Amount:   md.Amount,
		Receiver: md.Receivers[md.TokenID],
	}

	defer func() {
		if shouldContinue {
			if err != nil {
				content, err := json.Marshal(rejectedConvertRequest)
				if err != nil {
					return
				}
				contents = append(contents, content)
			}
			resInst, err = buildInstruction(
				metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta,
				errorType, contents, action.TxReqID, shardID, err,
			)
			if err != nil {
				return
			}
			err = nil
		}
	}()

	if _, found := resUnifiedTokenInfos[md.UnifiedTokenID]; !found {
		errorType = NotFoundTokenIDInNetworkError
		return
	}
	if vault, found := resUnifiedTokenInfos[md.UnifiedTokenID][md.NetworkID]; !found {
		errorType = NotFoundNetworkIDError
		return
	} else {
		if vault.tokenID.String() != md.TokenID.String() {
			errorType = NotFoundTokenIDInNetworkError
			return
		}
		err = vault.convert(md.Amount, md.UnifiedTokenID == common.PRVCoinID)
		if err != nil {
			Logger.log.Warnf("Invalid convert amount error: %v tx %s", err, action.TxReqID.String())
			errorType = InvalidConvertAmountError
			return
		}
		resUnifiedTokenInfos[md.UnifiedTokenID][md.NetworkID] = vault
		acceptedContent := metadataBridge.AcceptedConvertTokenToUnifiedToken{
			ConvertTokenToUnifiedTokenRequest: *md,
			TxReqID:                           action.TxReqID,
		}
		content, e := json.Marshal(acceptedContent)
		if e != nil {
			shouldContinue = false
			err = e
			return
		}
		contents = append(contents, content)
	}
	return
}

func (sp *stateProducer) shield(
	contentStr string, unifiedTokenInfos map[common.Hash]map[uint]*Vault,
	ac *metadata.AccumulatedValues, shardID byte,
	stateDBs map[int]*statedb.StateDB,
) (resInst [][]string, resUnifiedTokenInfos map[common.Hash]map[uint]*Vault, err error) {
	var errorType int
	var contents [][]byte
	shouldContinue := true
	resUnifiedTokenInfos = unifiedTokenInfos
	action := metadataCommon.NewAction()
	md := &metadataBridge.ShieldRequest{}
	action.Meta = md
	err = action.FromString(contentStr)
	if err != nil {
		err = NewBridgeAggErrorWithValue(OtherError, err)
		return
	}
	defer func() {
		if shouldContinue {
			if err != nil {
				Logger.log.Warnf("tx %s shield with err %v", action.TxReqID.String(), err)
				contents = append(contents, action.TxReqID.Bytes())
			}
			resInst, err = buildInstruction(
				metadataCommon.ShieldUnifiedTokenRequestMeta,
				errorType, contents, action.TxReqID, shardID, err,
			)
			if err != nil {
				return
			}
			err = nil
		}
	}()

	vaults, err := CloneVaults(resUnifiedTokenInfos, md.IncTokenID)
	if err != nil {
		errorType = NotFoundTokenIDInNetworkError
		return
	}
	tempAC := new(metadata.AccumulatedValues)
	*tempAC = *ac
	var receiveShardID byte
	acceptedShieldRequestData := make([]metadataBridge.AcceptedShieldRequestData, len(md.Data))
	acceptedShieldRequestRewardData := make([]metadataBridge.AcceptedShieldRequestData, len(md.Data))

	shouldSkip := false
	for index, data := range md.Data {
		if shouldSkip {
			break
		}
		switch data.NetworkID {
		case common.BSCNetworkID, common.ETHNetworkID, common.PLGNetworkID:
			blockHash := rCommon.Hash{}
			e := blockHash.UnmarshalText([]byte(data.BlockHash))
			if e != nil {
				shouldSkip = true
				errorType = OtherError
				err = e
				continue
			}
			actualAmount, reward, receivingShardID, token, uniqTX, et, e := shieldEVM(
				md.IncTokenID, data.NetworkID, tempAC, shardID,
				action.TxReqID, vaults, stateDBs,
				action.ExtraData[index], blockHash, data.TxIndex,
			)
			if e != nil {
				shouldSkip = true
				errorType = et
				err = e
				continue
			}
			acceptedShieldRequestData[index].IssuingAmount = actualAmount - reward
			acceptedShieldRequestRewardData[index].IssuingAmount = reward
			acceptedShieldRequestData[index].ExternalTokenID = token
			acceptedShieldRequestData[index].UniqTx = uniqTX
			acceptedShieldRequestData[index].NetworkID = data.NetworkID
			acceptedShieldRequestRewardData[index].NetworkID = data.NetworkID
			receiveShardID = receivingShardID
		case common.DefaultNetworkID:
			errorType = OtherError
			shouldSkip = true
			return
		default:
			errorType = OtherError
			shouldSkip = true
			return
		}
	}
	if shouldSkip {
		if errorType == OtherError {
			shouldContinue = false
		}
		return
	}
	contents, err = buildAcceptedShieldContents(
		acceptedShieldRequestData, acceptedShieldRequestRewardData,
		md.PaymentAddress, md.IncTokenID, action.TxReqID, receiveShardID,
	)
	if err != nil {
		shouldContinue = false
		return
	}
	resUnifiedTokenInfos[md.IncTokenID] = vaults
	ac = tempAC
	return
}

func (sp *stateProducer) unshield(
	contentStr string, unifiedTokenInfos map[common.Hash]map[uint]*Vault, beaconHeight uint64, shardID byte,
	stateDB *statedb.StateDB,
) (resInsts [][]string, resUnifiedTokenInfos map[common.Hash]map[uint]*Vault, err error) {
	action := metadataCommon.NewAction()
	md := &metadataBridge.BurningRequest{}
	action.Meta = md
	err = action.FromString(contentStr)
	if err != nil {
		err = NewBridgeAggErrorWithValue(OtherError, err)
		return
	}
	var errorType int
	var contents [][]byte
	var burningInst []string
	shouldContinue := true
	defer func() {
		if shouldContinue {
			insts, err := buildInstruction(
				metadataCommon.UnshieldUnifiedTokenRequestMeta,
				errorType, contents, action.TxReqID, shardID, err,
			)
			if err != nil {
				return
			}
			err = nil
			resInsts = append(resInsts, insts...)
			if len(burningInst) != 0 {
				resInsts = append(resInsts, burningInst)
			}
		}
	}()

	//var prefix, contractAddress string
	//var burningMetaType int

	/*switch md.NetworkID {*/
	/*case common.ETHNetworkID:*/
	/*prefix = utils.EmptyString*/
	/*contractAddress = config.Param().PRVERC20ContractAddressStr*/
	/*if *md.IsDepositToSC {*/
	/*burningMetaType = metadata.BurningConfirmForDepositToSCMetaV2*/
	/*} else {*/
	/*if md.TokenID == common.PRVCoinID {*/
	/*burningMetaType = metadata.BurningPRVERC20ConfirmMeta*/
	/*} else {*/
	/*burningMetaType = metadata.BurningConfirmMetaV2*/
	/*}*/
	/*}*/
	/*case common.BSCNetworkID:*/
	/*prefix = common.BSCPrefix*/
	/*contractAddress = config.Param().PRVBEP20ContractAddressStr*/
	/*if *md.IsDepositToSC {*/
	/*burningMetaType = metadata.BurningPBSCConfirmForDepositToSCMeta*/
	/*} else {*/
	/*if md.TokenID == common.PRVCoinID {*/
	/*burningMetaType = metadata.BurningPRVBEP20ConfirmMeta*/
	/*} else {*/
	/*burningMetaType = metadata.BurningBSCConfirmMeta*/
	/*}*/
	/*}*/
	/*prefix = common.BSCPrefix*/
	/*case common.PLGNetworkID:*/
	/*prefix = common.PLGPrefix*/
	/*if *md.IsDepositToSC {*/
	/*burningMetaType = metadata.BurningPLGForDepositToSCRequestMeta*/
	/*} else {*/
	/*burningMetaType = metadata.BurningPLGConfirmMeta*/
	/*}*/
	/*prefix = common.PLGPrefix*/
	/*case common.DefaultNetworkID:*/
	/*err = errors.New("Cannot get info from default networkID")*/
	/*return*/
	/*default:*/
	/*err = errors.New("Cannot detect networkID")*/
	/*return*/
	/*}*/

	/*var externalTokenID []byte*/
	/*if md.TokenID.String() != common.PRVIDStr {*/
	/*// Convert to external tokenID*/
	/*externalTokenID, err = metadata.FindExternalTokenID(stateDB, md.TokenID, prefix, burningMetaType)*/
	/*if err != nil {*/
	/*errorType = NotFoundTokenIDInNetworkError*/
	/*return*/
	/*}*/
	/*} else {*/
	/*temp := rCommon.HexToAddress(contractAddress)*/
	/*externalTokenID = temp[:]*/
	/*}*/

	/*resUnifiedTokenInfos = unifiedTokenInfos*/
	/*vault, err := GetVault(resUnifiedTokenInfos, md.TokenID, md.NetworkID)*/
	/*if err != nil {*/
	/*errorType = NotFoundNetworkIDError*/
	/*return*/
	/*}*/

	/*actualAmount, err := vault.unshield(md.BurningAmount, md.ExpectedAmount)*/
	/*if err != nil {*/
	/*Logger.log.Warnf("Calculate unshield amount error: %v tx %s", err, action.TxReqID.String())*/
	/*errorType = CalculateUnshieldAmountError*/
	/*return*/
	/*}*/
	/*amount := big.NewInt(0).SetUint64(actualAmount)*/
	/*fee := md.BurningAmount - actualAmount*/
	/*incTokenID := vault.tokenID*/
	/*if md.TokenID == common.PRVCoinID {*/
	/*incTokenID = common.PRVCoinID*/
	/*}*/
	/*if md.TokenID != common.PRVCoinID {*/
	/*amount.Mul(amount, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(vault.Decimal())), nil))*/
	/*amount.Div(amount, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(config.Param().BridgeAggParam.BaseDecimal)), nil))*/
	/*if !amount.IsUint64() {*/
	/*Logger.log.Warnf("Calculate actual unshield amount is out of range uint64")*/
	/*errorType = OutOfRangeUni64Error*/
	/*err = fmt.Errorf("Out of range uint64")*/
	/*return*/
	/*}*/

	/*if bytes.Equal(externalTokenID, append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...)) {*/
	/*amount = amount.Mul(amount, big.NewInt(1000000000))*/
	/*}*/
	/*}*/
	/*acceptedContent := metadata.AcceptedUnshieldRequest{*/
	/*TokenID:       md.TokenID,*/
	/*Amount:        actualAmount,*/
	/*Fee:           fee,*/
	/*NetworkID:     md.NetworkID,*/
	/*TxReqID:       action.TxReqID,*/
	/*BurnerAddress: md.BurnerAddress,*/
	/*}*/
	/*content, err = json.Marshal(acceptedContent)*/
	/*if err != nil {*/
	/*shouldContinue = false*/
	/*return*/
	/*}*/

	/*h := big.NewInt(0).SetUint64(beaconHeight)*/
	/*burningInst = []string{*/
	/*strconv.Itoa(burningMetaType),*/
	/*strconv.Itoa(int(common.BridgeShardID)),*/
	/*base58.Base58Check{}.Encode(externalTokenID, 0x00),*/
	/*md.RemoteAddress,*/
	/*base58.Base58Check{}.Encode(amount.Bytes(), 0x00),*/
	/*action.TxReqID.String(),*/
	/*base58.Base58Check{}.Encode(incTokenID[:], 0x00),*/
	/*base58.Base58Check{}.Encode(h.Bytes(), 0x00),*/
	/*}*/
	return
}
