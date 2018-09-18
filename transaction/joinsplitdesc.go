package transaction

import (
	"github.com/ninjadotorg/cash-prototype/privacy/client"
	"github.com/ninjadotorg/cash-prototype/privacy/proto/zksnark"
)

// JoinSplitDesc stores the UTXO of a transaction
type JoinSplitDesc struct {
	Anchor          []byte             `json:"Anchor"`          // 32 bytes
	Nullifiers      [][]byte           `json:"Nullifiers"`      // len == 2, 32 bytes each element
	Commitments     [][]byte           `json:"Commitments"`     // len == 2, 32 bytes each element
	Proof           *zksnark.PHGRProof `json:"Proof"`           // G_A, G_APrime, G_B, G_C, G_CPrime, G_K, G_H == 33 bytes each, G_BPrime 65 bytes
	EncryptedData   [][]byte           `json:"EncryptedData"`   // len == 2
	EphemeralPubKey []byte             `json:"EphemeralPubKey"` // 32 bytes
	HSigSeed        []byte             `json:"HSigSeed"`        // 32 bytes
	Type            string             `json:"Type"`            // unit type (coin or bond) which used in tx
	Reward          uint64             `json:"Reward"`          // For coinbase tx
	Vmacs           [][]byte           `json:"Vmacs"`           // len == 2, 32 bytes

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
