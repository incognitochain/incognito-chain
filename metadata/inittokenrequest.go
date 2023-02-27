package metadata

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy/coin"
)

type InitTokenRequest struct {
	OTAStr      string
	TxRandomStr string
	Amount      uint64
	TokenName   string
	TokenSymbol string
	MetadataBase
}

type InitTokenReqAction struct {
	Meta    InitTokenRequest `json:"meta"`
	TxReqID common.Hash      `json:"txReqID"`
	TokenID common.Hash      `json:"tokenID"`
}

type InitTokenAcceptedInst struct {
	OTAStr        string      `json:"OTAStr"`
	TxRandomStr   string      `json:"TxRandomStr"`
	Amount        uint64      `json:"Amount"`
	TokenID       common.Hash `json:"TokenID"`
	TokenName     string      `json:"TokenName"`
	TokenSymbol   string      `json:"TokenSymbol"`
	TokenType     int         `json:"TokenType"`
	ShardID       byte        `json:"ShardID"`
	RequestedTxID common.Hash `json:"txReqID"`
}

func ParseInitTokenInstContent(instContentStr string) (*InitTokenReqAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instContentStr)
	if err != nil {
		return nil, NewMetadataTxError(InitTokenRequestDecodeInstructionError, err)
	}
	var initPTokenReqAction InitTokenReqAction
	err = json.Unmarshal(contentBytes, &initPTokenReqAction)
	if err != nil {
		return nil, NewMetadataTxError(InitTokenRequestUnmarshalJsonError, err)
	}
	return &initPTokenReqAction, nil
}

func ParseInitTokenInstAcceptedContent(instContentStr string) (*InitTokenAcceptedInst, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instContentStr)
	if err != nil {
		return nil, NewMetadataTxError(InitTokenRequestDecodeInstructionError, err)
	}
	var initPTokenReqAction InitTokenAcceptedInst
	err = json.Unmarshal(contentBytes, &initPTokenReqAction)
	if err != nil {
		return nil, NewMetadataTxError(InitTokenRequestUnmarshalJsonError, err)
	}
	return &initPTokenReqAction, nil
}

func NewInitTokenRequest(otaStr string, txRandomStr string, amount uint64, tokenName, tokenSymbol string, metaType int) (*InitTokenRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	initPTokenMeta := &InitTokenRequest{
		OTAStr:      otaStr,
		TxRandomStr: txRandomStr,
		TokenName:   tokenName,
		TokenSymbol: tokenSymbol,
		Amount:      amount,
	}
	initPTokenMeta.MetadataBase = metadataBase
	return initPTokenMeta, nil
}

func (iReq InitTokenRequest) ValidateMetadataByItself() bool {
	return iReq.Type == InitTokenRequestMeta
}

func (iReq InitTokenRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

//ValidateSanityData performs the following verifications:
//	1. Check transaction type and tx version
//	2. Check the addressV2 is valid
//	3. Check if the amount is not 0
//	4. Check tokenName and tokenSymbol
func (iReq InitTokenRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	//Step 1
	if tx.GetType() != common.TxNormalType {
		return false, false, NewMetadataTxError(InitTokenRequestValidateSanityDataError, fmt.Errorf("tx InitTokenRequest must have type `%v`", common.TxNormalType))
	}
	if tx.GetVersion() != 2 {
		return false, false, NewMetadataTxError(InitTokenRequestValidateSanityDataError, fmt.Errorf("metadata %v only supports tx ver 2", InitTokenRequestMeta))
	}

	//Step 2
	recvPubKey, _, err := coin.ParseOTAInfoFromString(iReq.OTAStr, iReq.TxRandomStr)
	if err != nil {
		return false, false, NewMetadataTxError(InitTokenRequestValidateSanityDataError, fmt.Errorf("cannot parse OTA params (%v, %v): %v", iReq.OTAStr, iReq.TxRandomStr, err))
	}
	recvKeyBytes := recvPubKey.ToBytesS()
	senderShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	receiverShardID := common.GetShardIDFromLastByte(recvKeyBytes[len(recvKeyBytes)-1])
	if senderShardID != receiverShardID {
		return false, false, NewMetadataTxError(InitTokenRequestValidateSanityDataError, fmt.Errorf("sender shardID and receiver shardID mismatch: %v != %v", senderShardID, receiverShardID))
	}

	//Step 3
	if iReq.Amount == 0 {
		return false, false, NewMetadataTxError(InitTokenRequestValidateSanityDataError, fmt.Errorf("initialized amount must not be 0"))
	}

	//Step 4
	if len(iReq.TokenName) == 0 {
		return false, false, NewMetadataTxError(InitTokenRequestValidateSanityDataError, fmt.Errorf("tokenName must not be empty"))
	}
	if len(iReq.TokenSymbol) == 0 {
		return false, false, NewMetadataTxError(InitTokenRequestValidateSanityDataError, fmt.Errorf("tokenSymbol must not be empty"))
	}

	return true, true, nil
}

//Hash returns the hash of all components in the request.
func (iReq InitTokenRequest) Hash() *common.Hash {
	record := iReq.MetadataBase.Hash().String()
	record += iReq.OTAStr
	record += iReq.TxRandomStr
	record += iReq.TokenName
	record += iReq.TokenSymbol
	record += strconv.FormatUint(iReq.Amount, 10)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

//genTokenID generates a (deterministically) random tokenID for the request transaction.
//From now on, users cannot generate their own tokenID.
//The generated tokenID is calculated as the hash of the following components:
//	- The Tx hash
//	- The shardID at which the request is sent
func (iReq *InitTokenRequest) genTokenID(tx Transaction, shardID byte) *common.Hash {
	record := tx.Hash().String()
	record += strconv.FormatUint(uint64(shardID), 10)

	tokenID := common.HashH([]byte(record))
	return &tokenID
}

//BuildReqActions builds an request action content from the shard chain to the beacon chain.
func (iReq *InitTokenRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	tokenID := GenTokenIDFromRequest(tx.Hash().String(), shardID)
	txReqID := tx.Hash()
	actionContent := map[string]interface{}{
		"meta":    *iReq,
		"txReqID": *txReqID,
		"tokenID": *tokenID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, NewMetadataTxError(InitTokenRequestBuildReqActionsError, err)
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(InitTokenRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

//CalculateSize returns the size of the request.
func (iReq *InitTokenRequest) CalculateSize() uint64 {
	return calculateSize(iReq)
}

func (iReq *InitTokenRequest) ToCompactBytes() ([]byte, error) {
	return toCompactBytes(iReq)
}

func (iReq *InitTokenRequest) GetOTADeclarations() []OTADeclaration {
	pk, _, err := coin.ParseOTAInfoFromString(iReq.OTAStr, iReq.TxRandomStr)
	if err != nil {
		return []OTADeclaration{}
	}
	result := OTADeclaration{PublicKey: pk.ToBytes(), TokenID: common.ConfidentialAssetID}
	return []OTADeclaration{result}
}
