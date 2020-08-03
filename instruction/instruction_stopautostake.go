package instruction

import (
	"fmt"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
)

type StopAutoStakeInstruction struct {
	PublicKeys []string
}

func NewStopAutoStakeInstructionWithValue(publicKeys []string) *StopAutoStakeInstruction {
	return &StopAutoStakeInstruction{PublicKeys: publicKeys}
}

func NewStopAutoStakeInstruction() *StopAutoStakeInstruction {
	return &StopAutoStakeInstruction{}
}

func (s *StopAutoStakeInstruction) GetType() string {
	return STOP_AUTO_STAKE_ACTION
}

func (s *StopAutoStakeInstruction) ToString() []string {
	stopAutoStakeInstructionStr := []string{STOP_AUTO_STAKE_ACTION}
	stopAutoStakeInstructionStr = append(stopAutoStakeInstructionStr, strings.Join(s.PublicKeys, SPLITTER))
	return stopAutoStakeInstructionStr
}

func ValidateAndImportStopAutoStakeInstructionFromString(instruction []string) (*StopAutoStakeInstruction, error) {
	if err := ValidateStopAutoStakeInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportStopAutoStakeInstructionFromString(instruction), nil
}

func ImportStopAutoStakeInstructionFromString(instruction []string) *StopAutoStakeInstruction {
	stopAutoStakeInstruction := NewStopAutoStakeInstruction()
	if len(instruction[1]) > 0 {
		publicKeys := strings.Split(instruction[1], SPLITTER)
		stopAutoStakeInstruction.PublicKeys = publicKeys
	}
	return stopAutoStakeInstruction
}

func ValidateStopAutoStakeInstructionSanity(instruction []string) error {
	if len(instruction) != 2 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != STOP_AUTO_STAKE_ACTION {
		return fmt.Errorf("invalid stop auto stake action, %+v", instruction)
	}
	return nil
}

func (saI *StopAutoStakeInstruction) InsertIntoStateDB(sDB *statedb.StateDB) error {
	pkStructs, err := incognitokey.CommitteeBase58KeyListToStruct(saI.PublicKeys)
	if err != nil {
		return err
	}
	// TODO:
	// Instead of preprocessing for create input for storeStakerInfo, create function update stakerinfo
	asMap := map[string]bool{}
	for _, pk := range saI.PublicKeys {
		asMap[pk] = true
	}
	return statedb.StoreStakerInfo(
		sDB,
		pkStructs,
		map[string]privacy.PaymentAddress{}, //Empty map cuz we just update auto staking flag
		asMap,
		map[string]common.Hash{},
	)
}
