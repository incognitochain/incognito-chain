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

// PortalCustodianDepositAction - shard validator creates instruction that contain this action content
// it will be append to ShardToBeaconBlock
type RelayingHeaderAction struct {
	Meta    RelayingHeader
	TxReqID common.Hash
	ShardID byte
}

// PortalCustodianDepositContent - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and refund status
type RelayingHeaderContent struct {
	IncogAddressStr string
	Header          string
	BlockHeight     uint64
	TxReqID         common.Hash
}

// PortalCustodianDepositStatus - Beacon tracks status of custodian deposit tx into db
type RelayingHeaderStatus struct {
	Status          byte
	IncogAddressStr string
	Header          string
	BlockHeight     uint64
}

func NewRelayingHeader(metaType int, incognitoAddrStr string, header string, blockHeight uint64) (*RelayingHeader, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	custodianDepositMeta := &RelayingHeader{
		IncogAddressStr: incognitoAddrStr,
		Header:          header,
		BlockHeight:     blockHeight,
	}
	custodianDepositMeta.MetadataBase = metadataBase
	return custodianDepositMeta, nil
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

func (headerRelaying RelayingHeader) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	//if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
	//	return true, true, nil
	//}

	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(headerRelaying.IncogAddressStr)
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
	if headerRelaying.BlockHeight < 1 {
		return false, false, errors.New("BlockHeight must be greater than 0")
	}

	// check header
	headerBytes, err := base64.StdEncoding.DecodeString(headerRelaying.Header)
	if err != nil || len(headerBytes) == 0 {
		return false, false, errors.New("header is invalid")
	}

	return true, true, nil
}

func (headerRelaying RelayingHeader) ValidateMetadataByItself() bool {
	return headerRelaying.Type == RelayingHeaderMeta
}

func (headerRelaying RelayingHeader) Hash() *common.Hash {
	record := headerRelaying.MetadataBase.Hash().String()
	record += headerRelaying.IncogAddressStr
	record += headerRelaying.Header

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (headerRelaying *RelayingHeader) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := RelayingHeaderAction{
		Meta:    *headerRelaying,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(RelayingHeaderMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (headerRelaying *RelayingHeader) CalculateSize() uint64 {
	return calculateSize(headerRelaying)
}
