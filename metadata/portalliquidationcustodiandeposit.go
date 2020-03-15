package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/wallet"
	"reflect"
	"strconv"
)

type PortalLiquidationCustodianDeposit struct {
	MetadataBase
	IncogAddressStr string
	PTokenId string
	RemoteAddress string
	DepositedAmount uint64
}

type PortalLiquidationCustodianDepositAction struct {
	Meta    PortalLiquidationCustodianDeposit
	TxReqID common.Hash
	ShardID byte
}

type PortalLiquidationCustodianDepositContent struct {
	IncogAddressStr string
	PTokenId string
	DepositedAmount uint64
	RemoteAddress string
	TxReqID         common.Hash
	ShardID         byte
}

func NewPortalLiquidationCustodianDeposit(metaType int, incognitoAddrStr string, pToken string, remoteAddress string, amount uint64) (*PortalLiquidationCustodianDeposit , error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	custodianDepositMeta := &PortalLiquidationCustodianDeposit {
		IncogAddressStr: incognitoAddrStr,
		PTokenId: pToken,
		DepositedAmount: amount, 
		RemoteAddress: remoteAddress,
	}
	custodianDepositMeta.MetadataBase = metadataBase
	return custodianDepositMeta, nil
}

func (custodianDeposit PortalLiquidationCustodianDeposit) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (custodianDeposit PortalLiquidationCustodianDeposit) ValidateSanityData(bcr BlockchainRetriever, txr Transaction, beaconHeight uint64) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
		return true, true, nil
	}

	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(custodianDeposit.IncogAddressStr)
	if err != nil {
		return false, false, errors.New("IncogAddressStr of custodian incorrect")
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
	if !txr.IsCoinsBurning(bcr, beaconHeight) {
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
	if len(custodianDeposit.RemoteAddress) == 0 {
		return false, false, errors.New("remote addresses should be at least one")
	}

	if !common.IsPortalToken(custodianDeposit.PTokenId) {
		return false, false, errors.New("TokenID in remote address is invalid")
	}

	return true, true, nil
}

func (custodianDeposit PortalLiquidationCustodianDeposit) ValidateMetadataByItself() bool {
	return custodianDeposit.Type == PortalLiquidationCustodianDepositMeta
}

func (custodianDeposit PortalLiquidationCustodianDeposit) Hash() *common.Hash {
	record := custodianDeposit.MetadataBase.Hash().String()
	record += custodianDeposit.IncogAddressStr
	record += custodianDeposit.RemoteAddress
	record += custodianDeposit.PTokenId
	record += strconv.FormatUint(custodianDeposit.DepositedAmount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (custodianDeposit *PortalLiquidationCustodianDeposit) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := PortalLiquidationCustodianDepositAction{
		Meta:    *custodianDeposit,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalLiquidationCustodianDepositMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (custodianDeposit *PortalLiquidationCustodianDeposit) CalculateSize() uint64 {
	return calculateSize(custodianDeposit)
}
