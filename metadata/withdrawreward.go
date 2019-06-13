package metadata

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type WithDrawRewardRequest struct {
	privacy.PaymentAddress
	MetadataBase
}

func NewWithDrawRewardRequestFromRPC(data map[string]interface{}) (Metadata, error) {
	metadataBase := MetadataBase{
		Type: WithDrawRewardRequestMeta,
	}
	requesterPaymentStr := data["PaymentAddress"].(string)
	for key, value := range data {
		fmt.Printf("[ndh]- - - - Key %+v; value %+v\n", key, value)
	}
	requesterPublicKeySet, err := wallet.Base58CheckDeserialize(requesterPaymentStr)
	if err != nil {
		return nil, err
	}
	return &WithDrawRewardRequest{
		MetadataBase:   metadataBase,
		PaymentAddress: requesterPublicKeySet.KeySet.PaymentAddress,
	}, nil
}

type WithDrawRewardResponse struct {
	MetadataBase
	TxRequest *common.Hash
}

func NewWithDrawRewardResponse(txRequestID *common.Hash) (Metadata, error) {
	metadataBase := MetadataBase{
		Type: WithDrawRewardResponseMeta,
	}
	return &WithDrawRewardResponse{
		MetadataBase: metadataBase,
		TxRequest:    txRequestID,
	}, nil
}

func (withDrawRewardResponse *WithDrawRewardResponse) Hash() *common.Hash {
	return withDrawRewardResponse.TxRequest
}

func (withDrawRewardRequest *WithDrawRewardRequest) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	//TODO - check fee?
	return true
}

func (withDrawRewardRequest *WithDrawRewardRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	if txr.IsPrivacy() {
		return false, errors.New("This transaction can not be privacy")
	}
	value, err := db.GetCommitteeReward(withDrawRewardRequest.PaymentAddress.Pk)
	if (err != nil) || (value == 0) {
		return false, errors.New("Not enough reward")
	}
	receivers, _ := txr.GetReceivers()
	if len(receivers) > 0 {
		return false, errors.New("This metadata just for request withdraw reward")
	}
	return true, nil
}

func (withDrawRewardRequest *WithDrawRewardRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (withDrawRewardRequest *WithDrawRewardRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (withDrawRewardResponse *WithDrawRewardResponse) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	//TODO - check fee?
	return true
}

func (withDrawRewardResponse *WithDrawRewardResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	if txr.IsPrivacy() {
		return false, errors.New("This transaction can not be privacy")
	}
	receivers, amounts := txr.GetReceivers()
	if len(receivers) != 1 {
		return false, errors.New("Wrong receiver")
	}
	value, err := db.GetCommitteeReward(receivers[0])
	if (err != nil) || (value == 0) {
		return false, errors.New("Not enough reward")
	}
	if value != amounts[0] {
		return false, errors.New("Wrong amounts")
	}
	return true, nil
}

func (withDrawRewardResponse *WithDrawRewardResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (withDrawRewardResponse *WithDrawRewardResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}
