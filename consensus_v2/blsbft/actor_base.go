package blsbft

import (
	"errors"

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
