package blsbft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/bridgesig"
	"reflect"
	"sort"
	"time"

	lru "github.com/hashicorp/golang-lru"
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
	proposeHistory  *lru.Cache

	receiveBlockByHeight map[uint64][]*ProposeBlockInfo  //blockHeight -> blockInfo
	receiveBlockByHash   map[string]*ProposeBlockInfo    //blockHash -> blockInfo
	voteHistory          map[uint64]types.BlockInterface // bestview height (previsous height )-> block
	//votedTimeslot        map[int64]bool
	blockVersion int
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
	a.receiveBlockByHash = make(map[string]*ProposeBlockInfo)
	a.receiveBlockByHeight = make(map[uint64][]*ProposeBlockInfo)
	a.voteHistory = make(map[uint64]types.BlockInterface)
	//a.votedTimeslot = make(map[int64]bool)
	a.committeeChain = committeeChain
	a.blockVersion = blockVersion
	a.proposeHistory, err = lru.New(1000)
	if err != nil {
		panic(err) //must not error
	}
	return a
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

func (a *actorV2) Stop() error {
	if a.isStarted {
		a.logger.Info("stop bls-bftv3 consensus for chain", a.chainKey)
	}
	a.isStarted = false
	return nil
}

func (a *actorV2) Destroy() {
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
			fmt.Println(err)
			return
		}
		msgPropose.PeerID = msgBFT.PeerID
		a.proposeMessageCh <- msgPropose
	case MSG_VOTE:
		var msgVote BFTVote
		err := json.Unmarshal(msgBFT.Content, &msgVote)
		if err != nil {
			fmt.Println(err)
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
		a.logger.Info("start bls-bftv3 consensus for chain", a.chainKey)
	}
	a.isStarted = true
	return nil
}

func (a *actorV2) run() error {
	go func() {
		//init view maps
		ticker := time.Tick(200 * time.Millisecond)
		cleanMemTicker := time.Tick(5 * time.Minute)
		a.logger.Infof("init bls-bft-v2 consensus for chain %+v", a.chainKey)

		for { //actor loop
			if !a.isStarted { //sleep if this process is not start
				time.Sleep(time.Second)
				continue
			}

			select {
			case <-a.destroyCh:
				a.logger.Infof("exit bls-bft-%+v consensus for chain %+v", a.blockVersion, a.chainKey)
				close(a.destroyCh)
				return
			case proposeMsg := <-a.proposeMessageCh:
				err := a.handleProposeMsg(proposeMsg)
				if err != nil {
					a.logger.Debug(err)
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
				currentTimeSlot := common.CalculateTimeSlot(a.currentTime)

				newTimeSlot := false
				if a.currentTimeSlot != currentTimeSlot {
					newTimeSlot = true
				}

				a.currentTimeSlot = currentTimeSlot
				bestView := a.chain.GetBestView()

				//set round for monitor
				round := a.currentTimeSlot - common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime())
				monitor.SetGlobalParam("RoundKey", fmt.Sprintf("%d_%d", bestView.GetHeight(), round))

				signingCommittees, committees, proposerPk, committeeViewHash, err := a.getCommitteesAndCommitteeViewHash()
				if err != nil {
					a.logger.Info(err)
					continue
				}

				userKeySet := a.getUserKeySetForSigning(signingCommittees, a.userKeySet)
				shouldListen, shouldPropose, userProposeKey := a.isUserKeyProposer(
					common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()),
					proposerPk,
					userKeySet,
				)

				if newTimeSlot { //for logging
					a.logger.Info("")
					a.logger.Info("======================================================")
					a.logger.Info("")
					if shouldListen {
						a.logger.Infof("%v TS: %v, LISTEN BLOCK %v, Round %v", a.chainKey, common.CalculateTimeSlot(a.currentTime), bestView.GetHeight()+1, round)
					}
					if shouldPropose {
						a.logger.Infof("%v TS: %v, PROPOSE BLOCK %v, Round %v", a.chainKey, common.CalculateTimeSlot(a.currentTime), bestView.GetHeight()+1, round)
					}
				}

				if shouldPropose {
					a.proposeHistory.Add(fmt.Sprintf("%d", a.currentTimeSlot), 1)
					// Proposer Rule: check propose block connected to bestview (longest chain rule 1)
					// and re-propose valid block with smallest timestamp (including already propose in the past) (rule 2)
					sort.Slice(a.receiveBlockByHeight[bestView.GetHeight()+1], func(i, j int) bool {
						return a.receiveBlockByHeight[bestView.GetHeight()+1][i].block.GetProduceTime() < a.receiveBlockByHeight[bestView.GetHeight()+1][j].block.GetProduceTime()
					})

					var proposeBlock types.BlockInterface = nil
					for _, v := range a.receiveBlockByHeight[bestView.GetHeight()+1] {
						if v.isValid {
							proposeBlock = v.block
							break
						}
					}

					if createdBlk, err := a.proposeBlock(userProposeKey, proposerPk, proposeBlock, committees, committeeViewHash); err != nil {
						a.logger.Critical(UnExpectedError, errors.New("can't propose block"))
						a.logger.Critical(err)
					} else {
						a.logger.Infof("[dcs] proposer block %v round %v time slot %v blockTimeSlot %v with hash %v", createdBlk.GetHeight(), createdBlk.GetRound(), a.currentTimeSlot, common.CalculateTimeSlot(createdBlk.GetProduceTime()), createdBlk.Hash().String())
					}
				}

				validProposeBlocks := a.getValidProposeBlocks(bestView)
				for _, v := range validProposeBlocks {
					if err := a.validateBlock(bestView, v); err == nil {
						err = a.voteValidBlock(v)
						if err != nil {
							a.logger.Debug(err)
						}
					}
				}

				/*
					Check for 2/3 vote to commit
				*/
				for k, v := range a.receiveBlockByHash {
					a.processIfBlockGetEnoughVote(k, v)
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
				if _, ok := a.proposeHistory.Get(fmt.Sprintf("%d", a.currentTimeSlot)); !ok {
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

func (a *actorV2) processIfBlockGetEnoughVote(
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
	a.logger.Infof("Process Block With enough votes, %+v, %+v", *v.block.Hash(), v.block.GetHeight())
	//already in chain
	bestView := a.chain.GetBestView()
	view := a.chain.GetViewByHash(*v.block.Hash())
	if view != nil && bestView.GetHash().String() != view.GetHash().String() {
		//e.Logger.Infof("Get View By Hash Fail, %+v, %+v", *v.block.Hash(), v.block.GetHeight())
		return
	}

	//not connected previous block
	view = a.chain.GetViewByHash(v.block.GetPrevHash())
	if view == nil {
		//e.Logger.Infof("Get Previous View By Hash Fail, %+v, %+v", v.block.GetPrevHash(), v.block.GetHeight()-1)
		return
	}

	v = a.validateVotes(v)

	if !v.isCommitted {
		if v.validVotes > 2*len(v.signingCommittees)/3 {
			a.logger.Infof("Commit block %v , height: %v", blockHash, v.block.GetHeight())
			var err error
			if a.chain.IsBeaconChain() {
				err = a.processWithEnoughVotesBeaconChain(v)

			} else {
				err = a.processWithEnoughVotesShardChain(v)
			}
			if err != nil {
				a.logger.Error(err)
				return
			}
			v.isCommitted = true
		}
	}
}

func (a *actorV2) validateVotes(v *ProposeBlockInfo) *ProposeBlockInfo {
	validVote := 0
	errVote := 0

	committees := make(map[string]int)
	if len(v.votes) != 0 {
		for i, v := range v.signingCommittees {
			committees[v.GetMiningKeyBase58(common.BlsConsensus)] = i
		}
	}

	for id, vote := range v.votes {
		dsaKey := []byte{}
		if vote.IsValid == 0 {
			if value, ok := committees[vote.Validator]; ok {
				dsaKey = v.signingCommittees[value].MiningPubKey[common.BridgeConsensus]
			} else {
				a.logger.Error("Receive vote from nonCommittee member")
				continue
			}
			if len(dsaKey) == 0 {
				a.logger.Error("canot find dsa key")
				continue
			}

			err := vote.validateVoteOwner(dsaKey)
			if err != nil {
				a.logger.Error(dsaKey)
				a.logger.Error(err)
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

	a.logger.Info("Number of Valid Vote", validVote, "| Number Of Error Vote", errVote)
	v.hasNewVote = false
	for key, value := range v.votes {
		if value.IsValid == -1 {
			delete(v.votes, key)
		}
	}

	v.addBlockInfo(
		v.block,
		v.committees,
		v.signingCommittees,
		v.userKeySet,
		v.proposerMiningKeyBase58,
		validVote,
		errVote,
	)

	return v
}
func (a *actorV2) processWithEnoughVotesBeaconChain(
	v *ProposeBlockInfo,
) error {
	validationData, err := a.createBLSAggregatedSignatures(v.signingCommittees, v.block.GetValidationField(), v.votes)
	if err != nil {
		a.logger.Error(err)
		return err
	}
	v.block.(blockValidation).AddValidationField(validationData)

	if err := a.chain.InsertAndBroadcastBlock(v.block); err != nil {
		return err
	}

	delete(a.receiveBlockByHash, v.block.GetPrevHash().String())

	return nil
}

func (a *actorV2) processWithEnoughVotesShardChain(v *ProposeBlockInfo) error {

	validationData, err := a.createBLSAggregatedSignatures(v.signingCommittees, v.block.GetValidationField(), v.votes)
	if err != nil {
		a.logger.Error(err)
		return err
	}
	v.block.(blockValidation).AddValidationField(validationData)

	// validate and previous block
	if previousProposeBlockInfo, ok := a.receiveBlockByHash[v.block.GetPrevHash().String()]; ok &&
		previousProposeBlockInfo != nil && previousProposeBlockInfo.block != nil {

		previousProposeBlockInfo = a.validateVotes(previousProposeBlockInfo)

		previousValidationData, err := a.createBLSAggregatedSignatures(
			previousProposeBlockInfo.signingCommittees,
			previousProposeBlockInfo.block.GetValidationField(),
			previousProposeBlockInfo.votes)
		if err != nil {
			a.logger.Error(err)
			return err
		}

		previousProposeBlockInfo.block.(blockValidation).AddValidationField(previousValidationData)
		if err := a.chain.InsertAndBroadcastBlockWithPrevValidationData(v.block, previousValidationData); err != nil {
			return err
		}

		delete(a.receiveBlockByHash, previousProposeBlockInfo.block.GetPrevHash().String())
	}

	if err := a.chain.InsertAndBroadcastBlock(v.block); err != nil {
		return err
	}

	loggedCommittee, _ := incognitokey.CommitteeKeyListToString(v.signingCommittees)
	a.logger.Infof("Successfully Insert Block \n "+
		"ChainID %+v | Height %+v, Hash %+v, Version %+v \n"+
		"Committee %+v", a.chain, v.block.GetHeight(), *v.block.Hash(), v.block.GetVersion(), loggedCommittee)

	return nil
}

func (a *actorV2) createBLSAggregatedSignatures(
	committees []incognitokey.CommitteePublicKey,
	tempValidationData string,
	votes map[string]*BFTVote,
) (string, error) {
	committeeBLSString, err := incognitokey.ExtractPublickeysFromCommitteeKeyList(committees, common.BlsConsensus)
	if err != nil {
		return "", err
	}

	aggSig, brigSigs, validatorIdx, err := a.combineVotes(votes, committeeBLSString)
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

//voteValidBlock this function should be use to vote for valid block only
func (a *actorV2) voteValidBlock(
	proposeBlockInfo *ProposeBlockInfo,
) error {
	//if valid then vote
	committeeBLSString, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(proposeBlockInfo.signingCommittees, common.BlsConsensus)
	for _, userKey := range proposeBlockInfo.userKeySet {
		pubKey := userKey.GetPublicKey()
		//TODO: @hung will this trick has bad effect on other vote cases
		// When node is not connect to highway (drop connection/startup), propose and vote a block will prevent voting for any other blocks having same height but larger timestamp (rule1)
		// In case number of validator is 22, we need to make 22 turn to propose the old smallest timestamp block
		// To prevent this, proposer will not vote unless receiving at least one vote (look at receive vote event)
		if pubKey.GetMiningKeyBase58(a.GetConsensusName()) == proposeBlockInfo.proposerMiningKeyBase58 {
			continue
		}
		if common.IndexOfStr(pubKey.GetMiningKeyBase58(a.GetConsensusName()), committeeBLSString) != -1 {
			err := a.sendVote(&userKey, proposeBlockInfo.block, proposeBlockInfo.signingCommittees)
			if err != nil {
				a.logger.Error(err)
				return NewConsensusError(UnExpectedError, err)
			} else {
				proposeBlockInfo.isVoted = true
			}
		}
	}

	return nil
}

func createVote(
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
	vote.Validator = userBLSPk
	vote.PrevBlockHash = block.GetPrevHash().String()
	err = vote.signVote(userKey)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	return vote, nil
}

func (a *actorV2) proposeBlock(
	userMiningKey signatureschemes2.MiningKey,
	proposerPk incognitokey.CommitteePublicKey,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
) (types.BlockInterface, error) {
	time1 := time.Now()
	b58Str, _ := proposerPk.ToBase58()
	var err error

	if a.chain.IsBeaconChain() {
		block, err = a.proposeBeaconBlock(
			b58Str,
			block,
			committees,
			committeeViewHash,
		)
	} else {
		block, err = a.proposeShardBlock(
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
		a.logger.Infof("create block %v hash %v, propose time %v, produce time %v", block.GetHeight(), block.Hash().String(), block.(types.BlockInterface).GetProposeTime(), block.(types.BlockInterface).GetProduceTime())
	} else {
		a.logger.Infof("create block fail, time: %v", time.Since(time1).Seconds())
		return nil, NewConsensusError(BlockCreationError, errors.New("block is nil"))
	}

	var validationData consensustypes.ValidationData
	validationData.ProducerBLSSig, _ = userMiningKey.BriSignData(block.Hash().GetBytes())
	validationDataString, _ := consensustypes.EncodeValidationData(validationData)
	block.(blockValidation).AddValidationField(validationDataString)
	blockData, _ := json.Marshal(block)

	var proposeCtn = new(BFTPropose)
	proposeCtn.Block = blockData
	proposeCtn.PeerID = a.node.GetSelfPeerID().String()
	msg, _ := a.makeBFTProposeMsg(proposeCtn, a.chainKey, a.currentTimeSlot, block.GetHeight())
	go a.ProcessBFTMsg(msg.(*wire.MessageBFT))
	go a.node.PushMessageToChain(msg, a.chain)

	return block, nil
}

func (a *actorV2) proposeBeaconBlock(
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
		a.logger.Info("CreateNewBlock")
		block, err = a.chain.CreateNewBlock(a.blockVersion, b58Str, 1, a.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	} else {
		a.logger.Infof("CreateNewBlockFromOldBlock, Block Height %+v")
		block, err = a.chain.CreateNewBlockFromOldBlock(block, b58Str, a.currentTime, committees, committeeViewHash)
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
) (types.BlockInterface, error) {
	var err error
	var newBlock types.BlockInterface
	var committeesFromBeaconHash []incognitokey.CommitteePublicKey
	if block != nil {
		_, committeesFromBeaconHash, err = a.getCommitteeForBlock(block)
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
		a.logger.Info("CreateNewBlock")
		newBlock, err = a.chain.CreateNewBlock(a.blockVersion, b58Str, 1, a.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	} else {
		a.logger.Infof("CreateNewBlockFromOldBlock, Block Height %+v hash %+v", block.GetHeight(), block.Hash().String())
		newBlock, err = a.chain.CreateNewBlockFromOldBlock(block, b58Str, a.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	}
	return newBlock, err
}

func (a *actorV2) preValidateVote(blockHash []byte, vote *BFTVote, candidate []byte) error {
	data := []byte{}
	data = append(data, blockHash...)
	data = append(data, vote.Bls...)
	data = append(data, vote.Bri...)
	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, vote.Confirmation, candidate)
	return err
}

// getCommitteeForBlock base on the block version to retrieve the right committee list
func (a *actorV2) getCommitteeForBlock(
	v types.BlockInterface,
) ([]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error) {
	committees := []incognitokey.CommitteePublicKey{}
	signingCommittees := []incognitokey.CommitteePublicKey{}
	var err error
	proposerIndex := -1
	if a.blockVersion == types.MULTI_VIEW_VERSION || a.chain.IsBeaconChain() {
		committees = a.chain.GetBestView().GetCommittee()
	} else {
		committees, err = a.
			committeeChain.
			CommitteesFromViewHashForShard(v.CommitteeFromBlock(), byte(a.chainID))
		if err != nil {
			return signingCommittees, committees, err
		}
		_, proposerIndex, err = a.chain.GetProposerByTimeSlotFromCommitteeList(
			common.CalculateTimeSlot(v.GetProposeTime()),
			committees,
		)
		if err != nil {
			return signingCommittees, committees, err
		}
	}

	signingCommittees = a.chain.GetSigningCommittees(
		proposerIndex, committees, v.GetVersion())

	return signingCommittees, committees, err
}

func (a *actorV2) sendVote(userKey *signatureschemes2.MiningKey, block types.BlockInterface, signingCommittees []incognitokey.CommitteePublicKey) error {

	Vote, err := createVote(userKey, block, signingCommittees)
	if err != nil {
		a.logger.Error(err)
		return NewConsensusError(UnExpectedError, err)
	}

	msg, err := a.makeBFTVoteMsg(Vote, a.chainKey, a.currentTimeSlot, block.GetHeight())
	if err != nil {
		a.logger.Error(err)
		return NewConsensusError(UnExpectedError, err)
	}

	a.voteHistory[block.GetHeight()] = block
	//a.votedTimeslot[common.CalculateTimeSlot(block.GetProposeTime())] = true

	a.logger.Info(a.chainKey, "sending vote...")

	go a.node.PushMessageToChain(msg, a.chain)

	return nil
}

func (a *actorV2) getUserKeySetForSigning(
	signingCommittees []incognitokey.CommitteePublicKey, userKeySet []signatureschemes2.MiningKey,
) []signatureschemes2.MiningKey {
	res := []signatureschemes2.MiningKey{}
	if a.chain.IsBeaconChain() {
		res = userKeySet
	} else {
		validCommittees := make(map[string]bool)
		for _, v := range signingCommittees {
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

func (a *actorV2) getCommitteesAndCommitteeViewHash() (
	[]incognitokey.CommitteePublicKey,
	[]incognitokey.CommitteePublicKey,
	incognitokey.CommitteePublicKey, common.Hash, error,
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
				committeeViewHash, err
		}
	}

	proposerPk, proposerIndex, err := a.chain.GetProposerByTimeSlotFromCommitteeList(
		a.currentTimeSlot,
		committees,
	)
	if err != nil {
		return []incognitokey.CommitteePublicKey{},
			[]incognitokey.CommitteePublicKey{},
			incognitokey.CommitteePublicKey{},
			committeeViewHash, err
	}

	signingCommittees = a.chain.GetSigningCommittees(
		proposerIndex, committees, a.blockVersion)

	return signingCommittees, committees, proposerPk, committeeViewHash, err
}

func (a *actorV2) handleProposeMsg(proposeMsg BFTPropose) error {
	blockIntf, err := a.chain.UnmarshalBlock(proposeMsg.Block)
	if err != nil || blockIntf == nil {
		a.logger.Debug(err)
		return err
	}
	block := blockIntf.(types.BlockInterface)
	blkHash := block.Hash().String()

	blkCPk := incognitokey.CommitteePublicKey{}
	blkCPk.FromBase58(block.GetProducer())
	proposerMiningKeyBase58 := blkCPk.GetMiningKeyBase58(a.GetConsensusName())

	signingCommittees, committees, err := a.getCommitteeForBlock(block)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	userKeySet := a.getUserKeySetForSigning(signingCommittees, a.userKeySet)
	if len(userKeySet) == 0 {
		a.logger.Infof("HandleProposeMsg, Block Hash %+v, Block Height %+v, round %+v, NOT in round for voting",
			*block.Hash(), block.GetHeight(), block.GetRound())
		// Log only
		if !a.chain.IsBeaconChain() {
			_, proposerIndex, _ := a.chain.GetProposerByTimeSlotFromCommitteeList(
				common.CalculateTimeSlot(block.GetProposeTime()),
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

	if v, ok := a.receiveBlockByHash[blkHash]; !ok {
		proposeBlockInfo := newProposeBlockForProposeMsg(
			block, committees, signingCommittees, userKeySet, proposerMiningKeyBase58)
		a.receiveBlockByHash[blkHash] = proposeBlockInfo
		a.logger.Info("Receive block ", block.Hash().String(), "height", block.GetHeight(), ",block timeslot ", common.CalculateTimeSlot(block.GetProposeTime()))
		a.receiveBlockByHeight[block.GetHeight()] = append(a.receiveBlockByHeight[block.GetHeight()], a.receiveBlockByHash[blkHash])
	} else {
		a.receiveBlockByHash[blkHash].addBlockInfo(
			block, committees, signingCommittees, userKeySet, proposerMiningKeyBase58, v.validVotes, v.errVotes)
	}

	if block.GetHeight() <= a.chain.GetBestViewHeight() {
		a.logger.Debug("Receive block create from old view. Rejected!")
		return errors.New("Receive block create from old view. Rejected!")
	}

	proposeView := a.chain.GetViewByHash(block.GetPrevHash())
	if proposeView == nil {
		a.logger.Infof("Request sync block from node %s from %s to %s", proposeMsg.PeerID, block.GetPrevHash().String(), block.GetPrevHash().Bytes())
		a.node.RequestMissingViewViaStream(proposeMsg.PeerID, [][]byte{block.GetPrevHash().Bytes()}, a.chain.GetShardID(), a.chain.GetChainName())
	}
	return nil
}

func (a *actorV2) handleVoteMsg(voteMsg BFTVote) error {
	voteMsg.IsValid = 0
	if b, ok := a.receiveBlockByHash[voteMsg.BlockHash]; ok { //if received block is already initiated
		if _, ok := b.votes[voteMsg.Validator]; !ok { // and not receive validatorA vote
			//TODO: @hung only store vote from known validator?
			b.votes[voteMsg.Validator] = &voteMsg // store it
			vid, v := a.getValidatorIndex(b.signingCommittees, voteMsg.Validator)
			if v != nil {
				vbase58, _ := v.ToBase58()
				a.logger.Infof("%v Receive vote (%d) for block %s from validator %d %v", a.chainKey, len(a.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, vid, vbase58)
			} else {
				a.logger.Infof("%v Receive vote (%d) for block %v from unknown validator %v", a.chainKey, len(a.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, voteMsg.Validator)
			}
			b.hasNewVote = true
		}

		if !b.proposerSendVote {
			for _, userKey := range a.userKeySet {
				pubKey := userKey.GetPublicKey()
				if b.block != nil && pubKey.GetMiningKeyBase58(a.GetConsensusName()) == b.proposerMiningKeyBase58 { // if this node is proposer and not sending vote
					var err error
					if err = a.validateBlock(a.chain.GetBestView(), b); err != nil {
						err = a.voteValidBlock(b)
						if err != nil {
							a.logger.Debug(err)
						}
					} else {
						a.logger.Debug(err)
					}
					//TODO: @hung only send vote if userKey in proposeBlockInfo.UserKeySet list?
					if err == nil {
						bestViewHeight := a.chain.GetBestView().GetHeight()
						if b.block.GetHeight() == bestViewHeight+1 { // and if the propose block is still connected to bestview
							err := a.sendVote(&userKey, b.block, b.signingCommittees) // => send vote
							if err != nil {
								a.logger.Error(err)
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
		a.receiveBlockByHash[voteMsg.BlockHash] = newBlockInfoForVoteMsg()
		a.receiveBlockByHash[voteMsg.BlockHash].votes[voteMsg.Validator] = &voteMsg
		a.logger.Infof("Chain %v, Receive vote (%d) for block %v from validator %v", a.chainKey, len(a.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, voteMsg.Validator)
	}
	return nil
}

func (a *actorV2) handleCleanMem() {

	for h, _ := range a.receiveBlockByHeight {
		if h <= a.chain.GetFinalView().GetHeight() {
			delete(a.receiveBlockByHeight, h)
		}
	}

	for h, _ := range a.voteHistory {
		if h <= a.chain.GetFinalView().GetHeight() {
			delete(a.voteHistory, h)
		}
	}

	for h, proposeBlk := range a.receiveBlockByHash {
		if time.Now().Sub(proposeBlk.receiveTime) > time.Minute {
			//delete(a.votedTimeslot, proposeBlk.block.GetProposeTime())
			delete(a.receiveBlockByHash, h)
		}
	}

}

// getValidProposeBlocks validate received proposed block and return valid proposed block
// Special case: in case block is already inserted, try to send vote (avoid slashing)
// 1. by pass nil block
// 2. just validate recently
// 3. not in current time slot
// 4. not connect to best view
func (a *actorV2) getValidProposeBlocks(bestView multiview.View) []*ProposeBlockInfo {
	//Check for valid block to vote
	validProposeBlock := []*ProposeBlockInfo{}
	//get all block that has height = bestview height  + 1(rule 2 & rule 3) (
	bestViewHeight := bestView.GetHeight()
	for h, proposeBlockInfo := range a.receiveBlockByHash {
		if proposeBlockInfo.block == nil {
			continue
		}

		//// check if this time slot has been voted
		//if a.votedTimeslot[common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime())] {
		//	continue
		//}

		//special case: if we insert block too quick, before voting
		//=> vote for this block (within TS,but block is inserted into bestview)
		//this special case by pass validate with consensus rules
		if proposeBlockInfo.block.GetHeight() == bestViewHeight && !proposeBlockInfo.isVoted {
			//already validate and vote for this proposed block
			if !proposeBlockInfo.isValid {
				if err := a.validatePreSignBlock(proposeBlockInfo); err != nil {
					continue
				}
			}
			a.voteValidBlock(proposeBlockInfo)
			continue
		}

		//not validate if we do it recently
		if time.Since(proposeBlockInfo.lastValidateTime).Seconds() < 1 {
			continue
		}

		// check if propose block in within TS
		if common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) != a.currentTimeSlot {
			continue
		}

		//if the block height is not next height or current height
		if proposeBlockInfo.block.GetHeight() != bestViewHeight+1 {
			continue
		}

		// check if producer time > proposer time
		if common.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime()) > a.currentTimeSlot {
			continue
		}

		if proposeBlockInfo.block.GetHeight() < a.chain.GetFinalView().GetHeight() {
			//delete(a.votedTimeslot, proposeBlockInfo.block.GetProposeTime())
			delete(a.receiveBlockByHash, h)
		}

		validProposeBlock = append(validProposeBlock, proposeBlockInfo)
	}
	//rule 1: get history of vote for this height, vote if (round is lower than the vote before) or (round is equal but new proposer) or (there is no vote for this height yet)
	sort.Slice(validProposeBlock, func(i, j int) bool {
		return validProposeBlock[i].block.GetProduceTime() < validProposeBlock[j].block.GetProduceTime()
	})
	return validProposeBlock
}

func (a *actorV2) validateBlock(bestView multiview.View, proposeBlockInfo *ProposeBlockInfo) error {

	proposeBlockInfo.lastValidateTime = time.Now()

	bestViewHeight := bestView.GetHeight()
	blkCreateTimeSlot := common.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime())

	shouldVote := a.validateConsensusRules(bestViewHeight, proposeBlockInfo)

	if !shouldVote {
		a.logger.Debugf("Can't vote for this block %v height %v timeslot %v",
			proposeBlockInfo.block.Hash().String(), proposeBlockInfo.block.GetHeight(), blkCreateTimeSlot)
		return errors.New("Can't vote for this block")
	}

	if proposeBlockInfo.isVoted {
		return nil
	}

	//already validate and vote for this proposed block
	if !proposeBlockInfo.isValid {
		if err := a.validatePreSignBlock(proposeBlockInfo); err != nil {
			return err
		}
	}

	proposeBlockInfo.isValid = true

	return nil
}

//validateConsensusRules validate block, block is valid when one of these conditions hold
// 1. block connect to best view (== bestViewHeight + 1) and first time receive this height
// 2. blockHeight = lastVoteBlockHeight && blockCreationTime < lastVoteBlockCreationTime
// 3. blockCreationTime = lastVoteBlockCreationTime && blockProposeTime > lastVoteBlockProposeTime
// 4. block has new committees (assign from beacon) than lastVoteBlock
func (a *actorV2) validateConsensusRules(bestViewHeight uint64, proposeBlockInfo *ProposeBlockInfo) bool {
	blkCreateTimeSlot := common.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime())
	if lastVotedBlk, ok := a.voteHistory[bestViewHeight+1]; ok {
		if blkCreateTimeSlot < common.CalculateTimeSlot(lastVotedBlk.GetProduceTime()) { //blkCreateTimeSlot is smaller than voted block => vote for this blk
			return true
		} else if blkCreateTimeSlot == common.CalculateTimeSlot(lastVotedBlk.GetProduceTime()) && common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) > common.CalculateTimeSlot(lastVotedBlk.GetProposeTime()) { //blk is old block (same round), but new proposer(larger timeslot) => vote again
			return true
		} else if proposeBlockInfo.block.CommitteeFromBlock().String() != lastVotedBlk.CommitteeFromBlock().String() { //blkCreateTimeSlot is larger or equal than voted block
			return true
		} // if not swap committees => do nothing
	} else { //there is no vote for this height yet
		return true
	}

	return false
}

func (a *actorV2) validatePreSignBlock(proposeBlockInfo *ProposeBlockInfo) error {

	blkCreateTimeSlot := common.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime())

	//not connected
	view := a.chain.GetViewByHash(proposeBlockInfo.block.GetPrevHash())
	if view == nil {
		a.logger.Infof("previous view for this block %v height %v timeslot %v is null",
			proposeBlockInfo.block.Hash().String(), proposeBlockInfo.block.GetHeight(), blkCreateTimeSlot)
		return errors.New("View not connect")
	}

	_, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	a.logger.Infof("validate block: %+v \n", proposeBlockInfo.block.Hash().String())
	if err := a.chain.ValidatePreSignBlock(proposeBlockInfo.block, proposeBlockInfo.signingCommittees, proposeBlockInfo.committees); err != nil {
		a.logger.Error(err)
		return err
	}

	return nil
}

func (a *actorV2) BlockVersion() int {
	return a.blockVersion
}

func (a *actorV2) combineVotes(votes map[string]*BFTVote, committees []string) (aggSig []byte, brigSigs [][]byte, validatorIdx []int, err error) {
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
		return nil, nil, nil, NewConsensusError(CombineSignatureError, errors.New("len(validatorIdx) == 0"))
	}

	sort.Ints(validatorIdx)
	for _, idx := range validatorIdx {
		blsSigList = append(blsSigList, votes[committees[idx]].Bls)
		brigSigs = append(brigSigs, votes[committees[idx]].Bri)
	}

	aggSig, err = blsmultisig.Combine(blsSigList)
	if err != nil {
		return nil, nil, nil, NewConsensusError(CombineSignatureError, err)
	}

	return
}

func (a *actorV2) makeBFTProposeMsg(proposeCtn *BFTPropose, chainKey string, ts int64, height uint64) (wire.Message, error) {
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
