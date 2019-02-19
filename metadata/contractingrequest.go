package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
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
) *ContractingRequest {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	contractingReq := &ContractingRequest{
		CurrencyType:      currencyType,
		BurnedConstAmount: burnedConstAmount,
		BurnerAddress:     burnerAddress,
	}
	contractingReq.MetadataBase = metadataBase
	return contractingReq
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
	if !txr.IsCoinsBurning() {
		return false, false, nil
	}
	// TODO: compare BurnedConstAmount to vout value
	// TODO: check buner address is the one in vin
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
