package pdexv3

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type MintNft struct {
	nftID       string
	otaReceiver string
	metadataCommon.MetadataBase
}

func NewMintNft() *MintNft {
	return &MintNft{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3MintNft,
		},
	}
}

func NewMintNftWithValue(nftID string, otaReceiver string) *MintNft {
	return &MintNft{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3MintNft,
		},
		nftID:       nftID,
		otaReceiver: otaReceiver,
	}
}

func (mintNft *MintNft) CheckTransactionFee(tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (mintNft *MintNft) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (mintNft *MintNft) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	otaReceiver := privacy.OTAReceiver{}
	err := otaReceiver.FromString(mintNft.otaReceiver)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !otaReceiver.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("ReceiveAddress is not valid"))
	}
	nftID, err := common.Hash{}.NewHashFromStr(mintNft.nftID)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if nftID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("TxReqID should not be empty"))
	}
	return true, true, nil
}

func (mintNft *MintNft) ValidateMetadataByItself() bool {
	return mintNft.Type == metadataCommon.Pdexv3MintNft
}

func (mintNft *MintNft) Hash() *common.Hash {
	record := mintNft.MetadataBase.Hash().String()
	record += mintNft.nftID
	record += mintNft.otaReceiver
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (mintNft *MintNft) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(mintNft)
}

func (mintNft *MintNft) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		NftID       string `json:"NftID"`
		OtaReceiver string `json:"OtaReceiver"`
		metadataCommon.MetadataBase
	}{
		NftID:        mintNft.nftID,
		OtaReceiver:  mintNft.otaReceiver,
		MetadataBase: mintNft.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (mintNft *MintNft) UnmarshalJSON(data []byte) error {
	temp := struct {
		NftID       string `json:"NftID"`
		OtaReceiver string `json:"OtaReceiver"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	mintNft.otaReceiver = temp.OtaReceiver
	mintNft.nftID = temp.NftID
	mintNft.MetadataBase = temp.MetadataBase
	return nil
}

func (mintNft *MintNft) OtaReceiver() string {
	return mintNft.otaReceiver
}

func (mintNft *MintNft) NftID() string {
	return mintNft.nftID
}

func (mintNft *MintNft) VerifyMinerCreatedTxBeforeGettingInBlock(
	mintData *metadataCommon.MintData,
	shardID byte,
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	ac *metadataCommon.AccumulatedValues,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
) (bool, error) {
	return true, nil
}
