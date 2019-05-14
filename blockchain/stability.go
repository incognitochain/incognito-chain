package blockchain

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

type accumulativeValues struct {
	bondsSold            uint64
	govTokensSold        uint64
	incomeFromBonds      uint64
	incomeFromGOVTokens  uint64
	dcbTokensSoldByUSD   uint64
	dcbTokensSoldByETH   uint64
	constantsBurnedByETH uint64
	buyBackCoins         uint64
	totalFee             uint64
	totalSalary          uint64
	totalBeaconSalary    uint64
	totalShardSalary     uint64
	totalRefundAmt       uint64
	totalOracleRewards   uint64
	saleDataMap          map[string]*component.SaleData
}

func getStabilityInfoByHeight(blockchain *BlockChain, beaconHeight uint64) (*StabilityInfo, error) {
	stabilityInfoBytes, dbErr := blockchain.config.DataBase.FetchStabilityInfoByHeight(beaconHeight)
	if dbErr != nil {
		return nil, dbErr
	}
	if len(stabilityInfoBytes) == 0 { // not found
		return nil, nil
	}
	var stabilityInfo StabilityInfo
	unmarshalErr := json.Unmarshal(stabilityInfoBytes, &stabilityInfo)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return &stabilityInfo, nil
}

func isGOVFundEnough(
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
	expense uint64,
) bool {
	govFund := beaconBestState.StabilityInfo.SalaryFund
	income := accumulativeValues.incomeFromBonds + accumulativeValues.incomeFromGOVTokens + accumulativeValues.totalFee
	totalExpensed := accumulativeValues.buyBackCoins + accumulativeValues.totalSalary + accumulativeValues.totalRefundAmt + accumulativeValues.totalOracleRewards
	return govFund+income > expense+totalExpensed
}

// build actions from txs and ins at shard
func buildStabilityActions(
	txs []metadata.Transaction,
	bc *BlockChain,
	shardID byte,
	producerAddress *privacy.PaymentAddress,
	shardBlockHeight uint64,
	beaconBlocks []*BeaconBlock,
	beaconHeight uint64,
) ([][]string, error) {
	actions := [][]string{}
	for _, tx := range txs {
		meta := tx.GetMetadata()
		if meta != nil {
			actionPairs, err := meta.BuildReqActions(tx, bc, shardID)
			if err != nil {
				continue
			}
			actions = append(actions, actionPairs...)
		}
	}

	// build salary update action
	totalFee := getShardBlockFee(txs)
	totalSalary, err := getShardBlockSalary(txs, bc, beaconHeight)
	shardSalary := math.Ceil(float64(totalSalary) / 2)
	beaconSalary := math.Floor(float64(totalSalary) / 2)
	//fmt.Println("SA: fee&salary", totalFee, totalSalary, shardSalary, beaconSalary)
	if err != nil {
		return nil, err
	}

	if totalFee != 0 || totalSalary != 0 {
		salaryUpdateActions, _ := createShardBlockSalaryUpdateAction(uint64(beaconSalary), uint64(shardSalary), totalFee, producerAddress, shardBlockHeight)
		actions = append(actions, salaryUpdateActions...)
	}

	return actions, nil
}

// build instructions at beacon chain before syncing to shards
func (blockChain *BlockChain) buildStabilityInstructions(
	shardID byte,
	shardBlockInstructions [][]string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) ([][]string, error) {
	instructions := [][]string{}

	for _, inst := range shardBlockInstructions {
		if len(inst) == 0 {
			continue
		}
		// TODO: will improve the condition later
		if inst[0] == StakeAction || inst[0] == SwapAction || inst[0] == RandomAction {
			continue
		}
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			return [][]string{}, err
		}
		contentStr := inst[1]
		newInst := [][]string{}
		switch metaType {
		case metadata.BuyFromGOVRequestMeta:
			newInst, err = buildInstructionsForBuyBondsFromGOVReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.BuyGOVTokenRequestMeta:
			newInst, err = buildInstructionsForBuyGOVTokensReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.CrowdsaleRequestMeta:
			newInst, err = buildInstructionsForCrowdsaleRequest(shardID, contentStr, beaconBestState, accumulativeValues, blockChain)

		case metadata.TradeActivationMeta:
			newInst, err = buildInstructionsForTradeActivation(shardID, contentStr)

		case metadata.BuyBackRequestMeta:
			newInst, err = buildInstructionsForBuyBackBondsReq(shardID, contentStr, beaconBestState, accumulativeValues, blockChain)

		case metadata.IssuingRequestMeta:
			newInst, err = buildInstructionsForIssuingReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.ContractingRequestMeta:
			newInst, err = buildInstructionsForContractingReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.ShardBlockSalaryRequestMeta:
			newInst, err = buildInstForShardBlockSalaryReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.OracleFeedMeta:
			newInst, err = buildInstForOracleFeedReq(shardID, contentStr, beaconBestState)

		case metadata.UpdatingOracleBoardMeta:
			newInst, err = buildInstForUpdatingOracleBoardReq(shardID, contentStr, beaconBestState)

		default:
			continue
		}
		if err != nil {
			Logger.log.Error(err)
			continue
		}

		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
	}
	// update component in beststate

	return instructions, nil
}

func (blockgen *BlkTmplGenerator) buildStabilityResponseTxsFromInstructions(
	beaconBlocks []*BeaconBlock,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) ([]metadata.Transaction, error) {
	// TODO(@0xbunyip): refund bonds in multiple blocks since many refund instructions might come at once and UTXO picking order is not perfect
	unspentTokens := map[string]([]transaction.TxTokenVout){}
	tradeActivated := map[string]bool{}
	resTxs := []metadata.Transaction{}
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			// TODO: will improve the condition later
			var tx metadata.Transaction
			var err error
			txs := []metadata.Transaction{}

			if l[0] == SwapAction {
				fmt.Println("SA: swap instruction ", l, beaconBlock.Header.Height, blockgen.chain.BestState.Beacon.ShardCommittee)
				for _, v := range strings.Split(l[2], ",") {
					tx, err := blockgen.buildReturnStakingAmountTx(v, producerPrivateKey)
					if err != nil {
						Logger.log.Error("SA:", err)
						continue
					}
					resTxs = append(resTxs, tx)
				}

			}

			if l[0] == StakeAction || l[0] == RandomAction {
				continue
			}
			if len(l) <= 2 {
				continue
			}
		}
	}
	return resTxs, nil
}

func (blockgen *BlkTmplGenerator) buildStabilityResponseTxsAtShardOnly(txs []metadata.Transaction, producerPrivateKey *privacy.PrivateKey) ([]metadata.Transaction, error) {
	respTxs := []metadata.Transaction{}
	removeIds := []int{}
	multisigsRegTxs := []metadata.Transaction{}
	for i, tx := range txs {
		var respTx metadata.Transaction
		var err error

		switch tx.GetMetadataType() {
		case metadata.MultiSigsRegistrationMeta:
			multisigsRegTxs = append(multisigsRegTxs, tx)
		}

		if err != nil {
			// Remove this tx if cannot create corresponding response
			removeIds = append(removeIds, i)
		} else if respTx != nil {
			respTxs = append(respTxs, respTx)
		}
	}

	err := blockgen.registerMultiSigsAddresses(multisigsRegTxs)
	if err != nil {
		return nil, err
	}

	return respTxs, nil
}
