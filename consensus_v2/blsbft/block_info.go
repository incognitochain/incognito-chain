package blsbft

import (
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ProposeBlockInfo struct {
	block                   types.BlockInterface
	receiveTime             time.Time
	committees              []incognitokey.CommitteePublicKey
	signingCommittees       []incognitokey.CommitteePublicKey
	userKeySet              []signatureschemes2.MiningKey
	votes                   map[string]*BFTVote //pk->BFTVote
	isValid                 bool
	hasNewVote              bool
	isVoted                 bool
	isCommitted             bool
	validVotes              int
	errVotes                int
	proposerSendVote        bool
	proposerMiningKeyBase58 string
	lastValidateTime        time.Time
}

//NewProposeBlockInfoValue : new propose block info
func newProposeBlockForProposeMsg(
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	signingCommittes []incognitokey.CommitteePublicKey,
	userKeySet []signatureschemes2.MiningKey,
	proposerMiningKeyBase58 string,
) *ProposeBlockInfo {
	return &ProposeBlockInfo{
		block:                   block,
		votes:                   make(map[string]*BFTVote),
		committees:              incognitokey.DeepCopy(committees),
		signingCommittees:       incognitokey.DeepCopy(signingCommittes),
		userKeySet:              signatureschemes2.DeepCopyMiningKeyArray(userKeySet),
		proposerMiningKeyBase58: proposerMiningKeyBase58,
	}
}

func (proposeBlockInfo *ProposeBlockInfo) addBlockInfo(
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	signingCommittes []incognitokey.CommitteePublicKey,
	userKeySet []signatureschemes2.MiningKey,
	proposerMiningKeyBase58 string,
	validVotes, errVotes int,
) {
	proposeBlockInfo.block = block
	proposeBlockInfo.committees = incognitokey.DeepCopy(committees)
	proposeBlockInfo.signingCommittees = incognitokey.DeepCopy(signingCommittes)
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
