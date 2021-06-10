package blsbft

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
)

type actorBase struct {
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
}

func NewActorBase() *actorBase {
	return &actorBase{}
}

func NewActorBaseWithValue(
	chain Chain,
	chainKey string, chainID int,
	node NodeInterface, logger common.Logger,
) *actorBase {
	res := NewActorBase()
	res.chain = chain
	res.chainKey = chainKey
	res.chainID = chainID
	res.node = node
	res.logger = logger
	res.destroyCh = make(chan struct{})
	res.proposeMessageCh = make(chan BFTPropose)
	res.voteMessageCh = make(chan BFTVote)
	return res
}

func (actorBase *actorBase) IsStarted() bool {
	return actorBase.isStarted
}

func (actorBase *actorBase) GetConsensusName() string {
	return consensusName
}

func (actorBase *actorBase) GetChainKey() string {
	return actorBase.chainKey
}
func (actorBase *actorBase) GetChainID() int {
	return actorBase.chainID
}

func (actorBase *actorBase) Stop() error {
	if actorBase.isStarted {
		actorBase.logger.Info("stop bls-bft consensus for chain", actorBase.chainKey)
		actorBase.isStarted = false
		actorBase.destroyCh <- struct{}{}
		return nil
	}
	return NewConsensusError(ConsensusAlreadyStoppedError, errors.New(actorBase.chainKey))
}

func (actorBase *actorBase) processBFTMsg(msg *wire.MessageBFT) {
	switch msg.Type {
	case MSG_PROPOSE:
		var msgPropose BFTPropose
		err := json.Unmarshal(msg.Content, &msgPropose)
		if err != nil {
			fmt.Println(err)
			return
		}
		actorBase.proposeMessageCh <- msgPropose
	case MSG_VOTE:
		var msgVote BFTVote
		err := json.Unmarshal(msg.Content, &msgVote)
		if err != nil {
			fmt.Println(err)
			return
		}
		actorBase.voteMessageCh <- msgVote
	default:
		actorBase.logger.Critical("???")
		return
	}
}

func (actorBase *actorBase) preValidateVote(blockHash []byte, Vote *vote, candidate []byte) error {
	data := []byte{}
	data = append(data, blockHash...)
	data = append(data, Vote.BLS...)
	data = append(data, Vote.BRI...)
	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, Vote.Confirmation, candidate)
	return err
}

func (actorBase *actorBase) LoadUserKeys(miningKey []signatureschemes2.MiningKey) {
	actorBase.userKeySet = miningKey
	return
}

func (actorBase *actorBase) GetUserPublicKey() *incognitokey.CommitteePublicKey {
	if actorBase.userKeySet != nil {
		key := actorBase.userKeySet[0].GetPublicKey()
		return key
	}
	return nil
}

func (actorBase *actorBase) SignData(data []byte) (string, error) {
	result, err := actorBase.userKeySet[0].BriSignData(data) //, 0, []blsmultisig.PublicKey{e.UserKeySet.PubKey[common.BlsConsensus]})
	if err != nil {
		return "", NewConsensusError(SignDataError, err)
	}

	return base58.Base58Check{}.Encode(result, common.Base58Version), nil
}

func (actorBase *actorBase) ValidateData(data []byte, sig string, publicKey string) error {
	sigByte, _, err := base58.Base58Check{}.Decode(sig)
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	publicKeyByte := []byte(publicKey)
	// if err != nil {
	// 	return consensus.NewConsensusError(consensus.UnExpectedError, err)
	// }
	//fmt.Printf("ValidateData data %v, sig %v, publicKey %v\n", data, sig, publicKeyByte)
	dataHash := new(common.Hash)
	dataHash.NewHash(data)
	_, err = bridgesig.Verify(publicKeyByte, dataHash.GetBytes(), sigByte) //blsmultisig.Verify(sigByte, data, []int{0}, []blsmultisig.PublicKey{publicKeyByte})
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	return nil
}

func (actorBase *actorBase) combineVotes(votes map[string]vote, committee []string) (aggSig []byte, brigSigs [][]byte, validatorIdx []int, err error) {
	var blsSigList [][]byte
	for validator, _ := range votes {
		validatorIdx = append(validatorIdx, common.IndexOfStr(validator, committee))
	}
	sort.Ints(validatorIdx)
	for _, idx := range validatorIdx {
		blsSigList = append(blsSigList, votes[committee[idx]].BLS)
		brigSigs = append(brigSigs, votes[committee[idx]].BRI)
	}

	aggSig, err = blsmultisig.Combine(blsSigList)
	if err != nil {
		return nil, nil, nil, NewConsensusError(CombineSignatureError, err)
	}
	return
}

func (actorBase *actorBase) Run() error {
	panic("Imelement this function")
}
