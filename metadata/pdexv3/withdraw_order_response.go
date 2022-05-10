package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// WithdrawOrderStatus containns the info tracked by feature statedb, which is then displayed in RPC status queries.
// For refunded `add order` requests, all fields except Status are ignored
type WithdrawOrderStatus struct {
	Status         int         `json:"Status"`
	TokenID        common.Hash `json:"TokenID"`
	WithdrawAmount uint64      `json:"Amount"`
}

type WithdrawOrderResponse struct {
	Status      int         `json:"Status"`
	RequestTxID common.Hash `json:"RequestTxID"`
	metadataCommon.MetadataBase
}

type AcceptedWithdrawOrder struct {
	PoolPairID string              `json:"PoolPairID"`
	OrderID    string              `json:"OrderID"`
	TokenID    common.Hash         `json:"TokenID"`
	Receiver   privacy.OTAReceiver `json:"Receiver"`
	Amount     uint64              `json:"Amount"`
	AccessOTA  []byte              `json:"AccessOTA,omitempty"`
}

func (md AcceptedWithdrawOrder) GetType() int {
	return metadataCommon.Pdexv3WithdrawOrderRequestMeta
}

func (md AcceptedWithdrawOrder) GetStatus() int {
	return WithdrawOrderAcceptedStatus
}

type RejectedWithdrawOrder struct {
	PoolPairID string `json:"PoolPairID"`
	OrderID    string `json:"OrderID"`
	AccessOTA  []byte `json:"AccessOTA,omitempty"`
}

func (md RejectedWithdrawOrder) GetType() int {
	return metadataCommon.Pdexv3WithdrawOrderRequestMeta
}

func (md RejectedWithdrawOrder) GetStatus() int {
	return WithdrawOrderRejectedStatus
}

func (res WithdrawOrderResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (res WithdrawOrderResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *metadataCommon.MintData, shardID byte, tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, ac *metadataCommon.AccumulatedValues, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever) (bool, error) {
	// look for the instruction associated with this response
	matchedInstructionIndex := -1

	for i, inst := range mintData.Insts {
		// match common data from instruction before parsing accepted / refunded metadata
		// use layout from instruction.Action
		if mintData.InstsUsed[i] > 0 ||
			inst[0] != strconv.Itoa(metadataCommon.Pdexv3WithdrawOrderRequestMeta) ||
			inst[1] != strconv.Itoa(res.Status) ||
			inst[2] != strconv.Itoa(int(shardID)) ||
			inst[3] != res.RequestTxID.String() {
			// upon any error, skip to next instruction
			continue
		}
		switch res.Status {
		case WithdrawOrderAcceptedStatus:
			var mdHolder struct {
				Content AcceptedWithdrawOrder
			}
			err := json.Unmarshal([]byte(inst[4]), &mdHolder)
			if err != nil {
				metadataCommon.Logger.Log.Warnf("Error matching instruction %s as accepted withdrawOrder - %v", inst[4], err)
				continue
			}
			md := mdHolder.Content
			valid, msg := validMintForInstruction(md.Receiver, md.Amount, md.TokenID, tx)
			if valid {
				matchedInstructionIndex = i
				break
			} else {
				metadataCommon.Logger.Log.Warnf(msg)
			}
		default:
			metadataCommon.Logger.Log.Warnf("Unrecognized WithdrawOrder status %v for response", res.Status)
		}
	}

	if matchedInstructionIndex == -1 {
		return false, fmt.Errorf("Instruction not found for WithdrawOrder Response TX %s", tx.Hash().String())
	}
	mintData.InstsUsed[matchedInstructionIndex] = 1
	return true, nil
}

func (res WithdrawOrderResponse) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (res WithdrawOrderResponse) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	return true, true, nil
}

func (res WithdrawOrderResponse) ValidateMetadataByItself() bool {
	return res.Type == metadataCommon.Pdexv3WithdrawOrderResponseMeta
}

func (res WithdrawOrderResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(res)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (res *WithdrawOrderResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(res)
}
