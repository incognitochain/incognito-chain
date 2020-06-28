package instruction

import (
	"fmt"
	"strconv"
	"strings"
)

type SwapInstruction struct {
	InPublicKeys  []string
	OutPublicKeys []string
	ChainID       int
	// old slashing, never used
	PunishedPublicKeys string
	// this field is only for replace committee
	NewRewardReceivers []string
	IsReplace          bool
}

func NewSwapInstructionWithValue(inPublicKeys []string, outPublicKeys []string, chainID int, punishedPublicKeys string, newRewardReceivers []string) *SwapInstruction {
	return &SwapInstruction{InPublicKeys: inPublicKeys, OutPublicKeys: outPublicKeys, ChainID: chainID, PunishedPublicKeys: punishedPublicKeys, NewRewardReceivers: newRewardReceivers}
}

func NewSwapInstruction() *SwapInstruction {
	return &SwapInstruction{}
}

func (s *SwapInstruction) GetType() string {
	return SWAP_ACTION
}

func (s *SwapInstruction) ToString() []string {
	swapInstructionStr := []string{SWAP_ACTION}
	swapInstructionStr = append(swapInstructionStr, strings.Join(s.InPublicKeys, SPLITTER))
	swapInstructionStr = append(swapInstructionStr, strings.Join(s.OutPublicKeys, SPLITTER))
	if s.ChainID == BEACON_CHAIN_ID {
		swapInstructionStr = append(swapInstructionStr, BEACON_INST)
		if len(s.NewRewardReceivers) > 0 {
			swapInstructionStr = append(swapInstructionStr, "")
		}
	} else {
		swapInstructionStr = append(swapInstructionStr, SHARD_INST)
		swapInstructionStr = append(swapInstructionStr, fmt.Sprintf("%v", s.ChainID))
	}
	swapInstructionStr = append(swapInstructionStr, s.PunishedPublicKeys)
	if len(s.NewRewardReceivers) > 0 {
		swapInstructionStr = append(swapInstructionStr, strings.Join(s.NewRewardReceivers, SPLITTER))
	}
	return swapInstructionStr
}

func (s *SwapInstruction) SetInPublicKeys(inPublicKeys []string) *SwapInstruction {
	s.InPublicKeys = inPublicKeys
	return s
}

func (s *SwapInstruction) SetOutPublicKeys(outPublicKeys []string) *SwapInstruction {
	s.OutPublicKeys = outPublicKeys
	return s
}

func (s *SwapInstruction) SetChainID(chainID int) *SwapInstruction {
	s.ChainID = chainID
	return s
}

func (s *SwapInstruction) SetPunishedPublicKeys(punishedPublicKeys string) *SwapInstruction {
	s.PunishedPublicKeys = punishedPublicKeys
	return s
}

func (s *SwapInstruction) SetNewRewardReceivers(newRewardReceivers []string) *SwapInstruction {
	s.NewRewardReceivers = newRewardReceivers
	return s
}

func (s *SwapInstruction) SetIsReplace(isReplace bool) *SwapInstruction {
	s.IsReplace = isReplace
	return s
}

func ValidateAndImportSwapInstructionFromString(instruction []string) (*SwapInstruction, error) {
	if err := ValidateSwapInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportSwapInstructionFromString(instruction), nil
}

func ImportSwapInstructionFromString(instruction []string) *SwapInstruction {
	swapInstruction := NewSwapInstruction()
	if len(instruction[1]) > 0 {
		swapInstruction.SetInPublicKeys(strings.Split(instruction[1], SPLITTER))
	}
	if len(instruction[2]) > 0 {
		swapInstruction.SetOutPublicKeys(strings.Split(instruction[2], SPLITTER))
	}
	if len(instruction) == 7 {
		swapInstruction.SetIsReplace(true)
		swapInstruction.SetNewRewardReceivers(strings.Split(instruction[6], SPLITTER))
	} else {
		swapInstruction.SetIsReplace(false)
		swapInstruction.SetNewRewardReceivers([]string{})
	}
	chain := instruction[3]
	if chain == BEACON_INST {
		swapInstruction.SetChainID(BEACON_CHAIN_ID)
		if len(instruction[4]) > 0 {
			swapInstruction.SetPunishedPublicKeys(instruction[4])
		}
	} else {
		chainID, _ := strconv.Atoi(instruction[4])
		swapInstruction.SetChainID(chainID)
		if len(instruction[5]) > 0 {
			swapInstruction.SetPunishedPublicKeys(instruction[5])
		}
	}
	return swapInstruction
}

// validate swap instruction sanity
// new reward receiver only present in replace committee
// beaconproducer.go: 356 - 367
func ValidateSwapInstructionSanity(instruction []string) error {
	if len(instruction) != 5 || len(instruction) != 6 || len(instruction) != 7 {
		return fmt.Errorf("invalid instruction length, %+v, %+v", len(instruction), instruction)
	}
	if instruction[0] != SWAP_ACTION {
		return fmt.Errorf("invalid swap action, %+v", instruction)
	}
	// beacon instruction
	if len(instruction) == 5 && instruction[3] != BEACON_INST {
		return fmt.Errorf("invalid swap beacon instruction, %+v", instruction)
	}
	// shard instruction
	if len(instruction) == 6 {
		if instruction[3] != SHARD_INST {
			return fmt.Errorf("invalid swap shard instruction, %+v", instruction)
		} else {
			_, err := strconv.Atoi(instruction[4])
			if err != nil {
				return fmt.Errorf("invalid swap shard id, %+v, %+v", err, instruction)
			}
		}
	}
	return nil
}
