package blockchain

import (
	"encoding/json"
	"strconv"

	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type ShardBlockSalaryInfo struct {
	ShardBlockSalary uint64
	ShardBlockFee    uint64
	PayToAddress     *privacy.PaymentAddress
	ShardBlockHeight uint64
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
	payToAddress *privacy.PaymentAddress,
	shardBlockHeight uint64,
) ([][]string, error) {
	shardBlockSalaryInfo := ShardBlockSalaryInfo{
		ShardBlockSalary: shardBlockSalary,
		ShardBlockFee:    shardBlockFee,
		PayToAddress:     payToAddress,
		ShardBlockHeight: shardBlockHeight,
	}
	shardBlockSalaryInfoBytes, err := json.Marshal(shardBlockSalaryInfo)
	if err != nil {
		return [][]string{}, err
	}
	action := []string{strconv.Itoa(metadata.ShardBlockSalaryRequestMeta), string(shardBlockSalaryInfoBytes)}
	return [][]string{action}, nil
}

func buildInstForShardBlockSalaryReq(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) ([][]string, error) {
	var shardBlockSalaryInfo ShardBlockSalaryInfo
	err := json.Unmarshal([]byte(contentStr), &shardBlockSalaryInfo)
	if err != nil {
		return nil, err
	}
	instructions := [][]string{}
	instType := string("")
	accumulativeValues.totalFee += shardBlockSalaryInfo.ShardBlockFee
	if !isGOVFundEnough(beaconBestState, accumulativeValues, shardBlockSalaryInfo.ShardBlockSalary) {
		instType = "fundNotEnough"
	} else {
		instType = "accepted"
		accumulativeValues.totalSalary += shardBlockSalaryInfo.ShardBlockSalary
	}
	returnedInst := []string{
		strconv.Itoa(metadata.ShardBlockSalaryRequestMeta),
		strconv.Itoa(int(shardID)),
		instType,
		contentStr,
	}
	instructions = append(instructions, returnedInst)
	return instructions, nil
}

func (blockgen *BlkTmplGenerator) buildSalaryRes(
	instType string,
	contentStr string,
	blkProducerPrivateKey *privacy.SpendingKey,
) ([]metadata.Transaction, error) {
	if instType == "fundNotEnough" {
		return nil, nil
	}
	var shardBlockSalaryInfo ShardBlockSalaryInfo
	err := json.Unmarshal([]byte(contentStr), &shardBlockSalaryInfo)
	if err != nil {
		return nil, err
	}
	salaryResMeta := metadata.NewShardBlockSalaryRes(
		shardBlockSalaryInfo.ShardBlockHeight,
		*shardBlockSalaryInfo.PayToAddress,
		metadata.ShardBlockSalaryResponseMeta,
	)
	salaryResTx := new(transaction.Tx)
	err = salaryResTx.InitTxSalary(
		shardBlockSalaryInfo.ShardBlockSalary,
		shardBlockSalaryInfo.PayToAddress,
		blkProducerPrivateKey,
		blockgen.chain.GetDatabase(),
		salaryResMeta,
	)
	if err != nil {
		return nil, err
	}
	return []metadata.Transaction{salaryResTx}, nil
}
