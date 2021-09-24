package blsbft

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
)

type VoteMessageEnvironment struct {
	userKey           *signatureschemes2.MiningKey
	signingCommittees []incognitokey.CommitteePublicKey
	portalParamV4     portalv4.PortalParams
}

func NewVoteMessageEnvironment(userKey *signatureschemes2.MiningKey, signingCommittees []incognitokey.CommitteePublicKey, portalParamV4 portalv4.PortalParams) *VoteMessageEnvironment {
	return &VoteMessageEnvironment{userKey: userKey, signingCommittees: signingCommittees, portalParamV4: portalParamV4}
}

type IVoteRule interface {
	ValidateVote(*ProposeBlockInfo) *ProposeBlockInfo
	CreateVote(*VoteMessageEnvironment, types.BlockInterface) (*BFTVote, error)
}

type VoteRule struct {
	logger common.Logger
}

func NewVoteRule(logger common.Logger) *VoteRule {
	return &VoteRule{logger: logger}
}

func (v VoteRule) ValidateVote(proposeBlockInfo *ProposeBlockInfo) *ProposeBlockInfo {
	validVote := 0
	errVote := 0

	committees := make(map[string]int)
	if len(proposeBlockInfo.votes) != 0 {
		for i, v := range proposeBlockInfo.signingCommittees {
			committees[v.GetMiningKeyBase58(common.BlsConsensus)] = i
		}
	}

	for id, vote := range proposeBlockInfo.votes {
		dsaKey := []byte{}
		if vote.IsValid == 0 {
			if value, ok := committees[vote.Validator]; ok {
				dsaKey = proposeBlockInfo.signingCommittees[value].MiningPubKey[common.BridgeConsensus]
			} else {
				v.logger.Error("Receive vote from nonCommittee member")
				continue
			}
			if len(dsaKey) == 0 {
				v.logger.Error("canot find dsa key")
				continue
			}

			err := vote.validateVoteOwner(dsaKey)
			if err != nil {
				v.logger.Error(dsaKey)
				v.logger.Error(err)
				proposeBlockInfo.votes[id].IsValid = -1
				errVote++
			} else {
				proposeBlockInfo.votes[id].IsValid = 1
				validVote++
			}
		} else {
			validVote++
		}
	}

	v.logger.Info("Number of Valid Vote", validVote, "| Number Of Error Vote", errVote)
	proposeBlockInfo.hasNewVote = false
	for key, value := range proposeBlockInfo.votes {
		if value.IsValid == -1 {
			delete(proposeBlockInfo.votes, key)
		}
	}

	proposeBlockInfo.addBlockInfo(
		proposeBlockInfo.block,
		proposeBlockInfo.committees,
		proposeBlockInfo.signingCommittees,
		proposeBlockInfo.userKeySet,
		validVote,
		errVote,
	)

	return proposeBlockInfo
}

func (v VoteRule) CreateVote(env *VoteMessageEnvironment, block types.BlockInterface) (*BFTVote, error) {

	vote, err := createVote(env.userKey, block, env.signingCommittees, env.portalParamV4)
	if err != nil {
		v.logger.Error(err)
		return nil, err
	}

	return vote, nil
}

func createVote(
	userKey *signatureschemes2.MiningKey,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	portalParamsV4 portalv4.PortalParams,
) (*BFTVote, error) {
	var vote = new(BFTVote)
	bytelist := []blsmultisig.PublicKey{}
	selfIdx := 0
	userBLSPk := userKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus)
	for i, v := range committees {
		if v.GetMiningKeyBase58(common.BlsConsensus) == userBLSPk {
			selfIdx = i
		}
		bytelist = append(bytelist, v.MiningPubKey[common.BlsConsensus])
	}

	blsSig, err := userKey.BLSSignData(block.Hash().GetBytes(), selfIdx, bytelist)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	bridgeSig := []byte{}
	if metadata.HasBridgeInstructions(block.GetInstructions()) {
		bridgeSig, err = userKey.BriSignData(block.Hash().GetBytes())
		if err != nil {
			return nil, NewConsensusError(UnExpectedError, err)
		}
	}

	// check and sign on unshielding external tx for Portal v4
	portalSigs, err := portalprocessv4.CheckAndSignPortalUnshieldExternalTx(userKey.PriKey[common.BridgeConsensus], block.GetInstructions(), portalParamsV4)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}

	vote.BLS = blsSig
	vote.BRI = bridgeSig
	vote.PortalSigs = portalSigs
	vote.BlockHash = block.Hash().String()
	vote.Validator = userBLSPk
	vote.PrevBlockHash = block.GetPrevHash().String()
	err = vote.signVote(userKey)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	return vote, nil
}
