package metadata

import (
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"

	// "errors"

	"github.com/constant-money/constant-chain/database"
)

type ShardBlockSalaryResponse struct {
	MetadataBase
	ShardBlockHeight         uint64
	ProducerAddress          *privacy.PaymentAddress
	ShardBlockSalaryInfoHash *common.Hash
}

type ShardBlockRewardInfo struct {
	ShardReward uint64
	Epoch       uint64
}

type ShardBlockRewardMeta struct {
	MetadataBase
}

type AcceptedBlockRewardInfo struct {
	ShardID          byte
	TxsFee           map[common.Hash]uint64
	ShardBlockHeight uint64
}

// func NewShardBlockSalaryRes(
// 	shardBlockHeight uint64,
// 	producerAddress privacy.PaymentAddress,
// 	shardBlockSalaryInfoHash common.Hash,
// 	metaType int,
// ) *ShardBlockSalaryRes {
// 	metadataBase := MetadataBase{
// 		Type: metaType,
// 	}
// 	return &ShardBlockSalaryRes{
// 		ShardBlockHeight:         shardBlockHeight,
// 		ProducerAddress:          producerAddress,
// 		ShardBlockSalaryInfoHash: shardBlockSalaryInfoHash,
// 		MetadataBase:             metadataBase,
// 	}
// }

func NewShardBlockRewardMeta() *ShardBlockRewardMeta {
	metadataBase := MetadataBase{
		Type: ShardBlockReward,
	}
	return &ShardBlockRewardMeta{
		MetadataBase: metadataBase,
	}
}

func BuildInstForShardReward(reward, epoch uint64, shardID byte) ([][]string, error) {
	resIns := [][]string{}
	shardBlockRewardInfo := ShardBlockRewardInfo{
		Epoch:       epoch,
		ShardReward: reward,
	}

	contentStr, err := json.Marshal(shardBlockRewardInfo)
	if err != nil {
		return nil, err
	}

	returnedInst := []string{
		strconv.Itoa(ShardBlockRewardRequestMeta),
		strconv.Itoa(int(shardID)),
		"shardRewardInst",
		string(contentStr),
	}
	resIns = append(resIns, returnedInst)
	return resIns, nil
}

func NewShardBlockRewardInfoFromString(inst string) (*ShardBlockRewardInfo, error) {
	Ins := &ShardBlockRewardInfo{}
	err := json.Unmarshal([]byte(inst), Ins)
	if err != nil {
		return nil, err
	}
	return Ins, nil
}

// func (sbsRes *ShardBlockSalaryRes) CheckTransactionFee(tr Transaction, minFee uint64) bool {
// 	// no need to have fee for this tx
// 	return true
// }

// func (sbsRes *ShardBlockSalaryRes) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
// 	// no need to validate tx with blockchain, just need to validate with request tx (via RequestedTxID) in current block
// 	return false, nil
// }

func (shardBlockRewardMeta *ShardBlockRewardMeta) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with request tx (via RequestedTxID) in current block
	return false, nil
}

// func (sbsRes *ShardBlockSalaryRes) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
// 	if len(sbsRes.ProducerAddress.Pk) == 0 {
// 		return false, false, errors.New("Wrong request info's producer address")
// 	}
// 	if len(sbsRes.ProducerAddress.Tk) == 0 {
// 		return false, false, errors.New("Wrong request info's producer address")
// 	}
// 	// if sbsRes.ShardBlockHeight == 0 {
// 	// 	return false, false, errors.New("Wrong request info's shard block height")
// 	// }
// 	return false, true, nil
// }

func (shardBlockRewardMeta *ShardBlockRewardMeta) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (shardBlockRewardMeta *ShardBlockRewardMeta) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	// The validation just need to check at tx level, so returning true here
	return true, true, nil
}

// func (sbsRes *ShardBlockSalaryRes) ValidateMetadataByItself() bool {
// 	// The validation just need to check at tx level, so returning true here
// 	return true
// }

// func (sbsRes *ShardBlockSalaryRes) VerifyMinerCreatedTxBeforeGettingInBlock(
// 	insts [][]string,
// 	instUsed []int,
// 	shardID byte,
// 	tx Transaction,
// 	bcr BlockchainRetriever,
// ) (bool, error) {
// 	instIdx := -1
// 	var shardBlockSalaryInfo ShardBlockSalaryInfo
// 	for i, inst := range insts {
// 		if instUsed[i] > 0 {
// 			continue
// 		}
// 		if inst[0] != strconv.Itoa(ShardBlockSalaryRequestMeta) {
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
// 			return false, err
// 		}
// 		if !bytes.Equal(shardBlockSalaryInfo.InfoHash[:], sbsRes.ShardBlockSalaryInfoHash[:]) {
// 			continue
// 		}
// 		instIdx = i
// 		instUsed[i] += 1
// 		break
// 	}
// 	if instIdx == -1 {
// 		return false, errors.Errorf("no instruction found for ShardBlockSalaryRes tx %s", tx.Hash().String())
// 	}
// 	if (!bytes.Equal(shardBlockSalaryInfo.PayToAddress.Pk[:], sbsRes.ProducerAddress.Pk[:])) ||
// 		(!bytes.Equal(shardBlockSalaryInfo.PayToAddress.Tk[:], sbsRes.ProducerAddress.Tk[:])) {
// 		return false, errors.Errorf("Producer address in ShardBlockSalaryRes tx %s is not matched to instruction's", tx.Hash().String())
// 	}
// 	if shardBlockSalaryInfo.ShardBlockHeight != sbsRes.ShardBlockHeight {
// 		return false, errors.Errorf("ShardBlockHeight in ShardBlockSalaryRes tx %s is not matched to instruction's", tx.Hash().String())
// 	}
// 	if shardBlockSalaryInfo.ShardBlockSalary != tx.CalculateTxValue() {
// 		return false, errors.Errorf("Salary amount in ShardBlockSalaryRes tx %s is not matched to instruction's", tx.Hash().String())
// 	}
// 	return true, nil
// }
// func (shardBlockSalaryRequest *ShardBlockSalaryRequest) GetStringFormat() ([]string, error) {
// 	content, err := json.Marshal(shardBlockSalaryRequest)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return []string{
// 		strconv.Itoa(ShardBlockSalaryRequestMeta),
// 		strconv.Itoa(BeaconOnly),
// 		string(content),
// 	}, nil
// }

func NewAcceptedBlockRewardInfo(
	shardID byte,
	txsFee map[common.Hash]uint64,
	shardBlockHeight uint64,
) *AcceptedBlockRewardInfo {
	return &AcceptedBlockRewardInfo{
		ShardID:          shardID,
		TxsFee:           txsFee,
		ShardBlockHeight: shardBlockHeight,
	}
}

func NewAcceptedBlockRewardInfoFromStr(
	inst string,
) (*AcceptedBlockRewardInfo, error) {
	Ins := &AcceptedBlockRewardInfo{}
	err := json.Unmarshal([]byte(inst), Ins)
	if err != nil {
		return nil, err
	}
	return Ins, nil
}

func (blockRewardInfo *AcceptedBlockRewardInfo) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(blockRewardInfo)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(AcceptedBlockRewardInfoMeta),
		strconv.Itoa(BeaconOnly),
		string(content),
	}, nil
}

// func NewShardBlockSalaryRequestFromStr(inst string) (*ShardBlockSalaryRequest, error) {
// 	Ins := &ShardBlockSalaryRequest{}
// 	err := json.Unmarshal([]byte(inst), Ins)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return Ins, nil
// }
