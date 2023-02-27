package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type MintPDEXGenesisResponse struct {
	metadataCommon.MetadataBase
	MintingPaymentAddress string `json:"MintingPaymentAddress"`
	MintingAmount         uint64 `json:"MintingAmount"`
	SharedRandom          []byte `json:"SharedRandom"`
}

type MintBlockRewardContent struct {
	PoolPairID string      `json:"PoolPairID"`
	Amount     uint64      `json:"Amount"`
	TokenID    common.Hash `json:"TokenID"`
}

type DistributeMiningOrderRewardContent struct {
	PoolPairID    string      `json:"PoolPairID"`
	MakingTokenID common.Hash `json:"MakingTokenID"`
	Amount        uint64      `json:"Amount"`
	TokenID       common.Hash `json:"TokenID"`
}

type MintPDEXGenesisContent struct {
	MintingPaymentAddress string `json:"MintingPaymentAddress"`
	MintingAmount         uint64 `json:"MintingAmount"`
	ShardID               byte   `json:"ShardID"`
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
	return mintResponse.Type == metadataCommon.Pdexv3MintPDEXGenesisMeta
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

func (mintResponse *MintPDEXGenesisResponse) ToCompactBytes() ([]byte, error) {
	return metadataCommon.ToCompactBytes(mintResponse)
}

func (mintResponse MintPDEXGenesisResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	mintData *metadataCommon.MintData,
	shardID byte, tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	ac *metadataCommon.AccumulatedValues,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
) (bool, error) {
	// verify mining tx with the request tx
	idx := -1
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not MintPDEXGenesisResponse instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || (instMetaType != strconv.Itoa(metadataCommon.Pdexv3MintPDEXGenesisMeta)) {
			continue
		}
		instReqStatus := inst[2]
		if instReqStatus != RequestAcceptedChainStatus {
			continue
		}

		contentBytes := []byte(inst[3])
		var instContent MintPDEXGenesisContent
		err := json.Unmarshal(contentBytes, &instContent)
		if err != nil {
			continue
		}
		shardIDFromInst := instContent.ShardID

		if shardID != shardIDFromInst {
			continue
		}

		isMinted, mintCoin, assetID, err := tx.GetTxMintData()
		if err != nil || !isMinted || *assetID != common.PDEXCoinID {
			continue
		}

		keyWallet, err := wallet.Base58CheckDeserialize(instContent.MintingPaymentAddress)
		if err != nil {
			continue
		}
		paymentAddress := keyWallet.KeySet.PaymentAddress
		if ok := mintCoin.CheckCoinValid(paymentAddress, mintResponse.SharedRandom, instContent.MintingAmount); !ok {
			continue
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("No MintingPDEXGenesis instruction found for MintPDEXGenesisResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (mintResponse *MintPDEXGenesisResponse) SetSharedRandom(r []byte) {
	mintResponse.SharedRandom = r
}
