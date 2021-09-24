package blsbft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
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
	// previous block hash -> a map of next block block time slot -> corresponding re-propose hash signature

	nextBlockFinalityProof map[string]map[int64]string

	proposeRule        IProposeRule
	consensusValidator IConsensusValidator
	VoteRule           IVoteRule
	blockVersion       int
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
	a.nextBlockFinalityProof = make(map[string]map[int64]string)
	//a.bodyHashes = make(map[uint64]map[string]bool)
	//a.votedTimeslot = make(map[int64]bool)
	a.committeeChain = committeeChain
	a.blockVersion = blockVersion
	a.proposeHistory, err = lru.New(1000)
	a.proposeRule = NewProposeRuleLemma2(
		logger,
		make(map[string]map[int64]string),
		chain,
	)
	a.consensusValidator = NewConsensusValidator(
		logger,
		chain,
	)
	a.VoteRule = NewVoteRule(
		logger,
	)
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

func (a *actorV2) run() error {
	go func() {
		//init view maps
		ticker := time.Tick(200 * time.Millisecond)
		cleanMemTicker := time.Tick(5 * time.Minute)
		a.logger.Infof("init bls-bft-%+v consensus for chain %+v", a.blockVersion, a.chainKey)

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

					var proposeBlockInfo = NewProposeBlockInfo()
					for _, v := range a.receiveBlockByHeight[bestView.GetHeight()+1] {
						if v.isValid {
							proposeBlockInfo = v
							break
						}
					}

					var finalityProof = NewFinalityProof()
					var isValidRePropose bool = false
					if proposeBlockInfo.block != nil {
						finalityProof, isValidRePropose = a.proposeRule.GetValidFinalityProof(proposeBlockInfo.block, a.currentTimeSlot)
					}
					if createdBlk, err := a.proposeBlock(
						userProposeKey,
						proposerPk,
						proposeBlockInfo,
						committees,
						committeeViewHash,
						isValidRePropose,
					); err != nil {
						a.logger.Error(UnExpectedError, errors.New("can't propose block"), err)
					} else {
						if isValidRePropose {
							a.logger.Infof("Get Finality Proof | New Block %+v, %+v, Finality Proof %+v",
								createdBlk.GetHeight(), createdBlk.Hash().String(), finalityProof.ReProposeHashSignature)
						}

						env := NewSendProposeBlockEnvironment(
							finalityProof,
							isValidRePropose,
							userProposeKey,
							a.node.GetSelfPeerID().String(),
						)
						bftProposeMessage, err := a.proposeRule.CreateProposeBFTMessage(env, createdBlk)
						if err != nil {
							a.logger.Error("Create BFT Propose Message Failed", err)
						} else {
							err = a.sendBFTProposeMsg(bftProposeMessage)
							if err != nil {
								a.logger.Error("Send BFT Propose Message Failed", err)
							}
							a.logger.Infof("[dcs] proposer block %v round %v time slot %v blockTimeSlot %v with hash %v", createdBlk.GetHeight(), createdBlk.GetRound(), a.currentTimeSlot, common.CalculateTimeSlot(createdBlk.GetProduceTime()), createdBlk.Hash().String())
						}
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
	blockHash string, proposeBlockInfo *ProposeBlockInfo,
) {
	//no vote
	if proposeBlockInfo.hasNewVote == false {
		return
	}

	//no block
	if proposeBlockInfo.block == nil {
		return
	}
	a.logger.Infof("Process Block With enough votes, %+v, %+v", *proposeBlockInfo.block.Hash(), proposeBlockInfo.block.GetHeight())
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

	proposeBlockInfo = a.VoteRule.ValidateVote(proposeBlockInfo)

	if !proposeBlockInfo.isCommitted {
		if proposeBlockInfo.validVotes > 2*len(proposeBlockInfo.signingCommittees)/3 {
			a.logger.Infof("Commit block %v , height: %v", blockHash, proposeBlockInfo.block.GetHeight())
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
			proposeBlockInfo.isCommitted = true
		}
	}
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

		previousProposeBlockInfo = a.VoteRule.ValidateVote(previousProposeBlockInfo)

		rawPreviousValidationData, err := a.createBLSAggregatedSignatures(
			previousProposeBlockInfo.signingCommittees,
			previousProposeBlockInfo.block.GetValidationField(),
			previousProposeBlockInfo.votes)
		if err != nil {
			a.logger.Error(err)
			return err
		}

		previousProposeBlockInfo.block.(blockValidation).AddValidationField(rawPreviousValidationData)
		if err := a.chain.InsertAndBroadcastBlockWithPrevValidationData(v.block, rawPreviousValidationData); err != nil {
			return err
		}

		previousValidationData, _ := consensustypes.DecodeValidationData(rawPreviousValidationData)
		a.logger.Infof("Block %+v broadcast with previous block %+v, previous block number of signatures %+v",
			v.block.GetHeight(), previousProposeBlockInfo.block.GetHeight(), len(previousValidationData.ValidatiorsIdx))

		delete(a.receiveBlockByHash, previousProposeBlockInfo.block.GetPrevHash().String())
	} else {

		if err := a.chain.InsertAndBroadcastBlock(v.block); err != nil {
			return err
		}
	}
	loggedCommittee, _ := incognitokey.CommitteeKeyListToString(v.signingCommittees)
	a.logger.Infof("Successfully Insert Block \n "+
		"ChainID %+v | Height %+v, Hash %+v, Version %+v \n"+
		"Committee %+v", a.chain, v.block.GetHeight(), *v.block.Hash(), v.block.GetVersion(), loggedCommittee)

	// @NOTICE: debug mode only, this data should only be used for debugging
	if err := a.chain.StoreFinalityProof(v.block, v.finalityProof, v.reProposeHashSignature); err != nil {
		a.logger.Errorf("Store Finality Proof error %+v", err)
	}
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

	aggSig, brigSigs, validatorIdx, portalSigs, err := a.combineVotes(votes, committeeBLSString)
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
	return validationData, err
}

//VoteValidBlock this function should be use to vote for valid block only
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
			err := a.sendVote(&userKey, proposeBlockInfo.block, proposeBlockInfo.signingCommittees, a.chain.GetPortalParamsV4(0))
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
		a.logger.Infof("create block %v hash %v, propose time %v, produce time %v", block.GetHeight(), block.Hash().String(), block.(types.BlockInterface).GetProposeTime(), block.(types.BlockInterface).GetProduceTime())
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

		//debug only
		proposerCommitteePK, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{b58Str})
		proposerKeySet := proposerCommitteePK[0].GetMiningKeyBase58(a.GetConsensusName())
		proposerKeySetIndex, proposerKeySetSubsetID := blockchain.GetSubsetIDByKey(committees, proposerKeySet, a.GetConsensusName())
		newBlock, err = a.chain.CreateNewBlock(a.blockVersion, b58Str, 1, a.currentTime, committees, committeeViewHash)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
		a.logger.Infof("CreateNewBlock, Block Height %+v, Block Hash %+v | "+
			"Producer Index %+v, Producer SubsetID %+v", newBlock.GetHeight(), newBlock.Hash().String(),
			proposerKeySetIndex, proposerKeySetSubsetID)
	} else {
		//debug only
		proposerCommitteePK, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{b58Str})
		proposerKeySet := proposerCommitteePK[0].GetMiningKeyBase58(a.GetConsensusName())
		proposerKeySetIndex, proposerKeySetSubsetID := blockchain.GetSubsetIDByKey(committees, proposerKeySet, a.GetConsensusName())
		producerCommitteePK, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{block.GetProducer()})
		producerKeySet := producerCommitteePK[0].GetMiningKeyBase58(a.GetConsensusName())
		producerKeySetIndex, producerKeySetSubsetID := blockchain.GetSubsetIDByKey(committees, producerKeySet, a.GetConsensusName())

		a.logger.Infof("CreateNewBlockFromOldBlock, Block Height %+v hash %+v | "+
			"Producer Index %+v, Producer SubsetID %+v | "+
			"Proposer Index %+v, Proposer SubsetID %+v ", block.GetHeight(), block.Hash().String(),
			producerKeySetIndex, producerKeySetSubsetID, proposerKeySetIndex, proposerKeySetSubsetID)
		newBlock, err = a.chain.CreateNewBlockFromOldBlock(block, b58Str, a.currentTime, isValidRePropose)
		if err != nil {
			return nil, NewConsensusError(BlockCreationError, err)
		}
	}

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
	validationData.ProducerBLSSig, _ = userMiningKey.BriSignData(block.Hash().GetBytes())
	validationDataString, _ := consensustypes.EncodeValidationData(validationData)
	block.(blockValidation).AddValidationField(validationDataString)

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
		_, proposerIndex = a.chain.GetProposerByTimeSlotFromCommitteeList(
			common.CalculateTimeSlot(v.GetProposeTime()),
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
	Vote, err := a.VoteRule.CreateVote(env, block)
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

	proposerPk, proposerIndex := a.chain.GetProposerByTimeSlotFromCommitteeList(
		a.currentTimeSlot,
		committees,
	)

	signingCommittees = a.chain.GetSigningCommittees(
		proposerIndex, committees, a.blockVersion)

	return signingCommittees, committees, proposerPk, committeeViewHash, err
}

func (a *actorV2) handleProposeMsg(proposeMsg BFTPropose) error {

	blockInfo, err := a.chain.UnmarshalBlock(proposeMsg.Block)
	if err != nil || blockInfo == nil {
		a.logger.Debug(err)
		return err
	}

	block := blockInfo.(types.BlockInterface)

	blockHash := block.Hash().String()
	producerCommitteePublicKey := incognitokey.CommitteePublicKey{}
	producerCommitteePublicKey.FromBase58(block.GetProducer())
	producerMiningKeyBase58 := producerCommitteePublicKey.GetMiningKeyBase58(a.GetConsensusName())
	signingCommittees, committees, err := a.getCommitteeForBlock(block)
	if err != nil {
		a.logger.Error(err)
		return err
	}
	userKeySet := a.getUserKeySetForSigning(signingCommittees, a.userKeySet)
	previousBlock, err := a.chain.GetBlockByHash(block.GetPrevHash())
	if err != nil {
		a.logger.Error(err)
		return err
	}
	if len(userKeySet) == 0 {
		a.logger.Infof("HandleProposeMsg, Block Hash %+v, Block Height %+v, round %+v, NOT in round for voting",
			*block.Hash(), block.GetHeight(), block.GetRound())
		// Log only
		if !a.chain.IsBeaconChain() {
			_, proposerIndex := a.chain.GetProposerByTimeSlotFromCommitteeList(
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

	if v, ok := a.receiveBlockByHash[blockHash]; !ok {
		err := a.handleNewProposeMsg(
			proposeMsg,
			block,
			previousBlock,
			committees,
			signingCommittees,
			userKeySet,
			producerMiningKeyBase58,
		)
		if err != nil {
			a.logger.Error(err)
			return err
		}
	} else {
		a.receiveBlockByHash[blockHash].addBlockInfo(
			block, committees, signingCommittees, userKeySet, v.validVotes, v.errVotes)
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

func (a *actorV2) handleNewProposeMsg(
	proposeMsg BFTPropose,
	block types.BlockInterface,
	previousBlock types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	signingCommittees []incognitokey.CommitteePublicKey,
	userKeySet []signatureschemes2.MiningKey,
	producerPublicBLSMiningKey string,
) error {

	blockHash := block.Hash().String()
	env := NewProposeMessageEnvironment(
		block,
		previousBlock,
		committees,
		signingCommittees,
		userKeySet,
		producerPublicBLSMiningKey,
	)
	proposeBlockInfo, err := a.proposeRule.HandleBFTProposeMessage(env, &proposeMsg)
	if err != nil {
		a.logger.Errorf("Fail to apply lemma 2, block %+v, %+v, "+
			"error %+v", block.GetHeight(), block.Hash().String(), err)
		return err
	}

	a.receiveBlockByHash[blockHash] = proposeBlockInfo
	a.logger.Info("Receive block ", block.Hash().String(), "height", block.GetHeight(), ",block timeslot ", common.CalculateTimeSlot(block.GetProposeTime()))
	a.receiveBlockByHeight[block.GetHeight()] = append(a.receiveBlockByHeight[block.GetHeight()], a.receiveBlockByHash[blockHash])

	return nil
}

func (a *actorV2) handleVoteMsg(voteMsg BFTVote) error {

	voteMsg.IsValid = 0
	if b, ok := a.receiveBlockByHash[voteMsg.BlockHash]; ok { //if received block is already initiated
		if _, ok := b.votes[voteMsg.Validator]; !ok { // and not receive validatorA vote
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
					if err == nil {
						bestViewHeight := a.chain.GetBestView().GetHeight()
						if b.block.GetHeight() == bestViewHeight+1 { // and if the propose block is still connected to bestview
							err := a.sendVote(&userKey, b.block, b.signingCommittees, a.chain.GetPortalParamsV4(0)) // => send vote
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
			//delete(a.bodyHashes, h)
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

	a.proposeRule.HandleCleanMem(a.chain.GetFinalView().GetHeight())

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

	validProposeBlock, tryVoteInsertedBlocks, invalidProposeBlocks := a.consensusValidator.FilterValidProposeBlockInfo(
		bestViewHeight,
		a.chain.GetFinalView().GetHeight(),
		a.currentTimeSlot,
		a.receiveBlockByHash,
	)

	for _, invalidProposeBlock := range invalidProposeBlocks {
		delete(a.receiveBlockByHash, invalidProposeBlock)
	}

	for _, tryVoteInsertedBlock := range tryVoteInsertedBlocks {
		if !tryVoteInsertedBlock.isValid {
			if err := a.validatePreSignBlock(tryVoteInsertedBlock); err != nil {
				continue
			}
		}
		a.voteValidBlock(tryVoteInsertedBlock)
	}

	return validProposeBlock
}

func (a *actorV2) validateBlock(bestView multiview.View, proposeBlockInfo *ProposeBlockInfo) error {

	bestViewHeight := bestView.GetHeight()
	lastVotedBlock, isVoted := a.voteHistory[bestViewHeight+1]
	blockProduceTimeSlot := common.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime())

	isValid, err := a.consensusValidator.ValidateBlock(lastVotedBlock, isVoted, proposeBlockInfo)
	if err != nil {
		return err
	}

	if !isValid {
		a.logger.Debugf("can't vote for this block %v height %v timeslot %v",
			proposeBlockInfo.block.Hash().String(), proposeBlockInfo.block.GetHeight(), blockProduceTimeSlot)
		return errors.New("can't vote for this block")
	}

	proposeBlockInfo.isValid = true

	return nil
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

func (a *actorV2) combineVotes(votes map[string]*BFTVote, committees []string) (aggSig []byte, brigSigs [][]byte, validatorIdx []int, portalSigs []*portalprocessv4.PortalSig, err error) {
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

//func (a *actorV2) validateVotes(v *ProposeBlockInfo) *ProposeBlockInfo {
//	validVote := 0
//	errVote := 0
//
//	committees := make(map[string]int)
//	if len(v.votes) != 0 {
//		for i, v := range v.signingCommittees {
//			committees[v.GetMiningKeyBase58(common.BlsConsensus)] = i
//		}
//	}
//
//	for id, vote := range v.votes {
//		dsaKey := []byte{}
//		if vote.IsValid == 0 {
//			if value, ok := committees[vote.Validator]; ok {
//				dsaKey = v.signingCommittees[value].MiningPubKey[common.BridgeConsensus]
//			} else {
//				a.logger.Error("Receive vote from nonCommittee member")
//				continue
//			}
//			if len(dsaKey) == 0 {
//				a.logger.Error("canot find dsa key")
//				continue
//			}
//
//			err := vote.validateVoteOwner(dsaKey)
//			if err != nil {
//				a.logger.Error(dsaKey)
//				a.logger.Error(err)
//				v.votes[id].IsValid = -1
//				errVote++
//			} else {
//				v.votes[id].IsValid = 1
//				validVote++
//			}
//		} else {
//			validVote++
//		}
//	}
//
//	a.logger.Info("Number of Valid Vote", validVote, "| Number Of Error Vote", errVote)
//	v.hasNewVote = false
//	for key, value := range v.votes {
//		if value.IsValid == -1 {
//			delete(v.votes, key)
//		}
//	}
//
//	v.addBlockInfo(
//		v.block,
//		v.committees,
//		v.signingCommittees,
//		v.userKeySet,
//		validVote,
//		errVote,
//	)
//
//	return v
//}
//
//func (a *actorV2) getValidProposeBlocks(bestView multiview.View) []*ProposeBlockInfo {
//	//Check for valid block to vote
//	validProposeBlock := []*ProposeBlockInfo{}
//	//get all block that has height = bestview height  + 1(rule 2 & rule 3) (
//	bestViewHeight := bestView.GetHeight()
//	for h, proposeBlockInfo := range a.receiveBlockByHash {
//		if proposeBlockInfo.block == nil {
//			continue
//		}
//
//		//// check if this time slot has been voted
//		//if a.votedTimeslot[common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime())] {
//		//	continue
//		//}
//
//		//special case: if we insert block too quick, before voting
//		//=> vote for this block (within TS,but block is inserted into bestview)
//		//this special case by pass validate with consensus rules
//		if proposeBlockInfo.block.GetHeight() == bestViewHeight && !proposeBlockInfo.isVoted {
//			//already validate and vote for this proposed block
//			if !proposeBlockInfo.isValid {
//				if err := a.validatePreSignBlock(proposeBlockInfo); err != nil {
//					continue
//				}
//			}
//			a.voteValidBlock(proposeBlockInfo)
//			continue
//		}
//
//		//not validate if we do it recently
//		if time.Since(proposeBlockInfo.lastValidateTime).Seconds() < 1 {
//			continue
//		}
//
//		// check if propose block in within TS
//		if common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) != a.currentTimeSlot {
//			continue
//		}
//
//		//if the block height is not next height or current height
//		if proposeBlockInfo.block.GetHeight() != bestViewHeight+1 {
//			continue
//		}
//
//		// check if producer time > proposer time
//		if common.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime()) > a.currentTimeSlot {
//			continue
//		}
//
//		// lemma 2
//		if proposeBlockInfo.isValidLemma2Proof {
//			if proposeBlockInfo.block.GetFinalityHeight() != proposeBlockInfo.block.GetHeight()-1 {
//				a.logger.Errorf("Block %+v %+v, is valid for lemma 2, expect finality height %+v, got %+v",
//					proposeBlockInfo.block.GetHeight(), proposeBlockInfo.block.Hash().String(),
//					proposeBlockInfo.block.GetHeight(), proposeBlockInfo.block.GetFinalityHeight())
//				continue
//			}
//		}
//		if !proposeBlockInfo.isValidLemma2Proof {
//			if proposeBlockInfo.block.GetFinalityHeight() != 0 {
//				a.logger.Errorf("Block %+v %+v, root hash %+v, previous block hash %+v, is invalid for lemma 2, expect finality height %+v, got %+v",
//					proposeBlockInfo.block.GetHeight(), proposeBlockInfo.block.Hash().String(), proposeBlockInfo.block.GetAggregateRootHash(), proposeBlockInfo.block.GetPrevHash().String(),
//					0, proposeBlockInfo.block.GetFinalityHeight())
//				continue
//			}
//		}
//
//		//TODO @hung continue?
//		if proposeBlockInfo.block.GetHeight() < a.chain.GetFinalView().GetHeight() {
//			//delete(a.votedTimeslot, proposeBlockInfo.block.GetProposeTime())
//			delete(a.receiveBlockByHash, h)
//		}
//
//		validProposeBlock = append(validProposeBlock, proposeBlockInfo)
//	}
//	//rule 1: get history of vote for this height, vote if (round is lower than the vote before) or (round is equal but new proposer) or (there is no vote for this height yet)
//	sort.Slice(validProposeBlock, func(i, j int) bool {
//		return validProposeBlock[i].block.GetProduceTime() < validProposeBlock[j].block.GetProduceTime()
//	})
//	return validProposeBlock
//}
//
//func (a *actorV2) validateBlock(bestView multiview.View, proposeBlockInfo *ProposeBlockInfo) error {
//
//	proposeBlockInfo.lastValidateTime = time.Now()
//
//	bestViewHeight := bestView.GetHeight()
//	blkCreateTimeSlot := common.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime())
//
//	shouldVote := a.validateConsensusRules(bestViewHeight, proposeBlockInfo)
//
//	if !shouldVote {
//		a.logger.Debugf("Can't vote for this block %v height %v timeslot %v",
//			proposeBlockInfo.block.Hash().String(), proposeBlockInfo.block.GetHeight(), blkCreateTimeSlot)
//		return errors.New("Can't vote for this block")
//	}
//
//	if proposeBlockInfo.isVoted {
//		return nil
//	}
//
//	//already validate and vote for this proposed block
//	if !proposeBlockInfo.isValid {
//		a.logger.Infof("validate block: %+v \n", proposeBlockInfo.block.Hash().String())
//		if err := a.chain.ValidatePreSignBlock(proposeBlockInfo.block, proposeBlockInfo.signingCommittees, proposeBlockInfo.committees); err != nil {
//			a.logger.Error(err)
//			return err
//		}
//	}
//
//	proposeBlockInfo.isValid = true
//
//	return nil
//}

////validateConsensusRules validate block, block is valid when one of these conditions hold
//// 1. block connect to best view (== bestViewHeight + 1) and first time receive this height
//// 2. blockHeight = lastVoteBlockHeight && blockCreationTime < lastVoteBlockCreationTime
//// 3. blockCreationTime = lastVoteBlockCreationTime && blockProposeTime > lastVoteBlockProposeTime
//// 4. block has new committees (assign from beacon) than lastVoteBlock
//func (a *actorV2) validateConsensusRules(bestViewHeight uint64, proposeBlockInfo *ProposeBlockInfo) bool {
//	blkCreateTimeSlot := common.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime())
//	if lastVotedBlk, ok := a.voteHistory[bestViewHeight+1]; ok {
//		if blkCreateTimeSlot < common.CalculateTimeSlot(lastVotedBlk.GetProduceTime()) { //blkCreateTimeSlot is smaller than voted block => vote for this blk
//			return true
//		} else if blkCreateTimeSlot == common.CalculateTimeSlot(lastVotedBlk.GetProduceTime()) && common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) > common.CalculateTimeSlot(lastVotedBlk.GetProposeTime()) { //blk is old block (same round), but new proposer(larger timeslot) => vote again
//			return true
//		} else if proposeBlockInfo.block.CommitteeFromBlock().String() != lastVotedBlk.CommitteeFromBlock().String() { //blkCreateTimeSlot is larger or equal than voted block
//			return true
//		} // if not swap committees => do nothing
//	} else { //there is no vote for this height yet
//		return true
//	}
//
//	return false
//}

//func (a *actorV2) sendBFTProposeMsg(
//	finalityProof *FinalityProof,
//	reProposeHashSignature string,
//	isValidRePropose bool,
//	block types.BlockInterface,
//) error {
//
//	blockData, _ := json.Marshal(block)
//	var bftPropose = new(BFTPropose)
//	if block.GetVersion() >= types.BLOCK_PRODUCINGV3_VERSION {
//		if isValidRePropose {
//			bftPropose.FinalityProof = *finalityProof
//		} else {
//			bftPropose.FinalityProof = *NewFinalityProof()
//		}
//		bftPropose.ReProposeHashSignature = reProposeHashSignature
//	}
//	bftPropose.Block = blockData
//	bftPropose.PeerID = a.node.GetSelfPeerID().String()
//	msg, _ := a.makeBFTProposeMsg(bftPropose, a.chainKey, a.currentTimeSlot)
//	go a.ProcessBFTMsg(msg.(*wire.MessageBFT))
//	go a.node.PushMessageToChain(msg, a.chain)
//
//	return nil
//}
//func (a *actorV2) getValidFinalityProof(block types.BlockInterface) (*FinalityProof, bool) {
//
//	if block == nil {
//		return NewFinalityProof(), false
//	}
//
//	finalityData, ok := a.nextBlockFinalityProof[block.GetPrevHash().String()]
//	if !ok {
//		return NewFinalityProof(), false
//	}
//
//	finalityProof := NewFinalityProof()
//
//	producerTime := block.GetProduceTime()
//	producerTimeTimeSlot := common.CalculateTimeSlot(producerTime)
//	currentTimeSlot := a.currentTimeSlot
//
//	if currentTimeSlot-producerTimeTimeSlot > MAX_FINALITY_PROOF {
//		return finalityProof, false
//	}
//
//	for i := producerTimeTimeSlot; i < currentTimeSlot; i++ {
//		reProposeHashSignature, ok := finalityData[i]
//		if !ok {
//			return NewFinalityProof(), false
//		}
//		finalityProof.AddProof(reProposeHashSignature)
//	}
//
//	return finalityProof, true
//}
//
//// block can apply lemma 2 when
//// Can be applied: first block of next height (compare to bestview) or re-propose from first block of next height
//// Can't be applied: not first block of next height and not re-proposed from first block of next height
//func (a *actorV2) handleNewProposeMsgLemma2(
//	proposeMsg BFTPropose,
//	previousBlock types.BlockInterface,
//	block types.BlockInterface,
//	committees []incognitokey.CommitteePublicKey,
//	signingCommittees []incognitokey.CommitteePublicKey,
//	userKeySet []signatureschemes2.MiningKey,
//	producerPublicBLSMiningKey string,
//) (*ProposeBlockInfo, error) {
//
//	isValidLemma2 := false
//	var err error
//	var isReProposeFirstBlockNextHeight = false
//	var isFirstBlockNextHeight = false
//
//	isFirstBlockNextHeight = a.isFirstBlockNextHeight(previousBlock, block)
//	if isFirstBlockNextHeight {
//		err := a.verifyLemma2FirstBlockNextHeight(proposeMsg, block)
//		if err != nil {
//			return nil, err
//		}
//		isValidLemma2 = true
//	} else {
//		isReProposeFirstBlockNextHeight = a.isReProposeFromFirstBlockNextHeight(previousBlock, block, committees)
//		if isReProposeFirstBlockNextHeight {
//			isValidLemma2, err = a.verifyLemma2ReProposeBlockNextHeight(proposeMsg, block, committees)
//			if err != nil {
//				return nil, err
//			}
//		}
//	}
//
//	proposeBlockInfo := newProposeBlockForProposeMsgLemma2(
//		&proposeMsg,
//		block,
//		committees,
//		signingCommittees,
//		userKeySet,
//		producerPublicBLSMiningKey,
//		isValidLemma2,
//	)
//
//	if !isValidLemma2 {
//		a.logger.Infof("Receive Invalid Block for lemma 2, block %+v, %+v",
//			block.GetHeight(), block.Hash().String())
//	}
//
//	if isValidLemma2 {
//		if err := a.addFinalityProof(block, proposeMsg.ReProposeHashSignature, proposeMsg.FinalityProof); err != nil {
//			return nil, err
//		}
//		a.logger.Infof("Receive Valid Block for lemma 2, block %+v, %+v",
//			block.GetHeight(), block.Hash().String())
//	}
//
//	return proposeBlockInfo, nil
//}
//
//// isFirstBlockNextHeight verify firstBlockNextHeight
//// producer timeslot is proposer timeslot
//// producer is proposer
//// producer timeslot = previous proposer timeslot + 1
//func (a *actorV2) isFirstBlockNextHeight(
//	previousBlock types.BlockInterface,
//	block types.BlockInterface,
//) bool {
//
//	if block.GetProposeTime() != block.GetProduceTime() {
//		return false
//	}
//
//	if block.GetProposer() != block.GetProducer() {
//		return false
//	}
//
//	previousProposerTimeSlot := common.CalculateTimeSlot(previousBlock.GetProposeTime())
//	producerTimeSlot := common.CalculateTimeSlot(block.GetProduceTime())
//
//	if producerTimeSlot != previousProposerTimeSlot+1 {
//		return false
//	}
//
//	return true
//}
//
//// isReProposeFromFirstBlockNextHeight verify a block is re-propose from first block next height
//// producer timeslot is first block next height
//// proposer timeslot > producer timeslot
//// proposer is correct
//func (a *actorV2) isReProposeFromFirstBlockNextHeight(
//	previousBlock types.BlockInterface,
//	block types.BlockInterface,
//	committees []incognitokey.CommitteePublicKey,
//) bool {
//
//	previousProposerTimeSlot := common.CalculateTimeSlot(previousBlock.GetProposeTime())
//	producerTimeSlot := common.CalculateTimeSlot(block.GetProduceTime())
//	proposerTimeSlot := common.CalculateTimeSlot(block.GetProposeTime())
//
//	if producerTimeSlot != previousProposerTimeSlot+1 {
//		return false
//	}
//
//	if proposerTimeSlot <= producerTimeSlot {
//		return false
//	}
//
//	wantProposer, _ := a.chain.GetProposerByTimeSlotFromCommitteeList(proposerTimeSlot, committees)
//	wantProposerBase58, _ := wantProposer.ToBase58()
//	if block.GetProposer() != wantProposerBase58 {
//		return false
//	}
//
//	return true
//}
//
//func (a *actorV2) verifyLemma2FirstBlockNextHeight(
//	proposeMsg BFTPropose,
//	block types.BlockInterface,
//) error {
//
//	isValid, err := verifyReProposeHashSignatureFromBlock(proposeMsg.ReProposeHashSignature, block)
//	if err != nil {
//		return err
//	}
//	if !isValid {
//		return fmt.Errorf("Invalid FirstBlockNextHeight ReproposeHashSignature %+v, proposer %+v",
//			proposeMsg.ReProposeHashSignature, block.GetProposer())
//	}
//
//	finalityHeight := block.GetFinalityHeight()
//	previousBlockHeight := block.GetHeight() - 1
//	if finalityHeight != previousBlockHeight {
//		return fmt.Errorf("Invalid FirstBlockNextHeight FinalityHeight expect %+v, but got %+v",
//			previousBlockHeight, finalityHeight)
//	}
//
//	return nil
//}
//
//func (a *actorV2) verifyLemma2ReProposeBlockNextHeight(
//	proposeMsg BFTPropose,
//	block types.BlockInterface,
//	committees []incognitokey.CommitteePublicKey,
//) (bool, error) {
//
//	isValid, err := verifyReProposeHashSignatureFromBlock(proposeMsg.ReProposeHashSignature, block)
//	if err != nil {
//		return false, err
//	}
//	if !isValid {
//		return false, fmt.Errorf("Invalid ReProposeBlockNextHeight ReproposeHashSignature %+v, proposer %+v",
//			proposeMsg.ReProposeHashSignature, block.GetProposer())
//	}
//
//	isValidProof, err := a.verifyFinalityProof(proposeMsg, block, committees)
//	if err != nil {
//		return false, err
//	}
//
//	finalityHeight := block.GetFinalityHeight()
//	if isValidProof {
//		previousBlockHeight := block.GetHeight() - 1
//		if finalityHeight != previousBlockHeight {
//			return false, fmt.Errorf("Invalid ReProposeBlockNextHeight FinalityHeight expect %+v, but got %+v",
//				previousBlockHeight, finalityHeight)
//		}
//	} else {
//		if finalityHeight != 0 {
//			return false, fmt.Errorf("Invalid ReProposeBlockNextHeight FinalityHeight expect %+v, but got %+v",
//				0, finalityHeight)
//		}
//	}
//
//	return isValidProof, nil
//}
//
//func (a *actorV2) verifyFinalityProof(
//	proposeMsg BFTPropose,
//	block types.BlockInterface,
//	committees []incognitokey.CommitteePublicKey,
//) (bool, error) {
//
//	finalityProof := proposeMsg.FinalityProof
//
//	previousBlockHash := block.GetPrevHash()
//	producer := block.GetProducer()
//	rootHash := block.GetAggregateRootHash()
//	beginTimeSlot := common.CalculateTimeSlot(block.GetProduceTime())
//	currentTimeSlot := common.CalculateTimeSlot(block.GetProposeTime())
//
//	if int(currentTimeSlot-beginTimeSlot) != len(finalityProof.ReProposeHashSignature) {
//		a.logger.Infof("Failed to verify finality proof, expect number of proof %+v, but got %+v",
//			int(currentTimeSlot-beginTimeSlot), len(finalityProof.ReProposeHashSignature))
//		return false, nil
//	}
//
//	proposerBase58List := []string{}
//	for reProposeTimeSlot := beginTimeSlot; reProposeTimeSlot < currentTimeSlot; reProposeTimeSlot++ {
//		reProposer, _ := a.chain.GetProposerByTimeSlotFromCommitteeList(reProposeTimeSlot, committees)
//		reProposerBase58, _ := reProposer.ToBase58()
//		proposerBase58List = append(proposerBase58List, reProposerBase58)
//	}
//
//	err := finalityProof.Verify(
//		previousBlockHash,
//		producer,
//		beginTimeSlot,
//		proposerBase58List,
//		rootHash,
//	)
//	if err != nil {
//		return false, err
//	}
//
//	return true, nil
//}
//
//func (a *actorV2) addFinalityProof(
//	block types.BlockInterface,
//	reProposeHashSignature string,
//	proof FinalityProof,
//) error {
//	previousHash := block.GetPrevHash()
//	beginTimeSlot := common.CalculateTimeSlot(block.GetProduceTime())
//	currentTimeSlot := common.CalculateTimeSlot(block.GetProposeTime())
//
//	if currentTimeSlot-beginTimeSlot > MAX_FINALITY_PROOF {
//		return nil
//	}
//
//	nextBlockFinalityProof, ok := a.nextBlockFinalityProof[previousHash.String()]
//	if !ok {
//		nextBlockFinalityProof = make(map[int64]string)
//	}
//
//	nextBlockFinalityProof[currentTimeSlot] = reProposeHashSignature
//	a.logger.Infof("Add Finality Proof | Block %+v, %+v, Current Block Sig for Timeslot: %+v",
//		block.GetHeight(), block.Hash().String(), currentTimeSlot)
//
//	index := 0
//	var err error
//	for timeSlot := beginTimeSlot; timeSlot < currentTimeSlot; timeSlot++ {
//		_, ok := nextBlockFinalityProof[timeSlot]
//		if !ok {
//			nextBlockFinalityProof[timeSlot], err = proof.GetProofByIndex(index)
//			if err != nil {
//				return err
//			}
//			a.logger.Infof("Add Finality Proof | Block %+v, %+v, Previous Proof for Timeslot: %+v",
//				block.GetHeight(), block.Hash().String(), timeSlot)
//		}
//		index++
//	}
//
//	a.nextBlockFinalityProof[previousHash.String()] = nextBlockFinalityProof
//
//	return nil
//}
