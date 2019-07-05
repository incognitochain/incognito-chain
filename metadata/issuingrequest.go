package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	privacy "github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

// only centralized website can send this type of tx
type IssuingRequest struct {
	ReceiverAddress privacy.PaymentAddress
	DepositedAmount uint64
	TokenID         common.Hash
	TokenName       string
	MetadataBase
}

func NewIssuingRequest(
	receiverAddress privacy.PaymentAddress,
	depositedAmount uint64,
	tokenID common.Hash,
	tokenName string,
	metaType int,
) (*IssuingRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	issuingReq := &IssuingRequest{
		ReceiverAddress: receiverAddress,
		DepositedAmount: depositedAmount,
		TokenID:         tokenID,
		TokenName:       tokenName,
	}
	issuingReq.MetadataBase = metadataBase
	return issuingReq, nil
}

func NewIssuingRequestFromMap(data map[string]interface{}) (Metadata, error) {
	tokenID, err := common.Hash{}.NewHashFromStr(data["TokenID"].(string))
	if err != nil {
		return nil, errors.Errorf("TokenID incorrect")
	}

	tokenName, ok := data["TokenName"].(string)
	if !ok {
		return nil, errors.Errorf("TokenName incorrect")
	}

	depositedAmt := uint64(data["DepositedAmount"].(float64))
	keyWallet, err := wallet.Base58CheckDeserialize(data["ReceiveAddress"].(string))
	if err != nil {
		return nil, errors.Errorf("ReceiveAddress incorrect")
	}

	return NewIssuingRequest(
		keyWallet.KeySet.PaymentAddress,
		depositedAmt,
		*tokenID,
		tokenName,
		IssuingRequestMeta,
	)
}

func (iReq *IssuingRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	if !bytes.Equal(txr.GetSigPubKey(), common.CentralizedWebsitePubKey) {
		return false, errors.New("the issuance request must be called by centralized website")
	}

	bridgeTokenExisted, err := db.IsBridgeTokenExisted(iReq.TokenID)
	if err != nil {
		return false, err
	}

	privacyCustomTokenExisted := db.PrivacyCustomTokenIDExisted(iReq.TokenID)
	privacyCustomTokenCrossShardExisted := db.PrivacyCustomTokenIDCrossShardExisted(iReq.TokenID)
	if !bridgeTokenExisted && (privacyCustomTokenExisted || privacyCustomTokenCrossShardExisted) {
		return false, errors.New("another custom token was already existed with the same token id")
	}
	return true, nil
}

func (iReq *IssuingRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(iReq.ReceiverAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's receiver address")
	}
	if iReq.DepositedAmount == 0 {
		return false, false, errors.New("Wrong request info's deposited amount")
	}
	if iReq.Type != IssuingRequestMeta {
		return false, false, errors.New("Wrong request info's meta type")
	}
	if len(iReq.TokenID) != common.HashSize {
		return false, false, errors.New("Wrong request info's token id")
	}
	if iReq.TokenName == "" {
		return false, false, errors.New("Wrong request info's token name")
	}
	return true, true, nil
}

func (iReq *IssuingRequest) ValidateMetadataByItself() bool {
	if iReq.Type != IssuingRequestMeta {
		return false
	}
	return true
}

func (iReq *IssuingRequest) Hash() *common.Hash {
	record := iReq.ReceiverAddress.String()
	record += iReq.TokenID.String()
	record += string(iReq.DepositedAmount)
	record += iReq.TokenName
	record += iReq.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iReq *IssuingRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := map[string]interface{}{
		"meta": *iReq,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(IssuingRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (iReq *IssuingRequest) CalculateSize() uint64 {
	return calculateSize(iReq)
}
