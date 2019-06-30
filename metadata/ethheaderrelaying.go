package metadata

import (
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	rCommon "github.com/incognitochain/incognito-chain/ethrelaying/common"
	"github.com/incognitochain/incognito-chain/ethrelaying/core/types"
	privacy "github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type ETHHeaderRelaying struct {
	RelayerAddress privacy.PaymentAddress
	ETHHeaders     []*types.Header
	MetadataBase
}

func NewETHHeaderRelaying(
	relayerAddress privacy.PaymentAddress,
	ethHeaders []*types.Header,
	metaType int,
) *ETHHeaderRelaying {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	ethHeaderRelaying := &ETHHeaderRelaying{
		RelayerAddress: relayerAddress,
		ETHHeaders:     ethHeaders,
	}
	ethHeaderRelaying.MetadataBase = metadataBase
	return ethHeaderRelaying
}

func NewETHHeaderRelayingFromMap(data map[string]interface{}) (Metadata, error) {
	ethHeadersRaw, ok := data["ETHHeaders"].([]interface{})
	if !ok {
		return nil, errors.Errorf("ETHHeaders incorrect")
	}
	ethHeaders := []*types.Header{}
	for _, item := range ethHeadersRaw {
		ethHeaderRaw, ok := item.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("ETHHeader incorrect")
		}
		ethHeader := &types.Header{
			ParentHash:  rCommon.HexToHash(ethHeaderRaw["parentHash"].(string)),
			UncleHash:   rCommon.HexToHash(ethHeaderRaw["sha3Uncles"].(string)),
			Coinbase:    rCommon.HexToAddress(ethHeaderRaw["miner"].(string)),
			Root:        rCommon.HexToHash(ethHeaderRaw["stateRoot"].(string)),
			TxHash:      rCommon.HexToHash(ethHeaderRaw["transactionsRoot"].(string)),
			ReceiptHash: rCommon.HexToHash(ethHeaderRaw["receiptsRoot"].(string)),
			Bloom:       types.BytesToBloom(rCommon.FromHex(ethHeaderRaw["logsBloom"].(string))),
			Difficulty:  rCommon.HexToHash(ethHeaderRaw["difficulty"].(string)).Big(),
			Number:      rCommon.HexToHash(ethHeaderRaw["number"].(string)).Big(),
			GasLimit:    rCommon.HexToHash(ethHeaderRaw["gasLimit"].(string)).Big().Uint64(),
			GasUsed:     rCommon.HexToHash(ethHeaderRaw["gasUsed"].(string)).Big().Uint64(),
			Time:        rCommon.HexToHash(ethHeaderRaw["timestamp"].(string)).Big().Uint64(),
			Extra:       rCommon.FromHex(ethHeaderRaw["extraData"].(string)),
			MixDigest:   rCommon.HexToHash(ethHeaderRaw["mixHash"].(string)),
			Nonce:       types.EncodeNonce(rCommon.HexToHash(ethHeaderRaw["nonce"].(string)).Big().Uint64()),
		}
		fmt.Println("huhu num: ", ethHeader.Number)
		ethHeaders = append(ethHeaders, ethHeader)
	}

	relayerAddr, ok := data["RelayerAddress"].(privacy.PaymentAddress)
	if !ok {
		return nil, errors.Errorf("RelayerAddress incorrect")
	}

	return NewETHHeaderRelaying(
		relayerAddr,
		ethHeaders,
		ETHHeaderRelayingMeta,
	), nil
}

func (e *ETHHeaderRelaying) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (e *ETHHeaderRelaying) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(e.RelayerAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's relayer address")
	}
	ethHeadersLen := len(e.ETHHeaders)
	if ethHeadersLen == 0 || ethHeadersLen > 10 {
		return false, false, errors.New("Wrong request info's eth headers length")
	}
	return true, true, nil
}

func (e *ETHHeaderRelaying) ValidateMetadataByItself() bool {
	if e.Type != ETHHeaderRelayingMeta {
		return false
	}
	return true
}

func (e *ETHHeaderRelaying) Hash() *common.Hash {
	record := e.RelayerAddress.String()
	for _, item := range e.ETHHeaders {
		record += item.ParentHash.String()
		record += item.UncleHash.String()
		record += item.Coinbase.String()
		record += item.Root.String()
		record += item.TxHash.String()
		record += item.ReceiptHash.String()
		record += string(item.Bloom[:])
		record += item.Difficulty.String()
		record += item.Number.String()
		record += strconv.FormatUint(item.GasLimit, 10)
		record += strconv.FormatUint(item.GasUsed, 10)
		record += strconv.FormatUint(item.Time, 10)
		record += string(item.Extra)
		record += item.MixDigest.String()
		record += string(item.Nonce[:])
	}
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (e *ETHHeaderRelaying) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	return [][]string{}, nil
}

func (e *ETHHeaderRelaying) CalculateSize() uint64 {
	return calculateSize(e)
}
