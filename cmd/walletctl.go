package main

import (
	"errors"
	"github.com/incognitochain/incognito-chain/wallet"
	"log"
	"os"
	"path/filepath"
)

var walletObj *wallet.Wallet

func loadWallet() (*wallet.Wallet, error) {
	var walletObj *wallet.Wallet
	walletObj = &wallet.Wallet{}
	walletObj.SetConfig(&wallet.WalletConfig{
		DataDir:        cfg.DataDir,
		DataFile:       cfg.WalletName,
		DataPath:       filepath.Join(cfg.DataDir, cfg.WalletName),
		IncrementalFee: 0,
	})
	err := walletObj.LoadWallet(cfg.WalletPassphrase)
	return walletObj, err
}

func createWallet() error {
	var walletObj *wallet.Wallet
	walletObj = &wallet.Wallet{}
	walletObj.SetConfig(&wallet.WalletConfig{
		DataDir:        cfg.DataDir,
		DataFile:       cfg.WalletName,
		DataPath:       filepath.Join(cfg.DataDir, cfg.WalletName),
		IncrementalFee: 0,
	})
	if _, err := os.Stat(walletObj.GetConfig().DataPath); os.IsNotExist(err) {
		err1 := walletObj.Init(cfg.WalletPassphrase, 0, cfg.WalletName)
		if err1 != nil {
			log.Println(err)
			return nil
		}
		walletObj.Save(cfg.WalletPassphrase)
		log.Printf("Create wallet successfully with name: %s", cfg.WalletName)
		return nil
	} else {
		return errors.New("Exist wallet with name %s\n")
	}
}

func listAccounts() (interface{}, error) {
	walletObj, err := loadWallet()
	if err != nil {
		return nil, err
	}
	accounts := walletObj.ListAccounts()
	result := []string{}
	for accountName, _ := range accounts {
		result = append(result, accountName)
	}
	return result, err
}

func getAccount(accountName string) (interface{}, error) {
	var err error
	walletObj, err = loadWallet()
	if err != nil {
		return nil, err
	}
	accounts := walletObj.ListAccounts()
	for _, account := range accounts {
		if accountName == account.Name {
			result := make(map[string]interface{})
			result["Name"] = accountName
			result["PrivateKey"] = account.Key.Base58CheckSerialize(wallet.PriKeyType)
			result["PaymentAddress"] = account.Key.Base58CheckSerialize(wallet.PaymentAddressType)
			result["ReadonlyKey"] = account.Key.Base58CheckSerialize(wallet.ReadonlyKeyType)
			return result, nil
		}
	}
	return nil, errors.New("Not found")
}

func createAccount(accountName string, shardID *byte) (interface{}, error) {
	account, _ := getAccount(accountName)
	if account != nil {
		return nil, errors.New("Existed account")
	}

	if walletObj != nil {
		account1, err := walletObj.CreateNewAccount(accountName, shardID)
		if err != nil {
			return nil, err
		}
		err = walletObj.Save(cfg.WalletPassphrase)
		if err != nil {
			return nil, err
		}
		if account1 == nil {
			return nil, errors.New("Can not create account")
		}
		log.Printf("Create account '%s' successfully", accountName)
		log.Printf("Private key: %s", account1.Key.Base58CheckSerialize(wallet.PriKeyType))
		log.Printf("Payment address: %s", account1.Key.Base58CheckSerialize(wallet.PaymentAddressType))
		log.Printf("Readonly key: %s", account1.Key.Base58CheckSerialize(wallet.ReadonlyKeyType))
		return account, nil
	}
	return nil, errors.New("Can not load wallet")
}
