package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

func (blockchain *BlockChain) GetStakingAmountShard() uint64 {
	return blockchain.config.ChainParams.StakingAmountShard
}

func (blockchain *BlockChain) GetDatabase() incdb.Database {
	return blockchain.config.DataBase
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

func (blockchain *BlockChain) GetShardIDFromTx(txid string) (byte, error) {
	txHash, err := (&common.Hash{}).NewHashFromStr(txid)
	if err != nil {
		return 0, NewBlockChainError(GetShardIDFromTxError, err)
	}
	shardID, _, _, _, err := blockchain.GetTransactionByHash(*txHash)
	if err != nil {
		return 0, NewBlockChainError(GetShardIDFromTxError, err)
	}
	return shardID, nil
}

func (blockchain *BlockChain) GetTxValue(txid string) (uint64, error) {
	txHash, err := (&common.Hash{}).NewHashFromStr(txid)
	if err != nil {
		return 0, NewBlockChainError(GetValueFromTxError, err)
	}
	_, _, _, tx, err := blockchain.GetTransactionByHash(*txHash)
	if err != nil {
		return 0, NewBlockChainError(GetShardIDFromTxError, err)
	}
	return tx.CalculateTxValue(), nil
}

func ListPubKeyFromListPayment(listPaymentAddresses []privacy.PaymentAddress) [][]byte {
	pubKeys := make([][]byte, 0)
	for _, i := range listPaymentAddresses {
		pubKeys = append(pubKeys, i.Pk)
	}
	return pubKeys
}

func (blockchain *BlockChain) GetAllCommitteeValidatorCandidate() (map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error) {
	SC := make(map[byte][]incognitokey.CommitteePublicKey)
	SPV := make(map[byte][]incognitokey.CommitteePublicKey)
	if blockchain.IsTest {
		return SC, SPV, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, nil
	}
	beaconBestState := BeaconBestState{}
	beaconBestStateBytes, err := rawdbv2.GetBeaconBestState(blockchain.GetDatabase())
	if err != nil {
		return SC, SPV, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, nil
	} else {
		if err := json.Unmarshal(beaconBestStateBytes, &beaconBestState); err != nil {
			Logger.log.Error(err)
			return SC, SPV, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, nil
		}
	}
	for shardID, committee := range beaconBestState.GetShardCommittee() {
		SC[shardID] = append([]incognitokey.CommitteePublicKey{}, committee...)
	}
	for shardID, pendingValidator := range beaconBestState.GetShardPendingValidator() {
		SPV[shardID] = append([]incognitokey.CommitteePublicKey{}, pendingValidator...)
	}
	BC := beaconBestState.BeaconCommittee
	BPV := beaconBestState.BeaconPendingValidator
	CBWFCR := beaconBestState.CandidateBeaconWaitingForCurrentRandom
	CBWFNR := beaconBestState.CandidateBeaconWaitingForNextRandom
	CSWFCR := beaconBestState.CandidateShardWaitingForCurrentRandom
	CSWFNR := beaconBestState.CandidateShardWaitingForNextRandom
	return SC, SPV, BC, BPV, CBWFCR, CBWFNR, CSWFCR, CSWFNR, nil
}

func (blockchain *BlockChain) GetAllCommitteeValidatorCandidateFlattenListFromDatabase() ([]string, error) {
	beaconBestState := BeaconBestState{}
	beaconBestStateBytes, err := rawdbv2.GetBeaconBestState(blockchain.GetDatabase())
	if err != nil {
		return nil, err
	} else {
		if err := json.Unmarshal(beaconBestStateBytes, &beaconBestState); err != nil {
			return nil, err
		}
	}
	res := []string{}
	for _, committee := range beaconBestState.GetShardCommittee() {
		committeeStr, err := incognitokey.CommitteeKeyListToString(committee)
		if err != nil {
			return nil, err
		}
		res = append(res, committeeStr...)
	}
	for _, pendingValidator := range beaconBestState.GetShardPendingValidator() {
		pendingValidatorStr, err := incognitokey.CommitteeKeyListToString(pendingValidator)
		if err != nil {
			return nil, err
		}
		res = append(res, pendingValidatorStr...)
	}
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconCommittee)
	if err != nil {
		return nil, err
	}
	res = append(res, beaconCommitteeStr...)

	beaconPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconPendingValidator)
	if err != nil {
		return nil, err
	}
	res = append(res, beaconPendingValidatorStr...)

	candidateBeaconWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.CandidateBeaconWaitingForCurrentRandom)
	if err != nil {
		return nil, err
	}
	res = append(res, candidateBeaconWaitingForCurrentRandomStr...)

	candidateBeaconWaitingForNextRandomStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.CandidateBeaconWaitingForNextRandom)
	if err != nil {
		return nil, err
	}
	res = append(res, candidateBeaconWaitingForNextRandomStr...)

	candidateShardWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.CandidateShardWaitingForCurrentRandom)
	if err != nil {
		return nil, err
	}
	res = append(res, candidateShardWaitingForCurrentRandomStr...)

	candidateShardWaitingForNextRandomStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.CandidateShardWaitingForNextRandom)
	if err != nil {
		return nil, err
	}
	res = append(res, candidateShardWaitingForNextRandomStr...)
	return res, nil
}

func (blockchain *BlockChain) GetStakingTx(shardID byte) map[string]string {
	return blockchain.BestState.Shard[shardID].GetStakingTx()
}
func (blockchain *BlockChain) GetAutoStakingList() map[string]bool {
	return blockchain.BestState.Beacon.GetAutoStakingList()
}

func (blockchain *BlockChain) GetCentralizedWebsitePaymentAddress() string {
	if blockchain.config.ChainParams.Net == Testnet {
		return blockchain.config.ChainParams.CentralizedWebsitePaymentAddress
	}
	if blockchain.config.ChainParams.Net == Mainnet {
		beaconHeight := blockchain.GetBeaconHeight()
		if beaconHeight >= 243500 {
			// use new address
			return "12S6jZ6sjJaqsuMJKS6jG7gvE9eHUXGWa2B2dNC7PwyEYJkL6cE53Uzk926HrQMEv2i2oBvKP2GDTC6tzU9dYSVH5X3w9P58VWqux4F"
		} else {
			// use original address
			return blockchain.config.ChainParams.CentralizedWebsitePaymentAddress
		}
	}
	return ""
}

func (blockchain *BlockChain) GetBeaconHeightBreakPointBurnAddr() uint64 {
	return blockchain.config.ChainParams.BeaconHeightBreakPointBurnAddr
}

func (blockchain *BlockChain) GetBurningAddress(beaconHeight uint64) string {
	breakPoint := blockchain.GetBeaconHeightBreakPointBurnAddr()
	if beaconHeight == 0 {
		beaconHeight = blockchain.GetBeaconHeight()
	}
	if beaconHeight <= breakPoint {
		return common.BurningAddress
	}

	return common.BurningAddress2
}
