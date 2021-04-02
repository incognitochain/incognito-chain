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
	currentTime     int64
	currentTimeSlot int64
	proposeHistory  *lru.Cache

	receiveBlockByHeight map[uint64][]*ProposeBlockInfo  //blockHeight -> blockInfo
	receiveBlockByHash   map[string]*ProposeBlockInfo    //blockHash -> blockInfo
	voteHistory          map[uint64]types.BlockInterface // bestview height (previsous height )-> block
}

func (actorV2 *actorV2) Destroy() {
	actorV2.actorBase.Destroy()
	close(actorV2.destroyCh)
}

func NewActorV2() *actorV2 {
	return &actorV2{}
}

func NewActorV2WithValue(
	chain blockchain.Chain,
	chainKey string, chainID int,
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
	res.proposeHistory, err = lru.New(1000)
	if err != nil {
		panic(err) //must not error
	}
	//run?
	return res
}

func (actorV2 *actorV2) run() error {
	go func() {
		var err error
		//init view maps
		ticker := time.Tick(200 * time.Millisecond)
		cleanMemTicker := time.Tick(5 * time.Minute)
		actorV2.logger.Info("init bls-bftv3 consensus for chain", actorV2.chainKey)

		for { //actor loop
			if !actorV2.isStarted { //sleep if this process is not start
				time.Sleep(time.Second)
				continue
			}
			select {
			case <-actorV2.destroyCh:
				actorV2.logger.Info("exit bls-bftv3 consensus for chain", actorV2.chainKey)
				return
			case proposeMsg := <-actorV2.proposeMessageCh:
				blockIntf, err := actorV2.chain.UnmarshalBlock(proposeMsg.Block)
				if err != nil || blockIntf == nil {
					actorV2.logger.Info(err)
					continue
				}
				block := blockIntf.(types.BlockInterface)
				blkHash := block.Hash().String()

				committees, err := actorV2.getCommitteeForBlock(block)
				if err != nil {
					actorV2.logger.Debug(err)
					continue
				}

				res, _ := incognitokey.CommitteeKeyListToString(committees)
				actorV2.logger.Infof("######### Shard %+v, BlockHeight %+v, Committee %+v", actorV2.chain.GetShardID(), block.GetHeight(), res)

				blkCPk := incognitokey.CommitteePublicKey{}
				blkCPk.FromBase58(block.GetProducer())
				proposerMiningKeyBas58 := blkCPk.GetMiningKeyBase58(actorV2.GetConsensusName())

				if _, ok := actorV2.receiveBlockByHash[blkHash]; !ok {
					proposeBlockInfo := newProposeBlockForProposeMsg(block, committees, nil, nil, proposerMiningKeyBas58) //@tin
					actorV2.receiveBlockByHash[blkHash] = proposeBlockInfo
					actorV2.logger.Info("Receive block ", block.Hash().String(), "height", block.GetHeight(), ",block timeslot ", common.CalculateTimeSlot(block.GetProposeTime()))
					actorV2.receiveBlockByHeight[block.GetHeight()] = append(actorV2.receiveBlockByHeight[block.GetHeight()], actorV2.receiveBlockByHash[blkHash])
				} else {
					actorV2.receiveBlockByHash[blkHash].addBlockInfo(block, committees, nil, nil, proposerMiningKeyBas58, 0, 0) //@tin
				}

				if block.GetHeight() <= actorV2.chain.GetBestViewHeight() {
					actorV2.logger.Info("Receive block create from old view. Rejected!")
					continue
				}

				proposeView := actorV2.chain.GetViewByHash(block.GetPrevHash())
				if proposeView == nil {
					actorV2.logger.Infof("Request sync block from node %s from %s to %s", proposeMsg.PeerID, block.GetPrevHash().String(), block.GetPrevHash().Bytes())
					actorV2.node.RequestMissingViewViaStream(proposeMsg.PeerID, [][]byte{block.GetPrevHash().Bytes()}, actorV2.chain.GetShardID(), actorV2.chain.GetChainName())
				}

			case voteMsg := <-actorV2.voteMessageCh:
				voteMsg.IsValid = 0
				if b, ok := actorV2.receiveBlockByHash[voteMsg.BlockHash]; ok { //if receiveblock is already initiated
					if _, ok := b.votes[voteMsg.Validator]; !ok { // and not receive validatorA vote
						b.votes[voteMsg.Validator] = &voteMsg // store it
						vid, v := getValidatorIndex(actorV2.chain.GetBestView(), voteMsg.Validator)
						if v != nil {
							vbase58, _ := v.ToBase58()
							actorV2.logger.Infof("%v Receive vote (%d) for block %s from validator %d %v", actorV2.chainKey, len(actorV2.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, vid, vbase58)
						} else {
							actorV2.logger.Infof("%v Receive vote (%d) for block from unknown validator", actorV2.chainKey, len(actorV2.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, voteMsg.Validator)
						}
						b.hasNewVote = true
					}

					if !b.proposerSendVote {
						for _, userKey := range actorV2.userKeySet {
							pubKey := userKey.GetPublicKey()
							if b.block != nil && pubKey.GetMiningKeyBase58(actorV2.GetConsensusName()) == b.proposerMiningKeyBase58 { // if this node is proposer and not sending vote
								err := actorV2.validateAndVote(b) //validate in case we get malicious block
								if err == nil {
									bestViewHeight := actorV2.chain.GetBestView().GetHeight()
									if b.block.GetHeight() == bestViewHeight+1 { // and if the propose block is still connected to bestview
										view := actorV2.chain.GetViewByHash(b.block.GetPrevHash())
										err := actorV2.sendVote(&userKey, b.block, view.GetCommittee()) // => send vote
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
					vid, v := getValidatorIndex(actorV2.chain.GetBestView(), voteMsg.Validator)
					if v != nil {
						vbase58, _ := v.ToBase58()
						actorV2.logger.Infof("%v Receive vote (%d) for block %s from validator %d %v", actorV2.chainKey, len(actorV2.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, vid, vbase58)
					} else {
						actorV2.logger.Infof("%v Receive vote (%d) for block from unknown validator", actorV2.chainKey, len(actorV2.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, voteMsg.Validator)
					}
				}
			case <-cleanMemTicker:
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
						delete(actorV2.receiveBlockByHash, h)
					}
				}
			case <-ticker:
				if !actorV2.chain.IsReady() {
					continue
				}
				actorV2.currentTime = time.Now().Unix()

				newTimeSlot := false
				if actorV2.currentTimeSlot != common.CalculateTimeSlot(actorV2.currentTime) {
					newTimeSlot = true
				}

				actorV2.currentTimeSlot = common.CalculateTimeSlot(actorV2.currentTime)
				bestView := actorV2.chain.GetBestView()

				//set round for monitor
				round := actorV2.currentTimeSlot - common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime())
				monitor.SetGlobalParam("RoundKey", fmt.Sprintf("%d_%d", bestView.GetHeight(), round))

				committeeViewHash := common.Hash{}
				committees := []incognitokey.CommitteePublicKey{}
				proposerPk := incognitokey.CommitteePublicKey{}
				var userProposeKey signatureschemes2.MiningKey
				shouldPropose := false
				shouldListen := true

				if actorV2.chainID == BEACON_CHAIN_ID {
					proposerPk, _ = bestView.GetProposerByTimeSlot(actorV2.currentTimeSlot, 2)
					committees = actorV2.chain.GetBestView().GetCommittee()
				} else {
					committeeViewHash = *actorV2.committeeChain.FinalView().GetHash()
					committees, err = actorV2.CommitteeChain.CommitteesFromViewHashForShard(committeeViewHash, byte(actorV2.chainID))
					if err != nil {
						actorV2.logger.Error(err)
					}
					proposerPk = actorV2.CommitteeChain.ProposerByTimeSlot(byte(actorV2.chainID), actorV2.currentTimeSlot, committees)
				}

				for _, userKey := range actorV2.userKeySet {
					userPk := userKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus)
					if proposerPk.GetMiningKeyBase58(common.BlsConsensus) == userPk {
						shouldListen = false
						if common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()) != actorV2.currentTimeSlot { // current timeslot is not add to view, and this user is proposer of this timeslot
							//using block hash as key of best view -> check if this best view we propose or not
							if _, ok := actorV2.proposeHistory.Get(fmt.Sprintf("%s%d", actorV2.currentTimeSlot)); !ok {
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
						actorV2.logger.Infof("%v TS: %v, LISTEN BLOCK %v, Round %v", actorV2.chainKey, common.CalculateTimeSlot(actorV2.currentTime), bestView.GetHeight()+1, actorV2.currentTimeSlot-common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()))
					}
					if shouldPropose {
						actorV2.logger.Infof("%v TS: %v, PROPOSE BLOCK %v, Round %v", actorV2.chainKey, common.CalculateTimeSlot(actorV2.currentTime), bestView.GetHeight()+1, actorV2.currentTimeSlot-common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()))
					}
				}

				if shouldPropose {
					actorV2.proposeHistory.Add(fmt.Sprintf("%s%d", actorV2.currentTimeSlot), 1)
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

				//Check for valid block to vote
				validProposeBlock := []*ProposeBlockInfo{}
				//get all block that has height = bestview height  + 1(rule 2 & rule 3) (
				bestViewHeight := bestView.GetHeight()
				for h, proposeBlockInfo := range actorV2.receiveBlockByHash {
					if proposeBlockInfo.block == nil {
						continue
					}

					// e.Logger.Infof("[Monitor] bestview height %v, finalview height %v, block height %v %v", bestViewHeight, e.Chain.GetFinalView().GetHeight(), proposeBlockInfo.block.GetHeight(), proposeBlockInfo.block.GetProduceTime())
					if proposeBlockInfo.block.GetHeight() == bestViewHeight+1 {
						validProposeBlock = append(validProposeBlock, proposeBlockInfo)
					} else {
						if proposeBlockInfo.block.Hash().String() == bestView.GetHash().String() {
							validProposeBlock = append(validProposeBlock, proposeBlockInfo)
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

					validProposeBlock = append(validProposeBlock, proposeBlockInfo)

					if proposeBlockInfo.block.GetHeight() < actorV2.chain.GetFinalView().GetHeight() {
						delete(actorV2.receiveBlockByHash, h)
					}

				}
				//rule 1: get history of vote for this height, vote if (round is lower than the vote before) or (round is equal but new proposer) or (there is no vote for this height yet)
				sort.Slice(validProposeBlock, func(i, j int) bool {
					return validProposeBlock[i].block.GetProduceTime() < validProposeBlock[j].block.GetProduceTime()
				})
				for _, v := range validProposeBlock {
					blkCreateTimeSlot := common.CalculateTimeSlot(v.block.GetProduceTime())
					bestViewHeight := bestView.GetHeight()

					if lastVotedBlk, ok := actorV2.voteHistory[bestViewHeight+1]; ok {
						if blkCreateTimeSlot < common.CalculateTimeSlot(lastVotedBlk.GetProduceTime()) { //blkCreateTimeSlot is smaller than voted block => vote for this blk
							actorV2.validateAndVote(v)
						} else if blkCreateTimeSlot == common.CalculateTimeSlot(lastVotedBlk.GetProduceTime()) && common.CalculateTimeSlot(v.block.GetProposeTime()) > common.CalculateTimeSlot(lastVotedBlk.GetProposeTime()) { //blk is old block (same round), but new proposer(larger timeslot) => vote again
							actorV2.validateAndVote(v)
						} else if v.block.CommitteeFromBlock().String() != lastVotedBlk.CommitteeFromBlock().String() { //blkCreateTimeSlot is larger or equal than voted block
							actorV2.validateAndVote(v)
						} // if not swap committees => do nothing
					} else { //there is no vote for this height yet
						actorV2.validateAndVote(v)
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

func getValidatorIndex(view multiview.View, validator string) (int, *incognitokey.CommitteePublicKey) {
	for id, c := range view.GetCommittee() {
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

	validVote := 0
	errVote := 0
	for id, vote := range v.votes {
		dsaKey := []byte{}
		if vote.IsValid == 0 {
			cid, committeePk := getValidatorIndex(view, vote.Validator)
			if committeePk != nil {
				dsaKey = committeePk.MiningPubKey[common.BridgeConsensus]
			} else {
				actorV2.logger.Error("Receive vote from nonCommittee member")
				continue
			}
			if len(dsaKey) == 0 {
				actorV2.logger.Error(fmt.Sprintf("Cannot find dsa key from vote of %d", cid))
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
	// TODO: @tin/0xkumi check here again
	for key, value := range v.votes {
		if value.IsValid == -1 {
			delete(v.votes, key)
		}
	}

	if !v.isCommitted {
		if validVote > 2*len(v.committees)/3 {
			v.isCommitted = true
			actorV2.logger.Infof("Commit block %v , height: %v", blockHash, v.block.GetHeight())
			if actorV2.chainID == BEACON_CHAIN_ID {
				actorV2.processWithEnoughVotesBeaconChain(v)
			} else {
				actorV2.processWithEnoughVotesShardChain(v)
			}
		}
	}
}

func (actorV2 *actorV2) processWithEnoughVotesBeaconChain(
	v *ProposeBlockInfo,
) {
	validationData, err := createBLSAggregatedSignatures(v.committees, v.block.GetValidationField(), v.votes)
	if err != nil {
		actorV2.logger.Error(err)
		return
	}
	v.block.(blockValidation).AddValidationField(validationData)

	go actorV2.chain.InsertAndBroadcastBlock(v.block)

	delete(actorV2.receiveBlockByHash, v.block.GetPrevHash().String())
}

func (actorV2 *actorV2) processWithEnoughVotesShardChain(
	v *ProposeBlockInfo,
) {
	// validationData at present block
	validationData, err := createBLSAggregatedSignatures(v.committees, v.block.GetValidationField(), v.votes)
	if err != nil {
		actorV2.logger.Error(err)
		return
	}
	v.block.(blockValidation).AddValidationField(validationData)

	// validate and previous block
	if previousProposeBlockInfo, ok := actorV2.receiveBlockByHash[v.block.GetPrevHash().String()]; ok &&
		previousProposeBlockInfo != nil && previousProposeBlockInfo.block != nil {
		previousValidationData, err := createBLSAggregatedSignatures(
			previousProposeBlockInfo.committees,
			previousProposeBlockInfo.block.GetValidationField(),
			previousProposeBlockInfo.votes)

		if err != nil {
			actorV2.logger.Error(err)
			return
		}

		previousProposeBlockInfo.block.(blockValidation).AddValidationField(previousValidationData)
		go actorV2.chain.InsertAndBroadcastBlockWithPrevValidationData(v.block, previousValidationData)
		delete(actorV2.receiveBlockByHash, previousProposeBlockInfo.block.GetPrevHash().String())
	} else {
		go actorV2.chain.InsertAndBroadcastBlock(v.block)
	}
}

func createBLSAggregatedSignatures(committees []incognitokey.CommitteePublicKey, tempValidationData string, votes map[string]*BFTVote) (string, error) {
	committeeBLSString, err := incognitokey.ExtractPublickeysFromCommitteeKeyList(committees, common.BlsConsensus)
	if err != nil {
		return "", err
	}
	aggSig, brigSigs, validatorIdx, err := CombineVotes(votes, committeeBLSString)
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

func (actorV2 *actorV2) validateAndVote(
	v *ProposeBlockInfo,
) error {
	//already vote for this proposed block
	if v.sendVote {
		return nil
	}

	if v.isVoted {
		return nil
	}

	if !v.isValid {
		//not connected
		view := actorV2.chain.GetViewByHash(v.block.GetPrevHash())
		if view == nil {
			actorV2.logger.Info("view is null")
			return errors.New("View not connect")
		}

		//TODO: using context to validate block
		_, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if err := actorV2.chain.ValidatePreSignBlock(v.block, v.committees, v.signingCommittes); err != nil {
			actorV2.logger.Error(err)
			return err
		}
	}
	v.isValid = true

	//if valid then vote
	for _, userKey := range actorV2.userKeySet {
		Vote, err := CreateVote(&userKey, v.block, v.committees)
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
		actorV2.logger.Info(actorV2.chainKey, "sending vote...")
		go actorV2.processBFTMsg(msg.(*wire.MessageBFT))
		go actorV2.node.PushMessageToChain(msg, actorV2.chain)
	}

	return nil
}

func CreateVote(userKey *signatureschemes2.MiningKey, block types.BlockInterface, committees []incognitokey.CommitteePublicKey) (*BFTVote, error) {
	var Vote = new(BFTVote)
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
	Vote.Bls = blsSig
	Vote.Bri = bridgeSig
	Vote.BlockHash = block.Hash().String()

	userPk := userKey.GetPublicKey()
	Vote.Validator = userPk.GetMiningKeyBase58(common.BlsConsensus)
	Vote.PrevBlockHash = block.GetPrevHash().String()
	err = Vote.signVote(userKey)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	return Vote, nil
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
		actorV2.logger.Infof("create block %v hash %v, propose time %v, produce time %v", block.GetHeight(), block.Hash().String(), block.(types.BlockInterface).GetProposeTime(), block.(types.BlockInterface).GetProduceTime())
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
		block, err = actorV2.chain.CreateNewBlock(2, b58Str, 1, actorV2.currentTime, committees, committeeViewHash)
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
	var err1 error
	var committeesFromBeaconHash []incognitokey.CommitteePublicKey

	if block != nil {
		committeesFromBeaconHash, err1 = actorV2.getCommitteeForBlock(block)
		if err1 != nil {
			return block, NewConsensusError(BlockCreationError, err1)
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
		block, err = actorV2.chain.CreateNewBlock(2, b58Str, 1, actorV2.currentTime, committees, committeeViewHash)
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

func (actorV2 *actorV2) processBFTMsg(msgBFT *wire.MessageBFT) {
	switch msgBFT.Type {
	case MsgPropose:
		var msgPropose BFTPropose
		err := json.Unmarshal(msgBFT.Content, &msgPropose)
		if err != nil {
			actorV2.logger.Error(err)
			return
		}
		msgPropose.PeerID = msgBFT.PeerID
		actorV2.proposeMessageCh <- msgPropose
	case MsgVote:
		var msgVote BFTVote
		err := json.Unmarshal(msgBFT.Content, &msgVote)
		if err != nil {
			actorV2.logger.Error(err)
			return
		}
		actorV2.voteMessageCh <- msgVote
	default:
		actorV2.logger.Critical("Unknown BFT message type")
		return
	}
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

func (actorV2 *actorV2) getCommitteeForBlock(v types.BlockInterface) ([]incognitokey.CommitteePublicKey, error) {
	var err error = nil
	var committees []incognitokey.CommitteePublicKey
	if !actorV2.chain.IsBeaconChain() {
		committees, err = actorV2.CommitteeChain.CommitteesFromViewHashForShard(v.CommitteeFromBlock(), byte(actorV2.chain.GetShardID()))
	} else {
		committees = actorV2.chain.GetBestView().GetCommittee()
	}
	return committees, err
}

func (s *BFTVote) signVote(key *signatureschemes2.MiningKey) error {
	data := []byte{}
	data = append(data, s.BlockHash...)
	data = append(data, s.Bls...)
	data = append(data, s.Bri...)
	data = common.HashB(data)
	var err error
	s.Confirmation, err = key.BriSignData(data)
	return err
}

func (s *BFTVote) validateVoteOwner(ownerPk []byte) error {
	data := []byte{}
	data = append(data, s.BlockHash...)
	data = append(data, s.Bls...)
	data = append(data, s.Bri...)
	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, s.Confirmation, ownerPk)
	return err
}

/*func ExtractBridgeValidationData(block types.BlockInterface) ([][]byte, []int, error) {*/
//valData, err := consensustypes.DecodeValidationData(block.GetValidationField())
//if err != nil {
//return nil, nil, NewConsensusError(UnExpectedError, err)
//}
//return valData.BridgeSig, valData.ValidatiorsIdx, nil
/*}*/

func (actorV2 *actorV2) sendVote(userKey *signatureschemes2.MiningKey, block types.BlockInterface, committees []incognitokey.CommitteePublicKey) error {
	Vote, err := CreateVote(userKey, block, committees)
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
