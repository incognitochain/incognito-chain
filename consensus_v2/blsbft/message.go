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
	PeerID   string
	Block    json.RawMessage
	TimeSlot uint64
}

type BFTVote struct {
	RoundKey      string
	PrevBlockHash string
	BlockHash     string
	Validator     string
	BLS           []byte
	BRI           []byte
	Confirmation  []byte
	IsValid       int // 0 not process, 1 valid, -1 not valid
	TimeSlot      uint64

	// Portal v4
	PortalSigs []*portalprocessv4.PortalSig
}

type BFTRequestBlock struct {
	BlockHash string
	PeerID    string
}

func (s *BFTVote) signVote(key *signatureschemes2.MiningKey) error {

	data := []byte{}
	portalSigsBytes, _ := json.Marshal(s.PortalSigs)

	data = append(data, s.BlockHash...)
	data = append(data, s.BLS...)
	data = append(data, s.BRI...)
	data = append(data, []byte(s.Validator)...)
	data = append(data, portalSigsBytes...)

	data = common.HashB(data)
	var err error
	s.Confirmation, err = key.BriSignData(data)
	return err
}

// TODO: @hung validate vote sender with vote owner to avoid false vote, use this schema to detect byzantine node
func (s *BFTVote) validateVoteOwner(ownerPk []byte) error {

	data := []byte{}
	portalSigsBytes, _ := json.Marshal(s.PortalSigs)

	data = append(data, s.BlockHash...)
	data = append(data, s.BLS...)
	data = append(data, s.BRI...)
	data = append(data, []byte(s.Validator)...)
	data = append(data, portalSigsBytes...)

	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, s.Confirmation, ownerPk)
	return err
}
