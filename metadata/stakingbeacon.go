package metadata

import (
	"bytes"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type PromoteBeaconMetadata struct {
	MetadataBase
	FunderPaymentAddress         string
	RewardReceiverPaymentAddress string
	StakingAmount                uint64
	CommitteePublicKey           string
}

func NewPromoteBeaconMetadata(
	stakingType int,
	funderPaymentAddress string,
	rewardReceiverPaymentAddress string,
	stakingAmount uint64,
	committeePublicKey string,
) (
	*PromoteBeaconMetadata,
	error,
) {
	if (stakingType != ShardStakingMeta) && (stakingType != BeaconStakingMeta) {
		return nil, errors.New("invalid staking type")
	}
	metadataBase := NewMetadataBase(stakingType)
	return &PromoteBeaconMetadata{
		MetadataBase:                 *metadataBase,
		FunderPaymentAddress:         funderPaymentAddress,
		RewardReceiverPaymentAddress: rewardReceiverPaymentAddress,
		StakingAmount:                stakingAmount,
		CommitteePublicKey:           committeePublicKey,
	}, nil
}

// + no need to IsInBase58ShortFormat because error is already check below by FromString
// + what IsInBase58ShortFormat does is the same as FromString does but for an array
func (sm *PromoteBeaconMetadata) ValidateMetadataByItself() bool {
	rewardReceiverPaymentAddress := sm.RewardReceiverPaymentAddress
	rewardReceiverWallet, err := wallet.Base58CheckDeserialize(rewardReceiverPaymentAddress)
	if err != nil || rewardReceiverWallet == nil {
		return false
	}
	if !incognitokey.IsInBase58ShortFormat([]string{sm.CommitteePublicKey}) {
		return false
	}
	CommitteePublicKey := new(incognitokey.CommitteePublicKey)
	if err := CommitteePublicKey.FromString(sm.CommitteePublicKey); err != nil {
		return false
	}
	if !CommitteePublicKey.CheckSanityData() {
		return false
	}
	return (sm.Type == BeaconStakingMeta)
}

// TODO modify validate function
func (stakingMetadata PromoteBeaconMetadata) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	/* Check if beacon already promote
	- get all beacon validator
	- make sure there is no duplicated mining public key
	*/

	/* Check if promoting is valid
	- check if shard validator exist
	*/

	return true, nil
}

// ValidateSanityData
// Have only one receiver
// Have only one amount corresponding to receiver
// Receiver Is Burning Address
func (stakingMetadata PromoteBeaconMetadata) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	if tx.IsPrivacy() {
		return false, false, errors.New("staking Transaction Is No Privacy Transaction")
	}
	isBurned, burnCoin, tokenID, err := tx.GetTxBurnData()
	if err != nil {
		return false, false, errors.New("Error Cannot get burn data from tx")
	}
	if !isBurned {
		return false, false, errors.New("Error Staking tx should be a burn tx")
	}
	if !bytes.Equal(tokenID[:], common.PRVCoinID[:]) {
		return false, false, errors.New("Error Staking tx should transfer PRV only")
	}
	amount := burnCoin.GetValue()

	if stakingMetadata.Type == BeaconStakingMeta && amount < config.Param().StakingAmountBeacon {
		return false, false, errors.Errorf("invalid Stake Beacon Amount, got %v, required >= %v", amount, config.Param().StakingAmountBeacon)
	}

	if _, err := AssertPaymentAddressAndTxVersion(stakingMetadata.FunderPaymentAddress, tx.GetVersion()); err != nil {
		return false, false, errors.New(fmt.Sprintf("invalid funder address: %v", err))
	}

	if _, err := AssertPaymentAddressAndTxVersion(stakingMetadata.RewardReceiverPaymentAddress, tx.GetVersion()); err != nil {
		return false, false, errors.New(fmt.Sprintf("invalid reward receiver address: %v", err))
	}

	CommitteePublicKey := new(incognitokey.CommitteePublicKey)
	err = CommitteePublicKey.FromString(stakingMetadata.CommitteePublicKey)
	if err != nil {
		return false, false, err
	}
	if !CommitteePublicKey.CheckSanityData() {
		return false, false, errors.New("Invalid Commitee Public Key of Candidate who join consensus")
	}

	return true, true, nil
}
func (stakingMetadata PromoteBeaconMetadata) GetType() int {
	return stakingMetadata.Type
}

func (stakingMetadata *PromoteBeaconMetadata) CalculateSize() uint64 {
	return calculateSize(stakingMetadata)
}

func (stakingMetadata PromoteBeaconMetadata) GetBeaconStakeAmount() uint64 {
	return stakingMetadata.StakingAmount
}
