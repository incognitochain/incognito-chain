package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/basemeta"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

// PortalRedeemRequestV2 - portal user redeem requests to get public token by burning ptoken
// metadata - redeem request - create normal tx with this metadata
type PortalRedeemRequestV2 struct {
	basemeta.MetadataBase
	UniqueRedeemID        string
	TokenID               string // pTokenID in incognito chain
	RedeemAmount          uint64
	RedeemerIncAddressStr string
	RemoteAddress         string // btc/bnb/etc address
	RedeemFee             uint64 // redeem fee in PRV, 0.01% redeemAmount in PRV
}

// PortalRedeemRequest - portal user redeem requests to get public token by burning ptoken
// metadata - redeem request - create normal tx with this metadata
type PortalRedeemRequest struct {
	basemeta.MetadataBase
	UniqueRedeemID          string
	TokenID                 string // pTokenID in incognito chain
	RedeemAmount            uint64
	RedeemerIncAddressStr   string
	RemoteAddress           string // btc/bnb/etc address
	RedeemFee               uint64 // redeem fee in PRV, 0.01% redeemAmount in PRV
	RedeemerExternalAddress string
}

// PortalRedeemRequestAction - shard validator creates instruction that contain this action content
type PortalRedeemRequestAction struct {
	Meta    PortalRedeemRequestV2
	TxReqID common.Hash
	ShardID byte
}

// PortalRedeemRequestAction - shard validator creates instruction that contain this action content
type PortalRedeemRequestActionV3 struct {
	Meta        PortalRedeemRequest
	TxReqID     common.Hash
	ShardID     byte
	ShardHeight uint64
}

// PortalRedeemRequestContent - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and rejected status
type PortalRedeemRequestContent struct {
	UniqueRedeemID          string
	TokenID                 string // pTokenID in incognito chain
	RedeemAmount            uint64
	RedeemerIncAddressStr   string
	RemoteAddress           string                                   // btc/bnb/etc address
	RedeemFee               uint64                                   // redeem fee in PRV, 0.01% redeemAmount in PRV
	MatchingCustodianDetail []*statedb.MatchingRedeemCustodianDetail // key: incAddressCustodian
	TxReqID                 common.Hash
	ShardID                 byte
	ShardHeight             uint64
	RedeemerExternalAddress string
}

// PortalRedeemRequestStatus - Beacon tracks status of redeem request into db
type PortalRedeemRequestStatus struct {
	Status                  byte
	UniqueRedeemID          string
	TokenID                 string // pTokenID in incognito chain
	RedeemAmount            uint64
	RedeemerIncAddressStr   string
	RemoteAddress           string                                   // btc/bnb/etc address
	RedeemFee               uint64                                   // redeem fee in PRV, 0.01% redeemAmount in PRV
	MatchingCustodianDetail []*statedb.MatchingRedeemCustodianDetail // key: incAddressCustodian
	TxReqID                 common.Hash
	ShardID                 byte
	ShardHeight             uint64
	BeaconHeight            uint64
	RedeemerExternalAddress string
}

func NewPortalRedeemRequest(
	metaType int,
	uniqueRedeemID string,
	tokenID string,
	redeemAmount uint64,
	incAddressStr string,
	remoteAddr string,
	redeemFee uint64,
	redeemerExternalAddress string,
) (*PortalRedeemRequest, error) {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	requestPTokenMeta := &PortalRedeemRequest{
		UniqueRedeemID:          uniqueRedeemID,
		TokenID:                 tokenID,
		RedeemAmount:            redeemAmount,
		RedeemerIncAddressStr:   incAddressStr,
		RemoteAddress:           remoteAddr,
		RedeemFee:               redeemFee,
		RedeemerExternalAddress: redeemerExternalAddress,
	}
	requestPTokenMeta.MetadataBase = metadataBase
	return requestPTokenMeta, nil
}

func (redeemReq PortalRedeemRequest) ValidateTxWithBlockChain(
	txr basemeta.Transaction,
	chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (redeemReq PortalRedeemRequest) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, txr basemeta.Transaction) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
		if !txr.IsCoinsBurning(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight) {
			return false, false, errors.New("txnormal in tx redeem request must be coin burning tx")
		}
		// validate value transfer of tx for redeem fee in prv
		if redeemReq.RedeemFee != txr.CalculateTxValue() {
			return false, false, errors.New("redeem fee amount should be equal to the tx value")
		}
		return true, true, nil
	}

	// validate RedeemerIncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(redeemReq.RedeemerIncAddressStr)
	if err != nil {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRedeemRequestParamError, errors.New("Requester incognito address is invalid"))
	}
	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRedeemRequestParamError, errors.New("Requester incognito address is invalid"))
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incAddr.Pk[:]) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRedeemRequestParamError, errors.New("Requester incognito address is not signer"))
	}

	// check tx type
	if txr.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, errors.New("tx redeem request must be TxCustomTokenPrivacyType")
	}

	if !txr.IsCoinsBurning(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight) {
		return false, false, errors.New("txprivacytoken in tx redeem request must be coin burning tx")
	}

	// validate redeem amount
	minAmount := common.MinAmountPortalPToken[redeemReq.TokenID]
	if redeemReq.RedeemAmount < minAmount {
		return false, false, fmt.Errorf("redeem amount should be larger or equal to %v", minAmount)
	}

	// validate redeem fee
	if redeemReq.RedeemFee <= 0 {
		return false, false, errors.New("redeem fee should be larger than 0")
	}

	// validate value transfer of tx for redeem amount in ptoken
	if redeemReq.RedeemAmount != txr.CalculateTxValue() {
		return false, false, errors.New("redeem amount should be equal to the tx value")
	}

	// validate tokenID
	if redeemReq.TokenID != txr.GetTokenID().String() {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRedeemRequestParamError, errors.New("TokenID in metadata is not matched to tokenID in tx"))
	}
	// check tokenId is portal token or not
	if !!chainRetriever.IsPortalToken(beaconHeight, redeemReq.TokenID) {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRedeemRequestParamError, errors.New("TokenID is not in portal tokens list"))
	}

	//validate RemoteAddress
	if len(redeemReq.RemoteAddress) == 0 {
		return false, false, basemeta.NewMetadataTxError(basemeta.PortalRedeemRequestParamError, errors.New("Remote address is invalid"))
	}
	if !IsValidPortalRemoteAddress(chainRetriever, redeemReq.RemoteAddress, redeemReq.TokenID, beaconHeight) {
		return false, false, fmt.Errorf("Remote address %v is not a valid address of tokenID %v", redeemReq.RemoteAddress, redeemReq.TokenID)
	}

	if beaconHeight >= chainRetriever.GetBCHeightBreakPointPortalV3() {
		// validate metadata type
		if redeemReq.Type != basemeta.PortalRedeemRequestMetaV3 {
			return false, false, fmt.Errorf("Metadata type should be %v", basemeta.PortalRedeemRequestMetaV3)
		}

		//validate RedeemerExternalAddress
		if len(redeemReq.RedeemerExternalAddress) == 0 {
			return false, false, basemeta.NewMetadataTxError(basemeta.PortalRedeemRequestParamError, errors.New("Redeemer address for liquidating is invalid"))
		}
		if isValid, err := ValidatePortalExternalAddress(common.ETHChainName, "", redeemReq.RedeemerExternalAddress); !isValid || err != nil {
			return false, false, fmt.Errorf("RedeemerExternalAddress %v is not a valid address of ethereum network", redeemReq.RedeemerExternalAddress)
		}
	}

	return true, true, nil
}

func (redeemReq PortalRedeemRequest) ValidateMetadataByItself() bool {
	return redeemReq.Type == basemeta.PortalRedeemRequestMeta || redeemReq.Type == basemeta.PortalRedeemRequestMetaV3
}

func (redeemReq PortalRedeemRequest) Hash() *common.Hash {
	record := redeemReq.MetadataBase.Hash().String()
	record += redeemReq.UniqueRedeemID
	record += redeemReq.TokenID
	record += strconv.FormatUint(redeemReq.RedeemAmount, 10)
	record += strconv.FormatUint(redeemReq.RedeemFee, 10)
	record += redeemReq.RedeemerIncAddressStr
	record += redeemReq.RemoteAddress
	record += redeemReq.RedeemerExternalAddress
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (redeemReq *PortalRedeemRequest) BuildReqActions(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	if redeemReq.Type == basemeta.PortalRedeemRequestMeta {
		actionContent := PortalRedeemRequestAction{
			Meta: PortalRedeemRequestV2{
				MetadataBase:          redeemReq.MetadataBase,
				UniqueRedeemID:        redeemReq.UniqueRedeemID,
				TokenID:               redeemReq.TokenID,
				RedeemAmount:          redeemReq.RedeemAmount,
				RedeemerIncAddressStr: redeemReq.RedeemerIncAddressStr,
				RemoteAddress:         redeemReq.RemoteAddress,
				RedeemFee:             redeemReq.RedeemFee,
			},
			TxReqID: *tx.Hash(),
			ShardID: shardID,
		}
		actionContentBytes, err := json.Marshal(actionContent)
		if err != nil {
			return [][]string{}, err
		}
		actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
		action := []string{strconv.Itoa(redeemReq.Type), actionContentBase64Str}
		return [][]string{action}, nil
	} else if redeemReq.Type == basemeta.PortalRedeemRequestMetaV3 {
		actionContent := PortalRedeemRequestActionV3{
			Meta:        *redeemReq,
			TxReqID:     *tx.Hash(),
			ShardID:     shardID,
			ShardHeight: shardHeight,
		}
		actionContentBytes, err := json.Marshal(actionContent)
		if err != nil {
			return [][]string{}, err
		}
		actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
		action := []string{strconv.Itoa(redeemReq.Type), actionContentBase64Str}
		return [][]string{action}, nil
	}

	return nil, nil
}

func (redeemReq *PortalRedeemRequest) CalculateSize() uint64 {
	return basemeta.CalculateSize(redeemReq)
}
