package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
)

// whoever can send this type of tx
type ContractingRequest struct {
	BurnerAddress     privacy.PaymentAddress
	BurnedConstAmount uint64      // must be equal to vout value
	CurrencyType      common.Hash // USD or ETH for now
	MetadataBase
}

func NewContractingRequest(
	burnerAddress privacy.PaymentAddress,
	burnedConstAmount uint64,
	currencyType common.Hash,
	metaType int,
) (*ContractingRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	contractingReq := &ContractingRequest{
		CurrencyType:      currencyType,
		BurnedConstAmount: burnedConstAmount,
		BurnerAddress:     burnerAddress,
	}
	contractingReq.MetadataBase = metadataBase
	return contractingReq, nil
}

func NewContractingRequestFromMap(data map[string]interface{}) (Metadata, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(data["BurnerAddress"].(string))
	if err != nil {
		return nil, errors.Errorf("BurnerAddress incorrect")
	}

	burnedConstAmount := uint64(data["BurnedConstAmount"].(float64))

	currencyType, err := common.NewHashFromStr(data["CurrencyType"].(string))
	if err != nil {
		return nil, errors.Errorf("CurrencyType incorrect")
	}
	return NewContractingRequest(
		keyWallet.KeySet.PaymentAddress,
		burnedConstAmount,
		*currencyType,
		IssuingRequestMeta,
	)
}

func (cReq *ContractingRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (cReq *ContractingRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if cReq.Type == ContractingRequestMeta {
		return false, false, errors.New("Wrong request info's meta type")
	}
	if len(cReq.BurnerAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's burner address")
	}
	if cReq.BurnedConstAmount == 0 {
		return false, false, errors.New("Wrong request info's deposited amount")
	}
	if len(cReq.CurrencyType) != common.HashSize {
		return false, false, errors.New("Wrong request info's currency type")
	}

	if !txr.IsCoinsBurning() {
		return false, false, nil
	}
	if cReq.BurnedConstAmount != txr.CalculateTxValue() {
		return false, false, nil
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], cReq.BurnerAddress.Pk[:]) {
		return false, false, nil
	}
	return true, true, nil
}

func (cReq *ContractingRequest) ValidateMetadataByItself() bool {
	return cReq.Type != ContractingRequestMeta
}

func (cReq *ContractingRequest) Hash() *common.Hash {
	record := cReq.MetadataBase.Hash().String()
	record += cReq.BurnerAddress.String()
	record += cReq.CurrencyType.String()
	record += string(cReq.BurnedConstAmount)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (cReq *ContractingRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	if bytes.Equal(cReq.CurrencyType[:], common.USDAssetID[:]) {
		return [][]string{}, nil
	}
	actionContent := map[string]interface{}{
		"txReqId": *(tx.Hash()),
		"meta":    *cReq,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(ContractingRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}
