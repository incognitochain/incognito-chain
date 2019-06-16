package bft2

import (
	"fmt"
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
	TIMEOUT = 60 * time.Second
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
	GetLastBlockTimeStamp() uint64
	GetBlkMinTime() time.Duration
	IsReady() bool
	GetHeight() uint64
	GetCommitteeSize() int
	GetNodePubKeyIndex() int
	GetLastProposerIndex() int
	GetNodePubKey() string
	CreateNewBlock() BlockInterface
	ValidateBlock(interface{}) bool
	ValidateSignature(interface{}, string) bool
	InsertBlk(interface{}, bool)
}

type BlockInterface interface {
	GetHeight() uint64
	GetRound() uint64
	GetProducer() string
	Hash() string
}

type ProposeMsg struct {
	Block BlockInterface
	RoundKey string
}

type PrepareMsg struct {
	IsOk bool
	From string
	Sig string
	BlkHash string
	RoundKey string
}

type CommitMsg struct {
	WillCommit bool
	From string
	Sig string
	BlkHash string
	RoundKey string
}

type BFTEngine struct {
	Chain  ChainInterface
	PeerID string
	Round      uint64
	NextHeight uint64
	State        string
	Block        BlockInterface
	NextStateCh  chan string
	ProposeMsgCh chan ProposeMsg
	PrepareMsgCh chan PrepareMsg
	CommitMsgCh  chan CommitMsg
	
	PrepareMsgs  map[string]map[string]bool
	CommitMsgs  map[string]map[string]bool
	Blocks map[string]BlockInterface
}

func (e *BFTEngine) Start() {
	e.PrepareMsgs = map[string]map[string]bool{}
	e.CommitMsgs = map[string]map[string]bool{}
	e.Blocks = map[string]BlockInterface{}
	
	e.ProposeMsgCh = make(chan ProposeMsg)
	e.PrepareMsgCh = make(chan PrepareMsg)
	e.CommitMsgCh = make(chan CommitMsg)
	e.NextStateCh = make(chan string)

	ticker := time.Tick(100 * time.Millisecond)
	
	go func() {
		for { //action react pattern
			select {
			case s := <-e.NextStateCh:
				e.nextState(s)
			case b := <-e.ProposeMsgCh:
				e.Blocks[b.RoundKey] = b.Block
			case sig := <-e.PrepareMsgCh:
				if e.Chain.ValidateSignature(e.Block, sig.Sig) {
					if e.PrepareMsgs[sig.RoundKey] == nil {
						e.PrepareMsgs[sig.RoundKey] = map[string]bool{}
					}
					e.PrepareMsgs[sig.RoundKey][sig.From] = sig.IsOk
				}
			case commit := <-e.CommitMsgCh:
				if e.CommitMsgs[commit.RoundKey] == nil {
					e.CommitMsgs[commit.RoundKey] = map[string]bool{}
				}
				e.CommitMsgs[commit.RoundKey][commit.From] = commit.WillCommit
				
				//fmt.Println(e.PeerID, " get commit ",commit.RoundKey,commit.From,commit.WillCommit)
			
			case <-ticker:
				if e.Chain.IsReady() {
					if !e.viewIsInTimeFrame() {
						e.nextState(NEWROUND)
					}
				}
				switch e.State {
				case LISTEN:
					roundKey := fmt.Sprint(e.NextHeight, "_", e.Round)
					if e.Blocks[roundKey] != nil &&  e.Chain.ValidateBlock(e.Blocks[roundKey]) {
						e.Block = e.Blocks[roundKey]
						e.nextState(PREPARE)
					}
				case PREPARE:
					roundKey := fmt.Sprint(e.NextHeight, "_", e.Round)
					if e.Block != nil && e.getMajorityVote(e.PrepareMsgs[roundKey]) != 0 {
						e.nextState(COMMIT)
					}
				case COMMIT:
					roundKey := fmt.Sprint(e.NextHeight, "_", e.Round)
					if e.Block != nil && e.getMajorityVote(e.CommitMsgs[roundKey]) == 1 {
						e.Chain.InsertBlk(e.Block, true)
						e.nextState(NEWROUND)
					}
					if e.Block != nil && e.getMajorityVote(e.CommitMsgs[roundKey]) == -1 {
						e.Chain.InsertBlk(e.Block, false)
						e.nextState(NEWROUND)
					}
				}
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
