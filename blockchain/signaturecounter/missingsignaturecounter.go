package signaturecounter

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"log"
	"reflect"
	"sync"

	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type Penalty struct {
	MinPercent   uint
	Time         int64
	ForceUnstake bool
}

type MissingSignature struct {
	ActualTotal uint
	Missing     uint
}

func NewMissingSignature() MissingSignature {
	return MissingSignature{
		ActualTotal: 0,
		Missing:     0,
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
	AddPreviousMissignSignature(prevValidationData string, committees []incognitokey.CommitteePublicKey) error
	GetAllSlashingPenaltyWithActualTotalBlock() map[string]Penalty
	GetAllSlashingPenaltyWithExpectedTotalBlock(map[string]uint) map[string]Penalty
	Reset(committees []string)
	CommitteeChange(committees []string)
	Copy() IMissingSignatureCounter
}

type MissingSignatureCounter struct {
	missingSignature map[string]MissingSignature
	penalties        []Penalty
	indexData        [][]int
	lock             *sync.RWMutex
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
		if idx >= len(tempToBeSignedCommittees) {
			return fmt.Errorf("Idx = %+v, greater than len(toBeSignedCommittees) = %+v", idx, len(tempToBeSignedCommittees))
		}
		signedCommittees[tempToBeSignedCommittees[idx]] = struct{}{}
	}
	for _, toBeSignedCommittee := range tempToBeSignedCommittees {
		missingSignature, ok := s.missingSignature[toBeSignedCommittee]
		if !ok {
			// skip toBeSignedCommittee not in current list
			continue
		}
		missingSignature.ActualTotal++

		if _, ok := signedCommittees[toBeSignedCommittee]; !ok {
			missingSignature.Missing++
		}
		s.missingSignature[toBeSignedCommittee] = missingSignature
	}

	s.indexData = append(s.indexData, validationData.ValidatiorsIdx)
	return nil
}

func (s *MissingSignatureCounter) AddPreviousMissignSignature(data string, toBeSignedCommittees []incognitokey.CommitteePublicKey) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.indexData) == 0 {
		log.Println("cache data is empty")
		return nil
	}

	prevValidatorIndexCache := s.indexData[len(s.indexData)-1]

	prevValidationData, err := consensustypes.DecodeValidationData(data)
	if err != nil {
		return err
	}
	tempToBeSignedCommittees, _ := incognitokey.CommitteeKeyListToString(toBeSignedCommittees)
	uncountCommittees := make(map[string]struct{})

	if len(prevValidationData.ValidatiorsIdx) <= len(prevValidatorIndexCache) {
		return nil
	}

	//find index that is not count in previous validator index
	for _, idx := range prevValidationData.ValidatiorsIdx {
		if idx >= len(tempToBeSignedCommittees) {
			return fmt.Errorf("Idx = %+v, greater than len(toBeSignedCommittees) = %+v", idx, len(tempToBeSignedCommittees))
		}
		if common.IndexOfInt(idx, prevValidatorIndexCache) == -1 {
			uncountCommittees[tempToBeSignedCommittees[idx]] = struct{}{}
		}
	}

	//revert missing counter
	for _, toBeSignedCommittee := range tempToBeSignedCommittees {
		missingSignature, ok := s.missingSignature[toBeSignedCommittee]
		if !ok {
			// skip toBeSignedCommittee not in current list
			continue
		}
		if _, ok := uncountCommittees[toBeSignedCommittee]; ok {
			if missingSignature.Missing > 0 {
				log.Println("sig counter: add sig counter for ", toBeSignedCommittee)
				missingSignature.Missing--
			}
		}

		s.missingSignature[toBeSignedCommittee] = missingSignature
	}
	s.indexData[len(s.indexData)-1] = prevValidationData.ValidatiorsIdx
	return nil
}

func (s MissingSignatureCounter) GetAllSlashingPenaltyWithActualTotalBlock() map[string]Penalty {
	s.lock.RLock()
	defer s.lock.RUnlock()

	penalties := make(map[string]Penalty)
	for key, numberOfMissingSig := range s.missingSignature {
		penalty := getSlashingPenalty(numberOfMissingSig.Missing, numberOfMissingSig.ActualTotal, s.penalties)
		if !penalty.IsEmpty() {
			penalties[key] = penalty
		}
	}
	return penalties
}

func (s MissingSignatureCounter) GetAllSlashingPenaltyWithExpectedTotalBlock(expectedTotalBlocks map[string]uint) map[string]Penalty {
	s.lock.RLock()
	defer s.lock.RUnlock()

	penalties := make(map[string]Penalty)
	for key, expectedTotalBlock := range expectedTotalBlocks {
		var penalty Penalty
		missingSignature, ok := s.missingSignature[key]
		if !ok {
			penalty = getSlashingPenalty(expectedTotalBlock, expectedTotalBlock, s.penalties)
		} else {
			signedBlock := missingSignature.ActualTotal - missingSignature.Missing
			missingBlock := uint(0)
			if signedBlock > expectedTotalBlock {
				missingBlock = 0
			} else {
				missingBlock = expectedTotalBlock - signedBlock
			}
			penalty = getSlashingPenalty(missingBlock, expectedTotalBlock, s.penalties)
		}
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

func (s *MissingSignatureCounter) Reset(committees []string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	missingSignature := make(map[string]MissingSignature)
	for _, v := range committees {
		missingSignature[v] = NewMissingSignature()
	}

	s.missingSignature = missingSignature
}

func (s *MissingSignatureCounter) CommitteeChange(newCommittees []string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	missingSignature := make(map[string]MissingSignature)
	for _, v := range newCommittees {
		res, ok := s.missingSignature[v]
		if !ok {
			missingSignature[v] = NewMissingSignature()
		} else {
			missingSignature[v] = res
		}
	}

	s.missingSignature = missingSignature
}

func (s *MissingSignatureCounter) Copy() IMissingSignatureCounter {
	s.lock.RLock()
	defer s.lock.RUnlock()

	newS := &MissingSignatureCounter{
		missingSignature: make(map[string]MissingSignature),
		penalties:        make([]Penalty, len(s.penalties)),
		indexData:        make([][]int, len(s.indexData)),
		lock:             new(sync.RWMutex),
	}
	copy(newS.penalties, s.penalties)
	copy(newS.indexData, s.indexData)

	for k, v := range s.missingSignature {
		newS.missingSignature[k] = v
	}
	return newS
}
