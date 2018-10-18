package main

import (
	"log"
	"github.com/ninjadotorg/cash-prototype/common"
	"os"
	"fmt"
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
		argsWithoutProg := os.Args[3:]
		fmt.Printf("Params %v\n", argsWithoutProg)
		if cfg.Command == InitWalletCmd {
			walletPassphrase := cfg.WalletPassphrase
			walletName := cfg.WalletName
		}
	} else {
		log.Println("Parse params error", err.Error())
	}
}
