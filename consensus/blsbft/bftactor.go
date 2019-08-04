package blsbft

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/consensus/chain"
	"github.com/incognitochain/incognito-chain/consensus/multisigschemes/bls"
)

const (
	CONSENSUSNAME = "BLSBFT"
)

type ProposeMsg struct {
	ChainKey   string
	Block      chain.BlockInterface
	ContentSig string
	Pubkey     string
	Timestamp  int64
	RoundKey   string
}

type PrepareMsg struct {
	ChainKey   string
	IsOk       bool
	Pubkey     string
	ContentSig string
	BlkHash    string
	RoundKey   string
	Timestamp  int64
}

type SigStatus struct {
	IsOk       bool
	Verified   bool
	SigContent string
}

type BLSBFT struct {
	ChainKey   string
	Chain      chain.ChainInterface
	PeerID     string
	Round      int
	NextHeight uint64

	// UserKeySet        *incognitokey.KeySet
	UserKeySet      *bls.KeySet
	State           string
	NotYetSendAgree bool

	Block chain.BlockInterface

	ProposeMsgCh chan ProposeMsg
	PrepareMsgCh chan PrepareMsg
	StopCh       chan int

	PrepareMsgs map[string]map[string]SigStatus

	Blocks map[string]chain.BlockInterface

	isOngoing bool
}

func (e *BLSBFT) IsOngoing() bool {
	return e.isOngoing
}

func (e *BLSBFT) GetConsensusName() string {
	return CONSENSUSNAME
}

func (e *BLSBFT) ReceiveProposeMsg(msg interface{}) {
	e.ProposeMsgCh <- msg.(ProposeMsg)
}

func (e *BLSBFT) ReceivePrepareMsg(msg interface{}) {
	e.PrepareMsgCh <- msg.(PrepareMsg)
}

func (e *BLSBFT) Stop() {
	select {
	case <-e.StopCh:
		return
	default:
		close(e.StopCh)
	}
}

func (e *BLSBFT) Start() {
	e.isOngoing = false
	e.StopCh = make(chan int)
	e.PrepareMsgs = map[string]map[string]SigStatus{}
	e.Blocks = map[string]chain.BlockInterface{}

	e.ProposeMsgCh = make(chan ProposeMsg)
	e.PrepareMsgCh = make(chan PrepareMsg)

	ticker := time.Tick(100 * time.Millisecond)

	//TODO: clean up buffer msgs
	go func() {
		for { //actor loop
			select {
			case <-e.StopCh:
				return
			case b := <-e.ProposeMsgCh:
				round := b.Block.GetRound()
				if round < e.Round {
					continue
				}
				if e.Block != nil {
					if e.Block.GetHeight() == b.Block.GetHeight() && e.Round == round {
						e.Blocks[b.RoundKey] = b.Block
					}

				}

			case sig := <-e.PrepareMsgCh:
				if e.PrepareMsgs[sig.RoundKey] == nil {
					e.PrepareMsgs[sig.RoundKey] = map[string]SigStatus{}
				}
				e.PrepareMsgs[sig.RoundKey][sig.Pubkey] = SigStatus{sig.IsOk, false, sig.ContentSig}

			case <-ticker:
				if e.Chain.GetPubKeyCommitteeIndex(e.UserKeySet.GetPubkeyB58()) == -1 {
					continue
				}

				if !e.Chain.IsReady() {
					continue
				}

				if !e.isInTimeFrame() || e.State == "" {
					e.enterNewRound()
				}

				switch e.State {
				case LISTEN:
					// timeout or vote nil?
					roundKey := fmt.Sprint(e.NextHeight, "_", e.Round)
					if e.Blocks[roundKey] != nil && e.Chain.ValidatePreSignBlock(e.Blocks[roundKey]) != nil {
						e.Block = e.Blocks[roundKey]
						e.enterAgreePhase()
					}

				case AGREE:

					if e.NotYetSendAgree {
						e.validateAndSendVote()
					}

					roundKey := fmt.Sprint(e.NextHeight, "_", e.Round)
					if e.Block != nil && e.getMajorityVote(e.PrepareMsgs[roundKey]) == 1 {
						//TODO: aggregate sigs
						e.Chain.InsertBlk(e.Block, true)
						e.enterNewRound()
					}
					//if e.Block != nil && e.getMajorityVote(e.PrepareMsgs[roundKey]) == -1 {
					//	e.Chain.InsertBlk(e.Block, false)
					//	e.enterNewRound()
					//}
				}

			}
		}
	}()
}

func (e *BLSBFT) enterProposePhase() {
	if !e.isInTimeFrame() || e.State == PROPOSE {
		return
	}
	e.setState(PROPOSE)

	block := e.Chain.CreateNewBlock(int(e.Round))
	e.Block = block

	blockData, _ := json.Marshal(e.Block)
	msg, _ := MakeBFTProposeMsg(blockData, e.ChainKey, e.UserKeySet)
	go e.Chain.PushMessageToValidator(msg)
	e.enterAgreePhase()

}

func (e *BLSBFT) enterListenPhase() {
	if !e.isInTimeFrame() || e.State == LISTEN {
		return
	}
	e.setState(LISTEN)
}

func (e *BLSBFT) enterAgreePhase() {
	if !e.isInTimeFrame() || e.State == AGREE {
		return
	}
	e.setState(AGREE)
	e.validateAndSendVote()
}

func (e *BLSBFT) enterNewRound() {
	//if chain is not ready,  return
	if !e.Chain.IsReady() {
		e.State = ""
		return
	}

	//if already running a round for current timeframe
	if e.isInTimeFrame() && e.State != NEWROUND {
		return
	}
	e.setState(NEWROUND)
	e.waitForNextRound()

	e.Round = e.getCurrentRound()
	e.NextHeight = e.Chain.GetHeight() + 1
	e.Block = nil

	// if e.Chain.GetNodePubKeyCommitteeIndex() == (e.Chain.GetLastProposerIndex()+1+int(e.Round))%e.Chain.GetCommitteeSize() {
	// 	fmt.Println("BFT: new round propose")
	// 	e.enterProposePhase()
	// } else {
	// 	fmt.Println("BFT: new round listen")
	// 	e.enterListenPhase()
	// }

}

func init() {
	consensus.RegisterConsensus(common.BLSBFT, BLSBFT{})
}
