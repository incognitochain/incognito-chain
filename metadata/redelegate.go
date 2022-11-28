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

type ReDelegateMetadata struct {
	MetadataBaseWithSignature
	CommitteePublicKey string
	NewDelegate        string
	// CommitteePublicKey CommitteePublicKeys of a candidate who join consensus, base58CheckEncode
	// CommitteePublicKey string <= encode byte <= mashal struct
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

func NewReDelegateMetadata(committeePublicKey, newDelegate string) (*ReDelegateMetadata, error) {
	// if stopStakingType != StopAutoStakingMeta {
	// 	return nil, errors.New("invalid stop staking type")
	// }
	metadataBase := NewMetadataBaseWithSignature(ReDelegateMeta)
	return &ReDelegateMetadata{
		MetadataBaseWithSignature: *metadataBase,
		CommitteePublicKey:        committeePublicKey,
		NewDelegate:               newDelegate,
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

// ValidateTxWithBlockChain Validate Condition to Request Stop AutoStaking With Blockchain
// - Requested Committee Publickey is in candidate, pending validator,
// - Requested Committee Publickey is in staking tx list,
// - Requester (sender of tx) must be address, which create staking transaction for current requested committee public key
// - Not yet requested to stop auto-restaking
func (redelegateMetadata ReDelegateMetadata) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	newDelegate := redelegateMetadata.NewDelegate
	delegateList := beaconViewRetriever.GetAllBeaconValidatorCandidateFlattenList()
	requestedPublicKey := redelegateMetadata.CommitteePublicKey
	if !(common.IndexOfStr(newDelegate, delegateList) > -1) {
		return false, NewMetadataTxError(StopAutoStakingRequestNotInCommitteeListError, fmt.Errorf("Committee Publickey %+v not found in any committee list of current beacon beststate", newDelegate))
	}

	stakerInfo, has, err := beaconViewRetriever.GetStakerInfo(requestedPublicKey)
	if err != nil {
		return false, NewMetadataTxError(StopAutoStakingRequestNotInCommitteeListError, err)
	}
	if !has {
		return false, NewMetadataTxError(StopAutoStakingRequestStakingTransactionNotFoundError, fmt.Errorf("No Committe Publickey %+v found in StakingTx of Shard %+v", newDelegate, shardID))
	}
	stakingTxHash := stakerInfo.TxStakingID()

	_, _, _, _, stakingTx, err := chainRetriever.GetTransactionByHash(stakingTxHash)
	if err != nil {
		return false, NewMetadataTxError(StopAutoStakingRequestStakingTransactionNotFoundError, err)
	}

	stakingMetadata := stakingTx.GetMetadata().(*StakingMetadata)
	funderPaymentAddress := stakingMetadata.FunderPaymentAddress
	funderWallet, err := wallet.Base58CheckDeserialize(funderPaymentAddress)
	if err != nil || funderWallet == nil {
		return false, errors.New("Invalid Funder Payment Address, Failed to Deserialized Into Key Wallet")
	}

	if ok, err := redelegateMetadata.MetadataBaseWithSignature.VerifyMetadataSignature(funderWallet.KeySet.PaymentAddress.Pk, tx); !ok || err != nil {
		return false, NewMetadataTxError(StopAutoStakingRequestInvalidTransactionSenderError, fmt.Errorf("CheckAuthorizedSender fail"))
	}

	if stakerInfo.Delegate() == newDelegate {
		return false, NewMetadataTxError(StopAutoStakingRequestNoAutoStakingAvaiableError, fmt.Errorf("Cannot replace with the same key"))
	}
	return true, nil
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
		return false, false, errors.New("Error StopAutoStaking tx should be a burn tx")
	}
	if !bytes.Equal(tokenID[:], common.PRVCoinID[:]) {
		return false, false, errors.New("Error StopAutoStaking tx should transfer PRV only")
	}
	if redelegateMetadata.Type != ReDelegateMeta && burnCoin.GetValue() != StopAutoStakingAmount {
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
