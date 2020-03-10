package metadata

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"strconv"
)

// PortalRedeemRequest - portal user redeem requests to get public token by burning ptoken
// metadata - redeem request - create normal tx with this metadata
type PortalLiquidateCustodian struct {
	MetadataBase
	UniqueRedeemID         string
	TokenID                string // pTokenID in incognito chain
	RedeemPubTokenAmount   uint64
	RedeemerIncAddressStr  string
	CustodianIncAddressStr string
}

// PortalRedeemRequestAction - shard validator creates instruction that contain this action content
// it will be append to ShardToBeaconBlock
//type PortalRedeemRequestAction struct {
//	Meta    PortalRedeemRequest
//	TxReqID common.Hash
//	ShardID byte
//}

// PortalLiquidateCustodianContent - Beacon builds a new instruction with this content after detecting custodians run away
// It will be appended to beaconBlock
type PortalLiquidateCustodianContent struct {
	MetadataBase
	UniqueRedeemID         string
	TokenID                string // pTokenID in incognito chain
	RedeemPubTokenAmount   uint64
	RedeemerIncAddressStr  string
	CustodianIncAddressStr string
	ShardID                byte
}

//// PortalRedeemRequestStatus - Beacon tracks status of redeem request into db
//type PortalRedeemRequestStatus struct {
//	Status                  byte
//	UniqueRedeemID          string
//	TokenID                 string // pTokenID in incognito chain
//	RedeemAmount            uint64
//	RedeemerIncAddressStr   string
//	RemoteAddress           string // btc/bnb/etc address
//	RedeemFee               uint64 // ptoken fee
//	MatchingCustodianDetail map[string]*lvdb.MatchingRedeemCustodianDetail   // key: incAddressCustodian
//	TxReqID                 common.Hash
//}

func NewPortalLiquidateCustodian(
	metaType int,
	uniqueRedeemID string,
	tokenID string,
	redeemAmount uint64,
	redeemerIncAddressStr string,
	custodianIncAddressStr string) (*PortalLiquidateCustodian, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	liquidCustodianMeta := &PortalLiquidateCustodian{
		UniqueRedeemID:         uniqueRedeemID,
		TokenID:                tokenID,
		RedeemPubTokenAmount:   redeemAmount,
		RedeemerIncAddressStr:  redeemerIncAddressStr,
		CustodianIncAddressStr: custodianIncAddressStr,
	}
	liquidCustodianMeta.MetadataBase = metadataBase
	return liquidCustodianMeta, nil
}

func (liqCustodian PortalLiquidateCustodian) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (liqCustodian PortalLiquidateCustodian) ValidateSanityData(bcr BlockchainRetriever, txr Transaction, beaconHeight uint64) (bool, bool, error) {
	//// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	//if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
	//	return true, true, nil
	//}
	//
	//// validate RedeemerIncAddressStr
	//keyWallet, err := wallet.Base58CheckDeserialize(liqCustodian.RedeemerIncAddressStr)
	//if err != nil {
	//	return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("Requester incognito address is invalid"))
	//}
	//incAddr := keyWallet.KeySet.PaymentAddress
	//if len(incAddr.Pk) == 0 {
	//	return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("Requester incognito address is invalid"))
	//}
	//if !bytes.Equal(txr.GetSigPubKey()[:], incAddr.Pk[:]) {
	//	return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("Requester incognito address is not signer"))
	//}
	//
	//// check tx type
	//if txr.GetType() != common.TxCustomTokenPrivacyType {
	//	return false, false, errors.New("tx redeem request must be TxCustomTokenPrivacyType")
	//}
	//
	//if !txr.IsCoinsBurning(bcr, beaconHeight) {
	//	return false, false, errors.New("tx redeem request must be coin burning tx")
	//}
	//
	//// validate redeem amount
	//if liqCustodian.RedeemAmount == 0 {
	//	return false, false, errors.New("redeem amount should be larger than 0")
	//}
	//
	//// validate redeem fee
	//if liqCustodian.RedeemFee == 0 {
	//	return false, false, errors.New("redeem fee should be larger than 0")
	//}
	//
	//minFee, err := getMinRedeemFeeByRedeemAmount(liqCustodian.RedeemAmount)
	//if err != nil {
	//	return false, false, err
	//}
	//
	//if liqCustodian.RedeemFee < minFee {
	//	return false, false, fmt.Errorf("redeem fee should be larger than min fee %v\n", minFee)
	//}
	//
	//// validate value transfer of tx
	//if liqCustodian.RedeemAmount+liqCustodian.RedeemFee != txr.CalculateTxValue() {
	//	return false, false, errors.New("deposit amount should be equal to the tx value")
	//}
	//
	//// validate tokenID
	//if liqCustodian.TokenID != txr.GetTokenID().String() {
	//	return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("TokenID in metadata is not matched to tokenID in tx"))
	//}
	//// check tokenId is portal token or not
	//if !IsPortalToken(liqCustodian.TokenID) {
	//	return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("TokenID is not in portal tokens list"))
	//}
	//
	////validate RemoteAddress
	//// todo:
	//if len(liqCustodian.RemoteAddress) == 0 {
	//	return false, false, NewMetadataTxError(PortalRedeemRequestParamError, errors.New("Remote address is invalid"))
	//}

	return true, true, nil
}

func (liqCustodian PortalLiquidateCustodian) ValidateMetadataByItself() bool {
	return liqCustodian.Type == PortalLiquidateCustodianMeta
}

func (liqCustodian PortalLiquidateCustodian) Hash() *common.Hash {
	record := liqCustodian.MetadataBase.Hash().String()
	record += liqCustodian.UniqueRedeemID
	record += liqCustodian.TokenID
	record += strconv.FormatUint(liqCustodian.RedeemPubTokenAmount, 10)
	record += liqCustodian.RedeemerIncAddressStr
	record += liqCustodian.CustodianIncAddressStr
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

//func (liqCustodian *PortalLiquidateCustodian) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
//	actionContent := PortalRedeemRequestAction{
//		Meta:    *liqCustodian,
//		TxReqID: *tx.Hash(),
//		ShardID: shardID,
//	}
//	actionContentBytes, err := json.Marshal(actionContent)
//	if err != nil {
//		return [][]string{}, err
//	}
//	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
//	action := []string{strconv.Itoa(PortalRedeemRequestMeta), actionContentBase64Str}
//	return [][]string{action}, nil
//}

func (liqCustodian *PortalLiquidateCustodian) CalculateSize() uint64 {
	return calculateSize(liqCustodian)
}
