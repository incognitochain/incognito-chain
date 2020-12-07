package signaturecounter

import (
	"reflect"
	"sync"

	"github.com/incognitochain/incognito-chain/consensus/consensustypes"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type Penalty struct {
	MinPercent   uint
	Time         int64
	ForceUnstake bool
}

type MissingSignature struct {
	Total   uint
	Missing uint
}

func NewMissingSignature() MissingSignature {
	return MissingSignature{
		Total:   0,
		Missing: 0,
	}
}

var defaultRule = []Penalty{
	{
		MinPercent:   50,
		Time:         0,
		ForceUnstake: true,
	},
}

func NewPenalty() Penalty {
	return Penalty{}
}

func (p Penalty) IsEmpty() bool {
	return reflect.DeepEqual(p, NewPenalty())
}

type IMissingSignatureCounter interface {
	MissingSignature() map[string]MissingSignature
	Penalties() []Penalty
	AddMissingSignature(validationData string, committees []incognitokey.CommitteePublicKey) error
	GetAllSlashingPenalty() map[string]Penalty
	GetSlashingPenalty(key *incognitokey.CommitteePublicKey) (bool, Penalty, error)
	Reset(committees []string)
	Copy() IMissingSignatureCounter
}

type MissingSignatureCounter struct {
	missingSignature map[string]MissingSignature
	penalties        []Penalty

	lock *sync.RWMutex
}

func (s *MissingSignatureCounter) Penalties() []Penalty {
	return s.penalties
}

func (s *MissingSignatureCounter) MissingSignature() map[string]MissingSignature {
	s.lock.RLock()
	defer s.lock.RUnlock()
	missingSignature := make(map[string]MissingSignature)
	for k, v := range s.missingSignature {
		missingSignature[k] = v
	}
	return missingSignature
}

func NewDefaultSignatureCounter(committees []string) *MissingSignatureCounter {
	missingSignature := make(map[string]MissingSignature)
	for _, v := range committees {
		missingSignature[v] = NewMissingSignature()
	}
	return &MissingSignatureCounter{
		missingSignature: missingSignature,
		penalties:        defaultRule,
		lock:             new(sync.RWMutex),
	}
}

func (s *MissingSignatureCounter) AddMissingSignature(data string, toBeSignedCommittees []incognitokey.CommitteePublicKey) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	validationData, err := consensustypes.DecodeValidationData(data)
	if err != nil {
		return err
	}
	tempToBeSignedCommittees, _ := incognitokey.CommitteeKeyListToString(toBeSignedCommittees)
	signedCommittees := make(map[string]struct{})
	for _, idx := range validationData.ValidatiorsIdx {
		signedCommittees[tempToBeSignedCommittees[idx]] = struct{}{}
	}
	for _, toBeSignedCommittee := range tempToBeSignedCommittees {
		missingSignature, ok := s.missingSignature[toBeSignedCommittee]
		if !ok {
			// skip toBeSignedCommittee not in current list
			continue
		}
		missingSignature.Total++
		if _, ok := signedCommittees[toBeSignedCommittee]; !ok {
			missingSignature.Missing++
		}
		s.missingSignature[toBeSignedCommittee] = missingSignature
	}
	return nil
}

func (s MissingSignatureCounter) GetAllSlashingPenalty() map[string]Penalty {
	s.lock.RLock()
	defer s.lock.RUnlock()

	penalties := make(map[string]Penalty)
	for key, numberOfMissingSig := range s.missingSignature {
		penalty := getSlashingPenalty(numberOfMissingSig.Missing, numberOfMissingSig.Total, s.penalties)
		if !penalty.IsEmpty() {
			penalties[key] = penalty
		}
	}
	return penalties
}

func getSlashingPenalty(numberOfMissingSig uint, total uint, penalties []Penalty) Penalty {
	penalty := NewPenalty()
	if total == 0 {
		return penalty
	}
	missedPercent := numberOfMissingSig * 100 / total
	for _, penaltyLevel := range penalties {
		if missedPercent >= penaltyLevel.MinPercent {
			penalty = penaltyLevel
		}
	}
	return penalty
}

func (s MissingSignatureCounter) GetSlashingPenalty(key *incognitokey.CommitteePublicKey) (bool, Penalty, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	tempKey, err := key.ToBase58()
	if err != nil {
		return false, NewPenalty(), err
	}
	missingSignature, ok := s.missingSignature[tempKey]
	if !ok {
		return false, NewPenalty(), nil
	}
	penalty := getSlashingPenalty(missingSignature.Missing, missingSignature.Total, s.penalties)
	if penalty.IsEmpty() {
		return false, NewPenalty(), nil
	}
	return true, penalty, nil
}

func (s *MissingSignatureCounter) Reset(committees []string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	missingSignature := make(map[string]MissingSignature)
	for _, v := range committees {
		missingSignature[v] = NewMissingSignature()
	}

	s.missingSignature = missingSignature
}

func (s *MissingSignatureCounter) Copy() IMissingSignatureCounter {
	s.lock.RLock()
	defer s.lock.RUnlock()

	newS := &MissingSignatureCounter{
		missingSignature: make(map[string]MissingSignature),
		penalties:        make([]Penalty, len(s.penalties)),
		lock:             new(sync.RWMutex),
	}
	copy(newS.penalties, s.penalties)
	for k, v := range s.missingSignature {
		newS.missingSignature[k] = v
	}
	return newS
}
