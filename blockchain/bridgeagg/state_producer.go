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
	"github.com/incognitochain/incognito-chain/config"
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
		if ok, err := statedb.CheckTokenIDExisted(sDBs, unifiedTokenID); err != nil || !ok {
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
			if ok, err := statedb.CheckTokenIDExisted(sDBs, vault.TokenID()); err != nil || !ok {
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
	contentStr string, unifiedTokenInfos map[common.Hash]map[uint]*Vault,
	ac *metadata.AccumulatedValues, shardID byte,
	stateDBs map[int]*statedb.StateDB,
) (resInst []string, resUnifiedTokenInfos map[common.Hash]map[uint]*Vault, err error) {
	var errorType int
	var content []byte
	shouldContinue := true
	resUnifiedTokenInfos = unifiedTokenInfos
	action, err := metadata.ParseEVMIssuingInstContent(contentStr)
	if err != nil {
		shouldContinue = false
		err = NewBridgeAggErrorWithValue(OtherError, err)
		return
	}
	md := action.Meta
	defer func() {
		if shouldContinue {
			resInst, err = buildInstruction(
				metadataCommon.IssuingUnifiedTokenRequestMeta,
				errorType, content, action.TxReqID, shardID, err,
			)
			if err != nil {
				return
			}
			err = nil
		}
	}()

	var listTxUsed [][]byte
	var contractAddress, prefix string
	var isTxHashIssued func(stateDB *statedb.StateDB, uniqueEthTx []byte) (bool, error)

	switch md.NetworkID {
	case common.ETHNetworkID:
		listTxUsed = ac.UniqETHTxsUsed
		contractAddress = config.Param().EthContractAddressStr
		prefix = utils.EmptyString
		isTxHashIssued = statedb.IsETHTxHashIssued
		if md.IncTokenID == common.PRVCoinID {
			contractAddress = config.Param().PRVERC20ContractAddressStr
			listTxUsed = ac.UniqPRVEVMTxsUsed
			isTxHashIssued = statedb.IsPRVEVMTxHashIssued
		}
	case common.BSCNetworkID:
		listTxUsed = ac.UniqBSCTxsUsed
		contractAddress = config.Param().BscContractAddressStr
		prefix = common.BSCPrefix
		isTxHashIssued = statedb.IsBSCTxHashIssued
		if md.IncTokenID == common.PRVCoinID {
			contractAddress = config.Param().PRVBEP20ContractAddressStr
			listTxUsed = ac.UniqPRVEVMTxsUsed
			prefix = utils.EmptyString
			isTxHashIssued = statedb.IsPRVEVMTxHashIssued
		}
	case common.PLGNetworkID:
		listTxUsed = ac.UniqPLGTxsUsed
		contractAddress = config.Param().PlgContractAddressStr
		prefix = common.PLGPrefix
		isTxHashIssued = statedb.IsPLGTxHashIssued
	case common.DefaultNetworkID:
		shouldContinue = false
		errorType = OtherError
		err = errors.New("Cannot get info from default networkID")
		return
	default:
		shouldContinue = false
		errorType = OtherError
		err = errors.New("Cannot detect networkID")
	}
	vault, err := GetVault(resUnifiedTokenInfos, md.IncTokenID, md.NetworkID)
	if err != nil {
		errorType = NotFoundTokenIDInNetworkError
		return
	}
	amount, receivingShardID, addressStr, token, uniqTx, err := metadata.ExtractIssueEVMData(
		stateDBs[common.BeaconChainID], shardID, listTxUsed, contractAddress, prefix, isTxHashIssued, action,
	)
	if err != nil {
		errorType = FailToExtractDataError
		return
	}
	err = metadata.VerifyTokenPair(stateDBs, ac, vault.tokenID, token)
	if err != nil {
		errorType = FailToVerifyTokenPairError
		return
	}
	actualAmount, err := vault.shield(amount)
	if err != nil {
		Logger.log.Warnf("Calculate shield amount error: %v tx %s", err, action.TxReqID.String())
		errorType = CalculateShieldAmountError
		return
	}
	resUnifiedTokenInfos[action.Meta.IncTokenID][action.Meta.NetworkID] = vault
	reward := actualAmount - amount

	switch action.Meta.NetworkID {
	case common.ETHNetworkID:
		if md.IncTokenID == common.PRVCoinID {
			ac.UniqPRVEVMTxsUsed = append(ac.UniqPRVEVMTxsUsed, uniqTx)
		} else {
			ac.UniqETHTxsUsed = append(ac.UniqETHTxsUsed, uniqTx)
		}
	case common.BSCNetworkID:
		if md.IncTokenID == common.PRVCoinID {
			ac.UniqPRVEVMTxsUsed = append(ac.UniqPRVEVMTxsUsed, uniqTx)
		} else {
			ac.UniqBSCTxsUsed = append(ac.UniqBSCTxsUsed, uniqTx)
		}
	case common.PLGNetworkID:
		ac.UniqPLGTxsUsed = append(ac.UniqPLGTxsUsed, uniqTx)
	}
	issuingAcceptedInst := metadata.IssuingEVMAcceptedInst{
		ShardID:         receivingShardID,
		IssuingAmount:   amount,
		ReceiverAddrStr: addressStr,
		IncTokenID:      md.IncTokenID,
		TxReqID:         action.TxReqID,
		UniqTx:          uniqTx,
		ExternalTokenID: []byte(common.UnifiedTokenPrefix),
		NetworkID:       md.NetworkID,
		Reward:          reward,
	}
	content, err = json.Marshal(issuingAcceptedInst)
	if err != nil {
		shouldContinue = false
		return
	}
	ac.DBridgeTokenPair[vault.tokenID.String()] = token
	return
}

func (sp *stateProducer) unshield(
	contentStr string, unifiedTokenInfos map[common.Hash]map[uint]*Vault, beaconHeight uint64, shardID byte,
	stateDB *statedb.StateDB,
) (resInsts [][]string, resUnifiedTokenInfos map[common.Hash]map[uint]*Vault, err error) {
	action := metadataCommon.NewAction()
	md := &metadata.BurningRequest{}
	action.Meta = md
	err = action.FromString(contentStr)
	if err != nil {
		err = NewBridgeAggErrorWithValue(OtherError, err)
		return
	}
	var errorType int
	var content []byte
	var burningInst []string
	shouldContinue := true
	defer func() {
		if shouldContinue {
			inst, err := buildInstruction(
				metadataCommon.BurningUnifiedTokenRequestMeta,
				errorType, content, action.TxReqID, shardID, err,
			)
			if err != nil {
				return
			}
			err = nil
			resInsts = append(resInsts, inst)
			if len(burningInst) != 0 {
				resInsts = append(resInsts, burningInst)
			}
		}
	}()

	var prefix, contractAddress string
	var burningMetaType int

	switch md.NetworkID {
	case common.ETHNetworkID:
		prefix = utils.EmptyString
		contractAddress = config.Param().PRVERC20ContractAddressStr
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
		prefix = common.BSCPrefix
		contractAddress = config.Param().PRVBEP20ContractAddressStr
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
		prefix = common.PLGPrefix
		if *md.IsDepositToSC {
			burningMetaType = metadata.BurningPLGForDepositToSCRequestMeta
		} else {
			burningMetaType = metadata.BurningPLGConfirmMeta
		}
		prefix = common.PLGPrefix
	case common.DefaultNetworkID:
		err = errors.New("Cannot get info from default networkID")
		return
	default:
		err = errors.New("Cannot detect networkID")
		return
	}

	var externalTokenID []byte
	if md.TokenID.String() != common.PRVIDStr {
		// Convert to external tokenID
		externalTokenID, err = metadata.FindExternalTokenID(stateDB, md.TokenID, prefix, burningMetaType)
		if err != nil {
			errorType = NotFoundTokenIDInNetworkError
			return
		}
	} else {
		temp := rCommon.HexToAddress(contractAddress)
		externalTokenID = temp[:]
	}

	resUnifiedTokenInfos = unifiedTokenInfos
	vault, err := GetVault(resUnifiedTokenInfos, md.TokenID, md.NetworkID)
	if err != nil {
		errorType = NotFoundNetworkIDError
		return
	}

	actualAmount, err := vault.unshield(md.BurningAmount, md.ExpectedAmount)
	if err != nil {
		Logger.log.Warnf("Calculate unshield amount error: %v tx %s", err, action.TxReqID.String())
		errorType = CalculateUnshieldAmountError
		return
	}
	amount := big.NewInt(0).SetUint64(actualAmount)
	fee := md.BurningAmount - actualAmount
	incTokenID := vault.tokenID
	if md.TokenID == common.PRVCoinID {
		incTokenID = common.PRVCoinID
	}
	if md.TokenID != common.PRVCoinID {
		if bytes.Equal(externalTokenID, append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...)) {
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
	content, err = json.Marshal(acceptedContent)
	if err != nil {
		shouldContinue = false
		return
	}

	h := big.NewInt(0).SetUint64(beaconHeight)
	burningInst = []string{
		strconv.Itoa(burningMetaType),
		strconv.Itoa(int(common.BridgeShardID)),
		base58.Base58Check{}.Encode(externalTokenID, 0x00),
		md.RemoteAddress,
		base58.Base58Check{}.Encode(amount.Bytes(), 0x00),
		action.TxReqID.String(),
		base58.Base58Check{}.Encode(incTokenID[:], 0x00),
		base58.Base58Check{}.Encode(h.Bytes(), 0x00),
	}
	return
}
