package instruction

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

type RequestShardSwapInstruction struct {
	InPublicKeys        []string
	InPublicKeyStructs  []incognitokey.CommitteePublicKey
	OutPublicKeys       []string
	OutPublicKeyStructs []incognitokey.CommitteePublicKey
	ChainID             int
	Epoch               uint64
	RandomNumber        int64
}

func NewRequestShardSwapInstructionWithValue(inPublicKeys []string, outPublicKeys []string, chainID int, epoch uint64, randomNumber int64) *RequestShardSwapInstruction {
	return &RequestShardSwapInstruction{
		InPublicKeys:  inPublicKeys,
		OutPublicKeys: outPublicKeys,
		ChainID:       chainID,
		Epoch:         epoch,
		RandomNumber:  randomNumber,
	}
}

func NewRequestShardSwapInstruction() *RequestShardSwapInstruction {
	return &RequestShardSwapInstruction{}
}

func (s *RequestShardSwapInstruction) GetType() string {
	return REQUEST_SHARD_SWAP_ACTION
}

func (s *RequestShardSwapInstruction) ToString() []string {
	RequestShardSwapInstructionStr := []string{REQUEST_SHARD_SWAP_ACTION}
	RequestShardSwapInstructionStr = append(RequestShardSwapInstructionStr, strings.Join(s.InPublicKeys, SPLITTER))
	RequestShardSwapInstructionStr = append(RequestShardSwapInstructionStr, strings.Join(s.OutPublicKeys, SPLITTER))
	RequestShardSwapInstructionStr = append(RequestShardSwapInstructionStr, fmt.Sprintf("%v", s.ChainID))
	RequestShardSwapInstructionStr = append(RequestShardSwapInstructionStr, fmt.Sprintf("%v", s.Epoch))
	RequestShardSwapInstructionStr = append(RequestShardSwapInstructionStr, fmt.Sprintf("%v", s.RandomNumber))
	return RequestShardSwapInstructionStr
}

func (s *RequestShardSwapInstruction) SetInPublicKeys(inPublicKeys []string) (*RequestShardSwapInstruction, error) {
	s.InPublicKeys = inPublicKeys
	inPublicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(inPublicKeys)
	if err != nil {
		return nil, err
	}
	s.InPublicKeyStructs = inPublicKeyStructs
	return s, nil
}

func (s *RequestShardSwapInstruction) SetOutPublicKeys(outPublicKeys []string) (*RequestShardSwapInstruction, error) {
	s.OutPublicKeys = outPublicKeys
	outPublicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(outPublicKeys)
	if err != nil {
		return nil, err
	}
	s.OutPublicKeyStructs = outPublicKeyStructs
	return s, nil
}

func (s *RequestShardSwapInstruction) SetChainID(chainID int) *RequestShardSwapInstruction {
	s.ChainID = chainID
	return s
}

func ValidateAndImportRequestShardSwapInstructionFromString(instruction []string) (*RequestShardSwapInstruction, error) {
	if err := ValidateRequestShardSwapInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportRequestShardSwapInstructionFromString(instruction), nil
}

func ImportRequestShardSwapInstructionFromString(instruction []string) *RequestShardSwapInstruction {
	RequestShardSwapInstruction := NewRequestShardSwapInstruction()
	inPublicKey := []string{}
	outPublicKey := []string{}

	if len(instruction[1]) > 0 {
		inPublicKey = strings.Split(instruction[1], SPLITTER)
	}
	RequestShardSwapInstruction, _ = RequestShardSwapInstruction.SetInPublicKeys(inPublicKey)
	if len(instruction[2]) > 0 {
		outPublicKey = strings.Split(instruction[2], SPLITTER)
	}
	RequestShardSwapInstruction, _ = RequestShardSwapInstruction.SetOutPublicKeys(outPublicKey)
	RequestShardSwapInstruction.ChainID, _ = strconv.Atoi(instruction[3])
	RequestShardSwapInstruction.Epoch, _ = strconv.ParseUint(instruction[4], 10, 64)
	RequestShardSwapInstruction.RandomNumber, _ = strconv.ParseInt(instruction[5], 10, 64)

	return RequestShardSwapInstruction
}

// validate RequestShardSwap instruction sanity
// new reward receiver only present in replace committee
// beaconproducer.go: 356 - 367
func ValidateRequestShardSwapInstructionSanity(instruction []string) error {
	if len(instruction) != 6 {
		return fmt.Errorf("invalid instruction length %+v, %+v, expect %+v", len(instruction), instruction, 6)
	}
	if instruction[0] != REQUEST_SHARD_SWAP_ACTION {
		return fmt.Errorf("invalid RequestShardSwap action, %+v", instruction)
	}
	_, err1 := incognitokey.CommitteeBase58KeyListToStruct(strings.Split(instruction[1], ","))
	if err1 != nil {
		return fmt.Errorf("invalid RequestShardSwap in public key type, %+v, %+v", err1, instruction)
	}
	_, err2 := incognitokey.CommitteeBase58KeyListToStruct(strings.Split(instruction[2], ","))
	if err2 != nil {
		return fmt.Errorf("invalid RequestShardSwap out public key type, %+v, %+v", err2, instruction)
	}
	_, err3 := strconv.Atoi(instruction[3])
	if err3 != nil {
		return fmt.Errorf("invalid RequestShardSwap shardID, %+v, %+v", err3, instruction)
	}
	_, err4 := strconv.ParseUint(instruction[4], 10, 64)
	if err4 != nil {
		return fmt.Errorf("invalid RequestShardSwap epoch, %+v, %+v", err4, instruction)
	}
	_, err5 := strconv.ParseInt(instruction[5], 10, 64)
	if err5 != nil {
		return fmt.Errorf("invalid RequestShardSwap random number, %+v, %+v", err5, instruction)
	}
	return nil
}
