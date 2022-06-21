package bridge

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type ModifyRewardReserve struct {
	metadataCommon.MetadataBaseWithSignature
	Vaults map[common.Hash][]Vault `json:"Vaults"` // unifiedTokenID -> list vaults
}

type AcceptedModifyRewardReserve struct {
	Vaults  map[common.Hash][]Vault `json:"Vaults"` // unifiedTokenID -> list vaults
	TxReqID common.Hash             `json:"TxReqID"`
}

func NewModifyRewardReserve() *ModifyRewardReserve {
	return &ModifyRewardReserve{}
}

func NewModifyRewardReserveWithValue(vaults map[common.Hash][]Vault) *ModifyRewardReserve {
	metadataBase := metadataCommon.NewMetadataBaseWithSignature(metadataCommon.BridgeAggModifyRewardReserveMeta)
	request := &ModifyRewardReserve{}
	request.MetadataBaseWithSignature = *metadataBase
	request.Vaults = vaults
	return request
}

func (request *ModifyRewardReserve) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (request *ModifyRewardReserve) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	// validate IncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(config.Param().BridgeAggParam.AdminAddress)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyRewardReserveValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}
	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyRewardReserveValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}

	if ok, err := request.MetadataBaseWithSignature.VerifyMetadataSignature(incAddr.Pk, tx); err != nil || !ok {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyRewardReserveValidateSanityDataError, errors.New("Sender is unauthorized"))
	}

	// check tx type and version
	if tx.GetType() != common.TxNormalType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyRewardReserveValidateSanityDataError, errors.New("Tx bridge agg modify list tokens must be TxNormalType"))
	}

	if tx.GetVersion() != 2 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyRewardReserveValidateSanityDataError, errors.New("Tx bridge agg modify list tokens must be version 2"))
	}

	usedTokenIDs := make(map[common.Hash]bool)
	if len(request.Vaults) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyRewardReserveValidateSanityDataError, errors.New("Length of unifiedTokens cannot be 0"))
	}
	for unifiedTokenID, vaults := range request.Vaults {
		if len(vaults) == 0 {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyRewardReserveValidateSanityDataError, errors.New("Length of vaults cannot be 0"))
		}
		if unifiedTokenID.IsZeroValue() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyRewardReserveValidateSanityDataError, fmt.Errorf("unifiedTokenID can not be empty"))
		}
		if usedTokenIDs[unifiedTokenID] {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyRewardReserveValidateSanityDataError, fmt.Errorf("Found duplicate tokenID %s", unifiedTokenID.String()))
		}
		usedTokenIDs[unifiedTokenID] = true
		for _, vault := range vaults {
			if vault.TokenID().IsZeroValue() {
				return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyRewardReserveValidateSanityDataError, fmt.Errorf("vault tokenID can not be empty"))
			}
			if usedTokenIDs[vault.TokenID()] {
				return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyRewardReserveValidateSanityDataError, fmt.Errorf("Found duplicate tokenID %s", vault.TokenID().String()))
			}
			usedTokenIDs[vault.TokenID()] = true
		}
	}

	return true, true, nil
}

func (request *ModifyRewardReserve) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.BridgeAggModifyRewardReserveMeta
}

func (request *ModifyRewardReserve) Hash() *common.Hash {
	record := request.MetadataBaseWithSignature.Hash().String()
	if request.Sig != nil && len(request.Sig) != 0 {
		record += string(request.Sig)
	}
	contentBytes, _ := json.Marshal(request)
	hashParams := common.HashH(contentBytes)
	record += hashParams.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (request *ModifyRewardReserve) HashWithoutSig() *common.Hash {
	record := request.MetadataBaseWithSignature.Hash().String()
	contentBytes, _ := json.Marshal(request.Vaults)
	hashParams := common.HashH(contentBytes)
	record += hashParams.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (request *ModifyRewardReserve) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *ModifyRewardReserve) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	content, err := metadataCommon.NewActionWithValue(request, *tx.Hash(), nil).StringSlice(metadataCommon.BridgeAggModifyRewardReserveMeta)
	return [][]string{content}, err
}
