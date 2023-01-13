package instruction

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/log/proto"
	protolog "github.com/incognitochain/incognito-chain/log/proto"
)

var (
	ErrAssignBeaconInstruction = errors.New("assign instruction error")
)

const (
	COMMITTEE_POOL = iota
	PENDING_POOL
	WAITING_POOL
	LOCKING_POOL
)

var (
	poolName = map[int]string{
		COMMITTEE_POOL: "committee",
		PENDING_POOL:   "pending",
		WAITING_POOL:   "waiting",
		LOCKING_POOL:   "locking",
	}
	poolID = map[string]int{
		"committee": COMMITTEE_POOL,
		"pending":   PENDING_POOL,
		"waiting":   WAITING_POOL,
		"locking":   LOCKING_POOL,
	}
)

// AssignBeaconInstruction :
// Assign instruction format:
// ["assign action", publickeys, shard or beacon chain, shard_id]
type AssignBeaconInstruction struct {
	BeaconCandidates       string
	BeaconCandidatesStruct incognitokey.CommitteePublicKey
	FromPool               int
	ToPool                 int
	Reason                 int
	instructionBase
}

func NewAssignBeaconInstructionWithValue(from, to, reason int, candidates string) (*AssignBeaconInstruction, error) {
	assignInstruction := NewAssignBeaconInstruction()
	if _, ok := poolName[from]; !ok {
		return nil, fmt.Errorf("%+v: invalid from pool ID, %+v", ErrAssignBeaconInstruction, from)
	}
	if _, ok := poolName[to]; !ok {
		return nil, fmt.Errorf("%+v: invalid to pool ID, %+v", ErrAssignBeaconInstruction, to)
	}
	assignInstruction.SetFromTo(from, to)
	assignInstruction.SetBeaconCandidates(candidates)
	return assignInstruction, nil
}

func NewAssignBeaconInstruction() *AssignBeaconInstruction {
	return &AssignBeaconInstruction{
		instructionBase: instructionBase{
			featureID: proto.FID_CONSENSUS_SHARD,
			logOnly:   false,
		},
	}
}

func (a *AssignBeaconInstruction) GetType() string {
	return ASSIGN_BEACON_ACTION
}

func (a *AssignBeaconInstruction) JustLog() bool {
	return a.logOnly
}

func (a *AssignBeaconInstruction) IsEmpty() bool {
	return reflect.DeepEqual(a, NewAssignBeaconInstruction()) ||
		len(a.BeaconCandidates) == 0
}

func (a *AssignBeaconInstruction) ToString() []string {
	assignInstructionStr := []string{ASSIGN_BEACON_ACTION}
	assignInstructionStr = append(assignInstructionStr, a.BeaconCandidates)
	assignInstructionStr = append(assignInstructionStr, poolName[a.FromPool])
	assignInstructionStr = append(assignInstructionStr, poolName[a.ToPool])
	assignInstructionStr = append(assignInstructionStr, strconv.FormatInt(int64(a.Reason), 10))
	return assignInstructionStr
}

func (a *AssignBeaconInstruction) FromLog(fLog *protolog.FeatureLog) (err error) {
	rawData := string(fLog.Data)
	dataStr := strings.Split(rawData, ".")
	a, err = ImportAssignBeaconInstructionFromString(dataStr)
	return err
}

func (a *AssignBeaconInstruction) SetFromTo(from, to int) *AssignBeaconInstruction {
	a.FromPool = from
	a.ToPool = to
	return a
}

func (a *AssignBeaconInstruction) SetReason(reason int) *AssignBeaconInstruction {
	a.Reason = reason
	return a
}

func (a *AssignBeaconInstruction) SetBeaconCandidates(candidates string) *AssignBeaconInstruction {
	a.BeaconCandidates = candidates
	return a
}

func ValidateAndImportAssignBeaconInstructionFromString(instruction []string) (*AssignBeaconInstruction, error) {
	if err := ValidateAssignBeaconInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportAssignBeaconInstructionFromString(instruction)
}

// ImportAssignBeaconInstructionFromString is unsafe method
func ImportAssignBeaconInstructionFromString(instruction []string) (*AssignBeaconInstruction, error) {
	assignIntruction := NewAssignBeaconInstruction()
	assignIntruction.SetBeaconCandidates(instruction[1])
	fromCommittee := instruction[2]
	toCommittee := instruction[3]
	reasonStr := instruction[4]
	fromID := poolID[fromCommittee]
	toID := poolID[toCommittee]
	assignIntruction.SetFromTo(fromID, toID)
	reason, _ := strconv.ParseInt(reasonStr, 10, 64)
	assignIntruction.Reason = int(reason)
	fmt.Printf("%v\n", assignIntruction)
	beaconValidatorStruct, err := incognitokey.CommitteeBase58KeyListToStruct([]string{assignIntruction.BeaconCandidates})
	if err != nil {
		return nil, err
	}

	assignIntruction.BeaconCandidatesStruct = beaconValidatorStruct[0]
	fmt.Printf("%v\n", assignIntruction)
	return assignIntruction, err
}

// ValidateAssignBeaconInstructionSanity ...
func ValidateAssignBeaconInstructionSanity(instruction []string) error {
	if len(instruction) != 5 {
		return fmt.Errorf("%+v: invalid length, %+v", ErrAssignBeaconInstruction, instruction)
	}
	if instruction[0] != ASSIGN_BEACON_ACTION {
		return fmt.Errorf("%+v: invalid assign action, %+v", ErrAssignBeaconInstruction, instruction)
	}
	if _, ok := poolID[instruction[2]]; !ok {
		return fmt.Errorf("%+v: invalid from pool ID, %+v", ErrAssignBeaconInstruction, instruction)
	}
	if _, ok := poolID[instruction[3]]; !ok {
		return fmt.Errorf("%+v: invalid to pool ID, %+v", ErrAssignBeaconInstruction, instruction)
	}
	return nil
}
