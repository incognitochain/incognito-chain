package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/wallet"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/pkg/errors"
)

type BeaconSalaryInfo struct {
	BeaconSalary      uint64
	PayToAddress      *privacy.PaymentAddress
	BeaconBlockHeight uint64
	InfoHash          *common.Hash
}

func (s *BeaconSalaryInfo) hash() *common.Hash {
	record := string(s.BeaconSalary) + string(s.BeaconBlockHeight)
	record += s.PayToAddress.String()
	hash := common.HashH([]byte(record))
	return &hash
}

type ShardBlockSalaryInfo struct {
	BeaconBlockSalary uint64
	ShardBlockSalary  uint64
	ShardBlockFee     uint64
	PayToAddress      *privacy.PaymentAddress
	ShardBlockHeight  uint64
	InfoHash          *common.Hash
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
	beaconBlockSalary uint64,
	shardBlockSalary uint64,
	shardBlockFee uint64,
	payToAddress *privacy.PaymentAddress,
	shardBlockHeight uint64,
) *common.Hash {
	record := string(shardBlockSalary) + string(shardBlockFee) + string(shardBlockHeight) + string(beaconBlockSalary)
	record += payToAddress.String()
	hash := common.HashH([]byte(record))
	return &hash
}

// Type Content
// Content: shardBlockSalaryInfo
func createShardBlockSalaryUpdateAction(
	beaconBlockSalary uint64,
	shardBlockSalary uint64,
	shardBlockFee uint64,
	payToAddress *privacy.PaymentAddress,
	shardBlockHeight uint64,
) ([][]string, error) {
	infoHash := hashShardBlockSalaryInfo(beaconBlockSalary, shardBlockSalary, shardBlockFee, payToAddress, shardBlockHeight)
	shardBlockSalaryInfo := ShardBlockSalaryInfo{
		BeaconBlockSalary: beaconBlockSalary,
		ShardBlockSalary:  shardBlockSalary,
		ShardBlockFee:     shardBlockFee,
		PayToAddress:      payToAddress,
		ShardBlockHeight:  shardBlockHeight,
		InfoHash:          infoHash,
	}
	shardBlockSalaryInfoBytes, err := json.Marshal(shardBlockSalaryInfo)
	if err != nil {
		return [][]string{}, err
	}
	action := []string{strconv.Itoa(metadata.ShardBlockSalaryRequestMeta), string(shardBlockSalaryInfoBytes)}
	return [][]string{action}, nil
}

func buildInstForBeaconSalary(salary, beaconHeight uint64, payToAddress *privacy.PaymentAddress) ([]string, error) {

	beaconSalaryInfo := BeaconSalaryInfo{
		BeaconBlockHeight: beaconHeight,
		PayToAddress:      payToAddress,
		BeaconSalary:      salary,
	}
	//fmt.Println("SA: beaconSalaryInfo", beaconSalaryInfo)
	beaconSalaryInfo.InfoHash = beaconSalaryInfo.hash()

	contentStr, err := json.Marshal(beaconSalaryInfo)
	if err != nil {
		return nil, err
	}

	returnedInst := []string{
		strconv.Itoa(metadata.BeaconSalaryRequestMeta),
		strconv.Itoa(int(common.GetShardIDFromLastByte(payToAddress.Bytes()[len(payToAddress.Bytes())-1]))),
		"beaconSalaryInst",
		string(contentStr),
	}

	return returnedInst, nil
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

	if !isGOVFundEnough(beaconBestState, accumulativeValues, shardBlockSalaryInfo.ShardBlockSalary+shardBlockSalaryInfo.BeaconBlockSalary) {
		instType = "fundNotEnough"
	} else if shardBlockSalaryInfo.ShardBlockSalary == 0 {
		instType = "zeroSalaryAmount"
	} else {
		instType = "accepted"
		accumulativeValues.totalSalary += shardBlockSalaryInfo.ShardBlockSalary
		accumulativeValues.totalSalary += shardBlockSalaryInfo.BeaconBlockSalary
		accumulativeValues.totalBeaconSalary += shardBlockSalaryInfo.BeaconBlockSalary
		accumulativeValues.totalShardSalary += shardBlockSalaryInfo.ShardBlockSalary
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

func (blockgen *BlkTmplGenerator) buildBeaconSalaryRes(
	instType string,
	contentStr string,
	blkProducerPrivateKey *privacy.PrivateKey,
) ([]metadata.Transaction, error) {
	//TODO: create tx for return beacon salary
	var beaconSalaryInfo BeaconSalaryInfo
	err := json.Unmarshal([]byte(contentStr), &beaconSalaryInfo)
	if err != nil {
		return nil, err
	}
	if beaconSalaryInfo.PayToAddress == nil || beaconSalaryInfo.InfoHash == nil {
		return nil, errors.Errorf("Can not Parse from contentStr")
	}

	salaryResMeta := metadata.NewBeaconBlockSalaryRes(
		beaconSalaryInfo.BeaconBlockHeight,
		*beaconSalaryInfo.PayToAddress,
		*beaconSalaryInfo.InfoHash,
		metadata.BeaconSalaryResponseMeta,
	)

	salaryResTx := new(transaction.Tx)
	err = salaryResTx.InitTxSalary(
		beaconSalaryInfo.BeaconSalary,
		beaconSalaryInfo.PayToAddress,
		blkProducerPrivateKey,
		blockgen.chain.GetDatabase(),
		salaryResMeta,
	)
	//fmt.Println("SA: beacon salary", beaconSalaryInfo, salaryResTx.CalculateTxValue(), salaryResTx.Hash().String())
	if err != nil {
		return nil, err
	}
	return []metadata.Transaction{salaryResTx}, nil
}

func (blockgen *BlkTmplGenerator) buildReturnStakingAmountTx(
	swaperPubKey string,
	blkProducerPrivateKey *privacy.PrivateKey,
) (metadata.Transaction, error) {
	addressBytes := blockgen.chain.config.UserKeySet.PaymentAddress.Pk
	shardID := common.GetShardIDFromLastByte(addressBytes[len(addressBytes)-1])
	_, committeeShardID := blockgen.chain.BestState.Beacon.GetPubkeyRole(base58.Base58Check{}.Encode(addressBytes, 0x00), 0)

	fmt.Println("SA: get tx for ", swaperPubKey, blockgen.chain.BestState.Shard[committeeShardID].StakingTx, committeeShardID)
	tx, ok := blockgen.chain.BestState.Shard[shardID].StakingTx[swaperPubKey]
	if !ok {
		return nil, NewBlockChainError(UnExpectedError, errors.New("No staking tx in best state"))
	}

	var txHash = &common.Hash{}
	(&common.Hash{}).Decode(txHash, tx)

	blockHash, index, err := blockgen.chain.config.DataBase.GetTransactionIndexById(txHash)
	if err != nil {
		abc := NewBlockChainError(UnExpectedError, err)
		Logger.log.Error(abc)
		return nil, abc
	}
	block, err1, _ := blockgen.chain.GetShardBlockByHash(blockHash)
	if err1 != nil {
		Logger.log.Errorf("ERROR", err1, "NO Transaction in block with hash &+v", blockHash, "and index", index, "contains", block.Body.Transactions[index])
		return nil, NewBlockChainError(UnExpectedError, err1)
	}

	txData := block.Body.Transactions[index]

	keyWallet, _ := wallet.Base58CheckDeserialize(txData.GetMetadata().(*metadata.StakingMetadata).PaymentAddress)
	paymentShardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1])

	fmt.Println("SA: build salary tx", txData.GetMetadata().(*metadata.StakingMetadata).PaymentAddress, paymentShardID, committeeShardID)

	if paymentShardID != committeeShardID {
		return nil, NewBlockChainError(UnExpectedError, errors.New("Not from this shard"))
	}

	returnStakingMeta := metadata.NewReturnStaking(
		tx,
		keyWallet.KeySet.PaymentAddress,
		metadata.ReturnStakingMeta,
	)

	returnStakingTx := new(transaction.Tx)
	err1 = returnStakingTx.InitTxSalary(
		txData.CalculateTxValue(),
		&keyWallet.KeySet.PaymentAddress,
		blkProducerPrivateKey,
		blockgen.chain.GetDatabase(),
		returnStakingMeta,
	)
	//modify the type of the salary transaction
	returnStakingTx.Type = common.TxReturnStakingType

	if err1 != nil {
		return nil, err1
	}
	return returnStakingTx, nil

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
	beaconSalaryResMeta := metadata.NewShardBlockSalaryRes(
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
		beaconSalaryResMeta,
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
