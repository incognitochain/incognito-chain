package main

import (
	"encoding/json"
	"github.com/ninjadotorg/constant/common"
	"log"
)

var (
	cfg *params
)

func main() {
	// Show version at startup.
	log.Printf("Version %s\n", "0.0.1")

	// load component
	tcfg, err := loadParams()
	if err != nil {
		log.Println("Parse component error", err.Error())
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
				err := createWallet()
				if err != nil {
					log.Println(err)
					return
				}
			}
		case ListWalletAccountCmd:
			{
				if cfg.WalletPassphrase == "" || cfg.WalletName == "" {
					log.Println("Wrong param")
					return
				}
				accounts, err := listAccounts()
				if err != nil {
					log.Println(err)
					return
				}
				result, err := parseToJsonString(accounts)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println(string(result))
			}
		case GetWalletAccountCmd:
			{
				if cfg.WalletPassphrase == "" || cfg.WalletName == "" || cfg.WalletAccountName == "" {
					log.Println("Wrong param")
					return
				}
				account, err := getAccount(cfg.WalletAccountName)
				if err != nil {
					log.Println(err)
					return
				}
				result, err := parseToJsonString(account)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println(string(result))
			}
		case CreateWalletAccountCmd:
			{
				if cfg.WalletPassphrase == "" || cfg.WalletName == "" || cfg.WalletAccountName == "" {
					log.Println("Wrong param")
					return
				}
				account, err := createAccount(cfg.WalletAccountName)
				if err != nil {
					log.Println(err)
					return
				}
				result, err := parseToJsonString(account)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println(string(result))
			}
		}
	} else {
		log.Println("Parse component error", err.Error())
	}
}

func parseToJsonString(data interface{}) ([]byte, error) {
	result, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return result, nil
}
