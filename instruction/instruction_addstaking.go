package instruction

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

type AddStakingInstruction struct {
	CommitteePublicKeys       []string
	StakingAmount             []uint64
	StakingTxIDs              []string
	CommitteePublicKeysStruct []incognitokey.CommitteePublicKey
}

func NewAddStakingInstructionWithValue(publicKeys []string, stakingAmount []uint64, txIDs []string) *AddStakingInstruction {
	res := &AddStakingInstruction{
		StakingAmount: stakingAmount,
		StakingTxIDs:  txIDs,
	}
	res.SetPublicKeys(publicKeys)
	return res
}

func NewAddStakingInstruction() *AddStakingInstruction {
	return &AddStakingInstruction{}
}

func (s *AddStakingInstruction) GetType() string {
	return ADD_STAKING_ACTION
}

func (s *AddStakingInstruction) IsEmpty() bool {
	return reflect.DeepEqual(s, NewAddStakingInstruction()) || len(s.CommitteePublicKeys) == 0
}

func (s *AddStakingInstruction) SetPublicKeys(publicKeys []string) (*AddStakingInstruction, error) {
	s.CommitteePublicKeys = publicKeys
	publicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return nil, err
	}
	s.CommitteePublicKeysStruct = publicKeyStructs
	return s, nil
}

func (s *AddStakingInstruction) SetTxIDs(txIDs []string) (*AddStakingInstruction, error) {
	s.StakingTxIDs = txIDs
	return s, nil
}

func (s *AddStakingInstruction) SetStakingAmount(stakingAmount []uint64) (*AddStakingInstruction, error) {
	s.StakingAmount = stakingAmount
	return s, nil
}

func (s *AddStakingInstruction) ToString() []string {
	addStakingInstructionStr := []string{ADD_STAKING_ACTION}
	stakingAmountStr := []string{}
	for _, v := range s.StakingAmount {
		str := strconv.FormatUint(v, 10)
		stakingAmountStr = append(stakingAmountStr, str)
	}
	addStakingInstructionStr = append(addStakingInstructionStr, strings.Join(s.CommitteePublicKeys, SPLITTER))
	addStakingInstructionStr = append(addStakingInstructionStr, strings.Join(stakingAmountStr, SPLITTER))
	addStakingInstructionStr = append(addStakingInstructionStr, strings.Join(s.StakingTxIDs, SPLITTER))
	return addStakingInstructionStr
}

func ValidateAndImportAddStakingInstructionFromString(instruction []string) (*AddStakingInstruction, error) {
	if err := ValidateAddStakingInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportAddStakingInstructionFromString(instruction), nil
}
func BuildAddStakingInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateAddStakingInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportAddStakingInstructionFromString(instruction), nil
}

func ImportAddStakingInstructionFromString(instruction []string) *AddStakingInstruction {
	addStakingInstruction := NewAddStakingInstruction()
	if len(instruction[1]) > 0 {
		publicKeys := strings.Split(instruction[1], SPLITTER)
		addStakingInstruction, _ = addStakingInstruction.SetPublicKeys(publicKeys)
	}
	stakingAmountStr := strings.Split(instruction[2], SPLITTER)
	stakingAmount := make([]uint64, len(stakingAmountStr))
	for i, v := range stakingAmountStr {
		amount, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil
		}
		stakingAmount[i] = amount
	}
	addStakingInstruction.SetStakingAmount(stakingAmount)
	if len(instruction[3]) > 0 {
		txIDs := strings.Split(instruction[3], SPLITTER)
		addStakingInstruction, _ = addStakingInstruction.SetTxIDs(txIDs)
	}
	return addStakingInstruction
}

func ValidateAddStakingInstructionSanity(instruction []string) error {
	if len(instruction) != 4 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != ADD_STAKING_ACTION {
		return fmt.Errorf("invalid stop auto stake action, %+v", instruction)
	}
	publicKeys := strings.Split(instruction[1], SPLITTER)
	_, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return err
	}
	return nil
}

func (s *AddStakingInstruction) DeleteSingleElement(index int) {
	s.CommitteePublicKeys = append(s.CommitteePublicKeys[:index], s.CommitteePublicKeys[index+1:]...)
	s.CommitteePublicKeysStruct = append(s.CommitteePublicKeysStruct[:index], s.CommitteePublicKeysStruct[index+1:]...)
	s.StakingAmount = append(s.StakingAmount[:index], s.StakingAmount[index+1:]...)
	s.StakingTxIDs = append(s.StakingTxIDs[:index], s.StakingTxIDs[index+1:]...)
}
