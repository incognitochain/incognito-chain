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
	consensusName = common.BlsConsensus
)

//
const (
	timeout             = 36 * time.Second       // must be at least twice the time of block creation
	maxNetworkDelayTime = 150 * time.Millisecond // in ms
)
