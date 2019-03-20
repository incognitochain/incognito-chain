package wallet

import "github.com/constant-money/constant-chain/common"

type WalletLogger struct {
	log common.Logger
}

func (walletLogger *WalletLogger) Init(inst common.Logger) {
	walletLogger.log = inst
}

// Global instant to use
var Logger = WalletLogger{}
