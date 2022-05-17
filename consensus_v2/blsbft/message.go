package blsbft

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/config"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"

	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
)

const (
	MSG_PROPOSE   = "propose"
	MSG_VOTE      = "vote"
	MsgRequestBlk = "getblk"
)

type BFTPropose struct {
	PeerID                 string
	Block                  json.RawMessage
	ReProposeHashSignature string
	FinalityProof          FinalityProof
	BestBlockConsensusData map[int]BestBlockConsensusData
}

type BestBlockConsensusData struct {
	BlockHash      common.Hash
	BlockHeight    uint64
	FinalityHeight uint64
	Proposer       string
	ProposerTime   int64
	ValidationData string
}

type BFTVote struct {
	RoundKey           string
	PrevBlockHash      string
	BlockHeight        uint64
	BlockHash          string //this is propose block hash
	Validator          string
	BLS                []byte
	BRI                []byte
	Confirmation       []byte
	IsValid            int // 0 not process, 1 valid, -1 not valid
	ProduceTimeSlot    int64
	ProposeTimeSlot    int64
	CommitteeFromBlock common.Hash
	ChainID            int
	// Portal v4
	PortalSigs []*portalprocessv4.PortalSig
}

type BFTRequestBlock struct {
	BlockHash string
	PeerID    string
}

func (s *BFTVote) isEmptyDataForByzantineDetector() bool {

	if s.BlockHeight == 0 || s.ProduceTimeSlot == 0 || s.ProposeTimeSlot == 0 {
		return true
	}

	return false
}

func (s *BFTVote) signVote(key *signatureschemes2.MiningKey) error {

	data := []byte{}

	if s.BlockHeight < config.Param().ConsensusParam.ByzantineDetectorHeight {
		data = append(data, s.BlockHash...)
		data = append(data, s.BLS...)
		data = append(data, s.BRI...)
	} else {
		data = append(data, s.BlockHash...)
		data = append(data, s.BLS...)
		data = append(data, s.BRI...)
		data = append(data, common.Uint64ToBytes(s.BlockHeight)...)
		data = append(data, common.Int64ToBytes(s.ProduceTimeSlot)...)
		data = append(data, common.Int64ToBytes(s.ProposeTimeSlot)...)
		data = append(data, []byte(s.Validator)...)
		data = append(data, []byte(s.PrevBlockHash)...)
		data = append(data, s.CommitteeFromBlock[:]...)
		data = append(data, common.Int64ToBytes(int64(s.ChainID))...)
	}

	data = common.HashB(data)
	var err error
	s.Confirmation, err = key.BriSignData(data)
	return err
}

func (s *BFTVote) validateVoteOwner(ownerPk []byte) error {

	data := []byte{}

	if s.BlockHeight < config.Param().ConsensusParam.ByzantineDetectorHeight {
		data = append(data, s.BlockHash...)
		data = append(data, s.BLS...)
		data = append(data, s.BRI...)
	} else {
		data = append(data, s.BlockHash...)
		data = append(data, s.BLS...)
		data = append(data, s.BRI...)
		data = append(data, common.Uint64ToBytes(s.BlockHeight)...)
		data = append(data, common.Int64ToBytes(s.ProduceTimeSlot)...)
		data = append(data, common.Int64ToBytes(s.ProposeTimeSlot)...)
		data = append(data, []byte(s.Validator)...)
		data = append(data, []byte(s.PrevBlockHash)...)
		data = append(data, s.CommitteeFromBlock[:]...)
		data = append(data, common.Int64ToBytes(int64(s.ChainID))...)
	}

	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, s.Confirmation, ownerPk)
	return err
}
