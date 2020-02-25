package metadata

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
)

type StopAutoStakingMetadata struct {
	MetadataBase
	CommitteePublicKey string
}

func NewStopAutoStakingMetadata(stopStakingType int, committeePublicKey string) (*StopAutoStakingMetadata, error) {
	if stopStakingType != StopAutoStakingMeta {
		return nil, errors.New("invalid stop staking type")
	}
	metadataBase := NewMetadataBase(stopStakingType)
	return &StopAutoStakingMetadata{
		MetadataBase:       *metadataBase,
		CommitteePublicKey: committeePublicKey,
	}, nil
}

/*
 */
func (stopAutoStakingMetadata *StopAutoStakingMetadata) ValidateMetadataByItself() bool {
	CommitteePublicKey := new(incognitokey.CommitteePublicKey)
	if err := CommitteePublicKey.FromString(stopAutoStakingMetadata.CommitteePublicKey); err != nil {
		return false
	}
	if !CommitteePublicKey.CheckSanityData() {
		return false
	}
	return (stopAutoStakingMetadata.Type == StopAutoStakingMeta)
}

//ValidateTxWithBlockChain Validate Condition to Request Stop AutoStaking With Blockchain
//- Requested Committee Publickey is in candidate, pending validator,
//- Requested Committee Publickey is in staking tx list,
//- Requester (sender of tx) must be address, which create staking transaction for current requested committee public key
//- Not yet requested to stop auto-restaking
func (stopAutoStakingMetadata StopAutoStakingMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, stateDB *statedb.StateDB) (bool, error) {
	stopStakingMetadata, ok := txr.GetMetadata().(*StopAutoStakingMetadata)
	if !ok {
		return false, NewMetadataTxError(StopAutoStakingRequestTypeAssertionError, fmt.Errorf("Expect *StopAutoStakingMetadata type but get %+v", reflect.TypeOf(txr.GetMetadata())))
	}
	requestedPublicKey := stopStakingMetadata.CommitteePublicKey
	committees, err := bcr.GetAllCommitteeValidatorCandidateFlattenListFromDatabase()
	if err != nil {
		return false, NewMetadataTxError(StopAutoStakingRequestNotInCommitteeListError, err)
	}
	// if not found
	if !(common.IndexOfStr(requestedPublicKey, committees) > -1) {
		return false, NewMetadataTxError(StopAutoStakingRequestNotInCommitteeListError, fmt.Errorf("Committee Publickey %+v not found in any committee list of current beacon beststate", requestedPublicKey))
	}
	stakingTx := bcr.GetStakingTx(shardID)
	if tempStakingTxHash, ok := stakingTx[requestedPublicKey]; !ok {
		return false, NewMetadataTxError(StopAutoStakingRequestStakingTransactionNotFoundError, fmt.Errorf("No Committe Publickey %+v found in StakingTx of Shard %+v", requestedPublicKey, shardID))
	} else {
		stakingTxHash, err := common.Hash{}.NewHashFromStr(tempStakingTxHash)
		if err != nil {
			return false, err
		}
		_, _, _, stakingTx, err := bcr.GetTransactionByHash(*stakingTxHash)
		if err != nil {
			return false, NewMetadataTxError(StopAutoStakingRequestStakingTransactionNotFoundError, err)
		}
		if !bytes.Equal(stakingTx.GetSender(), txr.GetSender()) {
			return false, NewMetadataTxError(StopAutoStakingRequestInvalidTransactionSenderError, fmt.Errorf("Expect %+v to send stop auto staking request but get %+v", stakingTx.GetSender(), txr.GetSender()))
		}
	}
	autoStakingList := bcr.GetAutoStakingList()
	if isAutoStaking, ok := autoStakingList[stopStakingMetadata.CommitteePublicKey]; !ok {
		return false, NewMetadataTxError(StopAutoStakingRequestNoAutoStakingAvaiableError, fmt.Errorf("Committe Publickey %+v already request stop auto re-staking", stopStakingMetadata.CommitteePublicKey))
	} else {
		if !isAutoStaking {
			return false, NewMetadataTxError(StopAutoStakingRequestAlreadyStopError, fmt.Errorf("Auto Staking for Committee Public Key %+v already stop", stopAutoStakingMetadata.CommitteePublicKey))
		}
	}
	return true, nil
}

/*
	// Have only one receiver
	// Have only one amount corresponding to receiver
	// Receiver Is Burning Address
	//
*/
func (stopAutoStakingMetadata StopAutoStakingMetadata) ValidateSanityData(bcr BlockchainRetriever, txr Transaction, beaconHeight uint64) (bool, bool, error) {
	if txr.IsPrivacy() {
		return false, false, errors.New("Stop AutoStaking Request Transaction Is No Privacy Transaction")
	}
	onlyOne, pubkey, amount := txr.GetUniqueReceiver()
	if !onlyOne {
		return false, false, errors.New("staking Transaction Should Have 1 Output Amount crossponding to 1 Receiver")
	}

	// get burning address
	burningAddress := bcr.GetBurningAddress(beaconHeight)
	keyWalletBurningAdd, err := wallet.Base58CheckDeserialize(burningAddress)
	if err != nil {
		return false, false, err
	}
	if !bytes.Equal(pubkey, keyWalletBurningAdd.KeySet.PaymentAddress.Pk) {
		return false, false, errors.New("receiver Should be Burning Address")
	}
	if stopAutoStakingMetadata.Type != StopAutoStakingMeta && amount != StopAutoStakingAmount {
		return false, false, errors.New("receiver amount should be zero")
	}
	CommitteePublicKey := new(incognitokey.CommitteePublicKey)
	err = CommitteePublicKey.FromString(stopAutoStakingMetadata.CommitteePublicKey)
	if err != nil {
		return false, false, err
	}
	if !CommitteePublicKey.CheckSanityData() {
		return false, false, errors.New("Invalid Commitee Public Key of Candidate who join consensus")
	}
	return true, true, nil
}
func (stopAutoStakingMetadata StopAutoStakingMetadata) GetType() int {
	return stopAutoStakingMetadata.Type
}

func (stopAutoStakingMetadata *StopAutoStakingMetadata) CalculateSize() uint64 {
	return calculateSize(stopAutoStakingMetadata)
}
