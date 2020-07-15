package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
	"reflect"
	"strconv"
)

type PortalLiquidationCustodianDepositV2 struct {
	MetadataBase
	IncogAddressStr      string
	PTokenId             string
	DepositedAmount      uint64
	FreeCollateralAmount uint64
}

type PortalLiquidationCustodianDepositActionV2 struct {
	Meta    PortalLiquidationCustodianDepositV2
	TxReqID common.Hash
	ShardID byte
}

type PortalLiquidationCustodianDepositContentV2 struct {
	IncogAddressStr      string
	PTokenId             string
	DepositedAmount      uint64
	FreeCollateralAmount uint64
	TxReqID              common.Hash
	ShardID              byte
}

type LiquidationCustodianDepositStatusV2 struct {
	TxReqID              common.Hash
	IncogAddressStr      string
	PTokenId             string
	DepositAmount        uint64
	FreeCollateralAmount uint64
	Status               byte
}

func NewLiquidationCustodianDepositStatusV2(txReqID common.Hash, incogAddressStr string, PTokenId string, depositAmount uint64, freeCollateralAmount uint64, status byte) *LiquidationCustodianDepositStatusV2 {
	return &LiquidationCustodianDepositStatusV2{TxReqID: txReqID, IncogAddressStr: incogAddressStr, PTokenId: PTokenId, DepositAmount: depositAmount, FreeCollateralAmount: freeCollateralAmount, Status: status}
}

func NewPortalLiquidationCustodianDepositV2(metaType int, incognitoAddrStr string, pToken string, amount uint64, freeCollateralAmount uint64) (*PortalLiquidationCustodianDepositV2, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	custodianDepositMeta := &PortalLiquidationCustodianDepositV2{
		IncogAddressStr:      incognitoAddrStr,
		PTokenId:             pToken,
		DepositedAmount:      amount,
		FreeCollateralAmount: freeCollateralAmount,
	}
	custodianDepositMeta.MetadataBase = metadataBase
	return custodianDepositMeta, nil
}

func (custodianDeposit PortalLiquidationCustodianDepositV2) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (custodianDeposit PortalLiquidationCustodianDepositV2) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
		return true, true, nil
	}

	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(custodianDeposit.IncogAddressStr)
	if err != nil {
		return false, false, errors.New("IncogAddressStr of custodian incorrect")
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("wrong custodian incognito address")
	}
	// check burning tx
	isBurned, burnCoin, burnedTokenID, err := txr.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, errors.New("Error This is not Tx Burn")
	}
	// validate amount deposit
	if custodianDeposit.DepositedAmount == 0 || custodianDeposit.DepositedAmount != burnCoin.GetValue(){
		return false, false, errors.New("deposit amount should be larger than 0 and equal burn value")
	}
	// check tx type and token burn
	if txr.GetType() != common.TxNormalType || !bytes.Equal(burnedTokenID.Bytes(), common.PRVCoinID[:]) {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}
	// check tx type
	if txr.GetType() != common.TxNormalType || !bytes.Equal(burnedTokenID.Bytes(), common.PRVCoinID[:]) {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}

	if !common.IsPortalToken(custodianDeposit.PTokenId) {
		return false, false, errors.New("TokenID in remote address is invalid")
	}

	return true, true, nil
}

func (custodianDeposit PortalLiquidationCustodianDepositV2) ValidateMetadataByItself() bool {
	return custodianDeposit.Type == PortalLiquidationCustodianDepositMetaV2
}

func (custodianDeposit PortalLiquidationCustodianDepositV2) Hash() *common.Hash {
	record := custodianDeposit.MetadataBase.Hash().String()
	record += custodianDeposit.IncogAddressStr
	record += custodianDeposit.PTokenId
	record += strconv.FormatUint(custodianDeposit.DepositedAmount, 10)
	record += strconv.FormatUint(custodianDeposit.FreeCollateralAmount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (custodianDeposit *PortalLiquidationCustodianDepositV2) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte) ([][]string, error) {
	actionContent := PortalLiquidationCustodianDepositActionV2{
		Meta:    *custodianDeposit,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalLiquidationCustodianDepositMetaV2), actionContentBase64Str}
	return [][]string{action}, nil
}

func (custodianDeposit *PortalLiquidationCustodianDepositV2) CalculateSize() uint64 {
	return calculateSize(custodianDeposit)
}
