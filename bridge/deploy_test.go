package bridge

import (
	"testing"
)

func TestBurn(t *testing.T) {
	txID := ""
	err := Burn(txID)
	if err != nil {
		t.Error(err)
	}
}

func TestDeposit(t *testing.T) {
	err := Deposit()
	if err != nil {
		t.Error(err)
	}
}

func TestDeploy(t *testing.T) {
	err := Deploy()
	if err != nil {
		t.Error(err)
	}
}
