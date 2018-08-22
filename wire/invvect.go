package wire

import "github.com/ninjadotorg/cash-prototype/common"

// InvType represents the allowed types of inventory vectors.  See InvVect.
type InvType uint32

// These constants define the various supported inventory vector types.
const (
	InvTypeError InvType = 0
	InvTypeTx    InvType = 1
	InvTypeBlock InvType = 2
	//InvTypeFilteredBlock InvType = 3
	//InvTypeWitnessBlock         InvType = InvTypeBlock | InvWitnessFlag
	//InvTypeWitnessTx            InvType = InvTypeTx | InvWitnessFlag
	//InvTypeFilteredWitnessBlock InvType = InvTypeFilteredBlock | InvWitnessFlag
)

// InvVect defines a bitcoin inventory vector which is used to describe data,
// as specified by the Type field, that a peer wants, has, or does not have to
// another peer.
type InvVect struct {
	Type InvType     // Type of data
	Hash common.Hash // Hash of the data
}

// NewInvVect returns a new InvVect using the provided type and hash.
func NewInvVect(typ InvType, hash *common.Hash) *InvVect {
	return &InvVect{
		Type: typ,
		Hash: *hash,
	}
}
