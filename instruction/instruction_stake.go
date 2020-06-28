package instruction

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"strings"
)

type StakeInstruction struct {
	PublicKeys       []string
	PublicKeyStructs []incognitokey.CommitteePublicKey
	Chain            string
	TxStakes         []string
	RewardReceivers  []string
	AutoStakingFlag  []bool
}

func NewStakeInstructionWithValue(publicKeys []string, chain string, txStakes []string, rewardReceivers []string, autoStakingFlag []bool) *StakeInstruction {
	return &StakeInstruction{PublicKeys: publicKeys, Chain: chain, TxStakes: txStakes, RewardReceivers: rewardReceivers, AutoStakingFlag: autoStakingFlag}
}

func NewStakeInstruction() *StakeInstruction {
	return &StakeInstruction{}
}

func (s *StakeInstruction) GetType() string {
	return STAKE_ACTION
}

func (s *StakeInstruction) SetPublicKeys(publicKeys []string) (*StakeInstruction, error) {
	s.PublicKeys = publicKeys
	publicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return nil, err
	}
	s.PublicKeyStructs = publicKeyStructs
	return s, nil
}

func (s *StakeInstruction) SetChain(chain string) *StakeInstruction {
	s.Chain = chain
	return s
}

func (s *StakeInstruction) SetTxStakes(txStakes []string) *StakeInstruction {
	s.TxStakes = txStakes
	return s
}

func (s *StakeInstruction) SetRewardReceivers(rewardReceivers []string) *StakeInstruction {
	s.RewardReceivers = rewardReceivers
	return s
}

func (s *StakeInstruction) SetAutoStakingFlag(autoStakingFlag []bool) *StakeInstruction {
	s.AutoStakingFlag = autoStakingFlag
	return s
}

func (s *StakeInstruction) ToString() []string {
	stakeInstructionStr := []string{STAKE_ACTION}
	stakeInstructionStr = append(stakeInstructionStr, strings.Join(s.PublicKeys, SPLITTER))
	stakeInstructionStr = append(stakeInstructionStr, s.Chain)
	stakeInstructionStr = append(stakeInstructionStr, strings.Join(s.TxStakes, SPLITTER))
	stakeInstructionStr = append(stakeInstructionStr, strings.Join(s.RewardReceivers, SPLITTER))
	tempStopAutoStakeFlag := []string{}
	for _, v := range s.AutoStakingFlag {
		if v == true {
			tempStopAutoStakeFlag = append(tempStopAutoStakeFlag, TRUE)
		} else {
			tempStopAutoStakeFlag = append(tempStopAutoStakeFlag, FALSE)
		}
	}
	stakeInstructionStr = append(stakeInstructionStr, strings.Join(tempStopAutoStakeFlag, SPLITTER))
	return stakeInstructionStr
}

func ValidateAndImportStakeInstructionFromString(instruction []string) (*StakeInstruction, error) {
	if err := ValidateStakeInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportStakeInstructionFromString(instruction), nil
}

// ImportStakeInstructionFromString is unsafe method
func ImportStakeInstructionFromString(instruction []string) *StakeInstruction {
	stakeInstruction := NewStakeInstruction()
	stakeInstruction, _ = stakeInstruction.SetPublicKeys(strings.Split(instruction[1], SPLITTER))
	stakeInstruction.SetTxStakes(strings.Split(instruction[3], SPLITTER))
	stakeInstruction.SetRewardReceivers(strings.Split(instruction[4], SPLITTER))
	tempAutoStakes := strings.Split(instruction[5], SPLITTER)
	autoStakeFlags := []bool{}
	for _, v := range tempAutoStakes {
		if v == TRUE {
			autoStakeFlags = append(autoStakeFlags, true)
		} else {
			autoStakeFlags = append(autoStakeFlags, false)
		}
	}
	stakeInstruction.SetAutoStakingFlag(autoStakeFlags)
	stakeInstruction.SetChain(instruction[2])
	return stakeInstruction
}

// validate stake instruction sanity
// beaconprocess.go: 1122 - 1165
// beaconproducer.go: 386
func ValidateStakeInstructionSanity(instruction []string) error {
	if len(instruction) != 6 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != STAKE_ACTION {
		return fmt.Errorf("invalid stake action, %+v", instruction)
	}
	if instruction[2] != SHARD_INST && instruction[2] != BEACON_INST {
		return fmt.Errorf("invalid chain id, %+v", instruction)
	}
	publicKeys := strings.Split(instruction[1], SPLITTER)
	txStakes := strings.Split(instruction[3], SPLITTER)
	rewardReceivers := strings.Split(instruction[4], SPLITTER)
	autoStakings := strings.Split(instruction[5], SPLITTER)
	_, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return fmt.Errorf("invalid public key type,err %+v, %+v", err, instruction)
	}
	if len(publicKeys) != len(txStakes) {
		return fmt.Errorf("invalid public key & tx stake length, %+v", instruction)
	}
	if len(rewardReceivers) != len(txStakes) {
		return fmt.Errorf("invalid reward receivers & tx stake length, %+v", instruction)
	}
	if len(rewardReceivers) != len(autoStakings) {
		return fmt.Errorf("invalid reward receivers & tx auto staking length, %+v", instruction)
	}
	return nil
}
