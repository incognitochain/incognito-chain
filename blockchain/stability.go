package blockchain

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
)

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
	return actions, nil
}

// build instructions at beacon chain before syncing to shards
func (blockChain *BlockChain) buildStabilityInstructions(
	shardID byte,
	shardBlockInstructions [][]string,
	beaconBestState *BestStateBeacon,
) ([][]string, error) {
	instructions := [][]string{}

	for _, inst := range shardBlockInstructions {
		if len(inst) == 0 {
			continue
		}
		if inst[0] == StakeAction || inst[0] == SwapAction || inst[0] == RandomAction {
			continue
		}
		// metaType, err := strconv.Atoi(inst[0])
		// if err != nil {
		// 	return [][]string{}, err
		// }
		newInst := [][]string{}
		// switch metaType {
		// // case metadata.IssuingRequestMeta:
		// // 	newInst, err = buildInstructionsForIssuingReq(shardID, contentStr, beaconBestState, accumulativeValues)

		// // case metadata.ContractingRequestMeta:
		// // 	newInst, err = buildInstructionsForContractingReq(shardID, contentStr, beaconBestState, accumulativeValues)

		// default:
		// 	continue
		// }
		// if err != nil {
		// 	Logger.log.Error(err)
		// 	continue
		// }
		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
	}
	return instructions, nil
}

func (blockgen *BlkTmplGenerator) buildResponseTxsFromBeaconInstructions(
	beaconBlocks []*BeaconBlock,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) ([]metadata.Transaction, error) {
	resTxs := []metadata.Transaction{}
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
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
			shardToProcess, err := strconv.Atoi(l[1])
			if err != nil {
				continue
			}
			if shardToProcess == int(shardID) {
				metaType, err := strconv.Atoi(l[0])
				if err != nil {
					return nil, err
				}
				// var newIns []string
				switch metaType {
				case metadata.BeaconSalaryRequestMeta:
					txs, err := blockgen.buildBeaconSalaryRes(l[0], l[3], producerPrivateKey)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, txs...)
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
