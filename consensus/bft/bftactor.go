package bft

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/wire"
	"time"
)

type BFTCore struct {
	Name       string
	Chain      consensus.ChainInterface
	PeerID     string
	Round      uint64
	NextHeight uint64

	UserKeySet *cashec.KeySet
	State      string
	Block      consensus.BlockInterface

	ProposeMsgCh chan *wire.MessageBFTProposeV2
	PrepareMsgCh chan *wire.MessageBFTPrepareV2
	StopCh       chan int

	PrepareMsgs map[string]map[string]bool
	Blocks      map[string]consensus.BlockInterface

	IsRunning bool
}

func (e *BFTCore) IsRun() bool {
	return e.IsRunning
}

func (e *BFTCore) GetInfo() string {
	return ""
}

func (e *BFTCore) ReceiveMsg(msg wire.Message) {
	switch msg.MessageType() {
	case wire.CmdBFTPropose:
		e.ProposeMsgCh <- msg.(*wire.MessageBFTProposeV2)
	case wire.CmdBFTPrepare:
		e.PrepareMsgCh <- msg.(*wire.MessageBFTPrepareV2)

	}
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
	e.Blocks = map[string]consensus.BlockInterface{}

	e.ProposeMsgCh = make(chan *wire.MessageBFTProposeV2)
	e.PrepareMsgCh = make(chan *wire.MessageBFTPrepareV2)

	ticker := time.Tick(100 * time.Millisecond)

	go func() {
		for {
			select {
			case <-e.StopCh: //stop protocol -> break actor loop
				return
			case b := <-e.ProposeMsgCh:
				block := json.Unmarshal()
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
					roundKey := fmt.Sprint(e.NextHeight, "_", e.Round)
					if e.Blocks[roundKey] != nil && e.Chain.ValidateBlock(e.Blocks[roundKey]) {
						e.Block = e.Blocks[roundKey]
						e.enterPreparePhase()
					}
				case PREPARE:
					roundKey := fmt.Sprint(e.NextHeight, "_", e.Round)
					if e.Block != nil && e.getMajorityVote(e.PrepareMsgs[roundKey]) == 1 {
						e.Chain.InsertBlk(e.Block, true)
						e.enterNewRound()
					}
					if e.Block != nil && e.getMajorityVote(e.PrepareMsgs[roundKey]) == -1 {
						e.Chain.InsertBlk(e.Block, false)
						e.enterNewRound()
					}
				}
			}
		}
	}()
}
