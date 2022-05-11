package bridge

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

const EVENT_TYPE = "from_contract"

type IssuingWasmRequest struct {
	ShieldTx   string
	IncTokenID common.Hash
	NetworkID  uint `json:"NetworkID,omitempty"`
	metadataCommon.MetadataBase
}

type IssuingWasmReqAction struct {
	Meta             IssuingWasmRequest `json:"meta"`
	TxReqID          common.Hash        `json:"txReqId"`
	ContractId       string             `json:"contractId"`
	IncognitoAddrStr string             `json:"incognitoAddrStr"`
	Token            []byte             `json:"token"`
	Amount           uint64             `json:"amount"`
}

func NewIssuingWasmRequest(
	shieldTx string,
	incTokenID common.Hash,
	networkID uint,
	metaType int,
) (*IssuingWasmRequest, error) {
	metadataBase := metadataCommon.MetadataBase{
		Type: metaType,
	}
	issuingWasmReq := &IssuingWasmRequest{
		ShieldTx:   shieldTx,
		IncTokenID: incTokenID,
		NetworkID:  networkID,
	}
	issuingWasmReq.MetadataBase = metadataBase
	return issuingWasmReq, nil
}

func NewIssuingWasmRequestFromMap(
	data map[string]interface{},
	networkID uint,
	metatype int,
) (*IssuingWasmRequest, error) {
	shieldTx := data["ShieldTx"].(string)

	incTokenID, err := common.Hash{}.NewHashFromStr(data["IncTokenID"].(string))
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestNewIssuingEVMRequestFromMapError, errors.Errorf("TokenID incorrect"))
	}

	req, _ := NewIssuingWasmRequest(
		shieldTx,
		*incTokenID,
		networkID,
		metatype,
	)
	return req, nil
}

func ParseWasmIssuingInstContent(instContentStr string) (*IssuingWasmReqAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instContentStr)
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestDecodeInstructionError, err)
	}
	var issuingWasmReqAction IssuingWasmReqAction
	err = json.Unmarshal(contentBytes, &issuingWasmReqAction)
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestDecodeInstructionError, err)
	}
	return &issuingWasmReqAction, nil
}

func (iReq IssuingWasmRequest) ValidateMetadataByItself() bool {
	if iReq.Type != metadataCommon.IssuingTerraRequestMeta {
		return false
	}

	_, _, _, _, err := iReq.VerifyWasmTx()
	if err != nil {
		metadataCommon.Logger.Log.Error(metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestValidateTxWithBlockChainError, err))
		return false
	}
	return true
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Event struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

type Log struct {
	Events []Event `json:"events"`
}

type GetWasmTransaction struct {
	Height string `json:"height"`
	TxHash string `json:"txHash"`
	Data   string `json:"data"`
	RawLog string `json:"raw_log"`
	Logs   []Log  `json:"logs"`
}

type Header struct {
	Height string `json:"height"`
}

type Block struct {
	Header `json:"header"`
}

type GetWasmBlock struct {
	Block `json:"block"`
}

// VerifyWasmTx
// returns (contractId, externalTokenId, incognitoAddrStr, amount )
func (iReq *IssuingWasmRequest) VerifyWasmTx() (string, string, string, uint64, error) {
	hosts, _, minWasmConfirmationBlocks, _ := GetWasmInfoByMetadataType()
	for _, h := range hosts {
		url := fmt.Sprintf("%s/txs/%s", h, iReq.ShieldTx)
		// get transaction
		body, err := makeGetRequest(url)
		if err != nil {
			continue
		}

		tx := GetWasmTransaction{}
		err = json.Unmarshal(body, &tx)
		if err != nil {
			return "", "", "", 0, err
		}

		if len(tx.Logs) == 0 {
			return "", "", "", 0, errors.New("Invalid shield transaction")
		}
		height, err := strconv.ParseUint(tx.Height, 10, 64)
		if err != nil {
			return "", "", "", 0, err
		}

		// get latest block height
		url = fmt.Sprintf("%s/block/latest", h)
		body, err = makeGetRequest(url)
		if err != nil {
			continue
		}
		block := GetWasmBlock{}
		err = json.Unmarshal(body, &block)
		if err != nil {
			return "", "", "", 0, err
		}
		latestHeight, err := strconv.ParseUint(block.Block.Header.Height, 10, 64)
		if err != nil {
			return "", "", "", 0, err
		}
		if latestHeight < height+uint64(minWasmConfirmationBlocks) {
			return "", "", "", 0, errors.New("The shield transaction not finality yet")
		}
		for _, logData := range tx.Logs {
			for _, event := range logData.Events {
				if event.Type == EVENT_TYPE && len(event.Attributes) == 4 {
					if event.Attributes[0].Key != "contract_address" {
						continue
					}
					if event.Attributes[1].Key != "token" {
						continue
					}
					if event.Attributes[2].Key != "incognitoAddress" {
						continue
					}
					if event.Attributes[3].Key != "value" {
						continue
					}
					amount, err := strconv.ParseUint(event.Attributes[3].Value, 10, 64)
					if err != nil {
						continue
					}

					return event.Attributes[0].Value,
						event.Attributes[1].Value,
						event.Attributes[2].Value,
						amount, nil
				}
			}
		}
	}

	return "", "", "", 0, errors.New("Hosts are not available")
}

func (iReq IssuingWasmRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (iReq IssuingWasmRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	if len(iReq.ShieldTx) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestValidateSanityDataError, errors.New("Wrong request info's shield tx id"))
	}

	return true, true, nil
}

func (iReq IssuingWasmRequest) Hash() *common.Hash {
	record := iReq.ShieldTx
	record += iReq.MetadataBase.Hash().String()
	record += iReq.IncTokenID.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iReq *IssuingWasmRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	contractId, externalToken, incognitoAddrStr, amount, err := iReq.VerifyWasmTx()
	if err != nil {
		return [][]string{}, metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestBuildReqActionsError, err)
	}

	txReqID := *(tx.Hash())
	actionContent := map[string]interface{}{
		"meta":             *iReq,
		"txReqId":          txReqID,
		"contractId":       contractId,
		"incognitoAddrStr": incognitoAddrStr,
		"token":            externalToken,
		"amount":           amount,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestBuildReqActionsError, err)
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(iReq.Type), actionContentBase64Str}

	return [][]string{action}, nil
}

func (iReq *IssuingWasmRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(iReq)
}

func makeGetRequest(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return body, nil
}
