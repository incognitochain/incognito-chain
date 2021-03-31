package blsbft

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/wire"
)

type actorBase struct {
	chain    ChainInterface
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

func (actorBase *actorBase) Destroy() {
	actorBase.Stop()
}

func (actorBase *actorBase) IsOngoing() bool {
	panic("Implement this function")
}

func (actorBase *actorBase) Stop() error {
	if actorBase.isStarted {
		actorBase.logger.Info("stop bls-bft consensus for chain", actorBase.chainKey)
		actorBase.isStarted = false
		return nil
	}
	return NewConsensusError(ConsensusAlreadyStoppedError, errors.New(actorBase.chainKey))
}

func (actorBase *actorBase) ProcessBFTMsg(msg *wire.MessageBFT) {
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
