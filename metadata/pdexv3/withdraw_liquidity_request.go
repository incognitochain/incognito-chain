package pdexv3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/utils"
)

type WithdrawLiquidityRequest struct {
	metadataCommon.MetadataBase
	poolPairID   string
	nftID        string
	otaReceivers map[string]string
	shareAmount  uint64
}

func NewWithdrawLiquidityRequest() *WithdrawLiquidityRequest {
	return &WithdrawLiquidityRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3WithdrawLiquidityRequestMeta,
		},
	}
}

func NewWithdrawLiquidityRequestWithValue(
	poolPairID, nftID string,
	otaReceivers map[string]string,
	shareAmount uint64,
) *WithdrawLiquidityRequest {
	return &WithdrawLiquidityRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3WithdrawLiquidityRequestMeta,
		},
		poolPairID:   poolPairID,
		nftID:        nftID,
		otaReceivers: otaReceivers,
		shareAmount:  shareAmount,
	}
}

func (request *WithdrawLiquidityRequest) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconViewRetriever.GetHeight()) {
		return false, fmt.Errorf("Feature pdexv3 has not been activated yet")
	}
	pdexv3StateCached := chainRetriever.GetPdexv3Cached(beaconViewRetriever.BlockHash())
	err := beaconViewRetriever.IsValidNftID(chainRetriever.GetBeaconChainDatabase(), pdexv3StateCached, request.nftID)
	if err != nil {
		return false, err
	}
	err = beaconViewRetriever.IsValidPdexv3ShareAmount(chainRetriever.GetBeaconChainDatabase(), pdexv3StateCached, request.poolPairID, request.nftID, request.shareAmount)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (request *WithdrawLiquidityRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconHeight) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Feature pdexv3 has not been activated yet"))
	}
	if request.poolPairID == utils.EmptyString {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Pool pair id should not be empty"))
	}
	nftID, err := common.Hash{}.NewHashFromStr(request.nftID)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if nftID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("NftID should not be empty"))
	}
	if request.otaReceivers[request.nftID] == utils.EmptyString {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("NftID's ota receiver can not be empty"))
	}
	for tokenID, otaReceiverStr := range request.otaReceivers {
		_, err := common.Hash{}.NewHashFromStr(tokenID)
		if err != nil {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
		}
		otaReceiver := privacy.OTAReceiver{}
		err = otaReceiver.FromString(otaReceiverStr)
		if err != nil {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
		}
		if !otaReceiver.IsValid() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiveNft is not valid"))
		}
		if otaReceiver.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("otaReceiver shardID is different from txShardID"))
		}
	}
	if request.shareAmount == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("shareAmount can not be 0"))
	}

	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDENotBurningTxError, err)
	}
	if !bytes.Equal(burnedTokenID[:], nftID[:]) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Wrong request info's token id, it should be equal to tx's token id"))
	}
	if burnCoin.GetValue() != 1 {
		err := fmt.Errorf("Burnt amount is not valid expect %v but get %v", 1, burnCoin.GetValue())
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if tx.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("Tx type must be custom token privacy type"))
	}
	if nftID.String() == common.PRVCoinID.String() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidTxTypeError, errors.New("Invalid NftID"))
	}
	return true, true, nil
}

func (request *WithdrawLiquidityRequest) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.Pdexv3WithdrawLiquidityRequestMeta
}

func (request *WithdrawLiquidityRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&request)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *WithdrawLiquidityRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *WithdrawLiquidityRequest) ToCompactBytes() ([]byte, error) {
	return metadataCommon.ToCompactBytes(request)
}

func (request *WithdrawLiquidityRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PoolPairID   string            `json:"PoolPairID"`
		NftID        string            `json:"NftID"`
		OtaReceivers map[string]string `json:"OtaReceivers"`
		ShareAmount  uint64            `json:"ShareAmount"`
		metadataCommon.MetadataBase
	}{
		PoolPairID:   request.poolPairID,
		NftID:        request.nftID,
		OtaReceivers: request.otaReceivers,
		ShareAmount:  request.shareAmount,
		MetadataBase: request.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (request *WithdrawLiquidityRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID   string            `json:"PoolPairID"`
		NftID        string            `json:"NftID"`
		OtaReceivers map[string]string `json:"OtaReceivers"`
		ShareAmount  uint64            `json:"ShareAmount"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	request.poolPairID = temp.PoolPairID
	request.nftID = temp.NftID
	request.otaReceivers = temp.OtaReceivers
	request.shareAmount = temp.ShareAmount
	request.MetadataBase = temp.MetadataBase
	return nil
}

func (request *WithdrawLiquidityRequest) PoolPairID() string {
	return request.poolPairID
}

func (request *WithdrawLiquidityRequest) OtaReceivers() map[string]string {
	return request.otaReceivers
}

func (request *WithdrawLiquidityRequest) ShareAmount() uint64 {
	return request.shareAmount
}

func (request *WithdrawLiquidityRequest) NftID() string {
	return request.nftID
}

func (request *WithdrawLiquidityRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	for tokenID, val := range request.otaReceivers {
		tokenHash := common.PRVCoinID
		if tokenID != common.PRVIDStr {
			tokenHash = common.ConfidentialAssetID
		}
		otaReceiver := privacy.OTAReceiver{}
		otaReceiver.FromString(val)
		result = append(result, metadataCommon.OTADeclaration{
			PublicKey: otaReceiver.PublicKey.ToBytes(), TokenID: tokenHash,
		})
	}
	return result
}
