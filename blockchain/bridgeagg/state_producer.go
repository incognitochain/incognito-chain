package bridgeagg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridgeAgg "github.com/incognitochain/incognito-chain/metadata/bridgeagg"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/utils"
)

type stateProducer struct {
}

func (sp *stateProducer) modifyListTokens(
	contentStr string,
	unifiedTokenInfos map[common.Hash]map[uint]*Vault,
	sDBs map[int]*statedb.StateDB,
) ([]string, map[common.Hash]map[uint]*Vault, error) {
	action := metadataCommon.NewAction()
	md := &metadataBridgeAgg.ModifyListToken{}
	action.Meta = md
	err := action.FromString(contentStr)
	if err != nil {
		return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.BridgeAggModifyListTokenMeta,
		common.AcceptedStatusStr,
		action.ShardID,
		utils.EmptyString,
	)
	rejectContent := metadataCommon.NewRejectContentWithValue(action.TxReqID, 0, nil)

	for unifiedTokenID, vaults := range md.NewListTokens {
		if err := CheckTokenIDExisted(sDBs, unifiedTokenID); err != nil {
			rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetworkError].Code
			temp, err := inst.StringSliceWithRejectContent(rejectContent)
			if err != nil {
				return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
			}
			return temp, unifiedTokenInfos, nil
		}
		_, found := unifiedTokenInfos[unifiedTokenID]
		if !found {
			unifiedTokenInfos[unifiedTokenID] = make(map[uint]*Vault)
		}
		for _, vault := range vaults {
			if err := CheckTokenIDExisted(sDBs, vault.TokenID()); err != nil {
				rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetworkError].Code
				temp, err := inst.StringSliceWithRejectContent(rejectContent)
				if err != nil {
					return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
				}
				return temp, unifiedTokenInfos, nil
			}
			if _, found := unifiedTokenInfos[unifiedTokenID][vault.NetworkID()]; !found {
				unifiedTokenInfos[unifiedTokenID][vault.NetworkID()] = NewVaultWithValue(
					*statedb.NewBridgeAggVaultState(), []byte{}, vault.TokenID(),
				)
			} else {
				newRewardReserve := vault.RewardReserve
				v := unifiedTokenInfos[unifiedTokenID][vault.NetworkID()]
				lastUpdatedRewardReserve := v.LastUpdatedRewardReserve()
				currentRewardReserve := v.CurrentRewardReserve()
				if newRewardReserve < lastUpdatedRewardReserve-currentRewardReserve {
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
				unifiedTokenInfos[unifiedTokenID][vault.NetworkID()] = v
			}
		}
	}
	acceptedContent := metadataBridgeAgg.AcceptedModifyListToken{
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
	contentStr string, unifiedTokenInfos map[common.Hash]map[uint]*Vault, sDBs map[int]*statedb.StateDB,
) ([]string, map[common.Hash]map[uint]*Vault, error) {
	action := metadataCommon.NewAction()
	md := &metadataBridgeAgg.ConvertTokenToUnifiedTokenRequest{}
	action.Meta = md
	err := action.FromString(contentStr)
	if err != nil {
		return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta,
		common.AcceptedStatusStr,
		action.ShardID,
		utils.EmptyString,
	)
	rejectContent := metadataCommon.NewRejectContentWithValue(action.TxReqID, 0, md)
	if _, found := unifiedTokenInfos[md.UnifiedTokenID]; !found {
		rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetworkError].Code
		temp, err := inst.StringSliceWithRejectContent(rejectContent)
		if err != nil {
			return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
		}
		return temp, unifiedTokenInfos, nil
	}
	if vault, found := unifiedTokenInfos[md.UnifiedTokenID][md.NetworkID]; !found {
		rejectContent.ErrorCode = ErrCodeMessage[NotFoundNetworkIDError].Code
		temp, err := inst.StringSliceWithRejectContent(rejectContent)
		if err != nil {
			return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
		}
		return temp, unifiedTokenInfos, nil
	} else {
		if vault.tokenID.String() != md.TokenID.String() {
			rejectContent.ErrorCode = ErrCodeMessage[NotFoundTokenIDInNetworkError].Code
			temp, err := inst.StringSliceWithRejectContent(rejectContent)
			if err != nil {
				return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
			}
			return temp, unifiedTokenInfos, nil

		}
		err = vault.convert(md.Amount)
		if err != nil {
			Logger.log.Warnf("Invalid shield amount error: %v tx %s", err, action.TxReqID.String())
			rejectContent.ErrorCode = ErrCodeMessage[InvalidConvertAmountError].Code
			temp, e := inst.StringSliceWithRejectContent(rejectContent)
			if e != nil {
				return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, e)
			}
			return temp, unifiedTokenInfos, nil

		}
		unifiedTokenInfos[md.UnifiedTokenID][md.NetworkID] = vault
		acceptedContent := metadataBridgeAgg.AcceptedConvertTokenToUnifiedToken{
			ConvertTokenToUnifiedTokenRequest: *md,
			TxReqID:                           action.TxReqID,
		}
		contentBytes, err := json.Marshal(acceptedContent)
		if err != nil {
			return []string{}, unifiedTokenInfos, err
		}
		inst.Content = base64.StdEncoding.EncodeToString(contentBytes)
	}
	return inst.StringSlice(), unifiedTokenInfos, nil
}

func (sp *stateProducer) shield(
	action ShieldAction, unifiedTokenInfos map[common.Hash]map[uint]*Vault, ac *metadata.AccumulatedValues,
) ([]string, map[common.Hash]map[uint]*Vault, error) {
	tempInst := metadataCommon.NewInstruction()
	if err := tempInst.FromStringSlice(action.Content); err != nil {
		return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	contentBytes, err := base64.StdEncoding.DecodeString(tempInst.Content)
	if err != nil {
		return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	content := metadata.IssuingEVMAcceptedInst{}
	err = json.Unmarshal(contentBytes, &content)
	if err != nil {
		return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.IssuingUnifiedTokenRequestMeta,
		common.AcceptedStatusStr,
		content.ShardID,
		utils.EmptyString,
	)
	rejectContent := metadataCommon.NewRejectContentWithValue(content.TxReqID, 0, nil)
	vault := unifiedTokenInfos[content.IncTokenID][content.NetworkID] // check available before
	actualAmount, err := vault.shield(content.IssuingAmount)
	if err != nil {
		Logger.log.Warnf("Calculate shield amount error: %v tx %s", err, content.TxReqID.String())
		rejectContent.ErrorCode = ErrCodeMessage[CalculateShieldAmountError].Code
		temp, e := inst.StringSliceWithRejectContent(rejectContent)
		if e != nil {
			return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
		}
		return temp, unifiedTokenInfos, nil
	}
	unifiedTokenInfos[content.IncTokenID][content.NetworkID] = vault
	// build instruction content
	content.Reward = actualAmount - content.IssuingAmount
	content.IssuingAmount = actualAmount
	content.UniqTx = action.UniqTx

	contentBytes, err = json.Marshal(content)
	if err != nil {
		return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	switch content.NetworkID {
	case common.ETHNetworkID:
		if content.IncTokenID == common.PRVCoinID {
			ac.UniqPRVEVMTxsUsed = append(ac.UniqPRVEVMTxsUsed, action.UniqTx)
		} else {
			ac.UniqETHTxsUsed = append(ac.UniqETHTxsUsed, action.UniqTx)
		}
	case common.BSCNetworkID:
		if content.IncTokenID == common.PRVCoinID {
			ac.UniqPRVEVMTxsUsed = append(ac.UniqPRVEVMTxsUsed, action.UniqTx)
		} else {
			ac.UniqBSCTxsUsed = append(ac.UniqBSCTxsUsed, action.UniqTx)
		}
	case common.PLGNetworkID:
		ac.UniqPLGTxsUsed = append(ac.UniqPLGTxsUsed, action.UniqTx)
	}
	inst.Content = base64.StdEncoding.EncodeToString(contentBytes)
	return inst.StringSlice(), unifiedTokenInfos, nil
}

func (sp *stateProducer) unshield(
	action UnshieldAction, unifiedTokenInfos map[common.Hash]map[uint]*Vault, beaconHeight uint64,
) ([][]string, map[common.Hash]map[uint]*Vault, error) {
	md := action.Content
	var prefix string
	vault := unifiedTokenInfos[md.TokenID][md.NetworkID] // check available before

	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.BurningUnifiedTokenRequestMeta,
		common.AcceptedStatusStr,
		common.BridgeShardID,
		utils.EmptyString,
	)
	rejectContent := metadataCommon.NewRejectContentWithValue(action.TxReqID, 0, &md)
	actualAmount, err := vault.unshield(md.BurningAmount, md.ExpectedAmount)
	if err != nil {
		Logger.log.Warnf("Calculate unshield amount error: %v tx %s", err, action.TxReqID.String())
		rejectContent.ErrorCode = ErrCodeMessage[CalculateUnshieldAmountError].Code
		temp, e := inst.StringSliceWithRejectContent(rejectContent)
		if e != nil {
			return [][]string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
		}
		return [][]string{temp}, unifiedTokenInfos, nil
	}
	amount := big.NewInt(0).SetUint64(actualAmount)
	fee := md.BurningAmount - actualAmount
	incTokenID := vault.tokenID
	if md.TokenID == common.PRVCoinID {
		incTokenID = common.PRVCoinID
	}
	var burningMetaType int
	res := [][]string{}
	switch md.NetworkID {
	case common.ETHNetworkID:
		if *md.IsDepositToSC {
			burningMetaType = metadata.BurningConfirmForDepositToSCMetaV2
		} else {
			if md.TokenID == common.PRVCoinID {
				burningMetaType = metadata.BurningPRVERC20ConfirmMeta
			} else {
				burningMetaType = metadata.BurningConfirmMetaV2
			}
		}
	case common.BSCNetworkID:
		if *md.IsDepositToSC {
			burningMetaType = metadata.BurningPBSCConfirmForDepositToSCMeta
		} else {
			if md.TokenID == common.PRVCoinID {
				burningMetaType = metadata.BurningPRVBEP20ConfirmMeta
			} else {
				burningMetaType = metadata.BurningBSCConfirmMeta
			}
		}
		prefix = common.BSCPrefix
	case common.PLGNetworkID:
		if *md.IsDepositToSC {
			burningMetaType = metadata.BurningPLGForDepositToSCRequestMeta
		} else {
			burningMetaType = metadata.BurningPLGConfirmMeta
		}
		prefix = common.PLGPrefix
	case common.DefaultNetworkID:
		return res, unifiedTokenInfos, errors.New("Cannot unshield with default networkID")
	default:
		return res, unifiedTokenInfos, errors.New("Cannot recognize networkID")
	}
	if md.TokenID != common.PRVCoinID {
		if bytes.Equal(action.ExternalTokenID, append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...)) {
			amount = amount.Mul(amount, big.NewInt(1000000000))
		}
	}
	acceptedContent := metadata.AcceptedUnshieldRequest{
		TokenID:       md.TokenID,
		Amount:        actualAmount,
		Fee:           fee,
		NetworkID:     md.NetworkID,
		TxReqID:       action.TxReqID,
		BurnerAddress: md.BurnerAddress,
	}
	contentBytes, err := json.Marshal(acceptedContent)
	if err != nil {
		return [][]string{}, unifiedTokenInfos, err
	}
	inst.Content = base64.StdEncoding.EncodeToString(contentBytes)
	res = append(res, inst.StringSlice())
	h := big.NewInt(0).SetUint64(beaconHeight)
	burningConfirmInst := []string{
		strconv.Itoa(burningMetaType),
		strconv.Itoa(int(common.BridgeShardID)),
		base58.Base58Check{}.Encode(action.ExternalTokenID, 0x00),
		md.RemoteAddress,
		base58.Base58Check{}.Encode(amount.Bytes(), 0x00),
		action.TxReqID.String(),
		base58.Base58Check{}.Encode(incTokenID[:], 0x00),
		base58.Base58Check{}.Encode(h.Bytes(), 0x00),
	}
	res = append(res, burningConfirmInst)
	return res, unifiedTokenInfos, nil
}
