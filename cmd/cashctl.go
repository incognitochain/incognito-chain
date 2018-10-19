package main

import (
	"log"
	"os"
	"path/filepath"
	"encoding/json"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/wallet"
	"github.com/pkg/errors"
)

var (
	cfg *config
)

func main() {
	// Show version at startup.
	log.Printf("Version %s\n", "0.0.1")

	// load config
	tcfg, err := loadConfig()
	if err != nil {
		log.Println("Parse params error", err.Error())
		return
	}
	cfg = tcfg

	log.Printf("Process cmd: %s", cfg.Command)
	if ok, err := common.SliceExists(CmdList, cfg.Command); ok || err == nil {
		switch cfg.Command {
		case CreateWalletCmd:
			{
				if cfg.WalletPassphrase == "" || cfg.WalletName == "" {
					log.Println("Wrong param")
					return
				}
				createWallet()
			}
		case ListWalletAccountCmd:
			{
				if cfg.WalletPassphrase == "" || cfg.WalletName == "" {
					log.Println("Wrong param")
					return
				}
				accounts, _ := listAccounts()
				result, _ := json.MarshalIndent(accounts, "", "\t")
				log.Println(string(result))
			}
		case GetWalletAccountCmd:
			{
				if cfg.WalletPassphrase == "" || cfg.WalletName == "" || cfg.WalletAccountName == "" {
					log.Println("Wrong param")
					return
				}
				accounts, err := getAccount()
				if err != nil {
					log.Println(err)
					return
				}
				result, _ := json.MarshalIndent(accounts, "", "\t")
				log.Println(string(result))
			}
		}
	} else {
		log.Println("Parse params error", err.Error())
	}
}

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

func createWallet() {
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
	} else {
		log.Printf("Exist wallet with name %s\n", )
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
