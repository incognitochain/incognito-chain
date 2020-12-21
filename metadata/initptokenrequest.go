package metadata

import (
	"fmt"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

type InitPTokenRequest struct {
	ReceiverAddress privacy.PaymentAddress
	Amount 					uint64
	TokenID         common.Hash
	TokenName       string
	TokenSymbol 		string
	MetadataBase
}

type InitPTokenReqAction struct {
	Meta    InitPTokenRequest `json:"meta"`
	TxReqID common.Hash       `json:"txReqId"`
}

type InitPTokenAcceptedInst struct {
	ShardID         byte                   `json:"shardId"`
	Amount 					uint64                 `json:"amount"`
	ReceiverAddr    privacy.PaymentAddress `json:"receiverAddrStr"`
	IncTokenID      common.Hash            `json:"incTokenId"`
	IncTokenName    string                 `json:"incTokenName"`
	IncTokenSymbol  string                 `json:"incTokenSymbol"`
	TxReqID         common.Hash            `json:"txReqId"`
}

func ParseInitPTokenInstContent(instContentStr string) (*InitPTokenReqAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instContentStr)
	if err != nil {
		return nil, NewMetadataTxError(InitPTokenRequestDecodeInstructionError, err)
	}
	var initPTokenReqAction InitPTokenReqAction
	err = json.Unmarshal(contentBytes, &initPTokenReqAction)
	if err != nil {
		return nil, NewMetadataTxError(InitPTokenRequestUnmarshalJsonError, err)
	}
	return &initPTokenReqAction, nil
}

func NewInitPTokenRequest(
	receiverAddress privacy.PaymentAddress,
	amount uint64,
	tokenID common.Hash,
	tokenName string,
	tokenSymbol string,
	metaType int,
) (*InitPTokenRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	initPTokenReq := &InitPTokenRequest{
		ReceiverAddress: receiverAddress,
		Amount: 				 amount,
		TokenID:         tokenID,
		TokenName:       tokenName,
		TokenSymbol: 		 tokenSymbol,
	}
	initPTokenReq.MetadataBase = metadataBase
	return initPTokenReq, nil
}

func NewInitPTokenRequestFromMap(data map[string]interface{}) (Metadata, error) {
	tokenID, err := common.Hash{}.NewHashFromStr(data["TokenID"].(string))
	if err != nil {
		return nil, NewMetadataTxError(InitPTokenRequestNewInitPTokenRequestFromMapError, errors.New("TokenID incorrect"))
	}

	tokenName, ok := data["TokenName"].(string)
	if !ok {
		return nil, NewMetadataTxError(InitPTokenRequestNewInitPTokenRequestFromMapError, errors.New("TokenName incorrect"))
	}

	tokenSymbol, ok := data["TokenSymbol"].(string)
	if !ok {
		return nil, NewMetadataTxError(InitPTokenRequestNewInitPTokenRequestFromMapError, errors.New("TokenSymbol incorrect"))
	}

	amt, err := common.AssertAndConvertStrToNumber(data["Amount"])
	if err != nil {
		return nil, NewMetadataTxError(InitPTokenRequestNewInitPTokenRequestFromMapError, errors.New("Amount incorrect"))
	}

	keyWallet, err := wallet.Base58CheckDeserialize(data["ReceiveAddress"].(string))
	if err != nil {
		return nil, NewMetadataTxError(InitPTokenRequestNewInitPTokenRequestFromMapError, errors.New("ReceiveAddress incorrect"))
	}

	return NewInitPTokenRequest(
		keyWallet.KeySet.PaymentAddress,
		amt,
		*tokenID,
		tokenName,
		tokenSymbol,
		InitPTokenRequestMeta,
	)
}

func (iReq InitPTokenRequest) ValidateTxWithBlockChain(
	tx Transaction,
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	if statedb.PrivacyTokenIDExisted(transactionStateDB, iReq.TokenID) {
		return false, NewMetadataTxError(InitPTokenRequestValidateTxWithBlockChainError, errors.New(fmt.Sprintf("The token id (%s) was already existed", iReq.TokenID.String())))
	}
	return true, nil
}

func (iReq InitPTokenRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	if len(iReq.ReceiverAddress.Pk) == 0 {
		return false, false, NewMetadataTxError(InitPTokenRequestValidateSanityDataError, errors.New("Wrong request info's receiver address"))
	}
	if iReq.Amount == 0 {
		return false, false, errors.New("Wrong request info's amount")
	}
	if iReq.TokenName == "" {
		return false, false, NewMetadataTxError(InitPTokenRequestValidateSanityDataError, errors.New("Wrong request info's token name"))
	}
	if iReq.TokenSymbol == "" {
		return false, false, NewMetadataTxError(InitPTokenRequestValidateSanityDataError, errors.New("Wrong request info's token symbol"))
	}
	if iReq.TokenID.String() == "" {
		return false, false, NewMetadataTxError(InitPTokenRequestValidateSanityDataError, errors.New("Wrong request info's token ID"))
	}
	return true, true, nil
}

func (iReq InitPTokenRequest) ValidateMetadataByItself() bool {
	return iReq.Type == InitPTokenRequestMeta
}

func (iReq InitPTokenRequest) Hash() *common.Hash {
	record := iReq.ReceiverAddress.String()
	record += iReq.TokenID.String()
	// TODO: @hung change to record += fmt.Sprint(iReq.DepositedAmount)
	record += fmt.Sprint(iReq.Amount)
	record += iReq.TokenName
	record += iReq.TokenSymbol
	record += iReq.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iReq *InitPTokenRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte) ([][]string, error) {
	txReqID := *(tx.Hash())
	actionContent := map[string]interface{}{
		"meta":    *iReq,
		"txReqId": txReqID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, NewMetadataTxError(InitPTokenRequestBuildReqActionsError, err)
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(InitPTokenRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (iReq *InitPTokenRequest) CalculateSize() uint64 {
	return calculateSize(iReq)
}
