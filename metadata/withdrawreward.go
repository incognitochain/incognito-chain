package metadata

import (
	"fmt"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
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
	requesterPublicKeySet, err := wallet.Base58CheckDeserialize(requesterPaymentStr)
	if err != nil {
		panic(err.Error())
		return nil, err
	}
	fmt.Printf("\n\n\n\n\n\n\n\n %+v----------------- \n\n\n\n\n", requesterPaymentStr)
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
	// if txr.IsPrivacy() {
	// 	return false, errors.New("This transaction can not be privacy")
	// }
	// sender := txr.GetSender()
	// value, err := db.GetCommitteeReward(sender[:33])
	// if (err != nil) || (value == 0) {
	// 	return false, errors.New("Not enough reward")
	// }
	// receivers, _ := txr.GetReceivers()
	// if len(receivers) > 0 {
	// 	return false, errors.New("This metadata just for request withdraw reward")
	// }
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
