package blsbft

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metrics/monitor"
	"github.com/incognitochain/incognito-chain/wire"
)

type actorV1 struct {
	actorBase
	roundData struct {
		timeStart         time.Time
		block             types.BlockInterface
		blockHash         common.Hash
		blockValidateData consensustypes.ValidationData
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

func (actorV1 *actorV1) Stop() error {
	err := actorV1.actorBase.Stop()
	if err != nil {
		return NewConsensusError(ConsensusAlreadyStoppedError, err)
	}
	actorV1.isOngoing = false
	return nil
}

func (actorV1 *actorV1) Run() error {
	if actorV1.isStarted {
		return NewConsensusError(ConsensusAlreadyStartedError, errors.New(actorV1.chainKey))
	}
	actorV1.destroyCh = make(chan struct{})
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
								v := vote{
									BLS:          voteMsg.Bls,
									BRI:          voteMsg.Bri,
									Confirmation: voteMsg.Confirmation,
								}
								if err := actorV1.preValidateVote(blockHash.GetBytes(), &v, committee[validatorIdx].MiningPubKey[common.BridgeConsensus]); err != nil {
									actorV1.logger.Error(err)
									return
								}
								if len(voteMsg.Bri) != 0 {
									if err := validateSingleBriSig(&blockHash, voteMsg.Bri, committee[validatorIdx].MiningPubKey[common.BridgeConsensus]); err != nil {
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
									msg.(*wire.MessageBFT).ChainKey = actorV1.chainKey
									msg.(*wire.MessageBFT).Content = voteCtnBytes
									msg.(*wire.MessageBFT).Type = MsgVote
									// TODO uncomment here when switch to non-highway mode
									// e.Node.PushMessageToChain(msg, e.Chain)
								}()
								actorV1.addVote(voteMsg)
							}(msg, actorV1.roundData.blockHash, append([]incognitokey.CommitteePublicKey{}, actorV1.roundData.committee...))
							continue
						} else {
							actorV1.roundData.lockVotes.Unlock()
							continue
						}
					}
				}
				actorV1.addEarlyVote(msg)

			case <-ticker.C:
				monitor.SetGlobalParam("RoundKey", getRoundKey(actorV1.roundData.nextHeight, actorV1.roundData.round), "Phase", actorV1.roundData.state)
				inCommitteeList := false
				for _, userKey := range actorV1.userKeySet {
					pubKey := userKey.GetPublicKey()
					if common.IndexOfStr(pubKey.GetMiningKeyBase58(consensusName), actorV1.roundData.committeeBLS.stringList) != -1 {
						inCommitteeList = true
						break
					}
				}

				if !inCommitteeList {
					actorV1.enterNewRound()
					continue
				}

				if !actorV1.chain.IsReady() {
					actorV1.isOngoing = false
					//fmt.Println("CONSENSUS: ticker 1")
					continue
				}

				if !actorV1.isInTimeFrame() || actorV1.roundData.state == "" {
					actorV1.enterNewRound()
				}

				switch actorV1.roundData.state {
				case listenPhase:
					if actorV1.chain.CurrentHeight() == actorV1.roundData.nextHeight {
						actorV1.enterNewRound()
						continue
					}
					roundKey := getRoundKey(actorV1.roundData.nextHeight, actorV1.roundData.round)
					if actorV1.blocks[roundKey] != nil {
						monitor.SetGlobalParam("ReceiveBlockTime", time.Since(actorV1.roundData.timeStart).Seconds())
						if err := actorV1.chain.ValidatePreSignBlock(actorV1.blocks[roundKey], []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}); err != nil {
							delete(actorV1.blocks, roundKey)
							actorV1.logger.Error(err)
							continue
						}

						if actorV1.roundData.block == nil {
							actorV1.roundData.block = actorV1.blocks[roundKey]
							actorV1.roundData.blockHash = *actorV1.roundData.block.Hash()
							valData, err := consensustypes.DecodeValidationData(actorV1.roundData.block.GetValidationField())
							if err != nil {
								actorV1.logger.Error(err)
								continue
							}
							actorV1.roundData.blockValidateData = *valData
							actorV1.enterVotePhase()
						}
					}
				case votePhase:
					actorV1.logger.Info("Case: In vote phase")
					if actorV1.roundData.notYetSendVote {
						err := actorV1.sendVote()
						if err != nil {
							actorV1.logger.Error(err)
							continue
						}
					}
					if !(new(common.Hash).IsEqual(&actorV1.roundData.blockHash)) && actorV1.isHasMajorityVotes() {
						actorV1.roundData.lockVotes.Lock()
						aggSig, brigSigs, validatorIdx, err := actorV1.combineVotes(actorV1.roundData.votes, actorV1.roundData.committeeBLS.stringList)
						actorV1.roundData.lockVotes.Unlock()
						if err != nil {
							actorV1.logger.Error(err)
							continue
						}

						actorV1.roundData.blockValidateData.AggSig = aggSig
						actorV1.roundData.blockValidateData.BridgeSig = brigSigs
						actorV1.roundData.blockValidateData.ValidatiorsIdx = validatorIdx

						validationDataString, _ := consensustypes.EncodeValidationData(actorV1.roundData.blockValidateData)
						actorV1.roundData.block.(blockValidation).AddValidationField(validationDataString)

						//TODO: check issue invalid sig when swap
						//TODO 0xakk0r0kamui trace who is malicious node if ValidateCommitteeSig return false
						err = ValidateCommitteeSig(actorV1.roundData.block, actorV1.roundData.committee)
						if err != nil {
							actorV1.logger.Error(err)
							actorV1.logger.Errorf("actorV1.RoundData.Block.GetValidationField()=%+v\n", actorV1.roundData.block.GetValidationField())
							actorV1.logger.Errorf("actorV1.RoundData.Committee=%+v\n", actorV1.roundData.committee)
							for _, member := range actorV1.roundData.committee {
								actorV1.logger.Errorf("member.MiningPubKey[%+v] %+v\n", consensusName, base58.Base58Check{}.Encode(member.MiningPubKey[consensusName], common.Base58Version))
							}
							continue
						}

						isBeacon := false
						if actorV1.chain.IsBeaconChain() {
							isBeacon = true
						}
						go actorV1.node.PushBlockToAll(actorV1.roundData.block, "", isBeacon)
						if err := actorV1.chain.InsertBlock(actorV1.roundData.block, false); err != nil {
							actorV1.logger.Error(err)
							if blockchainError, ok := err.(*blockchain.BlockChainError); ok {
								if blockchainError.Code != blockchain.ErrCodeMessage[blockchain.DuplicateShardBlockError].Code {
									actorV1.logger.Error(err)
								}
							}
							continue

						}
						monitor.SetGlobalParam("CommitTime", time.Since(time.Unix(actorV1.chain.GetLastBlockTimeStamp(), 0)).Seconds())
						// e.Node.PushMessageToAll()
						actorV1.logger.Infof("Commit block (%d votes) %+v hash=%+v \n Wait for next round", len(actorV1.roundData.votes), actorV1.roundData.block.GetHeight(), actorV1.roundData.block.Hash().String())
						actorV1.enterNewRound()
					}
				}
			}
		}
	}()
	return nil
}

func (actorV1 *actorV1) enterProposePhase(keyset *signatureschemes2.MiningKey) {
	if !actorV1.isInTimeFrame() || actorV1.roundData.state == proposePhase {
		return
	}
	actorV1.setState(proposePhase)
	actorV1.isOngoing = true
	block, err := actorV1.createNewBlock(keyset)
	monitor.SetGlobalParam("CreateTime", time.Since(actorV1.roundData.timeStart).Seconds())
	if err != nil {
		actorV1.isOngoing = false
		actorV1.logger.Error("can't create block", err)
		return
	}

	if actorV1.chain.CurrentHeight()+1 != block.GetHeight() {
		return
	}
	var validationData consensustypes.ValidationData
	validationData.ProducerBLSSig, _ = keyset.BriSignData(block.Hash().GetBytes())
	validationDataString, err := consensustypes.EncodeValidationData(validationData)
	if err != nil {
		actorV1.logger.Errorf("Encode validation data failed %+v", err)
	}
	block.(blockValidation).AddValidationField(validationDataString)

	actorV1.roundData.block = block
	actorV1.roundData.blockHash = *block.Hash()
	actorV1.roundData.blockValidateData = validationData

	blockData, _ := json.Marshal(actorV1.roundData.block)
	msg, _ := actorV1.makeBFTProposeMsg(blockData, actorV1.chainKey, keyset)
	go actorV1.node.PushMessageToChain(msg, actorV1.chain)
}

func (actorV1 *actorV1) enterListenPhase() {
	if !actorV1.isInTimeFrame() || actorV1.roundData.state == listenPhase {
		return
	}
	actorV1.setState(listenPhase)
}

func (actorV1 *actorV1) enterVotePhase() {
	if !actorV1.isInTimeFrame() || actorV1.roundData.state == votePhase || actorV1.roundData.block == nil {
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
	actorV1.logger.Info(actorV1.chainKey, "sending vote...")
}

func (actorV1 *actorV1) enterNewRound() {
	//if chain is not ready,  return
	if !actorV1.chain.IsReady() {
		actorV1.roundData.state = ""
		return
	}
	//if already running a round for current timeframe
	if actorV1.isInTimeFrame() && (actorV1.roundData.state != newround && actorV1.roundData.state != "") {
		fmt.Println("CONSENSUS", actorV1.isInTimeFrame(), actorV1.getCurrentRound(), actorV1.getTimeSinceLastBlock().Seconds(), actorV1.roundData.state)
		return
	}

	actorV1.isOngoing = false
	actorV1.setState("")
	if actorV1.waitForNextRound() {
		return
	}
	actorV1.setState(newround)
	actorV1.initRoundData()
	actorV1.logger.Info("")
	actorV1.logger.Info("============================================")
	actorV1.logger.Info("")

	for _, userKey := range actorV1.userKeySet {
		pubKey := userKey.GetPublicKey()
		if actorV1.chain.GetPubKeyCommitteeIndex(pubKey.GetMiningKeyBase58(consensusName)) == GetProposerIndexByRound(actorV1.chain.GetLastProposerIndex(), actorV1.roundData.round, actorV1.chain.GetCommitteeSize()) {
			actorV1.logger.Infof("%v TS: %v, PROPOSE BLOCK %v, Round %v", actorV1.chainKey, 0, actorV1.roundData.nextHeight, actorV1.roundData.round)
			actorV1.enterProposePhase(&userKey)
			actorV1.enterVotePhase()
			return
		}
	}

	//if not propose => check for listen
	for _, userKey := range actorV1.userKeySet {
		pubKey := userKey.GetPublicKey()
		if common.IndexOfStr(pubKey.GetMiningKeyBase58(consensusName), actorV1.roundData.committeeBLS.stringList) != -1 {
			actorV1.logger.Infof("%v TS: %v, LISTEN BLOCK %v, Round %v", actorV1.chainKey, 0, actorV1.roundData.nextHeight, actorV1.roundData.round)
			actorV1.enterListenPhase()
			break
		}
	}

}

func (actorV1 *actorV1) addVote(voteMsg BFTVote) {
	actorV1.roundData.lockVotes.Lock()
	defer actorV1.roundData.lockVotes.Unlock()
	v := vote{
		BLS:          voteMsg.Bls,
		BRI:          voteMsg.Bri,
		Confirmation: voteMsg.Confirmation,
	}
	actorV1.roundData.votes[voteMsg.Validator] = v
	actorV1.logger.Warn("vote added...")
	return
}

func (actorV1 *actorV1) addEarlyVote(voteMsg BFTVote) {
	actorV1.lockEarlyVotes.Lock()
	defer actorV1.lockEarlyVotes.Unlock()
	if _, ok := actorV1.earlyVotes[voteMsg.RoundKey]; !ok {
		actorV1.earlyVotes[voteMsg.RoundKey] = make(map[string]vote)
	}
	v := vote{
		BLS:          voteMsg.Bls,
		BRI:          voteMsg.Bri,
		Confirmation: voteMsg.Confirmation,
	}
	actorV1.earlyVotes[voteMsg.RoundKey][voteMsg.Validator] = v
	return
}

func (actorV1 *actorV1) createNewBlock(userKey *signatureschemes2.MiningKey) (types.BlockInterface, error) {

	var errCh chan error
	var block types.BlockInterface = nil
	errCh = make(chan error)
	timeout := time.NewTimer(timeout / 2).C

	go func() {
		time1 := time.Now()
		var err error
		commitee := actorV1.chain.GetCommittee()
		pk := userKey.GetPublicKey()
		base58Str, err := commitee[actorV1.chain.GetPubKeyCommitteeIndex(pk.GetMiningKeyBase58(consensusName))].ToBase58()
		if err != nil {
			actorV1.logger.Error("UserKeySet is wrong", err)
			errCh <- err
			return
		}

		block, err = actorV1.chain.CreateNewBlock(1, base58Str, int(actorV1.roundData.round), actorV1.roundData.timeStart.Unix(), []incognitokey.CommitteePublicKey{}, common.Hash{})
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

func NewActorV1WithValue(
	chain Chain,
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

func (actorV1 *actorV1) BlockVersion() int {
	return types.BFT_VERSION
}

func (actorV1) sendVote() error {
	panic("not implement")
}

func (actorV1) makeBFTProposeMsg(blockData []byte, chainKey string, keyset *signatureschemes2.MiningKey) (wire.Message, error) {
	panic("not implement")
}

func GetProposerIndexByRound(lastId, round, committeeSize int) int {
	//return (lastId + round) % committeeSize
	return 0
}

func (actorV1 *actorV1) getTimeSinceLastBlock() time.Duration {
	return time.Since(time.Unix(int64(actorV1.chain.GetLastBlockTimeStamp()), 0))
}

func (actorV1 *actorV1) waitForNextRound() bool {
	timeSinceLastBlk := actorV1.getTimeSinceLastBlock()
	if timeSinceLastBlk >= actorV1.chain.GetMinBlkInterval() {
		return false
	} else {
		//fmt.Println("\n\nWait for", e.Chain.GetMinBlkInterval()-timeSinceLastBlk, "\n\n")
		return true
	}
}

func (actorV1 *actorV1) setState(state string) {
	actorV1.roundData.state = state
}

func (actorV1 *actorV1) getCurrentRound() int {
	round := int((actorV1.getTimeSinceLastBlock().Seconds() - float64(actorV1.chain.GetMinBlkInterval().Seconds())) / timeout.Seconds())
	if round < 0 {
		return 1
	}

	return round + 1
}

func (actorV1 *actorV1) isInTimeFrame() bool {
	if actorV1.chain.CurrentHeight()+1 != actorV1.roundData.nextHeight {
		return false
	}

	if actorV1.getCurrentRound() != actorV1.roundData.round {
		return false
	}

	return true
}

func (actorV1 *actorV1) isHasMajorityVotes() bool {
	// e.RoundData.lockVotes.Lock()
	// defer e.RoundData.lockVotes.Unlock()
	actorV1.lockEarlyVotes.Lock()
	defer actorV1.lockEarlyVotes.Unlock()
	roundKey := getRoundKey(actorV1.roundData.nextHeight, actorV1.roundData.round)
	earlyVote, ok := actorV1.earlyVotes[roundKey]
	committeeSize := len(actorV1.roundData.committee)
	if ok {
		wg := sync.WaitGroup{}
		blockHashBytes := actorV1.roundData.blockHash.GetBytes()
		for k, v := range earlyVote {
			wg.Add(1)
			go func(validatorKey string, voteData vote) {
				defer wg.Done()
				validatorIdx := common.IndexOfStr(validatorKey, actorV1.roundData.committeeBLS.stringList)
				if err := actorV1.preValidateVote(blockHashBytes, &voteData, actorV1.roundData.committee[validatorIdx].MiningPubKey[common.BridgeConsensus]); err == nil {
					actorV1.roundData.lockVotes.Lock()
					actorV1.roundData.votes[validatorKey] = voteData
					actorV1.roundData.lockVotes.Unlock()
				} else {
					actorV1.logger.Error(err)
				}
			}(k, v)
		}
		wg.Wait()
		if len(actorV1.roundData.votes) > 2*committeeSize/3 {
			delete(actorV1.earlyVotes, roundKey)
		}
	}
	monitor.SetGlobalParam("NVote", len(actorV1.roundData.votes))
	if len(actorV1.roundData.votes) > 2*committeeSize/3 {
		return true
	}
	return false
}

func getRoundKey(nextHeight uint64, round int) string {
	return fmt.Sprint(nextHeight, "_", round)
}

func parseRoundKey(roundKey string) (uint64, int) {
	stringArray := strings.Split(roundKey, "_")
	if len(stringArray) != 2 {
		return 0, 0
	}
	height, err := strconv.Atoi(stringArray[0])
	if err != nil {
		return 0, 0
	}
	round, err := strconv.Atoi(stringArray[1])
	if err != nil {
		return 0, 0
	}
	return uint64(height), round
}

func ExtractBridgeValidationData(block types.BlockInterface) ([][]byte, []int, error) {
	valData, err := consensustypes.DecodeValidationData(block.GetValidationField())
	if err != nil {
		return nil, nil, NewConsensusError(UnExpectedError, err)
	}
	return valData.BridgeSig, valData.ValidatiorsIdx, nil
}

func (actorV1 *actorV1) UpdateCommitteeBLSList() {
	committee := actorV1.chain.GetCommittee()
	if !reflect.DeepEqual(actorV1.roundData.committee, committee) {
		actorV1.roundData.committee = committee
		actorV1.roundData.committeeBLS.byteList = []blsmultisig.PublicKey{}
		actorV1.roundData.committeeBLS.stringList = []string{}
		for _, member := range actorV1.roundData.committee {
			actorV1.roundData.committeeBLS.byteList = append(actorV1.roundData.committeeBLS.byteList, member.MiningPubKey[consensusName])
		}
		committeeBLSString, err := incognitokey.ExtractPublickeysFromCommitteeKeyList(actorV1.roundData.committee, consensusName)
		if err != nil {
			actorV1.logger.Error(err)
			return
		}
		actorV1.roundData.committeeBLS.stringList = committeeBLSString
	}
}

func (actorV1 *actorV1) initRoundData() {
	roundKey := getRoundKey(actorV1.roundData.nextHeight, actorV1.roundData.round)
	if _, ok := actorV1.blocks[roundKey]; ok {
		delete(actorV1.blocks, roundKey)
	}
	actorV1.roundData.nextHeight = actorV1.chain.CurrentHeight() + 1
	actorV1.roundData.round = actorV1.getCurrentRound()
	actorV1.roundData.votes = make(map[string]vote)
	actorV1.roundData.block = nil
	actorV1.roundData.blockHash = common.Hash{}
	actorV1.roundData.notYetSendVote = true
	actorV1.roundData.timeStart = time.Now()
	actorV1.roundData.lastProposerIndex = actorV1.chain.GetLastProposerIndex()
	actorV1.UpdateCommitteeBLSList()
	actorV1.setState(newround)
}

func NewActorWithValue(
	chain, committeeChain Chain, version int,
	chainID, blockVersion int, chainName string,
	node NodeInterface, logger common.Logger,
) Actor {
	var res Actor
	switch version {
	case types.BFT_VERSION:
		res = NewActorV1WithValue(chain, chainName, chainID, node, logger)
	case types.MULTI_VIEW_VERSION:
		res = NewActorV2WithValue(chain, committeeChain, chainName, chainID, blockVersion, node, logger)
	default:
		panic("Bft version is not valid")
	}
	return res
}
