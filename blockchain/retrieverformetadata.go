package blockchain

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

func (blockchain *BlockChain) GetStakingAmountShard() uint64 {
	return blockchain.config.ChainParams.StakingAmountShard
}

func (blockchain *BlockChain) GetDatabase() database.DatabaseInterface {
	return blockchain.config.DataBase
}

func (blockchain *BlockChain) GetShardIDFromTx(txid string) (byte, error) {
	var txHash = &common.Hash{}
	(&common.Hash{}).Decode(txHash, txid)

	blockHash, _, err := blockchain.config.DataBase.GetTransactionIndexById(*txHash)
	if err != nil {
		return 0, NewBlockChainError(UnExpectedError, err)
	}
	block, _, err1 := blockchain.GetShardBlockByHash(blockHash)
	if err1 != nil {
		return 0, NewBlockChainError(UnExpectedError, err1)
	}

	return block.Header.ShardID, nil
}

func (blockchain *BlockChain) GetTxValue(txid string) (uint64, error) {
	var txHash = &common.Hash{}
	(&common.Hash{}).Decode(txHash, txid)

	blockHash, index, err := blockchain.config.DataBase.GetTransactionIndexById(*txHash)
	if err != nil {
		return 0, NewBlockChainError(UnExpectedError, err)
	}
	block, _, err1 := blockchain.GetShardBlockByHash(blockHash)
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

func ListPubKeyFromListPayment(listPaymentAddresses []privacy.PaymentAddress) [][]byte {
	pubKeys := make([][]byte, 0)
	for _, i := range listPaymentAddresses {
		pubKeys = append(pubKeys, i.Pk)
	}
	return pubKeys
}

func (blockchain *BlockChain) GetAllCommitteeValidatorCandidate() (map[byte][]string, map[byte][]string, []string, []string, []string, []string, []string, []string, error) {
	SC := make(map[byte][]string)
	SPV := make(map[byte][]string)
	if blockchain.IsTest {
		return SC, SPV, []string{}, []string{}, []string{}, []string{}, []string{}, []string{}, nil
	}
	beaconBestState := BeaconBestState{}
	temp, err := blockchain.config.DataBase.FetchBeaconBestState()
	if err != nil {
		return SC, SPV, []string{}, []string{}, []string{}, []string{}, []string{}, []string{}, err
	} else {
		if err := json.Unmarshal(temp, &beaconBestState); err != nil {
			Logger.log.Error(err)
			return SC, SPV, []string{}, []string{}, []string{}, []string{}, []string{}, []string{}, err
		}
	}
	for shardID, committee := range beaconBestState.GetShardCommittee() {
		SC[shardID] = incognitokey.CommitteeKeyListToString(committee)
	}
	for shardID, committee := range beaconBestState.GetShardPendingValidator() {
		SPV[shardID] = incognitokey.CommitteeKeyListToString(committee)
	}
	BC := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconCommittee)
	BPV := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconPendingValidator)
	CBWFCR := incognitokey.CommitteeKeyListToString(beaconBestState.CandidateBeaconWaitingForCurrentRandom)
	CBWFNR := incognitokey.CommitteeKeyListToString(beaconBestState.CandidateBeaconWaitingForNextRandom)
	CSWFCR := incognitokey.CommitteeKeyListToString(beaconBestState.CandidateShardWaitingForCurrentRandom)
	CSWFNR := incognitokey.CommitteeKeyListToString(beaconBestState.CandidateShardWaitingForNextRandom)
	return SC, SPV, BC, BPV, CBWFCR, CBWFNR, CSWFCR, CSWFNR, nil
}

func (blockchain *BlockChain) GetAllCommitteeValidatorCandidateFlattenListFromDatabase() ([]string, error) {
	beaconBestState := BeaconBestState{}
	temp, err := blockchain.config.DataBase.FetchBeaconBestState()
	if err != nil {
		return nil, err
	} else {
		if err := json.Unmarshal(temp, &beaconBestState); err != nil {
			return nil, err
		}
	}
	res := []string{}
	for _, committee := range beaconBestState.GetShardCommittee() {
		res = append(res, incognitokey.CommitteeKeyListToString(committee)...)
	}
	for _, pendingValidator := range beaconBestState.GetShardPendingValidator() {
		res = append(res, incognitokey.CommitteeKeyListToString(pendingValidator)...)
	}
	res = append(res, incognitokey.CommitteeKeyListToString(beaconBestState.BeaconCommittee)...)
	res = append(res, incognitokey.CommitteeKeyListToString(beaconBestState.BeaconPendingValidator)...)
	res = append(res, incognitokey.CommitteeKeyListToString(beaconBestState.CandidateBeaconWaitingForCurrentRandom)...)
	res = append(res, incognitokey.CommitteeKeyListToString(beaconBestState.CandidateBeaconWaitingForNextRandom)...)
	res = append(res, incognitokey.CommitteeKeyListToString(beaconBestState.CandidateShardWaitingForCurrentRandom)...)
	res = append(res, incognitokey.CommitteeKeyListToString(beaconBestState.CandidateShardWaitingForNextRandom)...)
	return res, nil
}

func (blockchain *BlockChain) GetStakingTx(shardID byte) map[string]string {
	return blockchain.BestState.Shard[shardID].GetStakingTx()
}
func (blockchain *BlockChain) GetAutoStakingList() map[string]bool {
	return blockchain.BestState.Beacon.GetAutoStakingList()
}
