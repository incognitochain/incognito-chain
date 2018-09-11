package transaction

import (
	"github.com/ninjadotorg/cash-prototype/privacy/proto/zksnark"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
)

// JoinSplitDesc stores the UTXO of a transaction
// TODO(@0xbunyip): add randomSeed, MACs and epk
type JoinSplitDesc struct {
	Anchor        	[]byte             `json:"Anchor"`
	Nullifiers    	[][]byte           `json:"Nullifiers"`
	Commitments   	[][]byte           `json:"Commitments"`
	Proof         	*zksnark.PHGRProof `json:"Proof"`
	EncryptedData 	[][]byte           `json:"EncryptedData"`
	EphemeralPubKey []byte             `json:"EphemeralPubKey"`
	Type          	string             `json:"Type"`
	Reward        	uint64             `json:"Reward"` // For coinbase tx

	note *client.Note
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

func (self *JoinSplitDesc) SetNote(note *client.Note) {
	self.note = note
}

func (self *JoinSplitDesc) GetNote() *client.Note {
	return self.note
}
