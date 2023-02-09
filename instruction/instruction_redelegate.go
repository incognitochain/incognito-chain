package instruction

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ReDelegateInstruction struct {
	CommitteePublicKeys       []string
	CommitteePublicKeysStruct []incognitokey.CommitteePublicKey
	DelegateList              []string
	DelegateListStruct        []incognitokey.CommitteePublicKey
	DelegateUIDList           []string
}

func NewReDelegateInstructionWithValue(publicKeys, redelegateList []string, uIDs []string) *ReDelegateInstruction {
	res := &ReDelegateInstruction{}
	res.SetPublicKeys(publicKeys)
	res.SetDelegateList(redelegateList)
	copy(res.DelegateUIDList, uIDs)
	return res
}

func NewReDelegateInstruction() *ReDelegateInstruction {
	return &ReDelegateInstruction{}
}

func (s *ReDelegateInstruction) GetType() string {
	return RE_DELEGATE
}

func (s *ReDelegateInstruction) IsEmpty() bool {
	return reflect.DeepEqual(s, NewReDelegateInstruction()) || len(s.CommitteePublicKeys) == 0 || len(s.DelegateList) == 0
}

func (s *ReDelegateInstruction) SetPublicKeys(publicKeys []string) (*ReDelegateInstruction, error) {
	s.CommitteePublicKeys = publicKeys
	publicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return nil, err
	}
	s.CommitteePublicKeysStruct = publicKeyStructs
	return s, nil
}

func (s *ReDelegateInstruction) SetDelegateList(delegateList []string) (*ReDelegateInstruction, error) {
	s.DelegateList = delegateList
	publicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(delegateList)
	if err != nil {
		return nil, err
	}
	s.DelegateListStruct = publicKeyStructs
	return s, nil
}

func (s *ReDelegateInstruction) ToString() []string {
	redelegateInstructionStr := []string{RE_DELEGATE}
	redelegateInstructionStr = append(redelegateInstructionStr, strings.Join(s.CommitteePublicKeys, SPLITTER))
	redelegateInstructionStr = append(redelegateInstructionStr, strings.Join(s.DelegateList, SPLITTER))
	redelegateInstructionStr = append(redelegateInstructionStr, strings.Join(s.DelegateUIDList, SPLITTER))
	return redelegateInstructionStr
}

func ValidateAndImportReDelegateInstructionFromString(instruction []string) (*ReDelegateInstruction, error) {
	if err := ValidateReDelegateInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportReDelegateInstructionFromString(instruction), nil
}

func BuildReDelegateInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateReDelegateInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportReDelegateInstructionFromString(instruction), nil
}

func ImportReDelegateInstructionFromString(instruction []string) *ReDelegateInstruction {
	redelegateInstruction := NewReDelegateInstruction()
	if len(instruction[1]) > 0 {
		publicKeys := strings.Split(instruction[1], SPLITTER)
		redelegateInstruction, _ = redelegateInstruction.SetPublicKeys(publicKeys)
	}
	if len(instruction[2]) > 0 {
		publicKeys := strings.Split(instruction[2], SPLITTER)
		redelegateInstruction, _ = redelegateInstruction.SetDelegateList(publicKeys)
	}
	if len(instruction[3]) > 0 {
		publicKeys := strings.Split(instruction[3], SPLITTER)
		redelegateInstruction, _ = redelegateInstruction.SetDelegateList(publicKeys)
	}
	return redelegateInstruction
}

func ValidateReDelegateInstructionSanity(instruction []string) error {
	if len(instruction) != 4 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != RE_DELEGATE {
		return fmt.Errorf("invalid re delegate action, %+v", instruction)
	}
	publicKeys := strings.Split(instruction[1], SPLITTER)
	_, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return err
	}
	delegateList := strings.Split(instruction[2], SPLITTER)
	_, err = incognitokey.CommitteeBase58KeyListToStruct(delegateList)
	if err != nil {
		return err
	}
	delegateUIDList := strings.Split(instruction[3], SPLITTER)
	for _, uID := range delegateUIDList {
		_, err := common.Hash{}.NewHashFromStr(uID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ReDelegateInstruction) DeleteSingleElement(index int) {
	s.CommitteePublicKeys = append(s.CommitteePublicKeys[:index], s.CommitteePublicKeys[index+1:]...)
	s.CommitteePublicKeysStruct = append(s.CommitteePublicKeysStruct[:index], s.CommitteePublicKeysStruct[index+1:]...)
}
