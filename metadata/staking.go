package metadata

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
)

type StakingMetadata struct {
	MetadataBase
	FunderPaymentAddress         string
	RewardReceiverPaymentAddress string
	StakingAmount                uint64
	AutoReStaking                bool
	CommitteePublicKey           string
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
	autoReStaking bool,
) (
	*StakingMetadata,
	error,
) {
	if stakingType != ShardStakingMeta && stakingType != BeaconStakingMeta {
		return nil, errors.New("invalid staking type")
	}
	metadataBase := NewMetadataBase(stakingType)
	return &StakingMetadata{
		MetadataBase:                 *metadataBase,
		FunderPaymentAddress:         funderPaymentAddress,
		RewardReceiverPaymentAddress: rewardReceiverPaymentAddress,
		StakingAmount:                stakingAmountShard,
		CommitteePublicKey:           committeePublicKey,
		AutoReStaking:                autoReStaking,
	}, nil
}

func (meta *StakingMetadata) Hash() *common.Hash {
	if meta.Type == ShardStakingMeta {
		return meta.MetadataBase.Hash()
	}
	record := strconv.Itoa(meta.Type)
	data := []byte(record)
	data = append(data, []byte(meta.FunderPaymentAddress)...)
	data = append(data, []byte(meta.RewardReceiverPaymentAddress)...)
	data = append(data, []byte(fmt.Sprintf("%v", meta.StakingAmount))...)
	data = append(data, []byte(meta.CommitteePublicKey)...)
	data = append(data, []byte(fmt.Sprintf("%v", meta.AutoReStaking))...)
	hash := common.HashH(data)
	return &hash
}

func (meta *StakingMetadata) HashWithoutSig() *common.Hash {
	return meta.Hash()
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
	return sm.Type == ShardStakingMeta || sm.Type == BeaconStakingMeta
}

func (stakingMetadata StakingMetadata) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	SC, SPV, sSP, BC, BPV, BW, BL, CSWFCR, CSWFNR := beaconViewRetriever.GetAllStakers()
	tempStaker, err := incognitokey.CommitteeBase58KeyListToStruct([]string{stakingMetadata.CommitteePublicKey})
	if err != nil {
		return false, err
	}
	switch stakingMetadata.Type {
	case ShardStakingMeta:
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
		tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(BW, tempStaker)
		tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(BL, tempStaker)
		tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(CSWFCR, tempStaker)
		tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(CSWFNR, tempStaker)
		if len(tempStaker) == 0 {
			return false, errors.New("invalid Shard Staker, This pubkey may staked already")
		}
	case BeaconStakingMeta:
		/* Check if beacon already promote
		- get all beacon validator
		- make sure there is no duplicated mining public key
		*/
		tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(BC, tempStaker)
		if len(tempStaker) == 0 {
			return false, errors.New("invalid Beacon Staker, This pubkey may staked already. In BC")
		}
		tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(BPV, tempStaker)
		if len(tempStaker) == 0 {
			return false, errors.New("invalid Beacon Staker, This pubkey may staked already. In BP")
		}
		tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(BW, tempStaker)
		if len(tempStaker) == 0 {
			return false, errors.New("invalid Beacon Staker, This pubkey may staked already. In BW")
		}
		tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(BL, tempStaker)
		if len(tempStaker) == 0 {
			return false, errors.New("invalid Beacon Staker, This pubkey may staked already. In BL")
		}
		/* Check if promoting is valid
		- check if shard validator exist
		*/
		for _, committees := range SC {
			tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(committees, tempStaker)
		}
		for _, validators := range SPV {
			tempStaker = incognitokey.GetValidStakeStructCommitteePublicKey(validators, tempStaker)
		}
		if len(tempStaker) != 0 {
			return false, errors.New("invalid Staker, This pubkey is not in any shard committee or validator")
		}

	default:
		return false, errors.New("Not recognized staking type")
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
	if stakingMetadata.Type == ShardStakingMeta && amount != common.SHARD_STAKING_AMOUNT {
		return false, false, errors.New("invalid Stake Shard Amount")
	}
	if stakingMetadata.Type == BeaconStakingMeta && ((amount < common.BEACON_MIN_STAKING_AMOUNT) || (amount%common.SHARD_STAKING_AMOUNT != 0)) {
		return false, false, errors.New("invalid Stake Beacon Amount")
	}
	if stakingMetadata.Type == BeaconStakingMeta && !stakingMetadata.AutoReStaking {
		return false, false, errors.New("staking beacon must always set restaking")
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
func (stakingMetadata StakingMetadata) GetType() int {
	return stakingMetadata.Type
}

func (stakingMetadata *StakingMetadata) CalculateSize() uint64 {
	return calculateSize(stakingMetadata)
}

func (stakingMetadata StakingMetadata) GetBeaconStakeAmount() uint64 {
	return stakingMetadata.StakingAmount * 3
}

func (stakingMetadata StakingMetadata) GetShardStateAmount() uint64 {
	return stakingMetadata.StakingAmount
}
