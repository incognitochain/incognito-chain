package instruction

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ReDelegateInstruction struct {
	CommitteePublicKeys       []string
	CommitteePublicKeysStruct []incognitokey.CommitteePublicKey
	DelegateList              []string
	DelegateListStruct        []incognitokey.CommitteePublicKey
}

func NewReDelegateInstructionWithValue(publicKeys []string) *ReDelegateInstruction {
	res := &ReDelegateInstruction{}
	res.SetPublicKeys(publicKeys)
	return res
}

func NewReDelegateInstruction() *ReDelegateInstruction {
	return &ReDelegateInstruction{}
}

func (s *ReDelegateInstruction) GetType() string {
	return STOP_AUTO_STAKE_ACTION
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
	return redelegateInstructionStr
}

func ValidateAndImportReDelegateInstructionFromString(instruction []string) (*ReDelegateInstruction, error) {
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
		redelegateInstruction, _ = redelegateInstruction.SetPublicKeys(publicKeys)
	}
	return redelegateInstruction
}

func ValidateReDelegateInstructionSanity(instruction []string) error {
	if len(instruction) != 3 {
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
	_, err = incognitokey.CommitteeKeyListToStruct(delegateList)
	if err != nil {
		return err
	}
	return nil
}

func (s *ReDelegateInstruction) DeleteSingleElement(index int) {
	s.CommitteePublicKeys = append(s.CommitteePublicKeys[:index], s.CommitteePublicKeys[index+1:]...)
	s.CommitteePublicKeysStruct = append(s.CommitteePublicKeysStruct[:index], s.CommitteePublicKeysStruct[index+1:]...)
}
