package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

// RelayingHeader - relaying header chain
// metadata - create normal tx with this metadata
type RelayingHeader struct {
	MetadataBase
	IncogAddressStr string
	Header          string
	BlockHeight     uint64
}

// RelayingHeaderAction - shard validator creates instruction that contain this action content
// it will be append to ShardToBeaconBlock
type RelayingHeaderAction struct {
	Meta    RelayingHeader
	TxReqID common.Hash
	ShardID byte
}

// RelayingHeaderContent - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and refund status
type RelayingHeaderContent struct {
	IncogAddressStr string
	Header          string
	BlockHeight     uint64
	TxReqID         common.Hash
}

// RelayingHeaderStatus - Beacon tracks status of custodian deposit tx into db
type RelayingHeaderStatus struct {
	Status          byte
	IncogAddressStr string
	Header          string
	BlockHeight     uint64
}

func NewRelayingHeader(
	metaType int,
	incognitoAddrStr string,
	header string,
	blockHeight uint64,
) (*RelayingHeader, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	relayingHeader := &RelayingHeader{
		IncogAddressStr: incognitoAddrStr,
		Header:          header,
		BlockHeight:     blockHeight,
	}
	relayingHeader.MetadataBase = metadataBase
	return relayingHeader, nil
}

//todo
func (headerRelaying RelayingHeader) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (rh RelayingHeader) ValidateSanityData(bcr BlockchainRetriever, txr Transaction, beaconHeight uint64) (bool, bool, error) {
	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(rh.IncogAddressStr)
	if err != nil {
		return false, false, NewMetadataTxError(IssuingRequestNewIssuingRequestFromMapEror, errors.New("sender address is incorrect"))
	}
	incogAddr := keyWallet.KeySet.PaymentAddress
	if len(incogAddr.Pk) == 0 {
		return false, false, errors.New("wrong sender address")
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incogAddr.Pk[:]) {
		return false, false, errors.New("sender address is not signer tx")
	}

	// check tx type
	if txr.GetType() != common.TxNormalType {
		return false, false, errors.New("tx push header relaying must be TxNormalType")
	}

	// check block height
	if rh.BlockHeight < 1 {
		return false, false, errors.New("BlockHeight must be greater than 0")
	}

	// check header
	headerBytes, err := base64.StdEncoding.DecodeString(rh.Header)
	if err != nil || len(headerBytes) == 0 {
		return false, false, errors.New("header is invalid")
	}

	return true, true, nil
}

func (rh RelayingHeader) ValidateMetadataByItself() bool {
	return rh.Type == RelayingBNBHeaderMeta || rh.Type == RelayingBTCHeaderMeta
}

func (rh RelayingHeader) Hash() *common.Hash {
	record := rh.MetadataBase.Hash().String()
	record += rh.IncogAddressStr
	record += rh.Header
	record += strconv.Itoa(int(rh.BlockHeight))

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (rh *RelayingHeader) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := RelayingHeaderAction{
		Meta:    *rh,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(rh.Type), actionContentBase64Str}
	return [][]string{action}, nil
}

func (rh *RelayingHeader) CalculateSize() uint64 {
	return calculateSize(rh)
}
