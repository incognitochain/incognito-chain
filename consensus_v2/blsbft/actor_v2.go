package blsbft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb_consensus"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metrics/monitor"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
	"github.com/incognitochain/incognito-chain/wire"
)

type actorV2 struct {
	chain    Chain
	node     NodeInterface
	chainKey string
	chainID  int
	peerID   string

	userKeySet       []signatureschemes2.MiningKey
	bftMessageCh     chan wire.MessageBFT
	proposeMessageCh chan BFTPropose
	voteMessageCh    chan BFTVote

	isStarted bool
	destroyCh chan struct{}
	logger    common.Logger

	committeeChain  CommitteeChainHandler
	currentTime     int64
	currentTimeSlot int64

	proposeHistory     map[int64]struct{}
	receiveBlockByHash map[string]*ProposeBlockInfo    //blockHash -> blockInfo
	voteHistory        map[uint64]types.BlockInterface // bestview height (previsous height )-> block

	ruleDirector         *ActorV2RuleDirector
	blockVersion         int
	shouldPreparePropose bool
}

func NewActorV2() *actorV2 {
	return &actorV2{}
}

func NewActorV2WithValue(
	chain Chain,
	committeeChain CommitteeChainHandler,
	chainKey string, blockVersion, chainID int,
	node NodeInterface, logger common.Logger,
) *actorV2 {
	a := newActorV2WithValue(
		chain,
		committeeChain,
		chainKey,
		blockVersion,
		chainID,
		node,
		logger,
	)

	a.run()

	return a
}

func newActorV2WithValue(
	chain Chain,
	committeeChain CommitteeChainHandler,
	chainKey string, blockVersion, chainID int,
	node NodeInterface, logger common.Logger,
) *actorV2 {
	var err error
	a := NewActorV2()
	a.chain = chain
	a.chainKey = chainKey
	a.chainID = chainID
	a.node = node
	a.logger = logger
	a.destroyCh = make(chan struct{})
	a.proposeMessageCh = make(chan BFTPropose)
	a.voteMessageCh = make(chan BFTVote)
	a.proposeHistory, err = InitProposeHistory(chainID)
	if err != nil {
		panic(err) //must not error
	}
	a.receiveBlockByHash, err = InitReceiveBlockByHash(chainID)
	if err != nil {
		panic(err) //must not error
	}
	a.voteHistory, err = InitVoteHistory(chainID)
	if err != nil {
		panic(err) //must not error
	}
	a.committeeChain = committeeChain
	a.blockVersion = blockVersion
	SetBuilderContext(config.Param().ConsensusParam.Lemma2Height)
	a.ruleDirector = NewActorV2RuleDirector()
	a.ruleDirector.initRule(ActorRuleBuilderContext, a.chain.GetBestView().GetBeaconHeight(), chain, logger)
	if err != nil {
		panic(err) //must not error
	}
	return a
}

func (a *actorV2) GetSortedReceiveBlockByHeight(blockHeight uint64) []*ProposeBlockInfo {
	tmp := []*ProposeBlockInfo{}
	for _, proposeInfo := range a.receiveBlockByHash {
		if proposeInfo.block.GetHeight() == blockHeight {
			tmp = append(tmp, proposeInfo)
		}
	}
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].block.GetProduceTime() < tmp[j].block.GetProduceTime()
	})
	return tmp
}

func InitReceiveBlockByHash(chainID int) (map[string]*ProposeBlockInfo, error) {

	data, err := rawdb_consensus.GetAllReceiveBlockByHash(
		rawdb_consensus.GetConsensusDatabase(),
		chainID,
	)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*ProposeBlockInfo)

	for k, v := range data {
		var block types.BlockInterface
		if chainID == common.BeaconChainID {
			block = types.NewBeaconBlock()
		} else {
			block = types.NewShardBlock()
		}
		proposeBlockInfo := &ProposeBlockInfo{
			block: block,
		}
		err := json.Unmarshal(v, proposeBlockInfo)
		if err != nil {
			return nil, err
		}

		//restore votes by block hash
		votes, prevote, err := GetVotesByBlockHashFromDB(proposeBlockInfo.block.ProposeHash().String())
		if err != nil {
			return nil, err
		}

		proposeBlockInfo.Votes = votes
		proposeBlockInfo.PreVotes = prevote
		res[k] = proposeBlockInfo
	}

	return res, nil
}

func AddVoteByBlockHashToDB(blockHash string, bftVote BFTVote) error {
	data, err := json.Marshal(bftVote)
	if err != nil {
		return err
	}

	if err = rawdb_consensus.StoreVoteByBlockHash(rawdb_consensus.GetConsensusDatabase(), blockHash, bftVote.Validator, data); err != nil {
		return err
	}

	return nil
}

func (a *actorV2) AddReceiveBlockByHash(blockHash string, proposeBlockInfo *ProposeBlockInfo) error {

	a.receiveBlockByHash[blockHash] = proposeBlockInfo

	data, err := json.Marshal(proposeBlockInfo)
	if err != nil {
		return err
	}

	if err := rawdb_consensus.StoreReceiveBlockByHash(
		rawdb_consensus.GetConsensusDatabase(),
		a.chainID,
		blockHash,
		data,
	); err != nil {
		return err
	}
	return nil
}

func GetVotesByBlockHashFromDB(proposeHash string) (map[string]*BFTVote, map[string]*BFTVote, error) {
	votes, err := rawdb_consensus.GetVotesByBlockHash(rawdb_consensus.GetConsensusDatabase(), proposeHash)
	if err != nil {
		return nil, nil, err
	}
	vote := map[string]*BFTVote{}
	preVote := map[string]*BFTVote{}
	for validator, vData := range votes {
		v := &BFTVote{}
		err := json.Unmarshal(vData, v)
		if err != nil {
			continue
		}
		if v.Phase == "" || v.Phase == "vote" {
			vote[validator] = v
		} else {
			preVote[validator] = v
		}

	}
	return vote, preVote, nil
}

func (a *actorV2) GetReceiveBlockByHash(blockHash string) (*ProposeBlockInfo, bool) {
	res, ok := a.receiveBlockByHash[blockHash]
	return res, ok
}

func (a *actorV2) CleanReceiveBlockByHash(blockHash string) error {

	if err := rawdb_consensus.DeleteReceiveBlockByHash(
		rawdb_consensus.GetConsensusDatabase(),
		a.chainID,
		blockHash,
	); err != nil {
		return err
	}

	delete(a.receiveBlockByHash, blockHash)

	if err := rawdb_consensus.DeleteVotesByHash(rawdb_consensus.GetConsensusDatabase(), blockHash); err != nil {
		return err
	}
	return nil
}

func InitVoteHistory(chainID int) (map[uint64]types.BlockInterface, error) {

	data, err := rawdb_consensus.GetAllVoteHistory(
		rawdb_consensus.GetConsensusDatabase(),
		chainID,
	)
	if err != nil {
		return nil, err
	}

	res := make(map[uint64]types.BlockInterface)

	for k, v := range data {
		if chainID == common.BeaconChainID {
			block := &types.BeaconBlock{}
			err := json.Unmarshal(v, block)
			if err != nil {
				return nil, err
			}
			res[k] = block
		} else {
			block := &types.ShardBlock{}
			err := json.Unmarshal(v, block)
			if err != nil {
				return nil, err
			}
			res[k] = block
		}
	}

	return res, nil
}

func (a *actorV2) AddVoteHistory(blockHeight uint64, block types.BlockInterface) error {

	a.voteHistory[blockHeight] = block

	var data []byte
	var err error
	if a.chainID == common.BeaconChainID {
		data, err = json.Marshal(block.(*types.BeaconBlock))
		if err != nil {
			return err
		}
	} else {
		data, err = json.Marshal(block.(*types.ShardBlock))
		if err != nil {
			return err
		}
	}

	if err := rawdb_consensus.StoreVoteHistory(
		rawdb_consensus.GetConsensusDatabase(),
		a.chainID,
		blockHeight,
		data,
	); err != nil {
		return err
	}

	return nil
}

func (a *actorV2) GetVoteHistory(blockHeight uint64) (types.BlockInterface, bool) {
	res, ok := a.voteHistory[blockHeight]
	return res, ok
}

func (a *actorV2) CleanVoteHistory(blockHeight uint64) error {

	if err := rawdb_consensus.DeleteVoteHistory(
		rawdb_consensus.GetConsensusDatabase(),
		a.chainID,
		blockHeight,
	); err != nil {
		return err
	}
	delete(a.voteHistory, blockHeight)

	return nil
}

func InitProposeHistory(chainID int) (map[int64]struct{}, error) {

	data, err := rawdb_consensus.GetAllProposeHistory(
		rawdb_consensus.GetConsensusDatabase(),
		chainID,
	)
	if err != nil {
		return nil, err
	}

	res := make(map[int64]struct{})

	for k := range data {
		res[k] = struct{}{}
	}

	return res, nil
}

func (a *actorV2) AddCurrentTimeSlotProposeHistory() error {

	a.proposeHistory[a.currentTimeSlot] = struct{}{}

	if err := rawdb_consensus.StoreProposeHistory(
		rawdb_consensus.GetConsensusDatabase(),
		a.chainID,
		a.currentTimeSlot,
	); err != nil {
		return err
	}

	return nil
}

func (a *actorV2) GetCurrentTimeSlotProposeHistory() bool {
	_, ok := a.proposeHistory[a.currentTimeSlot]
	return ok
}

func (a *actorV2) CleanProposeHistory(timeSlot int64) error {

	if err := rawdb_consensus.DeleteProposeHistory(
		rawdb_consensus.GetConsensusDatabase(),
		a.chainID,
		timeSlot,
	); err != nil {
		return err
	}

	delete(a.proposeHistory, timeSlot)

	return nil
}

func (a actorV2) GetConsensusName() string {
	return common.BlsConsensus
}

func (a actorV2) GetChainKey() string {
	return a.chainKey
}

func (a actorV2) GetChainID() int {
	return a.chainID
}

func (a actorV2) GetUserPublicKey() *incognitokey.CommitteePublicKey {
	if a.userKeySet != nil {
		key := a.userKeySet[0].GetPublicKey()
		return key
	}
	return nil
}

func (a actorV2) BlockVersion() int {
	return a.blockVersion
}

func (a *actorV2) SetBlockVersion(version int) {
	a.blockVersion = version
}

func (a *actorV2) Stop() error {
	if a.isStarted {
		a.logger.Infof("stop bls-bft-%+v consensus for chain %+v", a.blockVersion, a.chainKey)
	}
	a.isStarted = false
	return nil
}

func (a *actorV2) Destroy() {
	a.isStarted = false
	a.destroyCh <- struct{}{}
}

func (a actorV2) IsStarted() bool {
	return a.isStarted
}

func (a *actorV2) ProcessBFTMsg(msgBFT *wire.MessageBFT) {

	switch msgBFT.Type {
	case MSG_PROPOSE:
		var msgPropose BFTPropose
		err := json.Unmarshal(msgBFT.Content, &msgPropose)
		if err != nil {
			a.logger.Error(err)
			return
		}
		msgPropose.PeerID = msgBFT.PeerID
		a.proposeMessageCh <- msgPropose
	case MSG_VOTE:
		var msgVote BFTVote
		err := json.Unmarshal(msgBFT.Content, &msgVote)
		if err != nil {
			a.logger.Error(err)
			return
		}
		a.voteMessageCh <- msgVote
	default:
		a.logger.Criticalf("Unknown BFT message type %+v", msgBFT)
		return
	}
}

func (a *actorV2) LoadUserKeys(miningKey []signatureschemes2.MiningKey) {
	a.userKeySet = miningKey
	return
}

func (a actorV2) ValidateData(data []byte, sig string, publicKey string) error {
	sigByte, _, err := base58.Base58Check{}.Decode(sig)
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	publicKeyByte := []byte(publicKey)
	dataHash := new(common.Hash)
	dataHash.NewHash(data)
	_, err = bridgesig.Verify(publicKeyByte, dataHash.GetBytes(), sigByte) //blsmultisig.Verify(sigByte, data, []int{0}, []blsmultisig.PublicKey{publicKeyByte})
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	return nil
}

func (a *actorV2) SignData(data []byte) (string, error) {
	//, 0, []blsmultisig.PublicKey{e.UserKeySet.PubKey[common.BlsConsensus]})
	result, err := a.userKeySet[0].BriSignData(data)
	if err != nil {
		return "", NewConsensusError(SignDataError, err)
	}

	return base58.Base58Check{}.Encode(result, common.Base58Version), nil
}

func (a *actorV2) Start() error {
	if !a.isStarted {
		a.logger.Infof("start bls-bft-%+v consensus for chain %+v", a.blockVersion, a.chainKey)
	}
	a.isStarted = true
	return nil
}

func (a *actorV2) closeActor() {
	close(a.destroyCh)
	close(a.proposeMessageCh)
	close(a.voteMessageCh)
}

func (a *actorV2) run() error {
	go func() {
		//init view maps
		ticker := time.Tick(200 * time.Millisecond)
		cleanMemTicker := time.Tick(5 * time.Minute)
		a.logger.Infof("init bls-bft-%+v consensus for chain %+v", a.blockVersion, a.chainKey)
		for { //actor loop
			if !a.isStarted { //sleep if this process is not start
				time.Sleep(time.Second)
				select {
				case <-a.proposeMessageCh:
				case <-a.voteMessageCh:
				case <-a.destroyCh:
					a.logger.Infof("exit bls-bft-%+v consensus for chain %+v", a.blockVersion, a.chainKey)
					a.closeActor()
					return
				default:
				}

				continue
			}

			a.ruleDirector.updateRule(
				ActorRuleBuilderContext,
				a.ruleDirector.builder,
				a.chain.GetBestView().GetBeaconHeight(),
				a.chain,
				a.logger,
			)

			select {
			case <-a.destroyCh:
				a.logger.Infof("exit bls-bft-%+v consensus for chain %+v", a.blockVersion, a.chainKey)
				a.closeActor()
				return
			case proposeMsg := <-a.proposeMessageCh:
				err := a.handleProposeMsg(proposeMsg)
				if err != nil {
					a.logger.Error(err)
					continue
				}

			case voteMsg := <-a.voteMessageCh:
				err := a.handleVoteMsg(voteMsg)
				if err != nil {
					a.logger.Debug(err)
					continue
				}

			case <-cleanMemTicker:
				a.handleCleanMem()
				continue

			case <-ticker:
				if !a.chain.IsReady() {
					continue
				}
				a.currentTime = time.Now().Unix()
				bestView := a.chain.GetBestView()
				currentTimeSlot := bestView.CalculateTimeSlot(a.currentTime)
				if currentTimeSlot == bestView.GetCurrentTimeSlot() || bestView.PastHalfTimeslot(a.currentTime) {
					if a.shouldPreparePropose {
						a.chain.CollectTxs()
					}
				}

				newTimeSlot := false
				if a.currentTimeSlot != currentTimeSlot {
					newTimeSlot = true
				}

				a.currentTimeSlot = currentTimeSlot

				//set round for monitor
				round := a.currentTimeSlot - bestView.CalculateTimeSlot(bestView.GetBlock().GetProposeTime())
				monitor.SetGlobalParam("RoundKey", fmt.Sprintf("%d_%d", bestView.GetHeight(), round))

				signingCommittees, committees, proposerPk, committeeViewHash, proposerIndex, err := a.getCommitteesAndCommitteeViewHash()
				if err != nil {
					a.logger.Info(err)
					continue
				}

				userKeySet, prepareProposerIndex := a.getUserKeySetForSigning(signingCommittees, a.userKeySet, proposerIndex)
				if prepareProposerIndex != -1 {
					a.shouldPreparePropose = true
				}

				shouldListen, shouldPropose, userProposeKey := a.isUserKeyProposer(
					bestView.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()),
					proposerPk,
					userKeySet,
				)

				if newTimeSlot { //for logging
					a.logger.Info("")
					a.logger.Info("======================================================")
					a.logger.Info("")
					if shouldListen {
						a.logger.Infof("%v TS: %v, LISTEN BLOCK %v, Round %v", a.chainKey, a.currentTimeSlot, bestView.GetHeight()+1, round)
					}
					if shouldPropose {
						a.logger.Infof("%v TS: %v, PROPOSE BLOCK %v, Round %v", a.chainKey, a.currentTimeSlot, bestView.GetHeight()+1, round)
					}
				}

				if shouldPropose {
					if err := a.AddCurrentTimeSlotProposeHistory(); err != nil {
						a.logger.Errorf("add current time slot propose history")
					}
					// Proposer Rule: check propose block connected to bestview (longest chain rule 1)
					// and re-propose valid block with smallest timestamp (including already propose in the past) (rule 2)
					var proposeBlockInfo = NewProposeBlockInfo()
					for _, v := range a.GetSortedReceiveBlockByHeight(bestView.GetHeight() + 1) {
						if v.IsValid {
							proposeBlockInfo = v
							break
						}
					}

					var finalityProof = NewFinalityProof()
					var isEnoughLemma2Proof = false
					var failReason = ""
					if proposeBlockInfo.block != nil {
						finalityProof, isEnoughLemma2Proof, failReason = a.ruleDirector.builder.ProposeMessageRule().
							GetValidFinalityProof(proposeBlockInfo.block, a.currentTimeSlot)
						a.logger.Infof("Timeslot %+v, height %+v | Attempt to re-propose block height %+v, hash %+v, produce timeslot %+v,"+
							" is enough finality proof %+v, false reason %+v",
							a.currentTimeSlot, bestView.GetHeight()+1,
							proposeBlockInfo.block.GetHeight(), proposeBlockInfo.block.FullHashString(),
							proposeBlockInfo.block.GetProduceTime(), isEnoughLemma2Proof, failReason)
					} else {
						a.logger.Infof("Timeslot %+v, height %+v | Attempt to create new block",
							a.currentTimeSlot, bestView.GetHeight()+1)
					}

					if createdBlk, err := a.proposeBlock(
						userProposeKey,
						proposerPk,
						proposeBlockInfo,
						committees,
						committeeViewHash,
						isEnoughLemma2Proof,
					); err != nil {
						a.logger.Error(UnExpectedError, errors.New("can't propose block"), err)
					} else {
						if isEnoughLemma2Proof {
							a.logger.Infof("Get Finality Proof | New Block %+v, %+v, Finality Proof %+v",
								createdBlk.GetHeight(), createdBlk.FullHashString(), finalityProof.ReProposeHashSignature)
						}

						env := NewSendProposeBlockEnvironment(
							finalityProof,
							isEnoughLemma2Proof,
							userProposeKey,
							a.node.GetSelfPeerID().String(),
							a.chain.GetBlockConsensusData(),
						)
						bftProposeMessage, err := a.ruleDirector.builder.ProposeMessageRule().CreateProposeBFTMessage(env, createdBlk)
						if err != nil {
							a.logger.Error("Create BFT Propose Message Failed", err)
						} else {
							err = a.sendBFTProposeMsg(bftProposeMessage)
							if err != nil {
								a.logger.Error("Send BFT Propose Message Failed", err)
							}
							a.logger.Infof("[dcs] proposer block %v round %v time slot %v blockTimeSlot %v with hash %v", createdBlk.GetHeight(), createdBlk.GetRound(), a.currentTimeSlot, bestView.CalculateTimeSlot(createdBlk.GetProduceTime()), createdBlk.FullHashString())
						}
					}
				}

				validProposeBlocks := a.getValidProposeBlocks(bestView)
				for _, v := range validProposeBlocks {
					if err := a.validateBlock(bestView.GetHeight(), v); err == nil && v.IsValid && !v.IsVoted {
						err = a.voteValidBlock(v)
						if err != nil {
							a.logger.Debug(err)
						}
					}
				}

				/*
					Check for 2/3 vote to commit
				*/
				for _, v := range a.receiveBlockByHash {
					a.processIfBlockGetEnoughVote(v)
				}
			}
		}
	}()
	return nil
}

func (a *actorV2) isUserKeyProposer(
	bestViewTimeSlot int64,
	proposerPk incognitokey.CommitteePublicKey,
	userKeySet []signatureschemes2.MiningKey) (bool, bool, signatureschemes2.MiningKey) {

	var userProposeKey signatureschemes2.MiningKey
	shouldPropose := false
	shouldListen := true

	for _, userKey := range userKeySet {
		userPk := userKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus)
		if proposerPk.GetMiningKeyBase58(common.BlsConsensus) == userPk {
			shouldListen = false
			// current timeslot is not add to view, and this user is proposer of this timeslot
			if bestViewTimeSlot != a.currentTimeSlot {
				//using block hash as key of best view -> check if this best view we propose or not
				if ok := a.GetCurrentTimeSlotProposeHistory(); !ok {
					shouldPropose = true
					userProposeKey = userKey
				}
			}
		}
	}

	return shouldListen, shouldPropose, userProposeKey
}

func (a *actorV2) getValidatorIndex(committees []incognitokey.CommitteePublicKey, validator string) (int, *incognitokey.CommitteePublicKey) {
	for id, c := range committees {
		if validator == c.GetMiningKeyBase58(common.BlsConsensus) {
			return id, &c
		}
	}
	return -1, nil
}

func (a *actorV2) processIfBlockGetEnoughVote(proposeBlockInfo *ProposeBlockInfo,
) {
	//no vote
	if proposeBlockInfo.HasNewVote == false {
		return
	}

	//no block
	if proposeBlockInfo.block == nil {
		return
	}

	//already in chain
	bestView := a.chain.GetBestView()
	view := a.chain.GetViewByHash(*proposeBlockInfo.block.Hash())
	if view != nil && bestView.GetHash().String() != view.GetHash().String() {
		//e.Logger.Infof("Get View By Hash Fail, %+v, %+v", *v.block.Hash(), v.block.GetHeight())
		return
	}

	//not connected previous block
	view = a.chain.GetViewByHash(proposeBlockInfo.block.GetPrevHash())
	if view == nil {
		//e.Logger.Infof("Get Previous View By Hash Fail, %+v, %+v", v.block.GetPrevHash(), v.block.GetHeight()-1)
		return
	}

	proposeBlockInfo = a.ruleDirector.builder.VoteRule().ValidateVote(proposeBlockInfo)

	if !proposeBlockInfo.IsCommitted {
		a.logger.Infof("Process Block With enough votes, %+v, has %+v, expect > %+v (from total %v). Majority: %+v", proposeBlockInfo.block.FullHashString(), proposeBlockInfo.ValidVotes, 2*len(proposeBlockInfo.SigningCommittees)/3, len(proposeBlockInfo.SigningCommittees), proposeBlockInfo.ValidateFixNodeMajority())
		if proposeBlockInfo.ValidVotes > 2*len(proposeBlockInfo.SigningCommittees)/3 && proposeBlockInfo.ValidateFixNodeMajority() {
			a.logger.Infof("Commit block %v , height: %v", proposeBlockInfo.block.FullHashString(), proposeBlockInfo.block.GetHeight())
			var err error
			if a.chain.IsBeaconChain() {
				err = a.processWithEnoughVotesBeaconChain(proposeBlockInfo)
			} else {
				err = a.processWithEnoughVotesShardChain(proposeBlockInfo)
			}
			if err != nil {
				a.logger.Error(err)
				return
			}
			proposeBlockInfo.IsCommitted = true
		}
	}
}

func (a *actorV2) processWithEnoughVotesBeaconChain(
	v *ProposeBlockInfo,
) error {
	validationData, err := a.createBLSAggregatedSignatures(v.SigningCommittees, v.block.ProposeHash(), v.block.GetValidationField(), v.Votes)
	if err != nil {
		return err
	}
	v.block.(BlockValidation).AddValidationField(validationData)

	if err := a.ruleDirector.builder.InsertBlockRule().InsertBlock(v.block); err != nil {
		return err
	}

	return nil
}

func (a *actorV2) processWithEnoughVotesShardChain(v *ProposeBlockInfo) error {
	validationData, err := a.createBLSAggregatedSignatures(v.SigningCommittees, v.block.ProposeHash(), v.block.GetValidationField(), v.Votes)
	if err != nil {
		return err
	}
	isInsertWithPreviousData := false
	v.block.(BlockValidation).AddValidationField(validationData)
	// validate and add previous block validation data
	previousBlock, _ := a.chain.GetBlockByHash(v.block.GetPrevHash())
	if previousBlock != nil {
		if previousProposeBlockInfo, ok := a.GetReceiveBlockByHash(previousBlock.ProposeHash().String()); ok &&
			previousProposeBlockInfo != nil && previousProposeBlockInfo.block != nil {

			previousProposeBlockInfo = a.ruleDirector.builder.VoteRule().ValidateVote(previousProposeBlockInfo)

			rawPreviousValidationData, err := a.createBLSAggregatedSignatures(
				previousProposeBlockInfo.SigningCommittees,
				previousProposeBlockInfo.block.ProposeHash(),
				previousProposeBlockInfo.block.GetValidationField(),
				previousProposeBlockInfo.Votes)
			if err != nil {
				a.logger.Error("Create BLS Aggregated Signature for previous block propose info, height ", previousProposeBlockInfo.block.GetHeight(), " error", err)
			} else {
				previousProposeBlockInfo.block.(BlockValidation).AddValidationField(rawPreviousValidationData)
				if err := a.ruleDirector.builder.InsertBlockRule().InsertWithPrevValidationData(v.block, rawPreviousValidationData); err != nil {
					return err
				}
				isInsertWithPreviousData = true
				previousValidationData, _ := consensustypes.DecodeValidationData(rawPreviousValidationData)
				a.logger.Infof("Block %+v broadcast with previous block %+v, previous block number of signatures %+v",
					v.block.GetHeight(), previousProposeBlockInfo.block.GetHeight(), len(previousValidationData.ValidatiorsIdx))
			}
		}
	} else {
		a.logger.Info("Cannot find block by hash", v.block.GetPrevHash().String())
	}

	if !isInsertWithPreviousData {
		if err := a.ruleDirector.builder.InsertBlockRule().InsertBlock(v.block); err != nil {
			return err
		}
	}
	loggedCommittee, _ := incognitokey.CommitteeKeyListToString(v.SigningCommittees)
	a.logger.Infof("Successfully Insert Block \n "+
		"ChainID %+v | Height %+v, Hash %+v, Version %+v \n"+
		"Committee %+v", a.chain, v.block.GetHeight(), v.block.FullHashString(), v.block.GetVersion(), loggedCommittee)

	//// @NOTICE: debug mode only, this data should only be used for debugging
	//if v.block.GetVersion() >= types.LEMMA2_VERSION {
	//	if err := a.chain.StoreFinalityProof(v.block, v.FinalityProof, v.ReProposeHashSignature); err != nil {
	//		a.logger.Errorf("Store Finality Proof error %+v", err)
	//	}
	//}
	return nil
}

func (a *actorV2) createBLSAggregatedSignatures(
	committees []incognitokey.CommitteePublicKey,
	blockHash *common.Hash,
	tempValidationData string,
	votes map[string]*BFTVote,
) (string, error) {
	committeeBLSString, err := incognitokey.ExtractPublickeysFromCommitteeKeyList(committees, common.BlsConsensus)
	if err != nil {
		return "", err
	}

	aggSig, brigSigs, validatorIdx, portalSigs, err := CombineVotes(votes, committeeBLSString)
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
	valData.PortalSig = portalSigs
	validationData, _ := consensustypes.EncodeValidationData(*valData)

	//post verify after combine vote
	committeeBLSKeys := []blsmultisig.PublicKey{}
	for _, member := range committees {
		committeeBLSKeys = append(committeeBLSKeys, member.MiningPubKey[consensusName])
	}

	if err := validateBLSSig(blockHash, valData.AggSig, valData.ValidatiorsIdx, committeeBLSKeys); err != nil {
		blsPKList := []blsmultisig.PublicKey{}
		for _, pk := range committees {
			blsK := make([]byte, len(pk.MiningPubKey[common.BlsConsensus]))
			copy(blsK, pk.MiningPubKey[common.BlsConsensus])
			blsPKList = append(blsPKList, blsK)
		}
		for pk, vote := range votes {
			log.Println(common.IndexOfStr(vote.Validator, committeeBLSString), vote.Validator, vote.BLS)
			index := common.IndexOfStr(pk, committeeBLSString)
			if index != -1 {
				err := validateSingleBLSSig(blockHash, vote.BLS, index, blsPKList)
				if err != nil {
					a.logger.Errorf("Can not validate vote from validator %v, pk %v, blkHash from vote %v, blk hash %v ", index, pk, vote.BlockHash, blockHash.String())
					vote.IsValid = -1
				}
			}
		}
		return "", errors.New("ValidateCommitteeSig from combine signature fail")
	}

	return validationData, err
}

// VoteValidBlock this function should be use to vote for valid block only
func (a *actorV2) voteValidBlock(
	proposeBlockInfo *ProposeBlockInfo,
) error {
	//if valid then vote
	committeeBLSString, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(proposeBlockInfo.SigningCommittees, common.BlsConsensus)

	for _, userKey := range proposeBlockInfo.UserKeySet {
		pubKey := userKey.GetPublicKey()
		// TODO: @dung.v review, persist consensus data no longer require this code
		//// When node is not connect to highway (drop connection/startup), propose and vote a block will prevent voting for any other blocks having same height but larger timestamp (rule1)
		//// In case number of validator is 22, we need to make 22 turn to propose the old smallest timestamp block
		//// To prevent this, proposer will not vote unless receiving at least one vote (look at receive vote event)
		if pubKey.GetMiningKeyBase58(a.GetConsensusName()) == proposeBlockInfo.ProposerMiningKeyBase58 {
			continue
		}
		if common.IndexOfStr(pubKey.GetMiningKeyBase58(a.GetConsensusName()), committeeBLSString) != -1 {
			err := a.sendVote(&userKey, proposeBlockInfo.block, proposeBlockInfo.SigningCommittees, a.chain.GetPortalParamsV4(0))
			if err != nil {
				a.logger.Error(err)
				return NewConsensusError(UnExpectedError, err)
			} else {
				if !proposeBlockInfo.IsVoted { //not update database if field is already set
					proposeBlockInfo.IsVoted = true
					if err := a.AddReceiveBlockByHash(proposeBlockInfo.block.ProposeHash().String(), proposeBlockInfo); err != nil {
						return err
					}
				}

			}
		}

	}

	return nil
}

func (a *actorV2) proposeBlock(
	userMiningKey signatureschemes2.MiningKey,
	proposerPk incognitokey.CommitteePublicKey,
	proposeBlockInfo *ProposeBlockInfo,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
	isValidRePropose bool,
) (types.BlockInterface, error) {
	time1 := time.Now()
	b58Str, _ := proposerPk.ToBase58()
	var err error
	block := proposeBlockInfo.block

	if a.chain.IsBeaconChain() {
		block, err = a.proposeBeaconBlock(
			b58Str,
			block,
			committees,
			committeeViewHash,
			isValidRePropose,
		)
	} else {
		block, err = a.proposeShardBlock(
			b58Str,
			block,
			committees,
			committeeViewHash,
			isValidRePropose,
		)
	}

	if err != nil {
		return nil, NewConsensusError(BlockCreationError, err)
	}

	if block != nil {
		a.logger.Infof("create block %v hash %v, propose time %v, produce time %v", block.GetHeight(), block.FullHashString(), block.(types.BlockInterface).GetProposeTime(), block.(types.BlockInterface).GetProduceTime())
	} else {
		a.logger.Infof("create block fail, time: %v", time.Since(time1).Seconds())
		return nil, NewConsensusError(BlockCreationError, errors.New("block is nil"))
	}

	block, err = a.addValidationData(userMiningKey, block)
	if err != nil {
		a.logger.Errorf("Add validation data for new block failed", err)
	}

	return block, nil
}

func (a *actorV2) proposeBeaconBlock(
	b58Str string,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
	isValidRePropose bool,
) (types.BlockInterface, error) {
	var err error
	if block == nil {
		a.logger.Info("CreateNewBlock version", a.blockVersion)
		block, err = a.chain.CreateNewBlock(a.blockVersion, b58Str, 1, a.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	} else {
		a.logger.Infof("CreateNewBlockFromOldBlock, Block Height %+v")
		block, err = a.chain.CreateNewBlockFromOldBlock(block, b58Str, a.currentTime, isValidRePropose)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	}
	return block, err
}

func (a *actorV2) proposeShardBlock(
	b58Str string,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
	isValidRePropose bool,
) (types.BlockInterface, error) {

	var err error
	var newBlock types.BlockInterface
	var committeesFromBeaconHash []incognitokey.CommitteePublicKey
	if block != nil {
		_, committeesFromBeaconHash, err = a.getCommitteeForNewBlock(block)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	}
	isRePropose := true
	if block == nil {
		isRePropose = false
	}
	if block != nil && !reflect.DeepEqual(committeesFromBeaconHash, committees) {
		a.logger.Debugf("SHARD %+v, old block %+v, Attempt to create new block because of committee change",
			a.chainID, block.GetHeight())
		isRePropose = false
	}

	newBlock, err = a.ruleDirector.builder.CreateRule().CreateBlock(
		b58Str,
		block,
		committees,
		committeeViewHash,
		isValidRePropose,
		a.GetConsensusName(),
		a.blockVersion,
		a.currentTime,
		isRePropose,
	)

	return newBlock, err
}

func (a *actorV2) addValidationData(userMiningKey signatureschemes2.MiningKey, block types.BlockInterface) (types.BlockInterface, error) {

	var validationData consensustypes.ValidationData
	portalParam := a.chain.GetPortalParamsV4(0)
	portalSigs, err := portalprocessv4.CheckAndSignPortalUnshieldExternalTx(userMiningKey.PriKey[common.BridgeConsensus], block.GetInstructions(), portalParam)
	if err != nil {
		return block, NewConsensusError(UnExpectedError, err)
	}
	validationData.PortalSig = portalSigs
	validationData.ProducerBLSSig, _ = userMiningKey.BriSignData(block.ProposeHash().GetBytes())
	validationDataString, _ := consensustypes.EncodeValidationData(validationData)
	block.(BlockValidation).AddValidationField(validationDataString)

	return block, nil
}

func (a *actorV2) sendBFTProposeMsg(
	bftPropose *BFTPropose,
) error {

	msg, _ := a.makeBFTProposeMsg(bftPropose, a.chainKey, a.currentTimeSlot)
	go a.ProcessBFTMsg(msg.(*wire.MessageBFT))
	go a.node.PushMessageToChain(msg, a.chain)

	return nil
}

func (a *actorV2) preValidateVote(blockHash []byte, vote *BFTVote, candidate []byte) error {
	data := []byte{}
	data = append(data, blockHash...)
	data = append(data, vote.BLS...)
	data = append(data, vote.BRI...)
	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, vote.Confirmation, candidate)
	return err
}

// getCommitteeForBlock base on the block version to retrieve the right committee list
func (a *actorV2) getCommitteeForNewBlock(
	v types.BlockInterface,
) ([]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error) {
	committees := []incognitokey.CommitteePublicKey{}
	signingCommittees := []incognitokey.CommitteePublicKey{}
	var err error
	proposerIndex := -1
	if a.blockVersion == types.MULTI_VIEW_VERSION || a.chain.IsBeaconChain() {
		committees = a.chain.GetBestView().GetCommittee()
	} else {
		previousView := a.chain.GetViewByHash(v.GetPrevHash())
		committees, err = a.
			committeeChain.
			CommitteesFromViewHashForShard(v.CommitteeFromBlock(), byte(a.chainID))
		if err != nil {
			return signingCommittees, committees, err
		}
		_, proposerIndex = a.chain.GetProposerByTimeSlotFromCommitteeList(
			previousView.CalculateTimeSlot(v.GetProposeTime()),
			committees,
		)
	}

	signingCommittees = a.chain.GetSigningCommittees(
		proposerIndex, committees, v.GetVersion())

	return signingCommittees, committees, err
}

func (a *actorV2) sendVote(
	userKey *signatureschemes2.MiningKey,
	block types.BlockInterface,
	signingCommittees []incognitokey.CommitteePublicKey,
	portalParamV4 portalv4.PortalParams,
) error {

	env := NewVoteMessageEnvironment(
		userKey,
		signingCommittees,
		portalParamV4,
	)
	vote, err := a.ruleDirector.builder.VoteRule().CreateVote(a.chain, env, block)
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}

	msg, err := a.makeBFTVoteMsg(vote, a.chainKey, a.currentTimeSlot, block.GetHeight())
	if err != nil {
		a.logger.Error(err)
		return NewConsensusError(UnExpectedError, err)
	}

	if err := a.AddVoteHistory(block.GetHeight(), block); err != nil {
		a.logger.Errorf("add vote history error %+v", err)
	}

	a.logger.Info(a.chainKey, "sending vote...", block.FullHashString())

	go a.node.PushMessageToChain(msg, a.chain)

	return nil
}

func (a *actorV2) getUserKeySetForSigning(
	signingCommittees []incognitokey.CommitteePublicKey, userKeySet []signatureschemes2.MiningKey,
	proposerIndex int,
) ([]signatureschemes2.MiningKey, int) {
	prepareProposerIndex := -1
	res := []signatureschemes2.MiningKey{}
	if a.chain.IsBeaconChain() {
		res = userKeySet
	} else {
		validCommittees := make(map[string]int)
		for i, v := range signingCommittees {
			key := v.GetMiningKeyBase58(common.BlsConsensus)
			validCommittees[key] = i
		}
		for i, userKey := range userKeySet {
			userPk := userKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus)
			if ci, found := validCommittees[userPk]; found {
				res = append(res, userKey)
				if proposerIndex >= 0 && ci == proposerIndex+1 {
					prepareProposerIndex = i
				}
			}
		}
	}
	return res, prepareProposerIndex
}

func (a *actorV2) getCommitteesAndCommitteeViewHash() (
	[]incognitokey.CommitteePublicKey,
	[]incognitokey.CommitteePublicKey,
	incognitokey.CommitteePublicKey, common.Hash, int, error,
) {
	committeeViewHash := common.Hash{}
	committees := []incognitokey.CommitteePublicKey{}
	var err error
	signingCommittees := []incognitokey.CommitteePublicKey{}
	if a.blockVersion == types.MULTI_VIEW_VERSION || a.chain.IsBeaconChain() {
		committees = a.chain.GetBestView().GetCommittee()
	} else {
		committeeViewHash = *a.committeeChain.FinalView().GetHash()
		committees, err = a.
			committeeChain.
			CommitteesFromViewHashForShard(committeeViewHash, byte(a.chainID))
		if err != nil {
			return []incognitokey.CommitteePublicKey{},
				[]incognitokey.CommitteePublicKey{},
				incognitokey.CommitteePublicKey{},
				committeeViewHash, -1, err
		}
	}

	proposerPk, proposerIndex := a.chain.GetProposerByTimeSlotFromCommitteeList(
		a.currentTimeSlot,
		committees,
	)

	signingCommittees = a.chain.GetSigningCommittees(
		proposerIndex, committees, a.blockVersion)

	return signingCommittees, committees, proposerPk, committeeViewHash, proposerIndex, err
}

func (a *actorV2) handleProposeMsg(proposeMsg BFTPropose) error {

	blockInfo, err := a.chain.UnmarshalBlock(proposeMsg.Block)
	if err != nil || blockInfo == nil {
		return err
	}

	block := blockInfo.(types.BlockInterface)

	blockHash := block.ProposeHash().String()

	_, ok := a.GetReceiveBlockByHash(blockHash)
	if ok {
		return errors.New("Already receive block")
	}

	//update consensus data
	if proposeMsg.BestBlockConsensusData != nil {
		for sid, consensusData := range proposeMsg.BestBlockConsensusData {
			if sid == -1 {
				if a.chain.IsBeaconChain() {
					if err = a.chain.(*blockchain.BeaconChain).VerifyFinalityAndReplaceBlockConsensusData(consensusData); err != nil {
						a.logger.Error(err)
					}
				} else {
					if err = a.chain.(*blockchain.ShardChain).Blockchain.BeaconChain.VerifyFinalityAndReplaceBlockConsensusData(consensusData); err != nil {
						a.logger.Error(err)
					}
				}

			} else if sid >= 0 {
				if a.chain.IsBeaconChain() {
					if err = a.chain.(*blockchain.BeaconChain).Blockchain.ShardChain[sid].VerifyFinalityAndReplaceBlockConsensusData(consensusData); err != nil {
						a.logger.Error(err)
					}
				} else {
					if err = a.chain.(*blockchain.ShardChain).Blockchain.ShardChain[sid].VerifyFinalityAndReplaceBlockConsensusData(consensusData); err != nil {
						a.logger.Error(err)
					}
				}

			}
		}
	}

	previousView := a.chain.GetViewByHash(block.GetPrevHash())
	if previousView == nil {
		a.logger.Infof("Request sync block from node %s from %s to %s", proposeMsg.PeerID, block.GetPrevHash().String(), block.GetPrevHash().Bytes())
		a.node.RequestMissingViewViaStream(proposeMsg.PeerID, [][]byte{block.GetPrevHash().Bytes()}, a.chain.GetShardID(), a.chain.GetChainName())
		return err
	}

	if block.GetHeight() <= a.chain.GetBestViewHeight() {
		return errors.New("Receive block create from old view. Rejected!")
	}

	proposerCommitteePublicKey := incognitokey.CommitteePublicKey{}
	proposerCommitteePublicKey.FromBase58(block.GetProposer())
	proposerMiningKeyBase58 := proposerCommitteePublicKey.GetMiningKeyBase58(a.GetConsensusName())
	signingCommittees, committees, err := a.getCommitteeForNewBlock(block)
	if err != nil {
		return err
	}
	userKeySet, _ := a.getUserKeySetForSigning(signingCommittees, a.userKeySet, -1)

	if len(userKeySet) == 0 && block.GetVersion() < types.INSTANT_FINALITY_VERSION_V2 {
		a.logger.Infof("HandleProposeMsg, Block Hash %+v,  Block Height %+v, round %+v, NOT in round for voting",
			block.FullHashString(), block.GetHeight(), block.GetRound())
		// Log only
		if !a.chain.IsBeaconChain() {
			_, proposerIndex := a.chain.GetProposerByTimeSlotFromCommitteeList(
				previousView.CalculateTimeSlot(block.GetProposeTime()),
				committees,
			)
			subsetID := blockchain.GetSubsetID(proposerIndex)
			userKeySet := a.userKeySet[0].GetPublicKey().GetMiningKeyBase58(a.GetConsensusName())
			userKeySetIndex, userKeySetSubsetID := blockchain.GetSubsetIDByKey(committees, userKeySet, a.GetConsensusName())
			a.logger.Infof("HandleProposeMsg, Block Proposer Index %+v, Block Subset ID %+v, "+
				"UserKeyIndex %+v , UserKeySubsetID %+v",
				proposerIndex, subsetID, userKeySetIndex, userKeySetSubsetID)
		}
	}

	err = a.handleNewProposeMsg(
		proposeMsg,
		block,
		previousView,
		committees,
		signingCommittees,
		userKeySet,
		proposerMiningKeyBase58,
	)

	if err != nil {
		return err
	}

	return nil
}

func (a *actorV2) handleNewProposeMsg(
	proposeMsg BFTPropose,
	block types.BlockInterface,
	previousView multiview.View,
	committees []incognitokey.CommitteePublicKey,
	signingCommittees []incognitokey.CommitteePublicKey,
	userKeySet []signatureschemes2.MiningKey,
	proposerPublicBLSMiningKey string,
) error {

	blockHash := block.ProposeHash().String()
	env := NewProposeMessageEnvironment(
		block,
		previousView.GetBlock(),
		committees,
		signingCommittees,
		userKeySet,
		previousView.GetProposerLength(),
		proposerPublicBLSMiningKey,
	)

	newProposeBlockInfo, err := a.ruleDirector.builder.ProposeMessageRule().HandleBFTProposeMessage(env, &proposeMsg)
	if err != nil {
		a.logger.Errorf("Fail to HandleBFTProposeMessage, block %+v, %+v, "+
			"error %+v", block.GetHeight(), block.ProposeHash().String(), err)
		return err
	}
	if err := a.AddReceiveBlockByHash(blockHash, newProposeBlockInfo); err != nil {
		a.logger.Errorf("add receive block by hash error %+v", err)
	}
	a.logger.Info("Receive block ", block.FullHashString(), "height", block.GetHeight(), ",block timeslot ", a.currentTimeSlot)

	return nil
}

func (a *actorV2) handleVoteMsg(voteMsg BFTVote) error {

	if a.chainID != common.BeaconChainID {
		if err := ByzantineDetectorObject.Validate(
			a.chain.GetBestViewHeight(),
			&voteMsg,
		); err != nil {
			a.logger.Errorf("Found byzantine validator %+v, err %+v", voteMsg.Validator, err)
			return err
		}
	}

	if !a.ruleDirector.builder.HandleVoteMessageRule().IsHandle() {
		//a.logger.Critical("NO COLLECT VOTE")
		return nil
	}

	return a.processVoteMessage(voteMsg)
}

func (a *actorV2) processVoteMessage(voteMsg BFTVote) error {
	voteMsg.IsValid = 0
	if proposeBlockInfo, ok := a.GetReceiveBlockByHash(voteMsg.BlockHash); ok { //if received block is already initiated
		if _, ok := proposeBlockInfo.Votes[voteMsg.Validator]; !ok { // and not receive validatorA vote
			proposeBlockInfo.Votes[voteMsg.Validator] = &voteMsg // store it
			vid, v := a.getValidatorIndex(proposeBlockInfo.SigningCommittees, voteMsg.Validator)
			if v != nil {
				vbase58, _ := v.ToBase58()
				a.logger.Infof("%v Receive vote (%d) for block %s from validator %d %v", a.chainKey, len(a.receiveBlockByHash[voteMsg.BlockHash].Votes), voteMsg.BlockHash, vid, vbase58)
			} else {
				a.logger.Infof("%v Receive vote (%d) for block %v from unknown validator %v", a.chainKey, len(a.receiveBlockByHash[voteMsg.BlockHash].Votes), voteMsg.BlockHash, voteMsg.Validator)
			}
			proposeBlockInfo.HasNewVote = true
		}

		if !proposeBlockInfo.ProposerSendVote {
			for _, userKey := range a.userKeySet {
				pubKey := userKey.GetPublicKey()
				if proposeBlockInfo.block != nil && pubKey.GetMiningKeyBase58(a.GetConsensusName()) == proposeBlockInfo.ProposerMiningKeyBase58 { // if this node is proposer and not sending vote
					var err error
					if err = a.validateBlock(a.chain.GetBestView().GetHeight(), proposeBlockInfo); err == nil && proposeBlockInfo.IsValid {
						bestViewHeight := a.chain.GetBestView().GetHeight()
						if proposeBlockInfo.block.GetHeight() == bestViewHeight+1 { // and if the propose block is still connected to bestview
							err := a.sendVote(&userKey, proposeBlockInfo.block, proposeBlockInfo.SigningCommittees, a.chain.GetPortalParamsV4(0)) // => send vote
							if err != nil {
								a.logger.Error(err)
							} else {
								proposeBlockInfo.ProposerSendVote = true
								if err := a.AddReceiveBlockByHash(proposeBlockInfo.block.ProposeHash().String(), proposeBlockInfo); err != nil {
									return err
								}
							}
						}
					} else {
						a.logger.Debug(err)
					}
				}
			}
		}

	}

	// record new votes for restore
	if err := AddVoteByBlockHashToDB(voteMsg.BlockHash, voteMsg); err != nil {
		a.logger.Errorf("add receive block by hash error %+v", err)
	}
	return nil
}

func (a *actorV2) handleCleanMem() {

	for h := range a.voteHistory {
		if h <= a.chain.GetFinalView().GetHeight() {
			if err := a.CleanVoteHistory(h); err != nil {
				a.logger.Errorf("clean vote history error %+v", err)
			}
		}
	}

	for h, proposeBlk := range a.receiveBlockByHash {
		if time.Now().Sub(proposeBlk.ReceiveTime) > time.Minute && (proposeBlk.block == nil || proposeBlk.block.GetHeight() <= a.chain.GetFinalView().GetHeight()) {
			if err := a.CleanReceiveBlockByHash(h); err != nil {
				a.logger.Errorf("clean receive block by hash error %+v", err)
			}
		}
	}

	for timeSlot := range a.proposeHistory {
		if timeSlot < a.currentTimeSlot {
			if err := a.CleanProposeHistory(timeSlot); err != nil {
				a.logger.Errorf("clean propose history %+v", err)
			}
		}
	}

	a.ruleDirector.builder.ProposeMessageRule().HandleCleanMem(a.chain.GetFinalView().GetHeight())
	ByzantineDetectorObject.UpdateState(a.chain.GetFinalView().GetHeight(), a.chain.GetBestView().CalculateTimeSlot(a.chain.GetFinalView().GetBlock().GetProposeTime()))

}

// getValidProposeBlocks validate received proposed block and return valid proposed block
// Special case: in case block is already inserted, try to send vote (avoid slashing)
// 1. by pass nil block
// 2. just validate recently
// 3. not in current time slot
// 4. not connect to best view
func (a *actorV2) getValidProposeBlocks(bestView multiview.View) []*ProposeBlockInfo {
	//Check for valid block to vote
	bestViewHeight := bestView.GetHeight()
	bestViewProposeHash := *bestView.GetBlock().ProposeHash()
	validProposeBlock, tryVoteInsertedBlocks, invalidProposeBlocks := a.ruleDirector.builder.ValidatorRule().FilterValidProposeBlockInfo(
		bestViewProposeHash,
		bestViewHeight,
		a.chain.GetFinalView().GetHeight(),
		a.currentTimeSlot,
		a.receiveBlockByHash,
	)

	for _, invalidProposeBlock := range invalidProposeBlocks {
		if err := a.CleanReceiveBlockByHash(invalidProposeBlock); err != nil {
			a.logger.Errorf("clean receive block by hash error %+v", err)
		}
	}

	// HACK CASE, block already insert but still vote
	// if block is inserted => no need to validate again
	for _, tryVoteInsertedBlock := range tryVoteInsertedBlocks {
		if !tryVoteInsertedBlock.IsValid {
			err := a.validateBlock(tryVoteInsertedBlock.block.GetHeight()-1, tryVoteInsertedBlock)
			if err != nil {
				a.logger.Errorf("Block %+v try vote inserted block but invalid", tryVoteInsertedBlock.block.FullHashString())
				continue
			}
		}
		if tryVoteInsertedBlock.IsValid {
			a.voteValidBlock(tryVoteInsertedBlock)
		}
	}

	return validProposeBlock
}

func (a *actorV2) validateBlock(bestViewHeight uint64, proposeBlockInfo *ProposeBlockInfo) error {

	//not validate if we do it recently
	if time.Since(proposeBlockInfo.LastValidateTime).Seconds() < 1 {
		return nil
	}

	if proposeBlockInfo.IsValid {
		return nil
	}

	lastVotedBlock, isVoted := a.GetVoteHistory(bestViewHeight + 1)
	blockProduceTimeSlot := a.chain.GetBestView().CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime())

	isValid, err := a.ruleDirector.builder.ValidatorRule().ValidateBlock(lastVotedBlock, isVoted, proposeBlockInfo)
	if err != nil {
		return err
	}

	proposeBlockInfo.LastValidateTime = time.Now()

	if !isValid {
		a.logger.Debugf("can't vote for this block %v height %v timeslot %v",
			proposeBlockInfo.block.ProposeHash().String(), proposeBlockInfo.block.GetHeight(), blockProduceTimeSlot)
		return errors.New("can't vote for this block")
	}

	proposeBlockInfo.IsValid = true

	return nil
}

func (a *actorV2) validatePreSignBlock(proposeBlockInfo *ProposeBlockInfo) error {

	//not connected
	view := a.chain.GetViewByHash(proposeBlockInfo.block.GetPrevHash())
	if view == nil {
		previousView := a.chain.GetViewByHash(proposeBlockInfo.block.GetPrevHash())
		blkCreateTimeSlot := previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime())
		a.logger.Infof("previous view for this block %v height %v timeslot %v is null",
			proposeBlockInfo.block.ProposeHash().String(), proposeBlockInfo.block.GetHeight(), blkCreateTimeSlot)
		return errors.New("View not connect")
	}

	_, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	a.logger.Infof("validate block: %+v \n", proposeBlockInfo.block.ProposeHash().String())
	if err := a.chain.ValidatePreSignBlock(proposeBlockInfo.block, proposeBlockInfo.SigningCommittees, proposeBlockInfo.Committees); err != nil {
		a.logger.Error(err)
		return err
	}

	return nil
}

func CombineVotes(votes map[string]*BFTVote, committees []string) (aggSig []byte, brigSigs [][]byte, validatorIdx []int, portalSigs []*portalprocessv4.PortalSig, err error) {
	var blsSigList [][]byte
	for validator, vote := range votes {
		if vote.IsValid == 1 {
			index := common.IndexOfStr(validator, committees)
			if index != -1 {
				validatorIdx = append(validatorIdx, index)
			}
		}
	}

	if len(validatorIdx) == 0 {
		return nil, nil, nil, nil, NewConsensusError(CombineSignatureError, errors.New("len(validatorIdx) == 0"))
	}

	sort.Ints(validatorIdx)
	for _, idx := range validatorIdx {
		blsSigList = append(blsSigList, votes[committees[idx]].BLS)
		brigSigs = append(brigSigs, votes[committees[idx]].BRI)
		portalSigs = append(portalSigs, votes[committees[idx]].PortalSigs...)
	}

	aggSig, err = blsmultisig.Combine(blsSigList)
	if err != nil {
		return nil, nil, nil, nil, NewConsensusError(CombineSignatureError, err)
	}

	return
}

func (a *actorV2) makeBFTProposeMsg(
	proposeCtn *BFTPropose,
	chainKey string,
	ts int64,
) (wire.Message, error) {
	proposeCtnBytes, err := json.Marshal(proposeCtn)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = proposeCtnBytes
	msg.(*wire.MessageBFT).Type = MSG_PROPOSE
	msg.(*wire.MessageBFT).TimeSlot = ts
	msg.(*wire.MessageBFT).Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	msg.(*wire.MessageBFT).PeerID = proposeCtn.PeerID
	return msg, nil
}

func (a *actorV2) makeBFTVoteMsg(vote *BFTVote, chainKey string, ts int64, height uint64) (wire.Message, error) {
	voteCtnBytes, err := json.Marshal(vote)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = voteCtnBytes
	msg.(*wire.MessageBFT).Type = MSG_VOTE
	msg.(*wire.MessageBFT).TimeSlot = ts
	msg.(*wire.MessageBFT).Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	return msg, nil
}

func (a *actorV2) makeBFTRequestBlk(request BFTRequestBlock, peerID string, chainKey string) (wire.Message, error) {
	requestCtnBytes, err := json.Marshal(request)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = requestCtnBytes
	msg.(*wire.MessageBFT).Type = MsgRequestBlk
	return msg, nil
}

func ExtractPortalV4ValidationData(block types.BlockInterface) ([]*portalprocessv4.PortalSig, error) {
	valData, err := consensustypes.DecodeValidationData(block.GetValidationField())
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	return valData.PortalSig, nil
}
