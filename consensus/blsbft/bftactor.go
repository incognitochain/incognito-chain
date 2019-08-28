package blsbft

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
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
		Committee         []incognitokey.CommitteePublicKey
		CommitteeBLS      struct {
			StringList []string
			ByteList   []blsmultisig.PublicKey
		}
		LastProposerIndex int
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

	ticker := time.Tick(200 * time.Millisecond)
	e.logger.Info("start bls-bft consensus for chain", e.ChainKey)
	go func() {
		for { //actor loop
			select {
			case <-e.StopCh:
				return
			case proposeMsg := <-e.ProposeMessageCh:
				block, err := e.Chain.UnmarshalBlock(proposeMsg.Block)
				if err != nil {
					e.logger.Info(err)
					continue
				}
				blockRoundKey := getRoundKey(block.GetHeight(), block.GetRound())
				e.logger.Info("receive block", blockRoundKey, getRoundKey(e.RoundData.NextHeight, e.RoundData.Round))
				if block.GetHeight() >= e.RoundData.NextHeight {
					if e.RoundData.NextHeight == block.GetHeight() && e.RoundData.Round > block.GetRound() {
						e.logger.Error("wrong round")
						continue
					}
					if e.RoundData.Round == block.GetRound() && e.RoundData.Block != nil {

						e.Blocks[blockRoundKey] = block

						continue
					}
					e.Blocks[blockRoundKey] = block
				}
			case voteMsg := <-e.VoteMessageCh:
				e.logger.Info("receive vote", voteMsg.RoundKey, getRoundKey(e.RoundData.NextHeight, e.RoundData.Round))
				if getRoundKey(e.RoundData.NextHeight, e.RoundData.Round) == voteMsg.RoundKey {
					//validate single sig
					if e.RoundData.Block != nil {
						validatorIdx := common.IndexOfStr(voteMsg.Validator, e.RoundData.CommitteeBLS.StringList)
						if err := validateSingleBLSSig(e.RoundData.Block.Hash(), voteMsg.Vote.BLS, validatorIdx, e.RoundData.CommitteeBLS.ByteList); err != nil {
							e.logger.Error(err)
							continue
						}
						e.RoundData.Votes[voteMsg.Validator] = voteMsg.Vote
						e.logger.Warn("vote added...")
						continue
					}
				}
				if _, ok := e.EarlyVotes[voteMsg.RoundKey]; !ok {
					e.EarlyVotes[voteMsg.RoundKey] = make(map[string]vote)
				}
				e.EarlyVotes[voteMsg.RoundKey][voteMsg.Validator] = voteMsg.Vote

			case <-ticker:
				pubKey := e.UserKeySet.GetPublicKey()
				if common.IndexOfStr(pubKey.GetMiningKeyBase58(CONSENSUSNAME), e.RoundData.CommitteeBLS.StringList) == -1 {
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
					roundKey := getRoundKey(e.RoundData.NextHeight, e.RoundData.Round)
					if e.Blocks[roundKey] != nil {
						if err := e.validatePreSignBlock(e.Blocks[roundKey]); err != nil {
							e.logger.Error(err)
							time.Sleep(1 * time.Second)
							continue
						}
						if e.RoundData.Block == nil {
							e.RoundData.Block = e.Blocks[roundKey]
							valData, err := DecodeValidationData(e.RoundData.Block.GetValidationField())
							if err != nil {
								e.logger.Error(err)
								time.Sleep(1 * time.Second)
								continue
							}
							e.RoundData.BlockValidateData = *valData
							e.enterVotePhase()
						}
					}
				case VOTE:
					if e.RoundData.NotYetSendVote {
						err := e.sendVote()
						if err != nil {
							e.logger.Error(err)
							continue
						}
					}
					if e.RoundData.Block != nil && e.isHasMajorityVotes() {
						aggSig, brigSigs, validatorIdx, err := combineVotes(e.RoundData.Votes, e.RoundData.CommitteeBLS.StringList)
						if err != nil {
							e.logger.Error(err)
							time.Sleep(1 * time.Second)
							continue
						}

						e.RoundData.BlockValidateData.AggSig = aggSig
						e.RoundData.BlockValidateData.BridgeSig = brigSigs
						e.RoundData.BlockValidateData.ValidatiorsIdx = validatorIdx

						validationDataString, _ := EncodeValidationData(e.RoundData.BlockValidateData)
						e.RoundData.Block.(blockValidation).AddValidationField(validationDataString)

						err = e.ValidateCommitteeSig(e.RoundData.Block, e.RoundData.Committee)
						if err != nil {
							fmt.Println(e.RoundData.Block.GetValidationField())
							fmt.Print("\n")
							fmt.Println(e.RoundData.Committee)
							fmt.Print("\n")
							for _, member := range e.RoundData.Committee {
								fmt.Println(base58.Base58Check{}.Encode(member.MiningPubKey[CONSENSUSNAME], common.Base58Version))
							}
							e.logger.Critical(err)
							time.Sleep(1 * time.Second)
							return
						}

						if err := e.Chain.InsertBlk(e.RoundData.Block); err != nil {
							e.logger.Error(err)
							time.Sleep(1 * time.Second)
							continue
						}
						e.logger.Warn("Commit block! Wait for next round")

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
	time1 := time.Now()
	block, err := e.Chain.CreateNewBlock(int(e.RoundData.Round))
	e.logger.Info("create block", time.Since(time1).Seconds())

	if err != nil {
		e.logger.Error("can't create block", err)
		return
	}
	validationData := e.CreateValidationData(block)
	validationDataString, _ := EncodeValidationData(validationData)
	block.(blockValidation).AddValidationField(validationDataString)

	e.RoundData.Block = block
	e.RoundData.BlockValidateData = validationData

	blockData, _ := json.Marshal(e.RoundData.Block)
	msg, _ := MakeBFTProposeMsg(blockData, e.ChainKey, e.UserKeySet)
	e.logger.Info("push block", time.Since(time1).Seconds())
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
	e.logger.Info("enter voting phase")
	if !e.isInTimeFrame() || e.RoundData.State == VOTE {
		return
	}
	e.setState(VOTE)
	err := e.sendVote()
	if err != nil {
		e.logger.Error(err)
	}
}

func (e *BLSBFT) enterNewRound() {
	//if chain is not ready,  return
	if !e.Chain.IsReady() {
		fmt.Println("BLSBFT", "not ready", e.ChainKey)
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
	e.RoundData.NotYetSendVote = true
	e.RoundData.LastProposerIndex = e.Chain.GetLastProposerIndex()
	committee := e.Chain.GetCommittee()
	if !reflect.DeepEqual(e.RoundData.Committee, committee) {
		e.RoundData.Committee = committee
		for _, member := range e.RoundData.Committee {
			e.RoundData.CommitteeBLS.ByteList = append(e.RoundData.CommitteeBLS.ByteList, member.MiningPubKey[CONSENSUSNAME])
		}
		committeeBLSString, err := incognitokey.ExtractPublickeysFromCommitteeKeyList(e.RoundData.Committee, CONSENSUSNAME)
		if err != nil {
			e.logger.Error(err)
			return
		}
		e.RoundData.CommitteeBLS.StringList = committeeBLSString
	}

	e.logger.Info("")
	e.logger.Info("============================================")
	e.logger.Info("")

	pubKey := e.UserKeySet.GetPublicKey()
	if e.Chain.GetPubKeyCommitteeIndex(pubKey.GetMiningKeyBase58(CONSENSUSNAME)) == (e.Chain.GetLastProposerIndex()+e.RoundData.Round)%e.Chain.GetCommitteeSize() {
		e.logger.Info("BFT: new round => PROPOSE", e.RoundData.NextHeight, e.RoundData.Round)
		e.enterProposePhase()
	} else {
		e.logger.Info("BFT: new round => LISTEN", e.RoundData.NextHeight, e.RoundData.Round)
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
