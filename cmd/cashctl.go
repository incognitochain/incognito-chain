package main

import (
	"log"
	"github.com/ninjadotorg/cash-prototype/common"
	"os"
	"github.com/ninjadotorg/cash-prototype/wallet"
	"path/filepath"
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
		if cfg.WalletPassphrase == "" || cfg.WalletName == "" {
			log.Println("Wrong param")
			return
		}
		if cfg.Command == InitWalletCmd {
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
	} else {
		log.Println("Parse params error", err.Error())
	}
}
