package pdexv3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
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

type MintNftData struct {
	NftID       common.Hash `json:"NftID"`
	OtaReceiver string      `json:"OtaReceiver"`
	ShardID     byte        `json:"ShardID"`
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
	idx := -1
	metadataCommon.Logger.Log.Infof("Currently verifying ins: %v\n", mintNft)
	metadataCommon.Logger.Log.Infof("BUGLOG There are %v inst\n", len(mintData.Insts))
	for i, inst := range mintData.Insts {
		if len(inst) != 3 { // this is not PDEContribution instruction
			continue
		}

		metadataCommon.Logger.Log.Infof("BUGLOG currently processing inst: %v\n", inst)

		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || instMetaType != strconv.Itoa(metadataCommon.Pdexv3MintNft) {
			continue
		}

		var shardIDFromInst byte
		var receiverAddrStrFromInst string
		var receivingAmtFromInst uint64
		var receivingTokenIDStr string

		contentBytes := []byte(inst[2])
		var mintNftData MintNftData
		err := json.Unmarshal(contentBytes, &mintNftData)
		if err != nil {
			return false, err
		}
		shardIDFromInst = mintNftData.ShardID
		receiverAddrStrFromInst = mintNftData.OtaReceiver
		receivingTokenIDStr = mintNftData.NftID.String()
		receivingAmtFromInst = 1

		if shardID != shardIDFromInst {
			metadataCommon.Logger.Log.Infof("BUGLOG shardID: %v, %v\n", shardID, shardIDFromInst)
			continue
		}

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil {
			metadataCommon.Logger.Log.Error("ERROR - VALIDATION: an error occured while get tx mint data: ", err)
			return false, err
		}
		if !isMinted {
			metadataCommon.Logger.Log.Info("WARNING - VALIDATION: this is not Tx Mint: ")
			return false, errors.New("This is not tx mint")
		}
		pk := mintCoin.GetPublicKey().ToBytesS()
		paidAmount := mintCoin.GetValue()

		otaReceiver := coin.OTAReceiver{}
		err = otaReceiver.FromString(receiverAddrStrFromInst)
		if err != nil {
			return false, errors.New("Invalid ota receiver")
		}

		txR := mintCoin.(*coin.CoinV2).GetTxRandom()
		if !bytes.Equal(otaReceiver.PublicKey.ToBytesS(), pk[:]) ||
			receivingAmtFromInst != paidAmount ||
			!bytes.Equal(txR[:], otaReceiver.TxRandom[:]) ||
			receivingTokenIDStr != coinID.String() {
			return false, errors.New("Coin is invalid")
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		metadataCommon.Logger.Log.Debugf("[pdex] mint nft %s error", mintNft.nftID)
		return false, fmt.Errorf("Can't mint nft with hash %s", mintNft.nftID)
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}
