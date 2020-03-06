package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/wallet"
	"math"
	"reflect"
	"strconv"
)

// PortalRedeemRequest - portal user redeem requests to get public token by burning ptoken
// metadata - redeem request - create normal tx with this metadata
type PortalRedeemRequest struct {
	MetadataBase
	UniqueRedeemID string
	TokenID        string // pTokenID in incognito chain
	RedeemAmount   uint64
	IncAddressStr  string
	RemoteAddress  string // btc/bnb/etc address
	RedeemFee      uint64 // ptoken fee
}

// PortalRedeemRequestAction - shard validator creates instruction that contain this action content
// it will be append to ShardToBeaconBlock
type PortalRedeemRequestAction struct {
	Meta    PortalRedeemRequest
	TxReqID common.Hash
	ShardID byte
}

// PortalRedeemRequestContent - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and rejected status
type PortalRedeemRequestContent struct {
	UniqueRedeemID          string
	TokenID                 string // pTokenID in incognito chain
	RedeemAmount            uint64
	IncAddressStr           string
	RemoteAddress           string // btc/bnb/etc address
	RedeemFee               uint64 // ptoken fee, 0.01% redeemAmount
	MatchingCustodianDetail map[string]*lvdb.MatchingRedeemCustodianDetail   // key: incAddressCustodian
	TxReqID                 common.Hash
	ShardID                 byte
}

// PortalRedeemRequestStatus - Beacon tracks status of redeem request into db
type PortalRedeemRequestStatus struct {
	Status                  byte
	UniqueRedeemID          string
	TokenID                 string // pTokenID in incognito chain
	RedeemAmount            uint64
	IncAddressStr           string
	RemoteAddress           string // btc/bnb/etc address
	RedeemFee               uint64 // ptoken fee
	MatchingCustodianDetail map[string]*lvdb.MatchingRedeemCustodianDetail   // key: incAddressCustodian
	TxReqID                 common.Hash
}

func NewPortalRedeemRequest(
	metaType int,
	uniqueRedeemID string,
	tokenID string,
	redeemAmount uint64,
	incAddressStr string,
	remoteAddr string,
	redeemFee uint64) (*PortalRedeemRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	requestPTokenMeta := &PortalRedeemRequest{
		UniqueRedeemID: uniqueRedeemID,
		TokenID:        tokenID,
		RedeemAmount:   redeemAmount,
		IncAddressStr:  incAddressStr,
		RemoteAddress:  remoteAddr,
		RedeemFee:      redeemFee,
	}
	requestPTokenMeta.MetadataBase = metadataBase
	return requestPTokenMeta, nil
}

func getMinRedeemFeeByRedeemAmount(redeemAmount uint64) (uint64, error) {
	if redeemAmount == 0 {
		return 0, errors.New("redeem amount must be greater than 0")
	}
	// minRedeemFee = redeemAmount * 0.01%
	minRedeemFee := uint64(math.Floor(float64(redeemAmount) * 0.01 / 100))
	return minRedeemFee, nil
}

func (redeemReq PortalRedeemRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (redeemReq PortalRedeemRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction, beaconHeight uint64) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
		return true, true, nil
	}

	// validate IncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(redeemReq.IncAddressStr)
	if err != nil {
		return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("Requester incognito address is invalid"))
	}
	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("Requester incognito address is invalid"))
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incAddr.Pk[:]) {
		return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("Requester incognito address is not signer"))
	}

	// check tx type
	if txr.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, errors.New("tx redeem request must be TxCustomTokenPrivacyType")
	}

	if !txr.IsCoinsBurning(bcr, beaconHeight) {
		return false, false, errors.New("tx redeem request must be coin burning tx")
	}

	// validate redeem amount
	if redeemReq.RedeemAmount == 0 {
		return false, false, errors.New("redeem amount should be larger than 0")
	}

	// validate redeem fee
	if redeemReq.RedeemFee == 0 {
		return false, false, errors.New("redeem fee should be larger than 0")
	}

	minFee, err := getMinRedeemFeeByRedeemAmount(redeemReq.RedeemAmount)
	if err != nil {
		return false, false, err
	}

	if redeemReq.RedeemFee < minFee {
		return false, false, fmt.Errorf("redeem fee should be larger than min fee %v\n", minFee)
	}

	// validate value transfer of tx
	if redeemReq.RedeemAmount+redeemReq.RedeemFee != txr.CalculateTxValue() {
		return false, false, errors.New("deposit amount should be equal to the tx value")
	}

	// validate tokenID
	if redeemReq.TokenID != txr.GetTokenID().String() {
		return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("TokenID in metadata is not matched to tokenID in tx"))
	}
	// check tokenId is portal token or not
	if !IsPortalToken(redeemReq.TokenID) {
		return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("TokenID is not in portal tokens list"))
	}

	//validate RemoteAddress
	// todo:
	if len(redeemReq.RemoteAddress) == 0 {
		return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("Remote address is invalid"))
	}

	return true, true, nil
}

func (redeemReq PortalRedeemRequest) ValidateMetadataByItself() bool {
	return redeemReq.Type == PortalRedeemRequestMeta
}

func (redeemReq PortalRedeemRequest) Hash() *common.Hash {
	record := redeemReq.MetadataBase.Hash().String()
	record += redeemReq.UniqueRedeemID
	record += redeemReq.TokenID
	record += strconv.FormatUint(redeemReq.RedeemAmount, 10)
	record += strconv.FormatUint(redeemReq.RedeemFee, 10)
	record += redeemReq.IncAddressStr
	record += redeemReq.RemoteAddress
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (redeemReq *PortalRedeemRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := PortalRedeemRequestAction{
		Meta:    *redeemReq,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalRedeemRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (redeemReq *PortalRedeemRequest) CalculateSize() uint64 {
	return calculateSize(redeemReq)
}
