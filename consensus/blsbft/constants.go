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
	CONSENSUSNAME = common.BlsConsensus
)

//
const (
	TIMEOUT             = 20 * time.Second       // must be at least twice the time of block creation
	MaxNetworkDelayTime = 150 * time.Millisecond // in ms
)
