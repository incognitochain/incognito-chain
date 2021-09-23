package blsbft

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
)

type VoteMessageEnvironment struct {
	userKey           *signatureschemes2.MiningKey
	signingCommittees []incognitokey.CommitteePublicKey
	portalParamV4     portalv4.PortalParams
}

type IVoteRule interface {
	HandleBFTVoteMsg(*BFTVote) error
	ValidateVote(*ProposeBlockInfo) *ProposeBlockInfo
	SendVote(*VoteMessageEnvironment, types.BlockInterface) error
}
