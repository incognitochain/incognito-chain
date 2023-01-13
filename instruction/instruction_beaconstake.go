package instruction

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/common/math"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/log/proto"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

// BeaconStakeInstruction :
// BeaconStakeInstruction Format:
// ["STAKE_ACTION", list_public_keys, chain or beacon, list_txs, list_reward_addresses, list_autostaking_status(boolean)]
type BeaconStakeInstruction struct {
	PublicKeys            []string
	PublicKeyStructs      []incognitokey.CommitteePublicKey
	Chain                 string
	TxStakes              []string
	TxStakeHashes         []common.Hash
	RewardReceivers       []string
	RewardReceiverStructs []privacy.PaymentAddress
	AutoStakingFlag       []bool
	StakingAmount         []uint64
	FunderAddress         []string
	FunderAddressStructs  []privacy.PaymentAddress
	instructionBase
}

func NewBeaconStakeInstruction() *BeaconStakeInstruction {
	return &BeaconStakeInstruction{
		instructionBase: instructionBase{
			featureID: proto.FID_CONSENSUS_BEACON,
			logOnly:   false,
		},
	}
}

func (s *BeaconStakeInstruction) IsEmpty() bool {
	return reflect.DeepEqual(s, NewBeaconStakeInstruction()) ||
		len(s.PublicKeyStructs) == 0 && len(s.PublicKeys) == 0
}

func (s *BeaconStakeInstruction) GetType() string {
	return BEACON_STAKE_ACTION
}

func (s *BeaconStakeInstruction) SetPublicKeys(publicKeys []string) (*BeaconStakeInstruction, error) {
	s.PublicKeys = publicKeys
	publicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return nil, err
	}
	s.PublicKeyStructs = publicKeyStructs
	return s, nil
}

func (s *BeaconStakeInstruction) SetChain(chain string) *BeaconStakeInstruction {
	s.Chain = chain
	return s
}

func (s *BeaconStakeInstruction) SetTxStakes(txStakes []string) *BeaconStakeInstruction {
	s.TxStakes = txStakes
	for _, txStake := range txStakes {
		res, _ := common.Hash{}.NewHashFromStr(txStake)
		s.TxStakeHashes = append(s.TxStakeHashes, *res)
	}
	return s
}

func (s *BeaconStakeInstruction) SetFunderAddress(funderAddress []string) *BeaconStakeInstruction {
	s.FunderAddress = funderAddress
	for _, v := range funderAddress {
		wl, _ := wallet.Base58CheckDeserialize(v)
		s.FunderAddressStructs = append(s.FunderAddressStructs, wl.KeySet.PaymentAddress)
	}
	return s
}

func (s *BeaconStakeInstruction) SetRewardReceivers(rewardReceivers []string) *BeaconStakeInstruction {
	s.RewardReceivers = rewardReceivers
	for _, v := range rewardReceivers {
		wl, _ := wallet.Base58CheckDeserialize(v)
		s.RewardReceiverStructs = append(s.RewardReceiverStructs, wl.KeySet.PaymentAddress)
	}
	return s
}

func (s *BeaconStakeInstruction) SetAutoStakingFlag(autoStakingFlag []bool) *BeaconStakeInstruction {
	s.AutoStakingFlag = autoStakingFlag
	return s
}

func (s *BeaconStakeInstruction) ToString() []string {
	stakeInstructionStr := []string{BEACON_STAKE_ACTION}
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

	amountStr := []string{}
	for _, amount := range s.StakingAmount {
		amountStr = append(amountStr, fmt.Sprintf("%v", amount))
	}
	stakeInstructionStr = append(stakeInstructionStr, strings.Join(amountStr, SPLITTER))

	stakeInstructionStr = append(stakeInstructionStr, strings.Join(s.FunderAddress, SPLITTER))

	return stakeInstructionStr
}

func ValidateAndImportBeaconStakeInstructionFromString(instruction []string) (*BeaconStakeInstruction, error) {
	if err := ValidateBeaconStakeInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportBeaconStakeInstructionFromString(instruction), nil
}

// ImportBeaconStakeInstructionFromString is unsafe method
func ImportBeaconStakeInstructionFromString(instruction []string) *BeaconStakeInstruction {
	stakeInstruction := NewBeaconStakeInstruction()
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

	amount := []uint64{}
	amountStr := strings.Split(instruction[6], SPLITTER)
	for _, amounts := range amountStr {
		a, _ := math.ParseUint64(amounts)
		amount = append(amount, a)
	}
	stakeInstruction.StakingAmount = amount
	stakeInstruction.SetFunderAddress(strings.Split(instruction[7], SPLITTER))

	return stakeInstruction
}

// validate stake instruction sanity
// beaconprocess.go: 1122 - 1165
// beaconproducer.go: 386
func ValidateBeaconStakeInstructionSanity(instruction []string) error {
	if instruction[2] != BEACON_INST {
		return fmt.Errorf("invalid chain id, %+v", instruction)
	}

	if len(instruction) != 8 {
		log.Println(len(instruction), instruction[2])
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != BEACON_STAKE_ACTION {
		return fmt.Errorf("invalid stake action, %+v", instruction)
	}

	publicKeys := strings.Split(instruction[1], SPLITTER)
	txStakes := strings.Split(instruction[3], SPLITTER)
	for _, txStake := range txStakes {
		_, err := common.Hash{}.NewHashFromStr(txStake)
		if err != nil {
			return fmt.Errorf("invalid tx stake %+v", err)
		}
	}
	rewardReceivers := strings.Split(instruction[4], SPLITTER)
	for _, v := range rewardReceivers {
		_, err := wallet.Base58CheckDeserialize(v)
		if err != nil {
			return fmt.Errorf("invalid privacy payment address %+v", err)
		}
	}
	autoStakes := strings.Split(instruction[5], SPLITTER)
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
	if len(rewardReceivers) != len(autoStakes) {
		return fmt.Errorf("invalid reward receivers & tx auto staking length, %+v", instruction)
	}
	return nil
}

// ImportBeaconStakeInstructionFromString is unsafe method
func ImportInitBeaconStakeInstructionFromString(instruction []string) *BeaconStakeInstruction {
	stakeInstruction := NewBeaconStakeInstruction()
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
