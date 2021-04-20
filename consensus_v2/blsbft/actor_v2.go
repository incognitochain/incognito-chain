package blsbft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/metrics/monitor"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/wire"
)

type actorV2 struct {
	actorBase
	committeeChain  blockchain.Chain
	currentTime     int64
	currentTimeSlot int64
	proposeHistory  *lru.Cache

	receiveBlockByHeight map[uint64][]*ProposeBlockInfo  //blockHeight -> blockInfo
	receiveBlockByHash   map[string]*ProposeBlockInfo    //blockHash -> blockInfo
	voteHistory          map[uint64]types.BlockInterface // bestview height (previsous height )-> block
	bodyHashes           map[uint64]map[string]bool
	votedTimeslot        map[int64]bool
	blockVersion         int
}

func NewActorV2() *actorV2 {
	return &actorV2{}
}

func NewActorV2WithValue(
	chain blockchain.Chain,
	committeeChain blockchain.Chain,
	chainKey string, blockVersion, chainID int,
	node NodeInterface, logger common.Logger,
) *actorV2 {
	var err error
	res := NewActorV2()
	res.actorBase = *NewActorBaseWithValue(chain, chainKey, chainID, node, logger)
	res.proposeMessageCh = make(chan BFTPropose)
	res.voteMessageCh = make(chan BFTVote)
	res.receiveBlockByHash = make(map[string]*ProposeBlockInfo)
	res.receiveBlockByHeight = make(map[uint64][]*ProposeBlockInfo)
	res.voteHistory = make(map[uint64]types.BlockInterface)
	res.bodyHashes = make(map[uint64]map[string]bool)
	res.votedTimeslot = make(map[int64]bool)
	res.committeeChain = committeeChain
	res.blockVersion = blockVersion
	res.proposeHistory, err = lru.New(1000)
	if err != nil {
		panic(err) //must not error
	}
	return res
}

func (actorV2 *actorV2) Run() error {
	actorV2.isStarted = true
	go func() {
		//init view maps
		ticker := time.Tick(200 * time.Millisecond)
		cleanMemTicker := time.Tick(5 * time.Minute)
		actorV2.logger.Infof("init bls-bft-%+v, consensus for chain %+v", actorV2.blockVersion, actorV2.chainKey)

		for { //actor loop
			select {
			case <-actorV2.destroyCh:
				actorV2.logger.Infof("exit bls-bft-%+v consensus for chain %+v", actorV2.blockVersion, actorV2.chainKey)
				return
			case proposeMsg := <-actorV2.proposeMessageCh:
				err := actorV2.handleProposeMsg(proposeMsg)
				if err != nil {
					actorV2.logger.Debug(err)
					continue
				}

			case voteMsg := <-actorV2.voteMessageCh:
				err := actorV2.handleVoteMsg(voteMsg)
				if err != nil {
					actorV2.logger.Debug(err)
					continue
				}

			case <-cleanMemTicker:

				for h, _ := range actorV2.receiveBlockByHeight {
					if h <= actorV2.chain.GetFinalView().GetHeight() {
						delete(actorV2.bodyHashes, h)
					}
				}

				for h, _ := range actorV2.receiveBlockByHeight {
					if h <= actorV2.chain.GetFinalView().GetHeight() {
						delete(actorV2.receiveBlockByHeight, h)
					}
				}

				for h, _ := range actorV2.voteHistory {
					if h <= actorV2.chain.GetFinalView().GetHeight() {
						delete(actorV2.voteHistory, h)
					}
				}

				for h, proposeBlk := range actorV2.receiveBlockByHash {
					if time.Now().Sub(proposeBlk.receiveTime) > time.Minute {
						delete(actorV2.votedTimeslot, proposeBlk.block.GetProposeTime())
						delete(actorV2.receiveBlockByHash, h)
					}
				}

			case <-ticker:
				if !actorV2.chain.IsReady() {
					continue
				}
				actorV2.currentTime = time.Now().Unix()
				currentTimeSlot := common.CalculateTimeSlot(actorV2.currentTime)

				newTimeSlot := false
				if actorV2.currentTimeSlot != currentTimeSlot {
					newTimeSlot = true
				}

				actorV2.currentTimeSlot = currentTimeSlot
				bestView := actorV2.chain.GetBestView()

				//set round for monitor
				round := actorV2.currentTimeSlot - common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime())

				monitor.SetGlobalParam("RoundKey", fmt.Sprintf("%d_%d", bestView.GetHeight(), round))

				var userProposeKey signatureschemes2.MiningKey
				shouldPropose := false
				shouldListen := true

				signingCommittees, committees, proposerPk, committeeViewHash, err := actorV2.getCommitteesAndCommitteeViewHash()
				if err != nil {
					actorV2.logger.Info(err)
					continue
				}

				userKeySet := actorV2.getUserKeySetForSigning(signingCommittees, actorV2.userKeySet)
				for _, userKey := range userKeySet {
					userPk := userKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus)
					if proposerPk.GetMiningKeyBase58(common.BlsConsensus) == userPk {
						shouldListen = false
						if common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()) != actorV2.currentTimeSlot { // current timeslot is not add to view, and this user is proposer of this timeslot
							//using block hash as key of best view -> check if this best view we propose or not
							if _, ok := actorV2.proposeHistory.Get(fmt.Sprintf("%d", actorV2.currentTimeSlot)); !ok {
								shouldPropose = true
								userProposeKey = userKey
							}
						}
					}
				}

				if newTimeSlot { //for logging
					actorV2.logger.Info("")
					actorV2.logger.Info("======================================================")
					actorV2.logger.Info("")
					if shouldListen {
						actorV2.logger.Infof("%v TS: %v, LISTEN BLOCK %v, Round %v", actorV2.chainKey, common.CalculateTimeSlot(actorV2.currentTime), bestView.GetHeight()+1, round)
					}
					if shouldPropose {
						actorV2.logger.Infof("%v TS: %v, PROPOSE BLOCK %v, Round %v", actorV2.chainKey, common.CalculateTimeSlot(actorV2.currentTime), bestView.GetHeight()+1, round)
					}
				}

				if shouldPropose {
					actorV2.proposeHistory.Add(fmt.Sprintf("%d", actorV2.currentTimeSlot), 1)
					//Proposer Rule: check propose block connected to bestview(longest chain rule 1) and re-propose valid block with smallest timestamp (including already propose in the past) (rule 2)
					sort.Slice(actorV2.receiveBlockByHeight[bestView.GetHeight()+1], func(i, j int) bool {
						return actorV2.receiveBlockByHeight[bestView.GetHeight()+1][i].block.GetProduceTime() < actorV2.receiveBlockByHeight[bestView.GetHeight()+1][j].block.GetProduceTime()
					})

					var proposeBlock types.BlockInterface = nil
					for _, v := range actorV2.receiveBlockByHeight[bestView.GetHeight()+1] {
						if v.isValid {
							proposeBlock = v.block
							break
						}
					}

					if createdBlk, err := actorV2.proposeBlock(userProposeKey, proposerPk, proposeBlock, committees, committeeViewHash); err != nil {
						actorV2.logger.Critical(UnExpectedError, errors.New("can't propose block"))
						actorV2.logger.Critical(err)
					} else {
						actorV2.logger.Infof("proposer block %v round %v time slot %v blockTimeSlot %v with hash %v", createdBlk.GetHeight(), createdBlk.GetRound(), actorV2.currentTimeSlot, common.CalculateTimeSlot(createdBlk.GetProduceTime()), createdBlk.Hash().String())
					}
				}

				validProposeBlocks := actorV2.getValidProposeBlocks(bestView)
				for _, v := range validProposeBlocks {
					if err := actorV2.validateBlock(bestView, v); err == nil {
						err = actorV2.voteForBlock(v)
						if err != nil {
							actorV2.logger.Debug(err)
						}
					}
				}

				/*
					Check for 2/3 vote to commit
				*/
				for k, v := range actorV2.receiveBlockByHash {
					actorV2.processIfBlockGetEnoughVote(k, v)
				}
			}
		}
	}()
	return nil
}

func (actorV2 *actorV2) getValidatorIndex(committees []incognitokey.CommitteePublicKey, validator string) (int, *incognitokey.CommitteePublicKey) {
	for id, c := range committees {
		if validator == c.GetMiningKeyBase58(common.BlsConsensus) {
			return id, &c
		}
	}
	return -1, nil
}

func (actorV2 *actorV2) processIfBlockGetEnoughVote(
	blockHash string, v *ProposeBlockInfo,
) {
	//no vote
	if v.hasNewVote == false {
		return
	}

	//no block
	if v.block == nil {
		return
	}
	actorV2.logger.Infof("Process Block With enough votes, %+v, %+v", *v.block.Hash(), v.block.GetHeight())
	//already in chain
	bestView := actorV2.chain.GetBestView()
	view := actorV2.chain.GetViewByHash(*v.block.Hash())
	if view != nil && bestView.GetHash().String() != view.GetHash().String() {
		//e.Logger.Infof("Get View By Hash Fail, %+v, %+v", *v.block.Hash(), v.block.GetHeight())
		return
	}

	//not connected previous block
	view = actorV2.chain.GetViewByHash(v.block.GetPrevHash())
	if view == nil {
		//e.Logger.Infof("Get Previous View By Hash Fail, %+v, %+v", v.block.GetPrevHash(), v.block.GetHeight()-1)
		return
	}
	v = actorV2.validateVotes(v)

	if !v.isCommitted {
		if v.validVotes > 2*len(v.signingCommittes)/3 {
			v.isCommitted = true
			actorV2.logger.Infof("Commit block %v , height: %v", blockHash, v.block.GetHeight())
			err := actorV2.processWithEnoughVotes(v)
			if err != nil {
				actorV2.logger.Error(err)
				return
			}
		}
	}
}

func (actorV2 *actorV2) validateVotes(v *ProposeBlockInfo) *ProposeBlockInfo {
	validVote := 0
	errVote := 0

	committees := make(map[string]int)
	if len(v.votes) != 0 {
		for i, v := range v.signingCommittes {
			committees[v.GetMiningKeyBase58(common.BlsConsensus)] = i
		}
	}

	for id, vote := range v.votes {
		dsaKey := []byte{}
		if vote.IsValid == 0 {
			if value, ok := committees[vote.Validator]; ok {
				dsaKey = v.signingCommittes[value].MiningPubKey[common.BridgeConsensus]
			} else {
				actorV2.logger.Error("Receive vote from nonCommittee member")
				continue
			}
			if len(dsaKey) == 0 {
				actorV2.logger.Error("canot find dsa key")
				continue
			}

			err := vote.validateVoteOwner(dsaKey)
			if err != nil {
				actorV2.logger.Error(dsaKey)
				actorV2.logger.Error(err)
				v.votes[id].IsValid = -1
				errVote++
			} else {
				v.votes[id].IsValid = 1
				validVote++
			}
		} else {
			validVote++
		}
	}
	actorV2.logger.Info("Number of Valid Vote", validVote, "| Number Of Error Vote", errVote)
	v.hasNewVote = false
	//TODO: @tin/0xkumi check here again
	for key, value := range v.votes {
		if value.IsValid == -1 {
			delete(v.votes, key)
		}
	}

	v.addBlockInfo(
		v.block,
		v.committees,
		v.signingCommittes,
		v.userKeySet,
		v.proposerMiningKeyBase58,
		validVote, errVote,
	)

	return v
}

func (actorV2 *actorV2) processWithEnoughVotes(v *ProposeBlockInfo) error {

	validationData, err := actorV2.createBLSAggregatedSignatures(v.signingCommittes, v.block.GetValidationField(), v.votes)
	if err != nil {
		actorV2.logger.Error(err)
		return err
	}
	v.block.(blockValidation).AddValidationField(validationData)

	newPreviousValidationData := ""
	isBeacon := false

	if actorV2.chain.IsBeaconChain() {
		isBeacon = true
		delete(actorV2.receiveBlockByHash, v.block.GetPrevHash().String())
	} else {
		// validate and previous block
		if previousProposeBlockInfo, ok := actorV2.receiveBlockByHash[v.block.GetPrevHash().String()]; ok &&
			previousProposeBlockInfo != nil && previousProposeBlockInfo.block != nil {
			previousValidationData, err := actorV2.createBLSAggregatedSignatures(
				previousProposeBlockInfo.signingCommittes,
				previousProposeBlockInfo.block.GetValidationField(),
				previousProposeBlockInfo.votes)
			if err != nil {
				actorV2.logger.Error(err)
				return err
			}
			previousProposeBlockInfo = actorV2.validateVotes(previousProposeBlockInfo)
			newPreviousValidationData = previousValidationData

			previousProposeBlockInfo.block.(blockValidation).AddValidationField(previousValidationData)
			if err := actorV2.chain.ReplacePreviousValidationData(v.block.GetPrevHash(), previousValidationData); err != nil {
				return err
			}

			delete(actorV2.receiveBlockByHash, previousProposeBlockInfo.block.GetPrevHash().String())
		}
	}

	if len(v.userKeySet) != 0 {
		go actorV2.node.PushBlockToAll(v.block, newPreviousValidationData, isBeacon)
	}
	if err := actorV2.chain.InsertBlock(v.block, false); err != nil {
		return err
	}

	return nil
}

func (actorV2 *actorV2) createBLSAggregatedSignatures(committees []incognitokey.CommitteePublicKey, tempValidationData string, votes map[string]*BFTVote) (string, error) {
	committeeBLSString, err := incognitokey.ExtractPublickeysFromCommitteeKeyList(committees, common.BlsConsensus)
	if err != nil {
		return "", err
	}
	aggSig, brigSigs, validatorIdx, err := actorV2.combineVotes(votes, committeeBLSString)
	if err != nil {
		return "", err
	}

	valData, err := consensustypes.DecodeValidationData(tempValidationData)
	if err != nil {
		return "", err
	}

	valData.AggSig = aggSig
	valData.BridgeSig = brigSigs
	valData.ValidatiorsIdx = validatorIdx
	validationData, _ := consensustypes.EncodeValidationData(*valData)
	return validationData, err
}

func (actorV2 *actorV2) voteForBlock(
	v *ProposeBlockInfo,
) error {
	for _, userKey := range actorV2.userKeySet {
		Vote, err := actorV2.createVote(&userKey, v.block, v.signingCommittes)
		if err != nil {
			actorV2.logger.Error(err)
			return NewConsensusError(UnExpectedError, err)
		}

		msg, err := actorV2.makeBFTVoteMsg(Vote, actorV2.chainKey, actorV2.currentTimeSlot, v.block.GetHeight())
		if err != nil {
			actorV2.logger.Error(err)
			return NewConsensusError(UnExpectedError, err)
		}

		v.isVoted = true
		actorV2.voteHistory[v.block.GetHeight()] = v.block
		actorV2.votedTimeslot[common.CalculateTimeSlot(v.block.GetProposeTime())] = true
		actorV2.logger.Info(actorV2.chainKey, "sending vote...")
		go actorV2.processBFTMsg(msg.(*wire.MessageBFT))
		go actorV2.node.PushMessageToChain(msg, actorV2.chain)
	}

	return nil
}

func (actorV2 *actorV2) createVote(
	userKey *signatureschemes2.MiningKey,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey) (*BFTVote, error) {
	var vote = new(BFTVote)
	bytelist := []blsmultisig.PublicKey{}
	selfIdx := 0
	userBLSPk := userKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus)
	for i, v := range committees {
		if v.GetMiningKeyBase58(common.BlsConsensus) == userBLSPk {
			selfIdx = i
		}
		bytelist = append(bytelist, v.MiningPubKey[common.BlsConsensus])
	}

	blsSig, err := userKey.BLSSignData(block.Hash().GetBytes(), selfIdx, bytelist)
	if err != nil {

		return nil, NewConsensusError(UnExpectedError, err)
	}
	bridgeSig := []byte{}
	if metadata.HasBridgeInstructions(block.GetInstructions()) {
		bridgeSig, err = userKey.BriSignData(block.Hash().GetBytes())
		if err != nil {
			return nil, NewConsensusError(UnExpectedError, err)
		}
	}
	vote.Bls = blsSig
	vote.Bri = bridgeSig
	vote.BlockHash = block.Hash().String()

	userPk := userKey.GetPublicKey()
	vote.Validator = userPk.GetMiningKeyBase58(common.BlsConsensus)
	vote.PrevBlockHash = block.GetPrevHash().String()
	err = vote.signVote(userKey)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	return vote, nil
}

func (actorV2 *actorV2) proposeBlock(
	userMiningKey signatureschemes2.MiningKey,
	proposerPk incognitokey.CommitteePublicKey,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
) (types.BlockInterface, error) {
	time1 := time.Now()
	b58Str, _ := proposerPk.ToBase58()
	var err error

	if actorV2.chain.IsBeaconChain() {
		block, err = actorV2.proposeBeaconBlock(
			b58Str,
			block,
			committees,
			committeeViewHash,
		)
	} else {
		block, err = actorV2.proposeShardBlock(
			b58Str,
			block,
			committees,
			committeeViewHash,
		)
	}

	if err != nil {
		return nil, NewConsensusError(BlockCreationError, err)
	}

	if block != nil {
		actorV2.logger.Infof("[dcs] create block %v hash %v, propose time %v, produce time %v", block.GetHeight(), block.Hash().String(), block.(types.BlockInterface).GetProposeTime(), block.(types.BlockInterface).GetProduceTime())
	} else {
		actorV2.logger.Infof("create block fail, time: %v", time.Since(time1).Seconds())
		return nil, NewConsensusError(BlockCreationError, errors.New("block is nil"))
	}

	var validationData consensustypes.ValidationData
	validationData.ProducerBLSSig, _ = userMiningKey.BriSignData(block.Hash().GetBytes())
	validationDataString, _ := consensustypes.EncodeValidationData(validationData)
	block.(blockValidation).AddValidationField(validationDataString)
	blockData, _ := json.Marshal(block)

	var proposeCtn = new(BFTPropose)
	proposeCtn.Block = blockData
	proposeCtn.PeerID = actorV2.node.GetSelfPeerID().String()
	msg, _ := actorV2.makeBFTProposeMsg(proposeCtn, actorV2.chainKey, actorV2.currentTimeSlot, block.GetHeight())
	go actorV2.processBFTMsg(msg.(*wire.MessageBFT))
	go actorV2.node.PushMessageToChain(msg, actorV2.chain)

	return block, nil
}

func (actorV2 *actorV2) proposeBeaconBlock(
	b58Str string,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
) (types.BlockInterface, error) {
	var err error
	if block == nil {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, (time.Duration(common.TIMESLOT)*time.Second)/2)
		defer cancel()
		actorV2.logger.Info("CreateNewBlock")
		block, err = actorV2.chain.CreateNewBlock(actorV2.blockVersion, b58Str, 1, actorV2.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	} else {
		actorV2.logger.Infof("CreateNewBlockFromOldBlock, Block Height %+v")
		block, err = actorV2.chain.CreateNewBlockFromOldBlock(block, b58Str, actorV2.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	}
	return block, err
}

func (actorV2 *actorV2) proposeShardBlock(
	b58Str string,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
) (types.BlockInterface, error) {
	var err error
	var newBlock types.BlockInterface
	var committeesFromBeaconHash []incognitokey.CommitteePublicKey
	if block != nil {
		_, committeesFromBeaconHash, err = actorV2.getCommitteeForBlock(block)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	}

	// propose new block when
	// no previous proposed block
	// or previous proposed block has different committee with new committees
	if block == nil ||
		(block != nil && !reflect.DeepEqual(committeesFromBeaconHash, committees)) {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, (time.Duration(common.TIMESLOT)*time.Second)/2)
		defer cancel()
		actorV2.logger.Info("CreateNewBlock")
		newBlock, err = actorV2.chain.CreateNewBlock(actorV2.blockVersion, b58Str, 1, actorV2.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	} else {
		actorV2.logger.Infof("[dcs] CreateNewBlockFromOldBlock, Block Height %+v hash %+v", block.GetHeight(), block.Hash().String())
		newBlock, err = actorV2.chain.CreateNewBlockFromOldBlock(block, b58Str, actorV2.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	}
	return newBlock, err
}

func (actorV2 *actorV2) preValidateVote(blockHash []byte, vote *BFTVote, candidate []byte) error {
	data := []byte{}
	data = append(data, blockHash...)
	data = append(data, vote.Bls...)
	data = append(data, vote.Bri...)
	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, vote.Confirmation, candidate)
	return err
}

func (actorV2 *actorV2) getCommitteeForBlock(v types.BlockInterface) ([]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error) {
	var err error
	var committees, signingCommittees []incognitokey.CommitteePublicKey

	if actorV2.blockVersion == MultiViewsVersion || actorV2.chain.IsBeaconChain() {
		committees = actorV2.chain.GetBestView().GetCommittee()
		signingCommittees = committees
	} else {
		signingCommittees, committees, err = actorV2.
			committeeChain.
			CommitteesFromViewHashForShard(
				v.CommitteeFromBlock(),
				v.SubsetCommitteesFromBlock(),
				byte(actorV2.chain.GetShardID()),
				blockchain.MaxSubsetCommittees,
			)
	}
	return signingCommittees, committees, err
}

func (actorV2 *actorV2) sendVote(userKey *signatureschemes2.MiningKey, block types.BlockInterface, committees []incognitokey.CommitteePublicKey) error {
	Vote, err := actorV2.createVote(userKey, block, committees)
	if err != nil {
		actorV2.logger.Error(err)
		return NewConsensusError(UnExpectedError, err)
	}

	msg, err := actorV2.makeBFTVoteMsg(Vote, actorV2.chainKey, actorV2.currentTimeSlot, block.GetHeight())
	if err != nil {
		actorV2.logger.Error(err)
		return NewConsensusError(UnExpectedError, err)
	}
	actorV2.voteHistory[block.GetHeight()] = block
	actorV2.logger.Info(actorV2.chainKey, "sending vote...")
	go actorV2.node.PushMessageToChain(msg, actorV2.chain)
	return nil
}

func (actorV2 *actorV2) getUserKeySetForSigning(
	committees []incognitokey.CommitteePublicKey, userKeySet []signatureschemes2.MiningKey,
) []signatureschemes2.MiningKey {
	res := []signatureschemes2.MiningKey{}
	if actorV2.chain.IsBeaconChain() {
		res = userKeySet
	} else {
		validCommittees := make(map[string]bool)
		for _, v := range committees {
			key := v.GetMiningKeyBase58(common.BlsConsensus)
			validCommittees[key] = true
		}
		for _, userKey := range userKeySet {
			userPk := userKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus)
			if validCommittees[userPk] {
				res = append(res, userKey)
			}
		}
	}
	return res
}

func (actorV2 *actorV2) getCommitteesAndCommitteeViewHash() (
	[]incognitokey.CommitteePublicKey,
	[]incognitokey.CommitteePublicKey,
	incognitokey.CommitteePublicKey, common.Hash, error,
) {
	committeeViewHash := common.Hash{}
	committees := []incognitokey.CommitteePublicKey{}
	signingCommittees := []incognitokey.CommitteePublicKey{}
	proposerPk := incognitokey.CommitteePublicKey{}
	var err error

	if actorV2.blockVersion == MultiViewsVersion || actorV2.chain.IsBeaconChain() {
		proposerPk, _ = actorV2.chain.GetBestView().GetProposerByTimeSlot(actorV2.currentTimeSlot, 2)
		committees = actorV2.chain.GetBestView().GetCommittee()
		signingCommittees = committees
	} else {
		committeeViewHash = *actorV2.committeeChain.GetFinalView().GetHash()
		subsetViewHash := committeeViewHash
		signingCommittees, committees, err = actorV2.
			committeeChain.
			CommitteesFromViewHashForShard(committeeViewHash, subsetViewHash, byte(actorV2.chainID), blockchain.MaxSubsetCommittees)
		if err != nil {
			return signingCommittees, committees, proposerPk, committeeViewHash, err
		}
		proposerPk = actorV2.committeeChain.ProposerByTimeSlot(byte(actorV2.chainID), actorV2.currentTimeSlot, committees)
	}

	return signingCommittees, committees, proposerPk, committeeViewHash, err
}

func (actorV2 *actorV2) handleProposeMsg(proposeMsg BFTPropose) error {
	blockIntf, err := actorV2.chain.UnmarshalBlock(proposeMsg.Block)
	if err != nil || blockIntf == nil {
		actorV2.logger.Debug(err)
		return err
	}
	block := blockIntf.(types.BlockInterface)
	blkHash := block.Hash().String()

	signingCommittees, committees, err := actorV2.getCommitteeForBlock(block)
	if err != nil {
		actorV2.logger.Error(err)
		return err
	}

	userKeySet := actorV2.getUserKeySetForSigning(signingCommittees, actorV2.userKeySet)
	if len(userKeySet) == 0 {
		actorV2.logger.Debug("Not in round for voting")
	}

	blkCPk := incognitokey.CommitteePublicKey{}
	blkCPk.FromBase58(block.GetProducer())
	proposerMiningKeyBas58 := blkCPk.GetMiningKeyBase58(actorV2.GetConsensusName())

	if v, ok := actorV2.receiveBlockByHash[blkHash]; !ok {
		proposeBlockInfo := newProposeBlockForProposeMsg(
			block, committees, signingCommittees, userKeySet, proposerMiningKeyBas58)
		actorV2.receiveBlockByHash[blkHash] = proposeBlockInfo
		actorV2.logger.Info("Receive block ", block.Hash().String(), "height", block.GetHeight(), ",block timeslot ", common.CalculateTimeSlot(block.GetProposeTime()))
		actorV2.receiveBlockByHeight[block.GetHeight()] = append(actorV2.receiveBlockByHeight[block.GetHeight()], actorV2.receiveBlockByHash[blkHash])
	} else {
		actorV2.receiveBlockByHash[blkHash].addBlockInfo(
			block, committees, signingCommittees, userKeySet, proposerMiningKeyBas58, v.validVotes, v.errVotes)
	}

	if block.GetHeight() <= actorV2.chain.GetBestViewHeight() {
		actorV2.logger.Debug("Receive block create from old view. Rejected!")
		return err
	}

	proposeView := actorV2.chain.GetViewByHash(block.GetPrevHash())
	if proposeView == nil {
		actorV2.logger.Infof("Request sync block from node %s from %s to %s", proposeMsg.PeerID, block.GetPrevHash().String(), block.GetPrevHash().Bytes())
		actorV2.node.RequestMissingViewViaStream(proposeMsg.PeerID, [][]byte{block.GetPrevHash().Bytes()}, actorV2.chain.GetShardID(), actorV2.chain.GetChainName())
	}
	return nil
}

func (actorV2 *actorV2) handleVoteMsg(voteMsg BFTVote) error {
	voteMsg.IsValid = 0
	if b, ok := actorV2.receiveBlockByHash[voteMsg.BlockHash]; ok { //if receiveblock is already initiated
		if _, ok := b.votes[voteMsg.Validator]; !ok { // and not receive validatorA vote
			b.votes[voteMsg.Validator] = &voteMsg // store it
			vid, v := actorV2.getValidatorIndex(b.signingCommittes, voteMsg.Validator)
			if v != nil {
				vbase58, _ := v.ToBase58()
				actorV2.logger.Infof("%v Receive vote (%d) for block %s from validator %d %v", actorV2.chainKey, len(actorV2.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, vid, vbase58)
			} else {
				actorV2.logger.Infof("%v Receive vote (%d) for block %v from unknown validator %v", actorV2.chainKey, len(actorV2.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, voteMsg.Validator)
			}
			b.hasNewVote = true
		}

		if !b.proposerSendVote {
			for _, userKey := range actorV2.userKeySet {
				pubKey := userKey.GetPublicKey()
				if b.block != nil && pubKey.GetMiningKeyBase58(actorV2.GetConsensusName()) == b.proposerMiningKeyBase58 { // if this node is proposer and not sending vote
					var err error
					if err = actorV2.validateBlock(actorV2.chain.GetBestView(), b); err != nil {
						err = actorV2.voteForBlock(b)
						if err != nil {
							actorV2.logger.Debug(err)
						}
					} else {
						actorV2.logger.Debug(err)
					}
					if err == nil {
						bestViewHeight := actorV2.chain.GetBestView().GetHeight()
						if b.block.GetHeight() == bestViewHeight+1 { // and if the propose block is still connected to bestview
							err := actorV2.sendVote(&userKey, b.block, b.signingCommittes) // => send vote
							if err != nil {
								actorV2.logger.Error(err)
							} else {
								b.proposerSendVote = true
								b.sendVote = true
							}
						}
					}
				}
			}
		}
	} else {
		actorV2.receiveBlockByHash[voteMsg.BlockHash] = newBlockInfoForVoteMsg()
		actorV2.receiveBlockByHash[voteMsg.BlockHash].votes[voteMsg.Validator] = &voteMsg
		actorV2.logger.Infof("%v Receive vote (%d) for block %v from unknown validator %v", actorV2.chainKey, len(actorV2.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, voteMsg.Validator)
	}
	return nil
}

func (actorV2 *actorV2) getValidProposeBlocks(bestView multiview.View) []*ProposeBlockInfo {
	//Check for valid block to vote
	validProposeBlock := []*ProposeBlockInfo{}
	//get all block that has height = bestview height  + 1(rule 2 & rule 3) (
	bestViewHeight := bestView.GetHeight()
	for h, proposeBlockInfo := range actorV2.receiveBlockByHash {
		if proposeBlockInfo.block == nil {
			continue
		}

		if proposeBlockInfo.block.GetHeight() != bestViewHeight+1 {
			if proposeBlockInfo.block.GetHeight() != bestViewHeight {
				continue
			}
			if proposeBlockInfo.block.Hash().String() != bestView.GetHash().String() {
				continue
			}
		}

		//not validate if we do it recently
		if time.Since(proposeBlockInfo.lastValidateTime).Seconds() < 1 {
			continue
		}

		// check if propose block in within TS
		if common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) != actorV2.currentTimeSlot {
			continue
		}

		// check if producer time > proposer time
		if common.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime()) > actorV2.currentTimeSlot {
			continue
		}

		// check if this time slot has been voted
		if actorV2.votedTimeslot[common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime())] {
			continue
		}

		if proposeBlockInfo.block.GetHeight() < actorV2.chain.GetFinalView().GetHeight() {
			delete(actorV2.votedTimeslot, proposeBlockInfo.block.GetProposeTime())
			delete(actorV2.receiveBlockByHash, h)
		}

		validProposeBlock = append(validProposeBlock, proposeBlockInfo)
	}
	//rule 1: get history of vote for this height, vote if (round is lower than the vote before) or (round is equal but new proposer) or (there is no vote for this height yet)
	sort.Slice(validProposeBlock, func(i, j int) bool {
		return validProposeBlock[i].block.GetProduceTime() < validProposeBlock[j].block.GetProduceTime()
	})
	return validProposeBlock
}

func (actorV2 *actorV2) validateBlock(bestView multiview.View, proposeBlockInfo *ProposeBlockInfo) error {
	blkCreateTimeSlot := common.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime())
	bestViewHeight := bestView.GetHeight()
	shouldVote := false

	if lastVotedBlk, ok := actorV2.voteHistory[bestViewHeight+1]; ok {
		if blkCreateTimeSlot < common.CalculateTimeSlot(lastVotedBlk.GetProduceTime()) { //blkCreateTimeSlot is smaller than voted block => vote for this blk
			shouldVote = true
		} else if blkCreateTimeSlot == common.CalculateTimeSlot(lastVotedBlk.GetProduceTime()) && common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) > common.CalculateTimeSlot(lastVotedBlk.GetProposeTime()) { //blk is old block (same round), but new proposer(larger timeslot) => vote again
			shouldVote = true
		} else if proposeBlockInfo.block.CommitteeFromBlock().String() != lastVotedBlk.CommitteeFromBlock().String() { //blkCreateTimeSlot is larger or equal than voted block
			shouldVote = true
		} // if not swap committees => do nothing
	} else { //there is no vote for this height yet
		shouldVote = true
	}

	if !shouldVote {
		actorV2.logger.Infof("Can't vote for this block %v height %v timeslot %v",
			proposeBlockInfo.block.Hash().String(), proposeBlockInfo.block.GetHeight(), blkCreateTimeSlot)
		return errors.New("Can't vote for this block")
	}

	//already vote for this proposed block
	if proposeBlockInfo.sendVote {
		return errors.New("Already vote for this block")
	}

	if proposeBlockInfo.isVoted {
		return errors.New("Already vote for this block")
	}

	//already validate and vote for this proposed block
	if !proposeBlockInfo.isValid {
		//not connected
		view := actorV2.chain.GetViewByHash(proposeBlockInfo.block.GetPrevHash())
		if view == nil {
			actorV2.logger.Infof("previous view for this block %v height %v timeslot %v is null",
				proposeBlockInfo.block.Hash().String(), proposeBlockInfo.block.GetHeight(), blkCreateTimeSlot)
			return errors.New("View not connect")
		}

		if _, ok := actorV2.bodyHashes[proposeBlockInfo.block.GetHeight()][proposeBlockInfo.block.BodyHash().String()]; !ok {
			_, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			actorV2.logger.Infof("validate block: %+v \n", proposeBlockInfo.block.Hash().String())
			if err := actorV2.chain.ValidatePreSignBlock(proposeBlockInfo.block, proposeBlockInfo.signingCommittes, proposeBlockInfo.committees); err != nil {
				actorV2.logger.Error(err)
				return err
			}

			// Block is valid for commit
			if len(actorV2.bodyHashes[proposeBlockInfo.block.GetHeight()]) == 0 {
				actorV2.bodyHashes[proposeBlockInfo.block.GetHeight()] = make(map[string]bool)
			}
			actorV2.bodyHashes[proposeBlockInfo.block.GetHeight()][proposeBlockInfo.block.BodyHash().String()] = true
		}
		proposeBlockInfo.isValid = true
	}

	return nil
}

func (actorV2 *actorV2) BlockVersion() int {
	return actorV2.blockVersion
}
