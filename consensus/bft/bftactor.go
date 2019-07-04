package bft

import (
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
