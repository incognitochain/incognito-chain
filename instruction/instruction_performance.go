package instruction

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/log/proto"
)

type UpdatePerformanceInstruction struct {
	CommitteePublicKeys []string
	OldPerformance      []uint64
	NewPerformance      []uint64
	IsVoted             []bool
	instructionBase
}

func NewUpdatePerformanceInstructionWithValue(publicKeys []string, oldP, newP []uint64, voted []bool) *UpdatePerformanceInstruction {
	pksNew := []string{}
	for _, v := range publicKeys {
		pksNew = append(pksNew, v[len(v)-5:])
	}
	res := &UpdatePerformanceInstruction{
		CommitteePublicKeys: pksNew,
		OldPerformance:      oldP,
		NewPerformance:      newP,
		IsVoted:             voted,
		instructionBase: instructionBase{
			featureID: proto.FID_CONSENSUS_BEACON,
			logOnly:   true,
		},
	}
	return res
}

func NewUpdatePerformanceInstruction() *UpdatePerformanceInstruction {
	return &UpdatePerformanceInstruction{
		instructionBase: instructionBase{
			featureID: proto.FID_CONSENSUS_BEACON,
			logOnly:   true,
		},
	}
}

func (s *UpdatePerformanceInstruction) GetType() string {
	return UPDATE_PERF_ACTION
}

func (s *UpdatePerformanceInstruction) IsEmpty() bool {
	return reflect.DeepEqual(s, NewUpdatePerformanceInstruction()) || len(s.CommitteePublicKeys) == 0
}

func (s *UpdatePerformanceInstruction) SetPublicKeys(publicKeys []string) (*UpdatePerformanceInstruction, error) {
	s.CommitteePublicKeys = publicKeys

	return s, nil
}

func (s *UpdatePerformanceInstruction) SetPerf(oldP, newP []uint64) (*UpdatePerformanceInstruction, error) {
	s.OldPerformance = oldP
	s.NewPerformance = newP
	return s, nil
}

func (s *UpdatePerformanceInstruction) SetVoted(isVoted []bool) (*UpdatePerformanceInstruction, error) {
	s.IsVoted = isVoted
	return s, nil
}

func (s *UpdatePerformanceInstruction) ToString() []string {
	updatePerformanceInstructionStr := []string{UPDATE_PERF_ACTION}
	oldPerfStr := []string{}
	for _, v := range s.OldPerformance {
		str := strconv.FormatUint(v, 10)
		oldPerfStr = append(oldPerfStr, str)
	}
	newPerfStr := []string{}
	for _, v := range s.OldPerformance {
		str := strconv.FormatUint(v, 10)
		newPerfStr = append(newPerfStr, str)
	}
	isVotedStr := []string{}
	for _, v := range s.IsVoted {
		str := FALSE
		if v {
			str = TRUE
		}
		isVotedStr = append(isVotedStr, str)
	}
	updatePerformanceInstructionStr = append(updatePerformanceInstructionStr, strings.Join(s.CommitteePublicKeys, SPLITTER))
	updatePerformanceInstructionStr = append(updatePerformanceInstructionStr, strings.Join(oldPerfStr, SPLITTER))
	updatePerformanceInstructionStr = append(updatePerformanceInstructionStr, strings.Join(newPerfStr, SPLITTER))
	updatePerformanceInstructionStr = append(updatePerformanceInstructionStr, strings.Join(isVotedStr, SPLITTER))
	return updatePerformanceInstructionStr
}

func ValidateAndImportUpdatePerformanceInstructionFromString(instruction []string) (*UpdatePerformanceInstruction, error) {
	if err := ValidateUpdatePerformanceInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportUpdatePerformanceInstructionFromString(instruction), nil
}

func ImportUpdatePerformanceInstructionFromString(instruction []string) *UpdatePerformanceInstruction {
	updatePerfInstruction := NewUpdatePerformanceInstruction()
	if len(instruction[1]) > 0 {
		publicKeys := strings.Split(instruction[1], SPLITTER)
		updatePerfInstruction, _ = updatePerfInstruction.SetPublicKeys(publicKeys)
	}
	oldPerfStr := strings.Split(instruction[2], SPLITTER)
	oldPerf := make([]uint64, len(oldPerfStr))
	for i, v := range oldPerfStr {
		amount, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil
		}
		oldPerf[i] = amount
	}
	newPerfStr := strings.Split(instruction[3], SPLITTER)
	newPerf := make([]uint64, len(newPerfStr))
	for i, v := range newPerfStr {
		amount, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil
		}
		newPerf[i] = amount
	}
	updatePerfInstruction.SetPerf(oldPerf, newPerf)
	isVotedStr := strings.Split(instruction[4], SPLITTER)
	isVoted := make([]bool, len(isVotedStr))
	for i, v := range isVotedStr {
		voted := true
		if v == FALSE {
			voted = false
		}
		isVoted[i] = voted
	}
	updatePerfInstruction.SetVoted(isVoted)
	return updatePerfInstruction
}

func ValidateUpdatePerformanceInstructionSanity(instruction []string) error {
	if len(instruction) != 5 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != UPDATE_PERF_ACTION {
		return fmt.Errorf("invalid update performance action, %+v", instruction)
	}
	return nil
}
