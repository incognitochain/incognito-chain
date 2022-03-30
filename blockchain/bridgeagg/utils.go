package bridgeagg

import (
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
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/utils"
)

type ShieldStatus struct {
	Status    byte `json:"Status"`
	ErrorCode uint `json:"ErrorCode,omitempty"`
}

type ModifyListTokenStatus struct {
	Status    byte `json:"Status"`
	ErrorCode uint `json:"ErrorCode,omitempty"`
}

type ConvertStatus struct {
	Status    byte `json:"Status"`
	ErrorCode uint `json:"ErrorCode,omitempty"`
}

type VaultChange struct {
	IsChanged        bool
	IsReserveChanged bool
}

func NewVaultChange() *VaultChange {
	return &VaultChange{}
}

type StateChange struct {
	unifiedTokenID map[common.Hash]bool
	vaultChange    map[common.Hash]map[uint]VaultChange
}

func NewStateChange() *StateChange {
	return &StateChange{
		unifiedTokenID: make(map[common.Hash]bool),
		vaultChange:    make(map[common.Hash]map[uint]VaultChange),
	}
}

func CalculateActualAmount(x, y, deltaX uint64, operator byte) (uint64, error) {
	if operator != SubOperator && operator != AddOperator {
		return 0, errors.New("Cannot recognize operator")
	}
	if deltaX == 0 {
		return 0, errors.New("Cannot process with deltaX = 0")
	}
	if y == 0 {
		return deltaX, nil
	}
	if x == 0 {
		temp := deltaX + y - 1
		if temp < deltaX {
			return 0, errors.New("Out of range uint64")
		}
		return temp, nil
	}
	newX := big.NewInt(0) // x'
	actualAmount := big.NewInt(0)
	switch operator {
	case AddOperator:
		newX.Add(big.NewInt(0).SetUint64(x), big.NewInt(0).SetUint64(deltaX))
		temp := big.NewInt(0).Mul(big.NewInt(0).SetUint64(y), big.NewInt(0).SetUint64(deltaX))
		reward := temp.Div(temp, newX)
		actualAmount = big.NewInt(0).Add(big.NewInt(0).SetUint64(deltaX), reward)
		if actualAmount.Cmp(big.NewInt(0).SetUint64(deltaX)) < 0 {
			return 0, errors.New("actualAmount < deltaX")
		}
	case SubOperator:
		newX.Sub(big.NewInt(0).SetUint64(x), big.NewInt(0).SetUint64(deltaX))
		temp := big.NewInt(0).Mul(big.NewInt(0).SetUint64(y), big.NewInt(0).SetUint64(deltaX))
		fee := temp.Div(temp, newX)
		actualAmount = big.NewInt(0).Sub(big.NewInt(0).SetUint64(deltaX), fee)
		if actualAmount.Cmp(big.NewInt(0).SetUint64(deltaX)) > 0 {
			return 0, errors.New("actualAmount > deltaX")
		}
	default:
		return 0, errors.New("Cannot recognize operator")
	}
	if !actualAmount.IsUint64() {
		return 0, errors.New("Actual amount is not uint64")
	}
	return actualAmount.Uint64(), nil
}

func EstimateActualAmountByBurntAmount(x, y, burntAmount uint64) (uint64, error) {
	if burntAmount == 0 {
		return 0, errors.New("Cannot process with burntAmount = 0")
	}
	if y == 0 {
		if burntAmount > x {
			return 0, fmt.Errorf("BurntAmount %d is > x %d", burntAmount, x)
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

func GetVault(unifiedTokenInfos map[common.Hash]map[uint]*Vault, unifiedTokenID common.Hash, networkID uint) (*Vault, error) {
	if vaults, found := unifiedTokenInfos[unifiedTokenID]; found {
		if vault, found := vaults[networkID]; found {
			return vault, nil
		} else {
			return nil, NewBridgeAggErrorWithValue(NotFoundNetworkIDError, errors.New("Not found networkID"))
		}
	} else {
		return nil, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, errors.New("Not found unifiedTokenID"))
	}
}

func InsertTxHashIssuedByNetworkID(networkID uint, isPRV bool) func(*statedb.StateDB, []byte) error {
	if isPRV {
		return statedb.InsertPRVEVMTxHashIssued
	}
	switch networkID {
	case common.PLGNetworkID:
		return statedb.InsertPLGTxHashIssued
	case common.BSCNetworkID:
		return statedb.InsertBSCTxHashIssued
	case common.ETHNetworkID:
		return statedb.InsertETHTxHashIssued
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

func CloneVaults(
	unifiedTokenInfos map[common.Hash]map[uint]*Vault, unifiedTokenID common.Hash,
) (map[uint]*Vault, error) {
	if vaults, found := unifiedTokenInfos[unifiedTokenID]; found {
		res := make(map[uint]*Vault)
		for networkID, vault := range vaults {
			res[networkID] = vault.Clone()
		}
		return res, nil
	} else {
		return nil, fmt.Errorf("Can't find unifiedTokenID %s", unifiedTokenID.String())
	}
}

func shieldEVM(
	incTokenID common.Hash, networkID uint, ac *metadataCommon.AccumulatedValues,
	shardID byte, txReqID common.Hash,
	vaults map[uint]*Vault, stateDBs map[int]*statedb.StateDB, extraData []byte,
	blockHash rCommon.Hash, txIndex uint,
) (uint64, uint64, byte, []byte, []byte, int, error) {
	var txReceipt *types.Receipt
	err := json.Unmarshal(extraData, &txReceipt)
	if err != nil {
		return 0, 0, 0, nil, nil, OtherError, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestUnmarshalJsonError, err)
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
		if incTokenID == common.PRVCoinID {
			contractAddress = config.Param().PRVERC20ContractAddressStr
			listTxUsed = ac.UniqPRVEVMTxsUsed
			isTxHashIssued = statedb.IsPRVEVMTxHashIssued
		}
	case common.BSCNetworkID:
		listTxUsed = ac.UniqBSCTxsUsed
		contractAddress = config.Param().BscContractAddressStr
		prefix = common.BSCPrefix
		isTxHashIssued = statedb.IsBSCTxHashIssued
		if incTokenID == common.PRVCoinID {
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
		return 0, 0, 0, nil, nil, OtherError, NewBridgeAggErrorWithValue(OtherError, errors.New("Cannot get info from default networkID"))
	default:
		return 0, 0, 0, nil, nil, NotFoundNetworkIDError, NewBridgeAggErrorWithValue(OtherError, errors.New("Cannot detect networkID"))
	}
	vault, found := vaults[networkID]
	if !found {
		return 0, 0, 0, nil, nil, NotFoundNetworkIDError, NewBridgeAggErrorWithValue(NotFoundNetworkIDError, errors.New("Cannot detect networkID"))
	}
	amount, receivingShardID, _, token, uniqTx, err := metadataBridge.ExtractIssueEVMData(
		stateDBs[common.BeaconChainID], shardID, listTxUsed,
		contractAddress, prefix, isTxHashIssued, txReceipt, blockHash, txIndex,
	)
	if err != nil {
		return 0, 0, 0, nil, nil, FailToExtractDataError, NewBridgeAggErrorWithValue(FailToExtractDataError, err)
	}
	err = metadataBridge.VerifyTokenPair(stateDBs, ac, vault.tokenID, token)
	if err != nil {
		Logger.log.Info("[bridgeagg] 4")
		return 0, 0, 0, nil, nil, FailToVerifyTokenPairError, NewBridgeAggErrorWithValue(FailToVerifyTokenPairError, err)
	}

	if incTokenID != common.PRVCoinID {
		tmpAmount := big.NewInt(0).SetUint64(amount)
		tmpAmount.Mul(tmpAmount, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(config.Param().BridgeAggParam.BaseDecimal)), nil))
		tmpAmount.Div(tmpAmount, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(vault.Decimal())), nil))
		if !tmpAmount.IsUint64() {
			return 0, 0, 0, nil, nil, OutOfRangeUni64Error, NewBridgeAggErrorWithValue(OutOfRangeUni64Error, errors.New("Out of range uint64"))
		}
		amount = tmpAmount.Uint64()
	}

	actualAmount, err := vault.shield(amount)
	if err != nil {
		Logger.log.Warnf("Calculate shield amount error: %v tx %s", err, txReqID)
		return 0, 0, 0, nil, nil, CalculateShieldAmountError, NewBridgeAggErrorWithValue(CalculateShieldAmountError, err)
	}
	vaults[networkID] = vault
	reward := actualAmount - amount

	switch networkID {
	case common.ETHNetworkID:
		if incTokenID == common.PRVCoinID {
			ac.UniqPRVEVMTxsUsed = append(ac.UniqPRVEVMTxsUsed, uniqTx)
		} else {
			ac.UniqETHTxsUsed = append(ac.UniqETHTxsUsed, uniqTx)
		}
	case common.BSCNetworkID:
		if incTokenID == common.PRVCoinID {
			ac.UniqPRVEVMTxsUsed = append(ac.UniqPRVEVMTxsUsed, uniqTx)
		} else {
			ac.UniqBSCTxsUsed = append(ac.UniqBSCTxsUsed, uniqTx)
		}
	case common.PLGNetworkID:
		ac.UniqPLGTxsUsed = append(ac.UniqPLGTxsUsed, uniqTx)
	}
	ac.DBridgeTokenPair[vault.tokenID.String()] = token
	return actualAmount, reward, receivingShardID, token, uniqTx, 0, nil
}

func buildAcceptedShieldContents(
	shieldData, rewardData []metadataBridge.AcceptedShieldRequestData,
	paymentAddress privacy.PaymentAddress, incTokenID, txReqID common.Hash, shardID byte,
) ([][]byte, error) {
	contents := [][]byte{}
	acceptedContent := metadataBridge.AcceptedShieldRequest{
		Receiver:   paymentAddress.String(),
		IncTokenID: incTokenID,
		TxReqID:    txReqID,
		ShardID:    shardID,
		Data:       shieldData,
	}
	content, err := json.Marshal(acceptedContent)
	if err != nil {
		return contents, err
	}
	contents = append(contents, content)
	acceptedRewardContent := metadataBridge.AcceptedShieldRequest{
		Receiver:   paymentAddress.String(),
		IncTokenID: incTokenID,
		TxReqID:    txReqID,
		ShardID:    shardID,
		IsReward:   true,
		Data:       rewardData,
	}
	content, err = json.Marshal(acceptedRewardContent)
	if err != nil {
		return contents, err
	}
	contents = append(contents, content)
	return contents, nil
}
