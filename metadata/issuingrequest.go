package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	privacy "github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/pkg/errors"
)

// only centralized website can send this type of tx
type IssuingRequest struct {
	ReceiverAddress privacy.PaymentAddress
	DepositedAmount uint64
	AssetType       common.Hash // token id (one of types: Constant, BANK)
	CurrencyType    common.Hash // USD or ETH for now
	MetadataBase
}

func NewIssuingRequest(
	receiverAddress privacy.PaymentAddress,
	depositedAmount uint64,
	assetType common.Hash,
	currencyType common.Hash,
	metaType int,
) (*IssuingRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	issuingReq := &IssuingRequest{
		ReceiverAddress: receiverAddress,
		DepositedAmount: depositedAmount,
		AssetType:       assetType,
		CurrencyType:    currencyType,
	}
	issuingReq.MetadataBase = metadataBase
	return issuingReq, nil
}

func NewIssuingRequestFromMap(data map[string]interface{}) (Metadata, error) {
	n := new(big.Int)
	n, ok := n.SetString(data["DepositedAmount"].(string), 10)
	if !ok {
		return nil, errors.Errorf("DepositedAmount incorrect")
	}
	currencyType, err := common.NewHashFromStr(data["CurrencyType"].(string))
	if err != nil {
		return nil, errors.Errorf("CurrencyType incorrect")
	}

	depositedAmt := uint64(0)
	if bytes.Equal(currencyType[:], common.USDAssetID[:]) {
		depositedAmtStr := data["DepositedAmount"].(string)
		depositedAmtInt, err := strconv.Atoi(depositedAmtStr)
		if err != nil {
			return nil, err
		}
		depositedAmt = uint64(depositedAmtInt)
	} else {
		// Convert from Wei to MilliEther
		denominator := big.NewInt(common.WeiToMilliEtherRatio)
		n = n.Quo(n, denominator)
		if !n.IsUint64() {
			return nil, errors.Errorf("DepositedAmount cannot be converted into uint64")
		}
		depositedAmt = n.Uint64()
	}

	keyWallet, err := wallet.Base58CheckDeserialize(data["ReceiveAddress"].(string))
	if err != nil {
		return nil, errors.Errorf("ReceiveAddress incorrect")
	}

	assetType, err := common.NewHashFromStr(data["AssetType"].(string))
	if err != nil {
		return nil, errors.Errorf("AssetType incorrect")
	}

	return NewIssuingRequest(
		keyWallet.KeySet.PaymentAddress,
		depositedAmt,
		*assetType,
		*currencyType,
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
		return false, errors.New("The issuance request must be called by centralized website.")
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
	if len(iReq.AssetType) != common.HashSize {
		return false, false, errors.New("Wrong request info's asset type")
	}
	if len(iReq.CurrencyType) != common.HashSize {
		return false, false, errors.New("Wrong request info's currency type")
	}
	return true, true, nil
}

func (iReq *IssuingRequest) ValidateMetadataByItself() bool {
	if iReq.Type != IssuingRequestMeta {
		return false
	}
	if !bytes.Equal(iReq.CurrencyType[:], common.USDAssetID[:]) &&
		!bytes.Equal(iReq.CurrencyType[:], common.ETHAssetID[:]) {
		return false
	}
	if !bytes.Equal(iReq.AssetType[:], common.ConstantID[:]) &&
		!bytes.Equal(iReq.AssetType[:], common.DCBTokenID[:]) {
		return false
	}
	if bytes.Equal(iReq.CurrencyType[:], common.ETHAssetID[:]) &&
		!bytes.Equal(iReq.AssetType[:], common.DCBTokenID[:]) {
		return false
	}
	return true
}

func (iReq *IssuingRequest) Hash() *common.Hash {
	record := iReq.ReceiverAddress.String()
	record += iReq.AssetType.String()
	record += iReq.CurrencyType.String()
	record += string(iReq.DepositedAmount)
	record += iReq.MetadataBase.Hash().String()

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (iReq *IssuingRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	pkLastByte := iReq.ReceiverAddress.Pk[len(iReq.ReceiverAddress.Pk)-1]
	receiverShardID := common.GetShardIDFromLastByte(pkLastByte)
	actionContent := map[string]interface{}{
		"txReqId":         *(tx.Hash()),
		"receiverShardID": receiverShardID,
		"meta":            *iReq,
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
