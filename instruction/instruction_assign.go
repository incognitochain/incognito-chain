package instruction

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

var (
	ErrAssignInstruction = errors.New("assign instruction error")
)

type AssignInstruction struct {
	ChainID         int
	ShardCandidates []string
}

func NewAssignInstructionWithValue(chainID int, shardCandidates []string) *AssignInstruction {
	return &AssignInstruction{ChainID: chainID, ShardCandidates: shardCandidates}
}

func NewAssignInstruction() *AssignInstruction {
	return &AssignInstruction{}
}

func (a *AssignInstruction) GetType() string {
	return ASSIGN_ACTION
}

func (a *AssignInstruction) ToString() []string {
	assignInstructionStr := []string{ASSIGN_ACTION}
	assignInstructionStr = append(assignInstructionStr, strings.Join(a.ShardCandidates, SPLITTER))
	assignInstructionStr = append(assignInstructionStr, "shard")
	assignInstructionStr = append(assignInstructionStr, fmt.Sprintf("%v", a.ChainID))
	return assignInstructionStr
}

func (a *AssignInstruction) SetChainID(chainID int) *AssignInstruction {
	a.ChainID = chainID
	return a
}

func (a *AssignInstruction) SetShardCandidates(shardCandidates []string) *AssignInstruction {
	a.ShardCandidates = shardCandidates
	return a
}

func ValidateAndImportAssignInstructionFromString(instruction []string) (*AssignInstruction, error) {
	if err := ValidateAssignInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportAssignInstructionFromString(instruction), nil
}

// ImportAssignInstructionFromString is unsafe method
func ImportAssignInstructionFromString(instruction []string) *AssignInstruction {
	assignIntruction := NewAssignInstruction()
	tempShardID := instruction[3]
	chainID, _ := strconv.Atoi(tempShardID)
	assignIntruction.SetChainID(chainID)
	assignIntruction.SetShardCandidates(strings.Split(instruction[1], SPLITTER))
	return assignIntruction
}

//ValidateAssignInstructionSanity ...
func ValidateAssignInstructionSanity(instruction []string) error {
	if len(instruction) != 4 {
		return fmt.Errorf("%+v: invalid length, %+v", ErrAssignInstruction, instruction)
	}
	if instruction[0] != ASSIGN_ACTION {
		return fmt.Errorf("%+v: invalid assign action, %+v", ErrAssignInstruction, instruction)
	}
	if instruction[2] != SHARD_INST {
		return fmt.Errorf("%+v: invalid assign chain ID, %+v", ErrAssignInstruction, instruction)
	}
	if _, err := strconv.Atoi(instruction[3]); err != nil {
		return fmt.Errorf("%+v: invalid assign shard ID, err %+v, %+v", ErrAssignInstruction, err, instruction)
	}
	return nil
}

func (aI *AssignInstruction) InsertIntoStateDB(sDB *statedb.StateDB) error {
	candidates, err := incognitokey.CommitteeBase58KeyListToStruct(aI.ShardCandidates)
	if err != nil {
		return err
	}
	if aI.ChainID == BEACON_CHAIN_ID {
		err = statedb.StoreBeaconSubstituteValidator(sDB, candidates)
		if err != nil {
			return err
		}
	}
	err = statedb.StoreOneShardSubstitutesValidator(sDB, byte(aI.ChainID), candidates)
	if err != nil {
		return err
	}
	return nil
}
