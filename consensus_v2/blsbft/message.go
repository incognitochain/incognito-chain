package blsbft

import (
	"encoding/json"
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
}

type BFTVote struct {
	RoundKey      string
	PrevBlockHash string
	BlockHeight   uint64
	BlockHash     string
	Validator     string
	BLS           []byte
	BRI           []byte
	Confirmation  []byte
	IsValid       int // 0 not process, 1 valid, -1 not valid
	// TODO: @hung how to validate timeslot with block height anhd hash
	TimeSlot int64
	//TODO: @hung add committee from block
	CommitteeFromBlock common.Hash
	//TODO: @hung  add ChainID
	ChainID int
	// Portal v4
	PortalSigs []*portalprocessv4.PortalSig
}

type BFTRequestBlock struct {
	BlockHash string
	PeerID    string
}

func (s *BFTVote) signVote(key *signatureschemes2.MiningKey) error {
	data := []byte{}
	data = append(data, s.BlockHash...)
	data = append(data, s.BLS...)
	data = append(data, s.BRI...)
	data = append(data, common.Uint64ToBytes(s.BlockHeight)...)
	data = append(data, common.Int64ToBytes(s.TimeSlot)...)
	data = append(data, []byte(s.Validator)...)
	data = append(data, []byte(s.PrevBlockHash)...)
	data = common.HashB(data)
	var err error
	s.Confirmation, err = key.BriSignData(data)
	return err
}

func (s *BFTVote) validateVoteOwner(ownerPk []byte) error {
	data := []byte{}
	data = append(data, s.BlockHash...)
	data = append(data, s.BLS...)
	data = append(data, s.BRI...)
	data = append(data, common.Uint64ToBytes(s.BlockHeight)...)
	data = append(data, common.Int64ToBytes(s.TimeSlot)...)
	data = append(data, []byte(s.Validator)...)
	data = append(data, []byte(s.PrevBlockHash)...)
	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, s.Confirmation, ownerPk)
	return err
}
