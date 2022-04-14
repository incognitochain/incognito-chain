package blsbft

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/multiview"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ProposeBlockInfo struct {
	block                   types.BlockInterface
	ReceiveTime             time.Time
	Committees              []incognitokey.CommitteePublicKey
	SigningCommittees       []incognitokey.CommitteePublicKey
	UserKeySet              []signatureschemes2.MiningKey
	Votes                   map[string]*BFTVote //pk->BFTVote
	IsValid                 bool
	HasNewVote              bool
	IsVoted                 bool
	IsCommitted             bool
	ValidVotes              int
	ErrVotes                int
	ProposerSendVote        bool
	ProposerMiningKeyBase58 string
	LastValidateTime        time.Time
	ReProposeHashSignature  string
	IsValidLemma2Proof      bool
	FinalityProof           multiview.FinalityProof
}

func NewProposeBlockInfo() *ProposeBlockInfo {
	return &ProposeBlockInfo{}
}

func (p *ProposeBlockInfo) UnmarshalJSON(data []byte) error {
	type Alias ProposeBlockInfo
	if p.block.Type() == common.BeaconChainKey {
		tempBeaconBlock := struct {
			Block *types.BeaconBlock
			Alias *Alias
		}{}
		err := json.Unmarshal(data, &tempBeaconBlock)
		if err != nil {
			return err
		}
		p.block = tempBeaconBlock.Block
		p.ReceiveTime = tempBeaconBlock.Alias.ReceiveTime
		p.Committees = tempBeaconBlock.Alias.Committees
		p.SigningCommittees = tempBeaconBlock.Alias.SigningCommittees
		p.UserKeySet = tempBeaconBlock.Alias.UserKeySet
		p.Votes = tempBeaconBlock.Alias.Votes
		p.IsValid = tempBeaconBlock.Alias.IsValid
		p.HasNewVote = tempBeaconBlock.Alias.HasNewVote
		p.IsVoted = tempBeaconBlock.Alias.IsVoted
		p.IsCommitted = tempBeaconBlock.Alias.IsCommitted
		p.ValidVotes = tempBeaconBlock.Alias.ValidVotes
		p.ErrVotes = tempBeaconBlock.Alias.ErrVotes
		p.ProposerSendVote = tempBeaconBlock.Alias.ProposerSendVote
		p.ProposerMiningKeyBase58 = tempBeaconBlock.Alias.ProposerMiningKeyBase58
		p.LastValidateTime = tempBeaconBlock.Alias.LastValidateTime
		p.ReProposeHashSignature = tempBeaconBlock.Alias.ReProposeHashSignature
		p.IsValidLemma2Proof = tempBeaconBlock.Alias.IsValidLemma2Proof
		p.FinalityProof = tempBeaconBlock.Alias.FinalityProof

		return nil
	} else {
		tempShardBlock := struct {
			Block *types.ShardBlock
			Alias *Alias
		}{}
		err := json.Unmarshal(data, &tempShardBlock)
		if err != nil {
			return err
		}
		p.block = tempShardBlock.Block
		p.ReceiveTime = tempShardBlock.Alias.ReceiveTime
		p.Committees = tempShardBlock.Alias.Committees
		p.SigningCommittees = tempShardBlock.Alias.SigningCommittees
		p.UserKeySet = tempShardBlock.Alias.UserKeySet
		p.Votes = tempShardBlock.Alias.Votes
		p.IsValid = tempShardBlock.Alias.IsValid
		p.HasNewVote = tempShardBlock.Alias.HasNewVote
		p.IsVoted = tempShardBlock.Alias.IsVoted
		p.IsCommitted = tempShardBlock.Alias.IsCommitted
		p.ValidVotes = tempShardBlock.Alias.ValidVotes
		p.ErrVotes = tempShardBlock.Alias.ErrVotes
		p.ProposerSendVote = tempShardBlock.Alias.ProposerSendVote
		p.ProposerMiningKeyBase58 = tempShardBlock.Alias.ProposerMiningKeyBase58
		p.LastValidateTime = tempShardBlock.Alias.LastValidateTime
		p.ReProposeHashSignature = tempShardBlock.Alias.ReProposeHashSignature
		p.IsValidLemma2Proof = tempShardBlock.Alias.IsValidLemma2Proof
		p.FinalityProof = tempShardBlock.Alias.FinalityProof
		return nil
	}
}

func (p *ProposeBlockInfo) MarshalJSON() ([]byte, error) {
	type Alias ProposeBlockInfo
	_, isBeaconBlock := p.block.(*types.BeaconBlock)
	if !isBeaconBlock {
		data, err := json.Marshal(struct {
			Block *types.ShardBlock
			Alias *Alias
		}{
			Block: p.block.(*types.ShardBlock),
			Alias: &Alias{
				ReceiveTime:             p.ReceiveTime,
				Committees:              p.Committees,
				SigningCommittees:       p.SigningCommittees,
				UserKeySet:              p.UserKeySet,
				Votes:                   p.Votes,
				IsValid:                 p.IsValid,
				HasNewVote:              p.HasNewVote,
				IsVoted:                 p.IsVoted,
				IsCommitted:             p.IsCommitted,
				ValidVotes:              p.ValidVotes,
				ErrVotes:                p.ErrVotes,
				ProposerSendVote:        p.ProposerSendVote,
				ProposerMiningKeyBase58: p.ProposerMiningKeyBase58,
				LastValidateTime:        p.LastValidateTime,
				ReProposeHashSignature:  p.ReProposeHashSignature,
				IsValidLemma2Proof:      p.IsValidLemma2Proof,
				FinalityProof:           p.FinalityProof,
			},
		})
		if err != nil {
			return []byte{}, err
		}
		return data, nil
	} else {
		data, err := json.Marshal(struct {
			Block *types.BeaconBlock
			Alias *Alias
		}{
			Block: p.block.(*types.BeaconBlock),
			Alias: &Alias{
				ReceiveTime:             p.ReceiveTime,
				Committees:              p.Committees,
				SigningCommittees:       p.SigningCommittees,
				UserKeySet:              p.UserKeySet,
				Votes:                   p.Votes,
				IsValid:                 p.IsValid,
				HasNewVote:              p.HasNewVote,
				IsVoted:                 p.IsVoted,
				IsCommitted:             p.IsCommitted,
				ValidVotes:              p.ValidVotes,
				ErrVotes:                p.ErrVotes,
				ProposerSendVote:        p.ProposerSendVote,
				ProposerMiningKeyBase58: p.ProposerMiningKeyBase58,
				LastValidateTime:        p.LastValidateTime,
				ReProposeHashSignature:  p.ReProposeHashSignature,
				IsValidLemma2Proof:      p.IsValidLemma2Proof,
				FinalityProof:           p.FinalityProof,
			},
		})
		if err != nil {
			return []byte{}, err
		}
		return data, nil
	}
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
		ReceiveTime:             time.Now(),
		Votes:                   make(map[string]*BFTVote),
		Committees:              incognitokey.DeepCopy(committees),
		SigningCommittees:       incognitokey.DeepCopy(signingCommittes),
		UserKeySet:              signatureschemes2.DeepCopyMiningKeyArray(userKeySet),
		ProposerMiningKeyBase58: proposerMiningKeyBase58,
	}
}

func newProposeBlockForProposeMsgLemma2(
	proposeMsg *BFTPropose,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	signingCommittees []incognitokey.CommitteePublicKey,
	userKeySet []signatureschemes2.MiningKey,
	proposerMiningKeyBase58 string,
	isValidLemma2 bool,
) *ProposeBlockInfo {
	return &ProposeBlockInfo{
		block:                   block,
		ReceiveTime:             time.Now(),
		Votes:                   make(map[string]*BFTVote),
		Committees:              incognitokey.DeepCopy(committees),
		SigningCommittees:       incognitokey.DeepCopy(signingCommittees),
		UserKeySet:              signatureschemes2.DeepCopyMiningKeyArray(userKeySet),
		ProposerMiningKeyBase58: proposerMiningKeyBase58,
		IsValidLemma2Proof:      isValidLemma2,
		ReProposeHashSignature:  proposeMsg.ReProposeHashSignature,
		FinalityProof:           proposeMsg.FinalityProof,
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
	proposeBlockInfo.ReceiveTime = time.Now()
	proposeBlockInfo.Committees = incognitokey.DeepCopy(committees)
	proposeBlockInfo.SigningCommittees = incognitokey.DeepCopy(signingCommittes)
	proposeBlockInfo.UserKeySet = signatureschemes2.DeepCopyMiningKeyArray(userKeySet)
	proposeBlockInfo.ValidVotes = validVotes
	proposeBlockInfo.ErrVotes = errVotes
}

func newBlockInfoForVoteMsg(chainID int) *ProposeBlockInfo {
	proposeBlockInfo := &ProposeBlockInfo{
		Votes:      make(map[string]*BFTVote),
		HasNewVote: true,
	}
	if chainID == common.BeaconChainID {
		if proposeBlockInfo.block == nil {
			proposeBlockInfo.block = types.NewBeaconBlock()
		}
	} else {
		if proposeBlockInfo.block == nil {
			proposeBlockInfo.block = types.NewShardBlock()
		}
	}

	return proposeBlockInfo
}
