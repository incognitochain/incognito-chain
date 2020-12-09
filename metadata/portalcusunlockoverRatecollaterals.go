package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

type PortalUnlockOverRateCollaterals struct {
	MetadataBase
	CustodianAddressStr string
	TokenID             string
}

type PortalUnlockOverRateCollateralsAction struct {
	Meta    PortalUnlockOverRateCollaterals
	TxReqID common.Hash
	ShardID byte
}

type UnlockOverRateCollateralsRequestStatus struct {
	Status              byte
	CustodianAddressStr string
	TokenID             string
	UnlockedAmounts     map[string]uint64
}

func NewUnlockOverRateCollateralsRequestStatus(status byte, custodianAddr string, tokenID string, unlockAmount map[string]uint64) *UnlockOverRateCollateralsRequestStatus {
	return &UnlockOverRateCollateralsRequestStatus{Status: status, CustodianAddressStr: custodianAddr, TokenID: tokenID, UnlockedAmounts: unlockAmount}
}

func NewPortalUnlockOverRateCollaterals(metaType int, custodianAddr string, tokenID string) (*PortalUnlockOverRateCollaterals, error) {
	metadataBase := MetadataBase{Type: metaType}

	portalUnlockOverRateCollaterals := &PortalUnlockOverRateCollaterals{
		CustodianAddressStr: custodianAddr,
		TokenID:             tokenID,
	}

	portalUnlockOverRateCollaterals.MetadataBase = metadataBase

	return portalUnlockOverRateCollaterals, nil
}

type PortalUnlockOverRateCollateralsContent struct {
	CustodianAddressStr string
	TokenID             string
	TxReqID             common.Hash
	UnlockedAmounts     map[string]uint64
}

func (portalUnlockCs PortalUnlockOverRateCollaterals) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (portalUnlockCs PortalUnlockOverRateCollaterals) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(portalUnlockCs.CustodianAddressStr)
	if err != nil {
		return false, false, errors.New("CustodianAddressStr incorrect")
	}

	senderAddr := keyWallet.KeySet.PaymentAddress
	if len(senderAddr.Pk) == 0 {
		return false, false, errors.New("Custodian address invalid, custodian address must be incognito address")
	}

	if !bytes.Equal(txr.GetSigPubKey()[:], senderAddr.Pk[:]) {
		return false, false, errors.New("Custodian address is not signer tx")
	}

	if txr.GetType() != common.TxNormalType {
		return false, false, errors.New("Tx unlock over rate collaterals must be TxNormalType")
	}

	// check tokenId is portal token or not
	if !IsPortalToken(portalUnlockCs.TokenID) {
		return false, false, NewMetadataTxError(PortalUnlockOverRateCollateralsError, errors.New("TokenID is not in portal tokens list"))
	}

	return true, true, nil
}

func (portalUnlockCs PortalUnlockOverRateCollaterals) ValidateMetadataByItself() bool {
	return portalUnlockCs.Type == PortalUnlockOverRateCollateralsMeta
}

func (portalUnlockCs PortalUnlockOverRateCollaterals) Hash() *common.Hash {
	record := portalUnlockCs.MetadataBase.Hash().String()
	record += portalUnlockCs.TokenID
	record += portalUnlockCs.CustodianAddressStr
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (portalUnlockCs *PortalUnlockOverRateCollaterals) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalUnlockOverRateCollateralsAction{
		Meta:    *portalUnlockCs,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}

	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalUnlockOverRateCollateralsMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (portalUnlockCs *PortalUnlockOverRateCollaterals) CalculateSize() uint64 {
	return calculateSize(portalUnlockCs)
}
