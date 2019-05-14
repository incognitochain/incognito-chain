package blockchain

import (
	"encoding/json"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	privacy "github.com/constant-money/constant-chain/privacy"
)

func (blockchain *BlockChain) GetDatabase() database.DatabaseInterface {
	return blockchain.config.DataBase
}

func (blockchain *BlockChain) GetShardIDFromTx(txid string) (byte, error) {
	var txHash = &common.Hash{}
	(&common.Hash{}).Decode(txHash, txid)

	blockHash, _, err := blockchain.config.DataBase.GetTransactionIndexById(txHash)
	if err != nil {
		return 0, NewBlockChainError(UnExpectedError, err)
	}
	block, err1, _ := blockchain.GetShardBlockByHash(blockHash)
	if err1 != nil {
		return 0, NewBlockChainError(UnExpectedError, err1)
	}

	return block.Header.ShardID, nil
}

func (blockchain *BlockChain) GetTxValue(txid string) (uint64, error) {
	var txHash = &common.Hash{}
	(&common.Hash{}).Decode(txHash, txid)

	blockHash, index, err := blockchain.config.DataBase.GetTransactionIndexById(txHash)
	if err != nil {
		return 0, NewBlockChainError(UnExpectedError, err)
	}
	block, err1, _ := blockchain.GetShardBlockByHash(blockHash)
	if err1 != nil {
		return 0, NewBlockChainError(UnExpectedError, err1)
	}
	txData := block.Body.Transactions[index]
	return txData.CalculateTxValue(), nil
}

func (blockchain *BlockChain) GetTxChainHeight(tx metadata.Transaction) (uint64, error) {
	shardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	return blockchain.GetChainHeight(shardID), nil
}

func (blockchain *BlockChain) GetChainHeight(shardID byte) uint64 {
	return blockchain.BestState.Shard[shardID].ShardHeight
}

func (blockchain *BlockChain) GetBeaconHeight() uint64 {
	return blockchain.BestState.Beacon.BeaconHeight
}

func (blockchain *BlockChain) GetBoardPubKeys(boardType common.BoardType) [][]byte {
	if boardType == common.DCBBoard {
		return blockchain.GetDCBBoardPubKeys()
	} else {
		return blockchain.GetGOVBoardPubKeys()
	}
}

func (blockchain *BlockChain) GetDCBBoardPubKeys() [][]byte {
	pubkeys := [][]byte{}
	for _, addr := range blockchain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress {
		pubkeys = append(pubkeys, addr.Pk[:])
	}
	return pubkeys
}

func (blockchain *BlockChain) GetGOVBoardPubKeys() [][]byte {
	pubkeys := [][]byte{}
	for _, addr := range blockchain.BestState.Beacon.StabilityInfo.GOVGovernor.BoardPaymentAddress {
		pubkeys = append(pubkeys, addr.Pk[:])
	}
	return pubkeys
}

func (blockchain *BlockChain) GetBoardPaymentAddress(boardType common.BoardType) []privacy.PaymentAddress {
	if boardType == common.DCBBoard {
		return blockchain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress
	}
	return blockchain.BestState.Beacon.StabilityInfo.GOVGovernor.BoardPaymentAddress
}

func ListPubKeyFromListPayment(listPaymentAddresses []privacy.PaymentAddress) [][]byte {
	pubKeys := make([][]byte, 0)
	for _, i := range listPaymentAddresses {
		pubKeys = append(pubKeys, i.Pk)
	}
	return pubKeys
}

func (blockchain *BlockChain) GetDCBParams() component.DCBParams {
	return blockchain.BestState.Beacon.StabilityInfo.DCBConstitution.DCBParams
}

func (blockchain *BlockChain) GetGOVParams() component.GOVParams {
	return blockchain.BestState.Beacon.StabilityInfo.GOVConstitution.GOVParams
}

//// Reserve
func (blockchain *BlockChain) GetAssetPrice(assetID *common.Hash) uint64 {
	return blockchain.BestState.Beacon.GetAssetPrice(*assetID)
}

func (blockchain *BlockChain) GetAllCommitteeValidatorCandidate() (map[byte][]string, map[byte][]string, []string, []string, []string, []string, []string, []string) {
	beaconBestState := BestStateBeacon{}
	temp, err := blockchain.config.DataBase.FetchBeaconBestState()
	if err != nil {
		panic("Can't Fetch Beacon BestState")
	} else {
		if err := json.Unmarshal(temp, &beaconBestState); err != nil {
			Logger.log.Error(err)
			panic("Fail to unmarshal Beacon BestState")
		}
	}
	SC := beaconBestState.ShardCommittee
	SPV := beaconBestState.ShardPendingValidator
	BC := beaconBestState.BeaconCommittee
	BPV := beaconBestState.BeaconPendingValidator
	CBWFCR := beaconBestState.CandidateBeaconWaitingForCurrentRandom
	CBWFNR := beaconBestState.CandidateBeaconWaitingForNextRandom
	CSWFCR := beaconBestState.CandidateShardWaitingForCurrentRandom
	CSWFNR := beaconBestState.CandidateShardWaitingForNextRandom
	return SC, SPV, BC, BPV, CBWFCR, CBWFNR, CSWFCR, CSWFNR
}
