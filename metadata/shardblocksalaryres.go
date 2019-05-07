package metadata

import (
	"bytes"
	"encoding/json"

	// "errors"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/pkg/errors"
)

type ShardBlockSalaryRes struct {
	MetadataBase
	ShardBlockHeight         uint64
	ProducerAddress          privacy.PaymentAddress
	ShardBlockSalaryInfoHash common.Hash
}

type ShardBlockSalaryInfo struct {
	ShardBlockSalary uint64
	ShardBlockFee    uint64
	PayToAddress     *privacy.PaymentAddress
	ShardBlockHeight uint64
	InfoHash         *common.Hash
}

func NewShardBlockSalaryRes(
	shardBlockHeight uint64,
	producerAddress privacy.PaymentAddress,
	shardBlockSalaryInfoHash common.Hash,
	metaType int,
) *ShardBlockSalaryRes {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &ShardBlockSalaryRes{
		ShardBlockHeight:         shardBlockHeight,
		ProducerAddress:          producerAddress,
		ShardBlockSalaryInfoHash: shardBlockSalaryInfoHash,
		MetadataBase:             metadataBase,
	}
}

func (sbsRes *ShardBlockSalaryRes) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return true
}

func (sbsRes *ShardBlockSalaryRes) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with request tx (via RequestedTxID) in current block
	return false, nil
}

func (sbsRes *ShardBlockSalaryRes) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(sbsRes.ProducerAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's producer address")
	}
	if len(sbsRes.ProducerAddress.Tk) == 0 {
		return false, false, errors.New("Wrong request info's producer address")
	}
	// if sbsRes.ShardBlockHeight == 0 {
	// 	return false, false, errors.New("Wrong request info's shard block height")
	// }
	return false, true, nil
}

func (sbsRes *ShardBlockSalaryRes) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (sbsRes *ShardBlockSalaryRes) Hash() *common.Hash {
	record := sbsRes.ProducerAddress.String()
	record += string(sbsRes.ShardBlockHeight)
	record += sbsRes.ShardBlockSalaryInfoHash.String()

	// final hash
	record += sbsRes.MetadataBase.Hash().String()
	hash := common.HashH([]byte(record))
	return &hash
}

func (sbsRes *ShardBlockSalaryRes) VerifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	accumulatedData *component.UsedInstData,
) (bool, error) {
	instIdx := -1
	var shardBlockSalaryInfo ShardBlockSalaryInfo
	for i, inst := range insts {
		if instUsed[i] > 0 {
			continue
		}
		if inst[0] != strconv.Itoa(ShardBlockSalaryRequestMeta) {
			continue
		}
		if inst[1] != strconv.Itoa(int(shardID)) {
			continue
		}
		if inst[2] != "accepted" {
			continue
		}
		contentStr := inst[3]
		err := json.Unmarshal([]byte(contentStr), &shardBlockSalaryInfo)
		if err != nil {
			return false, err
		}
		if !bytes.Equal(shardBlockSalaryInfo.InfoHash[:], sbsRes.ShardBlockSalaryInfoHash[:]) {
			continue
		}
		instIdx = i
		instUsed[i] += 1
		break
	}
	if instIdx == -1 {
		return false, errors.Errorf("no instruction found for ShardBlockSalaryRes tx %s", tx.Hash().String())
	}
	if (!bytes.Equal(shardBlockSalaryInfo.PayToAddress.Pk[:], sbsRes.ProducerAddress.Pk[:])) ||
		(!bytes.Equal(shardBlockSalaryInfo.PayToAddress.Tk[:], sbsRes.ProducerAddress.Tk[:])) {
		return false, errors.Errorf("Producer address in ShardBlockSalaryRes tx %s is not matched to instruction's", tx.Hash().String())
	}
	if shardBlockSalaryInfo.ShardBlockHeight != sbsRes.ShardBlockHeight {
		return false, errors.Errorf("ShardBlockHeight in ShardBlockSalaryRes tx %s is not matched to instruction's", tx.Hash().String())
	}
	if shardBlockSalaryInfo.ShardBlockSalary != tx.CalculateTxValue() {
		return false, errors.Errorf("Salary amount in ShardBlockSalaryRes tx %s is not matched to instruction's", tx.Hash().String())
	}
	return true, nil
}
