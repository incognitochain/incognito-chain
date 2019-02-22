package blockchain

import (
	"encoding/json"
	"strconv"

	"github.com/ninjadotorg/constant/metadata"
)

type ShardBlockSalaryUpdateInfo struct {
	ShardBlockSalary uint64
	ShardBlockFee    uint64
}

func getShardBlockFee(txs []metadata.Transaction) uint64 {
	totalFee := uint64(0)
	for _, tx := range txs {
		totalFee += tx.GetTxFee()
	}
	return totalFee
}

func getShardBlockSalary(txs []metadata.Transaction, bestStateBeacon *BestStateBeacon) uint64 {
	salaryPerTx := bestStateBeacon.StabilityInfo.GOVConstitution.GOVParams.SalaryPerTx
	basicSalary := bestStateBeacon.StabilityInfo.GOVConstitution.GOVParams.BasicSalary
	return uint64(len(txs))*salaryPerTx + basicSalary
}

func createShardBlockSalaryUpdateAction(
	shardBlockSalary uint64,
	shardBlockFee uint64,
) ([][]string, error) {
	shardBlockSalaryUpdateInfo := ShardBlockSalaryUpdateInfo{
		ShardBlockSalary: shardBlockSalary,
		ShardBlockFee:    shardBlockFee,
	}
	shardBlockSalaryUpdateInfoBytes, err := json.Marshal(shardBlockSalaryUpdateInfo)
	if err != nil {
		return [][]string{}, err
	}
	action := []string{strconv.Itoa(metadata.ShardBlockSalaryUpdateMeta), string(shardBlockSalaryUpdateInfoBytes)}
	return [][]string{action}, nil
}

func buildInstForShardBlockSalaryUpdate(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) ([][]string, error) {
	var shardBlockSalaryUpdateInfo ShardBlockSalaryUpdateInfo
	err := json.Unmarshal([]byte(contentStr), &shardBlockSalaryUpdateInfo)
	if err != nil {
		return nil, err
	}
	instructions := [][]string{}
	accumulativeValues.totalFee += shardBlockSalaryUpdateInfo.ShardBlockFee
	accumulativeValues.totalSalary += shardBlockSalaryUpdateInfo.ShardBlockSalary
	returnedInst := []string{
		strconv.Itoa(metadata.ShardBlockSalaryUpdateMeta),
		strconv.Itoa(int(shardID)),
		contentStr,
	}
	instructions = append(instructions, returnedInst)
	return instructions, nil
}

// type SalaryResContent struct {
// 	ShardBlockSalary uint64
// 	PayToAddress     *privacy.PaymentAddress
// 	ShardBlockHeight uint64
// }

// func buildInstructionsForSalaryReq(
// 	shardID byte,
// 	contentStr string,
// 	beaconBestState *BestStateBeacon,
// 	accumulativeValues *accumulativeValues,
// ) ([][]string, error) {
// 	var salaryReqInfo SalaryReqInfo
// 	err := json.Unmarshal([]byte(contentStr), &salaryReqInfo)
// 	if err != nil {
// 		return nil, err
// 	}
// 	instructions := [][]string{}
// 	stabilityInfo := beaconBestState.StabilityInfo
// 	govParams := stabilityInfo.GOVConstitution.GOVParams
// 	salaryPerTx := govParams.SalaryPerTx
// 	feePerKbTx := govParams.FeePerKbTx
// 	basicSalary := govParams.BasicSalary
// 	shardBlockSalary := basicSalary + salaryPerTx*salaryReqInfo.TotalTxs
// 	shardBlockFee := salaryReqInfo.TotalTxsSizeInKb * feePerKbTx
// 	accumulativeValues.totalFee += shardBlockFee
// 	accumulativeValues.totalSalary += shardBlockSalary
// 	salaryResContent := SalaryResContent{
// 		ShardBlockSalary: shardBlockSalary,
// 		PayToAddress:     salaryReqInfo.PayToAddress,
// 		ShardBlockHeight: salaryReqInfo.ShardBlockHeight,
// 	}
// 	salaryResContentBytes, err := json.Marshal(salaryResContent)
// 	if err != nil {
// 		return nil, err
// 	}
// 	returnedInst := []string{
// 		strconv.Itoa(metadata.SalaryResponseMeta),
// 		strconv.Itoa(int(shardID)),
// 		string(salaryResContentBytes),
// 	}
// 	instructions = append(instructions, returnedInst)
// 	return instructions, nil
// }

// func (blockgen *BlkTmplGenerator) buildSalaryRes(
// 	salaryResContentStr string,
// 	blkProducerPrivateKey *privacy.SpendingKey,
// ) ([]metadata.Transaction, error) {
// 	var salaryResContent SalaryResContent
// 	err := json.Unmarshal([]byte(salaryResContentStr), &salaryResContent)
// 	if err != nil {
// 		return nil, err
// 	}
// 	salaryResTx := new(transaction.Tx)
// 	err = salaryResTx.InitTxSalary(
// 		salaryResContent.ShardBlockSalary,
// 		salaryResContent.PayToAddress,
// 		blkProducerPrivateKey,
// 		blockgen.chain.GetDatabase(),
// 		nil,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return []metadata.Transaction{salaryResTx}, nil
// }
