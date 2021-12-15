package pdexv3

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/utils"
)

type ReceiverInfo struct {
	Address privacy.OTAReceiver `json:"Address"`
	Amount  uint64              `json:"Amount"`
}

// Check if the given OTA address is a valid address and has the expected shard ID
func isValidOTAReceiver(receiverAddress privacy.OTAReceiver, expectedShardID byte) (privacy.OTAReceiver, error) {
	if !receiverAddress.IsValid() {
		return receiverAddress, errors.New("ReceiverAddress is invalid")
	}

	pkb := receiverAddress.PublicKey.ToBytesS()
	currentShardID := common.GetShardIDFromLastByte(pkb[len(pkb)-1])
	if currentShardID != expectedShardID {
		return receiverAddress, errors.New("ReceiverAddress shard ID is wrong")
	}

	return receiverAddress, nil
}

type AccessOTA privacy.Point

func (ota AccessOTA) MarshalJSON() ([]byte, error) {
	temp := common.Hash((privacy.Point)(ota).ToBytes())
	return json.Marshal(temp)
}

func (ota *AccessOTA) UnmarshalJSON(data []byte) error {
	var temp common.Hash
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	p, err := (&privacy.Point{}).FromBytes([32]byte(temp))
	if p != nil {
		*ota = AccessOTA(*p)
	}
	return err
}

func (ota AccessOTA) Bytes() [32]byte {
	return privacy.Point(ota).ToBytes()
}

func (ota *AccessOTA) FromBytes(data [32]byte) error {
	_, err := (*privacy.Point)(ota).FromBytes(data)
	return err
}

type BurnInputReader interface {
	DerivableBurnInput(*statedb.StateDB) (map[common.Hash]privacy.Point, error)
}

func ValidPdexv3Access(burnOTA *AccessOTA, nextOTA AccessOTA, tx metadataCommon.Transaction, accessTokenID common.Hash, transactionStateDB *statedb.StateDB) (bool, error) {
	if accessTokenID != common.PRVCoinID {
		accessTokenID = common.ConfidentialAssetID
	}
	// check that nextOTA is valid
	p := (*privacy.Point)(&nextOTA)
	if !p.PointValid() {
		return false, fmt.Errorf("invalid point - next AccessOTA %v", nextOTA.Bytes())
	}
	rawPubkey := p.ToBytesS()

	exists, otaStatus, err := statedb.HasOnetimeAddress(transactionStateDB, accessTokenID, rawPubkey)
	if err != nil {
		return false, err
	}
	// allow only OTAs from stored coins & not OTA-declarations
	if !exists {
		return false, fmt.Errorf("next AccessOTA %s does not exist", p.String())
	}
	if otaStatus == statedb.OTA_STATUS_OCCUPIED {
		return false, fmt.Errorf("next AccessOTA %s has invalid status %v", statedb.OTA_STATUS_OCCUPIED)
	}

	if burnOTA == nil {
		// accept metadata that declare nextAccessOTA the 1st time
		return true, nil
	}
	// check that burnOTA is valid
	{
		burnOTAPoint := (*privacy.Point)(burnOTA)
		if !burnOTAPoint.PointValid() {
			return false, fmt.Errorf("invalid point - burnt AccessOTA")
		}
		rawPubkey := burnOTAPoint.ToBytesS()

		exists, otaStatus, err := statedb.HasOnetimeAddress(transactionStateDB, accessTokenID, rawPubkey)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, fmt.Errorf("burnt AccessOTA %s does not exist", burnOTAPoint.String())
		}
		if otaStatus == statedb.OTA_STATUS_OCCUPIED {
			return false, fmt.Errorf("burnt AccessOTA %s has invalid status %v", statedb.OTA_STATUS_OCCUPIED)
		}

		// check that burnOTA is a tx input. Param tx cannot be nil
		btx, ok := tx.(BurnInputReader)
		if !ok {
			return false, fmt.Errorf("cannot identify burn input for tx %s", tx.Hash().String())
		}
		burnMap, err := btx.DerivableBurnInput(transactionStateDB)
		if err != nil {
			return false, fmt.Errorf("cannot identify burn input for tx %s: %v", tx.Hash().String(), err)
		}
		if burnInputCoinPubkey, exists := burnMap[common.PdexAccessCoinID]; !exists || !privacy.IsPointEqual(burnOTAPoint, &burnInputCoinPubkey) {
			return false, fmt.Errorf("burn missing: tx %s must burn AccessOTA %s from metadata", tx.Hash().String(), burnOTAPoint.String())
		}
		// check that burn output is PdexAccess, with sufficient amount
		isBurn, _, burnedTokenCoin, burnedTokenID, err := tx.GetTxFullBurnData()
		if !isBurn || *burnedTokenID != common.PdexAccessCoinID || burnedTokenCoin.GetValue() < MinPdexv3AccessBurn {
			return false, fmt.Errorf("invalid burn output (%v, %d) for Pdexv3 Access", burnedTokenID, burnedTokenCoin.GetValue())
		}
	}

	return true, nil
}

type AccessOption struct {
	NftID    common.Hash `json:"NftID"` // can not omitempty
	BurntOTA *AccessOTA  `json:"BurntOTA,omitempty"`
	AssetID  string      `json:"AssetID,omitempty"`
}

func NewAccessOption() *AccessOption {
	return &AccessOption{}
}

func NewAccessOptionWithValue(burntOTA *AccessOTA, nftID common.Hash, assetID string) *AccessOption {
	return &AccessOption{
		BurntOTA: burntOTA,
		NftID:    nftID,
		AssetID:  assetID,
	}
}

func (a *AccessOption) IsValid(
	tx metadataCommon.Transaction,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	transactionStateDB *statedb.StateDB,
	isWithdrawRequest bool,
) error {
	err := beaconViewRetriever.IsValidNftID(a.NftID.String())
	if err != nil {
		if isWithdrawRequest {
			if a.AssetID == utils.EmptyString || a.BurntOTA == nil {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("%v - %v", err, errors.New("burnt OTA or assetID is null")))
			}
			_, err = a.AssetHash()
			if err != nil {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
			}
			/*valid, err1 := ValidPdexv3Access(a.BurntOTA, tx, common.ConfidentialAssetID, transactionStateDB)*/
			/*if !valid {*/
			/*return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("%v - %v", err, err1))*/
			/*}*/
		}
		return nil
	}
	if a.BurntOTA != nil || a.AssetID != utils.EmptyString {
		return errors.New("NftID and (burnOTA or assetID) can not be existed at the same time")
	}
	return nil
}

func (a *AccessOption) IsEmpty() bool {
	return a.NftID.IsZeroValue() && a.BurntOTA == nil && a.AssetID == utils.EmptyString
}

func (a *AccessOption) UseNft() bool {
	return !a.NftID.IsZeroValue()
}

func (a *AccessOption) BurntOTAStringify() (string, error) {
	return "burnt-ota", nil
}

func (a *AccessOption) AssetHash() (common.Hash, error) {
	hash, err := common.Hash{}.NewHashFromStr(a.AssetID)
	return *hash, err
}

func (a *AccessOption) ValidateOtaReceivers(tx metadataCommon.Transaction, otaReceiver string, otaReceivers map[common.Hash]*privacy.OTAReceiver) error {
	if otaReceivers != nil && otaReceiver != utils.EmptyString {
		return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver and otaReceivers can not be existed at the same time"))
	}
	if a.UseNft() {
		receiver := privacy.OTAReceiver{}
		err := receiver.FromString(otaReceiver)
		if err != nil {
			return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
		}
		if !receiver.IsValid() {
			return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("RefundAddress is not valid"))
		}
		if receiver.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
			return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver shardID is different from txShardID"))
		}
	} else {
		isExistedPdexAccessToken := false
		for k, v := range otaReceivers {
			if k == common.PdexAccessCoinID {
				isExistedPdexAccessToken = true
			}
			if !v.IsValid() {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver is not valid"))
			}
			if v.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver shard id and tx shard id need to be similar"))
			}
		}
		if !isExistedPdexAccessToken {
			return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("is not existed pdex access token in otaReceivers"))
		}
	}
	return nil
}

func GenAssetID(otaReceiver privacy.OTAReceiver) common.Hash {
	pubKey := otaReceiver.PublicKey.ToBytes()
	return common.HashH(pubKey[:])
}
