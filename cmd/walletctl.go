package main

import (
	"path/filepath"
	"os"
	"github.com/ninjadotorg/cash/wallet"
	"log"
	"errors"
)

func loadWallet() (*wallet.Wallet, error) {
	var walletObj *wallet.Wallet
	walletObj = &wallet.Wallet{}
	walletObj.Config = &wallet.WalletConfig{
		DataDir:        cfg.DataDir,
		DataFile:       cfg.WalletName,
		DataPath:       filepath.Join(cfg.DataDir, cfg.WalletName),
		IncrementalFee: 0,
	}
	err := walletObj.LoadWallet(cfg.WalletPassphrase)
	return walletObj, err
}

func createWallet() error {
	var walletObj *wallet.Wallet
	walletObj = &wallet.Wallet{}
	walletObj.Config = &wallet.WalletConfig{
		DataDir:        cfg.DataDir,
		DataFile:       cfg.WalletName,
		DataPath:       filepath.Join(cfg.DataDir, cfg.WalletName),
		IncrementalFee: 0,
	}
	if _, err := os.Stat(walletObj.Config.DataPath); os.IsNotExist(err) {
		walletObj.Init(cfg.WalletPassphrase, 0, cfg.WalletName)
		walletObj.Save(cfg.WalletPassphrase)
		log.Printf("Create wallet successfully with name: %s", cfg.WalletName)
		return nil
	} else {
		return errors.New("Exist wallet with name %s\n", )
	}
}

func listAccounts() (interface{}, error) {
	walletObj, err := loadWallet()
	if err != nil {
		return nil, err
	}
	accounts := walletObj.ListAccounts()
	return accounts, err
}

func getAccount() (interface{}, error) {
	walletObj, err := loadWallet()
	if err != nil {
		return nil, err
	}
	accounts := walletObj.ListAccounts()
	for _, account := range accounts {
		if cfg.WalletAccountName == account.Name {
			return account, nil
		}
	}
	return nil, errors.New("Not found")
}
