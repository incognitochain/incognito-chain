package metadata

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/config"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type WithdrawDelegationRewardRequest struct {
	MetadataBaseWithSignature
	PaymentAddress privacy.PaymentAddress
	Version        int
}

func (withdrawRequestMetadata *WithdrawDelegationRewardRequest) UnmarshalJSON(data []byte) error {
	tmp := &struct {
		MetadataBaseWithSignature
		PaymentAddress privacy.PaymentAddress
		TokenID        common.Hash
		Version        int
	}{}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if tmp.PaymentAddress.Pk == nil && tmp.PaymentAddress.Tk == nil {
		tmpOld := &struct {
			MetadataBaseWithSignature
			privacy.PaymentAddress
			TokenID common.Hash
			Version int
		}{}
		if err := json.Unmarshal(data, &tmpOld); err != nil {
			return err
		}

		tmp.PaymentAddress.Tk = tmpOld.Tk
		tmp.PaymentAddress.Pk = tmpOld.Pk
	}

	withdrawRequestMetadata.MetadataBaseWithSignature = tmp.MetadataBaseWithSignature
	withdrawRequestMetadata.PaymentAddress = tmp.PaymentAddress
	withdrawRequestMetadata.Version = tmp.Version
	return nil
}

func (withdrawDelegationRewardRequest WithdrawDelegationRewardRequest) Hash() *common.Hash {
	bArr := append([]byte{}, withdrawDelegationRewardRequest.PaymentAddress.Bytes()...)
	if withdrawDelegationRewardRequest.Sig != nil && len(withdrawDelegationRewardRequest.Sig) != 0 {
		bArr = append(bArr, withdrawDelegationRewardRequest.Sig...)
	}
	txReqHash := common.HashH(bArr)
	return &txReqHash
}

func (withdrawDelegationRewardRequest WithdrawDelegationRewardRequest) HashWithoutSig() *common.Hash {
	bArr := withdrawDelegationRewardRequest.PaymentAddress.Bytes()
	txReqHash := common.HashH(bArr)
	return &txReqHash
}

func NewWithdrawDelegationRewardRequest(paymentAddStr string, version float64, metaType int) (*WithdrawDelegationRewardRequest, error) {
	metadataBase := NewMetadataBaseWithSignature(metaType)

	paymentAddWallet, err := wallet.Base58CheckDeserialize(paymentAddStr)
	if err != nil {
		return nil, errors.New("payment address is invalid")
	}
	ok, err := common.SliceExists(AcceptedWithdrawRewardRequestVersion, int(version))
	if !ok || err != nil {
		return nil, errors.Errorf("Invalid version %v", version)
	}

	withdrawRewardRequest := &WithdrawDelegationRewardRequest{
		MetadataBaseWithSignature: *metadataBase,
		PaymentAddress:            paymentAddWallet.KeySet.PaymentAddress,
		Version:                   int(version),
	}
	return withdrawRewardRequest, nil
}

func NewWithdrawDelegationRewardRequestFromRPC(data map[string]interface{}) (Metadata, error) {
	metadataBase := NewMetadataBaseWithSignature(WithdrawDelegationRewardRequestMeta)
	requesterPaymentStr, ok := data["PaymentAddress"].(string)
	if !ok {
		return nil, errors.New("Invalid payment address receiver")
	}

	requesterPublicKeySet, err := wallet.Base58CheckDeserialize(requesterPaymentStr)
	if err != nil {
		return nil, err
	}
	result := &WithdrawDelegationRewardRequest{
		MetadataBaseWithSignature: *metadataBase,
		PaymentAddress:            requesterPublicKeySet.KeySet.PaymentAddress,
		Version:                   common.SALARY_VER_FIX_HASH,
	}

	// versionFloat, ok := data["Version"].(float64)
	// if ok {
	// 	version := int(versionFloat)
	// 	result.Version = version
	// }
	if ok, err := common.SliceExists(AcceptedWithdrawRewardRequestVersion, result.Version); !ok || err != nil {
		return nil, errors.Errorf("Invalid version %d", result.Version)
	}
	return result, nil
}

func (withdrawDelegationRewardRequest WithdrawDelegationRewardRequest) CheckTransactionFee(tr Transaction, minFeePerKb uint64, minFeePerTx uint64, beaconHeight int64, stateDB *statedb.StateDB) bool {
	return true
}

func (withdrawDelegationRewardRequest WithdrawDelegationRewardRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	triggerMap := shardViewRetriever.GetTriggeredFeature()

	if triggerMap == nil || shardViewRetriever.GetTriggeredFeature()[config.DELEGATION_REWARD] == 0 {
		return false, fmt.Errorf("this feature is not supported yet!")
	}

	if tx.IsPrivacy() && tx.GetVersion() <= 1 {
		return false, fmt.Errorf("reward-withdraw request transaction version 1 should not be private")
	}

	//check authorized sender
	if ok, err := withdrawDelegationRewardRequest.MetadataBaseWithSignature.VerifyMetadataSignature(withdrawDelegationRewardRequest.PaymentAddress.Pk, tx); err != nil || !ok {
		return false, fmt.Errorf("public key in withdraw delegation reward request metadata is unauthorized. Error %v, OK %v", err, ok)
	}

	//check available reward
	if rewardAmount, err := chainRetriever.GetDelegationRewardAmount(beaconViewRetriever.GetBeaconConsensusStateDB(), withdrawDelegationRewardRequest.PaymentAddress.Pk); err != nil {
		return false, err
	} else if rewardAmount <= 0 {
		return false, errors.New("Not enough reward")
	} else {
		return true, nil
	}

	return false, errors.Errorf("The delegation reward feature is not ready, trigger feature info %+v, shard ID %+v height %+v", triggerMap, shardID, shardViewRetriever.GetHeight())
}

func (withdrawDelegationRewardRequest WithdrawDelegationRewardRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	if _, err := AssertPaymentAddressAndTxVersion(withdrawDelegationRewardRequest.PaymentAddress, tx.GetVersion()); err != nil {
		return false, false, err
	}
	return false, true, nil
}

func (withdrawDelegationRewardRequest WithdrawDelegationRewardRequest) ValidateMetadataByItself() bool {
	if ok, err := common.SliceExists(AcceptedWithdrawRewardRequestVersion, withdrawDelegationRewardRequest.Version); !ok || err != nil {
		Logger.log.Error(errors.Errorf("Invalid version %d", withdrawDelegationRewardRequest.Version))
		return false
	}
	return true
}

func (withdrawDelegationRewardRequest WithdrawDelegationRewardRequest) GetType() int {
	return withdrawDelegationRewardRequest.Type
}

func (withdrawDelegationRewardRequest *WithdrawDelegationRewardRequest) CalculateSize() uint64 {
	return calculateSize(withdrawDelegationRewardRequest)
}
