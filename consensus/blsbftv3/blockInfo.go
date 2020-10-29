package blsbftv3

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ProposeBlockInfo struct {
	block      types.BlockInterface
	committees []incognitokey.CommitteePublicKey
	votes      map[string]BFTVote //pk->BFTVote
	isValid    bool
	hasNewVote bool
}

//NewProposeBlockInfoValue : new propose block info
func newProposeBlockInfoValue(
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	votes map[string]BFTVote,
	isValid, hasNewVote bool,
) *ProposeBlockInfo {
	return &ProposeBlockInfo{
		block:      block,
		committees: committees,
		votes:      votes,
		isValid:    isValid,
		hasNewVote: hasNewVote,
	}
}

func (proposeBlockInfo *ProposeBlockInfo) newBlockInfo(
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey) {
	proposeBlockInfo.block = block
	proposeBlockInfo.committees = committees
}
