package blsbft

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

const (
	proposePhase  = "PROPOSE"
	listenPhase   = "LISTEN"
	votePhase     = "VOTE"
	newround      = "NEWROUND"
	bls           = "bls"
	bri           = "dsa"
	consensusName = common.BlsConsensus
)

//
const (
	timeout             = 20 * time.Second       // must be at least twice the time of block creation
	maxNetworkDelayTime = 150 * time.Millisecond // in ms
)
