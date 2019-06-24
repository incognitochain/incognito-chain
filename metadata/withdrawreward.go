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
	TokenID common.Hash
}

func NewWithDrawRewardRequestFromRPC(data map[string]interface{}) (Metadata, error) {
	metadataBase := MetadataBase{
		Type: WithDrawRewardRequestMeta,
	}
	requesterPaymentStr := data["PaymentAddress"].(string)
	requestTokenID := data["TokenID"].(string)
	tokenID, err := common.Hash{}.NewHashFromStr(requestTokenID)
	if err != nil {
		return nil, err
	}
	fmt.Printf("[ndh] - - request %+v PRV %+v\n", *tokenID, common.PRVCoinID)
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
		TokenID:        *tokenID,
	}, nil
}

type WithDrawRewardResponse struct {
	MetadataBase
	TxRequest *common.Hash
	TokenID   common.Hash
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
		return false, errors.New("This transaction is not private")
	}
	isPositive := false
	value, err := db.GetCommitteeReward(withDrawRewardRequest.PaymentAddress.Pk, withDrawRewardRequest.TokenID)
	if err != nil {
		return false, err
	}
	if value > 0 {
		isPositive = true
	}
	// }
	if !isPositive {
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
		return false, errors.New("This transaction is not private")
	}
	unique, requesterRes, amountRes, coinID := txr.GetTransferData()
	if !unique {
		return false, errors.New("Just one receiver")
	}
	value, err := db.GetCommitteeReward(requesterRes, *coinID)
	if (err != nil) || (value == 0) {
		return false, errors.New("Not enough reward")
	}
	if value != amountRes {
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
