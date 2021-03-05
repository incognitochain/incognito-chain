package metadata

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"strconv"
)

type InitPTokenRequest struct {
	OTAStr      string
	TxRandomStr string
	Amount      uint64
	TokenName   string
	TokenSymbol string
	MetadataBase
}

type InitPTokenReqAction struct {
	Meta    InitPTokenRequest `json:"meta"`
	TxReqID common.Hash       `json:"txReqID"`
	TokenID common.Hash       `json:"tokenID"`
}

type InitPTokenAcceptedInst struct {
	OTAStr        string      `json:"OTAStr"`
	TxRandomStr   string      `json:"TxRandomStr"`
	Amount        uint64      `json:"Amount"`
	TokenID       common.Hash `json:"TokenID"`
	ShardID       byte        `json:"ShardID"`
	RequestedTxID common.Hash `json:"txReqID"`
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

func NewPTokenInitRequest(convertingAddress string, txRandomStr string, amount uint64, tokenName, tokenSymbol string, metaType int) (*InitPTokenRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	initPTokenMeta := &InitPTokenRequest{
		OTAStr:      convertingAddress,
		TxRandomStr: txRandomStr,
		TokenName:   tokenName,
		TokenSymbol: tokenSymbol,
		Amount:      amount,
	}
	initPTokenMeta.MetadataBase = metadataBase
	return initPTokenMeta, nil
}

func (iReq InitPTokenRequest) ValidateMetadataByItself() bool {
	return iReq.Type == InitPTokenRequestMeta
}

func (iReq InitPTokenRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

//ValidateSanityData performs the following verifications:
//	1. Check transaction type
//	2. Check the addressV2 is valid
//	3. Check if the amount is not 0
//	4. Check tokenName and tokenSymbol
func (iReq InitPTokenRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	//Step 1
	if tx.GetType() != common.TxNormalType {
		return false, false, NewMetadataTxError(InitPTokenRequestValidateSanityDataError, fmt.Errorf("tx InitPTokenRequest must have type `%v`", common.TxNormalType))
	}

	//Step 2
	recvPubKey, _, err := coin.ParseOTAInfoFromString(iReq.OTAStr, iReq.TxRandomStr)
	if err != nil {
		return false, false, NewMetadataTxError(InitPTokenRequestValidateSanityDataError, fmt.Errorf("cannot parse OTA params (%v, %v): %v", iReq.OTAStr, iReq.TxRandomStr, err))
	}
	recvKeyBytes := recvPubKey.ToBytesS()
	senderShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	receiverShardID := common.GetShardIDFromLastByte(recvKeyBytes[len(recvKeyBytes) - 1])
	if senderShardID != receiverShardID {
		return false, false, NewMetadataTxError(InitPTokenRequestValidateSanityDataError, fmt.Errorf("sender shardID and receiver shardID mismatch: %v != %v", senderShardID, receiverShardID))
	}

	//Step 3
	if iReq.Amount == 0 {
		return false, false, NewMetadataTxError(InitPTokenRequestValidateSanityDataError, fmt.Errorf("initialized amount must not be 0"))
	}

	//Step 4
	if len(iReq.TokenName) == 0 {
		return false, false, NewMetadataTxError(InitPTokenRequestValidateSanityDataError, fmt.Errorf("tokenName must not be empty"))
	}
	if len(iReq.TokenSymbol) == 0 {
		return false, false, NewMetadataTxError(InitPTokenRequestValidateSanityDataError, fmt.Errorf("tokenSymbol must not be empty"))
	}

	return true, true, nil
}

//Hash returns the hash of all components in the request.
func (iReq InitPTokenRequest) Hash() *common.Hash {
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
//	- The InitPTokenRequest hash
//	- The Tx hash
//	- The shardID at which the request is sent
func (iReq *InitPTokenRequest) genTokenID(tx Transaction, shardID byte) *common.Hash {
	record := iReq.Hash().String()
	record += tx.Hash().String()
	record += strconv.FormatUint(uint64(shardID), 10)

	tokenID := common.HashH([]byte(record))
	return &tokenID
}

//BuildReqActions builds an request action content from the shard chain to the beacon chain.
func (iReq *InitPTokenRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	tokenID := iReq.genTokenID(tx, shardID)
	txReqID := tx.Hash()
	actionContent := map[string]interface{}{
		"meta":    *iReq,
		"txReqID": *txReqID,
		"tokenID": *tokenID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, NewMetadataTxError(InitPTokenRequestBuildReqActionsError, err)
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(InitPTokenRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

//CalculateSize returns the size of the request.
func (iReq *InitPTokenRequest) CalculateSize() uint64 {
	return calculateSize(iReq)
}
