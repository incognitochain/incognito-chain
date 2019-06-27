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

func TestCreateNewAccount(t *testing.T){
	data := []struct {
		accountName string
		shardID byte
	}{
		{"", byte(0)},
		{"Acc A", byte(1)},
		//{"Acc A", },
	}

	wallet := new(Wallet)
	wallet.Init("", 0, "")

	numAccount := len(wallet.MasterAccount.Child)

	for _, item := range data {
		Logger.log.Infof("item.accountName: ", item.accountName)
		wallet.CreateNewAccount(item.accountName, &item.shardID)
		newAccount := wallet.MasterAccount.Child[numAccount]

		assert.Equal(t, numAccount + 1, len(wallet.MasterAccount.Child))

		if item.accountName == "" {
			assert.Equal(t, "AccountWallet "+string(numAccount), newAccount.Name)
		}
	}
}

