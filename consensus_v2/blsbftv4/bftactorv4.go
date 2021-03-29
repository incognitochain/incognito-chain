package blsbftv4

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"

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

type BLSBFT_V4 struct {
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

	bodyHashes map[uint64]map[string]bool // TODO: @tin use this property for optimizing producing block process
}

func (e BLSBFT_V4) GetChainKey() string {
	return e.ChainKey
}

func (e BLSBFT_V4) GetChainID() int {
	return e.ChainID
}

func (e BLSBFT_V4) IsOngoing() bool {
	return e.isStarted
}

func (e BLSBFT_V4) IsStarted() bool {
	return e.isStarted
}

func (e *BLSBFT_V4) GetConsensusName() string {
	return common.BlsConsensus
}

func (e *BLSBFT_V4) Stop() error {
	if e.isStarted {
		e.Logger.Info("stop bls-bftv4 consensus for chain", e.ChainKey)
	}
	e.isStarted = false
	return nil
}

func (e *BLSBFT_V4) Start() error {
	if !e.isStarted {
		e.Logger.Info("start bls-bftv4 consensus for chain", e.ChainKey)
	}
	e.isStarted = true
	return nil
}

func (e *BLSBFT_V4) Destroy() {
	close(e.destroyCh)
}

func (e *BLSBFT_V4) getCommitteeForBlock(v types.BlockInterface) ([]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error) {
	var err error = nil
	var committees, signingCommittees []incognitokey.CommitteePublicKey
	if !e.Chain.IsBeaconChain() {
		signingCommittees, committees, err = e.CommitteeChain.CommitteesFromViewHashForShard(v.CommitteeFromBlock(), byte(e.Chain.GetShardID()), committeestate.MaxSubsetCommittees)
	} else {
		committees = e.Chain.GetBestView().GetCommittee()
		signingCommittees = committees
	}
	return signingCommittees, committees, err
}

func (e *BLSBFT_V4) getUserKeySetForSigning(
	committees []incognitokey.CommitteePublicKey, userKeySet []signatureschemes2.MiningKey,
) []signatureschemes2.MiningKey {
	res := []signatureschemes2.MiningKey{}
	if e.Chain.IsBeaconChain() {
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

func (e *BLSBFT_V4) verifyVersion() bool {
	if e.Chain.IsBeaconChain() {
		return e.Chain.CommitteeEngineVersion() == committeestate.DCS_VERSION
	} else {
		return e.CommitteeChain.CommitteeEngineVersion() == committeestate.DCS_VERSION
	}
}

func (e *BLSBFT_V4) run() error {
	go func() {
		//init view maps
		ticker := time.Tick(200 * time.Millisecond)
		e.Logger.Info("init bls-bftv4 consensus for chain", e.ChainKey)

		for { //actor loop
			if !e.verifyVersion() {
				continue
			}
			if !e.isStarted { //sleep if this process is not start
				time.Sleep(time.Second)
				continue
			}
			select {
			case <-e.destroyCh:
				e.Logger.Info("exit bls-bftv4 consensus for chain", e.ChainKey)
				return
			case proposeMsg := <-e.ProposeMessageCh:
				//fmt.Println("debug receive propose message", string(proposeMsg.Block))
				blockIntf, err := e.Chain.UnmarshalBlock(proposeMsg.Block)
				if err != nil || blockIntf == nil {
					e.Logger.Info(err)
					continue
				}
				block := blockIntf.(types.BlockInterface)
				blkHash := block.Hash().String()

				signingCommittees, committees, err := e.getCommitteeForBlock(block)
				if err != nil {
					e.Logger.Error(err)
					continue
				}

				userKeySet := e.getUserKeySetForSigning(signingCommittees, e.UserKeySet)
				if len(userKeySet) == 0 {
					e.Logger.Info("Not in round for voting")
				}

				committeesStr, _ := incognitokey.CommitteeKeyListToString(signingCommittees)
				e.Logger.Infof("######### Shard %+v, BlockHash %+v, BlockHeight %+v, committees %+v committeesFromBlock %+v",
					e.Chain.GetShardID(), block.Hash().String(), block.GetHeight(), committeesStr, block.CommitteeFromBlock().String())

				if _, ok := e.receiveBlockByHash[blkHash]; !ok {
					proposeBlockInfo := newProposeBlockForProposeMsg(block, committees, signingCommittees, userKeySet, make(map[string]*BFTVote))
					e.receiveBlockByHash[blkHash] = proposeBlockInfo
					e.Logger.Info("Receive block ", block.Hash().String(), "height", block.GetHeight(), ",block timeslot ", common.CalculateTimeSlot(block.GetProposeTime()))
					e.receiveBlockByHeight[block.GetHeight()] = append(e.receiveBlockByHeight[block.GetHeight()], e.receiveBlockByHash[blkHash])
				} else {
					e.receiveBlockByHash[blkHash].addBlockInfo(block, committees, signingCommittees, userKeySet, 0, 0)
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
						e.Logger.Infof("Receive vote for block %s (%d) from %v", voteMsg.BlockHash, len(e.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.Validator)
						b.hasNewVote = true
					}
				} else {
					e.receiveBlockByHash[voteMsg.BlockHash] = newBlockInfoForVoteMsg()
					e.receiveBlockByHash[voteMsg.BlockHash].votes[voteMsg.Validator] = &voteMsg
					e.Logger.Infof("[Monitor] receive vote for block %s (%d) from %v", voteMsg.BlockHash, len(e.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.Validator)
				}
				// e.Logger.Infof("receive vote for block %s (%d)", voteMsg.BlockHash, len(e.receiveBlockByHash[voteMsg.BlockHash].votes))

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
				var userProposeKey signatureschemes2.MiningKey
				shouldPropose := false
				shouldListen := true

				signingCommittees, _, proposerPk, committeeViewHash, err := e.getCommitteesAndCommitteeViewHash()
				if err != nil {
					e.Logger.Info(err)
					continue
				}
				userKeySet := e.getUserKeySetForSigning(signingCommittees, e.UserKeySet)
				for _, userKey := range userKeySet {
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

				if newTimeSlot && len(userKeySet) != 0 { //for logging
					e.Logger.Info("")
					e.Logger.Info("======================================================")
					e.Logger.Info("")
					if shouldListen {
						e.Logger.Infof("%v TS: %v, LISTEN BLOCK %v, ROUND %v", e.ChainKey, common.CalculateTimeSlot(e.currentTime), bestView.GetHeight()+1, e.currentTimeSlot-common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()))
					}
					if shouldPropose {
						e.Logger.Infof("%v TS: %v, PROPOSE BLOCK %v, ROUND %v", e.ChainKey, common.CalculateTimeSlot(e.currentTime), bestView.GetHeight()+1, e.currentTimeSlot-common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()))
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

					if createdBlk, err := e.proposeBlock(userProposeKey, proposerPk, proposeBlock, signingCommittees, committeeViewHash); err != nil {
						e.Logger.Critical(UnExpectedError, errors.New("can't propose block"))
						e.Logger.Critical(err)
					} else {
						e.Logger.Infof("proposer block %v round %v time slot %v blockTimeSlot %v with hash %v", createdBlk.GetHeight(), createdBlk.GetRound(), e.currentTimeSlot, common.CalculateTimeSlot(createdBlk.GetProduceTime()), createdBlk.Hash().String())
					}
				}

				//Check for valid block to vote
				validProposeBlock := []*ProposeBlockInfo{}
				//get all block that has height = bestview height  + 1(rule 2 & rule 3) (
				for h, proposeBlockInfo := range e.receiveBlockByHash {
					if proposeBlockInfo.block == nil {
						continue
					}
					bestViewHeight := bestView.GetHeight()
					// e.Logger.Infof("[Monitor] bestview height %v, finalview height %v, block height %v %v", bestViewHeight, e.Chain.GetFinalView().GetHeight(), proposeBlockInfo.block.GetHeight(), proposeBlockInfo.block.GetProduceTime())
					if proposeBlockInfo.block.GetHeight() == bestViewHeight+1 {
						validProposeBlock = append(validProposeBlock, proposeBlockInfo)
					} else {
						if proposeBlockInfo.block.Hash().String() == bestView.GetHash().String() {
							validProposeBlock = append(validProposeBlock, proposeBlockInfo)
						}
					}

					if proposeBlockInfo.block.GetHeight() < e.Chain.GetFinalView().GetHeight() {
						delete(e.receiveBlockByHash, h)
						delete(e.bodyHashes, proposeBlockInfo.block.GetHeight())
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

				for k, v := range e.receiveBlockByHash {
					if len(v.userKeySet) != 0 {
						e.processIfBlockGetEnoughVote(k, v)
					}
				}
			}
		}
	}()
	return nil
}

func NewInstance(chain ChainInterface, committeeChain CommitteeChainHandler, chainKey string, chainID int, node NodeInterface, logger common.Logger) *BLSBFT_V4 {
	var err error
	var newInstance = new(BLSBFT_V4)
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
	newInstance.bodyHashes = make(map[uint64]map[string]bool)
	newInstance.proposeHistory, err = lru.New(1000)
	if err != nil {
		panic(err) //must not error
	}
	newInstance.CommitteeChain = committeeChain
	newInstance.run()
	return newInstance
}

func (e *BLSBFT_V4) validateVotes(v *ProposeBlockInfo) *ProposeBlockInfo {
	validVote := 0
	errVote := 0

	committees := make(map[string]int)
	if len(v.votes) != 0 {
		for i, v := range v.signingCommittes {
			committees[v.GetMiningKeyBase58(common.BlsConsensus)] = i
		}
	}

	for _, vote := range v.votes {
		dsaKey := []byte{}
		if vote.IsValid == 0 {
			if value, ok := committees[vote.Validator]; ok {
				dsaKey = v.signingCommittes[value].MiningPubKey[common.BridgeConsensus]
			}
			if len(dsaKey) == 0 {
				e.Logger.Error("canot find dsa key")
			}
			err := vote.validateVoteOwner(dsaKey)
			if err != nil {
				e.Logger.Infof("key %+v can not vote for committees %+v \n", vote.Validator, committees)
				vote.IsValid = -1
				errVote++
				panic(err)
			} else {
				vote.IsValid = 1
				validVote++
			}
		} else {
			validVote++
		}
	}
	e.Logger.Info("Number of Valid Vote", validVote, "| Number Of Error Vote", errVote)
	v.hasNewVote = false
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
		validVote, errVote,
	)

	return v
}

func (e *BLSBFT_V4) processIfBlockGetEnoughVote(
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
	//e.Logger.Infof("Process Block With enough votes, %+v, %+v", *v.block.Hash(), v.block.GetHeight())
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
		e.Logger.Infof("Get Previous View By Hash Fail, %+v, %+v", v.block.GetPrevHash(), v.block.GetHeight()-1)
		return
	}
	v = e.validateVotes(v)

	if v.validVotes > 2*len(v.signingCommittes)/3 && !v.isCommitted {
		v.isCommitted = true
		e.Logger.Infof("Commit block %v , height: %v", blockHash, v.block.GetHeight())
		if e.ChainID == BEACON_CHAIN_ID {
			e.processWithEnoughVotesBeaconChain(v)
		} else {
			e.processWithEnoughVotesShardChain(v)
		}
	}
}

func (e *BLSBFT_V4) processWithEnoughVotesBeaconChain(
	v *ProposeBlockInfo,
) {
	validationData, err := createBLSAggregatedSignatures(v.signingCommittes, v.block.GetValidationField(), v.votes)
	if err != nil {
		e.Logger.Error(err)
		return
	}
	v.block.(blockValidation).AddValidationField(validationData)

	go e.Chain.InsertAndBroadcastBlock(v.block)

	delete(e.receiveBlockByHash, v.block.GetPrevHash().String())
}

func (e *BLSBFT_V4) processWithEnoughVotesShardChain(
	v *ProposeBlockInfo,
) {
	// validationData at present block
	validationData, err := createBLSAggregatedSignatures(v.signingCommittes, v.block.GetValidationField(), v.votes)
	if err != nil {
		e.Logger.Error(err)
		return
	}
	v.block.(blockValidation).AddValidationField(validationData)

	// validate and previous block
	if previousProposeBlockInfo, ok := e.receiveBlockByHash[v.block.GetPrevHash().String()]; ok &&
		previousProposeBlockInfo != nil && previousProposeBlockInfo.block != nil {

		signingCommittes, committees, err := e.getCommitteeForBlock(previousProposeBlockInfo.block)
		if err != nil {
			e.Logger.Error(err)
		}
		previousProposeBlockInfo.addBlockInfo(
			previousProposeBlockInfo.block, committees, signingCommittes,
			previousProposeBlockInfo.userKeySet,
			previousProposeBlockInfo.validVotes, previousProposeBlockInfo.errVotes,
		)
		previousProposeBlockInfo = e.validateVotes(previousProposeBlockInfo)

		previousValidationData, err := createBLSAggregatedSignatures(
			previousProposeBlockInfo.signingCommittes,
			previousProposeBlockInfo.block.GetValidationField(), previousProposeBlockInfo.votes)
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

func (e *BLSBFT_V4) validateAndVote(
	v *ProposeBlockInfo,
) error {

	//already vote for this proposed block
	if v.isVoted {
		return nil
	}

	//already validate and vote for this proposed block
	if !v.isValid {
		//not connected
		view := e.Chain.GetViewByHash(v.block.GetPrevHash())
		if view == nil {
			e.Logger.Info("view is null")
			return errors.New("View not connect")
		}

		if _, ok := e.bodyHashes[v.block.GetHeight()][v.block.BodyHash().String()]; !ok {
			//TODO: using context to validate block
			_, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			e.Logger.Infof("validateAndVote for block: %+v \n", v.block.Hash().String())
			if err := e.Chain.ValidatePreSignBlock(v.block, v.signingCommittes, v.committees); err != nil { //@tin fix here
				e.Logger.Error(err)
				return err
			}

			// Block is valid for commit
			v.isValid = true
			if len(e.bodyHashes[v.block.GetHeight()]) == 0 {
				e.bodyHashes[v.block.GetHeight()] = make(map[string]bool)
			}
			e.bodyHashes[v.block.GetHeight()][v.block.BodyHash().String()] = true
		}
	}

	//if valid then vote
	for _, userKey := range v.userKeySet {
		vote, err := CreateVote(&userKey, v.block, v.signingCommittes)
		if err != nil {
			e.Logger.Error("error:", err)
			return NewConsensusError(UnExpectedError, err)
		}

		msg, err := MakeBFTVoteMsg(vote, e.ChainKey, e.currentTimeSlot, v.block.GetHeight())
		if err != nil {
			e.Logger.Error("error:", err)
			return NewConsensusError(UnExpectedError, err)
		}
		e.voteHistory[v.block.GetHeight()] = v.block
		e.Logger.Info(e.ChainKey, "sending vote...")
		go e.ProcessBFTMsg(msg.(*wire.MessageBFT))
		go e.Node.PushMessageToChain(msg, e.Chain)
		v.isVoted = true
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

func (e *BLSBFT_V4) proposeBlock(
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

func (e *BLSBFT_V4) proposeBeaconBlock(
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
		block, err = e.Chain.CreateNewBlock(4, b58Str, 1, e.currentTime, committees, committeeViewHash)
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

func (e *BLSBFT_V4) proposeShardBlock(
	b58Str string,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
) (types.BlockInterface, error) {
	var err error
	var newBlock types.BlockInterface

	// propose new block when
	// no previous proposed block
	// or previous proposed block has different committee with new committees
	if block == nil {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, (time.Duration(common.TIMESLOT)*time.Second)/2)
		defer cancel()
		newBlock, err = e.Chain.CreateNewBlock(4, b58Str, 1, e.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
		e.Logger.Infof("proposeShardBlock hash %+v height %+v Timestamp: %+v \n", newBlock.Hash().String(), newBlock.GetHeight(), newBlock.GetProduceTime())
	} else {
		newBlock, err = e.Chain.CreateNewBlockFromOldBlock(block, b58Str, e.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
		e.Logger.Infof("CreateNewBlockFromOldBlock, Block Height %+v Hash %+v Timestamp %+v", newBlock.GetHeight(), newBlock.Hash().String(), newBlock.GetProduceTime())
	}
	return newBlock, err
}

func (e *BLSBFT_V4) ProcessBFTMsg(msgBFT *wire.MessageBFT) {
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

func (e *BLSBFT_V4) preValidateVote(blockHash []byte, Vote *BFTVote, candidate []byte) error {
	data := []byte{}
	data = append(data, blockHash...)
	data = append(data, Vote.BLS...)
	data = append(data, Vote.BRI...)
	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, Vote.Confirmation, candidate)
	return err
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

func (e *BLSBFT_V4) getCommitteesAndCommitteeViewHash() (
	[]incognitokey.CommitteePublicKey,
	[]incognitokey.CommitteePublicKey,
	incognitokey.CommitteePublicKey, common.Hash, error,
) {
	committeeViewHash := common.Hash{}
	committees := []incognitokey.CommitteePublicKey{}
	signingCommittees := []incognitokey.CommitteePublicKey{}
	proposerPk := incognitokey.CommitteePublicKey{}
	var err error
	if e.ChainID == BEACON_CHAIN_ID {
		proposerPk, _ = e.Chain.GetBestView().GetProposerByTimeSlot(e.currentTimeSlot, 2)
		committees = e.Chain.GetBestView().GetCommittee()
	} else {
		committeeViewHash = *e.CommitteeChain.FinalView().GetHash()
		signingCommittees, committees, err = e.
			CommitteeChain.
			CommitteesFromViewHashForShard(committeeViewHash, byte(e.ChainID), committeestate.MaxSubsetCommittees)
		if err != nil {
			return signingCommittees, committees, proposerPk, committeeViewHash, err
		}
		proposerPk = e.CommitteeChain.ProposerByTimeSlot(byte(e.ChainID), e.currentTimeSlot, committees)
	}
	return signingCommittees, committees, proposerPk, committeeViewHash, err
}
