package metadata

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
)

type StakingMetadata struct {
	MetadataBase
	FunderPaymentAddress         string
	RewardReceiverPaymentAddress string
	StakingAmountShard           uint64
	AutoReStaking                bool
	CommitteePublicKey           string
	Delegate                     string
	// CommitteePublicKey CommitteePublicKeys of a candidate who join consensus, base58CheckEncode
	// CommitteePublicKey string <= encode byte <= mashal struct
}

func NewStakingMetadata(
	stakingType int,
	funderPaymentAddress string,
	rewardReceiverPaymentAddress string,
	// candidatePaymentAddress string,
	stakingAmountShard uint64,
	committeePublicKey string,
	delegatePublicKey string,
	autoReStaking bool,
) (
	*StakingMetadata,
	error,
) {
	if (stakingType != ShardStakingMeta) && (stakingType != BeaconStakingMeta) {
		return nil, errors.New("invalid staking type")
	}
	metadataBase := NewMetadataBase(stakingType)
	return &StakingMetadata{
		MetadataBase:                 *metadataBase,
		FunderPaymentAddress:         funderPaymentAddress,
		RewardReceiverPaymentAddress: rewardReceiverPaymentAddress,
		StakingAmountShard:           stakingAmountShard,
		CommitteePublicKey:           committeePublicKey,
		AutoReStaking:                autoReStaking,
		Delegate:                     delegatePublicKey,
	}, nil
}

// + no need to IsInBase58ShortFormat because error is already check below by FromString
// + what IsInBase58ShortFormat does is the same as FromString does but for an array
func (sm *StakingMetadata) ValidateMetadataByItself() bool {
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
	if sm.Delegate != "" {
		if !incognitokey.IsInBase58ShortFormat([]string{sm.Delegate}) {
			return false
		}
		if err := CommitteePublicKey.FromString(sm.Delegate); err != nil {
			return false
		}
		if !CommitteePublicKey.CheckSanityData() {
			return false
		}
	}
	// only stake to shard
	return (sm.Type == ShardStakingMeta) || (sm.Type == BeaconStakingMeta)
}

// TODO modify validate function
func (stakingMetadata StakingMetadata) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	SC, SPV, sSP, BC, BPV, CBWFCR, CBWFNR, CSWFCR, CSWFNR, err := beaconViewRetriever.GetAllCommitteeValidatorCandidate()
	if err != nil {
		return false, err
	}
	tempStaker, err := incognitokey.CommitteeBase58KeyListToStruct([]string{stakingMetadata.CommitteePublicKey})
	if err != nil {
		return false, err
	}
	for _, committees := range SC {
		tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(committees, tempStaker)
	}
	for _, validators := range SPV {
		tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(validators, tempStaker)
	}
	for _, syncValidators := range sSP {
		tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(syncValidators, tempStaker)
	}
	tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(BC, tempStaker)
	tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(BPV, tempStaker)
	tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(CBWFCR, tempStaker)
	tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(CBWFNR, tempStaker)
	tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(CSWFCR, tempStaker)
	tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(CSWFNR, tempStaker)
	if len(tempStaker) == 0 {
		return false, errors.New("invalid Staker, This pubkey may staked already")
	}
	return true, nil
}

// ValidateSanityData
// Have only one receiver
// Have only one amount corresponding to receiver
// Receiver Is Burning Address
func (stakingMetadata StakingMetadata) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
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
	if stakingMetadata.Type == ShardStakingMeta && amount != config.Param().StakingAmountShard {
		return false, false, errors.New("invalid Stake Shard Amount")
	}
	// if stakingMetadata.Type == BeaconStakingMeta && amount != config.Param().StakingAmountShard*3 {
	// 	return false, false, errors.New("invalid Stake Beacon Amount")
	// }

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
func (stakingMetadata StakingMetadata) GetType() int {
	return stakingMetadata.Type
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
