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

// PortalCustodianDeposit - portal custodian deposit collateral (PRV)
type PortalCustodianDeposit struct {
	MetadataBase
	IncogAddressStr string
	RemoteAddresses map[string]string		// token symbol: remote address
	DepositedAmount uint64
}

type PortalCustodianDepositAction struct {
	Meta    PortalCustodianDeposit
	TxReqID common.Hash
	ShardID byte
}

// PortalCustodianDepositContent - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and refund status
type PortalCustodianDepositContent struct {
	IncogAddressStr string
	RemoteAddresses map[string]string
	DepositedAmount uint64
	TxReqID         common.Hash
	ShardID byte
}

// PortalCustodianDepositStatus - Beacon tracks status of custodian deposit tx into db
type PortalCustodianDepositStatus struct {
	Status byte
	IncogAddressStr string
	RemoteAddresses map[string]string
	DepositedAmount uint64
	TxReqID         common.Hash
}

func NewPortalCustodianDeposit(metaType int, incognitoAddrStr string, remoteAddrs map[string]string, amount uint64) (*PortalCustodianDeposit, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	custodianDepositMeta := &PortalCustodianDeposit{
		IncogAddressStr: incognitoAddrStr,
		RemoteAddresses: remoteAddrs,
		DepositedAmount: amount,
	}
	custodianDepositMeta.MetadataBase = metadataBase
	return custodianDepositMeta, nil
}

//todo
func (custodianDeposit PortalCustodianDeposit) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (custodianDeposit PortalCustodianDeposit) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	//if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
	//	return true, true, nil
	//}

	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(custodianDeposit.IncogAddressStr)
	if err != nil {
		return false, false, NewMetadataTxError(IssuingRequestNewIssuingRequestFromMapEror, errors.New("ContributorAddressStr incorrect"))
	}
	incogAddr := keyWallet.KeySet.PaymentAddress
	if len(incogAddr.Pk) == 0 {
		return false, false, errors.New("wrong custodian incognito address")
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incogAddr.Pk[:]) {
		return false, false, errors.New("custodian incognito address is not signer tx")
	}

	// check tx type
	if txr.GetType() != common.TxNormalType {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}

	// check burning tx
	if !txr.IsCoinsBurning(bcr) {
		return false, false, errors.New("must send coin to burning address")
	}

	// validate amount deposit
	if custodianDeposit.DepositedAmount == 0 {
		return false, false, errors.New("deposit amount should be larger than 0")
	}
	if custodianDeposit.DepositedAmount != txr.CalculateTxValue() {
		return false, false, errors.New("deposit amount should be equal to the tx value")
	}

	// validate remote addresses
	if len(custodianDeposit.RemoteAddresses) == 0 {
		return false, false, errors.New("remote addresses should be at least one")
	}
	for tokenSymbol, _ := range custodianDeposit.RemoteAddresses {
		isSupportedToken, err := common.SliceExists(PortalSupportedTokenSymbols, tokenSymbol)
		if err != nil || !isSupportedToken {
			return false, false, errors.New("remote address is invalid")
		}
	}

	return true, true, nil
}

func (custodianDeposit PortalCustodianDeposit) ValidateMetadataByItself() bool {
	return custodianDeposit.Type == PortalCustodianDepositMeta
}

func (custodianDeposit PortalCustodianDeposit) Hash() *common.Hash {
	record := custodianDeposit.MetadataBase.Hash().String()
	record += custodianDeposit.IncogAddressStr
	for tokenSymbol, rAddress := range custodianDeposit.RemoteAddresses {
		record += tokenSymbol
		record += rAddress
	}
	record += strconv.FormatUint(custodianDeposit.DepositedAmount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (custodianDeposit *PortalCustodianDeposit) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := PortalCustodianDepositAction{
		Meta:    *custodianDeposit,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PDEContributionMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (custodianDeposit *PortalCustodianDeposit) CalculateSize() uint64 {
	return calculateSize(custodianDeposit)
}
