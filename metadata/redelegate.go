package metadata

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type ReDelegateMetadata struct {
	MetadataBaseWithSignature
	CommitteePublicKey string
	NewDelegate        string
	DelegateUID        string
}

func (meta *ReDelegateMetadata) Hash() *common.Hash {
	record := strconv.Itoa(meta.Type)
	data := []byte(record)
	data = append(data, meta.Sig...)
	hash := common.HashH(data)
	return &hash
}

func (meta *ReDelegateMetadata) HashWithoutSig() *common.Hash {
	return meta.MetadataBase.Hash()
}

func NewReDelegateMetadata(committeePublicKey, newDelegate string, newDelegateUID string) (*ReDelegateMetadata, error) {
	metadataBase := NewMetadataBaseWithSignature(ReDelegateMeta)
	return &ReDelegateMetadata{
		MetadataBaseWithSignature: *metadataBase,
		CommitteePublicKey:        committeePublicKey,
		NewDelegate:               newDelegate,
		DelegateUID:               newDelegateUID,
	}, nil
}

/*
 */
func (redelegateMetadata *ReDelegateMetadata) ValidateMetadataByItself() bool {
	CommitteePublicKey := new(incognitokey.CommitteePublicKey)
	if err := CommitteePublicKey.FromString(redelegateMetadata.CommitteePublicKey); err != nil {
		return false
	}
	if !CommitteePublicKey.CheckSanityData() {
		return false
	}
	if err := CommitteePublicKey.FromString(redelegateMetadata.NewDelegate); err != nil {
		return false
	}
	if !CommitteePublicKey.CheckSanityData() {
		return false
	}
	return (redelegateMetadata.Type == ReDelegateMeta)
}

func (redelegateMetadata ReDelegateMetadata) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	triggerMap := shardViewRetriever.GetTriggeredFeature()
	if (triggerMap != nil) && (triggerMap[config.DELEGATION] <= shardViewRetriever.GetHeight()) {
		newDelegate := redelegateMetadata.NewDelegate
		beaconStakerInfo, has, err := beaconViewRetriever.GetBeaconStakerInfo(newDelegate)
		if (!has) || (err != nil) {
			return false, NewMetadataTxError(ReDelegateCommitteeNotFoundError, fmt.Errorf("Committee Publickey %+v not found in any committee list of current beacon beststate", newDelegate))
		}
		rawUID := fmt.Sprintf("%v-%v", newDelegate, beaconStakerInfo.BeaconConfirmHeight())
		uID := common.HashH([]byte(rawUID))
		if uID.String() != redelegateMetadata.DelegateUID {
			return false, NewMetadataTxError(ReDelegateCommitteeNotFoundError, fmt.Errorf("Committee Publickey %+v with Beacon confirm height %v not match with the UID in Metadata, expected %v, got %v", newDelegate, beaconStakerInfo.BeaconConfirmHeight(), redelegateMetadata.DelegateUID, uID.String()))
		}
		requestedPublicKey := redelegateMetadata.CommitteePublicKey
		processList := []string{}
		shardCommittee := shardViewRetriever.GetShardCommittee()
		sCStr, err := incognitokey.CommitteeKeyListToString(shardCommittee)
		if err != nil {
			return false, errors.Errorf("Can not convert list shard committee %+v to string, shard height %+v", shardCommittee, shardViewRetriever.GetHeight())
		}
		shardPending := shardViewRetriever.GetShardPendingValidator()
		sPStr, err := incognitokey.CommitteeKeyListToString(shardPending)
		if err != nil {
			return false, errors.Errorf("Can not convert list shard pending %+v to string, shard ID %v, shard height %+v", shardPending, shardViewRetriever.GetShardID(), shardViewRetriever.GetHeight())
		}
		processList = append(sCStr, sPStr...)
		if common.IndexOfStr(requestedPublicKey, processList) == -1 {
			return false, errors.Errorf("Can not found public key %v in list shard committeee %+v and list shard pending %+v, shard ID %v, shard height %v", requestedPublicKey, sCStr, sPStr, shardViewRetriever.GetShardID(), shardViewRetriever.GetHeight())
		}

		stakerInfo, has, err := beaconViewRetriever.GetStakerInfo(requestedPublicKey)
		if err != nil {
			return false, NewMetadataTxError(ReDelegateRequestNotInCommitteeListError, err)
		}
		if !has {
			return false, NewMetadataTxError(ReDelegateRequestNotInCommitteeListError, fmt.Errorf("No Staker Info of %+v found in Beacon statedb of Shard %+v", requestedPublicKey, shardID))
		}

		rewardReceiverAddress := stakerInfo.RewardReceiver()
		pkToValidateMetadata := rewardReceiverAddress.Pk
		stakingTxHash := stakerInfo.TxStakingID()

		_, _, _, _, stakingTx, err := chainRetriever.GetTransactionByHash(stakingTxHash)
		if err == nil {
			stakingMetadata := stakingTx.GetMetadata().(*StakingMetadata)
			funderPaymentAddress := stakingMetadata.FunderPaymentAddress

			funderWallet, err := wallet.Base58CheckDeserialize(funderPaymentAddress)
			if err != nil || funderWallet == nil {
				return false, errors.New("Invalid Funder Payment Address, Failed to Deserialized Into Key Wallet")
			}
			pkToValidateMetadata = funderWallet.KeySet.PaymentAddress.Pk
		}

		if ok, err := redelegateMetadata.MetadataBaseWithSignature.VerifyMetadataSignature(pkToValidateMetadata, tx); !ok || err != nil {
			return false, NewMetadataTxError(ReDelegateInvalidTxError, fmt.Errorf("CheckAuthorizedSender fail"))
		}

		if (stakerInfo.GetDelegate() == newDelegate) && (stakerInfo.GetDelegateUID() == redelegateMetadata.DelegateUID) {
			return false, NewMetadataTxError(ReDelegateInvalidTxError, fmt.Errorf("Cannot replace with the same key"))
		}
		return true, nil
	}
	return false, errors.Errorf("The redelegate feature is not ready, trigger feature info %+v, shard ID %+v height %+v", triggerMap, shardID, shardViewRetriever.GetHeight())
}

// Have only one receiver
// Have only one amount corresponding to receiver
// Receiver Is Burning Address
func (redelegateMetadata ReDelegateMetadata) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	isBurned, burnCoin, tokenID, err := tx.GetTxBurnData()
	if err != nil {
		return false, false, errors.New("Error Cannot get burn data from tx")
	}
	if !isBurned {
		return false, false, errors.New("Error ReDelegate tx should be a burn tx")
	}
	if !bytes.Equal(tokenID[:], common.PRVCoinID[:]) {
		return false, false, errors.New("Error ReDelegate tx should transfer PRV only")
	}
	if redelegateMetadata.Type != ReDelegateMeta && burnCoin.GetValue() != ReDelegateFee {
		return false, false, errors.New("receiver amount should be zero")
	}
	CommitteePublicKey := new(incognitokey.CommitteePublicKey)
	if err := CommitteePublicKey.FromString(redelegateMetadata.CommitteePublicKey); err != nil {
		return false, false, err
	}
	if !CommitteePublicKey.CheckSanityData() {
		return false, false, errors.New("Invalid Commitee Public Key of Candidate who join consensus")
	}
	return true, true, nil
}

func (redelegateMetadata ReDelegateMetadata) GetType() int {
	return redelegateMetadata.Type
}

func (redelegateMetadata *ReDelegateMetadata) CalculateSize() uint64 {
	return calculateSize(redelegateMetadata)
}
