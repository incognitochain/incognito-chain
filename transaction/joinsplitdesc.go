package transaction

import (
	"github.com/ninjadotorg/constant/privacy/client"
	"github.com/ninjadotorg/constant/privacy/proto/zksnark"
)

// JoinSplitDesc stores the UTXO of a transaction
type JoinSplitDesc struct {
	Anchor          [][]byte           `json:"Anchor"`          // len == 2, 32 bytes each
	Nullifiers      [][]byte           `json:"Nullifiers"`      // len == 2, 32 bytes each
	Commitments     [][]byte           `json:"Commitments"`     // len == 2, 32 bytes each
	Proof           *zksnark.PHGRProof `json:"Proof"`           // G_A, G_APrime, G_B, G_C, G_CPrime, G_K, G_H == 33 bytes each, G_BPrime 65 bytes
	EncryptedData   [][]byte           `json:"EncryptedData"`   // len == 2
	EphemeralPubKey []byte             `json:"EphemeralPubKey"` // 32 bytes
	HSigSeed        []byte             `json:"HSigSeed"`        // 32 bytes
	Type            string             `json:"Type"`            // asset type (constant coin or bond or d-token, g-token) which used in tx
	Reward          uint64             `json:"Reward"`          // For salary tx
	Vmacs           [][]byte           `json:"Vmacs"`           // len == 2, 32 bytes

	Note []*client.Note // decrypt data for EncryptedData
}

// EstimateJSDescSize returns the estimated size of a JoinSplitDesc in bytes
func EstimateJSDescSize() uint64 {
	var sizeAnchor uint64 = 32                                 // [32]byte
	var sizeNullifiers uint64 = 64                             // [2][32]byte
	var sizeCommitments uint64 = 64                            // [2][32]byte
	var sizeProof uint64 = 33*7 + 65                           // zksnark.PHGRProof
	var sizeEncryptedData uint64 = 2 * (8 + 32 + 32)           // [2][]byte, ignore memo
	var sizeEphemeralPubKey uint64 = client.EphemeralKeyLength // [32]byte
	var sizeHSigSeed uint64 = 32                               // [32]byte
	var sizeType uint64 = 8                                    // string
	var sizeReward uint64 = 8                                  // uint64
	var sizeVmacs uint64 = 64                                  // [2][32]byte
	return sizeAnchor + sizeNullifiers + sizeCommitments + sizeProof +
		sizeEncryptedData + sizeEphemeralPubKey + sizeHSigSeed + sizeType +
		sizeReward + sizeVmacs
}

func (desc *JoinSplitDesc) toString() string {
	var s string
	for _, anchor := range desc.Anchor {
		s += string(anchor)
	}
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
	self.Note = append(self.Note, note)
}

func (self *JoinSplitDesc) GetNote() []*client.Note {
	return self.Note
}
