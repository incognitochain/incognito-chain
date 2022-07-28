package instruction

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEnableFeatureInstruction = errors.New("enable feature instruction error")
)

//FeatureEnableInstruction :
// format: "finish_sync", "0", "key1,key2"
type EnableFeatureInstruction struct {
	Features []string
}

func NewEnableFeatureInstructionWithValue(feature []string) *EnableFeatureInstruction {
	featureEnableInstruction := &EnableFeatureInstruction{feature}
	return featureEnableInstruction
}

func NewEnableFeatureInstruction() *EnableFeatureInstruction {
	return &EnableFeatureInstruction{}
}

func (f *EnableFeatureInstruction) GetType() string {
	return ENABLE_FEATURE
}

func (f *EnableFeatureInstruction) IsEmpty() bool {
	return len(f.Features) == 0
}

func (f *EnableFeatureInstruction) ToString() []string {
	featureEnableInstruction := []string{ENABLE_FEATURE}
	featureEnableInstruction = append(featureEnableInstruction, strings.Join(f.Features, ","))
	return featureEnableInstruction
}

func ValidateAndImportEnableFeatureInstructionFromString(instruction []string) (*EnableFeatureInstruction, error) {
	if err := ValidateEnableFeatureInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportEnableFeatureInstructionFromString(instruction)
}

// ImportFeatureEnableInstructionFromString is unsafe method
func ImportEnableFeatureInstructionFromString(instruction []string) (*EnableFeatureInstruction, error) {
	featureEnableInstruction := NewEnableFeatureInstruction()
	featureName := instruction[1]
	featureEnableInstruction.Features = strings.Split(featureName, ",")
	return featureEnableInstruction, nil
}

//ValidateFeatureEnableInstructionSanity ...
func ValidateEnableFeatureInstructionSanity(instruction []string) error {
	if len(instruction) != 2 {
		return fmt.Errorf("%+v: invalid length, %+v", ErrEnableFeatureInstruction, instruction)
	}
	if instruction[0] != ENABLE_FEATURE {
		return fmt.Errorf("%+v: invalid enable feature action, %+v", ErrEnableFeatureInstruction, instruction)
	}
	return nil
}
