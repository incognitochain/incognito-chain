package pdexv3

import (
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type MintPDEXGenesisResponse struct {
	metadataCommon.MetadataBase
	MintingPaymentAddress string `json:"MintingPaymentAddress"`
	MintingAmount         uint64 `json:"MintingAmount"`
	SharedRandom          []byte `json:"SharedRandom"`
}

func NewPdexv3MintPDEXGenesisResponse(
	metaType int,
	mintingPaymentAddress string,
	mintingAmount uint64,
) *MintPDEXGenesisResponse {
	metadataBase := metadataCommon.NewMetadataBase(metaType)

	return &MintPDEXGenesisResponse{
		MetadataBase:          *metadataBase,
		MintingPaymentAddress: mintingPaymentAddress,
		MintingAmount:         mintingAmount,
	}
}

func (mintResponse MintPDEXGenesisResponse) CheckTransactionFee(
	tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB,
) bool {
	// no need to have fee for this tx
	return true
}

func (mintResponse MintPDEXGenesisResponse) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (mintResponse MintPDEXGenesisResponse) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	return false, true, nil
}

func (mintResponse MintPDEXGenesisResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return mintResponse.Type == metadataCommon.Pdexv3MintPDEXGenesisResponseMeta
}

func (mintResponse MintPDEXGenesisResponse) Hash() *common.Hash {
	record := mintResponse.MetadataBase.Hash().String()
	record += mintResponse.MintingPaymentAddress
	record += strconv.FormatUint(mintResponse.MintingAmount, 10)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (mintResponse *MintPDEXGenesisResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(mintResponse)
}

func (mintResponse MintPDEXGenesisResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	mintData *metadataCommon.MintData,
	shardID byte, tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	ac *metadataCommon.AccumulatedValues,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
) (bool, error) {
	// TODO: verify mining tx with the instruction
	return true, nil
}

func (mintResponse *MintPDEXGenesisResponse) SetSharedRandom(r []byte) {
	mintResponse.SharedRandom = r
}
