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
	metadataCommon.MetadataBaseWithSignature
	NewListTokens map[common.Hash][]statedb.BridgeAggConvertedTokenState `json:"NewListTokens"` // unifiedTokenID -> list tokenID
}

type AcceptedModifyListToken struct {
	ModifyListToken
	TxReqID common.Hash `json:"TxReqID"`
}

func NewModifyListToken() *ModifyListToken {
	return &ModifyListToken{}
}

func NewModifyListTokenWithValue(
	newListTokens map[common.Hash][]statedb.BridgeAggConvertedTokenState,
) *ModifyListToken {
	metadataBase := metadataCommon.NewMetadataBaseWithSignature(metadataCommon.BridgeAggModifyListTokenMeta)
	request := &ModifyListToken{}
	request.MetadataBaseWithSignature = *metadataBase
	request.NewListTokens = newListTokens
	return request
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

	usedTokenIDs := make(map[common.Hash]bool)
	for unifiedTokenID, convertTokens := range request.NewListTokens {
		if usedTokenIDs[unifiedTokenID] {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyListTokenValidateSanityDataError, errors.New("Found duplicate tokenID"))
		}
		usedTokenIDs[unifiedTokenID] = true
		for _, convertedToken := range convertTokens {
			if usedTokenIDs[convertedToken.TokenID()] {
				return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyListTokenValidateSanityDataError, errors.New("Found duplicate tokenID"))
			}
			usedTokenIDs[convertedToken.TokenID()] = true
		}
	}

	return true, true, nil
}

func (request *ModifyListToken) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.BridgeAggModifyListTokenMeta
}

func (request *ModifyListToken) Hash() *common.Hash {
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

func (request *ModifyListToken) HashWithoutSig() *common.Hash {
	record := request.MetadataBaseWithSignature.Hash().String()
	contentBytes, _ := json.Marshal(request.NewListTokens)
	hashParams := common.HashH(contentBytes)
	record += hashParams.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (request *ModifyListToken) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *ModifyListToken) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	content, err := metadataCommon.NewActionWithValue(request, *tx.Hash(), shardID).StringSlice(metadataCommon.BridgeAggModifyListTokenMeta)
	return [][]string{content}, err
}
