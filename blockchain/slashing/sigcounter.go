package slashing

import (
	"github.com/incognitochain/incognito-chain/consensus/blsbft"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type SignatureCounter struct {
	epoch            int
	missingSignature map[string]uint
	rule             map[int]int
}

func NewSignatureCounter() *SignatureCounter {
	return &SignatureCounter{
		missingSignature: make(map[string]uint),
		rule:             make(map[int]int),
	}
}

func NewSignatureCounterWithValue(epoch int, missingSignature map[string]uint, rule map[int]int) *SignatureCounter {
	return &SignatureCounter{epoch: epoch, missingSignature: missingSignature, rule: rule}
}

func (s SignatureCounter) AddMissingSignature(data string, committees []incognitokey.CommitteePublicKey) error {
	validationData, err := blsbft.DecodeValidationData(data)
	if err != nil {
		return err
	}
	tempCommittees, _ := incognitokey.CommitteeKeyListToString(committees)
	signedCommittees := make(map[string]struct{})
	for _, idx := range validationData.ValidatiorsIdx {
		signedCommittees[tempCommittees[idx]] = struct{}{}
	}
	for _, committee := range tempCommittees {
		if _, ok := signedCommittees[committee]; !ok {
			s.missingSignature[committee] += 1
		}
	}
	return nil
}

func (s SignatureCounter) GetAllSlashingPenalty() map[string]uint64 {
	panic("implement me")
}

func (s SignatureCounter) GetSlashingPenalty(key *incognitokey.CommitteePublicKey) {
	panic("implement me")
}

func (s SignatureCounter) Reset() {
	s.missingSignature = make(map[string]uint)
}
