package txpool

import "github.com/incognitochain/incognito-chain/common"

type TxPoolLogger struct {
	common.Logger
}

func (self *TxPoolLogger) Init(inst common.Logger) {
	self.Logger = inst
}

// Global instant to use
var Logger = TxPoolLogger{}
