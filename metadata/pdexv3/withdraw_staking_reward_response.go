package pdexv3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy/coin"
)

type WithdrawalStakingRewardResponse struct {
	metadataCommon.MetadataBase
	ReqTxID common.Hash `json:"ReqTxID"`
}

type DistributeStakingRewardContent struct {
	StakingPoolID string                 `json:"StakingPoolID"`
	Rewards       map[common.Hash]uint64 `json:"Rewards"`
}

func NewPdexv3WithdrawalStakingRewardResponse(
	metaType int,
	reqTxID common.Hash,
) *WithdrawalStakingRewardResponse {
	metadataBase := metadataCommon.NewMetadataBase(metaType)

	return &WithdrawalStakingRewardResponse{
		MetadataBase: *metadataBase,
		ReqTxID:      reqTxID,
	}
}

func (withdrawalResponse WithdrawalStakingRewardResponse) CheckTransactionFee(
	tx metadataCommon.Transaction, minFeePerKb uint64, minFeePerTx uint64, beaconHeight int64, db *statedb.StateDB,
) bool {
	// no need to have fee for this tx
	return true
}

func (withdrawalResponse WithdrawalStakingRewardResponse) ValidateTxWithBlockChain(
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

func (withdrawalResponse WithdrawalStakingRewardResponse) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	return false, true, nil
}

func (withdrawalResponse WithdrawalStakingRewardResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return withdrawalResponse.Type == metadataCommon.Pdexv3WithdrawStakingRewardResponseMeta
}

func (withdrawalResponse WithdrawalStakingRewardResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(withdrawalResponse)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (withdrawalResponse *WithdrawalStakingRewardResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(withdrawalResponse)
}

func (withdrawalResponse WithdrawalStakingRewardResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
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
		if len(inst) < 4 { // this is not WithdrawalStakingRewardRequest instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || (instMetaType != strconv.Itoa(metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta)) {
			continue
		}

		contentBytes := []byte(inst[3])
		var instContent WithdrawalStakingRewardContent
		err := json.Unmarshal(contentBytes, &instContent)
		if err != nil {
			continue
		}

		shardIDFromInst := instContent.ShardID
		txReqIDFromInst := instContent.TxReqID
		receiver := instContent.Receiver
		receiverAddress, err := isValidOTAReceiver(receiver.Address, shardIDFromInst)
		if err != nil {
			continue
		}

		if !bytes.Equal(withdrawalResponse.ReqTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}

		isMinted, mintCoin, assetID, err := tx.GetTxMintData()
		if err != nil {
			continue
		}
		if !isMinted {
			continue
		}
		pk := mintCoin.GetPublicKey().ToBytesS()
		paidAmount := mintCoin.GetValue()
		mintCoinV2, ok := mintCoin.(*coin.CoinV2)
		if !ok {
			metadataCommon.Logger.Log.Warnf("Mint coin is not CoinV2")
			continue
		}

		txR := mintCoinV2.GetTxRandom()
		publicKey := receiverAddress.PublicKey
		txRandom := receiverAddress.TxRandom

		if !bytes.Equal(publicKey.ToBytesS(), pk[:]) ||
			receiver.Amount != paidAmount ||
			!bytes.Equal(txR[:], txRandom[:]) ||
			instContent.TokenID != *assetID {
			continue
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("No WithdrawalStakingReward instruction found for WithdrawalStakingRewardResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}
