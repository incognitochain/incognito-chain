package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

// PortalUserRegister - User register porting public tokens
type PortalUserRegister struct {
	MetadataBase
	UniqueRegisterId string //
	IncogAddressStr  string
	PTokenId         string
	RegisterAmount   uint64
	PortingFee       uint64
}

type PortalUserRegisterAction struct {
	Meta    PortalUserRegister
	TxReqID common.Hash
	ShardID byte
}

type PortalUserRegisterActionV3 struct {
	Meta        PortalUserRegister
	TxReqID     common.Hash
	ShardID     byte
	ShardHeight uint64
}

type PortalPortingRequestContent struct {
	UniqueRegisterId string
	IncogAddressStr  string
	PTokenId         string
	RegisterAmount   uint64
	PortingFee       uint64
	Custodian        []*statedb.MatchingPortingCustodianDetail
	TxReqID          common.Hash
	ShardID          byte
	ShardHeight      uint64
}

type PortingRequestStatus struct {
	UniquePortingID string
	TxReqID         common.Hash
	TokenID         string
	PorterAddress   string
	Amount          uint64
	Custodians      []*statedb.MatchingPortingCustodianDetail
	PortingFee      uint64
	Status          int
	BeaconHeight    uint64
	ShardHeight     uint64
	ShardID         byte
}

func NewPortingRequestStatus(
	uniquePortingID string,
	txReqID common.Hash,
	tokenID string,
	porterAddress string,
	amount uint64,
	custodians []*statedb.MatchingPortingCustodianDetail,
	portingFee uint64,
	status int,
	beaconHeight uint64,
	shardHeight uint64,
	shardID byte) *PortingRequestStatus {
	return &PortingRequestStatus{UniquePortingID: uniquePortingID, TxReqID: txReqID, TokenID: tokenID, PorterAddress: porterAddress, Amount: amount, Custodians: custodians, PortingFee: portingFee, Status: status, BeaconHeight: beaconHeight, ShardHeight: shardHeight, ShardID: shardID}
}

func NewPortalUserRegister(uniqueRegisterId string, incogAddressStr string, pTokenId string, registerAmount uint64, portingFee uint64, metaType int) (*PortalUserRegister, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}

	portalUserRegisterMeta := &PortalUserRegister{
		UniqueRegisterId: uniqueRegisterId,
		IncogAddressStr:  incogAddressStr,
		PTokenId:         pTokenId,
		RegisterAmount:   registerAmount,
		PortingFee:       portingFee,
	}

	portalUserRegisterMeta.MetadataBase = metadataBase

	return portalUserRegisterMeta, nil
}

func (portalUserRegister PortalUserRegister) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (portalUserRegister PortalUserRegister) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	if len(portalUserRegister.IncogAddressStr) <= 0 {
		return false, false, errors.New("IncogAddressStr should be not empty")
	}

	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(portalUserRegister.IncogAddressStr)
	if err != nil {
		return false, false, NewMetadataTxError(IssuingRequestNewIssuingRequestFromMapEror, errors.New("ContributorAddressStr incorrect"))
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

	if len(portalUserRegister.UniqueRegisterId) <= 0 {
		return false, false, errors.New("UniqueRegisterId should be not empty")
	}

	// validate amount register
	minAmount := common.MinAmountPortalPToken[portalUserRegister.PTokenId]
	if portalUserRegister.RegisterAmount < minAmount {
		return false, false, fmt.Errorf("register amount should be larger or equal to %v", minAmount)
	}

	//validation porting fee
	if portalUserRegister.PortingFee == 0 {
		return false, false, errors.New("porting fee should be larger than 0")
	}

	if (portalUserRegister.PortingFee) != txr.CalculateTxValue() {
		return false, false, errors.New("Total of register amount and porting fee should be equal to the tx value")
	}

	// validate metadata type
	if beaconHeight >= chainRetriever.GetBCHeightBreakPointPortalV3() && portalUserRegister.Type != PortalRequestPortingMetaV3 {
		return false, false, fmt.Errorf("Metadata type should be %v", PortalRequestPortingMetaV3)
	}

	return true, true, nil
}

func (portalUserRegister PortalUserRegister) ValidateMetadataByItself() bool {
	return portalUserRegister.Type == PortalRequestPortingMeta || portalUserRegister.Type == PortalRequestPortingMetaV3
}

func (portalUserRegister PortalUserRegister) Hash() *common.Hash {
	record := portalUserRegister.MetadataBase.Hash().String()
	record += portalUserRegister.UniqueRegisterId
	record += portalUserRegister.PTokenId
	record += portalUserRegister.IncogAddressStr
	record += strconv.FormatUint(portalUserRegister.RegisterAmount, 10)
	record += strconv.FormatUint(portalUserRegister.PortingFee, 10)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (portalUserRegister *PortalUserRegister) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	if portalUserRegister.Type == PortalRequestPortingMeta {
		actionContent := PortalUserRegisterAction{
			Meta:    *portalUserRegister,
			TxReqID: *tx.Hash(),
			ShardID: shardID,
		}
		actionContentBytes, err := json.Marshal(actionContent)
		if err != nil {
			return [][]string{}, err
		}
		actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
		action := []string{strconv.Itoa(PortalRequestPortingMeta), actionContentBase64Str}
		return [][]string{action}, nil
	} else if portalUserRegister.Type == PortalRequestPortingMetaV3 {
		actionContent := PortalUserRegisterActionV3{
			Meta:        *portalUserRegister,
			TxReqID:     *tx.Hash(),
			ShardID:     shardID,
			ShardHeight: shardHeight,
		}
		actionContentBytes, err := json.Marshal(actionContent)
		if err != nil {
			return [][]string{}, err
		}
		actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
		action := []string{strconv.Itoa(PortalRequestPortingMetaV3), actionContentBase64Str}
		return [][]string{action}, nil
	}
	return nil, nil
}

func (portalUserRegister *PortalUserRegister) CalculateSize() uint64 {
	return calculateSize(portalUserRegister)
}
