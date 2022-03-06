package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/gagliardetto/solana-go"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

// whoever can send this type of tx
type BurningSOLRequest struct {
	BurningAmount uint64 // must be equal to vout value
	TokenID       common.Hash
	RemoteAddress string
	MetadataBase
}

type BurningSOLRequestAction struct {
	Meta    BurningSOLRequest
	TxReqID common.Hash
}

func NewBurningSOLRequest(
	burningAmount uint64,
	tokenID common.Hash,
	remoteAddress string,
	metaType int,
) (*BurningSOLRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	burningReq := &BurningSOLRequest{
		BurningAmount: burningAmount,
		TokenID:       tokenID,
		RemoteAddress: remoteAddress,
	}
	burningReq.MetadataBase = metadataBase
	return burningReq, nil
}

func (bReq BurningSOLRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (bReq BurningSOLRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	// if reflect.TypeOf(tx).String() == "*transaction.Tx" {
	// 	return true, true, nil
	// }

	if _, err := solana.PublicKeyFromBase58(bReq.RemoteAddress); err != nil {
		return false, false, NewMetadataTxError(metadataCommon.BurningSolReqValidateSanityDataError, fmt.Errorf("invalid remote address %v", err))
	}
	if bReq.BurningAmount == 0 {
		return false, false, NewMetadataTxError(metadataCommon.BurningSolReqValidateSanityDataError, errors.New("burned amount must be greater than zero"))
	}

	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, NewMetadataTxError(metadataCommon.BurningSolReqValidateSanityDataError, fmt.Errorf("it is not transaction burn. Error %v", err))
	}
	if !bytes.Equal(burnedTokenID[:], bReq.TokenID[:]) {
		return false, false, NewMetadataTxError(metadataCommon.BurningSolReqValidateSanityDataError, fmt.Errorf("wrong request info's token id and token burned"))
	}
	if bytes.Equal(bReq.TokenID[:], common.PRVCoinID[:]) {
		return false, false, NewMetadataTxError(metadataCommon.BurningSolReqValidateSanityDataError, fmt.Errorf("burned token must not be PRV"))
	}
	burnAmount := burnCoin.GetValue()
	if burnAmount != bReq.BurningAmount || burnAmount == 0 {
		return false, false, NewMetadataTxError(metadataCommon.BurningSolReqValidateSanityDataError, fmt.Errorf("burn amount is incorrect %v", burnAmount))
	}

	return true, true, nil
}

func (bReq BurningSOLRequest) ValidateMetadataByItself() bool {
	return bReq.Type == BurningSOLRequestMeta || bReq.Type == BurningSOLForDepositToSCRequestMeta
}

func (bReq *BurningSOLRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&bReq)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (bReq *BurningSOLRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		BurningAmount uint64      `json:"BurningAmount"`
		TokenID       common.Hash `json:"TokenID"`
		RemoteAddress string      `json:"RemoteAddress"`
		MetadataBase
	}{
		BurningAmount: bReq.BurningAmount,
		TokenID:       bReq.TokenID,
		RemoteAddress: bReq.RemoteAddress,
		MetadataBase:  bReq.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (bReq *BurningSOLRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		BurningAmount uint64      `json:"BurningAmount"`
		TokenID       common.Hash `json:"TokenID"`
		RemoteAddress string      `json:"RemoteAddress"`
		MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	bReq.BurningAmount = temp.BurningAmount
	bReq.TokenID = temp.TokenID
	bReq.RemoteAddress = temp.RemoteAddress
	bReq.MetadataBase = temp.MetadataBase

	return nil
}
func (bReq BurningSOLRequest) HashWithoutSig() *common.Hash {
	rawBytes, _ := json.Marshal(&bReq)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (bReq *BurningSOLRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := BurningSOLRequestAction{
		Meta:    *bReq,
		TxReqID: *tx.Hash(),
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(bReq.Type), actionContentBase64Str}
	return [][]string{action}, nil
}

func (bReq *BurningSOLRequest) CalculateSize() uint64 {
	return calculateSize(bReq)
}
