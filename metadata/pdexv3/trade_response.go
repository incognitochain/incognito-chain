package pdexv3

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// TradeStatus containns the info tracked by feature statedb, which is then displayed in RPC status queries.
// For refunded trade, all fields except Status are ignored
type TradeStatus struct {
	Status     int         `json:"Status"`
	BuyAmount  uint64      `json:"BuyAmount"`
	TokenToBuy common.Hash `json:"TokenToBuy"`
}

// TradeResponse is the metadata inside response tx for trade
type TradeResponse struct {
	Status      int         `json:"Status"`
	RequestTxID common.Hash `json:"RequestTxID"`
	metadataCommon.MetadataBase
}

// AcceptedTrade is added as Content for produced beacon Instructions after handling a trade successfully
type AcceptedTrade struct {
	Receiver     privacy.OTAReceiver      `json:"Receiver"`
	Amount       uint64                   `json:"Amount"`
	TradePath    []string                 `json:"TradePath"`
	TokenToBuy   common.Hash              `json:"TokenToBuy"`
	PairChanges  [][2]*big.Int            `json:"PairChanges"`
	RewardEarned []map[common.Hash]uint64 `json:"RewardEarned"`
	OrderChanges []map[string][2]*big.Int `json:"OrderChanges"`
}

func (md AcceptedTrade) GetType() int {
	return metadataCommon.Pdexv3TradeRequestMeta
}

func (md AcceptedTrade) GetStatus() int {
	return TradeAcceptedStatus
}

// RefundedTrade is added as Content for produced beacon instruction after failure to handle a trade
type RefundedTrade struct {
	Receiver privacy.OTAReceiver `json:"Receiver"`
	TokenID  common.Hash         `json:"TokenToSell"`
	Amount   uint64              `json:"Amount"`
}

func (md RefundedTrade) GetType() int {
	return metadataCommon.Pdexv3TradeRequestMeta
}

func (md RefundedTrade) GetStatus() int {
	return TradeRefundedStatus
}

func (res TradeResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFeePerKb uint64, minFeePerTx uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (res TradeResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *metadataCommon.MintData, shardID byte, tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, ac *metadataCommon.AccumulatedValues, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever) (bool, error) {
	// look for the instruction associated with this response
	matchedInstructionIndex := -1

	for i, inst := range mintData.Insts {
		// match common data from instruction before parsing accepted / refunded metadata
		// use layout from instruction.Action
		if mintData.InstsUsed[i] > 0 ||
			inst[0] != strconv.Itoa(metadataCommon.Pdexv3TradeRequestMeta) ||
			inst[1] != strconv.Itoa(res.Status) ||
			inst[2] != strconv.Itoa(int(shardID)) ||
			inst[3] != res.RequestTxID.String() {
			// upon any error, skip to next instruction
			continue
		}
		switch res.Status {
		case TradeAcceptedStatus:
			var mdHolder struct {
				Content AcceptedTrade
			}
			err := json.Unmarshal([]byte(inst[4]), &mdHolder)
			if err != nil {
				metadataCommon.Logger.Log.Warnf("Error matching instruction %s as accepted trade - %v", inst[4], err)
				continue
			}
			md := mdHolder.Content
			valid, msg := validMintForInstruction(md.Receiver, md.Amount, md.TokenToBuy, tx)
			if valid {
				matchedInstructionIndex = i
				break
			} else {
				metadataCommon.Logger.Log.Warnf(msg)
			}

		case TradeRefundedStatus:
			var mdHolder struct {
				Content RefundedTrade
			}
			err := json.Unmarshal([]byte(inst[4]), &mdHolder)
			if err != nil {
				metadataCommon.Logger.Log.Warnf("Error matching instruction %s as refunded trade - %v", inst[4], err)
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
			metadataCommon.Logger.Log.Warnf("Unrecognized trade status %v", res.Status)
		}
	}

	if matchedInstructionIndex == -1 {
		return false, fmt.Errorf("Instruction not found for Trade Response TX %s", tx.Hash().String())
	}
	mintData.InstsUsed[matchedInstructionIndex] = 1
	return true, nil
}

func validMintForInstruction(recv privacy.OTAReceiver, amount uint64, tokenID common.Hash, tx metadataCommon.Transaction) (bool, string) {
	minting, tempCoin, mintedTokenID, err := tx.GetTxMintData()
	if err != nil {
		return false, fmt.Sprintf("Error getting mint output - %v", err)
	}
	mintOutput, ok := tempCoin.(*privacy.CoinV2)
	if !minting || !ok {
		return false, fmt.Sprintf(
			"Unexpected mint output in response TX : minting = %v, version 2 = %v",
			minting, ok)
	}

	if recv.PublicKey.ToBytes() != mintOutput.GetPublicKey().ToBytes() ||
		amount != mintOutput.GetValue() ||
		recv.TxRandom != *mintOutput.GetTxRandom() ||
		tokenID != *mintedTokenID {
		return false, fmt.Sprintf("Mint output - instruction mismatch: [%v, %d, %v, %v] vs [%v, %d, %v, %v]",
			recv.PublicKey.ToBytes(), amount, recv.TxRandom, tokenID,
			mintOutput.GetPublicKey().ToBytes(), mintOutput.GetValue(), *mintOutput.GetTxRandom(), *mintedTokenID)
	}

	return true, ""
}

func (res TradeResponse) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (res TradeResponse) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	return true, true, nil
}

func (res TradeResponse) ValidateMetadataByItself() bool {
	return res.Type == metadataCommon.Pdexv3TradeResponseMeta
}

func (res TradeResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(res)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (res *TradeResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(res)
}
