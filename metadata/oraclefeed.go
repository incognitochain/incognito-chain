package metadata

import (
	"bytes"
	"errors"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type OracleFeed struct {
	FeederAddress privacy.PaymentAddress
	AssetType     common.Hash
	Price         uint64 // in USD
	MetadataBase
}

func NewOracleFeed(
	assetType common.Hash,
	price uint64,
	metaType int,
	feederAddress privacy.PaymentAddress,
) *OracleFeed {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &OracleFeed{
		AssetType:     assetType,
		Price:         price,
		MetadataBase:  metadataBase,
		FeederAddress: feederAddress,
	}
}

func (of *OracleFeed) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	govParams := bcr.GetGOVParams()
	oraclePubKeys := govParams.OracleNetwork.OraclePubKeys
	senderPubKey := txr.GetSigPubKey()
	for _, oraclePubKey := range oraclePubKeys {
		if bytes.Equal(oraclePubKey, senderPubKey) {
			return true, nil
		}
	}
	return true, errors.New("The oracle feeder is not belong to eligible oracles.")
}

func (of *OracleFeed) ValidateSanityData(
	bcr BlockchainRetriever,
	txr Transaction,
) (bool, bool, error) {
	if len(of.FeederAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if len(of.FeederAddress.Tk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if of.Price == 0 {
		return false, false, errors.New("Wrong oracle feed's price")
	}
	if len(of.AssetType) != common.HashSize {
		return false, false, errors.New("Wrong oracle feed's asset type")
	}
	return true, true, nil
}

func (of *OracleFeed) ValidateMetadataByItself() bool {
	if of.Type != OracleFeedMeta {
		return false
	}
	if !bytes.Equal(of.AssetType[:], common.DCBTokenID[:]) &&
		!bytes.Equal(of.AssetType[:], common.GOVTokenID[:]) &&
		!bytes.Equal(of.AssetType[:], common.CMBTokenID[:]) &&
		!bytes.Equal(of.AssetType[:], common.ConstantID[:]) &&
		!bytes.Equal(of.AssetType[:], common.ETHAssetID[:]) &&
		!bytes.Equal(of.AssetType[:], common.BTCAssetID[:]) &&
		!bytes.Equal(of.AssetType[:8], common.BondTokenID[:8]) {
		return false
	}
	return true
}

func (of *OracleFeed) Hash() *common.Hash {
	record := of.AssetType.String()
	record += of.FeederAddress.String()
	record += string(of.Price)
	record += of.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (of *OracleFeed) CalculateSize() uint64 {
	return calculateSize(of)
}
