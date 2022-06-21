package blsbft

import (
	"fmt"
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
	if len(proposeBlockInfo.Votes) != 0 {
		for i, v := range proposeBlockInfo.SigningCommittees {
			committees[v.GetMiningKeyBase58(common.BlsConsensus)] = i
		}
	}

	for id, vote := range proposeBlockInfo.Votes {
		dsaKey := []byte{}
		switch vote.IsValid {
		case 0:
			if value, ok := committees[vote.Validator]; ok {
				dsaKey = proposeBlockInfo.SigningCommittees[value].MiningPubKey[common.BridgeConsensus]
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
				proposeBlockInfo.Votes[id].IsValid = -1
				errVote++
			} else {
				proposeBlockInfo.Votes[id].IsValid = 1
				validVote++
			}
		case 1:
			validVote++
		case -1:
			errVote++
		}
	}

	v.logger.Info("Number of Valid Vote", validVote, "| Number Of Error Vote", errVote)
	proposeBlockInfo.HasNewVote = false
	proposeBlockInfo.ValidVotes = validVote
	proposeBlockInfo.ErrVotes = errVote

	for key, value := range proposeBlockInfo.Votes {
		if value.IsValid == -1 {
			delete(proposeBlockInfo.Votes, key)
		}
	}

	return proposeBlockInfo
}

func (v VoteRule) CreateVote(env *VoteMessageEnvironment, block types.BlockInterface) (*BFTVote, error) {

	vote, err := CreateVote(env.userKey, block, env.signingCommittees, env.portalParamV4)
	if err != nil {
		v.logger.Error(err)
		return nil, err
	}

	return vote, nil
}

func CreateVote(
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

	blsSig, err := userKey.BLSSignData(block.ProposeHash().GetBytes(), selfIdx, bytelist)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}

	bridgeSig := []byte{}
	if metadata.HasBridgeInstructions(block.GetInstructions()) {
		bridgeSig, err = userKey.BriSignData(block.Hash().GetBytes()) //proof is agg sig on block hash (not propose hash)
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
	vote.BlockHash = block.ProposeHash().String()
	vote.Validator = userBLSPk
	vote.ProduceTimeSlot = common.CalculateTimeSlot(block.GetProduceTime())
	vote.ProposeTimeSlot = common.CalculateTimeSlot(block.GetProposeTime())
	vote.PrevBlockHash = block.GetPrevHash().String()
	vote.BlockHeight = block.GetHeight()
	vote.CommitteeFromBlock = block.CommitteeFromBlock()
	vote.ChainID = block.GetShardID()
	err = vote.signVote(userKey)
	if err != nil {
		return nil, NewConsensusError(UnExpectedError, err)
	}
	return vote, nil
}

type NoVoteRule struct {
	logger common.Logger
}

func NewNoVoteRule(logger common.Logger) *NoVoteRule {
	return &NoVoteRule{logger: logger}
}

func (v NoVoteRule) ValidateVote(proposeBlockInfo *ProposeBlockInfo) *ProposeBlockInfo {
	validVote := 0
	errVote := 0

	committees := make(map[string]int)
	if len(proposeBlockInfo.Votes) != 0 {
		for i, v := range proposeBlockInfo.SigningCommittees {
			committees[v.GetMiningKeyBase58(common.BlsConsensus)] = i
		}
	}

	for id, vote := range proposeBlockInfo.Votes {
		dsaKey := []byte{}
		if vote.IsValid == 0 {
			if value, ok := committees[vote.Validator]; ok {
				dsaKey = proposeBlockInfo.SigningCommittees[value].MiningPubKey[common.BridgeConsensus]
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
				proposeBlockInfo.Votes[id].IsValid = -1
				errVote++
			} else {
				proposeBlockInfo.Votes[id].IsValid = 1
				validVote++
			}
		} else {
			validVote++
		}
	}

	v.logger.Info("Number of Valid Vote", validVote, "| Number Of Error Vote", errVote)
	proposeBlockInfo.HasNewVote = false
	for key, value := range proposeBlockInfo.Votes {
		if value.IsValid == -1 {
			delete(proposeBlockInfo.Votes, key)
		}
	}

	return proposeBlockInfo
}

func (i NoVoteRule) CreateVote(environment *VoteMessageEnvironment, block types.BlockInterface) (*BFTVote, error) {
	i.logger.Criticalf("NO VOTE")
	return nil, fmt.Errorf("No vote for block %+v, %+v", block.GetHeight(), block.Hash().String())
}

type IHandleVoteMessageRule interface {
	IsHandle() bool
}

type HandleVoteMessage struct {
}

func NewHandleVoteMessage() *HandleVoteMessage {
	return &HandleVoteMessage{}
}

func (h HandleVoteMessage) IsHandle() bool {
	return true
}

type NoHandleVoteMessage struct {
}

func NewNoHandleVoteMessage() *NoHandleVoteMessage {
	return &NoHandleVoteMessage{}
}

func (h NoHandleVoteMessage) IsHandle() bool {
	return false
}
