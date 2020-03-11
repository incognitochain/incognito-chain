package mlsag

import "github.com/incognitochain/incognito-chain/common"

type MlsagLogger struct {
	log common.Logger
}

func (mlsag *MlsagLogger) Init(inst common.Logger) {
	mlsag.log = inst
}

// Global instant to use
var Logger = MlsagLogger{}

const (
	HashSize = 32
)
