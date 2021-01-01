package instruction

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

var (
	ErrAssignSyncInstruction = errors.New("Assign sync instruction error")
)

//AssignSyncInstruction assign validators from candidates list to syncing pool
// format: ["assign_sync_action", publickeys, chain_id]
type AssignSyncInstruction struct {
	ChainID               int
	ShardCandidates       []string
	ShardCandidatesStruct []incognitokey.CommitteePublicKey
}

func NewAssignSyncInstructionWithValue(chainID int, shardCandidates []string) *AssignSyncInstruction {
	assignInstruction := &AssignSyncInstruction{}
	assignInstruction.SetChainID(chainID)
	assignInstruction.SetShardCandidates(shardCandidates)
	return assignInstruction
}

func NewAssignSyncInstruction() *AssignSyncInstruction {
	return &AssignSyncInstruction{}
}

func (a *AssignSyncInstruction) GetType() string {
	return ASSIGN_SYNC_ACTION
}

func (a *AssignSyncInstruction) IsEmpty() bool {
	return reflect.DeepEqual(a, NewAssignSyncInstruction()) ||
		len(a.ShardCandidates) == 0 && len(a.ShardCandidatesStruct) == 0
}

func (a *AssignSyncInstruction) ToString() []string {
	assignSyncInstructionStr := []string{ASSIGN_SYNC_ACTION}
	assignSyncInstructionStr = append(assignSyncInstructionStr, fmt.Sprintf("%v", a.ChainID))
	assignSyncInstructionStr = append(assignSyncInstructionStr, strings.Join(a.ShardCandidates, SPLITTER))
	return assignSyncInstructionStr
}

func (a *AssignSyncInstruction) SetChainID(chainID int) *AssignSyncInstruction {
	a.ChainID = chainID
	return a
}

func (a *AssignSyncInstruction) SetShardCandidates(shardCandidates []string) *AssignSyncInstruction {
	a.ShardCandidates = shardCandidates
	return a
}

func ValidateAndImportAssignSyncInstructionFromString(instruction []string) (*AssignSyncInstruction, error) {
	if err := ValidateAssignSyncInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportAssignSyncInstructionFromString(instruction)
}

// ImportAssignSyncInstructionFromString is unsafe method
func ImportAssignSyncInstructionFromString(instruction []string) (*AssignSyncInstruction, error) {
	assignSyncIntruction := NewAssignSyncInstruction()
	tempShardID := instruction[1]
	chainID, _ := strconv.Atoi(tempShardID)
	assignSyncIntruction.SetChainID(chainID)
	assignSyncIntruction.SetShardCandidates(strings.Split(instruction[2], SPLITTER))

	shardPendingValidatorStruct, err := incognitokey.CommitteeBase58KeyListToStruct(assignSyncIntruction.ShardCandidates)
	if err != nil {
		return nil, err
	}
	assignSyncIntruction.ShardCandidatesStruct = shardPendingValidatorStruct

	return assignSyncIntruction, err
}

//ValidateAssignSyncInstructionSanity ...
func ValidateAssignSyncInstructionSanity(instruction []string) error {
	if len(instruction) != 3 {
		return fmt.Errorf("%+v: invalid length, %+v", ErrAssignInstruction, instruction)
	}
	if instruction[0] != ASSIGN_SYNC_ACTION {
		return fmt.Errorf("%+v: invalid assign action, %+v", ErrAssignSyncInstruction, instruction)
	}
	shardID, err := strconv.Atoi(instruction[2])
	if err != nil {
		return fmt.Errorf("%+v: Can not parse shardid int type, %+v", ErrAssignSyncInstruction, instruction)
	}
	if shardID < -1 || shardID > 8 {
		return fmt.Errorf("%+v: invalid assign chain ID, %+v", ErrAssignSyncInstruction, instruction)
	}
	if _, err := strconv.Atoi(instruction[2]); err != nil {
		return fmt.Errorf("%+v: invalid assign shard ID, err %+v, %+v", ErrAssignSyncInstruction, err, instruction)
	}
	return nil
}
