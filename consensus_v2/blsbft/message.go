package blsbft

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/wire"
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
	Bls           []byte
	Bri           []byte
	Confirmation  []byte
	IsValid       int // 0 not process, 1 valid, -1 not valid
	TimeSlot      uint64
}

type BFTRequestBlock struct {
	BlockHash string
	PeerID    string
}

func (s *BFTVote) signVote(key *signatureschemes2.MiningKey) error {
	data := []byte{}
	data = append(data, s.BlockHash...)
	data = append(data, s.Bls...)
	data = append(data, s.Bri...)
	data = common.HashB(data)
	var err error
	s.Confirmation, err = key.BriSignData(data)
	return err
}

func (s *BFTVote) validateVoteOwner(ownerPk []byte) error {
	data := []byte{}
	data = append(data, s.BlockHash...)
	data = append(data, s.Bls...)
	data = append(data, s.Bri...)
	dataHash := common.HashH(data)
	err := validateSingleBriSig(&dataHash, s.Confirmation, ownerPk)
	return err
}

func (actorV1 *actorV1) ProcessBFTMsg(msgBFT *wire.MessageBFT) {
	switch msgBFT.Type {
	case MSG_PROPOSE:
		var msgPropose BFTPropose
		err := json.Unmarshal(msgBFT.Content, &msgPropose)
		if err != nil {
			actorV1.logger.Error(err)
			return
		}
		msgPropose.PeerID = msgBFT.PeerID
		actorV1.proposeMessageCh <- msgPropose
	case MSG_VOTE:
		var msgVote BFTVote
		err := json.Unmarshal(msgBFT.Content, &msgVote)
		if err != nil {
			actorV1.logger.Error(err)
			return
		}
		actorV1.voteMessageCh <- msgVote
	default:
		actorV1.logger.Critical("Unknown BFT message type")
		return
	}
}
