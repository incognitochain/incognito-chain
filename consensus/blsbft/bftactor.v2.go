package blsbft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus/blsbft"
	"github.com/incognitochain/incognito-chain/consensus_v2"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wire"
	"sort"
	"sync"
	"time"
)

type BLSBFT_V2 struct {
	Chain    ChainInterface
	Node     blsbft.NodeInterface
	ChainKey string
	PeerID   string

	UserKeySet   *MiningKey
	BFTMessageCh chan wire.MessageBFT
	isStarted    bool
	StopCh       chan struct{}
	Logger       common.Logger

	lockOnGoingBlocks sync.RWMutex

	currentTimeSlot  uint64
	proposeHistory   *lru.Cache
	ProposeMessageCh chan BFTPropose
	VoteMessageCh    chan BFTVote

	receiveBlockByHeight map[uint64][]*ProposeBlockInfo   //blockHeight -> blockInfo
	receiveBlockByHash   map[string]*ProposeBlockInfo     //blockHash -> blockInfo
	voteHistory          map[uint64]common.BlockInterface // bestview height (previsous height )-> block
}

type ProposeBlockInfo struct {
	block      common.BlockInterface
	votes      map[string]BFTVote //pk->BFTVote
	isValid    bool
	hasNewVote bool
}

func (e *BLSBFT_V2) GetConsensusName() string {
	return consensusName
}

func (e *BLSBFT_V2) Stop() error {
	if e.isStarted {
		select {
		case <-e.StopCh:
			return nil
		default:
			close(e.StopCh)
		}
		e.isStarted = false
	}
	return NewConsensusError(ConsensusAlreadyStoppedError, errors.New(e.ChainKey))
}

func (e *BLSBFT_V2) Start() error {
	if e.isStarted {
		return NewConsensusError(ConsensusAlreadyStartedError, errors.New(e.ChainKey))
	}
	e.isStarted = true
	e.StopCh = make(chan struct{})
	e.ProposeMessageCh = make(chan BFTPropose)
	e.VoteMessageCh = make(chan BFTVote)
	e.receiveBlockByHash = make(map[string]*ProposeBlockInfo)
	e.receiveBlockByHeight = make(map[uint64][]*ProposeBlockInfo)
	e.voteHistory = make(map[uint64]common.BlockInterface)
	var err error
	e.proposeHistory, err = lru.New(1000)
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	//init view maps
	ticker := time.Tick(200 * time.Millisecond)
	e.Logger.Info("start bls-bftv2 consensus for chain", e.ChainKey)
	go func() {
		for { //actor loop

			//e.Logger.Debug("Current time ", currentTime, "time slot ", currentTimeSlot)
			select {
			case <-e.StopCh:
				return
			case proposeMsg := <-e.ProposeMessageCh:
				block, err := e.Chain.UnmarshalBlock(proposeMsg.Block)
				if err != nil {
					e.Logger.Info(err)
					continue
				}
				blkHash := block.Hash().String()
				if _, ok := e.receiveBlockByHash[blkHash]; !ok {
					e.receiveBlockByHash[blkHash] = &ProposeBlockInfo{
						block:      block,
						votes:      make(map[string]BFTVote),
						hasNewVote: false,
					}
					e.receiveBlockByHeight[block.GetHeight()] = append(e.receiveBlockByHeight[block.GetHeight()], e.receiveBlockByHash[blkHash])
				} else {
					e.receiveBlockByHash[blkHash].block = block
				}
				e.Logger.Debug("Receive block ", block.Hash().String(), "height", block.GetHeight(), ",block timeslot ", block.GetTimeslot())
				if block.GetHeight() <= e.Chain.GetBestView().GetHeight() {
					e.Logger.Debug("Send proposer to update views. Propose view Height less than latest height: ", block.GetHeight(), "<=", e.Chain.GetBestView().GetHeight())
					//e.Node.NotifyOutdatedView(proposeMsg.PeerID, e.Chain.GetBestView().Hash().String())
					//TODO: update peerstate syncker
				}

				_, err = e.Chain.GetViewByHash(block.GetPreviousBlockHash())
				if err != nil {
					e.Logger.Debugf("Request sync block from node %s from %s to %s", proposeMsg.PeerID, block.GetPreviousBlockHash().String())
					e.Node.RequestSyncBlock(proposeMsg.PeerID, e.Chain.GetFinalView().Hash().String(), block.GetPreviousBlockHash().String())
				}

			case voteMsg := <-e.VoteMessageCh:
				voteMsg.isValid = 0
				if b, ok := e.receiveBlockByHash[voteMsg.BlockHash]; ok { //if receiveblock is already initiated
					if _, ok := b.votes[voteMsg.Validator]; !ok { // and not receive validatorA vote
						b.votes[voteMsg.Validator] = voteMsg // store it
						b.hasNewVote = true
					}
				} else {
					e.receiveBlockByHash[voteMsg.BlockHash] = &ProposeBlockInfo{
						votes:      make(map[string]BFTVote),
						hasNewVote: true,
					}
					e.receiveBlockByHash[voteMsg.BlockHash].votes[voteMsg.Validator] = voteMsg
				}
				e.Logger.Infof("receive vote for block %s (%d)", voteMsg.BlockHash, len(e.receiveBlockByHash[voteMsg.BlockHash].votes))

			case <-ticker:
				//TODO: syncker module should tell it is ready or not
				e.lockOnGoingBlocks.Lock()
				e.currentTimeSlot = common.GetTimeSlot(e.Chain.GetGenesisTime(), time.Now().Unix(), TIMESLOT)
				bestView := e.Chain.GetBestView()

				/*
					Check for whether we should propose block
				*/
				proposerPk := bestView.GetNextProposer(e.currentTimeSlot)
				userPk := e.GetUserPublicKey().GetMiningKeyBase58(common.BlsConsensus2)

				if proposerPk == userPk && bestView.GetTimeslot() != e.currentTimeSlot { // current timeslot is not add to view, and this user is proposer of this timeslot
					//using block hash as key of best view -> check if this best view we propose or not
					if _, ok := e.proposeHistory.Get(fmt.Sprintf("%s%d", e.currentTimeSlot)); !ok {

						e.proposeHistory.Add(fmt.Sprintf("%s%d", e.currentTimeSlot), 1)
						//Proposer Rule: check propose block connected to bestview(longest chain rule 1) and re-propose valid block with smallest timestamp (including already propose in the past) (rule 2)
						sort.Slice(e.receiveBlockByHeight[bestView.GetHeight()+1], func(i, j int) bool {
							return e.receiveBlockByHeight[bestView.GetHeight()+1][i].block.GetBlockTimestamp() < e.receiveBlockByHeight[bestView.GetHeight()+1][j].block.GetBlockTimestamp()
						})
						var proposeBlock common.BlockInterface = nil
						for _, v := range e.receiveBlockByHeight[bestView.GetHeight()+1] {
							if v.isValid {
								proposeBlock = v.block
								break
							}
						}
						e.Logger.Debug("prepare proposer block")
						if createdBlk, err := e.proposeBlock(proposeBlock); err != nil {
							e.Logger.Critical(UnExpectedError, errors.New("can't propose block"))
						} else {
							e.Logger.Debug("proposer block", createdBlk.GetHeight(), "time slot ", e.currentTimeSlot, " with hash", createdBlk.Hash().String())
							//if propose block is not in cache list, create new one
							if _, ok := e.receiveBlockByHash[createdBlk.Hash().String()]; !ok {
								e.receiveBlockByHash[createdBlk.Hash().String()] = &ProposeBlockInfo{isValid: true, block: createdBlk, votes: make(map[string]BFTVote)}
							}

						}
					}
				}

				/*
					Check for valid block to vote
				*/
				validProposeBlock := []*ProposeBlockInfo{}
				//get all block that has height = bestview height  + 1(rule 2 & rule 3) (
				for h, proposeBlockInfo := range e.receiveBlockByHash {
					if proposeBlockInfo.block == nil {
						continue
					}
					bestViewHeight := bestView.GetHeight()
					if proposeBlockInfo.block.GetHeight() == bestViewHeight+1 {
						validProposeBlock = append(validProposeBlock, proposeBlockInfo)
					}

					if proposeBlockInfo.block.GetHeight() < e.Chain.GetFinalView().GetHeight() {
						delete(e.receiveBlockByHash, h)
					}
				}

				//rule 1: get history of vote for this height, vote if (round is lower than the vote before) or (round is equal but new proposer) or (there is no vote for this height yet)
				sort.Slice(validProposeBlock, func(i, j int) bool {
					return validProposeBlock[i].block.GetRound() < validProposeBlock[j].block.GetRound()
				})
				for _, v := range validProposeBlock {
					blkCreateTimeSlot := v.block.GetCreateTimeslot()
					bestViewHeight := bestView.GetHeight()
					if lastVotedBlk, ok := e.voteHistory[bestViewHeight+1]; ok {
						if blkCreateTimeSlot < lastVotedBlk.GetCreateTimeslot() { //blkCreateTimeSlot is smaller than voted block => vote for this blk
							e.validateAndVote(v)
						} else if blkCreateTimeSlot == lastVotedBlk.GetCreateTimeslot() && v.block.GetTimeslot() > lastVotedBlk.GetTimeslot() { //blk is old block (same round), but new proposer(larger timeslot) => vote again
							e.validateAndVote(v)
						} //blkCreateTimeSlot is larger or equal than voted block => do nothing
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

				e.lockOnGoingBlocks.Unlock()
			}
		}
	}()
	return nil
}

func (e BLSBFT_V2) NewInstance(chain consensus.ChainViewManagerInterface, chainKey string, node consensus.NodeInterface, logger common.Logger) consensus.ConsensusInterface {
	var newInstance BLSBFT_V2
	newInstance.Chain = chain
	newInstance.ChainKey = chainKey
	newInstance.Node = node
	newInstance.UserKeySet = e.UserKeySet
	newInstance.Logger = logger
	return &newInstance
}

func init() {
	consensus.RegisterConsensus(common.BlsConsensus2, &BLSBFT_V2{})
}

func (e *BLSBFT_V2) processIfBlockGetEnoughVote(k string, v *ProposeBlockInfo) {
	//no vote
	if v.hasNewVote == false {
		return
	}

	//no block
	if v.block == nil {
		return
	}

	//already in chain
	_, err := e.Chain.GetViewByHash(*v.block.Hash())
	if err == nil {
		return
	}

	//not connected
	view, err := e.Chain.GetViewByHash(v.block.GetPreviousBlockHash())
	if err != nil {
		return
	}

	validVote := 0
	errVote := 0
	for _, vote := range v.votes {
		dsaKey := []byte{}
		if vote.isValid == 0 {
			for _, c := range view.GetCommittee() {
				//e.Logger.Error(vote.Validator, c.GetMiningKeyBase58(common.BlsConsensus2))
				if vote.Validator == c.GetMiningKeyBase58(common.BlsConsensus2) {
					dsaKey = c.MiningPubKey[common.BridgeConsensus]
				}
			}
			if len(dsaKey) == 0 {
				e.Logger.Error("canot find dsa key")
			}
			err := vote.validateVoteOwner(dsaKey)
			if err != nil {
				e.Logger.Error(dsaKey)
				e.Logger.Error(err)
				panic(1)
				vote.isValid = -1
				errVote++
			} else {
				vote.isValid = 1
				validVote++
			}
		}
	}
	//e.Logger.Debug(validVote, len(view.GetCommittee()), errVote)
	v.hasNewVote = false
	if validVote > 2*len(view.GetCommittee())/3 {
		e.Logger.Debug("Commit block", v.block.GetHeight())

		committeeBLSString, err := incognitokey.ExtractPublickeysFromCommitteeKeyList(view.GetCommittee(), common.BlsConsensus2)
		//fmt.Println(committeeBLSString)
		if err != nil {
			e.Logger.Error(err)
			return
		}
		aggSig, brigSigs, validatorIdx, err := combineVotes(v.votes, committeeBLSString)
		if err != nil {
			e.Logger.Error(err)
			return
		}

		valData, err := DecodeValidationData(v.block.GetValidationField())
		if err != nil {
			e.Logger.Error(err)
			return
		}

		valData.AggSig = aggSig
		valData.BridgeSig = brigSigs
		valData.ValidatiorsIdx = validatorIdx
		validationDataString, _ := EncodeValidationData(*valData)
		if err := v.block.(blockValidation).AddValidationField(validationDataString); err != nil {
			e.Logger.Error(err)
			return
		}

		if err := e.Chain.ConnectBlockAndAddView(v.block); err != nil {
			e.Logger.Error("Cannot add block to view")
		}
		go e.Node.BroadCastBlock(v.block)

		delete(e.receiveBlockByHash, k)
	}
}

func (e *BLSBFT_V2) validateAndVote(v *ProposeBlockInfo) error {
	//not connected
	view, err := e.Chain.GetViewByHash(v.block.GetPreviousBlockHash())
	if err != nil {
		return err
	}

	//check block valid,
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if _, err := view.ValidateBlockAndCreateNewView(ctx, v.block, true); err != nil {
		e.Logger.Error(err)
		return err
	}

	//if valid then vote
	var Vote = new(BFTVote)
	bytelist := []blsmultisig.PublicKey{}
	selfIdx := 0
	userBLSPk := e.GetUserPublicKey().GetMiningKeyBase58(common.BlsConsensus2)
	for i, v := range e.Chain.GetBestView().GetCommittee() {
		if v.GetMiningKeyBase58(common.BlsConsensus2) == userBLSPk {
			selfIdx = i
		}
		bytelist = append(bytelist, v.MiningPubKey[common.BlsConsensus2])
	}

	blsSig, err := e.UserKeySet.BLSSignData(v.block.Hash().GetBytes(), selfIdx, bytelist)
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	bridgeSig := []byte{}
	if metadata.HasBridgeInstructions(v.block.GetInstructions()) {
		bridgeSig, err = e.UserKeySet.BriSignData(v.block.Hash().GetBytes())
		if err != nil {
			return consensus.NewConsensusError(consensus.UnExpectedError, err)
		}
	}
	Vote.BLS = blsSig
	Vote.BRI = bridgeSig
	Vote.BlockHash = v.block.Hash().String()

	userPk := e.UserKeySet.GetPublicKey()
	Vote.Validator = userPk.GetMiningKeyBase58(common.BlsConsensus2)
	Vote.PrevBlockHash = v.block.GetPreviousBlockHash().String()
	err = Vote.signVote(e.UserKeySet)
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}

	msg, err := MakeBFTVoteMsg(Vote, e.ChainKey, e.currentTimeSlot)
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	//e.Logger.Info("sending vote...")
	go e.Node.PushMessageToChain(msg, e.Chain)
	go func() {
		e.VoteMessageCh <- *Vote
	}()
	v.isValid = true
	e.voteHistory[e.Chain.GetBestView().GetHeight()+1] = v.block
	return nil
}

func (e *BLSBFT_V2) proposeBlock(block consensus.BlockInterface) (consensus.BlockInterface, error) {
	time1 := time.Now()
	if block == nil {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		block, _ = e.Chain.GetBestView().CreateNewBlock(ctx, e.currentTimeSlot, e.UserKeySet.GetPublicKeyBase58())
	} else {
		block = e.Chain.GetBestView().CreateBlockFromOldBlockData(block)
	}
	if block != nil {
		e.Logger.Info("create block", block.GetHeight(), time.Since(time1).Seconds())
	} else {
		e.Logger.Info("create block", time.Since(time1).Seconds())
		return nil, consensus.NewConsensusError(consensus.BlockCreationError, errors.New("block creation timeout"))
	}

	validationData := e.CreateValidationData(block)
	validationDataString, _ := EncodeValidationData(validationData)
	block.(blockValidation).AddValidationField(validationDataString)
	blockData, _ := json.Marshal(block)
	var proposeCtn = new(BFTPropose)
	proposeCtn.Block = blockData
	msg, _ := MakeBFTProposeMsg(proposeCtn, e.ChainKey, e.currentTimeSlot)
	go e.Node.PushMessageToChain(msg, e.Chain)

	return block, nil
}

func (e *BLSBFT_V2) ProcessBFTMsg(msg interface{}, sender consensus.NodeSender) {
	msgBFT := msg.(*wire.MessageBFT)
	switch msgBFT.Type {
	case MSG_PROPOSE:
		var msgPropose BFTPropose
		err := json.Unmarshal(msgBFT.Content, &msgPropose)
		if err != nil {
			e.Logger.Error(err)
			return
		}
		msgPropose.PeerID = sender.GetID()
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

func (e *BLSBFT_V2) preValidateVote(blockHash []byte, Vote *BFTVote, candidate []byte) error {
	data := []byte{}
	data = append(data, blockHash...)
	data = append(data, Vote.BLS...)
	data = append(data, Vote.BRI...)
	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, Vote.Confirmation, candidate)
	return err
}

func (s *BFTVote) signVote(key *MiningKey) error {
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

func (e *BLSBFT_V2) ExtractBridgeValidationData(block consensus.BlockInterface) ([][]byte, []int, error) {
	valData, err := DecodeValidationData(block.GetValidationField())
	if err != nil {
		return nil, nil, consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	return valData.BridgeSig, valData.ValidatiorsIdx, nil
}
