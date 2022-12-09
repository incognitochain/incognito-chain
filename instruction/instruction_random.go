package instruction

import (
	"fmt"
	"strconv"
)

// Random Instruction which get nonce from bitcoin block
type RandomInstruction struct {
	randomNumber int64
}

func (s *RandomInstruction) RandomNumber() int64 {
	return s.randomNumber
}

func NewRandomInstructionWithValue(btcNonce int64) *RandomInstruction {
	return &RandomInstruction{randomNumber: btcNonce}
}

func NewRandomInstruction() *RandomInstruction {
	s := &RandomInstruction{}
	return s
}

func (s *RandomInstruction) GetType() string {
	return RANDOM_ACTION
}

func (s *RandomInstruction) ToString() []string {
	strs := []string{}
	strs = append(strs, RANDOM_ACTION)
	strs = append(strs, strconv.FormatInt(s.randomNumber, 10))
	strs = append(strs, "")
	strs = append(strs, "")
	strs = append(strs, "")
	return strs
}

func ValidateAndImportRandomInstructionFromString(instruction []string) (*RandomInstruction, error) {
	if err := ValidateRandomInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportRandomInstructionFromString(instruction), nil
}

func BuildRandomInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateRandomInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportRandomInstructionFromString(instruction), nil
}

// ImportRandomInstructionFromString is unsafe method
func ImportRandomInstructionFromString(instruction []string) *RandomInstruction {
	btcNonce, _ := strconv.ParseInt(instruction[1], 10, 64)
	r := NewRandomInstructionWithValue(btcNonce)
	return r
}

func ValidateRandomInstructionSanity(instruction []string) error {
	if len(instruction) < 2 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != RANDOM_ACTION {
		return fmt.Errorf("invalid random action, %+v", instruction)
	}
	if _, err := strconv.ParseInt(instruction[1], 10, 64); err != nil {
		return fmt.Errorf("invalid btc nonce value, %s", err)
	}
	return nil
}
