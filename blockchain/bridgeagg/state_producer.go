package bridgeagg

import (
	"encoding/base64"
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
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/incognitochain/incognito-chain/wallet"
)

type stateProducer struct {
}

func (sp *stateProducer) modifyRewardReserve(
	contentStr string,
	unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault,
	sDBs map[int]*statedb.StateDB,
	shardID byte,
) ([]string, map[common.Hash]map[common.Hash]*Vault, error) {
	action := metadataCommon.NewAction()
	md := &metadataBridge.ModifyRewardReserve{}
	action.Meta = md
	err := action.FromString(contentStr)
	if err != nil {
		return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	resVaults := make(map[common.Hash][]metadataBridge.Vault)
	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.BridgeAggModifyRewardReserveMeta,
		common.AcceptedStatusStr,
		shardID,
		utils.EmptyString,
	)
	rejectContent := metadataCommon.NewRejectContentWithValue(action.TxReqID, 0, nil)

	clonedUnifiedTokenInfos := CloneUnifiedTokenInfos(unifiedTokenInfos)
	for unifiedTokenID, vaults := range md.Vaults {
		if _, found := clonedUnifiedTokenInfos[unifiedTokenID]; !found {
			rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetworkError].Code
			temp, err := inst.StringSliceWithRejectContent(rejectContent)
			if err != nil {
				return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
			}
			return temp, unifiedTokenInfos, nil
		}
		for _, vault := range vaults {
			v, found := clonedUnifiedTokenInfos[unifiedTokenID][vault.TokenID]
			if !found {
				rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetworkError].Code
				temp, err := inst.StringSliceWithRejectContent(rejectContent)
				if err != nil {
					return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
				}
				return temp, unifiedTokenInfos, nil
			}
			err := v.updateRewardReserve(vault.RewardReserve, vault.IsPaused)
			if err != nil {
				Logger.log.Warnf("tx %s UpdatedRewardReserve is invalid err %v", action.TxReqID.String(), err)
				rejectContent.ErrorCode = ErrCodeMessage[InvalidRewardReserveError].Code
				temp, err := inst.StringSliceWithRejectContent(rejectContent)
				if err != nil {
					return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(InvalidRewardReserveError, err)
				}
				return temp, unifiedTokenInfos, nil
			}
			clonedUnifiedTokenInfos[unifiedTokenID][vault.TokenID] = v
			resVaults[unifiedTokenID] = append(resVaults[unifiedTokenID], metadataBridge.Vault{
				RewardReserve: vault.RewardReserve,
				TokenID:       vault.TokenID,
				IsPaused:      vault.IsPaused,
			})
		}
	}
	acceptedContent := metadataBridge.AcceptedModifyRewardReserve{
		Vaults:  resVaults,
		TxReqID: action.TxReqID,
	}
	contentBytes, err := json.Marshal(acceptedContent)
	if err != nil {
		return []string{}, unifiedTokenInfos, err
	}
	inst.Content = base64.StdEncoding.EncodeToString(contentBytes)
	unifiedTokenInfos = clonedUnifiedTokenInfos
	return inst.StringSlice(), unifiedTokenInfos, nil
}

func (sp *stateProducer) convert(
	contentStr string, unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault, sDBs map[int]*statedb.StateDB, shardID byte,
) (resInst [][]string, resUnifiedTokenInfos map[common.Hash]map[common.Hash]*Vault, err error) {
	var errorType int
	var contents [][]byte
	resUnifiedTokenInfos = unifiedTokenInfos
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

	if _, found := resUnifiedTokenInfos[md.UnifiedTokenID]; !found {
		errorType = NotFoundTokenIDInNetworkError
		err = errors.New("Cannot find unifiedTokenID")
		return
	}
	if vault, found := resUnifiedTokenInfos[md.UnifiedTokenID][md.TokenID]; !found {
		errorType = NotFoundTokenIDInNetworkError
		err = fmt.Errorf("Cannot find tokenID %s", md.TokenID.String())
		return
	} else {
		v := vault.Clone()
		mintAmount, e := v.convert(md.Amount)
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
		resUnifiedTokenInfos[md.UnifiedTokenID][md.TokenID] = v
		contents = append(contents, content)
	}
	return
}

func (sp *stateProducer) shield(
	contentStr string, unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault,
	ac *metadata.AccumulatedValues, shardID byte,
	stateDBs map[int]*statedb.StateDB,
) (
	resInst [][]string,
	resUnifiedTokenInfos map[common.Hash]map[common.Hash]*Vault,
	resAC *metadata.AccumulatedValues,
	err error,
) {
	var errorType int
	var contents [][]byte
	resUnifiedTokenInfos = unifiedTokenInfos
	resAC = ac
	action := metadataCommon.NewAction()
	md := &metadataBridge.ShieldRequest{}
	action.Meta = md
	err = action.FromString(contentStr)
	if err != nil {
		err = NewBridgeAggErrorWithValue(OtherError, err)
		return
	}
	defer func() {
		if err != nil {
			Logger.log.Warnf("tx %s shield with err %v", action.TxReqID.String(), err)
			contents = append(contents, action.TxReqID.Bytes())
		}
		resInst, err = buildInstruction(
			metadataCommon.IssuingUnifiedTokenRequestMeta,
			errorType, contents, action.TxReqID, shardID, err,
		)
		if err != nil {
			Logger.log.Warnf("Cannot build tx %s with err %v", action.TxReqID.String(), err)
		}
		err = nil
	}()

	vaults, e := CloneVaults(resUnifiedTokenInfos, md.UnifiedTokenID)
	if e != nil {
		err = e
		errorType = NotFoundTokenIDInNetworkError
		return
	}
	clonedAC := ac.Clone()
	var receiveShardID byte
	acceptedShieldRequestData := make([]metadataBridge.AcceptedShieldRequestData, len(md.Data))
	acceptedShieldRequestRewardData := make([]metadataBridge.AcceptedShieldRequestData, len(md.Data))
	var rewardAmount uint64
	paymentAddress := ""
	for index, data := range md.Data {
		vault, found := vaults[data.IncTokenID]
		if !found {
			err = fmt.Errorf("Not found tokenID %s", data.IncTokenID.String())
			errorType = NotFoundTokenIDInNetworkError
			return
		}
		if vault.NetworkID() != data.NetworkID {
			err = fmt.Errorf("Expect networkID %v but get %v", vault.NetworkID(), data.NetworkID)
			errorType = NotFoundNetworkIDError
			return
		}
		networkType, e := metadataBridge.GetNetworkTypeByNetworkID(data.NetworkID)
		if e != nil {
			err = e
			errorType = OtherError
			return
		}
		switch networkType {
		case common.EVMNetworkType:
			blockHash := rCommon.Hash{}
			err = blockHash.UnmarshalText([]byte(data.BlockHash))
			if err != nil {
				errorType = OtherError
				return
			}
			tempVault, actualAmount, reward, receivingShardID, token, uniqTX, addressStr, tempAC, et, e := shieldEVM(
				md.UnifiedTokenID, data.IncTokenID, data.NetworkID, clonedAC, shardID,
				action.TxReqID, vault, stateDBs,
				action.ExtraData[index], blockHash, data.TxIndex, paymentAddress,
			)
			if e != nil {
				errorType = et
				err = e
				return
			}
			if actualAmount == 0 {
				errorType = CalculateShieldAmountError
				err = errors.New("Actual receive amount from shield cannot be 0")
				return
			}
			paymentAddress = addressStr
			acceptedShieldRequestData[index].IssuingAmount = actualAmount - reward
			acceptedShieldRequestRewardData[index].IssuingAmount = reward
			acceptedShieldRequestData[index].ExternalTokenID = token
			acceptedShieldRequestData[index].UniqTx = uniqTX
			acceptedShieldRequestData[index].NetworkID = data.NetworkID
			acceptedShieldRequestRewardData[index].NetworkID = data.NetworkID
			acceptedShieldRequestData[index].IncTokenID = data.IncTokenID
			acceptedShieldRequestRewardData[index].IncTokenID = data.IncTokenID
			receiveShardID = receivingShardID
			rewardAmount += reward
			clonedAC = tempAC
			vault = tempVault
		default:
			errorType = OtherError
			err = errors.New("Not found networkType")
			return
		}
	}
	key, e := wallet.Base58CheckDeserialize(paymentAddress)
	if e != nil {
		err = e
		errorType = OtherError
		return
	}
	contents, err = buildAcceptedShieldContents(
		acceptedShieldRequestData, acceptedShieldRequestRewardData,
		key.KeySet.PaymentAddress, md.UnifiedTokenID, action.TxReqID, receiveShardID, rewardAmount != 0,
	)
	if err != nil {
		errorType = OtherError
		return
	}
	resUnifiedTokenInfos[md.UnifiedTokenID] = vaults
	resAC = clonedAC
	return
}

func (sp *stateProducer) unshield(
	contentStr string, unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault,
	beaconHeight uint64, shardID byte,
	stateDB *statedb.StateDB,
) (resInsts [][]string, resUnifiedTokenInfos map[common.Hash]map[common.Hash]*Vault, err error) {
	action := metadataCommon.NewAction()
	md := &metadataBridge.UnshieldRequest{}
	action.Meta = md
	err = action.FromString(contentStr)
	if err != nil {
		err = NewBridgeAggErrorWithValue(OtherError, err)
		return
	}
	resUnifiedTokenInfos = unifiedTokenInfos
	var errorType int
	var contents [][]byte
	var burningInsts [][]string
	defer func() {
		if err != nil {
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

	vaults, err := CloneVaults(resUnifiedTokenInfos, md.UnifiedTokenID)
	if err != nil {
		errorType = NotFoundTokenIDInNetworkError
		return
	}
	var listAcceptedUnshieldRequestData []metadataBridge.AcceptedUnshieldRequestData

	for index, data := range md.Data {
		vault, found := vaults[data.IncTokenID]
		if !found {
			errorType = NotFoundTokenIDInNetworkError
			err = fmt.Errorf("Not found tokenID %s", data.IncTokenID.String())
			return
		}
		networkType, e := metadataBridge.GetNetworkTypeByNetworkID(vault.NetworkID())
		if e != nil {
			errorType = NotFoundNetworkIDError
			err = errors.New("Not found networkID")
			burningInsts = [][]string{}
			return
		}
		switch networkType {
		case common.EVMNetworkType:
			tempVault, externalTokenID, unshieldAmount, amount, fee, burningMetaType, et, e := unshieldEVM(data, stateDB, vault, action.TxReqID)
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
			acceptedUnshieldRequestData := metadataBridge.AcceptedUnshieldRequestData{
				Fee:           fee,
				Amount:        amount,
				IncTokenID:    data.IncTokenID,
				IsDepositToSC: data.IsDepositToSC,
			}
			listAcceptedUnshieldRequestData = append(listAcceptedUnshieldRequestData, acceptedUnshieldRequestData)
		default:
			errorType = NotFoundNetworkIDError
			err = errors.New("Not found networkID")
			burningInsts = [][]string{}
			return
		}
	}
	resUnifiedTokenInfos[md.UnifiedTokenID] = vaults

	acceptedContent := metadataBridge.AcceptedUnshieldRequest{
		UnifiedTokenID: md.UnifiedTokenID,
		TxReqID:        action.TxReqID,
		Data:           listAcceptedUnshieldRequestData,
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
	unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault, beaconHeight uint64,
	sDBs map[int]*statedb.StateDB, ac *metadata.AccumulatedValues, checkpoint uint64,
) ([]string, map[common.Hash]map[common.Hash]*Vault, *metadata.AccumulatedValues, error) {
	var clonedUnifiedTokenInfos map[common.Hash]map[common.Hash]*Vault
	addToken := metadataBridge.AddToken{}
	configUnifiedTokens := config.UnifiedToken()
	var clonedAC *metadata.AccumulatedValues
	var newListTokens map[common.Hash]map[common.Hash]config.Vault

	//at beacon height after trigger punified , we get param from configUnifiedTokens (retrieving by date key, ex 20220422)
	if unifiedTokens, found := configUnifiedTokens[checkpoint]; found {
		clonedUnifiedTokenInfos = CloneUnifiedTokenInfos(unifiedTokenInfos)
		unifiedTokenIDs := make(map[string]bool)
		incTokenIDs := make(map[string]bool)
		newListTokens = make(map[common.Hash]map[common.Hash]config.Vault)
		clonedAC = ac.Clone()
		for unifiedTokenID, vaults := range unifiedTokens {
			if unifiedTokenID == common.PRVCoinID || unifiedTokenID == common.PDEXCoinID {
				Logger.log.Warnf("BridgeAggAddToken unifiedTokenID is prv or pdex")
				return []string{}, unifiedTokenInfos, ac, nil
			}
			if unifiedTokenID.IsZeroValue() {
				Logger.log.Warnf("BridgeAggAddToken Found empty unifiedTokenID")
				return []string{}, unifiedTokenInfos, ac, nil
			}
			newListTokens[unifiedTokenID] = make(map[common.Hash]config.Vault)
			if unifiedTokenIDs[unifiedTokenID.String()] {
				Logger.log.Warnf("BridgeAggAddToken Duplicate unifiedTokenID %s", unifiedTokenID.String())
				return []string{}, unifiedTokenInfos, ac, nil
			}
			if incTokenIDs[unifiedTokenID.String()] {
				Logger.log.Warnf("BridgeAggAddToken Duplicate incTokenID %s", unifiedTokenID.String())
				return []string{}, unifiedTokenInfos, ac, nil
			}
			isExisted, err := statedb.CheckTokenIDExisted(sDBs, unifiedTokenID)
			if err != nil {
				Logger.log.Warnf("BridgeAggAddToken WARNING: Error in finding unifiedTokenID %s", unifiedTokenID.String())
				return []string{}, unifiedTokenInfos, ac, nil
			}
			if isExisted {
				Logger.log.Warnf("BridgeAggAddToken WARNING: unifiedTokenID %s has existed", unifiedTokenID.String())
				return []string{}, unifiedTokenInfos, ac, nil
			}
			unifiedTokenIDs[unifiedTokenID.String()] = true
			if _, found := clonedUnifiedTokenInfos[unifiedTokenID]; !found {
				clonedUnifiedTokenInfos[unifiedTokenID] = make(map[common.Hash]*Vault)
			}
			externalTokenIDIndex := make(map[string]bool)
			for tokenID, vault := range vaults {
				if tokenID.IsZeroValue() {
					Logger.log.Warnf("BridgeAggAddToken WARNING: incTokenID cannot be empty")
					return []string{}, unifiedTokenInfos, ac, nil
				}
				if unifiedTokenIDs[tokenID.String()] {
					Logger.log.Warnf("BridgeAggAddToken Duplicate incTokenID with unifiedTokenID %s", tokenID.String())
					return []string{}, unifiedTokenInfos, ac, nil
				}
				err := validateConfigVault(sDBs, tokenID, vault)
				if err != nil {
					Logger.log.Warnf("BridgeAggAddToken Validate config vault fail by error %v", err)
					return []string{}, unifiedTokenInfos, ac, nil
				}
				if _, found := unifiedTokenInfos[unifiedTokenID][tokenID]; found {
					continue
				}
				externalTokenID, err := getExternalTokenIDByNetworkID(vault.ExternalTokenID, vault.NetworkID)
				if err != nil {
					Logger.log.Warnf("BridgeAggAddToken Cannot get externalTokenID error %v", err)
					return []string{}, unifiedTokenInfos, ac, nil
				}
				//TODO: @tin validate externalTokenID
				if externalTokenIDIndex[string(externalTokenID)] {
					Logger.log.Warnf("BridgeAggAddToken Duplicate externalTokenID")
					return []string{}, unifiedTokenInfos, ac, nil
				}
				externalTokenIDIndex[string(externalTokenID)] = true
				state := statedb.NewBridgeAggVaultStateWithValue(0, 0, 0, vault.ExternalDecimal, false, vault.NetworkID, tokenID)
				v := NewVaultWithValue(*state)
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
		return []string{}, unifiedTokenInfos, ac, nil
	}

	if len(addToken.NewListTokens) != 0 {
		temp, err := addToken.StringSlice()
		if err != nil {
			Logger.log.Warnf("BridgeAggAddToken Error in building instruction %v", err)
			return []string{}, unifiedTokenInfos, ac, nil
		}
		Logger.log.Info("BridgeAggAddToken instruction:", temp)
		return temp, clonedUnifiedTokenInfos, clonedAC, nil
	}
	return []string{}, unifiedTokenInfos, ac, nil
}
