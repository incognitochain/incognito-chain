package blockchain

import (
	"encoding/json"
	"strconv"

	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type SalaryReqInfo struct {
	PayToAddress     *privacy.PaymentAddress
	ShardID          byte
	ShardBlockHeight uint64
	TotalTxsSizeInKb uint64
	TotalTxs         uint64
}

func createSalaryReqAction(
	payToAddress *privacy.PaymentAddress,
	shardID byte,
	shardBlockHeight uint64,
	totalTxsSizeInKb uint64,
	totalTxs uint64,
) ([][]string, error) {
	salaryReqInfo := SalaryReqInfo{
		PayToAddress:     payToAddress,
		ShardID:          shardID,
		ShardBlockHeight: shardBlockHeight,
		TotalTxsSizeInKb: totalTxsSizeInKb,
		TotalTxs:         totalTxs,
	}
	salaryReqInfoBytes, err := json.Marshal(salaryReqInfo)
	if err != nil {
		return [][]string{}, err
	}
	action := []string{strconv.Itoa(metadata.SalaryRequestMeta), string(salaryReqInfoBytes)}
	return [][]string{action}, nil
}

type SalaryResContent struct {
	ShardBlockSalary uint64
	PayToAddress     *privacy.PaymentAddress
	ShardBlockHeight uint64
}

func buildInstructionsForSalaryReq(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) ([][]string, error) {
	var salaryReqInfo SalaryReqInfo
	err := json.Unmarshal([]byte(contentStr), &salaryReqInfo)
	if err != nil {
		return nil, err
	}
	instructions := [][]string{}
	stabilityInfo := beaconBestState.StabilityInfo
	govParams := stabilityInfo.GOVConstitution.GOVParams
	salaryPerTx := govParams.SalaryPerTx
	feePerKbTx := govParams.FeePerKbTx
	basicSalary := govParams.BasicSalary
	shardBlockSalary := basicSalary + salaryPerTx*salaryReqInfo.TotalTxs
	shardBlockFee := salaryReqInfo.TotalTxsSizeInKb * feePerKbTx
	accumulativeValues.totalFee += shardBlockFee
	accumulativeValues.totalSalary += shardBlockSalary
	salaryResContent := SalaryResContent{
		ShardBlockSalary: shardBlockSalary,
		PayToAddress:     salaryReqInfo.PayToAddress,
		ShardBlockHeight: salaryReqInfo.ShardBlockHeight,
	}
	salaryResContentBytes, err := json.Marshal(salaryResContent)
	if err != nil {
		return nil, err
	}
	returnedInst := []string{
		strconv.Itoa(metadata.SalaryResponseMeta),
		strconv.Itoa(int(shardID)),
		string(salaryResContentBytes),
	}
	instructions = append(instructions, returnedInst)
	return instructions, nil
}

func (blockgen *BlkTmplGenerator) buildSalaryRes(
	salaryResContentStr string,
	blkProducerPrivateKey *privacy.SpendingKey,
) ([]metadata.Transaction, error) {
	var salaryResContent SalaryResContent
	err := json.Unmarshal([]byte(salaryResContentStr), &salaryResContent)
	if err != nil {
		return nil, err
	}
	salaryResTx := new(transaction.Tx)
	err = salaryResTx.InitTxSalary(
		salaryResContent.ShardBlockSalary,
		salaryResContent.PayToAddress,
		blkProducerPrivateKey,
		blockgen.chain.GetDatabase(),
		nil,
	)
	if err != nil {
		return nil, err
	}
	return []metadata.Transaction{salaryResTx}, nil
}
