package bulletproofs

import "github.com/incognitochain/incognito-chain/common"

type logger struct {
	Log common.Logger
}

func (lg *logger) Init(inst common.Logger) {
	lg.Log = inst
}

// Global instant to use
var Logger = logger{}
