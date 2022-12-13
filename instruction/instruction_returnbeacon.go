package instruction

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ReturnBeaconStakeInstruction struct {
	PublicKeys       []string
	PublicKeysStruct []incognitokey.CommitteePublicKey
	ReturnAmounts    []uint64
	PercentReturns   []uint
	Reasons          []byte
}

func NewReturnBeaconStakeInsWithValue(
	publicKeys []string,
	reason []byte,
	amounts []uint64,
) *ReturnBeaconStakeInstruction {
	rsI := &ReturnBeaconStakeInstruction{}
	rsI, _ = rsI.SetPublicKeys(publicKeys)
	for _, _ = range publicKeys {
		rsI.PercentReturns = append(rsI.PercentReturns, 100)
	}
	rsI.SetReasons(reason)
	rsI.SetReturnAmounts(amounts)
	return rsI
}

func NewReturnBeaconStakeIns() *ReturnBeaconStakeInstruction {
	return &ReturnBeaconStakeInstruction{}
}

func (rsI *ReturnBeaconStakeInstruction) IsEmpty() bool {
	return reflect.DeepEqual(rsI, NewReturnStakeIns()) ||
		len(rsI.PublicKeysStruct) == 0 && len(rsI.PublicKeys) == 0
}

func (rsI *ReturnBeaconStakeInstruction) SetPublicKeys(publicKeys []string) (*ReturnBeaconStakeInstruction, error) {
	if publicKeys == nil {
		return nil, errors.New("Public Keys Are Null")
	}
	rsI.PublicKeys = publicKeys
	publicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return nil, err
	}
	rsI.PublicKeysStruct = publicKeyStructs
	return rsI, nil
}

func (rsI *ReturnBeaconStakeInstruction) SetPercentReturns(percentReturns []uint) error {
	rsI.PercentReturns = percentReturns
	return nil
}

func (rsI *ReturnBeaconStakeInstruction) SetReasons(reason []byte) error {
	rsI.Reasons = reason
	return nil
}
func (rsI *ReturnBeaconStakeInstruction) SetReturnAmounts(amounts []uint64) error {
	rsI.ReturnAmounts = amounts
	return nil
}

func (rsI *ReturnBeaconStakeInstruction) GetType() string {
	return RETURN_BEACON_ACTION
}

func (rsI *ReturnBeaconStakeInstruction) GetPercentReturns() []uint {
	return rsI.PercentReturns
}

func (rsI *ReturnBeaconStakeInstruction) GetPublicKey() []string {
	return rsI.PublicKeys
}

func (rsI *ReturnBeaconStakeInstruction) GetReason() []byte {
	return rsI.Reasons
}

func (rsI *ReturnBeaconStakeInstruction) GetReturnAmounts() []uint64 {
	return rsI.ReturnAmounts
}

func (rsI *ReturnBeaconStakeInstruction) ToString() []string {
	returnStakeInsStr := []string{RETURN_BEACON_ACTION}
	returnStakeInsStr = append(returnStakeInsStr, strings.Join(rsI.PublicKeys, SPLITTER))
	percentReturnsStr := make([]string, len(rsI.PercentReturns))
	for i, v := range rsI.PercentReturns {
		percentReturnsStr[i] = strconv.Itoa(int(v))
	}
	returnStakeInsStr = append(returnStakeInsStr, strings.Join(percentReturnsStr, SPLITTER))
	if len(rsI.Reasons) != 0 {
		reasonStrs := make([]string, len(rsI.Reasons))
		for i, v := range rsI.Reasons {
			reasonStrs[i] = strconv.Itoa(int(v))
		}
		returnStakeInsStr = append(returnStakeInsStr, strings.Join(reasonStrs, SPLITTER))
	}
	if len(rsI.ReturnAmounts) != 0 {
		returnAmountStrs := make([]string, len(rsI.ReturnAmounts))
		for i, v := range rsI.ReturnAmounts {
			returnAmountStrs[i] = strconv.FormatUint(v, 10)
		}
		returnStakeInsStr = append(returnStakeInsStr, strings.Join(returnAmountStrs, SPLITTER))
	}
	return returnStakeInsStr
}

func (rsI *ReturnBeaconStakeInstruction) AddNewRequest(publicKey string, amount uint64) {
	rsI.PublicKeys = append(rsI.PublicKeys, publicKey)
	publicKeyStruct, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{publicKey})
	rsI.PublicKeysStruct = append(rsI.PublicKeysStruct, publicKeyStruct[0])
	rsI.ReturnAmounts = append(rsI.ReturnAmounts, amount)
	rsI.PercentReturns = append(rsI.PercentReturns, 100)
}

func ValidateAndImportReturnBeaconStakingInstructionFromString(instruction []string) (*ReturnBeaconStakeInstruction, error) {
	if err := ValidateReturnBeaconStakingInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportReturnBeaconStakingInstructionFromString(instruction)
}

func BuildReturnBeaconStakingInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateReturnBeaconStakingInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportReturnBeaconStakingInstructionFromString(instruction)
}

func ImportReturnBeaconStakingInstructionFromString(instruction []string) (*ReturnBeaconStakeInstruction, error) {
	returnStakingIns := NewReturnBeaconStakeIns()
	var err error
	returnStakingIns, err = returnStakingIns.SetPublicKeys(strings.Split(instruction[1], SPLITTER))
	if err != nil {
		return nil, err
	}

	percentRetunrsStr := strings.Split(instruction[2], SPLITTER)
	percentReturns := make([]uint, len(percentRetunrsStr))
	for i, v := range percentRetunrsStr {
		tempPercent, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		percentReturns[i] = uint(tempPercent)
	}
	returnStakingIns.SetPercentReturns(percentReturns)

	reasonsStr := strings.Split(instruction[3], SPLITTER)
	reasons := make([]byte, len(reasonsStr))
	for i, v := range reasonsStr {
		reason, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		reasons[i] = byte(reason)
	}
	returnStakingIns.SetReasons(reasons)
	returnAmountsStr := strings.Split(instruction[4], SPLITTER)
	returnAmounts := make([]uint64, len(returnAmountsStr))
	for i, v := range returnAmountsStr {
		amount, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, err
		}
		returnAmounts[i] = amount
	}
	returnStakingIns.SetReturnAmounts(returnAmounts)

	return returnStakingIns, err
}

func ValidateReturnBeaconStakingInstructionSanity(instruction []string) error {
	if len(instruction) != 5 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != RETURN_BEACON_ACTION {
		return fmt.Errorf("invalid return staking action, %+v", instruction)
	}
	publicKeys := strings.Split(instruction[1], SPLITTER)
	_, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return err
	}
	percentRetunrsStr := strings.Split(instruction[2], SPLITTER)
	percentReturns := make([]uint, len(percentRetunrsStr))
	for i, v := range percentRetunrsStr {
		tempPercent, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid percent return %+v", err)
		}
		percentReturns[i] = uint(tempPercent)
	}
	if len(percentReturns) != len(publicKeys) {
		return fmt.Errorf("invalid reward percentReturns & public Keys length, %+v", instruction)
	}
	return nil
}
