package signaturecounter

import (
	"github.com/incognitochain/incognito-chain/consensus/consensustypes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"reflect"
	"sort"
	"sync"
)

type Penalty struct {
	MinRange     uint
	Time         int64
	ForceUnstake bool
}

var defaultRule = []Penalty{
	{
		MinRange:     800,
		Time:         302400,
		ForceUnstake: false,
	},
	{
		MinRange:     1500,
		Time:         302400 * 2,
		ForceUnstake: false,
	},
	{
		MinRange:     3000,
		Time:         302400 * 2,
		ForceUnstake: true,
	},
}

func NewPenalty() Penalty {
	return Penalty{}
}

func (p Penalty) IsEmpty() bool {
	return reflect.DeepEqual(p, NewPenalty())
}

type MissingSignatureCounter interface {
	MissingSignature() map[string]uint
	Penalties() []Penalty
	AddMissingSignature(validationData string, committees []incognitokey.CommitteePublicKey) error
	GetAllSlashingPenalty() map[string]Penalty
	GetSlashingPenalty(key *incognitokey.CommitteePublicKey) (bool, Penalty, error)
	Reset()
	Copy() MissingSignatureCounter
}

type SignatureCounter struct {
	missingSignature map[string]uint
	penalties        []Penalty

	lock *sync.RWMutex
}

func (s *SignatureCounter) Penalties() []Penalty {
	return s.penalties
}

func (s *SignatureCounter) MissingSignature() map[string]uint {
	s.lock.RLock()
	defer s.lock.RUnlock()
	missingSignature := make(map[string]uint)
	for k, v := range s.missingSignature {
		missingSignature[k] = v
	}
	return missingSignature
}

func NewDefaultSignatureCounter() *SignatureCounter {
	return &SignatureCounter{
		missingSignature: make(map[string]uint),
		penalties:        defaultRule,
		lock:             new(sync.RWMutex),
	}
}

func NewSignatureCounterWithValue(missingSignature map[string]uint, rule []Penalty) *SignatureCounter {
	sort.Slice(rule, func(i, j int) bool {
		return rule[i].MinRange < rule[j].MinRange
	})
	return &SignatureCounter{
		missingSignature: missingSignature,
		penalties:        rule,
		lock:             new(sync.RWMutex),
	}
}

func (s *SignatureCounter) AddMissingSignature(data string, committees []incognitokey.CommitteePublicKey) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	validationData, err := consensustypes.DecodeValidationData(data)
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
	s.lock.RLock()
	defer s.lock.RUnlock()

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
		if numberOfMissingSig >= penaltyLevel.MinRange {
			penalty = penaltyLevel
		}
	}
	return penalty
}

func (s SignatureCounter) GetSlashingPenalty(key *incognitokey.CommitteePublicKey) (bool, Penalty, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

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
	s.lock.Lock()
	defer s.lock.Unlock()

	s.missingSignature = make(map[string]uint)
}

func (s *SignatureCounter) Copy() MissingSignatureCounter {
	s.lock.RLock()
	defer s.lock.RUnlock()

	newS := &SignatureCounter{
		missingSignature: make(map[string]uint),
		penalties:        make([]Penalty, len(s.penalties)),
		lock:             new(sync.RWMutex),
	}
	copy(newS.penalties, s.penalties)
	for k, v := range s.missingSignature {
		newS.missingSignature[k] = v
	}
	return newS
}
