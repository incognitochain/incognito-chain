package blsbft

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb_consensus"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metrics/monitor"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
	"github.com/incognitochain/incognito-chain/wire"
	"log"
	"sort"
	"time"
)

//Tendermint-like consensus
//propose -> prevote -> vote -> commit

type actorV3 struct {
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
	receiveBlockByHash map[string]*ProposeBlockInfo    //blockProposeHash -> blockInfo
	voteHistory        map[uint64]types.BlockInterface // bestview height (previsous height )-> block

	blockVersion int

	currentBestViewHeight uint64
}

func NewActorV3() *actorV3 {
	return &actorV3{}
}

func NewActorV3WithValue(
	chain Chain,
	committeeChain CommitteeChainHandler,
	chainKey string, blockVersion, chainID int,
	node NodeInterface, logger common.Logger,
) *actorV3 {
	a := newActorV3WithValue(
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

func newActorV3WithValue(
	chain Chain,
	committeeChain CommitteeChainHandler,
	chainKey string, blockVersion, chainID int,
	node NodeInterface, logger common.Logger,
) *actorV3 {
	var err error
	a := NewActorV3()
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

	return a
}

func (a *actorV3) GetSortedReceiveBlockByHeight(blockHeight uint64) []*ProposeBlockInfo {
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

func (a *actorV3) AddReceiveBlockByHash(blockHash string, proposeBlockInfo *ProposeBlockInfo) error {

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

func (a *actorV3) GetReceiveBlockByHash(blockHash string) (*ProposeBlockInfo, bool) {
	res, ok := a.receiveBlockByHash[blockHash]
	return res, ok
}

func (a *actorV3) CleanReceiveBlockByHash(blockHash string) error {

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

func (a *actorV3) AddVoteHistory(blockHeight uint64, block types.BlockInterface) error {

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

func (a *actorV3) GetVoteHistory(blockHeight uint64) (types.BlockInterface, bool) {
	res, ok := a.voteHistory[blockHeight]
	return res, ok
}

func (a *actorV3) CleanVoteHistory(blockHeight uint64) error {

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

func (a *actorV3) AddCurrentTimeSlotProposeHistory() error {

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

func (a *actorV3) GetCurrentTimeSlotProposeHistory() bool {
	_, ok := a.proposeHistory[a.currentTimeSlot]
	return ok
}

func (a *actorV3) CleanProposeHistory(timeSlot int64) error {

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

func (a actorV3) GetConsensusName() string {
	return common.BlsConsensus
}

func (a actorV3) GetChainKey() string {
	return a.chainKey
}

func (a actorV3) GetChainID() int {
	return a.chainID
}

func (a actorV3) GetUserPublicKey() *incognitokey.CommitteePublicKey {
	if a.userKeySet != nil {
		key := a.userKeySet[0].GetPublicKey()
		return key
	}
	return nil
}

func (a actorV3) BlockVersion() int {
	return a.blockVersion
}

func (a *actorV3) SetBlockVersion(version int) {
	a.blockVersion = version
}

func (a *actorV3) Stop() error {
	if a.isStarted {
		a.logger.Infof("stop bls-bft-%+v consensus for chain %+v", a.blockVersion, a.chainKey)
	}
	a.isStarted = false
	return nil
}

func (a *actorV3) Destroy() {
	a.destroyCh <- struct{}{}
}

func (a actorV3) IsStarted() bool {
	return a.isStarted
}

func (a *actorV3) ProcessBFTMsg(msgBFT *wire.MessageBFT) {
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

func (a *actorV3) LoadUserKeys(miningKey []signatureschemes2.MiningKey) {
	a.userKeySet = miningKey
	return
}

func (a actorV3) ValidateData(data []byte, sig string, publicKey string) error {
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

func (a *actorV3) SignData(data []byte) (string, error) {
	//, 0, []blsmultisig.PublicKey{e.UserKeySet.PubKey[common.BlsConsensus]})
	result, err := a.userKeySet[0].BriSignData(data)
	if err != nil {
		return "", NewConsensusError(SignDataError, err)
	}

	return base58.Base58Check{}.Encode(result, common.Base58Version), nil
}

func (a *actorV3) Start() error {
	if !a.isStarted {
		a.logger.Infof("start bls-bft-%+v consensus for chain %+v", a.blockVersion, a.chainKey)
	}
	a.isStarted = true
	return nil
}

func (a *actorV3) isUserKeyProposer(
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

func (a *actorV3) getValidatorIndex(committees []incognitokey.CommitteePublicKey, validator string) (int, *incognitokey.CommitteePublicKey) {
	for id, c := range committees {
		if validator == c.GetMiningKeyBase58(common.BlsConsensus) {
			return id, &c
		}
	}
	return -1, nil
}

func (a *actorV3) createBLSAggregatedSignatures(
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

func (a *actorV3) addValidationData(userMiningKey signatureschemes2.MiningKey, block types.BlockInterface) (types.BlockInterface, error) {

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

func (a *actorV3) sendBFTProposeMsg(
	bftPropose *BFTPropose,
) error {

	msg, _ := a.makeBFTProposeMsg(bftPropose, a.chainKey, a.currentTimeSlot)
	go a.ProcessBFTMsg(msg.(*wire.MessageBFT))
	go a.node.PushMessageToChain(msg, a.chain)

	return nil
}

func (a *actorV3) preValidateVote(blockHash []byte, vote *BFTVote, candidate []byte) error {
	data := []byte{}
	data = append(data, blockHash...)
	data = append(data, vote.BLS...)
	data = append(data, vote.BRI...)
	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, vote.Confirmation, candidate)
	return err
}

// getCommitteeForBlock base on the block version to retrieve the right committee list
func (a *actorV3) getCommitteeForNewBlock(
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

func (a *actorV3) getUserKeySetForSigning(
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

func (a *actorV3) getCommitteesAndCommitteeViewHash() (
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

func (a *actorV3) handleCleanMem() {

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

}

func (a *actorV3) makeBFTProposeMsg(
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

func (a *actorV3) makeBFTVoteMsg(vote *BFTVote, chainKey string, ts int64, height uint64) (wire.Message, error) {
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

func (a *actorV3) makeBFTRequestBlk(request BFTRequestBlock, peerID string, chainKey string) (wire.Message, error) {
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

func (a *actorV3) run() error {
	go func() {
		//init view maps
		ticker := time.Tick(200 * time.Millisecond)
		cleanMemTicker := time.Tick(5 * time.Minute)
		a.logger.Infof("init bls-bft-%+v consensus for chain %+v", a.blockVersion, a.chainKey)
		time.Sleep(time.Duration(common.TIMESLOT-1) * time.Second)
		for { //actor loop
			if !a.isStarted { //sleep if this process is not start
				time.Sleep(time.Second)
				select {
				case <-a.proposeMessageCh:
				case <-a.voteMessageCh:
				default:
				}

				continue
			}

			select {
			case <-a.destroyCh:
				a.logger.Infof("exit bls-bft-%+v consensus for chain %+v", a.blockVersion, a.chainKey)
				close(a.destroyCh)
				return
			case proposeMsg := <-a.proposeMessageCh:
				if ActorRuleBuilderContext.HandleProposeRule != HANDLE_PROPOSE_MESSAGE_NORMAL {
					continue
				}
				err := a.handleProposeMsg(proposeMsg)
				if err != nil {
					a.logger.Error(err)
					continue
				}

			case voteMsg := <-a.voteMessageCh:
				if ActorRuleBuilderContext.HandleVoteRule != HANDLE_VOTE_MESSAGE_COLLECT {
					continue
				}
				switch voteMsg.Phase {
				case "prevote":
					err := a.handlePreVoteMsg(voteMsg)
					if err != nil {
						a.logger.Error(err)
						continue
					}
				case "vote":
					err := a.handleVoteMsg(voteMsg)
					if err != nil {
						a.logger.Error(err)
						continue
					}
				default:
					a.logger.Error("Cannot find vote type!")
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
				a.currentBestViewHeight = bestView.GetHeight()

				//set round for monitor
				round := a.currentTimeSlot - common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime())
				monitor.SetGlobalParam("RoundKey", fmt.Sprintf("%d_%d", bestView.GetHeight(), round))

				if newTimeSlot {
					a.logger.Info("")
					a.logger.Info("======================================================")
					if ActorRuleBuilderContext.CreateRule == CREATE_RULE_NORMAL {
						a.maybeProposeBlock()
					}
				}

				//validatingreceived data: propose, prevote, vote
				for _, proposeInfo := range a.receiveBlockByHash {
					if ActorRuleBuilderContext.ValidatorRule == VALIDATOR_NO_VALIDATE {
						break
					}

					//get propose info at current timeslot
					if proposeInfo.block != nil && common.CalculateTimeSlot(proposeInfo.block.GetProposeTime()) == a.currentTimeSlot {
						//validate the propose block
						err := a.validateBlock(proposeInfo)
						if err != nil {
							a.logger.Errorf("%v", err)
						}
						//validate pre vote this current propose block
						a.validatePreVote(proposeInfo)

						//validate vote this current propose block
						a.validateVote(proposeInfo)
					}
				}

				if ActorRuleBuilderContext.PreVoteRule == VOTE_RULE_VOTE {
					//prevote for this timeslot
					a.maybePreVoteMsg()
				}

				if ActorRuleBuilderContext.VoteRule == VOTE_RULE_VOTE {
					//vote for this timeslot
					a.maybeVoteMsg()
				}

				if ActorRuleBuilderContext.InsertRule == INSERT_AND_BROADCAST {
					//commit for this timeslot
					a.maybeCommit()
				}

			}
		}
	}()
	return nil
}

//get lock block hash, which is blockhash that we had send vote message
//so that, we will not prevote for other block
func (a *actorV3) getLockBlockHash(proposeBlockHeight uint64) (info *ProposeBlockInfo) {
	for _, proposeBlockInfo := range a.receiveBlockByHash {
		if proposeBlockInfo.block.GetHeight() == proposeBlockHeight && proposeBlockInfo.IsVoted {
			//get latest propose block info that has 2/3+ prevote => Proof of lock change
			if info != nil {
				if info.block.GetProposeTime() < proposeBlockInfo.block.GetProposeTime() {
					info = proposeBlockInfo
				}
			} else {
				info = proposeBlockInfo
			}
		}
	}
	return info
}

//job to validate propose block
func (a *actorV3) validateBlock(proposeBlockInfo *ProposeBlockInfo) error {
	//not validate if already valid
	if proposeBlockInfo.IsValid {
		return nil
	}

	//interval 1s
	if time.Since(proposeBlockInfo.LastValidateTime).Seconds() < 1 {
		return nil
	}

	//should be next block height
	if proposeBlockInfo.block.GetHeight() != a.currentBestViewHeight+1 {
		return errors.New("Not expected height!")
	}

	//should be the same with lock block hash or has valid POLC (POLC R > lockR)
	if !proposeBlockInfo.ValidPOLC {
		lockProposeInfo := a.getLockBlockHash(a.currentBestViewHeight + 1)
		if lockProposeInfo != nil && lockProposeInfo.block.Hash().String() != proposeBlockInfo.block.Hash().String() {
			return errors.New("Not expected locked blockhash!")
		}
	}

	proposeBlockInfo.LastValidateTime = time.Now()
	err := a.chain.ValidatePreSignBlock(proposeBlockInfo.block, proposeBlockInfo.SigningCommittees, proposeBlockInfo.Committees)
	if err != nil {
		return errors.New("Block is invalidated!")
	}
	proposeBlockInfo.IsValid = true
	return nil
}
