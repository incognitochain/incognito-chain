package bridgeagg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/utils"
)

type ShieldStatusData struct {
	Amount uint64 `json:"Amount"`
	Reward uint64 `json:"Reward"`
}

type ShieldStatus struct {
	Status    byte               `json:"Status"`
	Data      []ShieldStatusData `json:"Data,omitempty"`
	ErrorCode uint               `json:"ErrorCode,omitempty"`
}

type UnshieldStatusData struct {
	ReceivedAmount uint64 `json:"ReceivedAmount"`
	Fee            uint64 `json:"Fee"`
}

type UnshieldStatus struct {
	Status    byte                 `json:"Status"`
	Data      []UnshieldStatusData `json:"Data,omitempty"`
	ErrorCode uint                 `json:"ErrorCode,omitempty"`
}

type ModifyRewardReserveStatus struct {
	Status    byte `json:"Status"`
	ErrorCode uint `json:"ErrorCode,omitempty"`
}

type ConvertStatus struct {
	Status    byte `json:"Status"`
	ErrorCode uint `json:"ErrorCode,omitempty"`
}

type VaultChange struct {
	IsChanged bool
}

func NewVaultChange() *VaultChange {
	return &VaultChange{}
}

type StateChange struct {
	unifiedTokenID map[common.Hash]bool
	vaultChange    map[common.Hash]map[common.Hash]VaultChange
}

func NewStateChange() *StateChange {
	return &StateChange{
		unifiedTokenID: make(map[common.Hash]bool),
		vaultChange:    make(map[common.Hash]map[common.Hash]VaultChange),
	}
}

func CalculateDeltaY(x, y, deltaX uint64, operator byte, isPaused bool) (uint64, error) {
	if operator != SubOperator && operator != AddOperator {
		return 0, errors.New("Cannot recognize operator")
	}
	if deltaX == 0 {
		return 0, errors.New("Cannot process with deltaX = 0")
	}
	if y == 0 || isPaused {
		return 0, nil
	}
	if x == 0 {
		return y - 1, nil
	}
	newX := big.NewInt(0) // x'
	switch operator {
	case AddOperator:
		newX.Add(big.NewInt(0).SetUint64(x), big.NewInt(0).SetUint64(deltaX))
	case SubOperator:
		newX.Sub(big.NewInt(0).SetUint64(x), big.NewInt(0).SetUint64(deltaX))
	default:
		return 0, errors.New("Cannot recognize operator")
	}
	temp := big.NewInt(0).Mul(big.NewInt(0).SetUint64(y), big.NewInt(0).SetUint64(deltaX))
	deltaY := temp.Div(temp, newX)
	if !deltaY.IsUint64() {
		return 0, errors.New("Actual amount is not uint64")
	}
	return deltaY.Uint64(), nil
}

func CalculateShieldActualAmount(x, y, deltaX uint64, isPaused bool) (uint64, error) {
	deltaY, err := CalculateDeltaY(x, y, deltaX, AddOperator, isPaused)
	if err != nil {
		return 0, err
	}
	actualAmount := big.NewInt(0).Add(big.NewInt(0).SetUint64(deltaX), big.NewInt(0).SetUint64(deltaY))
	if actualAmount.Cmp(big.NewInt(0).SetUint64(deltaX)) < 0 {
		return 0, errors.New("actualAmount < deltaX")
	}
	if !actualAmount.IsUint64() {
		return 0, errors.New("Actual amount is not uint64")
	}
	return actualAmount.Uint64(), nil
}

func EstimateActualAmountByBurntAmount(x, y, burntAmount uint64, isPaused bool) (uint64, error) {
	if x == 0 || x == 1 {
		return 0, fmt.Errorf("x is 0 or 1")
	}
	if burntAmount == 0 {
		return 0, errors.New("Cannot process with burntAmount = 0")
	}
	if y == 0 || isPaused {
		if burntAmount > x {
			return 0, fmt.Errorf("BurntAmount %d is > x %d", burntAmount, x)
		}
		if burntAmount == x {
			burntAmount -= 1
		}
		if burntAmount == 0 {
			return 0, fmt.Errorf("Receive actualAmount is 0")
		}
		return burntAmount, nil
	}
	X := big.NewInt(0).SetUint64(x)
	Y := big.NewInt(0).SetUint64(y)
	Z := big.NewInt(0).SetUint64(burntAmount)
	t1 := big.NewInt(0).Add(X, Y)
	t1 = t1.Add(t1, Z)
	t2 := big.NewInt(0).Mul(X, X)
	temp := big.NewInt(0).Sub(Y, Z)
	temp = temp.Mul(temp, X)
	temp = temp.Mul(temp, big.NewInt(2))
	t2 = t2.Add(t2, temp)
	temp = big.NewInt(0).Add(Y, Z)
	temp = temp.Mul(temp, temp)
	t2 = t2.Add(t2, temp)
	t2 = big.NewInt(0).Sqrt(t2)

	A1 := big.NewInt(0).Add(t1, t2)
	A1 = A1.Div(A1, big.NewInt(2))
	A2 := big.NewInt(0).Sub(t1, t2)
	A2 = A2.Div(A2, big.NewInt(2))
	var a1, a2 uint64

	if A1.IsUint64() {
		a1 = A1.Uint64()
	}
	if A2.IsUint64() {
		a2 = A2.Uint64()
	}
	if a1 > burntAmount {
		a1 = 0
	}
	if a2 > burntAmount {
		a2 = 0
	}
	if a1 == 0 && a2 == 0 {
		return 0, fmt.Errorf("x %d y %d z %d cannot find solutions", x, y, burntAmount)
	}
	a := a1
	if a < a2 {
		a = a2
	}
	if a > x {
		return 0, fmt.Errorf("a %d is > x %d", a, x)
	}

	return a, nil
}

func GetVault(unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault, unifiedTokenID, incTokenID common.Hash) (*Vault, error) {
	if vaults, found := unifiedTokenInfos[unifiedTokenID]; found {
		if vault, found := vaults[incTokenID]; found {
			return vault, nil
		} else {
			return nil, NewBridgeAggErrorWithValue(NotFoundNetworkIDError, errors.New("Not found networkID"))
		}
	} else {
		return nil, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, errors.New("Not found unifiedTokenID"))
	}
}

func InsertTxHashIssuedByNetworkID(networkID uint) func(*statedb.StateDB, []byte) error {
	switch networkID {
	case common.PLGNetworkID:
		return statedb.InsertPLGTxHashIssued
	case common.BSCNetworkID:
		return statedb.InsertBSCTxHashIssued
	case common.ETHNetworkID:
		return statedb.InsertETHTxHashIssued
	case common.FTMNetworkID:
		return statedb.InsertFTMTxHashIssued
	}
	return nil
}

func buildInstruction(
	metaType int, errorType int,
	contents [][]byte, txReqID common.Hash,
	shardID byte, err error,
) ([][]string, error) {
	res := [][]string{}
	for _, content := range contents {
		inst := metadataCommon.NewInstructionWithValue(
			metaType,
			common.AcceptedStatusStr,
			shardID,
			utils.EmptyString,
		)
		if err != nil {
			rejectContent := metadataCommon.NewRejectContentWithValue(txReqID, ErrCodeMessage[errorType].Code, content)
			inst.Status = common.RejectedStatusStr
			rejectedInst := []string{}
			rejectedInst, err = inst.StringSliceWithRejectContent(rejectContent)
			if err != nil {
				return res, NewBridgeAggErrorWithValue(errorType, err)
			}
			res = append(res, rejectedInst)
		} else {
			inst.Content = base64.StdEncoding.EncodeToString(content)
			res = append(res, inst.StringSlice())
		}
	}
	return res, nil
}

func CloneVaultsByUnifiedTokenID(
	unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault, unifiedTokenID common.Hash,
) (map[common.Hash]*Vault, error) {
	if vaults, found := unifiedTokenInfos[unifiedTokenID]; found {
		res := make(map[common.Hash]*Vault)
		for networkID, vault := range vaults {
			res[networkID] = vault.Clone()
		}
		return res, nil
	} else {
		return nil, fmt.Errorf("Can't find unifiedTokenID %s", unifiedTokenID.String())
	}
}

func shieldEVM(
	unifiedTokenID, incTokenID common.Hash, networkID uint, ac *metadata.AccumulatedValues,
	shardID byte, txReqID common.Hash,
	vault *Vault, stateDBs map[int]*statedb.StateDB, extraData []byte,
	blockHash rCommon.Hash, txIndex uint, currentPaymentAddress string,
) (*Vault, uint64, uint64, byte, []byte, []byte, string, *metadata.AccumulatedValues, int, error) {
	var txReceipt *types.Receipt
	err := json.Unmarshal(extraData, &txReceipt)
	if err != nil {
		return nil, 0, 0, 0, nil, nil, "", ac, OtherError, NewBridgeAggErrorWithValue(OtherError, err)
	}
	var listTxUsed [][]byte
	var contractAddress, prefix string
	var isTxHashIssued func(stateDB *statedb.StateDB, uniqueEthTx []byte) (bool, error)

	switch networkID {
	case common.ETHNetworkID:
		listTxUsed = ac.UniqETHTxsUsed
		contractAddress = config.Param().EthContractAddressStr
		prefix = utils.EmptyString
		isTxHashIssued = statedb.IsETHTxHashIssued
	case common.BSCNetworkID:
		listTxUsed = ac.UniqBSCTxsUsed
		contractAddress = config.Param().BscContractAddressStr
		prefix = common.BSCPrefix
		isTxHashIssued = statedb.IsBSCTxHashIssued
	case common.PLGNetworkID:
		listTxUsed = ac.UniqPLGTxsUsed
		contractAddress = config.Param().PlgContractAddressStr
		prefix = common.PLGPrefix
		isTxHashIssued = statedb.IsPLGTxHashIssued
	case common.FTMNetworkID:
		listTxUsed = ac.UniqFTMTxsUsed
		contractAddress = config.Param().FtmContractAddressStr
		prefix = common.FTMPrefix
		isTxHashIssued = statedb.IsFTMTxHashIssued
	case common.DefaultNetworkID:
		return nil, 0, 0, 0, nil, nil, "", ac, OtherError, NewBridgeAggErrorWithValue(OtherError, errors.New("Cannot get info from default networkID"))
	default:
		return nil, 0, 0, 0, nil, nil, "", ac, NotFoundNetworkIDError, NewBridgeAggErrorWithValue(OtherError, errors.New("Cannot detect networkID"))
	}

	amount, receivingShardID, paymentAddress, token, uniqTx, err := metadataBridge.ExtractIssueEVMData(
		stateDBs[common.BeaconChainID], shardID, listTxUsed,
		contractAddress, prefix, isTxHashIssued, txReceipt, blockHash, txIndex,
	)
	if err != nil {
		return nil, 0, 0, 0, nil, nil, "", ac, FailToExtractDataError, NewBridgeAggErrorWithValue(FailToExtractDataError, err)
	}
	if currentPaymentAddress != "" && currentPaymentAddress != paymentAddress {
		return nil, 0, 0, 0, nil, nil, "", ac, FailToExtractDataError, NewBridgeAggErrorWithValue(FailToExtractDataError, errors.New("PaymentAddress from proofs need to be similar"))
	}
	err = metadataBridge.VerifyTokenPair(stateDBs, ac, incTokenID, token)
	if err != nil {
		return nil, 0, 0, 0, nil, nil, "", ac, FailToVerifyTokenPairError, NewBridgeAggErrorWithValue(FailToVerifyTokenPairError, err)
	}
	decimal := vault.ExtDecimal()
	if !bytes.Equal(append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), token) {
		if decimal > config.Param().BridgeAggParam.BaseDecimal {
			decimal = config.Param().BridgeAggParam.BaseDecimal
		}
	}
	tmpAmount, err := CalculateAmountByDecimal(*amount, decimal, AddOperator)
	if err != nil {
		return nil, 0, 0, 0, nil, nil, "", ac, OutOfRangeUni64Error, NewBridgeAggErrorWithValue(OutOfRangeUni64Error, err)
	}
	//tmpAmount is uint64 after this function

	actualAmount, err := vault.shield(tmpAmount.Uint64())
	if err != nil {
		Logger.log.Warnf("Calculate shield amount error: %v tx %s", err, txReqID)
		return nil, 0, 0, 0, nil, nil, "", ac, CalculateShieldAmountError, NewBridgeAggErrorWithValue(CalculateShieldAmountError, err)
	}
	reward := actualAmount - tmpAmount.Uint64()

	switch networkID {
	case common.ETHNetworkID:
		ac.UniqETHTxsUsed = append(ac.UniqETHTxsUsed, uniqTx)
	case common.BSCNetworkID:
		ac.UniqBSCTxsUsed = append(ac.UniqBSCTxsUsed, uniqTx)
	case common.PLGNetworkID:
		ac.UniqPLGTxsUsed = append(ac.UniqPLGTxsUsed, uniqTx)
	case common.FTMNetworkID:
		ac.UniqFTMTxsUsed = append(ac.UniqFTMTxsUsed, uniqTx)
	}
	ac.DBridgeTokenPair[unifiedTokenID.String()] = GetExternalTokenIDForUnifiedToken()
	return vault, actualAmount, reward, receivingShardID, token, uniqTx, paymentAddress, ac, 0, nil
}

func unshieldEVM(
	data metadataBridge.UnshieldRequestData, stateDB *statedb.StateDB, vault *Vault, txReqID common.Hash, isDepositToSC bool,
) (*Vault, []byte, *big.Int, uint64, uint64, int, int, error) {
	var prefix string
	var burningMetaType int

	switch vault.NetworkID() {
	case common.ETHNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningConfirmForDepositToSCMetaV2
		} else {
			burningMetaType = metadata.BurningConfirmMetaV2
		}
		prefix = utils.EmptyString
	case common.BSCNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningPBSCConfirmForDepositToSCMeta
		} else {
			burningMetaType = metadata.BurningBSCConfirmMeta
		}
		prefix = common.BSCPrefix
	case common.PLGNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningPLGConfirmForDepositToSCMeta
		} else {
			burningMetaType = metadata.BurningPLGConfirmMeta
		}
		prefix = common.PLGPrefix
	case common.FTMNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningFantomConfirmForDepositToSCMeta
		} else {
			burningMetaType = metadata.BurningFantomConfirmMeta
		}
		prefix = common.FTMPrefix
	case common.DefaultNetworkID:
		return nil, nil, nil, 0, 0, burningMetaType, OtherError, NewBridgeAggErrorWithValue(OtherError, errors.New("Cannot get info from default networkID"))
	default:
		return nil, nil, nil, 0, 0, burningMetaType, OtherError, NewBridgeAggErrorWithValue(OtherError, errors.New("Cannot detect networkID"))
	}

	// Convert to external tokenID
	externalTokenID, err := metadataBridge.FindExternalTokenID(stateDB, data.IncTokenID, prefix, burningMetaType)
	if err != nil {
		return nil, nil, nil, 0, 0, burningMetaType, NotFoundTokenIDInNetworkError, NewBridgeAggErrorWithValue(NotFoundNetworkIDError, err)
	}

	actualAmount, err := vault.unshield(data.BurningAmount, data.MinExpectedAmount)
	if err != nil {
		Logger.log.Warnf("Calculate unshield amount error: %v tx %s", err, txReqID.String())
		return nil, nil, nil, 0, 0, burningMetaType, CalculateUnshieldAmountError, NewBridgeAggErrorWithValue(CalculateUnshieldAmountError, err)
	}
	fee := data.BurningAmount - actualAmount
	decimal := vault.ExtDecimal()
	if !bytes.Equal(append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), externalTokenID) {
		if decimal > config.Param().BridgeAggParam.BaseDecimal {
			decimal = config.Param().BridgeAggParam.BaseDecimal
		}
	}
	unshieldAmount, err := CalculateAmountByDecimal(*big.NewInt(0).SetUint64(actualAmount), decimal, SubOperator)
	if err != nil {
		return nil, nil, nil, 0, 0, burningMetaType, OtherError, NewBridgeAggErrorWithValue(OtherError, err)
	}
	if unshieldAmount.Cmp(big.NewInt(0)) == 0 {
		return nil, nil, nil, 0, 0, burningMetaType, CalculateUnshieldAmountError, NewBridgeAggErrorWithValue(OtherError, errors.New("Cannot received unshield amount equal to 0"))
	}

	return vault, externalTokenID, unshieldAmount, actualAmount, fee, burningMetaType, 0, nil
}

func buildAcceptedShieldContents(
	shieldData, rewardData []metadataBridge.AcceptedShieldRequestData,
	paymentAddress privacy.PaymentAddress, unifiedTokenID, txReqID common.Hash, shardID byte,
	shouldBuildRewardContent bool,
) ([][]byte, error) {
	contents := [][]byte{}
	acceptedContent := metadataBridge.AcceptedShieldRequest{
		Receiver:       paymentAddress,
		UnifiedTokenID: unifiedTokenID,
		TxReqID:        txReqID,
		ShardID:        shardID,
		Data:           shieldData,
	}
	content, err := json.Marshal(acceptedContent)
	if err != nil {
		return contents, err
	}
	contents = append(contents, content)
	if shouldBuildRewardContent {
		acceptedRewardContent := metadataBridge.AcceptedShieldRequest{
			Receiver:       paymentAddress,
			UnifiedTokenID: unifiedTokenID,
			TxReqID:        txReqID,
			ShardID:        shardID,
			IsReward:       true,
			Data:           rewardData,
		}
		content, err = json.Marshal(acceptedRewardContent)
		if err != nil {
			return contents, err
		}
		contents = append(contents, content)
	}
	return contents, nil
}

func CalculateAmountByDecimal(amount big.Int, decimal uint, operator byte) (*big.Int, error) {
	res := big.NewInt(0).Set(&amount)
	switch operator {
	case AddOperator:
		res.Mul(res, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(config.Param().BridgeAggParam.BaseDecimal)), nil))
		res.Div(res, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil))
		if !res.IsUint64() {
			return nil, errors.New("Out of range unit64")
		}
	case SubOperator:
		res.Mul(res, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil))
		res.Div(res, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(config.Param().BridgeAggParam.BaseDecimal)), nil))
	default:
		return nil, errors.New("Cannot recognize operator")
	}

	return res, nil
}

func getBurningConfirmMeta(networkID int, isDepositToSC bool) (int, string, error) {
	var burningMetaType int
	var prefix string

	switch networkID {
	case common.ETHNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningConfirmForDepositToSCMetaV2
		} else {
			burningMetaType = metadata.BurningConfirmMetaV2
		}
		prefix = utils.EmptyString
	case common.BSCNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningPBSCConfirmForDepositToSCMeta
		} else {
			burningMetaType = metadata.BurningBSCConfirmMeta
		}
		prefix = common.BSCPrefix
	case common.PLGNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningPLGConfirmForDepositToSCMeta
		} else {
			burningMetaType = metadata.BurningPLGConfirmMeta
		}
		prefix = common.PLGPrefix
	case common.FTMNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningFantomConfirmForDepositToSCMeta
		} else {
			burningMetaType = metadata.BurningFantomConfirmMeta
		}
		prefix = common.FTMPrefix
	case common.DefaultNetworkID:
		return burningMetaType, prefix, NewBridgeAggErrorWithValue(OtherError, errors.New("Cannot get info from default networkID"))
	default:
		return burningMetaType, prefix, NewBridgeAggErrorWithValue(OtherError, errors.New("Cannot detect networkID"))
	}

	return burningMetaType, prefix, nil
}

type preProcessedUnshieldData struct {
	BurningAmount     uint64
	ReceivedAmount    uint64
	ExtReceivedAmount *big.Int
	Fee               uint64
	TokenID           common.Hash
	ExternalTokenID   common.Hash
}

// func preprocessUnshield(
// 	unifiedTokenID common.Hash,
// 	unshieldMeta metadataBridge.UnshieldRequest,
// 	vaults map[common.Hash]*Vault,
// 	percentFee float64,
// ) (
// 	acceptedData []preProcessedUnshieldData,
// 	isAddWaitingList bool,
// 	errorType int,
// 	err error,
// ) {
// 	isAddWaitingList = false

// 	// check vaults are enough for unshield req or add to waiting list
// 	for _, data := range unshieldMeta.Data {
// 		// check IncTokenID
// 		vault, found := vaults[data.IncTokenID]
// 		if !found {
// 			errorType = NotFoundTokenIDInNetworkError
// 			err = fmt.Errorf("Not found tokenID %s in unified tokenID %s", data.IncTokenID.String(), unifiedTokenID)
// 			return
// 		}

// 		// calculate unshield fee
// 		receivedAmount, fee, e := vault.calUnshieldFee(data.BurningAmount, data.MinExpectedAmount, percentFee)
// 		if e != nil {
// 			errorType = NotFoundTokenIDInNetworkError
// 			err = e
// 			return
// 		}

// 		// add to waiting list
// 		if fee > 0 {
// 			isAddWaitingList = true
// 		}

// 		acceptedData = append(
// 			acceptedData,
// 			preProcessedUnshieldData{
// 				BurningAmount:  data.BurningAmount,
// 				ReceivedAmount: receivedAmount,
// 				Fee:            fee,
// 				TokenID:        data.IncTokenID,
// 			})

// 		// // get networkType by NetworkID
// 		// networkType, _ := metadataBridge.GetNetworkTypeByNetworkID(vault.NetworkID())
// 		// switch networkType {
// 		// case common.EVMNetworkType:

// 		// // calculate infos for creating burning confirm instruction if don't add to waiting list
// 		// if !isAddWaitingList {
// 		// 	tempVault, externalTokenID, unshieldAmount, amount, fee, burningMetaType, et, e := unshieldEVM(data, stateDB, vault, action.TxReqID, md.IsDepositToSC)
// 		// 	if e != nil {
// 		// 		errorType = et
// 		// 		err = e
// 		// 		burningInsts = [][]string{}
// 		// 		return
// 		// 	}
// 		// 	if amount == 0 {
// 		// 		errorType = CalculateUnshieldAmountError
// 		// 		err = errors.New("Actual receive amount is 0 with unshielding request")
// 		// 		burningInsts = [][]string{}
// 		// 		return
// 		// 	}
// 		// 	vault = tempVault
// 		// 	h := big.NewInt(0).SetUint64(beaconHeight)
// 		// 	newTxReqID := common.HashH(append(action.TxReqID.Bytes(), common.IntToBytes(index)...))

// 		// }

// 		// burningInst := []string{
// 		// 	strconv.Itoa(burningMetaType),
// 		// 	strconv.Itoa(int(common.BridgeShardID)),
// 		// 	base58.Base58Check{}.Encode(externalTokenID, 0x00),
// 		// 	data.RemoteAddress,
// 		// 	base58.Base58Check{}.Encode(unshieldAmount.Bytes(), 0x00),
// 		// 	newTxReqID.String(),
// 		// 	base58.Base58Check{}.Encode(data.IncTokenID[:], 0x00),
// 		// 	base58.Base58Check{}.Encode(h.Bytes(), 0x00),
// 		// }
// 		// burningInsts = append(burningInsts, burningInst)
// 		// acceptedUnshieldRequestData := metadataBridge.UnshieldRequestData{
// 		// 	MinExpectedAmount: fee,
// 		// 	BurningAmount:     amount,
// 		// 	IncTokenID:        data.IncTokenID,
// 		// 	RemoteAddress:     data.RemoteAddress,
// 		// }
// 		// acceptedUnshieldRequestDatas = append(acceptedUnshieldRequestDatas, acceptedUnshieldRequestData)
// 		// default:
// 		// 	errorType = NotFoundNetworkIDError
// 		// 	err = errors.New("Not found networkID")
// 		// 	return
// 		// }
// 	}

// 	if !isAddWaitingList {
// 		for i, data := range unshieldMeta.Data {
// 			// check IncTokenID
// 			vault, _ := vaults[data.IncTokenID]

// 			tempVault, externalTokenID, unshieldAmount, amount, fee, burningMetaType, et, e := calExtReceivedAmtForUnshield(data, stateDB, vault, action.TxReqID, md.IsDepositToSC)
// 			if e != nil {
// 				errorType = et
// 				err = e
// 				burningInsts = [][]string{}
// 				return
// 			}
// 			if amount == 0 {
// 				errorType = CalculateUnshieldAmountError
// 				err = errors.New("Actual receive amount is 0 with unshielding request")
// 				burningInsts = [][]string{}
// 				return
// 			}
// 			vault = tempVault
// 			h := big.NewInt(0).SetUint64(beaconHeight)
// 			newTxReqID := common.HashH(append(action.TxReqID.Bytes(), common.IntToBytes(index)...))

// 		}
// 	}

// 	return

// }

// func calExtReceivedAmtForUnshield(
// 	data metadataBridge.UnshieldRequestData,
// 	stateDB *statedb.StateDB,
// 	vault *Vault,
// 	txReqID common.Hash,
// 	isDepositToSC bool,
// ) (*Vault, []byte, *big.Int, uint64, uint64, int, int, error) {
// 	var prefix string
// 	burningMetaType, prefix, err := getBurningConfirmMeta(int(vault.NetworkID()), isDepositToSC)
// 	if err != nil {
// 		return nil, nil, nil, 0, 0, burningMetaType, OtherError, NewBridgeAggErrorWithValue(OtherError, err)
// 	}

// 	// Convert to external tokenID
// 	externalTokenID, err := metadataBridge.FindExternalTokenID(stateDB, data.IncTokenID, prefix, burningMetaType)
// 	if err != nil {
// 		return nil, nil, nil, 0, 0, burningMetaType, NotFoundTokenIDInNetworkError, NewBridgeAggErrorWithValue(NotFoundNetworkIDError, err)
// 	}

// 	extDecimal := vault.ExtDecimal()
// 	if !bytes.Equal(append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), externalTokenID) {
// 		if extDecimal > config.Param().BridgeAggParam.BaseDecimal {
// 			extDecimal = config.Param().BridgeAggParam.BaseDecimal
// 		}
// 	}
// 	extReceivedAmount, err := CalculateAmountByDecimal(*big.NewInt(0).SetUint64(receivedAmt), extDecimal, SubOperator)
// 	if err != nil {
// 		return nil, nil, nil, 0, 0, burningMetaType, OtherError, NewBridgeAggErrorWithValue(OtherError, err)
// 	}
// 	if extReceivedAmount.Cmp(big.NewInt(0)) == 0 {
// 		return nil, nil, nil, 0, 0, burningMetaType, CalculateUnshieldAmountError, NewBridgeAggErrorWithValue(OtherError, errors.New("Cannot received unshield amount equal to 0"))
// 	}

// 	return vault, externalTokenID, extReceivedAmount, receivedAmt, fee, burningMetaType, 0, nil
// }

func CalculateIncDecimal(decimal, baseDecimal uint) uint {
	if decimal > baseDecimal {
		return baseDecimal
	}
	return decimal
}

func validateConfigVault(sDBs map[int]*statedb.StateDB, tokenID common.Hash, vault config.Vault) error {
	networkID := vault.NetworkID
	if networkID != common.BSCNetworkID && networkID != common.ETHNetworkID && networkID != common.PLGNetworkID && networkID != common.FTMNetworkID {
		return fmt.Errorf("Cannot find networkID %d", networkID)
	}
	if vault.ExternalDecimal == 0 {
		return fmt.Errorf("ExternalTokenID cannot be 0")
	}
	if vault.ExternalTokenID == utils.EmptyString {
		return fmt.Errorf("ExternalTokenID can not be empty")
	}
	if tokenID == common.PRVCoinID || tokenID == common.PDEXCoinID {
		return fmt.Errorf("incTokenID is prv or pdex")
	}
	bridgeTokenInfoIndex, externalTokenIDIndex, err := GetBridgeTokenIndex(sDBs[common.BeaconChainID])
	if err != nil {
		return err
	}
	externalTokenID, err := getExternalTokenIDByNetworkID(vault.ExternalTokenID, networkID)
	if err != nil {
		return err
	}
	if bridgeTokenInfoState, found := bridgeTokenInfoIndex[tokenID]; found {
		if !bytes.Equal(bridgeTokenInfoState.ExternalTokenID, externalTokenID) {
			return errors.New("ExternalTokenID is not valid with data from db")
		}
	} else {
		encodedExternalTokenID := base64.StdEncoding.EncodeToString(externalTokenID)
		if externalTokenIDIndex[encodedExternalTokenID] {
			return errors.New("ExternalTokenID has existed")
		}
		isExisted, err := statedb.CheckTokenIDExisted(sDBs, tokenID)
		if err != nil {
			return fmt.Errorf("WARNING: Error in finding tokenID %s", tokenID.String())
		}
		if isExisted {
			return fmt.Errorf("WARNING: tokenID %s has existed", tokenID.String())
		}
	}
	return nil
}

func getExternalTokenIDByNetworkID(externalTokenID string, networkID uint) ([]byte, error) {
	var res []byte
	var prefix string
	switch networkID {
	case common.ETHNetworkID:
		prefix = utils.EmptyString
	case common.BSCNetworkID:
		prefix = common.BSCPrefix
	case common.PLGNetworkID:
		prefix = common.PLGPrefix
	case common.FTMNetworkID:
		prefix = common.FTMPrefix
	default:
		return nil, fmt.Errorf("Invalid networkID %v", networkID)
	}
	networkType, err := metadataBridge.GetNetworkTypeByNetworkID(networkID)
	if err != nil {
		return nil, err
	}
	switch networkType {
	case common.EVMNetworkType:
		tokenAddr := rCommon.HexToAddress(externalTokenID)
		res = append([]byte(prefix), tokenAddr.Bytes()...)
	}
	return res, nil
}

func CloneUnifiedTokenInfos(unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault) map[common.Hash]map[common.Hash]*Vault {
	res := make(map[common.Hash]map[common.Hash]*Vault)
	for unifiedTokenID, vaults := range unifiedTokenInfos {
		res[unifiedTokenID] = make(map[common.Hash]*Vault)
		for tokenID, vault := range vaults {
			res[unifiedTokenID][tokenID] = vault.Clone()
		}
	}
	return res
}

func updateRewardReserve(lastUpdatedRewardReserve, currentRewardReserve, newRewardReserve uint64) (uint64, uint64, error) {
	if lastUpdatedRewardReserve == currentRewardReserve && lastUpdatedRewardReserve == newRewardReserve && newRewardReserve == 0 {
		return 0, 0, nil
	}
	var resLastUpdatedRewardReserve uint64
	tmp := big.NewInt(0).Sub(big.NewInt(0).SetUint64(lastUpdatedRewardReserve), big.NewInt(0).SetUint64(currentRewardReserve))
	if tmp.Cmp(big.NewInt(0).SetUint64(newRewardReserve)) >= 0 {
		return 0, 0, errors.New("deltaY is >= newRewardReserve")
	}

	resLastUpdatedRewardReserve = newRewardReserve
	tmpRewardReserve := big.NewInt(0).Sub(big.NewInt(0).SetUint64(newRewardReserve), tmp)
	if !tmpRewardReserve.IsUint64() {
		return 0, 0, errors.New("Out of range uint64")
	}
	return resLastUpdatedRewardReserve, tmpRewardReserve.Uint64(), nil
}

func GetExternalTokenIDForUnifiedToken() []byte {
	return []byte(common.UnifiedTokenPrefix)
}

func getPrefixByNetworkID(networkID uint) (string, error) {
	var prefix string
	switch networkID {
	case common.ETHNetworkID:
		prefix = utils.EmptyString
	case common.BSCNetworkID:
		prefix = common.BSCPrefix
	case common.PLGNetworkID:
		prefix = common.PLGPrefix
	case common.FTMNetworkID:
		prefix = common.FTMPrefix
	default:
		return utils.EmptyString, errors.New("Invalid networkID")
	}
	return prefix, nil
}

func CalculateReceivedAmount(amount uint64, tokenID common.Hash, decimal uint, networkID uint, sDB *statedb.StateDB) (uint64, error) {
	prefix, err := getPrefixByNetworkID(networkID)
	if err != nil {
		return 0, err
	}
	externalTokenID, err := GetExternalTokenIDByIncTokenID(tokenID, sDB)
	if err != nil {
		return 0, err
	}

	if !bytes.Equal(append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), externalTokenID) {
		if decimal > config.Param().BridgeAggParam.BaseDecimal {
			decimal = config.Param().BridgeAggParam.BaseDecimal
		}
	}
	unshieldAmount, err := CalculateAmountByDecimal(*big.NewInt(0).SetUint64(amount), decimal, SubOperator)
	if err != nil {
		return 0, err
	}
	if unshieldAmount.Cmp(big.NewInt(0)) == 0 {
		return 0, errors.New("Received amount is 0")
	}
	return unshieldAmount.Uint64(), nil
}

func CalculateMaxReceivedAmount(x, y uint64) (uint64, error) {
	if x <= 1 {
		return 0, nil
	}
	return x - 1, nil
}

func decreaseVaultAmount(v *statedb.BridgeAggVaultState, amount uint64) (*statedb.BridgeAggVaultState, error) {
	temp := v.Amount() - amount
	if temp > v.Amount() {
		return nil, errors.New("decrease out of range uint64")
	}
	v.SetAmount(temp)
	return v, nil
}

func increaseVaultAmount(v *statedb.BridgeAggVaultState, amount uint64) (*statedb.BridgeAggVaultState, error) {
	temp := v.Amount() + amount
	if temp < v.Amount() {
		return nil, errors.New("increase out of range uint64")
	}
	v.SetAmount(temp)
	return v, nil
}

func convert(v *statedb.BridgeAggVaultState, amount uint64) (*statedb.BridgeAggVaultState, uint64, error) {
	decimal := CalculateIncDecimal(v.ExtDecimal(), config.Param().BridgeAggParam.BaseDecimal)
	tmpAmount, err := CalculateAmountByDecimal(*big.NewInt(0).SetUint64(amount), decimal, AddOperator)
	if err != nil {
		return nil, 0, err
	}
	if tmpAmount.Cmp(big.NewInt(0)) == 0 {
		return nil, 0, fmt.Errorf("amount %d is not enough for converting", amount)
	}
	v, err = increaseVaultAmount(v, tmpAmount.Uint64())
	if err != nil {
		return nil, 0, err
	}
	return v, tmpAmount.Uint64(), nil
}

func shield(v *statedb.BridgeAggVaultState, amount uint64) (*statedb.BridgeAggVaultState, uint64, error) {
	v, err := increaseVaultAmount(v, amount)
	if err != nil {
		return v, 0, err
	}
	return v, amount, nil
}

func unshield(v *statedb.BridgeAggVaultState, amount, expectedAmount uint64) (*statedb.BridgeAggVaultState, uint64, error) {
	// actualAmount, err := EstimateActualAmountByBurntAmount(v.Reserve(), v.CurrentRewardReserve(), amount, v.IsPaused())
	// if err != nil {
	// 	return 0, err
	// }
	// if actualAmount < expectedAmount {
	// 	return 0, fmt.Errorf("actual amount %v < expected amount %v", actualAmount, expectedAmount)
	// }
	// err = v.increaseCurrentRewardReserve(amount - actualAmount)
	// if err != nil {
	// 	return 0, err
	// }
	v, err := decreaseVaultAmount(v, amount)
	if err != nil {
		return nil, 0, err
	}
	return v, expectedAmount, nil
}

// calculate actual received amount and actual fee
func calUnshieldFee(v *statedb.BridgeAggVaultState, burnAmount, minExpectedAmount uint64, percentFee float64) (*statedb.BridgeAggVaultState, uint64, uint64, error) {
	// vault has enough amount
	if v.Amount() >= burnAmount {
		return nil, burnAmount, 0, nil
	}

	// vault not enough amount
	// fee = shortageAmount * percentFee / 100
	shortageAmount := burnAmount - v.Amount()

	feeTmp, _ := new(big.Float).Mul(
		new(big.Float).SetFloat64(percentFee),
		new(big.Float).SetFloat64(float64(shortageAmount)),
	).Uint64()
	fee := feeTmp / 100
	if fee == 0 {
		fee = 1 // at least 1
	}
	if fee > burnAmount {
		return nil, 0, 0, fmt.Errorf("Needed fee %v larger than burn amount %v", fee, burnAmount)
	}

	// calculate actual received amount
	actualReceivedAmt := burnAmount - fee
	if actualReceivedAmt < minExpectedAmount {
		return nil, 0, 0, fmt.Errorf("Actual received amount %v less than min expected amount %v", actualReceivedAmt, minExpectedAmount)
	}

	return v, actualReceivedAmt, fee, nil
}
