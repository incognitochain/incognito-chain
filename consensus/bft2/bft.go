package bft2

import (
	"github.com/constant-money/constant-chain/wire"
	"time"
)

/*
	Sequence Number: blockheight + round
*/

const (
	PROPOSE  = "PROPOSE"
	LISTEN   = "LISTEN"
	PREPARE  = "PREPARE"
	COMMIT   = "COMMIT"
	NEWROUND = "NEWROUND"
)

const (
	TIMEOUT = 30 * time.Second
)

type ChainInterface interface {
	// list functions callback which are assigned from Server struct
	//GetPeerIDsFromPublicKey(string) []libp2p.ID
	//PushMessageToAll(wire.Message) error
	//PushMessageToPeer(wire.Message, libp2p.ID) error
	//PushMessageToShard(wire.Message, byte) error
	//PushMessageToBeacon(wire.Message) error
	//PushMessageToPbk(wire.Message, string) error
	//UpdateConsensusState(role string, userPbk string, currentShard *byte, beaconCommittee []string, shardCommittee map[byte][]string)
	PushMessageToValidator(wire.Message) error
	GetLastBlockTimeStamp() int64
	GetBlkMinTime() time.Duration
	IsReady() bool
	GetHeight() uint64
	GetCommitteeSize() int
	GetNodePubKeyIndex() int
	CreateNewBlock() BlockInterface
}

type BlockInterface interface {
	Validate() bool
}

type ProposeMsg struct {
	isOk bool
}

type PrepareMsg struct {
	isOk bool
}

type CommitMsg struct {
	isOk bool
}

var PrepareMsgs = []PrepareMsg{}
var CommitMsgs = []CommitMsg{}
var BlockMsgs = []BlockInterface{}

//type View struct {
//	PeerID        string
//
//}

type BFTEngine struct {
	Chain  ChainInterface
	PeerID string

	Round      int
	NextHeight uint64

	State        string
	Block        BlockInterface
	NextStateCh  chan string
	ProposeMsgCh chan ProposeMsg
	PrepareMsgCh chan PrepareMsg
	CommitMsgCh  chan CommitMsg
}

func (e *BFTEngine) Start() {
	e.ProposeMsgCh = make(chan ProposeMsg)
	e.PrepareMsgCh = make(chan PrepareMsg)
	e.CommitMsgCh = make(chan CommitMsg)
	e.NextStateCh = make(chan string)

	//var stateCache = make(map[uint64]BFTState)
	//var curentState = BFTState{}
	ticker := time.Tick(100 * time.Millisecond)

	// goroutine to update view
	go func() {
		for {
			select {
			case <-ticker:
				//round timeout
				if e.Chain.IsReady() {
					if !e.viewIsInTimeFrame() {
						e.NextStateCh <- NEWROUND
					}
				}
			}
		}
	}()

	go func() {
		for { //action react pattern
			select {
			case s := <-e.NextStateCh:
				e.nextState(s)
			case <-e.ProposeMsgCh:
			case <-e.PrepareMsgCh:
			case <-e.CommitMsgCh:

			}
		}
	}()

}

func (e *BFTEngine) nextState(nextState string) {
	switch nextState {
	case PROPOSE:
		e.handleProposePhase()
	case LISTEN:
		e.handleListenPhase()
	case PREPARE:
		e.handlePreparePhase()
	case COMMIT:
		e.handleCommitPhase()
	case NEWROUND:
		e.handleNewRoundPhase()
	}
}
