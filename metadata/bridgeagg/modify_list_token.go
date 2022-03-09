package bridgeagg

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type ModifyListToken struct {
	NewListTokens map[common.Hash][]common.Hash `json:"NewListTokens"` // unifiedTokenID -> list tokenID
	metadataCommon.MetadataBaseWithSignature
}

type AcceptedModifyListToken struct {
	NewListTokens map[common.Hash][]common.Hash `json:"NewListTokens"` // unifiedTokenID -> list tokenID
	TxReqID       common.Hash                   `json:"TxReqID"`
}

func NewModifyListToken() *ModifyListToken {
	return &ModifyListToken{}
}

func NewModifyListTokenWithValue(
	newListTokens map[common.Hash][]common.Hash,
) *ModifyListToken {
	metadataBase := metadataCommon.NewMetadataBaseWithSignature(metadataCommon.BridgeAggModifyListTokenMeta)
	return &ModifyListToken{
		NewListTokens:             newListTokens,
		MetadataBaseWithSignature: *metadataBase,
	}
}

func (request *ModifyListToken) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (request *ModifyListToken) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	// validate IncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(config.Param().BridgeAggParam.AdminAddress)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyListTokenValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}
	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyListTokenValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}

	if ok, err := request.MetadataBaseWithSignature.VerifyMetadataSignature(incAddr.Pk, tx); err != nil || !ok {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyListTokenValidateSanityDataError, errors.New("Sender is unauthorized"))
	}

	// check tx type and version
	if tx.GetType() != common.TxNormalType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyListTokenValidateSanityDataError, errors.New("Tx bridge agg modify list tokens must be TxNormalType"))
	}

	if tx.GetVersion() != 2 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyListTokenValidateSanityDataError, errors.New("Tx bridge agg modify list tokens must be version 2"))
	}
	return true, true, nil
}

func (request *ModifyListToken) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.BridgeAggModifyListTokenMeta
}

func (request *ModifyListToken) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&request)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *ModifyListToken) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *ModifyListToken) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	content, err := metadataCommon.NewActionWithValue(request, *tx.Hash()).StringSlice(metadataCommon.BridgeAggModifyListTokenMeta)
	return [][]string{content}, err
}
