package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	relaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

// PortalRequestPTokens - portal user requests ptoken (after sending pubToken to custodians)
type PortalRequestPTokens struct {
	MetadataBase
	UniquePortingID string
	TokenID         string
	IncogAddressStr string
	PortingAmount   uint64
	PortingProof    string
}

type PortalRequestPTokensAction struct {
	Meta    PortalRequestPTokens
	TxReqID common.Hash
	ShardID byte
}

func NewPortalRequestPTokens(metaType int, uniquePortingID string,
	tokenID string,
	incogAddressStr string,
	portingAmount uint64,
	portingProof string) (*PortalRequestPTokens, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	requestPTokenMeta := &PortalRequestPTokens{
		UniquePortingID: uniquePortingID,
		TokenID:         tokenID,
		IncogAddressStr: incogAddressStr,
		PortingAmount:   portingAmount,
		PortingProof:    portingProof,
	}
	requestPTokenMeta.MetadataBase = metadataBase
	return requestPTokenMeta, nil
}

//todo
func (reqPToken PortalRequestPTokens) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (reqPToken PortalRequestPTokens) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	//if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
	//	return true, true, nil
	//}

	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(reqPToken.IncogAddressStr)
	if err != nil {
		return false, false, NewMetadataTxError(PortalRequestPTokenParamError, errors.New("Requester incognito address is invalid"))
	}
	incogAddr := keyWallet.KeySet.PaymentAddress
	if len(incogAddr.Pk) == 0 {
		return false, false, NewMetadataTxError(PortalRequestPTokenParamError, errors.New("Requester incognito address is invalid"))
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incogAddr.Pk[:]) {
		return false, false, NewMetadataTxError(PortalRequestPTokenParamError, errors.New("Requester incognito address is not signer"))
	}

	// check tx type
	if txr.GetType() != common.TxNormalType {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}

	// validate amount deposit
	if reqPToken.PortingAmount == 0 {
		return false, false, errors.New("porting amount should be larger than 0")
	}

	// validate tokenID and porting proof
	if reqPToken.TokenID == PortalSupportedTokenMap["BTC"] {
		// token BTC
		// convert porting proof to BTC proof
		// todo:
	} else if reqPToken.TokenID == PortalSupportedTokenMap["BNB"] {
		// token BNB
		// convert porting proof to BNB proof

		_, err := relaying.ParseProofFromB64EncodeJsonStr(reqPToken.PortingProof)
		if err != nil {
			return false, false, NewMetadataTxError(PortalRequestPTokenParamError, err)
		}
	} else {
		return false, false, NewMetadataTxError(PortalRequestPTokenParamError, errors.New("TokenID is not supported currently on Portal"))
	}
	return true, true, nil
}

func (reqPToken PortalRequestPTokens) ValidateMetadataByItself() bool {
	return reqPToken.Type == PortalCustodianDepositMeta
}

func (reqPToken PortalRequestPTokens) Hash() *common.Hash {
	record := reqPToken.MetadataBase.Hash().String()
	record += reqPToken.UniquePortingID
	record += reqPToken.TokenID
	record += reqPToken.IncogAddressStr
	record += strconv.FormatUint(reqPToken.PortingAmount, 10)
	record += reqPToken.PortingProof
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (reqPToken *PortalRequestPTokens) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := PortalRequestPTokensAction{
		Meta:    *reqPToken,
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

func (reqPToken *PortalRequestPTokens) CalculateSize() uint64 {
	return calculateSize(reqPToken)
}
