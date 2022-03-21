package bridgeagg

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
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
	k := big.NewInt(0).Mul(big.NewInt(0).SetUint64(x), big.NewInt(0).SetUint64(y))
	newX := big.NewInt(0) // x'
	actualAmount := big.NewInt(0)
	switch operator {
	case AddOperator:
		newX.Add(big.NewInt(0).SetUint64(x), big.NewInt(0).SetUint64(deltaX))
		newY := big.NewInt(0).Div(k, newX) // y'
		reward := big.NewInt(0).Sub(big.NewInt(0).SetUint64(y), newY)
		actualAmount = big.NewInt(0).Add(big.NewInt(0).SetUint64(deltaX), reward)
		if actualAmount.Cmp(big.NewInt(0).SetUint64(deltaX)) < 0 {
			return 0, errors.New("actualAmount < deltaX")
		}
	case SubOperator:
		newX.Sub(big.NewInt(0).SetUint64(x), big.NewInt(0).SetUint64(deltaX))
		newY := big.NewInt(0).Div(k, newX) // y'
		fee := big.NewInt(0).Sub(newY, big.NewInt(0).SetUint64(y))
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

type ShieldAction struct {
	Content []string
	UniqTx  []byte
}

type UnshieldAction struct {
	Content         metadata.BurningRequest
	TxReqID         common.Hash
	ExternalTokenID []byte
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
