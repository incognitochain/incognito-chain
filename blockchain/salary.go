package blockchain

import (
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/pkg/errors"
)

type ShardBlockSalaryInfo struct {
	ShardBlockSalary uint64
	ShardBlockFee    uint64
	PayToAddress     *privacy.PaymentAddress
	ShardBlockHeight uint64
	InfoHash         *common.Hash
}

func getShardBlockFee(txs []metadata.Transaction) uint64 {
	totalFee := uint64(0)
	for _, tx := range txs {
		totalFee += tx.GetTxFee()
	}
	return totalFee
}

func getShardBlockSalary(txs []metadata.Transaction, bc *BlockChain, beaconHeight uint64) (uint64, error) {
	txLen := uint64(0)
	for _, tx := range txs {
		if !tx.IsSalaryTx() {
			txLen += 1
		}
	}

	stabilityInfo, err := getStabilityInfoByHeight(bc, beaconHeight)
	if err != nil {
		return 0, err
	}
	if stabilityInfo == nil {
		return 0, nil
	}
	salaryPerTx := stabilityInfo.GOVConstitution.GOVParams.SalaryPerTx
	basicSalary := stabilityInfo.GOVConstitution.GOVParams.BasicSalary
	return (txLen*salaryPerTx + basicSalary), nil
}

func hashShardBlockSalaryInfo(
	shardBlockSalary uint64,
	shardBlockFee uint64,
	payToAddress *privacy.PaymentAddress,
	shardBlockHeight uint64,
) *common.Hash {
	record := string(shardBlockSalary) + string(shardBlockFee) + string(shardBlockHeight)
	record += payToAddress.String()
	hash := common.HashH([]byte(record))
	return &hash
}

// Type Content
// Content: shardBlockSalaryInfo
func createShardBlockSalaryUpdateAction(
	shardBlockSalary uint64,
	shardBlockFee uint64,
	payToAddress *privacy.PaymentAddress,
	shardBlockHeight uint64,
) ([][]string, error) {
	infoHash := hashShardBlockSalaryInfo(shardBlockSalary, shardBlockFee, payToAddress, shardBlockHeight)
	shardBlockSalaryInfo := ShardBlockSalaryInfo{
		ShardBlockSalary: shardBlockSalary,
		ShardBlockFee:    shardBlockFee,
		PayToAddress:     payToAddress,
		ShardBlockHeight: shardBlockHeight,
		InfoHash:         infoHash,
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
	} else if shardBlockSalaryInfo.ShardBlockSalary == 0 {
		instType = "zeroSalaryAmount"
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
	blkProducerPrivateKey *privacy.PrivateKey,
) ([]metadata.Transaction, error) {
	if instType != "accepted" {
		return nil, nil
	}
	var shardBlockSalaryInfo ShardBlockSalaryInfo
	err := json.Unmarshal([]byte(contentStr), &shardBlockSalaryInfo)
	if err != nil {
		return nil, err
	}
	if shardBlockSalaryInfo.PayToAddress == nil || shardBlockSalaryInfo.InfoHash == nil {
		return nil, errors.Errorf("Can not Parse from contentStr")
	}
	salaryResMeta := metadata.NewShardBlockSalaryRes(
		shardBlockSalaryInfo.ShardBlockHeight,
		*shardBlockSalaryInfo.PayToAddress,
		*shardBlockSalaryInfo.InfoHash,
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

// func (bc *BlockChain) verifyShardBlockSalaryResTx(
// 	tx metadata.Transaction,
// 	insts [][]string,
// 	instUsed []int,
// 	shardID byte,
// ) error {
// 	meta, ok := tx.GetMetadata().(*metadata.ShardBlockSalaryRes)
// 	if !ok {
// 		return errors.Errorf("Could not parse ShardBlockSalaryRes metadata of tx %s", tx.Hash().String())
// 	}

// 	instIdx := -1
// 	var shardBlockSalaryInfo ShardBlockSalaryInfo
// 	for i, inst := range insts {
// 		if instUsed[i] > 0 {
// 			continue
// 		}
// 		if inst[0] != strconv.Itoa(metadata.ShardBlockSalaryRequestMeta) {
// 			continue
// 		}
// 		if inst[1] != strconv.Itoa(int(shardID)) {
// 			continue
// 		}
// 		if inst[2] != "accepted" {
// 			continue
// 		}
// 		contentStr := inst[3]
// 		err := json.Unmarshal([]byte(contentStr), &shardBlockSalaryInfo)
// 		if err != nil {
// 			return err
// 		}
// 		if !bytes.Equal(shardBlockSalaryInfo.InfoHash[:], meta.ShardBlockSalaryInfoHash[:]) {
// 			continue
// 		}
// 		instIdx = i
// 		instUsed[i] += 1
// 		break
// 	}
// 	if instIdx == -1 {
// 		return errors.Errorf("no instruction found for ShardBlockSalaryRes tx %s", tx.Hash().String())
// 	}
// 	if (!bytes.Equal(shardBlockSalaryInfo.PayToAddress.Pk[:], meta.ProducerAddress.Pk[:])) ||
// 		(!bytes.Equal(shardBlockSalaryInfo.PayToAddress.Tk[:], meta.ProducerAddress.Tk[:])) {
// 		return errors.Errorf("Producer address in ShardBlockSalaryRes tx %s is not matched to instruction's", tx.Hash().String())
// 	}
// 	if shardBlockSalaryInfo.ShardBlockHeight != meta.ShardBlockHeight {
// 		return errors.Errorf("ShardBlockHeight in ShardBlockSalaryRes tx %s is not matched to instruction's", tx.Hash().String())
// 	}
// 	if shardBlockSalaryInfo.ShardBlockSalary != tx.CalculateTxValue() {
// 		return errors.Errorf("Salary amount in ShardBlockSalaryRes tx %s is not matched to instruction's", tx.Hash().String())
// 	}
// 	return nil
// }
