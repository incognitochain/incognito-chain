package blsbftv3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/incognitochain/incognito-chain/metrics/monitor"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wire"
)

const (
	BEACON_CHAIN_ID = -1
)

type BLSBFT_V3 struct {
	CommitteeChain CommitteeChainHandler
	Chain          ChainInterface
	Node           NodeInterface
	ChainKey       string
	ChainID        int
	PeerID         string

	UserKeySet   []signatureschemes2.MiningKey
	BFTMessageCh chan wire.MessageBFT
	isStarted    bool
	destroyCh    chan struct{}
	Logger       common.Logger

	currentTime      int64
	currentTimeSlot  int64
	proposeHistory   *lru.Cache
	ProposeMessageCh chan BFTPropose
	VoteMessageCh    chan BFTVote

	receiveBlockByHeight map[uint64][]*ProposeBlockInfo  //blockHeight -> blockInfo
	receiveBlockByHash   map[string]*ProposeBlockInfo    //blockHash -> blockInfo
	voteHistory          map[uint64]types.BlockInterface // bestview height (previsous height )-> block
}

func (e BLSBFT_V3) GetVoteHistory() map[uint64]types.BlockInterface {
	return e.voteHistory
}

func (e BLSBFT_V3) GetReceiveBlockByHash() map[string]*ProposeBlockInfo {
	return e.receiveBlockByHash
}

func (e BLSBFT_V3) GetReceiveBlockByHeight() map[uint64][]*ProposeBlockInfo {
	return e.receiveBlockByHeight
}

func (e BLSBFT_V3) GetChainKey() string {
	return e.ChainKey
}

func (e BLSBFT_V3) GetChainID() int {
	return e.ChainID
}

func (e BLSBFT_V3) IsOngoing() bool {
	return e.isStarted
}

func (e BLSBFT_V3) IsStarted() bool {
	return e.isStarted
}

func (e *BLSBFT_V3) GetConsensusName() string {
	return common.BlsConsensus
}

func (e *BLSBFT_V3) Stop() error {
	if e.isStarted {
		e.Logger.Info("stop bls-bftv3 consensus for chain", e.ChainKey)
	}
	e.isStarted = false
	return nil
}

func (e *BLSBFT_V3) Start() error {
	if !e.isStarted {
		e.Logger.Info("start bls-bftv3 consensus for chain", e.ChainKey)
	}
	e.isStarted = true
	return nil
}

func (e *BLSBFT_V3) Destroy() {
	e.destroyCh <- struct{}{}
}

func (e *BLSBFT_V3) run() error {
	go func() {
		var err error
		//init view maps
		ticker := time.Tick(200 * time.Millisecond)
		cleanMemTicker := time.Tick(5 * time.Minute)
		e.Logger.Info("init bls-bftv3 consensus for chain", e.ChainKey)

		for { //actor loop
			if !e.isStarted { //sleep if this process is not start
				time.Sleep(time.Second)
				continue
			}
			select {
			case <-e.destroyCh:
				e.Logger.Info("exit bls-bftv3 consensus for chain", e.ChainKey)
				close(e.destroyCh)
				return
			case proposeMsg := <-e.ProposeMessageCh:
				//fmt.Println("debug receive propose message", string(proposeMsg.Block))
				blockIntf, err := e.Chain.UnmarshalBlock(proposeMsg.Block)
				if err != nil || blockIntf == nil {
					e.Logger.Error(err)
					continue
				}
				block := blockIntf.(types.BlockInterface)
				blkHash := block.Hash().String()

				committees, err := e.getCommitteeForBlock(block)
				if err != nil {
					e.Logger.Error(err)
					continue
				}

				blkCPk := incognitokey.CommitteePublicKey{}
				blkCPk.FromBase58(block.GetProducer())
				proposerMiningKeyBas58 := blkCPk.GetMiningKeyBase58(e.GetConsensusName())

				if _, ok := e.receiveBlockByHash[blkHash]; !ok {
					proposeBlockInfo := newProposeBlockForProposeMsg(block, committees, make(map[string]*BFTVote), false, proposerMiningKeyBas58)
					e.receiveBlockByHash[blkHash] = proposeBlockInfo
					e.Logger.Info("Receive block ", block.Hash().String(), "height", block.GetHeight(), ",block timeslot ", common.CalculateTimeSlot(block.GetProposeTime()))
					e.receiveBlockByHeight[block.GetHeight()] = append(e.receiveBlockByHeight[block.GetHeight()], e.receiveBlockByHash[blkHash])
				} else {
					e.receiveBlockByHash[blkHash].addBlockInfo(block, committees, proposerMiningKeyBas58)
				}

				if block.GetHeight() <= e.Chain.GetBestViewHeight() {
					e.Logger.Info("Receive block create from old view. Rejected!")
					continue
				}

				proposeView := e.Chain.GetViewByHash(block.GetPrevHash())
				if proposeView == nil {
					e.Logger.Infof("Request sync block from node %s from %s to %s", proposeMsg.PeerID, block.GetPrevHash().String(), block.GetPrevHash().Bytes())
					e.Node.RequestMissingViewViaStream(proposeMsg.PeerID, [][]byte{block.GetPrevHash().Bytes()}, e.Chain.GetShardID(), e.Chain.GetChainName())
				}

			case voteMsg := <-e.VoteMessageCh:
				voteMsg.IsValid = 0
				if b, ok := e.receiveBlockByHash[voteMsg.BlockHash]; ok { //if receiveblock is already initiated
					if _, ok := b.votes[voteMsg.Validator]; !ok { // and not receive validatorA vote
						b.votes[voteMsg.Validator] = &voteMsg // store it
						vid, v := getValidatorIndex(b.committees, voteMsg.Validator)
						if v != nil {
							vbase58, _ := v.ToBase58()
							e.Logger.Infof("%v Receive vote (%d) for block %s from validator %d %v", e.ChainKey, len(e.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, vid, vbase58)
						} else {
							e.Logger.Infof("%v Receive vote (%d) for block %v from unknown validator %v", e.ChainKey, len(e.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, voteMsg.Validator)
						}
						b.hasNewVote = true
					}

					if !b.proposerSendVote {
						for _, userKey := range e.UserKeySet {
							pubKey := userKey.GetPublicKey()
							if b.block != nil && pubKey.GetMiningKeyBase58(e.GetConsensusName()) == b.proposerMiningKeyBase58 { // if this node is proposer and not sending vote
								err := e.validateAndVote(b) //validate in case we get malicious block
								if err == nil {
									bestViewHeight := e.Chain.GetBestView().GetHeight()
									if b.block.GetHeight() == bestViewHeight+1 { // and if the propose block is still connected to bestview
										err := e.sendVote(&userKey, b.block, b.committees) // => send vote
										if err != nil {
											e.Logger.Error(err)
										} else {
											b.proposerSendVote = true
											b.isVoted = true
										}
									}
								}
							}
						}
					}
				} else {
					e.receiveBlockByHash[voteMsg.BlockHash] = newBlockInfoForVoteMsg()
					e.receiveBlockByHash[voteMsg.BlockHash].votes[voteMsg.Validator] = &voteMsg
					e.Logger.Infof("%v Receive vote (%d) for block %v from validator %v", e.ChainKey, len(e.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, voteMsg.Validator)
				}
			case <-cleanMemTicker:
				for h, _ := range e.receiveBlockByHeight {
					if h <= e.Chain.GetFinalView().GetHeight() {
						delete(e.receiveBlockByHeight, h)
					}
				}
				for h, _ := range e.voteHistory {
					if h <= e.Chain.GetFinalView().GetHeight() {
						delete(e.voteHistory, h)
					}
				}
				for h, proposeBlk := range e.receiveBlockByHash {
					if time.Now().Sub(proposeBlk.receiveTime) > time.Minute {
						delete(e.receiveBlockByHash, h)
					}
				}
			case <-ticker:
				if !e.Chain.IsReady() {
					continue
				}
				e.currentTime = time.Now().Unix()

				newTimeSlot := false
				if e.currentTimeSlot != common.CalculateTimeSlot(e.currentTime) {
					newTimeSlot = true
				}

				e.currentTimeSlot = common.CalculateTimeSlot(e.currentTime)
				bestView := e.Chain.GetBestView()

				//set round for monitor
				round := e.currentTimeSlot - common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime())
				monitor.SetGlobalParam("RoundKey", fmt.Sprintf("%d_%d", bestView.GetHeight(), round))

				committeeViewHash := common.Hash{}
				committees := []incognitokey.CommitteePublicKey{}
				proposerPk := incognitokey.CommitteePublicKey{}
				var userProposeKey signatureschemes2.MiningKey
				shouldPropose := false
				shouldListen := true

				if e.ChainID == BEACON_CHAIN_ID {
					proposerPk, _ = bestView.GetProposerByTimeSlot(e.currentTimeSlot, 2)
					committees = e.Chain.GetBestView().GetCommittee()
				} else {
					committeeViewHash = *e.CommitteeChain.FinalView().GetHash()
					committees, err = e.CommitteeChain.CommitteesFromViewHashForShard(committeeViewHash, byte(e.ChainID))
					if err != nil {
						e.Logger.Error(err)
					}
					proposerPk = e.CommitteeChain.ProposerByTimeSlot(byte(e.ChainID), e.currentTimeSlot, committees)
				}

				for _, userKey := range e.UserKeySet {
					userPk := userKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus)
					if proposerPk.GetMiningKeyBase58(common.BlsConsensus) == userPk {
						shouldListen = false
						if common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()) != e.currentTimeSlot { // current timeslot is not add to view, and this user is proposer of this timeslot
							//using block hash as key of best view -> check if this best view we propose or not
							if _, ok := e.proposeHistory.Get(fmt.Sprintf("%s%d", e.currentTimeSlot)); !ok {
								shouldPropose = true
								userProposeKey = userKey
							}
						}
					}
				}

				if newTimeSlot { //for logging
					e.Logger.Info("")
					e.Logger.Info("======================================================")
					e.Logger.Info("")
					if shouldListen {
						e.Logger.Infof("%v TS: %v, LISTEN BLOCK %v, Round %v", e.ChainKey, common.CalculateTimeSlot(e.currentTime), bestView.GetHeight()+1, e.currentTimeSlot-common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()))
					}
					if shouldPropose {
						e.Logger.Infof("%v TS: %v, PROPOSE BLOCK %v, Round %v", e.ChainKey, common.CalculateTimeSlot(e.currentTime), bestView.GetHeight()+1, e.currentTimeSlot-common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()))
					}
				}

				if shouldPropose {
					e.proposeHistory.Add(fmt.Sprintf("%s%d", e.currentTimeSlot), 1)
					//Proposer Rule: check propose block connected to bestview(longest chain rule 1) and re-propose valid block with smallest timestamp (including already propose in the past) (rule 2)
					sort.Slice(e.receiveBlockByHeight[bestView.GetHeight()+1], func(i, j int) bool {
						return e.receiveBlockByHeight[bestView.GetHeight()+1][i].block.GetProduceTime() < e.receiveBlockByHeight[bestView.GetHeight()+1][j].block.GetProduceTime()
					})

					var proposeBlock types.BlockInterface = nil
					for _, v := range e.receiveBlockByHeight[bestView.GetHeight()+1] {
						if v.isValid {
							proposeBlock = v.block
							break
						}
					}

					if createdBlk, err := e.proposeBlock(userProposeKey, proposerPk, proposeBlock, committees, committeeViewHash); err != nil {
						e.Logger.Critical(UnExpectedError, errors.New("can't propose block"))
						e.Logger.Critical(err)
					} else {
						e.Logger.Infof("proposer block %v round %v time slot %v blockTimeSlot %v with hash %v", createdBlk.GetHeight(), createdBlk.GetRound(), e.currentTimeSlot, common.CalculateTimeSlot(createdBlk.GetProduceTime()), createdBlk.Hash().String())
					}
				}

				//Check for valid block to vote
				//rule 2 & rule 3 (nextheight = bestview + 1)
				validProposeBlock := []*ProposeBlockInfo{}
				bestViewHeight := bestView.GetHeight()
				for h, proposeBlockInfo := range e.receiveBlockByHash {
					if proposeBlockInfo.block == nil {
						continue
					}

					//not validate if we do it recently
					if time.Since(proposeBlockInfo.lastValidateTime).Seconds() < 1 {
						continue
					}

					//if the block is bestview and we didnt vote yet => validate and vote
					if proposeBlockInfo.block.Hash().String() == bestView.GetHash().String() && !proposeBlockInfo.isVoted {
						e.validateAndVote(proposeBlockInfo)
						continue
					}

					//if the block height is not next height
					if proposeBlockInfo.block.GetHeight() != bestViewHeight+1 {
						continue
					}

					// check if propose block in within TS
					if common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) != e.currentTimeSlot {
						continue
					}

					validProposeBlock = append(validProposeBlock, proposeBlockInfo)

					if proposeBlockInfo.block.GetHeight() < e.Chain.GetFinalView().GetHeight() {
						delete(e.receiveBlockByHash, h)
					}

				}
				//rule 1: get history of vote for this height, vote if (round is lower than the vote before) or (round is equal but new proposer) or (there is no vote for this height yet)
				sort.Slice(validProposeBlock, func(i, j int) bool {
					return validProposeBlock[i].block.GetProduceTime() < validProposeBlock[j].block.GetProduceTime()
				})

				for _, v := range validProposeBlock {
					blkCreateTimeSlot := common.CalculateTimeSlot(v.block.GetProduceTime())
					bestViewHeight := bestView.GetHeight()
					if lastVotedBlk, ok := e.voteHistory[bestViewHeight+1]; ok {
						if blkCreateTimeSlot < common.CalculateTimeSlot(lastVotedBlk.GetProduceTime()) { //blkCreateTimeSlot is smaller than voted block => vote for this blk
							e.validateAndVote(v)
						} else if blkCreateTimeSlot == common.CalculateTimeSlot(lastVotedBlk.GetProduceTime()) && common.CalculateTimeSlot(v.block.GetProposeTime()) > common.CalculateTimeSlot(lastVotedBlk.GetProposeTime()) { //blk is old block (same round), but new proposer(larger timeslot) => vote again
							e.validateAndVote(v)
						} else if v.block.CommitteeFromBlock().String() != lastVotedBlk.CommitteeFromBlock().String() { //blkCreateTimeSlot is larger or equal than voted block
							e.validateAndVote(v)
						} // if not swap committees => do nothing
					} else { //there is no vote for this height yet
						e.validateAndVote(v)
					}
				}
				/*
					Check for 2/3 vote to commit
				*/
				for k, v := range e.receiveBlockByHash {
					e.processIfBlockGetEnoughVote(k, v)
				}
			}
		}
	}()
	return nil
}

func getValidatorIndex(committees []incognitokey.CommitteePublicKey, validator string) (int, *incognitokey.CommitteePublicKey) {
	for id, c := range committees {
		if validator == c.GetMiningKeyBase58(common.BlsConsensus) {
			return id, &c
		}
	}
	return -1, nil
}

func NewInstance(chain ChainInterface, committeeChain CommitteeChainHandler, chainKey string, chainID int, node NodeInterface, logger common.Logger) *BLSBFT_V3 {
	var err error
	var newInstance = new(BLSBFT_V3)
	newInstance.Chain = chain
	newInstance.ChainKey = chainKey
	newInstance.ChainID = chainID
	newInstance.Node = node
	newInstance.Logger = logger
	newInstance.destroyCh = make(chan struct{})
	newInstance.ProposeMessageCh = make(chan BFTPropose)
	newInstance.VoteMessageCh = make(chan BFTVote)
	newInstance.receiveBlockByHash = make(map[string]*ProposeBlockInfo)
	newInstance.receiveBlockByHeight = make(map[uint64][]*ProposeBlockInfo)
	newInstance.voteHistory = make(map[uint64]types.BlockInterface)
	newInstance.proposeHistory, err = lru.New(1000)
	if err != nil {
		panic(err) //must not error
	}
	newInstance.CommitteeChain = committeeChain
	newInstance.run()
	return newInstance
}

func (e *BLSBFT_V3) processIfBlockGetEnoughVote(
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
	e.Logger.Infof("Process Block IF enough votes, %+v, %+v", *v.block.Hash(), v.block.GetHeight())
	//already in chain
	bestView := e.Chain.GetBestView()
	view := e.Chain.GetViewByHash(*v.block.Hash())
	if view != nil && bestView.GetHash().String() != view.GetHash().String() {
		//e.Logger.Infof("Get View By Hash Fail, %+v, %+v", *v.block.Hash(), v.block.GetHeight())
		return
	}

	//not connected previous block
	view = e.Chain.GetViewByHash(v.block.GetPrevHash())
	if view == nil {
		//e.Logger.Infof("Get Previous View By Hash Fail, %+v, %+v", v.block.GetPrevHash(), v.block.GetHeight()-1)
		return
	}

	validVote := 0
	errVote := 0
	for id, vote := range v.votes {
		dsaKey := []byte{}
		if vote.IsValid == 0 {
			cid, committeePk := getValidatorIndex(v.committees, vote.Validator)
			if committeePk != nil {
				dsaKey = committeePk.MiningPubKey[common.BridgeConsensus]
			} else {
				e.Logger.Error("Receive vote from nonCommittee member")
				continue
			}
			if len(dsaKey) == 0 {
				e.Logger.Error(fmt.Sprintf("Cannot find dsa key from vote of %d", cid))
				continue
			}

			err := vote.validateVoteOwner(dsaKey)
			if err != nil {
				e.Logger.Error(dsaKey)
				e.Logger.Error(err)
				panic(1)
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
	e.Logger.Info("Number of Valid Vote", validVote, "| Number Of Error Vote", errVote)
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
			e.Logger.Infof("Commit block %v , height: %v", blockHash, v.block.GetHeight())
			if e.ChainID == BEACON_CHAIN_ID {
				e.processWithEnoughVotesBeaconChain(v)
			} else {
				e.processWithEnoughVotesShardChain(v)
			}
		}
	}
}

func (e *BLSBFT_V3) processWithEnoughVotesBeaconChain(
	v *ProposeBlockInfo,
) {
	validationData, err := createBLSAggregatedSignatures(v.committees, v.block.GetValidationField(), v.votes)
	if err != nil {
		e.Logger.Error(err)
		return
	}
	v.block.(blockValidation).AddValidationField(validationData)

	go e.Chain.InsertAndBroadcastBlock(v.block)

	delete(e.receiveBlockByHash, v.block.GetPrevHash().String())
}

func (e *BLSBFT_V3) processWithEnoughVotesShardChain(
	v *ProposeBlockInfo,
) {
	// validationData at present block
	validationData, err := createBLSAggregatedSignatures(v.committees, v.block.GetValidationField(), v.votes)
	if err != nil {
		e.Logger.Error(err)
		return
	}
	v.block.(blockValidation).AddValidationField(validationData)

	// validate and previous block
	if previousProposeBlockInfo, ok := e.receiveBlockByHash[v.block.GetPrevHash().String()]; ok &&
		previousProposeBlockInfo != nil && previousProposeBlockInfo.block != nil {
		previousValidationData, err := createBLSAggregatedSignatures(
			previousProposeBlockInfo.committees,
			previousProposeBlockInfo.block.GetValidationField(),
			previousProposeBlockInfo.votes)

		if err != nil {
			e.Logger.Error(err)
			return
		}

		previousProposeBlockInfo.block.(blockValidation).AddValidationField(previousValidationData)
		go e.Chain.InsertAndBroadcastBlockWithPrevValidationData(v.block, previousValidationData)
		delete(e.receiveBlockByHash, previousProposeBlockInfo.block.GetPrevHash().String())
	} else {
		go e.Chain.InsertAndBroadcastBlock(v.block)
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

func (e *BLSBFT_V3) validateAndVote(
	v *ProposeBlockInfo,
) error {
	v.lastValidateTime = time.Now()

	//already vote for this proposed block
	if v.isVoted {
		return nil
	}

	if !v.isValid {
		//not connected
		view := e.Chain.GetViewByHash(v.block.GetPrevHash())
		if view == nil {
			e.Logger.Info("view is null")
			return errors.New("View not connect")
		}

		//TODO: using context to validate block
		_, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if err := e.Chain.ValidatePreSignBlock(v.block, v.committees); err != nil {
			e.Logger.Error(err)
			return err
		}
	}
	v.isValid = true

	//if valid then vote
	committeeBLSString, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(v.committees, common.BlsConsensus)
	for _, userKey := range e.UserKeySet {
		pubKey := userKey.GetPublicKey()
		// When node is not connect to highway (drop connection/startup), propose and vote a block will prevent voting for any other blocks having same height but larger timestamp (rule1)
		// In case number of validator is 22, we need to make 22 turn to propose the old smallest timestamp block
		// To prevent this, proposer will not vote unless receiving at least one vote (look at receive vote event)
		if pubKey.GetMiningKeyBase58(e.GetConsensusName()) == v.proposerMiningKeyBase58 {
			continue
		}

		if common.IndexOfStr(pubKey.GetMiningKeyBase58(e.GetConsensusName()), committeeBLSString) != -1 {
			err := e.sendVote(&userKey, v.block, v.committees)
			if err != nil {
				e.Logger.Error(err)
				return NewConsensusError(UnExpectedError, err)
			} else {
				v.isVoted = true
			}
		}
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
	Vote.BLS = blsSig
	Vote.BRI = bridgeSig
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

func (e *BLSBFT_V3) proposeBlock(
	userMiningKey signatureschemes2.MiningKey,
	proposerPk incognitokey.CommitteePublicKey,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
) (types.BlockInterface, error) {
	time1 := time.Now()
	b58Str, _ := proposerPk.ToBase58()
	var err error

	if e.Chain.IsBeaconChain() {
		block, err = e.proposeBeaconBlock(
			b58Str,
			block,
			committees,
			committeeViewHash,
		)
	} else {
		block, err = e.proposeShardBlock(
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
		e.Logger.Infof("create block %v hash %v, propose time %v, produce time %v", block.GetHeight(), block.Hash().String(), block.(types.BlockInterface).GetProposeTime(), block.(types.BlockInterface).GetProduceTime())
	} else {
		e.Logger.Infof("create block fail, time: %v", time.Since(time1).Seconds())
		return nil, NewConsensusError(BlockCreationError, errors.New("block is nil"))
	}

	var validationData consensustypes.ValidationData
	validationData.ProducerBLSSig, _ = userMiningKey.BriSignData(block.Hash().GetBytes())
	validationDataString, _ := consensustypes.EncodeValidationData(validationData)
	block.(blockValidation).AddValidationField(validationDataString)
	blockData, _ := json.Marshal(block)

	var proposeCtn = new(BFTPropose)
	proposeCtn.Block = blockData
	proposeCtn.PeerID = e.Node.GetSelfPeerID().String()
	msg, _ := MakeBFTProposeMsg(proposeCtn, e.ChainKey, e.currentTimeSlot, block.GetHeight())
	go e.ProcessBFTMsg(msg.(*wire.MessageBFT))
	go e.Node.PushMessageToChain(msg, e.Chain)

	return block, nil
}

func (e *BLSBFT_V3) proposeBeaconBlock(
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
		e.Logger.Info("CreateNewBlock")
		block, err = e.Chain.CreateNewBlock(3, b58Str, 1, e.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	} else {
		e.Logger.Infof("CreateNewBlockFromOldBlock, Block Height %+v")
		block, err = e.Chain.CreateNewBlockFromOldBlock(block, b58Str, e.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	}
	return block, err
}

func (e *BLSBFT_V3) proposeShardBlock(
	b58Str string,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
) (types.BlockInterface, error) {
	var err error
	var err1 error
	var committeesFromBeaconHash []incognitokey.CommitteePublicKey

	if block != nil {
		committeesFromBeaconHash, err1 = e.getCommitteeForBlock(block)
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
		e.Logger.Info("CreateNewBlock")
		block, err = e.Chain.CreateNewBlock(3, b58Str, 1, e.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	} else {
		e.Logger.Infof("CreateNewBlockFromOldBlock, Block Height %+v")
		block, err = e.Chain.CreateNewBlockFromOldBlock(block, b58Str, e.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	}
	return block, err
}

func (e *BLSBFT_V3) ProcessBFTMsg(msgBFT *wire.MessageBFT) {
	switch msgBFT.Type {
	case MSG_PROPOSE:
		var msgPropose BFTPropose
		err := json.Unmarshal(msgBFT.Content, &msgPropose)
		if err != nil {
			e.Logger.Error(err)
			return
		}
		msgPropose.PeerID = msgBFT.PeerID
		e.ProposeMessageCh <- msgPropose
	case MSG_VOTE:
		var msgVote BFTVote
		err := json.Unmarshal(msgBFT.Content, &msgVote)
		if err != nil {
			e.Logger.Error(err)
			return
		}
		e.VoteMessageCh <- msgVote
	default:
		e.Logger.Critical("Unknown BFT message type")
		return
	}
}

func (e *BLSBFT_V3) preValidateVote(blockHash []byte, Vote *BFTVote, candidate []byte) error {
	data := []byte{}
	data = append(data, blockHash...)
	data = append(data, Vote.BLS...)
	data = append(data, Vote.BRI...)
	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, Vote.Confirmation, candidate)
	return err
}

func (e *BLSBFT_V3) getCommitteeForBlock(v types.BlockInterface) ([]incognitokey.CommitteePublicKey, error) {
	var err error = nil
	var committees []incognitokey.CommitteePublicKey
	if !e.Chain.IsBeaconChain() {
		committees, err = e.CommitteeChain.CommitteesFromViewHashForShard(v.CommitteeFromBlock(), byte(e.Chain.GetShardID()))
	} else {
		committees = e.Chain.GetBestView().GetCommittee()
	}
	return committees, err
}

func (s *BFTVote) signVote(key *signatureschemes2.MiningKey) error {
	data := []byte{}
	data = append(data, s.BlockHash...)
	data = append(data, s.BLS...)
	data = append(data, s.BRI...)
	data = common.HashB(data)
	var err error
	s.Confirmation, err = key.BriSignData(data)
	return err
}

func (s *BFTVote) validateVoteOwner(ownerPk []byte) error {
	data := []byte{}
	data = append(data, s.BlockHash...)
	data = append(data, s.BLS...)
	data = append(data, s.BRI...)
	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, s.Confirmation, ownerPk)
	return err
}

func ExtractBridgeValidationData(block types.BlockInterface) ([][]byte, []int, error) {
	valData, err := consensustypes.DecodeValidationData(block.GetValidationField())
	if err != nil {
		return nil, nil, NewConsensusError(UnExpectedError, err)
	}
	return valData.BridgeSig, valData.ValidatiorsIdx, nil
}

func (e *BLSBFT_V3) sendVote(userKey *signatureschemes2.MiningKey, block types.BlockInterface, committees []incognitokey.CommitteePublicKey) error {
	Vote, err := CreateVote(userKey, block, committees)
	if err != nil {
		e.Logger.Error(err)
		return NewConsensusError(UnExpectedError, err)
	}

	msg, err := MakeBFTVoteMsg(Vote, e.ChainKey, e.currentTimeSlot, block.GetHeight())
	if err != nil {
		e.Logger.Error(err)
		return NewConsensusError(UnExpectedError, err)
	}
	e.voteHistory[block.GetHeight()] = block
	e.Logger.Info(e.ChainKey, "sending vote...")
	go e.Node.PushMessageToChain(msg, e.Chain)
	return nil
}
