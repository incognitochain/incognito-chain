package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	eCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

// PortalLiquidationCustodianDepositV3 - custodian topup more token collaterals (ETH, ERC20) through the smart contract's bond
// submit the deposit proof
type PortalTopUpWaitingPortingRequestV3 struct {
	MetadataBase
	IncogAddressStr           string
	PortalTokenID             string
	CollateralTokenID         string
	DepositAmount             uint64
	FreeTokenCollateralAmount uint64 // topup from free token collaterals
	PortingID                 string

	// ETH proof
	BlockHash eCommon.Hash
	TxIndex   uint
	ProofStrs []string
}

type PortalTopUpWaitingPortingRequestActionV3 struct {
	Meta    PortalTopUpWaitingPortingRequestV3
	TxReqID common.Hash
	ShardID byte
}

type PortalTopUpWaitingPortingRequestContentV3 struct {
	IncogAddressStr           string
	PortalTokenID             string
	CollateralTokenID         string
	DepositAmount             uint64
	FreeTokenCollateralAmount uint64 // topup from free token collaterals
	PortingID                 string
	UniqExternalTxID          []byte
	TxReqID                   common.Hash
	ShardID                   byte
}

type PortalTopUpWaitingPortingRequestStatusV3 struct {
	IncogAddressStr           string
	PortalTokenID             string
	CollateralTokenID         string
	DepositAmount             uint64
	FreeTokenCollateralAmount uint64 // topup from free token collaterals
	PortingID                 string
	UniqExternalTxID          []byte
	TxReqID                   common.Hash
	Status                    byte
}

func NewPortalTopUpWaitingPortingRequestStatusV3(
	incognitoAddrStr string,
	portalTokenID string,
	collateralTokenID string,
	depositAmount uint64,
	freeTokenCollateralAmount uint64,
	portingId string,
	uniqExternalTokenID []byte,
	txReqID common.Hash,
	status byte) *PortalTopUpWaitingPortingRequestStatusV3 {
	return &PortalTopUpWaitingPortingRequestStatusV3{
		IncogAddressStr:           incognitoAddrStr,
		PortalTokenID:             portalTokenID,
		CollateralTokenID:         collateralTokenID,
		DepositAmount:             depositAmount,
		FreeTokenCollateralAmount: freeTokenCollateralAmount,
		PortingID:                 portingId,
		UniqExternalTxID:          uniqExternalTokenID,
		TxReqID:                   txReqID,
		Status:                    status,
	}
}

func NewPortalTopUpWaitingPortingRequestV3(
	metaType int,
	incognitoAddrStr string,
	portalTokenID string,
	collateralTokenID string,
	depositAmount uint64,
	freeTokenCollateralAmount uint64,
	portingID string,
	blockHash eCommon.Hash,
	txIndex uint,
	proofStrs []string,
) (*PortalTopUpWaitingPortingRequestV3, error) {
	custodianDepositMeta := &PortalTopUpWaitingPortingRequestV3{
		MetadataBase: MetadataBase{
			Type: metaType,
		},
		IncogAddressStr:           incognitoAddrStr,
		PortalTokenID:             portalTokenID,
		CollateralTokenID:         collateralTokenID,
		DepositAmount:             depositAmount,
		FreeTokenCollateralAmount: freeTokenCollateralAmount,
		PortingID:                 portingID,
		BlockHash:                 blockHash,
		TxIndex:                   txIndex,
		ProofStrs:                 proofStrs,
	}
	return custodianDepositMeta, nil
}

func (req PortalTopUpWaitingPortingRequestV3) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (req PortalTopUpWaitingPortingRequestV3) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(req.IncogAddressStr)
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

	// check PortalTokenID
	if !IsPortalToken(req.PortalTokenID) {
		return false, false, errors.New("TokenID in remote address is invalid")
	}

	// check CollateralTokenID
	if common.Has0xPrefix(req.CollateralTokenID) {
		return false, false, errors.New("CollateralTokenID shouldn't have 0x prefix")
	}
	if !IsSupportedTokenCollateralV3(chainRetriever, beaconHeight, req.CollateralTokenID) {
		return false, false, errors.New("CollateralTokenID is not portal collateral")
	}

	// validate amount deposit
	if req.DepositAmount > 0 {
		// validate deposit proof
		if len(req.BlockHash.Bytes()) == 0 {
			return false, false, errors.New("BlockHash should be not empty")
		}
		if len(req.ProofStrs) == 0 {
			return false, false, errors.New("ProofStrs should be not empty")
		}
	} else if req.FreeTokenCollateralAmount == 0 {
		return false, false, errors.New("both DepositedAmount and FreeCollateralAmount are zero")
	}

	// check portingID != nil
	if req.PortingID == "" {
		return false, false, errors.New("PortingID should not be empty")
	}

	return true, true, nil
}

func (req PortalTopUpWaitingPortingRequestV3) ValidateMetadataByItself() bool {
	return req.Type == PortalTopUpWaitingPortingRequestMetaV3
}

func (req PortalTopUpWaitingPortingRequestV3) Hash() *common.Hash {
	record := req.MetadataBase.Hash().String()
	record += req.IncogAddressStr
	record += req.PortalTokenID
	record += req.CollateralTokenID
	record += strconv.FormatUint(req.DepositAmount, 10)
	record += strconv.FormatUint(req.FreeTokenCollateralAmount, 10)
	record += req.PortingID

	record += req.BlockHash.String()
	record += strconv.FormatUint(uint64(req.TxIndex), 10)
	for _, proofStr := range req.ProofStrs {
		record += proofStr
	}
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (req *PortalTopUpWaitingPortingRequestV3) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalTopUpWaitingPortingRequestActionV3{
		Meta:    *req,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalTopUpWaitingPortingRequestMetaV3), actionContentBase64Str}
	return [][]string{action}, nil
}

func (req *PortalTopUpWaitingPortingRequestV3) CalculateSize() uint64 {
	return calculateSize(req)
}
