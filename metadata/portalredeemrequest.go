package metadata

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

// PortalRedeemRequest - portal user redeem requests to get public token by burning ptoken
// metadata - redeem request - create normal tx with this metadata
type PortalRedeemRequest struct {
	MetadataBase
	UniqueRedeemID        string
	TokenID               string // pTokenID in incognito chain
	RedeemAmount          uint64
	RedeemerIncAddressStr string
	RemoteAddress         string // btc/bnb/etc address
	RedeemFee             uint64 // redeem fee in PRV, 0.01% redeemAmount in PRV
}

// PortalRedeemRequestV3 - portal user redeem requests to get public token by burning ptoken
// metadata - redeem request - create normal tx with this metadata
type PortalRedeemRequestV3 struct {
	MetadataBase
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
	Meta    PortalRedeemRequest
	TxReqID common.Hash
	ShardID byte
}

// PortalRedeemRequestAction - shard validator creates instruction that contain this action content
type PortalRedeemRequestActionV3 struct {
	Meta        PortalRedeemRequestV3
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
) (*PortalRedeemRequestV3, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	requestPTokenMeta := &PortalRedeemRequestV3{
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

func (redeemReq PortalRedeemRequestV3) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (redeemReq PortalRedeemRequestV3) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	// check tx type
	if txr.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, fmt.Errorf("tx redeem request must be of type `tp`")
	}

	// check burned coins
	isBurned, burnedPRVCoin, burnedTokenCoin, burnedTokenID, err := txr.GetTxFullBurnData()
	if err != nil || !isBurned {
		return false, false, fmt.Errorf("this is not tx burn")
	}

	// check redeem fee
	if redeemReq.RedeemFee == 0 {
		return false, false, fmt.Errorf("redeem fee should be greater than 0")
	}
	if burnedPRVCoin == nil {
		return false, false, fmt.Errorf("must have a burned PRV output")
	}
	if redeemReq.RedeemFee != burnedPRVCoin.GetValue() {
		return false, false, fmt.Errorf("expect redeemFee to be %v, got %v", redeemReq.RedeemFee, burnedPRVCoin.GetValue())
	}

	// validate RedeemerIncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(redeemReq.RedeemerIncAddressStr)
	if err != nil {
		return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("Requester incognito address is invalid"))
	}

	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("Requester incognito address is invalid"))
	}

	// validate redeem amount
	minAmount, err := chainRetriever.GetMinAmountPortalToken(redeemReq.TokenID, beaconHeight, common.PortalVersion3)
	if err != nil {
		return false, false, fmt.Errorf("Error get min portal token amount: %v", err)
	}
	if redeemReq.RedeemAmount < minAmount {
		return false, false, fmt.Errorf("redeem amount should be larger or equal to %v", minAmount)
	}
	// validate value transfer of tx for redeem amount in ptoken
	if redeemReq.RedeemAmount != burnedTokenCoin.GetValue() {
		return false, false, fmt.Errorf("redeem amount should be equal to the tx value")
	}

	// validate tokenID
	if redeemReq.TokenID != burnedTokenID.String() {
		return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("TokenID in metadata is not matched to tokenID in tx"))
	}
	// check tokenId is portal token or not
	isPortalToken, err := chainRetriever.IsPortalToken(beaconHeight, redeemReq.TokenID, common.PortalVersion3)
	if !isPortalToken || err != nil {
		return false, false, fmt.Errorf("TokenID is not in portal tokens list")
	}

	//validate RemoteAddress
	isValidRemoteAddress, err := chainRetriever.IsValidPortalRemoteAddress(redeemReq.TokenID, redeemReq.RemoteAddress, beaconHeight, common.PortalVersion3)
	if err != nil || !isValidRemoteAddress {
		return false, false, fmt.Errorf("Remote address %v is not a valid address of tokenID %v - Error %v", redeemReq.RemoteAddress, redeemReq.TokenID, err)
	}

	if beaconHeight >= config.Param().BCHeightBreakPointPortalV3 {
		// validate metadata type
		if redeemReq.Type != PortalRedeemRequestMetaV3 {
			return false, false, fmt.Errorf("Metadata type should be %v", PortalRedeemRequestMetaV3)
		}

		//validate RedeemerExternalAddress
		if len(redeemReq.RedeemerExternalAddress) == 0 {
			return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("Redeemer address for liquidating is invalid"))
		}
		if isValid, err := ValidatePortalExternalAddress(common.ETHChainName, "", redeemReq.RedeemerExternalAddress); !isValid || err != nil {
			return false, false, fmt.Errorf("RedeemerExternalAddress %v is not a valid address of ethereum network", redeemReq.RedeemerExternalAddress)
		}
	}

	return true, true, nil
}

func (redeemReq PortalRedeemRequestV3) ValidateMetadataByItself() bool {
	return redeemReq.Type == PortalRedeemRequestMeta || redeemReq.Type == PortalRedeemRequestMetaV3
}

func (redeemReq PortalRedeemRequestV3) Hash() *common.Hash {
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

func (redeemReq *PortalRedeemRequestV3) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	if redeemReq.Type == PortalRedeemRequestMeta {
		actionContent := PortalRedeemRequestAction{
			Meta: PortalRedeemRequest{
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
	} else if redeemReq.Type == PortalRedeemRequestMetaV3 {
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

func (redeemReq *PortalRedeemRequestV3) CalculateSize() uint64 {
	return calculateSize(redeemReq)
}

func (redeemReq *PortalRedeemRequestV3) ToCompactBytes() ([]byte, error) {
	return toCompactBytes(redeemReq)
}
