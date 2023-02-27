package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy/coin"
)

// PDETradeRequest - privacy dex trade
type PDETradeRequest struct {
	TokenIDToBuyStr     string
	TokenIDToSellStr    string
	SellAmount          uint64 // must be equal to vout value
	MinAcceptableAmount uint64
	TradingFee          uint64
	TraderAddressStr    string
	TxRandomStr         string `json:"TxRandomStr,omitempty"`
	MetadataBase
}

type PDETradeRequestAction struct {
	Meta    PDETradeRequest
	TxReqID common.Hash
	ShardID byte
}

type TokenPoolValueOperation struct {
	Operator string
	Value    uint64
}

type PDETradeAcceptedContent struct {
	TraderAddressStr         string
	TxRandomStr              string `json:"TxRandomStr,omitempty"`
	TokenIDToBuyStr          string
	ReceiveAmount            uint64
	Token1IDStr              string
	Token2IDStr              string
	Token1PoolValueOperation TokenPoolValueOperation
	Token2PoolValueOperation TokenPoolValueOperation
	ShardID                  byte
	RequestedTxID            common.Hash
}

func NewPDETradeRequest(
	tokenIDToBuyStr string,
	tokenIDToSellStr string,
	sellAmount uint64,
	minAcceptableAmount uint64,
	tradingFee uint64,
	traderAddressStr string,
	txRandomStr string,
	metaType int,
) (*PDETradeRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	pdeTradeRequest := &PDETradeRequest{
		TokenIDToBuyStr:     tokenIDToBuyStr,
		TokenIDToSellStr:    tokenIDToSellStr,
		SellAmount:          sellAmount,
		MinAcceptableAmount: minAcceptableAmount,
		TradingFee:          tradingFee,
		TraderAddressStr:    traderAddressStr,
		TxRandomStr:         txRandomStr,
	}
	pdeTradeRequest.MetadataBase = metadataBase
	return pdeTradeRequest, nil
}

func (pc PDETradeRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (pc PDETradeRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	if chainRetriever.IsAfterPrivacyV2CheckPoint(beaconHeight) {
		return false, false, fmt.Errorf("metadata type %v is no longer supported, consider using %v instead", PDETradeRequestMeta, PDECrossPoolTradeRequestMeta)
	}

	_, err, ver := metadataCommon.CheckIncognitoAddress(pc.TraderAddressStr, pc.TxRandomStr)
	if err != nil {
		return false, false, err
	}
	if int8(ver) != tx.GetVersion() {
		return false, false, fmt.Errorf("payment address version (%v) and tx version (%v) mismatch", ver, tx.GetVersion())
	}

	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil {
		return false, false, err
	}
	if !isBurned {
		return false, false, errors.New("Error This is not Tx Burn")
	}
	if (pc.SellAmount + pc.TradingFee) != burnCoin.GetValue() {
		return false, false, errors.New("Error Selling amount should be equal to the burned amount")
	}
	if pc.SellAmount > burnCoin.GetValue() || pc.TradingFee > burnCoin.GetValue() {
		return false, false, errors.New("Neither selling amount nor trading fee allows to be larger than the tx value")
	}

	tokenIDToSell, err := common.Hash{}.NewHashFromStr(pc.TokenIDToSellStr)
	if err != nil {
		return false, false, NewMetadataTxError(IssuingRequestNewIssuingRequestFromMapEror, errors.New("TokenIDToSellStr incorrect"))
	}
	if !bytes.Equal(burnedTokenID[:], tokenIDToSell[:]) {
		return false, false, errors.New("Wrong request info's token id, it should be equal to tx's token id.")
	}

	if tx.GetType() == common.TxNormalType && pc.TokenIDToSellStr != common.PRVCoinID.String() {
		return false, false, errors.New("With tx normal privacy, the tokenIDStr should be PRV, not custom token.")
	}

	if tx.GetType() == common.TxCustomTokenPrivacyType && pc.TokenIDToSellStr == common.PRVCoinID.String() {
		return false, false, errors.New("With tx custome token privacy, the tokenIDStr should not be PRV, but custom token.")
	}

	return true, true, nil
}

func (pc PDETradeRequest) ValidateMetadataByItself() bool {
	return pc.Type == PDETradeRequestMeta
}

func (pc PDETradeRequest) Hash() *common.Hash {
	record := pc.MetadataBase.Hash().String()
	record += pc.TokenIDToBuyStr
	record += pc.TokenIDToSellStr
	record += pc.TraderAddressStr
	if len(pc.TxRandomStr) > 0 {
		record += pc.TxRandomStr
	}
	record += strconv.FormatUint(pc.SellAmount, 10)
	record += strconv.FormatUint(pc.MinAcceptableAmount, 10)
	record += strconv.FormatUint(pc.TradingFee, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (pc *PDETradeRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PDETradeRequestAction{
		Meta:    *pc,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}

	//Logger.log.Infof("BUGLOG4 actionContent: %v\n", string(actionContentBytes))

	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PDETradeRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (pc *PDETradeRequest) CalculateSize() uint64 {
	return calculateSize(pc)
}

func (pc *PDETradeRequest) ToCompactBytes() ([]byte, error) {
	return toCompactBytes(pc)
}

func (pc *PDETradeRequest) GetOTADeclarations() []OTADeclaration {
	pk, _, err := coin.ParseOTAInfoFromString(pc.TraderAddressStr, pc.TxRandomStr)
	if err != nil {
		return []OTADeclaration{}
	}
	sellingToken := common.ConfidentialAssetID
	if pc.TokenIDToSellStr == common.PRVIDStr {
		sellingToken = common.PRVCoinID
	}
	result := OTADeclaration{PublicKey: pk.ToBytes(), TokenID: sellingToken}
	return []OTADeclaration{result}
}
