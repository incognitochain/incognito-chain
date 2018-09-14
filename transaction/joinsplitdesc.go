package transaction

import (
	"github.com/ninjadotorg/cash-prototype/privacy/client"
	"github.com/ninjadotorg/cash-prototype/privacy/proto/zksnark"
)

// JoinSplitDesc stores the UTXO of a transaction
// TODO(@0xbunyip): add randomSeed, MACs and epk
type JoinSplitDesc struct {
	Anchor          []byte             `json:"Anchor"`
	Nullifiers      [][]byte           `json:"Nullifiers"`
	Commitments     [][]byte           `json:"Commitments"`
	Proof           *zksnark.PHGRProof `json:"Proof"`
	EncryptedData   [][]byte           `json:"EncryptedData"`
	EphemeralPubKey []byte             `json:"EphemeralPubKey"`
	HSigSeed        []byte             `json:"HSigSeed"`
	Type            string             `json:"Type"`   // unit type (coin or bond) which used in tx
	Reward          uint64             `json:"Reward"` // For coinbase tx
	Vmacs           [][]byte

	note []*client.Note // decrypt data for EncryptedData
}

func (desc *JoinSplitDesc) toString() string {
	s := string(desc.Anchor)
	for _, nf := range desc.Nullifiers {
		s += string(nf)
	}
	for _, cm := range desc.Commitments {
		s += string(cm)
	}
	s += desc.Proof.String()
	for _, data := range desc.EncryptedData {
		s += string(data)
	}
	return s
}

func (self *JoinSplitDesc) AppendNote(note *client.Note) {
	self.note = append(self.note, note)
}

func (self *JoinSplitDesc) GetNote() []*client.Note {
	return self.note
}
