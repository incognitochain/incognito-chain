package blsbftv4

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ProposeBlockInfo struct {
	block            types.BlockInterface
	committees       []incognitokey.CommitteePublicKey
	signingCommittes []incognitokey.CommitteePublicKey
	userKeySet       []signatureschemes2.MiningKey
	votes            map[string]*BFTVote //pk->BFTVote
	isValid          bool
	hasNewVote       bool
	isVoted          bool
	isCommitted      bool
	validVotes       int
	errVotes         int
}

//NewProposeBlockInfoValue : new propose block info
func newProposeBlockForProposeMsg(
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	signingCommittes []incognitokey.CommitteePublicKey,
	userKeySet []signatureschemes2.MiningKey,
	votes map[string]*BFTVote,
) *ProposeBlockInfo {
	return &ProposeBlockInfo{
		block:            block,
		committees:       incognitokey.DeepCopy(committees),
		signingCommittes: incognitokey.DeepCopy(signingCommittes),
		userKeySet:       signatureschemes2.DeepCopyMiningKeyArray(userKeySet),
		votes:            votes,
	}
}

func (proposeBlockInfo *ProposeBlockInfo) addBlockInfo(
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	signingCommittes []incognitokey.CommitteePublicKey,
	userKeySet []signatureschemes2.MiningKey,
	validVotes, errVotes int,
) {
	proposeBlockInfo.block = block
	proposeBlockInfo.committees = incognitokey.DeepCopy(committees)
	proposeBlockInfo.signingCommittes = incognitokey.DeepCopy(signingCommittes)
	proposeBlockInfo.userKeySet = signatureschemes2.DeepCopyMiningKeyArray(userKeySet)
	proposeBlockInfo.validVotes = validVotes
	proposeBlockInfo.errVotes = errVotes
}

func newBlockInfoForVoteMsg() *ProposeBlockInfo {
	return &ProposeBlockInfo{
		votes:      make(map[string]*BFTVote),
		hasNewVote: true,
	}
}
