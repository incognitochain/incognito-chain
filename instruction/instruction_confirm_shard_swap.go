package instruction

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ConfirmShardSwapInstruction struct {
	InPublicKeys        []string
	InPublicKeyStructs  []incognitokey.CommitteePublicKey
	OutPublicKeys       []string
	OutPublicKeyStructs []incognitokey.CommitteePublicKey
	ChainID             int
	Epoch               uint64
	RandomNumber        int64
}

func NewConfirmShardSwapInstructionWithValue(inPublicKeys []string, outPublicKeys []string, chainID int, epoch uint64, randomNumber int64) *ConfirmShardSwapInstruction {
	return &ConfirmShardSwapInstruction{
		InPublicKeys:  inPublicKeys,
		OutPublicKeys: outPublicKeys,
		ChainID:       chainID,
		Epoch:         epoch,
		RandomNumber:  randomNumber,
	}
}

func NewConfirmShardSwapInstruction() *ConfirmShardSwapInstruction {
	return &ConfirmShardSwapInstruction{}
}

func (s *ConfirmShardSwapInstruction) GetType() string {
	return CONFIRM_SHARD_SWAP_ACTION
}

func (s *ConfirmShardSwapInstruction) ToString() []string {
	ConfirmShardSwapInstructionStr := []string{CONFIRM_SHARD_SWAP_ACTION}
	ConfirmShardSwapInstructionStr = append(ConfirmShardSwapInstructionStr, strings.Join(s.InPublicKeys, SPLITTER))
	ConfirmShardSwapInstructionStr = append(ConfirmShardSwapInstructionStr, strings.Join(s.OutPublicKeys, SPLITTER))
	ConfirmShardSwapInstructionStr = append(ConfirmShardSwapInstructionStr, fmt.Sprintf("%v", s.ChainID))
	ConfirmShardSwapInstructionStr = append(ConfirmShardSwapInstructionStr, fmt.Sprintf("%v", s.Epoch))
	ConfirmShardSwapInstructionStr = append(ConfirmShardSwapInstructionStr, fmt.Sprintf("%v", s.RandomNumber))
	return ConfirmShardSwapInstructionStr
}

func (s *ConfirmShardSwapInstruction) IsEmpty() bool {
	return len(s.InPublicKeys) == 0 && len(s.OutPublicKeys) == 0
}
func (s *ConfirmShardSwapInstruction) SetInPublicKeys(inPublicKeys []string) (*ConfirmShardSwapInstruction, error) {
	s.InPublicKeys = inPublicKeys
	inPublicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(inPublicKeys)
	if err != nil {
		return nil, err
	}
	s.InPublicKeyStructs = inPublicKeyStructs
	return s, nil
}

func (s *ConfirmShardSwapInstruction) SetOutPublicKeys(outPublicKeys []string) (*ConfirmShardSwapInstruction, error) {
	s.OutPublicKeys = outPublicKeys
	outPublicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(outPublicKeys)
	if err != nil {
		return nil, err
	}
	s.OutPublicKeyStructs = outPublicKeyStructs
	return s, nil
}

func (s *ConfirmShardSwapInstruction) SetChainID(chainID int) *ConfirmShardSwapInstruction {
	s.ChainID = chainID
	return s
}

func ValidateAndImportConfirmShardSwapInstructionFromString(instruction []string) (*ConfirmShardSwapInstruction, error) {
	if err := ValidateConfirmShardSwapInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportConfirmShardSwapInstructionFromString(instruction), nil
}

func ImportConfirmShardSwapInstructionFromString(instruction []string) *ConfirmShardSwapInstruction {
	ConfirmShardSwapInstruction := NewConfirmShardSwapInstruction()
	inPublicKey := []string{}
	outPublicKey := []string{}

	if len(instruction[1]) > 0 {
		inPublicKey = strings.Split(instruction[1], SPLITTER)
	}
	ConfirmShardSwapInstruction, _ = ConfirmShardSwapInstruction.SetInPublicKeys(inPublicKey)
	if len(instruction[2]) > 0 {
		outPublicKey = strings.Split(instruction[2], SPLITTER)
	}
	ConfirmShardSwapInstruction, _ = ConfirmShardSwapInstruction.SetOutPublicKeys(outPublicKey)
	ConfirmShardSwapInstruction.ChainID, _ = strconv.Atoi(instruction[3])
	ConfirmShardSwapInstruction.Epoch, _ = strconv.ParseUint(instruction[4], 10, 64)
	ConfirmShardSwapInstruction.RandomNumber, _ = strconv.ParseInt(instruction[5], 10, 64)

	return ConfirmShardSwapInstruction
}

// validate ConfirmShardSwap instruction sanity
// new reward receiver only present in replace committee
// beaconproducer.go: 356 - 367
func ValidateConfirmShardSwapInstructionSanity(instruction []string) error {
	if len(instruction) != 6 {
		return fmt.Errorf("invalid instruction length %+v, %+v, expect %+v", len(instruction), instruction, 6)
	}
	if instruction[0] != REQUEST_SHARD_SWAP_ACTION {
		return fmt.Errorf("invalid ConfirmShardSwap action, %+v", instruction)
	}
	_, err1 := incognitokey.CommitteeBase58KeyListToStruct(strings.Split(instruction[1], ","))
	if err1 != nil {
		return fmt.Errorf("invalid ConfirmShardSwap in public key type, %+v, %+v", err1, instruction)
	}
	_, err2 := incognitokey.CommitteeBase58KeyListToStruct(strings.Split(instruction[2], ","))
	if err2 != nil {
		return fmt.Errorf("invalid ConfirmShardSwap out public key type, %+v, %+v", err2, instruction)
	}
	_, err3 := strconv.Atoi(instruction[3])
	if err3 != nil {
		return fmt.Errorf("invalid ConfirmShardSwap shardID, %+v, %+v", err3, instruction)
	}
	_, err4 := strconv.ParseUint(instruction[4], 10, 64)
	if err4 != nil {
		return fmt.Errorf("invalid ConfirmShardSwap epoch, %+v, %+v", err4, instruction)
	}
	_, err5 := strconv.ParseInt(instruction[5], 10, 64)
	if err5 != nil {
		return fmt.Errorf("invalid ConfirmShardSwap random number, %+v, %+v", err5, instruction)
	}
	return nil
}

func ConvertRequestToConfirmShardSwapInstruction(requestInstruction *RequestShardSwapInstruction) *ConfirmShardSwapInstruction {
	confirmShardSwapInstruction := NewConfirmShardSwapInstructionWithValue(
		requestInstruction.InPublicKeys,
		requestInstruction.OutPublicKeys,
		requestInstruction.ChainID,
		requestInstruction.Epoch,
		requestInstruction.RandomNumber,
	)
	return confirmShardSwapInstruction
}
