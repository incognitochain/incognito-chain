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

// RelayingBNBHeader - relaying header chain
// metadata - create normal tx with this metadata
type RelayingBNBHeader struct {
	MetadataBase
	IncogAddressStr string
	Header          string
	BlockHeight     uint64
}

// RelayingBNBHeaderAction - shard validator creates instruction that contain this action content
// it will be append to ShardToBeaconBlock
type RelayingBNBHeaderAction struct {
	Meta    RelayingBNBHeader
	TxReqID common.Hash
	ShardID byte
}

// PortalCustodianDepositContent - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and refund status
type RelayingBNBHeaderContent struct {
	IncogAddressStr string
	Header          string
	BlockHeight     uint64
	TxReqID         common.Hash
}

// PortalCustodianDepositStatus - Beacon tracks status of custodian deposit tx into db
type RelayingBNBHeaderStatus struct {
	Status          byte
	IncogAddressStr string
	Header          string
	BlockHeight     uint64
}

func NewRelayingBNBHeader(metaType int, incognitoAddrStr string, header string, blockHeight uint64) (*RelayingBNBHeader, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	custodianDepositMeta := &RelayingBNBHeader{
		IncogAddressStr: incognitoAddrStr,
		Header:          header,
		BlockHeight:     blockHeight,
	}
	custodianDepositMeta.MetadataBase = metadataBase
	return custodianDepositMeta, nil
}

//todo
func (headerRelaying RelayingBNBHeader) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (headerRelaying RelayingBNBHeader) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
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

func (headerRelaying RelayingBNBHeader) ValidateMetadataByItself() bool {
	return headerRelaying.Type == RelayingBNBHeaderMeta
}

func (headerRelaying RelayingBNBHeader) Hash() *common.Hash {
	record := headerRelaying.MetadataBase.Hash().String()
	record += headerRelaying.IncogAddressStr
	record += headerRelaying.Header
	record += strconv.Itoa(int(headerRelaying.BlockHeight))

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (headerRelaying *RelayingBNBHeader) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := RelayingBNBHeaderAction{
		Meta:    *headerRelaying,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(RelayingBNBHeaderMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (headerRelaying *RelayingBNBHeader) CalculateSize() uint64 {
	return calculateSize(headerRelaying)
}
