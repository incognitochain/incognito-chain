package mock

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type Consensus struct {
	Blockchain *blockchain.BlockChain
}

func (c *Consensus) ValidateProducerPosition(blk common.BlockInterface, lastProposerIdx int, committee []incognitokey.CommitteePublicKey, minCommitteeSize int) error {
	return nil
}

func (c *Consensus) ValidateProducerSig(block common.BlockInterface, consensusType string) error {
	return nil
}

func (c *Consensus) ValidateBlockCommitteSig(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
	return nil
}
