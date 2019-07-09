package main

const (
	createWalletCmd        = "createwallet"
	listWalletAccountCmd   = "listaccounts"
	getWalletAccountCmd    = "getaccount"
	createWalletAccountCmd = "createaccount"
	getPrivacyTokenID      = "getprivacytokenid"
	backupChain            = "backupchain"
	restoreChain           = "restorechain"
)

var CmdList = []string{createWalletCmd, listWalletAccountCmd, getWalletAccountCmd, createWalletAccountCmd, getPrivacyTokenID, backupChain, restoreChain}
