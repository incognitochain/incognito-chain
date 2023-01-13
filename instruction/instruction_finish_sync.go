package instruction

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/log/proto"
)

var (
	ErrFinishSyncInstruction = errors.New("finish sync instruction error")
)

// FinishSyncInstruction :
// format: "finish_sync", "0", "key1,key2"
type FinishSyncInstruction struct {
	ChainID          int
	PublicKeys       []string
	PublicKeysStruct []incognitokey.CommitteePublicKey
	instructionBase
}

func NewFinishSyncInstructionWithValue(chainID int, publicKeys []string) *FinishSyncInstruction {
	finishSyncInstruction := NewFinishSyncInstruction()
	finishSyncInstruction.SetChainID(chainID)
	if chainID == BEACON_CHAIN_ID {
		finishSyncInstruction.featureID = proto.FID_CONSENSUS_BEACON
	}
	finishSyncInstruction.SetPublicKeys(publicKeys)
	return finishSyncInstruction
}

func NewFinishSyncInstruction() *FinishSyncInstruction {
	return &FinishSyncInstruction{
		instructionBase: instructionBase{
			featureID: proto.FID_CONSENSUS_SHARD,
			logOnly:   false,
		},
	}
}

func (f *FinishSyncInstruction) GetType() string {
	return FINISH_SYNC_ACTION
}

func (f *FinishSyncInstruction) IsEmpty() bool {
	return reflect.DeepEqual(f, NewFinishSyncInstruction()) ||
		len(f.PublicKeys) == 0 && len(f.PublicKeysStruct) == 0
}

func (f *FinishSyncInstruction) ToString() []string {
	finishSyncInstructionStr := []string{FINISH_SYNC_ACTION}
	finishSyncInstructionStr = append(finishSyncInstructionStr, fmt.Sprintf("%v", f.ChainID))
	finishSyncInstructionStr = append(finishSyncInstructionStr, strings.Join(f.PublicKeys, SPLITTER))
	return finishSyncInstructionStr
}

func (f *FinishSyncInstruction) SetChainID(chainID int) *FinishSyncInstruction {
	f.ChainID = chainID
	return f
}

func (f *FinishSyncInstruction) SetPublicKeys(publicKeys []string) *FinishSyncInstruction {
	f.PublicKeys = publicKeys
	publicKeyStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	f.PublicKeysStruct = publicKeyStructs
	return f
}

func ValidateAndImportFinishSyncInstructionFromString(instruction []string) (*FinishSyncInstruction, error) {
	if err := ValidateFinishSyncInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportFinishSyncInstructionFromString(instruction)
}

// ImportFinishSyncInstructionFromString is unsafe method
func ImportFinishSyncInstructionFromString(instruction []string) (*FinishSyncInstruction, error) {
	finishSyncInstruction := NewFinishSyncInstruction()
	tempShardID := instruction[1]
	chainID, _ := strconv.Atoi(tempShardID)
	finishSyncInstruction.SetChainID(chainID)
	finishSyncInstruction.SetPublicKeys(strings.Split(instruction[2], SPLITTER))
	publicKeysStruct, err := incognitokey.CommitteeBase58KeyListToStruct(finishSyncInstruction.PublicKeys)
	if err != nil {
		return nil, err
	}
	finishSyncInstruction.PublicKeysStruct = publicKeysStruct
	return finishSyncInstruction, err
}

// ValidateFinishSyncInstructionSanity ...
func ValidateFinishSyncInstructionSanity(instruction []string) error {
	if len(instruction) != 3 {
		return fmt.Errorf("%+v: invalid length, %+v", ErrFinishSyncInstruction, instruction)
	}
	if instruction[0] != FINISH_SYNC_ACTION {
		return fmt.Errorf("%+v: invalid finish sync action, %+v", ErrFinishSyncInstruction, instruction)
	}
	if _, err := strconv.Atoi(instruction[1]); err != nil {
		return fmt.Errorf("%+v: invalid finish sync shard ID, err %+v, %+v", ErrFinishSyncInstruction, err, instruction)
	}
	return nil
}
