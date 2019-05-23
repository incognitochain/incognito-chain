package metadata

import (
	"bytes"
	"errors"
	"strings"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/wallet"
)

type StakingMetadata struct {
	MetadataBase
	PaymentAddress string
}

func NewStakingMetadata(stakingType int, paymentAdd string) (*StakingMetadata, error) {
	if stakingType != ShardStakingMeta && stakingType != BeaconStakingMeta {
		return nil, errors.New("invalid staking type")
	}
	metadataBase := NewMetadataBase(stakingType)

	return &StakingMetadata{*metadataBase, paymentAdd}, nil
}

/*
 */
func (sm *StakingMetadata) ValidateMetadataByItself() bool {
	return (sm.Type == ShardStakingMeta || sm.Type == BeaconStakingMeta)
}

func (sm *StakingMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	SC, SPV, BC, BPV, CBWFCR, CBWFNR, CSWFCR, CSWFNR := bcr.GetAllCommitteeValidatorCandidate()
	senderPubkeyString := base58.Base58Check{}.Encode(txr.GetSigPubKey(), common.ZeroByte)
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
		return false, errors.New("invalid Staker, This pubkey may staked already")
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
		return false, false, errors.New("staking Transaction Is No Privacy Transaction")
	}
	onlyOne, pubkey, amount := txr.GetUniqueReceiver()

	if !onlyOne {
		return false, false, errors.New("staking Transaction Should Have 1 Output Amount crossponding to 1 Receiver")
	}
	keyWalletBurningAdd, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	if !bytes.Equal(pubkey, keyWalletBurningAdd.KeySet.PaymentAddress.Pk) {
		return false, false, errors.New("receiver Should be Burning Address")
	}
	if sm.Type == ShardStakingMeta && amount != GetShardStateAmount() {
		return false, false, errors.New("invalid Stake Shard Amount")
	}
	if sm.Type == BeaconStakingMeta && amount != GetBeaconStakeAmount() {
		return false, false, errors.New("invalid Stake Beacon Amount")
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

func GetBeaconStakeAmount() uint64 {
	const STAKE_BEACON_AMOUNT = 20000
	return STAKE_BEACON_AMOUNT
}

func GetShardStateAmount() uint64 {
	const STAKE_SHARD_AMOUNT = 100
	return STAKE_SHARD_AMOUNT
}
