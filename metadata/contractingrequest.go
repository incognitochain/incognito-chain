package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/privacy"
)

// whoever can send this type of tx
type ContractingRequest struct {
	BurnerAddress privacy.PaymentAddress
	BurnedAmount  uint64 // must be equal to vout value
	TokenID       common.Hash
	MetadataBase
}

type ContractingReqAction struct {
	Meta    ContractingRequest `json:"meta"`
	TxReqID common.Hash        `json:"txReqId"`
}

func NewContractingRequest(
	burnerAddress privacy.PaymentAddress,
	burnedAmount uint64,
	tokenID common.Hash,
	metaType int,
) (*ContractingRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	contractingReq := &ContractingRequest{
		TokenID:       tokenID,
		BurnedAmount:  burnedAmount,
		BurnerAddress: burnerAddress,
	}
	contractingReq.MetadataBase = metadataBase
	return contractingReq, nil
}

func (cReq ContractingRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	bridgeTokenExisted, err := db.IsBridgeTokenExistedByType(cReq.TokenID, true)
	if err != nil {
		return false, err
	}
	if !bridgeTokenExisted {
		return false, errors.New("the burning token is not existed in bridge tokens")
	}
	return true, nil
}

func (cReq ContractingRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction, beaconHeight uint64) (bool, bool, error) {

	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	if reflect.TypeOf(txr).String() == "*transaction.Tx" {
		return true, true, nil
	}

	if cReq.Type != ContractingRequestMeta {
		return false, false, errors.New("Wrong request info's meta type")
	}
	if len(cReq.BurnerAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's burner address")
	}
	if cReq.BurnedAmount == 0 {
		return false, false, errors.New("Wrong request info's burned amount")
	}
	if !txr.IsCoinsBurning(bcr, beaconHeight) {
		return false, false, errors.New("Must send coin to burning address")
	}
	if cReq.BurnedAmount != txr.CalculateTxValue() {
		return false, false, errors.New("BurnedAmount incorrect")
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], cReq.BurnerAddress.Pk[:]) {
		return false, false, errors.New("BurnerAddress incorrect")
	}
	return true, true, nil
}

func (cReq ContractingRequest) ValidateMetadataByItself() bool {
	return cReq.Type == ContractingRequestMeta
}

func (cReq ContractingRequest) Hash() *common.Hash {
	record := cReq.MetadataBase.Hash().String()
	record += cReq.BurnerAddress.String()
	record += cReq.TokenID.String()
	record += string(cReq.BurnedAmount)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (cReq *ContractingRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := map[string]interface{}{
		"meta":          *cReq,
		"RequestedTxID": tx.Hash(),
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(ContractingRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (cReq *ContractingRequest) CalculateSize() uint64 {
	return calculateSize(cReq)
}
