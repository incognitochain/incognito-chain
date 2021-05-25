package blsbftv3

import (
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ProposeBlockInfo struct {
	receiveTime             time.Time
	block                   types.BlockInterface
	committees              []incognitokey.CommitteePublicKey
	votes                   map[string]*BFTVote //pk->BFTVote
	isValid                 bool
	hasNewVote              bool
	isVoted                 bool
	proposerSendVote        bool
	proposerMiningKeyBase58 string
	lastValidateTime        time.Time
	isCommitted             bool
}

//NewProposeBlockInfoValue : new propose block info
func newProposeBlockForProposeMsg(
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	votes map[string]*BFTVote,
	hasNewVote bool,
	proposerMiningKeyBase58 string,
) *ProposeBlockInfo {
	return &ProposeBlockInfo{
		block:                   block,
		committees:              incognitokey.DeepCopy(committees),
		votes:                   votes,
		hasNewVote:              hasNewVote,
		receiveTime:             time.Now(),
		proposerMiningKeyBase58: proposerMiningKeyBase58,
	}
}

func (proposeBlockInfo *ProposeBlockInfo) addBlockInfo(
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	proposerMiningKeyBase58 string) {
	proposeBlockInfo.block = block
	proposeBlockInfo.committees = incognitokey.DeepCopy(committees)
}

func newBlockInfoForVoteMsg() *ProposeBlockInfo {
	return &ProposeBlockInfo{
		votes:      make(map[string]*BFTVote),
		hasNewVote: true,
	}
}
