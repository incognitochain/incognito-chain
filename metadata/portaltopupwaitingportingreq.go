package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PortalTopUpWaitingPortingRequest struct {
	MetadataBase
	IncogAddressStr      string
	PortingID            string
	PTokenID             string
	DepositedAmount      uint64
	FreeCollateralAmount uint64
}

type PortalTopUpWaitingPortingRequestAction struct {
	Meta    PortalTopUpWaitingPortingRequest
	TxReqID common.Hash
	ShardID byte
}

type PortalTopUpWaitingPortingRequestContent struct {
	IncogAddressStr      string
	PortingID            string
	PTokenID             string
	DepositedAmount      uint64
	FreeCollateralAmount uint64
	TxReqID              common.Hash
	ShardID              byte
}

type PortalTopUpWaitingPortingRequestStatus struct {
	TxReqID              common.Hash
	IncogAddressStr      string
	PortingID            string
	PTokenID             string
	DepositAmount        uint64
	FreeCollateralAmount uint64
	Status               byte
}

func NewPortalTopUpWaitingPortingRequestStatus(
	txReqID common.Hash,
	portingID string,
	incogAddressStr string,
	pTokenID string,
	depositAmount uint64,
	freeCollateralAmount uint64,
	status byte,
) *PortalTopUpWaitingPortingRequestStatus {
	return &PortalTopUpWaitingPortingRequestStatus{
		TxReqID:              txReqID,
		PortingID:            portingID,
		IncogAddressStr:      incogAddressStr,
		PTokenID:             pTokenID,
		DepositAmount:        depositAmount,
		FreeCollateralAmount: freeCollateralAmount,
		Status:               status,
	}
}

func NewPortalTopUpWaitingPortingRequest(
	metaType int,
	portingID string,
	incogAddressStr string,
	pToken string,
	amount uint64,
	freeCollateralAmount uint64,
) (*PortalTopUpWaitingPortingRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	portalTopUpWaitingPortingRequestMeta := &PortalTopUpWaitingPortingRequest{
		PortingID:            portingID,
		IncogAddressStr:      incogAddressStr,
		PTokenID:             pToken,
		DepositedAmount:      amount,
		FreeCollateralAmount: freeCollateralAmount,
	}
	portalTopUpWaitingPortingRequestMeta.MetadataBase = metadataBase
	return portalTopUpWaitingPortingRequestMeta, nil
}

func (p PortalTopUpWaitingPortingRequest) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (p PortalTopUpWaitingPortingRequest) ValidateSanityData(
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	beaconHeight uint64,
	txr Transaction,
) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
		return true, true, nil
	}

	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(p.IncogAddressStr)
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
	// check tx type
	if txr.GetType() != common.TxNormalType || !bytes.Equal(burnedTokenID.Bytes(), common.PRVCoinID[:]) {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}
	// validate amount deposit
	if p.DepositedAmount == 0 || p.DepositedAmount != burnCoin.GetValue(){
		return false, false, errors.New("deposit amount should be larger than 0 and equal burn value")
	}

	if !common.IsPortalToken(p.PTokenID) {
		return false, false, errors.New("TokenID in remote address is invalid")
	}

	if p.PortingID == "" {
		return false, false, errors.New("Porting ID should not be empty")
	}

	return true, true, nil
}

func (p PortalTopUpWaitingPortingRequest) ValidateMetadataByItself() bool {
	return p.Type == PortalTopUpWaitingPortingRequestMeta
}

func (p PortalTopUpWaitingPortingRequest) Hash() *common.Hash {
	record := p.MetadataBase.Hash().String()
	record += p.PortingID
	record += p.IncogAddressStr
	record += p.PTokenID
	record += strconv.FormatUint(p.DepositedAmount, 10)
	record += strconv.FormatUint(p.FreeCollateralAmount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (p *PortalTopUpWaitingPortingRequest) BuildReqActions(
	tx Transaction,
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	shardID byte,
) ([][]string, error) {
	actionContent := PortalTopUpWaitingPortingRequestAction{
		Meta:    *p,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalTopUpWaitingPortingRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (p *PortalTopUpWaitingPortingRequest) CalculateSize() uint64 {
	return calculateSize(p)
}
