package wallet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInit(t *testing.T){
	data := []struct {
		passPhrase string
		numOfAccount uint32
		name string
	}{
		{"", uint32(2), "Wallet1"},
		{"12345678", uint32(0), "Wallet1"},
		{"12345678", uint32(3), ""},
	}

	wallet := new(Wallet)

	for _, item := range data {
		wallet.Init(item.passPhrase, item.numOfAccount, item.name)

		if item.numOfAccount == 0{
			assert.Equal(t, 1, len(wallet.MasterAccount.Child))
		} else {
			assert.Equal(t, int(item.numOfAccount), len(wallet.MasterAccount.Child))
		}

		if item.name == "" {
			assert.Equal(t, WalletNameDefault, wallet.Name)
		} else {
			assert.Equal(t, item.name, wallet.Name)
		}

		assert.Equal(t, item.passPhrase, wallet.PassPhrase)
		assert.Equal(t, SeedKeyLen, len(wallet.Seed))
		assert.Greater(t, len(wallet.Mnemonic), 0)
	}
}
