package metadata

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
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
	return meta.MetadataBase.Hash()
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
	if (addStakingMetadata.AddStakingAmount%1750 != 0) || (addStakingMetadata.AddStakingAmount < 1750) {
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
		return false, NewMetadataTxError(StopAutoStakingRequestTypeAssertionError, fmt.Errorf("Expect *AddStakingMetadata type but get %+v", reflect.TypeOf(tx.GetMetadata())))
	}
	requestedPublicKey := addStakingMetadataFromTx.CommitteePublicKey
	committees := beaconViewRetriever.GetAllBeaconValidatorCandidateFlattenList()

	// if not found
	if !(common.IndexOfStr(requestedPublicKey, committees) > -1) {
		return false, NewMetadataTxError(StopAutoStakingRequestNotInCommitteeListError, fmt.Errorf("Committee Publickey %+v not found in any committee list of current beacon beststate", requestedPublicKey))
	}
	stakerInfo, has, err := statedb.GetBeaconStakerInfo(beaconViewRetriever.GetBeaconConsensusStateDB(), requestedPublicKey)
	if err != nil {
		return false, NewMetadataTxError(StopAutoStakingRequestNotInCommitteeListError, err)
	}
	if !has {
		return false, NewMetadataTxError(StopAutoStakingRequestStakingTransactionNotFoundError, fmt.Errorf("No Committe Publickey %+v found in StakingTx of Shard %+v", requestedPublicKey, shardID))
	}
	stakingTxHash := stakerInfo.StakingInfo()
	var oneStakingTxHash common.Hash
	for k, _ := range stakingTxHash {
		oneStakingTxHash = k
	}
	_, _, _, _, stakingTx, err := chainRetriever.GetTransactionByHash(oneStakingTxHash)
	if err != nil {
		return false, NewMetadataTxError(StopAutoStakingRequestStakingTransactionNotFoundError, err)
	}

	stakingMetadata := stakingTx.GetMetadata().(*StakingMetadata)
	funderPaymentAddress := stakingMetadata.FunderPaymentAddress
	funderWallet, err := wallet.Base58CheckDeserialize(funderPaymentAddress)
	if err != nil || funderWallet == nil {
		return false, errors.New("Invalid Funder Payment Address, Failed to Deserialized Into Key Wallet")
	}

	if ok, err := addStakingMetadataFromTx.MetadataBaseWithSignature.VerifyMetadataSignature(funderWallet.KeySet.PaymentAddress.Pk, tx); !ok || err != nil {
		return false, NewMetadataTxError(StopAutoStakingRequestInvalidTransactionSenderError, fmt.Errorf("CheckAuthorizedSender fail"))
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
		return false, false, errors.New("Error StopAutoStaking tx should be a burn tx")
	}
	if !bytes.Equal(tokenID[:], common.PRVCoinID[:]) {
		return false, false, errors.New("Error StopAutoStaking tx should transfer PRV only")
	}
	if addStakingMetadata.Type != AddStakingMeta && (burnCoin.GetValue()%1750 != 0) && (burnCoin.GetValue() < 1750*3) {
		return false, false, errors.New("receiver amount invalid: " + fmt.Sprintln(burnCoin.GetValue()))
	}
	CommitteePublicKey := new(incognitokey.CommitteePublicKey)
	if err := CommitteePublicKey.FromString(addStakingMetadata.CommitteePublicKey); err != nil {
		return false, false, err
	}
	if !CommitteePublicKey.CheckSanityData() {
		return false, false, errors.New("Invalid Commitee Public Key of Candidate who join consensus")
	}
	return true, true, nil
}

func (addStakingMetadata AddStakingMetadata) GetType() int {
	return addStakingMetadata.Type
}

func (addStakingMetadata *AddStakingMetadata) CalculateSize() uint64 {
	return calculateSize(addStakingMetadata)
}
