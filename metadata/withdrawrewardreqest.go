package metadata

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type WithDrawRewardRequest struct {
	MetadataBaseWithSignature
	PaymentAddress privacy.PaymentAddress
	TokenID        common.Hash
	Version        int
}

func (withdrawRequestMetadata *WithDrawRewardRequest) UnmarshalJSON(data []byte) error {
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
	withdrawRequestMetadata.TokenID = tmp.TokenID
	withdrawRequestMetadata.Version = tmp.Version
	return nil
}

func (withDrawRewardRequest WithDrawRewardRequest) Hash() *common.Hash {
	if withDrawRewardRequest.Version == 1 {
		bArr := append(withDrawRewardRequest.PaymentAddress.Bytes(), withDrawRewardRequest.TokenID.GetBytes()...)
		if withDrawRewardRequest.Sig != nil && len(withDrawRewardRequest.Sig) != 0 {
			bArr = append(bArr, withDrawRewardRequest.Sig...)
		}
		txReqHash := common.HashH(bArr)
		return &txReqHash
	} else {
		record := strconv.Itoa(withDrawRewardRequest.Type)
		data := []byte(record)
		hash := common.HashH(data)
		return &hash
	}
}

func (withDrawRewardRequest WithDrawRewardRequest) HashWithoutSig() *common.Hash {
	if withDrawRewardRequest.Version == common.SALARY_VER_FIX_HASH {
		bArr := append(withDrawRewardRequest.PaymentAddress.Bytes(), withDrawRewardRequest.TokenID.GetBytes()...)
		txReqHash := common.HashH(bArr)
		return &txReqHash
	} else {
		record := strconv.Itoa(withDrawRewardRequest.Type)
		data := []byte(record)
		hash := common.HashH(data)
		return &hash
	}
}

func NewWithDrawRewardRequest(tokenIDStr string, paymentAddStr string, version float64, metaType int) (*WithDrawRewardRequest, error) {
	metadataBase := NewMetadataBaseWithSignature(metaType)
	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDStr)
	if err != nil {
		return nil, errors.New("token ID is invalid")
	}
	paymentAddWallet, err := wallet.Base58CheckDeserialize(paymentAddStr)
	if err != nil {
		return nil, errors.New("payment address is invalid")
	}
	ok, err := common.SliceExists(AcceptedWithdrawRewardRequestVersion, int(version))
	if !ok || err != nil {
		return nil, errors.Errorf("Invalid version %v", version)
	}

	withdrawRewardRequest := &WithDrawRewardRequest{
		MetadataBaseWithSignature: *metadataBase,
		TokenID:                   *tokenID,
		PaymentAddress:            paymentAddWallet.KeySet.PaymentAddress,
		Version:                   int(version),
	}
	return withdrawRewardRequest, nil
}

func NewWithDrawRewardRequestFromRPC(data map[string]interface{}) (Metadata, error) {
	metadataBase := NewMetadataBaseWithSignature(WithDrawRewardRequestMeta)
	requesterPaymentStr, ok := data["PaymentAddress"].(string)
	if !ok {
		return nil, errors.New("Invalid payment address receiver")
	}
	requestTokenID, ok := data["TokenID"].(string)
	if !ok {
		return nil, errors.New("Invalid token Id")
	}
	tokenID, err := common.Hash{}.NewHashFromStr(requestTokenID)
	if err != nil {
		return nil, err
	}
	requesterPublicKeySet, err := wallet.Base58CheckDeserialize(requesterPaymentStr)
	if err != nil {
		return nil, err
	}
	result := &WithDrawRewardRequest{
		MetadataBaseWithSignature: *metadataBase,
		PaymentAddress:            requesterPublicKeySet.KeySet.PaymentAddress,
		TokenID:                   *tokenID,
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

func (withDrawRewardRequest WithDrawRewardRequest) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, stateDB *statedb.StateDB) bool {
	return true
}

func (withDrawRewardRequest WithDrawRewardRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	if tx.IsPrivacy() && tx.GetVersion() <= 1 {
		return false, fmt.Errorf("reward-withdraw request transaction version 1 should not be private")
	}

	//check authorized sender
	if ok, err := withDrawRewardRequest.MetadataBaseWithSignature.VerifyMetadataSignature(withDrawRewardRequest.PaymentAddress.Pk, tx); err != nil || !ok {
		return false, fmt.Errorf("public key in withdraw reward request metadata is unauthorized. Error %v, OK %v", err, ok)
	}

	//check token valid (!= PRV)
	tokenIDReq := withDrawRewardRequest.TokenID
	isTokenValid := false
	var err error
	if !tokenIDReq.IsEqual(&common.PRVCoinID) {
		bridgeDB := beaconViewRetriever.GetBeaconFeatureStateDB()
		isTokenValid, err = statedb.IsBridgeToken(bridgeDB, tokenIDReq)
		if err != nil {
			return false, err
		} else {
			if !isTokenValid {
				isTokenValid = statedb.PrivacyTokenIDExisted(transactionStateDB, tokenIDReq)
			}
		}
		if !isTokenValid {
			return false, errors.New("Invalid TokenID, maybe this coin not available at current shard")
		}
	}
	pubKeyReqStr := base58.Base58Check{}.Encode(withDrawRewardRequest.PaymentAddress.Pk, common.Base58Version)
	//check available reward
	if rewardAmount, err := statedb.GetCommitteeReward(shardViewRetriever.GetShardRewardStateDB(), pubKeyReqStr, tokenIDReq); err != nil {
		return false, err
	} else if rewardAmount <= 0 {
		return false, errors.New("Not enough reward ")
	} else {
		return true, nil
	}
}

func (withDrawRewardRequest WithDrawRewardRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	if tx.GetVersion() == 2 && withDrawRewardRequest.TokenID.String() != common.PRVIDStr {
		return false, false, fmt.Errorf("withdrawing token is no longer supported in a transaction v2")
	}
	if _, err := AssertPaymentAddressAndTxVersion(withDrawRewardRequest.PaymentAddress, tx.GetVersion()); err != nil {
		return false, false, err
	}
	return false, true, nil
}

func (withDrawRewardRequest WithDrawRewardRequest) ValidateMetadataByItself() bool {
	if ok, err := common.SliceExists(AcceptedWithdrawRewardRequestVersion, withDrawRewardRequest.Version); !ok || err != nil {
		Logger.log.Error(errors.Errorf("Invalid version %d", withDrawRewardRequest.Version))
		return false
	}
	return true
}

func (withDrawRewardRequest WithDrawRewardRequest) GetType() int {
	return withDrawRewardRequest.Type
}

func (withDrawRewardRequest *WithDrawRewardRequest) CalculateSize() uint64 {
	return calculateSize(withDrawRewardRequest)
}

func (withDrawRewardRequest *WithDrawRewardRequest) ToCompactBytes() ([]byte, error) {
	return toCompactBytes(withDrawRewardRequest)
}
