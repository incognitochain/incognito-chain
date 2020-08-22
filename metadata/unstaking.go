package metadata

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

//UnStakingMetadata : unstaking metadata
type UnStakingMetadata struct {
	MetadataBase
	CommitteePublicKey string
}

//NewUnStakingMetadata : Constructor of UnStakingMetadata struct
func NewUnStakingMetadata(unStakingType int, committeePublicKey string) (*UnStakingMetadata, error) {
	if unStakingType != UnStakingMeta {
		return nil, errors.New("invalid stop staking type")
	}
	metadataBase := NewMetadataBase(unStakingType)
	return &UnStakingMetadata{
		MetadataBase:       *metadataBase,
		CommitteePublicKey: committeePublicKey,
	}, nil
}

//ValidateMetadataByItself Validate data format/type in unStakingMetadata
func (unStakingMetadata *UnStakingMetadata) ValidateMetadataByItself() bool {
	CommitteePublicKey := new(incognitokey.CommitteePublicKey)
	if err := CommitteePublicKey.FromString(unStakingMetadata.CommitteePublicKey); err != nil {
		return false
	}
	if !CommitteePublicKey.CheckSanityData() {
		return false
	}
	return (unStakingMetadata.Type == UnStakingMeta)
}

//ValidateTxWithBlockChain Validate Condition to Request Unstake With Blockchain
//- Requested Committee Publickey is in candidate, pending validator,
//- Requested Committee Publickey is in staking tx list,
//- Requester (sender of tx) must be address, which create staking transaction for current requested committee public key
func (unStakingMetadata UnStakingMetadata) ValidateTxWithBlockChain(tx Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {

	unStakeMetadata, ok := tx.GetMetadata().(*UnStakingMetadata)
	if !ok {
		return false, NewMetadataTxError(UnStakingRequestTypeAssertionError, fmt.Errorf("Expect *UnStakingMetadata type but get %+v", reflect.TypeOf(tx.GetMetadata())))
	}
	requestedPublicKey := unStakeMetadata.CommitteePublicKey
	committees, err := beaconViewRetriever.GetAllCommitteeValidatorCandidateFlattenListFromDatabase()
	if err != nil {
		return false, NewMetadataTxError(UnStakingRequestNotInCommitteeListError, err)
	}
	// if not found
	if !(common.IndexOfStr(requestedPublicKey, committees) > -1) {
		return false, NewMetadataTxError(UnStakingRequestNotInCommitteeListError, fmt.Errorf("Committee Publickey %+v not found in any committee list of current beacon beststate", requestedPublicKey))
	}

	_, has, err := statedb.GetStakerInfo(beaconViewRetriever.GetBeaconConsensusStateDB(), requestedPublicKey)
	if err != nil {
		Logger.log.Error(err)
		return false, NewMetadataTxError(UnStakingRequestGetStakerInfoError, err)
	}

	if !has {
		return false, NewMetadataTxError(UnStakingRequestNotFoundStakerInfoError, fmt.Errorf("Committee Publickey %+v has not staked yet", requestedPublicKey))
	}

	stakingTx, err := chainRetriever.GetShardStakingTx(shardViewRetriever.GetShardID(), shardViewRetriever.GetBeaconHeight())
	if err != nil {
		return false, NewMetadataTxError(UnStakingRequestNotInCommitteeListError, err)
	}
	if tempStakingTxHash, ok := stakingTx[requestedPublicKey]; ok {
		stakingTxHash, err := common.Hash{}.NewHashFromStr(tempStakingTxHash)
		if err != nil {
			return false, err
		}
		_, _, _, _, stakingTx, err := chainRetriever.GetTransactionByHash(*stakingTxHash)
		if err != nil {
			return false, NewMetadataTxError(UnStakingRequestStakingTransactionNotFoundError, err)
		}
		if !bytes.Equal(stakingTx.GetSender(), tx.GetSender()) {
			return false, NewMetadataTxError(UnStakingRequestInvalidTransactionSenderError, fmt.Errorf("Expect %+v to send stop auto staking request but get %+v", stakingTx.GetSender(), tx.GetSender()))
		}
	} else {
		return false, NewMetadataTxError(UnStakingRequestStakingTransactionNotFoundError, fmt.Errorf("No Committe Publickey %+v found in StakingTx of Shard %+v", requestedPublicKey, shardID))
	}

	autoStakingList := beaconViewRetriever.GetAutoStakingList()
	ok = false
	isAutoStaking := false
	if isAutoStaking, ok = autoStakingList[unStakingMetadata.CommitteePublicKey]; !ok {
		return false, NewMetadataTxError(UnStakingRequestNoAutoStakingAvaiableError, fmt.Errorf("Committe Publickey %+v already request stop auto re-staking", unStakingMetadata.CommitteePublicKey))
	}
	if !isAutoStaking {
		return false, NewMetadataTxError(UnStakingRequestAlreadyStopError, fmt.Errorf("Auto Staking for Committee Public Key %+v already stop", unStakingMetadata.CommitteePublicKey))
	}

	return true, nil
}

// ValidateSanityData :
// Have only one receiver
// Have only one amount corresponding to receiver
// Receiver Is Burning Address
func (unStakingMetadata UnStakingMetadata) ValidateSanityData(
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	if tx.IsPrivacy() {
		return false, false, errors.New("Stop AutoStaking Request Transaction Is No Privacy Transaction")
	}

	if unStakingMetadata.Type != UnStakingMeta {
		return false, false, errors.New("receiver amount should be zero")
	}
	CommitteePublicKey := new(incognitokey.CommitteePublicKey)
	err := CommitteePublicKey.FromString(unStakingMetadata.CommitteePublicKey)
	if err != nil {
		return false, false, err
	}
	if !CommitteePublicKey.CheckSanityData() {
		return false, false, errors.New("Invalid Commitee Public Key of Candidate who join consensus")
	}
	return true, true, nil
}

//GetType :
func (unStakingMetadata UnStakingMetadata) GetType() int {
	return unStakingMetadata.Type
}

//CalculateSize :
func (unStakingMetadata *UnStakingMetadata) CalculateSize() uint64 {
	return calculateSize(unStakingMetadata)
}
