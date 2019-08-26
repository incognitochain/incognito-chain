package blsbft

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

const (
	PROPOSE       = "PROPOSE"
	LISTEN        = "LISTEN"
	VOTE          = "VOTE"
	NEWROUND      = "NEWROUND"
	BLS           = "bls"
	BRI           = "dsa"
	CONSENSUSNAME = common.BLS_CONSENSUS
)

//
const (
	TIMEOUT             = 10 * time.Second
	MaxNetworkDelayTime = 150 * time.Millisecond // in ms
)
