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
	chainID byte,
	db database.DatabaseInterface,
) (bool, error) {
	govParams := bcr.GetGOVParams()
	oraclePubKeys := govParams.OracleNetwork.OraclePubKeys
	senderPubKey := txr.GetSigPubKey()
	for _, oraclePubKey := range oraclePubKeys {
		if bytes.Equal(oraclePubKey, senderPubKey) {
			return common.TrueValue, nil
		}
	}
	return common.TrueValue, errors.New("The oracle feeder is not belong to eligible oracles.")
}

func (of *OracleFeed) ValidateSanityData(
	bcr BlockchainRetriever,
	txr Transaction,
) (bool, bool, error) {
	if len(of.FeederAddress.Pk) == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's payment address")
	}
	if len(of.FeederAddress.Tk) == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's payment address")
	}
	if of.Price == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong oracle feed's price")
	}
	if len(of.AssetType) != common.HashSize {
		return common.FalseValue, common.FalseValue, errors.New("Wrong oracle feed's asset type")
	}
	return common.TrueValue, common.TrueValue, nil
}

func (of *OracleFeed) ValidateMetadataByItself() bool {
	if of.Type != OracleFeedMeta {
		return common.FalseValue
	}
	if !bytes.Equal(of.AssetType[:], common.DCBTokenID[:]) &&
		!bytes.Equal(of.AssetType[:], common.GOVTokenID[:]) &&
		!bytes.Equal(of.AssetType[:], common.CMBTokenID[:]) &&
		!bytes.Equal(of.AssetType[:], common.ConstantID[:]) &&
		!bytes.Equal(of.AssetType[:], common.ETHAssetID[:]) &&
		!bytes.Equal(of.AssetType[:], common.BTCAssetID[:]) &&
		!bytes.Equal(of.AssetType[:8], common.BondTokenID[:8]) {
		return common.FalseValue
	}
	return common.TrueValue
}

func (of *OracleFeed) Hash() *common.Hash {
	record := of.AssetType.String()
	record += string(of.FeederAddress.Bytes())
	record += string(of.Price)
	record += of.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
