package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/wallet"
	"reflect"
	"strconv"
)

type PortalLiquidationCustodianDeposit struct {
	basemeta.MetadataBase
	IncogAddressStr        string
	PTokenId               string
	DepositedAmount        uint64
	FreeCollateralSelected bool
}

type PortalLiquidationCustodianDepositAction struct {
	Meta    PortalLiquidationCustodianDeposit
	TxReqID common.Hash
	ShardID byte
}

type PortalLiquidationCustodianDepositContent struct {
	IncogAddressStr        string
	PTokenId               string
	DepositedAmount        uint64
	FreeCollateralSelected bool
	TxReqID                common.Hash
	ShardID                byte
}

type LiquidationCustodianDepositStatus struct {
	TxReqID                common.Hash
	IncogAddressStr        string
	PTokenId               string
	DepositAmount          uint64
	FreeCollateralSelected bool
	Status                 byte
}

func NewLiquidationCustodianDepositStatus(txReqID common.Hash, incogAddressStr string, PTokenId string, depositAmount uint64, freeCollateralSelected bool, status byte) *LiquidationCustodianDepositStatus {
	return &LiquidationCustodianDepositStatus{TxReqID: txReqID, IncogAddressStr: incogAddressStr, PTokenId: PTokenId, DepositAmount: depositAmount, FreeCollateralSelected: freeCollateralSelected, Status: status}
}

func NewPortalLiquidationCustodianDeposit(metaType int, incognitoAddrStr string, pToken string, amount uint64, freeCollateralSelected bool) (*PortalLiquidationCustodianDeposit, error) {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	custodianDepositMeta := &PortalLiquidationCustodianDeposit{
		IncogAddressStr:        incognitoAddrStr,
		PTokenId:               pToken,
		DepositedAmount:        amount,
		FreeCollateralSelected: freeCollateralSelected,
	}
	custodianDepositMeta.MetadataBase = metadataBase
	return custodianDepositMeta, nil
}

func (custodianDeposit PortalLiquidationCustodianDeposit) ValidateTxWithBlockChain(
	txr basemeta.Transaction,
	chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (custodianDeposit PortalLiquidationCustodianDeposit) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, txr basemeta.Transaction) (bool, bool, error) {
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
	if !txr.IsCoinsBurning(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight) {
		return false, false, errors.New("must send coin to burning address")
	}

	// validate amount deposit
	if custodianDeposit.DepositedAmount != txr.CalculateTxValue() {
		return false, false, errors.New("deposit amount should be equal to the tx value")
	}

	if !chainRetriever.IsPortalToken(beaconHeight, custodianDeposit.PTokenId) {
		return false, false, errors.New("TokenID in remote address is invalid")
	}

	return true, true, nil
}

func (custodianDeposit PortalLiquidationCustodianDeposit) ValidateMetadataByItself() bool {
	return custodianDeposit.Type == basemeta.PortalCustodianTopupMeta
}

func (custodianDeposit PortalLiquidationCustodianDeposit) Hash() *common.Hash {
	record := custodianDeposit.MetadataBase.Hash().String()
	record += custodianDeposit.IncogAddressStr
	record += custodianDeposit.PTokenId
	record += strconv.FormatUint(custodianDeposit.DepositedAmount, 10)
	record += strconv.FormatBool(custodianDeposit.FreeCollateralSelected)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (custodianDeposit *PortalLiquidationCustodianDeposit) BuildReqActions(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
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
	action := []string{strconv.Itoa(basemeta.PortalCustodianTopupMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (custodianDeposit *PortalLiquidationCustodianDeposit) CalculateSize() uint64 {
	return basemeta.CalculateSize(custodianDeposit)
}
