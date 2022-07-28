package blsbft

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/bridgesig"
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
	FinalityProof           FinalityProof
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
		p.IsValid = tempBeaconBlock.Alias.IsValid
		p.HasNewVote = true //force check 2/3+1 after init
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
		p.IsValid = tempShardBlock.Alias.IsValid
		p.HasNewVote = true //force check 2/3+1 after init
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

type FinalityProof struct {
	ReProposeHashSignature []string
}

func NewFinalityProof() *FinalityProof {
	return &FinalityProof{}
}

func (f *FinalityProof) AddProof(reProposeHashSig string) {
	f.ReProposeHashSignature = append(f.ReProposeHashSignature, reProposeHashSig)
}

func (f *FinalityProof) GetProofByIndex(index int) (string, error) {
	if index < 0 || index >= len(f.ReProposeHashSignature) {
		return "", fmt.Errorf("Proof index %+v, is not valid. Number of Proof %+v", index, len(f.ReProposeHashSignature))
	}
	proof := f.ReProposeHashSignature[index]
	if proof == "" {
		return "", fmt.Errorf("invalid proof zero length")
	}
	return f.ReProposeHashSignature[index], nil
}

func (f *FinalityProof) Verify(
	previousBlockHash common.Hash,
	producer string,
	beginTimeSlot int64,
	proposers []string,
	rootHash common.Hash,
) error {

	for i := 0; i < len(f.ReProposeHashSignature); i++ {
		reProposer := proposers[i]
		reProposeTimeSlot := beginTimeSlot + int64(i)
		sig := f.ReProposeHashSignature[i]

		isValid, err := verifyReProposeHashSignature(
			sig,
			previousBlockHash,
			producer,
			beginTimeSlot,
			reProposer,
			reProposeTimeSlot,
			rootHash,
		)
		if err != nil {
			return fmt.Errorf("verification failed verifyFinalityProof "+
				"Re-ProposeTimeSlot %+v, ReProposer %+v, error %+v",
				reProposeTimeSlot, reProposer, err)
		}
		if !isValid {
			return fmt.Errorf("invalid Signature verifyFinalityProof "+
				"Re-ProposeTimeSlot %+v, ReProposer %+v", reProposeTimeSlot, reProposer)
		}
	}

	return nil
}

//previousblockhash, producerTimeslot, Producer, proposerTimeslot, Proposer roothash
type ReProposeBlockInfo struct {
	PreviousBlockHash common.Hash
	Producer          string
	ProducerTimeSlot  int64
	Proposer          string
	ProposerTimeSlot  int64
	RootHash          common.Hash
}

func createReProposeHashSignature(privateKey []byte, block types.BlockInterface) (string, error) {

	reProposeBlockInfo := newReProposeBlockInfo(
		block.GetPrevHash(),
		block.GetProducer(),
		common.CalculateTimeSlot(block.GetProduceTime()),
		block.GetProposer(),
		common.CalculateTimeSlot(block.GetProposeTime()),
		block.GetAggregateRootHash(),
	)

	return reProposeBlockInfo.Sign(privateKey)
}

func verifyReProposeHashSignature(
	sig string,
	previousBlockHash common.Hash,
	producerBase58 string,
	producerTimeSlot int64,
	proposerBase58 string,
	proposerTimeSlot int64,
	rootHash common.Hash,
) (bool, error) {

	proposer := incognitokey.CommitteePublicKey{}

	_ = proposer.FromString(proposerBase58)
	publicKey := proposer.MiningPubKey[common.BridgeConsensus]

	reProposeBlockInfo := newReProposeBlockInfo(
		previousBlockHash,
		producerBase58,
		producerTimeSlot,
		proposerBase58,
		proposerTimeSlot,
		rootHash,
	)

	return reProposeBlockInfo.VerifySignature(sig, publicKey)
}

func verifyReProposeHashSignatureFromBlock(sig string, block types.BlockInterface) (bool, error) {
	return verifyReProposeHashSignature(
		sig,
		block.GetPrevHash(),
		block.GetProducer(),
		common.CalculateTimeSlot(block.GetProduceTime()),
		block.GetProposer(),
		common.CalculateTimeSlot(block.GetProposeTime()),
		block.GetAggregateRootHash(),
	)
}

func newReProposeBlockInfo(previousBlockHash common.Hash, producer string, producerTimeSlot int64, proposer string, proposerTimeSlot int64, rootHash common.Hash) *ReProposeBlockInfo {
	return &ReProposeBlockInfo{PreviousBlockHash: previousBlockHash, Producer: producer, ProducerTimeSlot: producerTimeSlot, Proposer: proposer, ProposerTimeSlot: proposerTimeSlot, RootHash: rootHash}
}

func (r ReProposeBlockInfo) Hash() common.Hash {
	data, _ := json.Marshal(&r)
	return common.HashH(data)
}

func (r ReProposeBlockInfo) Sign(privateKey []byte) (string, error) {

	hash := r.Hash()

	sig, err := bridgesig.Sign(privateKey, hash.Bytes())
	if err != nil {
		return "", err
	}

	sigBase58 := base58.Base58Check{}.Encode(sig, common.Base58Version)

	return sigBase58, nil
}

func (r ReProposeBlockInfo) VerifySignature(sigBase58 string, publicKey []byte) (bool, error) {

	hash := r.Hash()
	sig, _, _ := base58.Base58Check{}.Decode(sigBase58)

	isValid, err := bridgesig.Verify(publicKey, hash.Bytes(), sig)
	if err != nil {
		return false, err
	}

	return isValid, nil
}
