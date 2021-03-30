package blsbft

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"

	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"

	"github.com/incognitochain/incognito-chain/metrics/monitor"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
)

type actorV1 struct {
	actorBase
	roundData struct {
		timeStart         time.Time
		block             types.BlockInterface
		blockHash         common.Hash
		blockValidateData ValidationData
		lockVotes         sync.Mutex
		votes             map[string]vote
		round             int
		nextHeight        uint64
		state             string
		notYetSendVote    bool
		committee         []incognitokey.CommitteePublicKey
		committeeBLS      struct {
			stringList []string
			byteList   []blsmultisig.PublicKey
		}
		lastProposerIndex int
	}
	blocks         map[string]types.BlockInterface
	earlyVotes     map[string]map[string]vote
	lockEarlyVotes sync.Mutex
	isOngoing      bool
	stopCh         chan struct{}
}

func (actorV1 *actorV1) IsOngoing() bool {
	return actorV1.isOngoing
}

func (actorV1 *actorV1) Destroy() {
	actorV1.Stop()
}

func (actorV1 *actorV1) Stop() error {
	err := actorV1.actorBase.Stop()
	if err != nil {
		return NewConsensusError(ConsensusAlreadyStoppedError, err)
	}
	if actorV1.isStarted {
		actorV1.logger.Info("stop bls-bft consensus for chain", actorV1.chainKey)
		close(actorV1.stopCh)
		actorV1.isOngoing = false
		return nil
	}
	return NewConsensusError(ConsensusAlreadyStoppedError, errors.New(actorV1.chainKey))
}

func (actorV1 *actorV1) Start() error {
	if actorV1.isStarted {
		return NewConsensusError(ConsensusAlreadyStartedError, errors.New(actorV1.chainKey))
	}

	actorV1.isStarted = true
	actorV1.isOngoing = false
	actorV1.stopCh = make(chan struct{})
	actorV1.earlyVotes = make(map[string]map[string]vote)
	actorV1.blocks = map[string]types.BlockInterface{}
	actorV1.proposeMessageCh = make(chan BFTPropose)
	actorV1.voteMessageCh = make(chan BFTVote)
	actorV1.initRoundData()

	actorV1.logger.Info("start bls-bft consensus for chain", actorV1.chainKey)

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for { //actor loop
			select {
			case <-actorV1.stopCh:
				actorV1.logger.Info("Exit BFT")
				return
			case proposeMsg := <-actorV1.proposeMessageCh:
				block, err := actorV1.chain.UnmarshalBlock(proposeMsg.Block)
				if err != nil {
					actorV1.logger.Info(err)
					continue
				}
				blockRoundKey := getRoundKey(block.GetHeight(), block.GetRound())
				actorV1.logger.Info("receive block", blockRoundKey, getRoundKey(actorV1.roundData.nextHeight, actorV1.roundData.round))
				if block.GetHeight() == actorV1.roundData.nextHeight {
					if actorV1.roundData.round == block.GetRound() {
						if actorV1.roundData.block == nil {
							actorV1.blocks[blockRoundKey] = block
							continue
						}
					} else {
						if actorV1.roundData.round < block.GetRound() {
							actorV1.blocks[blockRoundKey] = block
							continue
						}
					}
					continue
				}
				if block.GetHeight() > actorV1.roundData.nextHeight {
					actorV1.blocks[blockRoundKey] = block
					continue
				}
			case msg := <-actorV1.voteMessageCh:
				actorV1.logger.Info("Receive vote for block", msg.RoundKey, getRoundKey(actorV1.roundData.nextHeight, actorV1.roundData.round))
				validatorIdx := common.IndexOfStr(msg.Validator, actorV1.roundData.committeeBLS.stringList)
				if validatorIdx == -1 {
					continue
				}
				height, round := parseRoundKey(msg.RoundKey)
				if height < actorV1.roundData.nextHeight {
					continue
				}
				if (height == actorV1.roundData.nextHeight) && (round < actorV1.roundData.round) {
					continue
				}
				if (height == actorV1.roundData.nextHeight) && (round == actorV1.roundData.round) {
					//validate single sig
					if !(new(common.Hash).IsEqual(&actorV1.roundData.blockHash)) {
						actorV1.roundData.lockVotes.Lock()
						if _, ok := actorV1.roundData.votes[msg.Validator]; !ok {
							// committeeArr := []incognitokey.CommitteePublicKey{}
							// committeeArr = append(committeeArr, actorV1.RoundData.Committee...)
							actorV1.roundData.lockVotes.Unlock()
							go func(voteMsg BFTVote, blockHash common.Hash, committee []incognitokey.CommitteePublicKey) {
								if err := actorV1.preValidateVote(blockHash.GetBytes(), &(voteMsg.Vote), committee[validatorIdx].MiningPubKey[common.BridgeConsensus]); err != nil {
									actorV1.logger.Error(err)
									return
								}
								if len(voteMsg.Vote.BRI) != 0 {
									if err := validateSingleBriSig(&blockHash, voteMsg.Vote.BRI, committee[validatorIdx].MiningPubKey[common.BridgeConsensus]); err != nil {
										actorV1.logger.Error(err)
										return
									}
								}
								go func() {
									voteCtnBytes, err := json.Marshal(voteMsg)
									if err != nil {
										actorV1.logger.Error(NewConsensusError(UnExpectedError, err))
										return
									}
									msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
									msg.(*wire.MessageBFT).ChainKey = actorV1.ChainKey
									msg.(*wire.MessageBFT).Content = voteCtnBytes
									msg.(*wire.MessageBFT).Type = MSG_VOTE
									// TODO uncomment here when switch to non-highway mode
									// e.Node.PushMessageToChain(msg, e.Chain)
								}()
								actorV1.addVote(voteMsg)
							}(msg, actorV1.RoundData.BlockHash, append([]incognitokey.CommitteePublicKey{}, e.RoundData.Committee...))
							continue
						} else {
							actorV1.RoundData.lockVotes.Unlock()
							continue
						}
					}
				}
				actorV1.addEarlyVote(msg)

			case <-ticker.C:
				monitor.SetGlobalParam("RoundKey", getRoundKey(actorV1.RoundData.NextHeight, actorV1.RoundData.Round), "Phase", actorV1.RoundData.State)
				inCommitteeList := false
				for _, userKey := range actorV1.UserKeySet {
					pubKey := userKey.GetPublicKey()
					if common.IndexOfStr(pubKey.GetMiningKeyBase58(consensusName), actorV1.RoundData.CommitteeBLS.StringList) != -1 {
						inCommitteeList = true
						break
					}
				}

				if !inCommitteeList {
					actorV1.enterNewRound()
					continue
				}

				if !actorV1.Chain.IsReady() {
					actorV1.isOngoing = false
					//fmt.Println("CONSENSUS: ticker 1")
					continue
				}

				if !actorV1.isInTimeFrame() || actorV1.RoundData.State == "" {
					actorV1.enterNewRound()
				}

				switch actorV1.RoundData.State {
				case listenPhase:
					if actorV1.Chain.CurrentHeight() == actorV1.RoundData.NextHeight {
						actorV1.enterNewRound()
						continue
					}
					roundKey := getRoundKey(actorV1.RoundData.NextHeight, actorV1.RoundData.Round)

					if actorV1.Blocks[roundKey] != nil {
						monitor.SetGlobalParam("ReceiveBlockTime", time.Since(actorV1.RoundData.TimeStart).Seconds())
						if err := actorV1.Chain.ValidatePreSignBlock(actorV1.Blocks[roundKey], []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}); err != nil {
							delete(actorV1.Blocks, roundKey)
							actorV1.logger.Error(err)
							continue
						}

						if actorV1.RoundData.Block == nil {
							actorV1.RoundData.Block = actorV1.Blocks[roundKey]
							actorV1.RoundData.BlockHash = *actorV1.RoundData.Block.Hash()
							valData, err := DecodeValidationData(actorV1.RoundData.Block.GetValidationField())
							if err != nil {
								actorV1.logger.Error(err)
								continue
							}
							actorV1.RoundData.BlockValidateData = *valData
							actorV1.enterVotePhase()
						}
					}
				case votePhase:
					actorV1.logger.Info("Case: In vote phase")
					if actorV1.RoundData.NotYetSendVote {
						err := actorV1.sendVote()
						if err != nil {
							actorV1.logger.Error(err)
							continue
						}
					}
					if !(new(common.Hash).IsEqual(&actorV1.RoundData.BlockHash)) && actorV1.isHasMajorityVotes() {
						actorV1.RoundData.lockVotes.Lock()
						aggSig, brigSigs, validatorIdx, err := combineVotes(actorV1.RoundData.Votes, actorV1.RoundData.CommitteeBLS.StringList)
						actorV1.RoundData.lockVotes.Unlock()
						if err != nil {
							actorV1.logger.Error(err)
							continue
						}

						actorV1.RoundData.BlockValidateData.AggSig = aggSig
						actorV1.RoundData.BlockValidateData.BridgeSig = brigSigs
						actorV1.RoundData.BlockValidateData.ValidatiorsIdx = validatorIdx

						validationDataString, _ := EncodeValidationData(actorV1.RoundData.BlockValidateData)
						actorV1.RoundData.Block.(blockValidation).AddValidationField(validationDataString)

						//TODO: check issue invalid sig when swap
						//TODO 0xakk0r0kamui trace who is malicious node if ValidateCommitteeSig return false
						err = ValidateCommitteeSig(actorV1.RoundData.Block, actorV1.RoundData.Committee)
						if err != nil {
							actorV1.logger.Error(err)
							actorV1.logger.Errorf("actorV1.RoundData.Block.GetValidationField()=%+v\n", actorV1.RoundData.Block.GetValidationField())
							actorV1.logger.Errorf("actorV1.RoundData.Committee=%+v\n", actorV1.RoundData.Committee)
							for _, member := range actorV1.RoundData.Committee {
								actorV1.logger.Errorf("member.MiningPubKey[%+v] %+v\n", consensusName, base58.Base58Check{}.Encode(member.MiningPubKey[consensusName], common.Base58Version))
							}
							continue
						}

						if err := actorV1.Chain.InsertAndBroadcastBlock(actorV1.RoundData.Block); err != nil {
							actorV1.logger.Error(err)
							if blockchainError, ok := err.(*blockchain.BlockChainError); ok {
								if blockchainError.Code != blockchain.ErrCodeMessage[blockchain.DuplicateShardBlockError].Code {
									actorV1.logger.Error(err)
								}
							}
							continue
						}
						monitor.SetGlobalParam("CommitTime", time.Since(time.Unix(e.Chain.GetLastBlockTimeStamp(), 0)).Seconds())
						// e.Node.PushMessageToAll()
						actorV1.logger.Infof("Commit block (%d votes) %+v hash=%+v \n Wait for next round", len(e.RoundData.Votes), e.RoundData.Block.GetHeight(), e.RoundData.Block.Hash().String())
						actorV1.enterNewRound()
					}
				}
			}
		}
	}()
	return nil
}

func (actorV1 *ActorV1) enterProposePhase(keyset *signatureschemes2.MiningKey) {
	if !actorV1.isInTimeFrame() || actorV1.RoundData.State == proposePhase {
		return
	}
	actorV1.setState(proposePhase)
	actorV1.isOngoing = true
	block, err := actorV1.createNewBlock(keyset)
	monitor.SetGlobalParam("CreateTime", time.Since(actorV1.RoundData.TimeStart).Seconds())
	if err != nil {
		actorV1.isOngoing = false
		actorV1.logger.Error("can't create block", err)
		return
	}

	if actorV1.Chain.CurrentHeight()+1 != block.GetHeight() {
		return
	}
	var validationData ValidationData
	validationData.ProducerBLSSig, _ = keyset.BriSignData(block.Hash().GetBytes())
	validationDataString, err := EncodeValidationData(validationData)
	if err != nil {
		actorV1.logger.Errorf("Encode validation data failed %+v", err)
	}
	block.(blockValidation).AddValidationField(validationDataString)

	actorV1.RoundData.Block = block
	actorV1.RoundData.BlockHash = *block.Hash()
	actorV1.RoundData.BlockValidateData = validationData

	blockData, _ := json.Marshal(actorV1.RoundData.Block)
	msg, _ := MakeBFTProposeMsg(blockData, actorV1.ChainKey, keyset)
	go actorV1.Node.PushMessageToChain(msg, actorV1.Chain)
}

func (actorV1 *ActorV1) enterListenPhase() {
	if !actorV1.isInTimeFrame() || actorV1.RoundData.State == listenPhase {
		return
	}
	actorV1.setState(listenPhase)
}

func (actorV1 *ActorV1) enterVotePhase() {
	if !actorV1.isInTimeFrame() || actorV1.RoundData.State == votePhase || actorV1.RoundData.Block == nil {
		return
	}
	actorV1.logger.Info("enter voting phase")
	actorV1.isOngoing = true
	actorV1.setState(votePhase)
	err := actorV1.sendVote()
	if err != nil {
		actorV1.logger.Error(err)
		return
	}
	actorV1.logger.Info(actorV1.ChainKey, "sending vote...")
}

func (actorV1 *ActorV1) enterNewRound() {
	//if chain is not ready,  return
	if !actorV1.Chain.IsReady() {
		actorV1.RoundData.State = ""
		return
	}
	//if already running a round for current timeframe
	if actorV1.isInTimeFrame() && (actorV1.RoundData.State != newround && actorV1.RoundData.State != "") {
		fmt.Println("CONSENSUS", actorV1.isInTimeFrame(), actorV1.getCurrentRound(), actorV1.getTimeSinceLastBlock().Seconds(), actorV1.RoundData.State)
		return
	}

	actorV1.isOngoing = false
	actorV1.setState("")
	if actorV1.waitForNextRound() {
		return
	}
	actorV1.setState(newround)
	actorV1.InitRoundData()
	actorV1.logger.Info("")
	actorV1.logger.Info("============================================")
	actorV1.logger.Info("")

	for _, userKey := range actorV1.UserKeySet {
		pubKey := userKey.GetPublicKey()
		if actorV1.Chain.GetPubKeyCommitteeIndex(pubKey.GetMiningKeyBase58(consensusName)) == GetProposerIndexByRound(actorV1.Chain.GetLastProposerIndex(), actorV1.RoundData.Round, actorV1.Chain.GetCommitteeSize()) {
			actorV1.logger.Infof("%v TS: %v, PROPOSE BLOCK %v, Round %v", actorV1.ChainKey, 0, actorV1.RoundData.NextHeight, actorV1.RoundData.Round)
			actorV1.enterProposePhase(&userKey)
			actorV1.enterVotePhase()
			return
		}
	}

	//if not propose => check for listen
	for _, userKey := range actorV1.UserKeySet {
		pubKey := userKey.GetPublicKey()
		if common.IndexOfStr(pubKey.GetMiningKeyBase58(consensusName), actorV1.RoundData.CommitteeBLS.StringList) != -1 {
			actorV1.logger.Infof("%v TS: %v, LISTEN BLOCK %v, Round %v", actorV1.ChainKey, 0, actorV1.RoundData.NextHeight, actorV1.RoundData.Round)
			actorV1.enterListenPhase()
			break
		}
	}

}

func (actorV1 *ActorV1) addVote(voteMsg BFTVote) {
	actorV1.RoundData.lockVotes.Lock()
	defer actorV1.RoundData.lockVotes.Unlock()
	actorV1.RoundData.Votes[voteMsg.Validator] = voteMsg.Vote
	actorV1.logger.Warn("vote added...")
	return
}

func (actorV1 *ActorV1) addEarlyVote(voteMsg BFTVote) {
	actorV1.lockEarlyVotes.Lock()
	defer actorV1.lockEarlyVotes.Unlock()
	if _, ok := actorV1.EarlyVotes[voteMsg.RoundKey]; !ok {
		actorV1.EarlyVotes[voteMsg.RoundKey] = make(map[string]vote)
	}
	actorV1.EarlyVotes[voteMsg.RoundKey][voteMsg.Validator] = voteMsg.Vote
	return
}

func (actorV1 *ActorV1) createNewBlock(userKey *signatureschemes2.MiningKey) (types.BlockInterface, error) {

	var errCh chan error
	var block types.BlockInterface = nil
	errCh = make(chan error)
	timeout := time.NewTimer(timeout / 2).C

	go func() {
		time1 := time.Now()
		var err error
		commitee := actorV1.Chain.GetCommittee()
		pk := userKey.GetPublicKey()
		base58Str, err := commitee[actorV1.Chain.GetPubKeyCommitteeIndex(pk.GetMiningKeyBase58(consensusName))].ToBase58()
		if err != nil {
			actorV1.logger.Error("UserKeySet is wrong", err)
			errCh <- err
			return
		}

		block, err = actorV1.Chain.CreateNewBlock(1, base58Str, int(actorV1.RoundData.Round), actorV1.RoundData.TimeStart.Unix(), []incognitokey.CommitteePublicKey{}, common.Hash{})
		if block != nil {
			actorV1.logger.Info("create block", block.GetHeight(), time.Since(time1).Seconds())
		} else {
			actorV1.logger.Info("create block", time.Since(time1).Seconds())
		}

		time.AfterFunc(100*time.Millisecond, func() {
			select {
			case <-errCh:
			default:
			}
		})
		errCh <- err
	}()
	select {
	case err := <-errCh:
		return block, err
	case <-timeout:
		if block != nil {
			actorV1.logger.Info("Create block has something wrong ", block.GetHeight())
		}
		return nil, NewConsensusError(BlockCreationError, errors.New("block creation timeout"))
	}
}

func NewInstanceV1WithValue(
	chain ChainInterface,
	chainKey string, chainID int,
	node NodeInterface, logger common.Logger,
) *actorV1 {
	var newInstance actorV1
	newInstance.chain = chain
	newInstance.chainKey = chainKey
	newInstance.chainID = chainID
	newInstance.node = node
	newInstance.logger = logger
	return &newInstance
}
