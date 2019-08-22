package blsbft

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
)

const (
	CONSENSUSNAME = common.BLS_CONSENSUS
)

type BLSBFT struct {
	Chain    blockchain.ChainInterface
	Node     consensus.NodeInterface
	ChainKey string
	PeerID   string

	UserKeySet       *MiningKey
	BFTMessageCh     chan wire.MessageBFT
	ProposeMessageCh chan BFTPropose
	VoteMessageCh    chan BFTVote

	RoundData struct {
		Block             common.BlockInterface
		BlockValidateData ValidationData
		Votes             map[string]vote
		Round             int
		NextHeight        uint64
		State             string
		NotYetSendVote    bool
	}

	Blocks     map[string]common.BlockInterface
	EarlyVotes map[string]map[string]vote
	isOngoing  bool
	isStarted  bool
	StopCh     chan struct{}
	logger     common.Logger
}

func (e *BLSBFT) IsOngoing() bool {
	return e.isOngoing
}

func (e *BLSBFT) GetConsensusName() string {
	return CONSENSUSNAME
}

func (e *BLSBFT) Stop() {
	if e.isStarted {
		select {
		case <-e.StopCh:
			return
		default:
			close(e.StopCh)
		}
		e.isStarted = false
	}
}

func (e *BLSBFT) Start() {
	if e.isStarted {
		return
	}
	e.isStarted = true
	e.isOngoing = false
	e.StopCh = make(chan struct{})
	e.EarlyVotes = make(map[string]map[string]vote)
	e.Blocks = map[string]common.BlockInterface{}

	e.ProposeMessageCh = make(chan BFTPropose)
	e.VoteMessageCh = make(chan BFTVote)

	ticker := time.Tick(100 * time.Millisecond)

	go func() {
		for { //actor loop
			select {
			case <-e.StopCh:
				return
			case proposeMsg := <-e.ProposeMessageCh:
				block, err := e.Chain.UnmarshalBlock(proposeMsg.Block)
				if err != nil {
					continue
				}
				round := block.GetRound()
				if round < e.RoundData.Round {
					continue
				}
				if e.RoundData.Block != nil {
					if e.RoundData.Block.GetHeight() == block.GetHeight() && e.RoundData.Round == round {
						e.Blocks[getRoundKey(e.RoundData.NextHeight, e.RoundData.Round)] = block
					}
				}
			case voteMsg := <-e.VoteMessageCh:
				if getRoundKey(e.RoundData.NextHeight, e.RoundData.Round) == voteMsg.RoundKey {
					//validate single sig
					if true {
						e.RoundData.Votes[voteMsg.Validator] = voteMsg.Vote
					}
				}
			case <-ticker:
				if e.Chain.GetPubKeyCommitteeIndex(e.UserKeySet.GetPublicKeyBase58()) == -1 {
					e.isOngoing = false
					continue
				}

				if !e.Chain.IsReady() {
					e.isOngoing = false
					continue
				}

				if !e.isInTimeFrame() || e.RoundData.State == "" {
					e.enterNewRound()
				}
				e.isOngoing = true
				switch e.RoundData.State {
				case LISTEN:
					// timeout or vote nil?
					roundKey := fmt.Sprint(e.RoundData.NextHeight, "_", e.RoundData.Round)
					if e.Blocks[roundKey] != nil && e.validatePreSignBlock(e.Blocks[roundKey], e.Chain.GetCommittee()) != nil {
						e.RoundData.Block = e.Blocks[roundKey]
						e.enterVotePhase()
					}
				case VOTE:

					if e.RoundData.NotYetSendVote {
						err := e.sendVote()
						if err != nil {
							e.logger.Error(err)
							continue
						}
					}
					if e.isHasMajorityVotes() {

						keyList, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(e.Chain.GetCommittee(), CONSENSUSNAME)
						aggSig, brigSigs, validatorIdx, err := combineVotes(e.RoundData.Votes, keyList)
						if err != nil {
							e.logger.Error(err)
							continue
						}

						e.RoundData.BlockValidateData.AggSig = aggSig
						e.RoundData.BlockValidateData.BridgeSig = brigSigs
						e.RoundData.BlockValidateData.ValidatiorsIdx = validatorIdx

						validationDataString, _ := EncodeValidationData(e.RoundData.BlockValidateData)
						e.RoundData.Block.(blockValidation).AddValidationField(validationDataString)

						e.Chain.InsertBlk(e.RoundData.Block)
						e.enterNewRound()
					}
				}

			}
		}
	}()
}

func (e *BLSBFT) enterProposePhase() {
	if !e.isInTimeFrame() || e.RoundData.State == PROPOSE {
		return
	}
	e.setState(PROPOSE)

	block := e.Chain.CreateNewBlock(int(e.RoundData.Round))
	validationData := e.CreateValidationData(block.Hash())
	validationDataString, _ := EncodeValidationData(validationData)
	block.(blockValidation).AddValidationField(validationDataString)

	e.RoundData.Block = block
	e.RoundData.BlockValidateData = validationData

	blockData, _ := json.Marshal(e.RoundData.Block)
	msg, _ := MakeBFTProposeMsg(blockData, e.ChainKey, e.UserKeySet)
	go e.Node.PushMessageToChain(msg, e.Chain)
	e.enterVotePhase()
}

func (e *BLSBFT) enterListenPhase() {
	if !e.isInTimeFrame() || e.RoundData.State == LISTEN {
		return
	}
	e.setState(LISTEN)
}

func (e *BLSBFT) enterVotePhase() {
	if !e.isInTimeFrame() || e.RoundData.State == VOTE {
		return
	}
	e.setState(VOTE)
	e.sendVote()
}

func (e *BLSBFT) enterNewRound() {
	//if chain is not ready,  return
	if !e.Chain.IsReady() {
		e.RoundData.State = ""
		return
	}

	//if already running a round for current timeframe
	if e.isInTimeFrame() && e.RoundData.State != NEWROUND {
		return
	}
	e.setState(NEWROUND)
	e.waitForNextRound()

	e.RoundData.NextHeight = e.Chain.CurrentHeight() + 1
	e.RoundData.Round = e.getCurrentRound()
	e.RoundData.Votes = make(map[string]vote)
	e.RoundData.Block = nil

	if e.Chain.GetPubKeyCommitteeIndex(e.UserKeySet.GetPublicKeyBase58()) == (e.Chain.GetLastProposerIndex()+1+e.RoundData.Round)%e.Chain.GetCommitteeSize() {
		e.logger.Info("BFT: new round propose")
		e.enterProposePhase()
	} else {
		e.logger.Info("BFT: new round listen")
		e.enterListenPhase()
	}

}

func (e BLSBFT) NewInstance(chain blockchain.ChainInterface, chainKey string, node consensus.NodeInterface, logger common.Logger) consensus.ConsensusInterface {
	var newInstance BLSBFT
	newInstance.Chain = chain
	newInstance.ChainKey = chainKey
	newInstance.Node = node
	newInstance.UserKeySet = e.UserKeySet
	newInstance.logger = logger
	return &newInstance
}

func init() {
	consensus.RegisterConsensus(common.BLS_CONSENSUS, &BLSBFT{})
}
