package metadata

import (
	"bytes"
	"errors"
	"strings"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/wallet"
)

type StakingMetadata struct {
	*MetadataBase
}

func NewStakingMetadata(stakingType int) (*StakingMetadata, error) {
	if stakingType != ShardStakingMeta && stakingType != BeaconStakingMeta {
		return nil, errors.New("Invalid staking type")
	}
	metadataBase := NewMetadataBase(stakingType)

	return &StakingMetadata{metadataBase}, nil
}

/*
 */
func (sm *StakingMetadata) ValidateMetadataByItself() bool {
	if !(sm.Type == ShardStakingMeta || sm.Type == BeaconStakingMeta) {
		return false
	}
	return true
}
func (sm *StakingMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	SC, SPV, BC, BPV, CBWFCR, CBWFNR, CSWFCR, CSWFNR := bcr.GetAllCommitteeValidatorCandidate()
	senderPubkeyString := base58.Base58Check{}.Encode(txr.GetSigPubKey(), byte(0x00))
	tempStaker := []string{senderPubkeyString}
	for _, committees := range SC {
		tempStaker = GetValidStaker(committees, tempStaker)
	}
	for _, validators := range SPV {
		tempStaker = GetValidStaker(validators, tempStaker)
	}
	tempStaker = GetValidStaker(BC, tempStaker)
	tempStaker = GetValidStaker(BPV, tempStaker)
	tempStaker = GetValidStaker(CBWFCR, tempStaker)
	tempStaker = GetValidStaker(CBWFNR, tempStaker)
	tempStaker = GetValidStaker(CSWFCR, tempStaker)
	tempStaker = GetValidStaker(CSWFNR, tempStaker)
	if len(tempStaker) == 0 {
		return false, errors.New("Invalid Staker, This pubkey may staked already")
	}
	return true, nil
}

/*
	// Have only one receiver
	// Have only one amount corresponding to receiver
	// Receiver Is Burning Address
	//
*/
func (sm *StakingMetadata) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if txr.IsPrivacy() {
		return false, false, errors.New("Staking Transaction Is No Privacy Transaction")
	}
	onlyOne, pubkey, amount := txr.GetUniqueReceiver()

	if !onlyOne {
		return false, false, errors.New("Staking Transaction Should Have 1 Output Amount crossponding to 1 Receiver")
	}
	keyWalletBurningAdd, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	if bytes.Compare(pubkey, keyWalletBurningAdd.KeySet.PaymentAddress.Pk) != 0 {
		return false, false, errors.New("Receiver Should be Burning Address")
	}
	if sm.Type == ShardStakingMeta && amount != STAKE_SHARD_AMOUNT {
		return false, false, errors.New("Invalid Stake Shard Amount")
	}
	if sm.Type == BeaconStakingMeta && amount != STAKE_BEACON_AMOUNT {
		return false, false, errors.New("Invalid Stake Beacon Amount")
	}
	return true, true, nil
}
func (sm *StakingMetadata) GetType() int {
	return sm.Type
}
func GetValidStaker(committees []string, stakers []string) []string {
	validStaker := []string{}
	for _, staker := range stakers {
		flag := false
		for _, committee := range committees {
			if strings.Compare(staker, committee) == 0 {
				flag = true
				break
			}
		}
		if !flag {
			validStaker = append(validStaker, staker)
		}
	}
	return validStaker
}

func (sm *StakingMetadata) CalculateSize() uint64 {
	return calculateSize(sm)
}
