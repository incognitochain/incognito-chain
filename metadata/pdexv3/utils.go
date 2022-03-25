package pdexv3

import (
	"encoding/base64"
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
	return json.Marshal(ota.ToBytesS())
}

func (ota *AccessOTA) UnmarshalJSON(data []byte) error {
	var b []byte
	err := json.Unmarshal(data, &b)
	if err != nil {
		return err
	}
	return ota.FromBytesS(b)
}

func (ota AccessOTA) ToBytes() [32]byte {
	return privacy.Point(ota).ToBytes()
}

func (ota *AccessOTA) FromBytes(data [32]byte) error {
	_, err := (*privacy.Point)(ota).FromBytes(data)
	return err
}

func (ota AccessOTA) ToBytesS() []byte {
	temp := ota.ToBytes()
	return temp[:]
}

func (ota *AccessOTA) FromBytesS(data []byte) error {
	if len(data) != 32 {
		return fmt.Errorf("Invalid AccessOTA byte length %d", len(data))
	}
	var temp [32]byte
	copy(temp[:], data)
	*ota = AccessOTA{}
	err := ota.FromBytes(temp)
	return err
}

func (ota *AccessOTA) String() string {
	return base64.StdEncoding.EncodeToString(ota.ToBytesS())
}

func (ota *AccessOTA) FromString(str string) error {
	temp, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}
	return ota.FromBytesS(temp)
}

type BurnInputReader interface {
	DerivableBurnInput(*statedb.StateDB) (map[common.Hash]privacy.Point, error)
}

func ValidPdexv3Access(burnOTA *AccessOTA, nextOTA privacy.Point, tx metadataCommon.Transaction, transactionStateDB *statedb.StateDB) (bool, error) {
	// check that nextOTA is valid
	if !nextOTA.PointValid() {
		return false, fmt.Errorf("invalid point - next AccessOTA %x", nextOTA.ToBytesS())
	}

	// reject OTAs already in db; already checked with txs double-spend verification
	// exists, otaStatus, err := statedb.HasOnetimeAddress(transactionStateDB, common.ConfidentialAssetID, rawPubkey)
	// if err != nil {
	// 	return false, err
	// }
	// if exists {
	// 	return false, fmt.Errorf("next AccessOTA %v already exists", nextOTA)
	// }

	if burnOTA == nil {
		// accept metadata that declare nextAccessOTA the 1st time
		return true, nil
	}
	// check that burnOTA is valid
	burnOTAPoint := (*privacy.Point)(burnOTA)
	if !burnOTAPoint.PointValid() {
		return false, fmt.Errorf("invalid point - burnt AccessOTA")
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

	return true, nil
}

type AccessOption struct {
	NftID    *common.Hash `json:"NftID,omitempty"`
	BurntOTA *AccessOTA   `json:"BurntOTA,omitempty"`
	AccessID *common.Hash `json:"AccessID,omitempty"`
}

func NewAccessOption() *AccessOption {
	return &AccessOption{}
}

func NewAccessOptionWithValue(burntOTA *AccessOTA, nftID, accessID *common.Hash) *AccessOption {
	return &AccessOption{
		BurntOTA: burntOTA,
		NftID:    nftID,
		AccessID: accessID,
	}
}

func (a *AccessOption) IsValid(
	tx metadataCommon.Transaction,
	receivers map[common.Hash]privacy.OTAReceiver,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	transactionStateDB *statedb.StateDB,
	isWithdrawalRequest bool,
	isNewAccessOTALpRequest bool,
	accessReceiverStr string,
) error {
	if a.NftID != nil {
		if a.NftID.IsZeroValue() {
			return fmt.Errorf("invalid NftID %s", a.NftID.String())
		}
		if a.BurntOTA != nil || a.AccessID != nil {
			return fmt.Errorf("invalid AccessOTA (%v, %v) when using NftID; expect none", a.BurntOTA, a.AccessID)
		}
		ok, err := beaconViewRetriever.IsValidPdexv3NftID(a.NftID.String())
		if err != nil || !ok {
			if err == nil {
				err = fmt.Errorf("NftID %s is not valid", a.NftID.String())
			}
			return err
		}
	} else {
		shouldValidateAccessReceiver := false
		if isWithdrawalRequest {
			if a.AccessID == nil || a.AccessID.IsZeroValue() {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("invalid AccessID (zero value)"))
			}
			if a.BurntOTA == nil {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("%v", errors.New("burnt OTA missing")))
			}
			shouldValidateAccessReceiver = true
		} else {
			if a.AccessID == nil {
				shouldValidateAccessReceiver = true
			}
		}
		if receivers == nil || len(receivers) == 0 {
			return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("%v", errors.New("OTA receivers missing")))
		}
		if shouldValidateAccessReceiver {
			accessReceiver := privacy.OTAReceiver{}
			if isNewAccessOTALpRequest {
				err := accessReceiver.FromString(accessReceiverStr)
				if err != nil {
					return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("%v", err))
				}
			} else {
				var exists bool
				accessReceiver, exists = receivers[common.PdexAccessCoinID]
				if !exists {
					return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("%v", errors.New("accessReceiver missing")))
				}
			}
			valid, err := ValidPdexv3Access(a.BurntOTA, accessReceiver.PublicKey, tx, transactionStateDB)
			if !valid {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
			}
		}
	}

	return nil
}

func (a *AccessOption) IsEmpty() bool {
	isEmptyNftID := true
	isEmptyBurntOTA := true
	isEmptyAccessID := true

	if a.NftID != nil {
		if !a.NftID.IsZeroValue() {
			isEmptyNftID = false
		}
	}
	if a.AccessID != nil {
		if !a.AccessID.IsZeroValue() {
			isEmptyAccessID = false
		}
	}
	if a.BurntOTA != nil && len(a.BurntOTA.ToBytesS()) > 0 {
		isEmptyBurntOTA = false
	}
	return isEmptyAccessID && isEmptyBurntOTA && isEmptyNftID
}

func (a *AccessOption) UseNft() bool {
	return a.NftID != nil && !a.NftID.IsZeroValue()
}

func (a *AccessOption) ValidateOtaReceivers(
	tx metadataCommon.Transaction, otaReceiver string,
	otaReceivers map[common.Hash]privacy.OTAReceiver,
	tokenHash common.Hash, isNewAccessOTALpRequest bool,
) error {
	if (otaReceivers == nil || len(otaReceivers) == 0) && otaReceiver == utils.EmptyString {
		return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver and otaReceivers can not be null at the same time"))
	}
	if a.UseNft() {
		if otaReceivers != nil && otaReceiver != utils.EmptyString {
			return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver and otaReceivers can not exist at the same time"))
		}
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
		if otaReceivers == nil || len(otaReceivers) == 0 {
			return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceivers can not be null or empty"))
		}
		if !isNewAccessOTALpRequest {
			if otaReceivers != nil && otaReceiver != utils.EmptyString {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver and otaReceivers can not exist at the same time"))
			}
		} else {
			if a.AccessID != nil {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("AccessID need to be null"))
			}
			o := privacy.OTAReceiver{}
			err := o.FromString(otaReceiver)
			if err != nil {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
			}
			if o.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver shard id and tx shard id need to be the same"))
			}
		}
		if _, found := otaReceivers[tokenHash]; !found {
			return errors.New("Can not find otaReceiver for burnt tokenID")
		}
		isExistedPdexAccessToken := false
		for k, v := range otaReceivers {
			if k == common.PdexAccessCoinID {
				isExistedPdexAccessToken = true
			}
			if !v.IsValid() {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver is not valid"))
			}
			if v.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver shard id and tx shard id need to be the same"))
			}
		}

		if isNewAccessOTALpRequest {
			if isExistedPdexAccessToken {
				temp, _ := otaReceivers[common.PdexAccessCoinID].String() // otaReceiver valid above
				if temp != otaReceiver {
					return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver and otaReceivers[PdexAccessCoinID] need to be the same"))
				}
			}
		} else {
			if !isExistedPdexAccessToken && a.AccessID == nil {
				return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("is not existed pdex access token in otaReceivers"))
			}
		}
	}
	return nil
}

func GenAccessID(otaReceiver privacy.OTAReceiver) common.Hash {
	pubKey := otaReceiver.PublicKey.ToBytes()
	return common.HashH(pubKey[:])
}

func GenAccessOTA(otaReceiver privacy.OTAReceiver) ([]byte, error) {
	tempAccessOTA := AccessOTA{}
	err := tempAccessOTA.FromBytesS(otaReceiver.PublicKey.ToBytesS())
	return tempAccessOTA.ToBytesS(), err
}

func GenAccessOTAByStr(otaReceiver string) ([]byte, error) {
	tempAccessOTA := AccessOTA{}
	tempOtaReceiver := privacy.OTAReceiver{}
	err := tempOtaReceiver.FromString(otaReceiver)
	if err != nil {
		return nil, err
	}
	err = tempAccessOTA.FromBytesS(tempOtaReceiver.PublicKey.ToBytesS())
	return tempAccessOTA.ToBytesS(), err
}
