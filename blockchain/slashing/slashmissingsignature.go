package slashing

import (
	"github.com/incognitochain/incognito-chain/consensus/blsbft"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"reflect"
	"sort"
)

type Penalty struct {
	minRange     uint
	time         int64
	forceUnstake bool
}

func NewPenalty() Penalty {
	return Penalty{}
}

func (p Penalty) MinMissingSignature() uint {
	return p.minRange
}

func (p Penalty) Time() int64 {
	return p.time
}

func (p Penalty) ForceUnstake() bool {
	return p.forceUnstake
}

func (p Penalty) IsEmpty() bool {
	return reflect.DeepEqual(p, NewPenalty())
}

var defaultRule = []Penalty{
	{
		minRange:     800,
		time:         302400,
		forceUnstake: false,
	},
	{
		minRange:     1500,
		time:         302400 * 2,
		forceUnstake: false,
	},
	{
		minRange:     3000,
		time:         302400 * 2,
		forceUnstake: true,
	},
}

type SlashMissingSignature interface {
	AddMissingSignature(validationData string, committees []incognitokey.CommitteePublicKey) error
	GetAllSlashingPenalty() map[string]Penalty
	GetSlashingPenalty(key *incognitokey.CommitteePublicKey) (bool, Penalty, error)
	Reset()
}

type SignatureCounter struct {
	missingSignature map[string]uint
	penalties        []Penalty
}

func (s *SignatureCounter) Penalties() []Penalty {
	return s.penalties
}

func (s *SignatureCounter) SetPenalties(penalties []Penalty) {
	s.penalties = penalties
}

func (s *SignatureCounter) MissingSignature() map[string]uint {
	return s.missingSignature
}

func (s *SignatureCounter) SetMissingSignature(missingSignature map[string]uint) {
	s.missingSignature = missingSignature
}

func NewDefaultSignatureCounter() *SignatureCounter {
	return &SignatureCounter{
		missingSignature: make(map[string]uint),
		penalties:        defaultRule,
	}
}

func NewSignatureCounterWithValue(missingSignature map[string]uint, rule []Penalty) *SignatureCounter {
	sort.Slice(rule, func(i, j int) bool {
		return rule[i].minRange < rule[j].minRange
	})
	return &SignatureCounter{missingSignature: missingSignature, penalties: rule}
}

func NewSignatureCounterWithPenalties(penalties []Penalty) *SignatureCounter {
	return &SignatureCounter{penalties: penalties}
}

func (s *SignatureCounter) AddMissingSignature(data string, committees []incognitokey.CommitteePublicKey) error {
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

func (s SignatureCounter) GetAllSlashingPenalty() map[string]Penalty {
	penalties := make(map[string]Penalty)
	for key, numberOfMissingSig := range s.missingSignature {
		penalty := getSlashingPenalty(numberOfMissingSig, s.penalties)
		if !penalty.IsEmpty() {
			penalties[key] = penalty
		}
	}
	return penalties
}

func getSlashingPenalty(numberOfMissingSig uint, penalties []Penalty) Penalty {
	penalty := NewPenalty()
	for _, penaltyLevel := range penalties {
		if numberOfMissingSig > penalty.minRange {
			penalty = penaltyLevel
		}
	}
	return penalty
}

func (s SignatureCounter) GetSlashingPenalty(key *incognitokey.CommitteePublicKey) (bool, Penalty, error) {
	tempKey, err := key.ToBase58()
	if err != nil {
		return false, NewPenalty(), err
	}
	numberOfMissingSig, ok := s.missingSignature[tempKey]
	if !ok {
		return false, NewPenalty(), nil
	}
	penalty := getSlashingPenalty(numberOfMissingSig, s.penalties)
	if penalty.IsEmpty() {
		return false, NewPenalty(), nil
	}
	return true, penalty, nil
}

func (s *SignatureCounter) Reset() {
	s.missingSignature = make(map[string]uint)
}
