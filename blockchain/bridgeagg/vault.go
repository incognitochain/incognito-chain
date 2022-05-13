package bridgeagg

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type Vault struct {
	statedb.BridgeAggVaultState
}

func NewVault() *Vault {
	return &Vault{}
}

func NewVaultWithValue(state statedb.BridgeAggVaultState) *Vault {
	return &Vault{
		BridgeAggVaultState: state,
	}
}

func (v *Vault) Clone() *Vault {
	res := &Vault{
		BridgeAggVaultState: *v.BridgeAggVaultState.Clone(),
	}
	return res
}

func (v *Vault) GetDiff(compareVault *Vault) (*Vault, *VaultChange, error) {
	vaultChange := NewVaultChange()
	if compareVault == nil {
		return nil, nil, errors.New("Compare vault is nul")
	}
	difVaultState, err := v.BridgeAggVaultState.GetDiff(&compareVault.BridgeAggVaultState)
	if err != nil {
		return nil, nil, err
	}
	if difVaultState != nil {
		vaultChange.IsChanged = true
	}
	if vaultChange.IsChanged {
		return v.Clone(), vaultChange, nil
	}
	return nil, nil, nil
}

func (v *Vault) decreaseAmount(amount uint64) error {
	temp := v.Amount() - amount
	if temp > v.Amount() {
		return errors.New("decrease out of range uint64")
	}
	v.SetAmount(temp)
	return nil
}

func (v *Vault) increaseAmount(amount uint64) error {
	temp := v.Amount() + amount
	if temp < v.Amount() {
		return errors.New("increase out of range uint64")
	}
	v.SetAmount(temp)
	return nil
}

func (v *Vault) convert(amount uint64) (uint64, error) {
	decimal := CalculateIncDecimal(v.ExtDecimal(), config.Param().BridgeAggParam.BaseDecimal)
	tmpAmount, err := CalculateAmountByDecimal(*big.NewInt(0).SetUint64(amount), decimal, AddOperator)
	if err != nil {
		return 0, err
	}
	if tmpAmount.Cmp(big.NewInt(0)) == 0 {
		return 0, fmt.Errorf("amount %d is not enough for converting", amount)
	}
	return tmpAmount.Uint64(), v.increaseAmount(tmpAmount.Uint64())
}

func (v *Vault) shield(amount uint64) (uint64, error) {
	// actualAmount, err := CalculateShieldActualAmount(v.Amount(), v.CurrentRewardReserve(), amount, v.IsPaused())
	// if err != nil {
	// 	return 0, err
	// }
	err := v.increaseAmount(amount)
	if err != nil {
		return 0, err
	}
	return amount, nil
}

func (v *Vault) unshield(amount, expectedAmount uint64) (uint64, error) {
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
	err := v.decreaseAmount(amount)
	if err != nil {
		return 0, err
	}
	return expectedAmount, nil
}

// calculate actual received amount and actual fee
func (v *Vault) calUnshieldFee(burnAmount, minExpectedAmount uint64, percentFee float64) (uint64, uint64, error) {
	// vault has enough amount
	if v.Amount() >= burnAmount {
		return burnAmount, 0, nil
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
		return 0, 0, fmt.Errorf("Needed fee %v larger than burn amount %v", fee, burnAmount)
	}

	// calculate actual received amount
	actualReceivedAmt := burnAmount - fee
	if actualReceivedAmt < minExpectedAmount {
		return 0, 0, fmt.Errorf("Actual received amount %v less than min expected amount %v", actualReceivedAmt, minExpectedAmount)
	}

	return actualReceivedAmt, fee, nil
}
