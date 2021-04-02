package blsbft

import (
	"encoding/json"
)

const (
	MsgPropose    = "propose"
	MsgVote       = "vote"
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
