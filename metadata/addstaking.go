package metadata

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/pkg/errors"
)

type AddStakingMetadata struct {
	MetadataBaseWithSignature
	CommitteePublicKey string
	AddStakingAmount   uint64
}

func (meta *AddStakingMetadata) Hash() *common.Hash {
	record := strconv.Itoa(meta.Type)
	data := []byte(record)
	data = append(data, meta.Sig...)
	data = append(data, []byte(meta.CommitteePublicKey)...)
	data = append(data, []byte(fmt.Sprintf("%v", meta.AddStakingAmount))...)
	hash := common.HashH(data)
	return &hash
}

func (meta *AddStakingMetadata) HashWithoutSig() *common.Hash {
	record := strconv.Itoa(meta.Type)
	data := []byte(record)
	data = append(data, []byte(meta.CommitteePublicKey)...)
	data = append(data, []byte(fmt.Sprintf("%v", meta.AddStakingAmount))...)
	hash := common.HashH(data)
	return &hash
}

func NewAddStakingMetadata(committeePublicKey string, addStakingAmount uint64) (*AddStakingMetadata, error) {
	metadataBase := NewMetadataBaseWithSignature(AddStakingMeta)
	return &AddStakingMetadata{
		MetadataBaseWithSignature: *metadataBase,
		CommitteePublicKey:        committeePublicKey,
		AddStakingAmount:          addStakingAmount,
	}, nil
}

/*
 */
func (addStakingMetadata *AddStakingMetadata) ValidateMetadataByItself() bool {
	CommitteePublicKey := new(incognitokey.CommitteePublicKey)
	if err := CommitteePublicKey.FromString(addStakingMetadata.CommitteePublicKey); err != nil {
		return false
	}
	if !CommitteePublicKey.CheckSanityData() {
		return false
	}
	if (addStakingMetadata.AddStakingAmount%common.SHARD_STAKING_AMOUNT != 0) || (addStakingMetadata.AddStakingAmount < (3 * common.SHARD_STAKING_AMOUNT)) {
		return false
	}
	return (addStakingMetadata.Type == AddStakingMeta)
}

// ValidateTxWithBlockChain Validate Condition to Request Stop AutoStaking With Blockchain
// - Requested Committee Publickey is in candidate, pending validator,
// - Requested Committee Publickey is in staking tx list,
// - Requester (sender of tx) must be address, which create staking transaction for current requested committee public key
// - Not yet requested to stop auto-restaking
func (addStakingMetadata AddStakingMetadata) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	addStakingMetadataFromTx, ok := tx.GetMetadata().(*AddStakingMetadata)
	if !ok {
		return false, NewMetadataTxError(ConsensusMetadataTypeAssertionError, fmt.Errorf("Expect *AddStakingMetadata type but get %+v", reflect.TypeOf(tx.GetMetadata())))
	}
	requestedPublicKey := addStakingMetadataFromTx.CommitteePublicKey

	stakerInfo, has, err := statedb.GetBeaconStakerInfo(beaconViewRetriever.GetBeaconConsensusStateDB(), requestedPublicKey)
	if err != nil {
		return false, NewMetadataTxError(AddStakingRequestNotInCommitteeListError, err)
	}
	if !has {
		return false, NewMetadataTxError(AddStakingCommitteeNotFoundError, fmt.Errorf("No Committee Publickey %+v found", requestedPublicKey))
	}

	lockStaker := beaconViewRetriever.GetBeaconLocking()
	lockStakerStr, _ := incognitokey.CommitteeKeyListToString(lockStaker)
	if common.IndexOfStr(requestedPublicKey, lockStakerStr) != -1 {
		return false, NewMetadataTxError(AddStakingRequestNotInCommitteeListError, fmt.Errorf("Committee Publickey %+v is in locking state", requestedPublicKey))
	}

	if ok, err := addStakingMetadataFromTx.MetadataBaseWithSignature.VerifyMetadataSignature(stakerInfo.FunderAddress().Pk, tx); !ok || err != nil {
		return false, NewMetadataTxError(ConsensusMetadataInvalidTransactionSenderError, fmt.Errorf("CheckAuthorizedSender fail"))
	}
	return true, nil
}

// Have only one receiver
// Have only one amount corresponding to receiver
// Receiver Is Burning Address
func (addStakingMetadata AddStakingMetadata) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	isBurned, burnCoin, tokenID, err := tx.GetTxBurnData()
	if err != nil {
		return false, false, errors.New("Error Cannot get burn data from tx")
	}
	if !isBurned {
		return false, false, errors.New("Error AddStaking tx should be a burn tx")
	}
	if !bytes.Equal(tokenID[:], common.PRVCoinID[:]) {
		return false, false, errors.New("Error AddStaking tx should transfer PRV only")
	}
	if (addStakingMetadata.Type != AddStakingMeta) || (burnCoin.GetValue()%common.SHARD_STAKING_AMOUNT != 0) || (burnCoin.GetValue() < common.SHARD_STAKING_AMOUNT*3) || (burnCoin.GetValue() != addStakingMetadata.AddStakingAmount) {
		return false, false, errors.Errorf("receiver amount invalid: %v, addStakeMeta %+v ", fmt.Sprint(burnCoin.GetValue()), addStakingMetadata)
	}
	CommitteePublicKey := new(incognitokey.CommitteePublicKey)
	if err := CommitteePublicKey.FromString(addStakingMetadata.CommitteePublicKey); err != nil {
		return false, false, err
	}
	if !CommitteePublicKey.CheckSanityData() {
		return false, false, errors.New("Invalid Committee Public Key of Candidate who join consensus")
	}
	return true, true, nil
}

func (addStakingMetadata AddStakingMetadata) GetType() int {
	return addStakingMetadata.Type
}

func (addStakingMetadata *AddStakingMetadata) CalculateSize() uint64 {
	return calculateSize(addStakingMetadata)
}

func (req *AddStakingMetadata) ToCompactBytes() ([]byte, error) {
	return metadataCommon.ToCompactBytes(req)
}
