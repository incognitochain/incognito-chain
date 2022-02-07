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

// AddOrderStatus containns the info tracked by feature statedb, which is then displayed in RPC status queries.
// For refunded `add order` requests, all fields except Status are ignored
type AddOrderStatus struct {
	Status  int    `json:"Status"`
	OrderID string `json:"OrderID"`
}

// AddOrderResponse is the metadata inside response tx for `add order` (applicable for refunded case only)
type AddOrderResponse struct {
	Status      int         `json:"Status"`
	RequestTxID common.Hash `json:"RequestTxID"`
	metadataCommon.MetadataBase
}

// AcceptedAddOrder is added as Content for produced beacon instruction after to handling an order successfully
type AcceptedAddOrder struct {
	PoolPairID     string                              `json:"PoolPairID"`
	OrderID        string                              `json:"OrderID"`
	NftID          *common.Hash                        `json:"NftID,omitempty"`
	AccessOTA      []byte                              `json:"AccessOTA,omitempty"`
	Token0Rate     uint64                              `json:"Token0Rate"`
	Token1Rate     uint64                              `json:"Token1Rate"`
	Token0Balance  uint64                              `json:"Token0Balance"`
	Token1Balance  uint64                              `json:"Token1Balance"`
	TradeDirection byte                                `json:"TradeDirection"`
	Receiver       [2]string                           `json:"Receiver"`
	RewardReceiver map[common.Hash]privacy.OTAReceiver `json:"RewardReceiver,omitempty"`
}

func (md AcceptedAddOrder) GetType() int {
	return metadataCommon.Pdexv3AddOrderRequestMeta
}

func (md AcceptedAddOrder) GetStatus() int {
	return OrderAcceptedStatus
}

// RefundedAddOrder is added as Content for produced beacon instruction after failure to handle an order
type RefundedAddOrder struct {
	Receiver privacy.OTAReceiver `json:"Receiver"`
	TokenID  common.Hash         `json:"TokenID"`
	Amount   uint64              `json:"Amount"`
}

func (md RefundedAddOrder) GetType() int {
	return metadataCommon.Pdexv3AddOrderRequestMeta
}

func (md RefundedAddOrder) GetStatus() int {
	return OrderRefundedStatus
}

func (res AddOrderResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (res AddOrderResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *metadataCommon.MintData, shardID byte, tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, ac *metadataCommon.AccumulatedValues, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever) (bool, error) {
	// look for the instruction associated with this response
	matchedInstructionIndex := -1

	for i, inst := range mintData.Insts {
		// match common data from instruction before parsing accepted / refunded metadata
		// use layout from instruction.Action
		if mintData.InstsUsed[i] > 0 ||
			inst[0] != strconv.Itoa(metadataCommon.Pdexv3AddOrderRequestMeta) ||
			inst[1] != strconv.Itoa(res.Status) ||
			inst[2] != strconv.Itoa(int(shardID)) ||
			inst[3] != res.RequestTxID.String() {
			// upon any error, skip to next instruction
			continue
		}
		switch res.Status {
		case OrderRefundedStatus:
			var mdHolder struct {
				Content RefundedAddOrder
			}
			err := json.Unmarshal([]byte(inst[4]), &mdHolder)
			if err != nil {
				metadataCommon.Logger.Log.Warnf("Error matching instruction %s as refunded order - %v", inst[4], err)
				continue
			}
			md := &mdHolder.Content
			valid, msg := validMintForInstruction(md.Receiver, md.Amount, md.TokenID, tx)
			if valid {
				matchedInstructionIndex = i
				break
			} else {
				metadataCommon.Logger.Log.Warnf(msg)
			}
		default:
			metadataCommon.Logger.Log.Warnf("Unrecognized AddOrder status %v for response", res.Status)
		}
	}

	if matchedInstructionIndex == -1 {
		return false, fmt.Errorf("Instruction not found for AddOrder Response TX %s", tx.Hash().String())
	}
	mintData.InstsUsed[matchedInstructionIndex] = 1
	return true, nil
}

func (res AddOrderResponse) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (res AddOrderResponse) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	return true, true, nil
}

func (res AddOrderResponse) ValidateMetadataByItself() bool {
	return res.Type == metadataCommon.Pdexv3AddOrderResponseMeta
}

func (res AddOrderResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(res)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (res *AddOrderResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(res)
}
