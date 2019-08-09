package metadata

import (
	"bytes"
	"errors"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/wallet"
)

type StakingMetadata struct {
	MetadataBase
	FunderPaymentAddress    string
	CandidatePaymentAddress string
	StakingAmountShard      uint64
}

func NewStakingMetadata(stakingType int, funderPaymentAddress string, candidatePaymentAddress string, stakingAmountShard uint64) (*StakingMetadata, error) {
	if stakingType != ShardStakingMeta && stakingType != BeaconStakingMeta {
		return nil, errors.New("invalid staking type")
	}
	metadataBase := NewMetadataBase(stakingType)
	return &StakingMetadata{
		MetadataBase:            *metadataBase,
		FunderPaymentAddress:    funderPaymentAddress,
		CandidatePaymentAddress: candidatePaymentAddress,
		StakingAmountShard:      stakingAmountShard,
	}, nil
}

/*
 */
func (stakingMetadata *StakingMetadata) ValidateMetadataByItself() bool {
	return (stakingMetadata.Type == ShardStakingMeta || stakingMetadata.Type == BeaconStakingMeta)
}

func (stakingMetadata *StakingMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	SC, SPV, BC, BPV, CBWFCR, CBWFNR, CSWFCR, CSWFNR := bcr.GetAllCommitteeValidatorCandidate()
	candidatePaymentAddress := stakingMetadata.CandidatePaymentAddress
	candidateWallet, err := wallet.Base58CheckDeserialize(candidatePaymentAddress)
	if err != nil || candidateWallet == nil {
		return false, errors.New("Can create wallet key from payment address")
	}
	pk := candidateWallet.KeySet.PaymentAddress.Pk
	senderPubkeyString := base58.Base58Check{}.Encode(pk, common.ZeroByte)
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
func (stakingMetadata *StakingMetadata) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
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
	if stakingMetadata.Type == ShardStakingMeta && amount != bcr.GetStakingAmountShard() {
		return false, false, errors.New("invalid Stake Shard Amount")
	}
	if stakingMetadata.Type == BeaconStakingMeta && amount != bcr.GetStakingAmountShard()*3 {
		return false, false, errors.New("invalid Stake Beacon Amount")
	}
	return true, true, nil
}
func (stakingMetadata *StakingMetadata) GetType() int {
	return stakingMetadata.Type
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

func (stakingMetadata *StakingMetadata) CalculateSize() uint64 {
	return calculateSize(stakingMetadata)
}

func (stakingMetadata StakingMetadata) GetBeaconStakeAmount() uint64 {
	return stakingMetadata.StakingAmountShard * 3
}

func (stakingMetadata StakingMetadata) GetShardStateAmount() uint64 {
	return stakingMetadata.StakingAmountShard
}
