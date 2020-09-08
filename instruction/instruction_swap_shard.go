package instruction

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

type SwapShardInstruction struct {
	InPublicKeys        []string
	InPublicKeyStructs  []incognitokey.CommitteePublicKey
	OutPublicKeys       []string
	OutPublicKeyStructs []incognitokey.CommitteePublicKey
	ChainID             int
	Height              uint64
	Type                int
}

func NewSwapShardInstructionWithValue(
	inPublicKeys, outPublicKeys []string,
	chainID, typeIns int, height uint64) *SwapShardInstruction {
	return &SwapShardInstruction{
		InPublicKeys:  inPublicKeys,
		OutPublicKeys: outPublicKeys,
		ChainID:       chainID,
		Height:        height,
		Type:          typeIns,
	}
}

func NewSwapShardInstruction() *SwapShardInstruction {
	return &SwapShardInstruction{}
}

func (s *SwapShardInstruction) GetType() string {
	return SWAP_SHARD_ACTION
}

func (s *SwapShardInstruction) ToString() []string {
	SwapShardInstructionStr := []string{SWAP_SHARD_ACTION}
	SwapShardInstructionStr = append(SwapShardInstructionStr, strings.Join(s.InPublicKeys, SPLITTER))
	SwapShardInstructionStr = append(SwapShardInstructionStr, strings.Join(s.OutPublicKeys, SPLITTER))
	SwapShardInstructionStr = append(SwapShardInstructionStr, fmt.Sprintf("%v", s.ChainID))
	SwapShardInstructionStr = append(SwapShardInstructionStr, fmt.Sprintf("%v", s.Height))
	SwapShardInstructionStr = append(SwapShardInstructionStr, fmt.Sprintf("%v", s.Type))
	return SwapShardInstructionStr
}

func (s *SwapShardInstruction) SetInPublicKeys(inPublicKeys []string) (*SwapShardInstruction, error) {
	s.InPublicKeys = inPublicKeys
	inPublicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(inPublicKeys)
	if err != nil {
		return nil, err
	}
	s.InPublicKeyStructs = inPublicKeyStructs
	return s, nil
}

func (s *SwapShardInstruction) SetOutPublicKeys(outPublicKeys []string) (*SwapShardInstruction, error) {
	s.OutPublicKeys = outPublicKeys
	outPublicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(outPublicKeys)
	if err != nil {
		return nil, err
	}
	s.OutPublicKeyStructs = outPublicKeyStructs
	return s, nil
}

func (s *SwapShardInstruction) SetChainID(chainID int) *SwapShardInstruction {
	s.ChainID = chainID
	return s
}

func (s *SwapShardInstruction) SetType(typeIns int) *SwapShardInstruction {
	s.Type = typeIns
	return s
}

func ValidateAndImportSwapShardInstructionFromString(instruction []string) (*SwapShardInstruction, error) {
	if err := ValidateSwapShardInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportSwapShardInstructionFromString(instruction), nil
}

func ImportSwapShardInstructionFromString(instruction []string) *SwapShardInstruction {
	swapShardInstruction := NewSwapShardInstruction()
	inPublicKey := []string{}
	outPublicKey := []string{}

	if len(instruction[1]) > 0 {
		inPublicKey = strings.Split(instruction[1], SPLITTER)
	}
	swapShardInstruction, _ = swapShardInstruction.SetInPublicKeys(inPublicKey)
	if len(instruction[2]) > 0 {
		outPublicKey = strings.Split(instruction[2], SPLITTER)
	}
	swapShardInstruction, _ = swapShardInstruction.SetOutPublicKeys(outPublicKey)
	swapShardInstruction.ChainID, _ = strconv.Atoi(instruction[3])
	swapShardInstruction.Height, _ = strconv.ParseUint(instruction[4], 10, 64)
	swapShardInstruction.Type, _ = strconv.Atoi(instruction[5])

	return swapShardInstruction
}

// validate SwapShard instruction sanity
// new reward receiver only present in replace committee
// beaconproducer.go: 356 - 367
func ValidateSwapShardInstructionSanity(instruction []string) error {
	if len(instruction) != 6 {
		return fmt.Errorf("invalid instruction length %+v, %+v, expect %+v", len(instruction), instruction, 6)
	}
	if instruction[0] != SWAP_SHARD_ACTION {
		return fmt.Errorf("invalid SwapShard action, %+v", instruction)
	}
	_, err1 := incognitokey.CommitteeBase58KeyListToStruct(strings.Split(instruction[1], ","))
	if err1 != nil {
		return fmt.Errorf("invalid SwapShard in public key type, %+v, %+v", err1, instruction)
	}
	_, err2 := incognitokey.CommitteeBase58KeyListToStruct(strings.Split(instruction[2], ","))
	if err2 != nil {
		return fmt.Errorf("invalid SwapShard out public key type, %+v, %+v", err2, instruction)
	}
	_, err3 := strconv.ParseUint(instruction[3], 10, 64)
	if err3 != nil {
		return fmt.Errorf("invalid SwapShard chainID, %+v, %+v", err3, instruction)
	}
	_, err4 := strconv.ParseUint(instruction[4], 10, 64)
	if err4 != nil {
		return fmt.Errorf("invalid RequestShardSwap height, %+v, %+v", err4, instruction)
	}
	_, err5 := strconv.ParseInt(instruction[5], 10, 64)
	if err5 != nil {
		return fmt.Errorf("invalid RequestShardSwap type, %+v, %+v", err5, instruction)
	}
	return nil
}
