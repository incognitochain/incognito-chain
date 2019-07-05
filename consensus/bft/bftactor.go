package bft

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/consensus/chain"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"time"
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

type BFTCore struct {
	ChainKey   string
	Chain      chain.ChainInterface
	PeerID     string
	Round      uint64
	NextHeight uint64

	UserKeySet *incognitokey.KeySet
	State      string
	Block      chain.BlockInterface

	ProposeMsgCh chan ProposeMsg
	PrepareMsgCh chan PrepareMsg
	StopCh       chan int

	PrepareMsgs map[string]map[string]bool
	Blocks      map[string]chain.BlockInterface

	IsRunning bool
}

func (e *BFTCore) IsRun() bool {
	return e.IsRunning
}

func (e *BFTCore) GetInfo() string {
	return ""
}

func (e *BFTCore) ReceiveProposeMsg(msg interface{}) {
	e.ProposeMsgCh <- msg.(ProposeMsg)
}

func (e *BFTCore) ReceivePrepareMsg(msg interface{}) {
	e.PrepareMsgCh <- msg.(PrepareMsg)
}

func (e *BFTCore) Stop() {
	if e.IsRunning {
		close(e.StopCh)
		e.IsRunning = false
	}
}

func (e *BFTCore) Start() {
	e.IsRunning = true
	e.StopCh = make(chan int)
	e.PrepareMsgs = map[string]map[string]bool{}
	e.Blocks = map[string]chain.BlockInterface{}

	e.ProposeMsgCh = make(chan ProposeMsg)
	e.PrepareMsgCh = make(chan PrepareMsg)

	ticker := time.Tick(100 * time.Millisecond)

	go func() {
		for {
			select {
			case <-e.StopCh: //stop protocol -> break actor loop
				return
			case b := <-e.ProposeMsgCh:
				e.Blocks[b.RoundKey] = b.Block
			case sig := <-e.PrepareMsgCh:
				if e.Chain.ValidateSignature(e.Block, sig.ContentSig) {
					if e.PrepareMsgs[sig.RoundKey] == nil {
						e.PrepareMsgs[sig.RoundKey] = map[string]bool{}
					}
					e.PrepareMsgs[sig.RoundKey][sig.Pubkey] = sig.IsOk
				}
			case <-ticker:
				if e.Chain.IsReady() {
					if !e.isInTimeFrame() {
						e.enterNewRound()
					}
				} else {
					//if not ready, stay in new round phase
					e.enterNewRound()
				}

				switch e.State {
				case LISTEN:
					//TODO: timeout or vote nil?
					roundKey := fmt.Sprint(e.NextHeight, "_", e.Round)
					if e.Blocks[roundKey] != nil && e.Chain.ValidateBlock(e.Blocks[roundKey]) {
						e.Block = e.Blocks[roundKey]
						e.enterPreparePhase()
					}
				case PREPARE:
					//retrieve all block with next height and check for majority vote
					roundKey := fmt.Sprint(e.NextHeight, "_", e.Round)
					if e.Block != nil && e.getMajorityVote(e.PrepareMsgs[roundKey]) == 1 {
						e.Chain.InsertBlk(&e.Block, true)
						e.enterNewRound()
					}
					if e.Block != nil && e.getMajorityVote(e.PrepareMsgs[roundKey]) == -1 {
						e.Chain.InsertBlk(&e.Block, false)
						e.enterNewRound()
					}
				}

			}
		}
	}()
}



// create new block (sequence number)
func (e *BFTCore) enterProposePhase() {
	if !e.isInTimeFrame() || e.State == PROPOSE {
		return //not in right time frame or already in propose phase
	}
	e.setState(PROPOSE)
	
	block := e.Chain.CreateNewBlock(int(e.Round))
	e.Block = block
	e.debug("start propose block", block)
	
	blockData, _ := json.Marshal(e.Block)
	msg, _ := MakeBFTProposeMsg(string(blockData), e.ChainKey, fmt.Sprint(e.NextHeight, "_", e.Round), e.UserKeySet)
	go e.Chain.PushMessageToValidator(msg)
	e.enterPreparePhase()
	
}

//listen for block
func (e *BFTCore) enterListenPhase() {
	if !e.isInTimeFrame() || e.State == LISTEN {
		return //not in right time frame or already in listen phase
	}
	e.setState(LISTEN)
	//e.debug("start listen block")
}

//send prepare message (signature of that message & sequence number) and wait for > 2/3 signature of nodes
//block for the message and sequence number
func (e *BFTCore) enterPreparePhase() {
	if !e.isInTimeFrame() || e.State == PREPARE {
		return //not in right time frame or already in prepare phase
	}
	e.setState(PREPARE)
	//e.debug("start prepare phase")
	//TODO: validate block isOK???
	msg, _ := MakeBFTPrepareMsg(true, e.ChainKey, e.Block.Hash().String(), fmt.Sprint(e.NextHeight, "_", e.Round), e.UserKeySet)
	go e.Chain.PushMessageToValidator(msg)
}

func (e *BFTCore) enterNewRound() {
	//if chain is not ready, return
	if !e.Chain.IsReady() {
		return
	}
	
	//if already running a round for current timeframe
	if e.isInTimeFrame() && e.State != NEWROUND {
		return
	}
	
	e.setState(NEWROUND)
	
	//wait for min blk time
	e.waitForNextRound()
	
	//move to next phase
	
	//create new round
	e.Round = e.getCurrentRound()
	e.NextHeight = e.Chain.GetHeight() + 1
	e.Block = nil
	
	if e.Chain.GetNodePubKeyIndex() == (e.Chain.GetLastProposerIndex()+1+int(e.Round))%e.Chain.GetCommitteeSize() {
		fmt.Println("BFT: new round propose")
		e.enterProposePhase()
	} else {
		fmt.Println("BFT: new round propose")
		e.enterListenPhase()
	}
	
}
