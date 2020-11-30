package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
)

//RestoreBeaconCommittee ...
func (beaconBestState *BeaconBestState) RestoreBeaconCommittee() error {

	committeePublicKey := statedb.GetBeaconCommittee(beaconBestState.consensusStateDB)
	beaconBestState.BeaconCommittee = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		beaconBestState.BeaconCommittee[i] = v
	}

	return nil
}

//RestoreShardCommittee ...
func (beaconBestState *BeaconBestState) RestoreShardCommittee() error {

	beaconBestState.ShardCommittee = make(map[byte][]incognitokey.CommitteePublicKey)
	for i := 0; i < beaconBestState.ActiveShards; i++ {
		committeePublicKey := statedb.GetOneShardCommittee(beaconBestState.consensusStateDB, byte(i))

		beaconBestState.ShardCommittee[byte(i)] = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
		for index, value := range committeePublicKey {
			beaconBestState.ShardCommittee[byte(i)][index] = value
		}
	}

	return nil
}

//RestoreBeaconPendingValidator ...
func (beaconBestState *BeaconBestState) RestoreBeaconPendingValidator() error {

	committeePublicKey := statedb.GetBeaconSubstituteValidator(beaconBestState.consensusStateDB)
	beaconBestState.BeaconPendingValidator = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		beaconBestState.BeaconPendingValidator[i] = v
	}
	return nil
}

//RestoreShardPendingValidator ...
func (beaconBestState *BeaconBestState) RestoreShardPendingValidator() error {

	beaconBestState.ShardPendingValidator = make(map[byte][]incognitokey.CommitteePublicKey)
	for i := 0; i < beaconBestState.ActiveShards; i++ {
		committeePublicKey := statedb.GetOneShardSubstituteValidator(beaconBestState.consensusStateDB, byte(i))
		beaconBestState.ShardPendingValidator[byte(i)] = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
		for index, value := range committeePublicKey {
			beaconBestState.ShardPendingValidator[byte(i)][index] = value
		}
	}
	return nil
}

//RestoreCandidateShardWaitingForCurrentRandom ...
func (beaconBestState *BeaconBestState) RestoreCandidateShardWaitingForCurrentRandom() error {

	//GetCurrentEpochCandidate
	committeePublicKey := statedb.GetCurrentEpochCandidate(beaconBestState.consensusStateDB)
	beaconBestState.CandidateShardWaitingForCurrentRandom = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		beaconBestState.CandidateShardWaitingForCurrentRandom[i] = v
	}
	return nil
}

//RestoreCandidateBeaconWaitingForCurrentRandom ...
func (beaconBestState *BeaconBestState) RestoreCandidateBeaconWaitingForCurrentRandom() error {

	//TODO: @tin
	// For further development, when beacon is round robin for community -> change here

	beaconBestState.CandidateBeaconWaitingForCurrentRandom = make([]incognitokey.CommitteePublicKey, 0)
	return nil
}

//RestoreCandidateShardWaitingForNextRandom ...
func (beaconBestState *BeaconBestState) RestoreCandidateShardWaitingForNextRandom() error {
	//GetNextEpochCandidate
	committeePublicKey := statedb.GetNextEpochCandidate(beaconBestState.consensusStateDB)
	beaconBestState.CandidateShardWaitingForNextRandom = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		beaconBestState.CandidateShardWaitingForNextRandom[i] = v
	}
	return nil
}

//RestoreCandidateBeaconWaitingForNextRandom ...
func (beaconBestState *BeaconBestState) RestoreCandidateBeaconWaitingForNextRandom() error {

	//TODO: @tin
	// For further development, when beacon is round robin for community -> change here

	beaconBestState.CandidateBeaconWaitingForNextRandom = make([]incognitokey.CommitteePublicKey, 0)
	return nil
}

//RecoverCommittee ...
func (shardBestState *ShardBestState) RestoreCommittee(shardID byte, chain *BlockChain) error {

	committeePublicKey := statedb.GetOneShardCommittee(shardBestState.consensusStateDB, shardID)

	shardBestState.ShardCommittee = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		shardBestState.ShardCommittee[i] = v
	}

	return nil
}

//RestorePendingValidators ...
func (shardBestState *ShardBestState) RestorePendingValidators(shardID byte, bc *BlockChain) error {

	committeePublicKey := statedb.GetOneShardSubstituteValidator(shardBestState.consensusStateDB, shardID)
	shardBestState.ShardPendingValidator = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		shardBestState.ShardPendingValidator[i] = v
	}
	return nil
}

// //
// func (shardBestState *ShardBestState) restoreViewFromHash(blockchain *BlockChain) error {
// 	return nil
// }

//RestoreBeaconViewStateFromHash ...
func (beaconBestState *BeaconBestState) RestoreBeaconViewStateFromHash(blockchain *BlockChain) error {
	err := beaconBestState.InitStateRootHash(blockchain)
	if err != nil {
		return err
	}

	//best block
	block, _, err := blockchain.GetBeaconBlockByHash(beaconBestState.BestBlockHash)
	if err != nil || block == nil {
		return err
	}
	beaconBestState.BestBlock = *block
	beaconBestState.BeaconHeight = block.GetHeight()
	if beaconBestState.RewardReceiver == nil {
		beaconBestState.RewardReceiver = make(map[string]privacy.PaymentAddress)
	}
	err = beaconBestState.RestoreBeaconCommittee()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreShardCommittee()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreBeaconPendingValidator()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreShardPendingValidator()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreCandidateBeaconWaitingForCurrentRandom()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreCandidateBeaconWaitingForNextRandom()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreCandidateShardWaitingForCurrentRandom()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreCandidateShardWaitingForNextRandom()
	if err != nil {
		panic(err)
	}

	beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(beaconBestState.ConsensusStateDBRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetBeaconChainDatabase()))
	if err != nil {
		panic(err)
	}
	beaconBestState.AutoStaking = NewMapStringBool()
	sids := []int{}
	for i := 0; i < beaconBestState.ActiveShards; i++ {
		sids = append(sids, i)
	}
	beaconBestState.AutoStaking.data = statedb.GetMapAutoStaking(beaconConsensusStateDB, sids)

	return nil
}

//RestoreBeaconViewStateFromHash ...
func (beaconBestState *BeaconBestState) RestoreDetailsBeaconViewStateFromHash(blockchain *BlockChain) error {

	err := beaconBestState.InitStateRootHash(blockchain)
	if err != nil {
		return err
	}
	var chainParamEpoch = blockchain.config.ChainParams.Epoch
	var randomTime = blockchain.config.ChainParams.RandomTime
	//best block
	block, _, err := blockchain.GetBeaconBlockByHash(beaconBestState.BestBlockHash)
	if err != nil || block == nil {
		return err
	}

	Logger.log.Infof("Shard State %v ", block.Body.ShardState )
	Logger.log.Infof("Beacon Consensus %v ", block.GetConsensusType() )
	Logger.log.Infof("Beacon Consensus 1 %v ", block.Header.ConsensusType )

	Logger.log.Infof("Beacon header %v ", block.Header )
	beaconBestState.BestBlock = *block
	beaconBestState.BeaconHeight = block.GetHeight()
	beaconBestState.PreviousBestBlockHash = block.GetPrevHash()
	beaconBestState.BestShardHash = make(map[byte]common.Hash)
	beaconBestState.BestShardHeight = make(map[byte]uint64)
	beaconBestState.ConsensusAlgorithm = block.Header.ConsensusType

	// Update new best new block hash
	for shardID, shardStates := range block.Body.ShardState {
		beaconBestState.BestShardHash[shardID] = shardStates[len(shardStates)-1].Hash
		beaconBestState.BestShardHeight[shardID] = shardStates[len(shardStates)-1].Height
	}
	beaconBestState.ShardConsensusAlgorithm = make(map[byte]string)
	for shardID := 0; shardID < beaconBestState.ActiveShards; shardID++ {
		beaconBestState.ShardConsensusAlgorithm[byte(shardID)] = common.BlsConsensus
	}

	beaconBestState.Epoch = block.Header.Epoch
	if beaconBestState.BeaconHeight%chainParamEpoch == 1 && beaconBestState.BeaconHeight != 1 {
		// Begin of each epoch
		beaconBestState.IsGetRandomNumber = false
		// Before get random from bitcoin
	} else if  beaconBestState.BeaconHeight % chainParamEpoch == randomTime {
		beaconBestState.CurrentRandomTimeStamp = block.Header.Timestamp
	}

	err = beaconBestState.RestoreBeaconCommittee()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreShardCommittee()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreBeaconPendingValidator()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreShardPendingValidator()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreCandidateBeaconWaitingForCurrentRandom()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreCandidateBeaconWaitingForNextRandom()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreCandidateShardWaitingForCurrentRandom()
	if err != nil {
		panic(err)
	}

	err = beaconBestState.RestoreCandidateShardWaitingForNextRandom()
	if err != nil {
		panic(err)
	}

	beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(beaconBestState.ConsensusStateDBRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetBeaconChainDatabase()))
	if err != nil {
		panic(err)
	}
	beaconBestState.AutoStaking = NewMapStringBool()
	sids := []int{}
	for i := 0; i < beaconBestState.ActiveShards; i++ {
		sids = append(sids, i)
	}
	beaconBestState.AutoStaking.data = statedb.GetMapAutoStaking(beaconConsensusStateDB, sids)


	if beaconBestState.RewardReceiver == nil {
		beaconBestState.RewardReceiver = make(map[string]privacy.PaymentAddress)
	}

	return nil
}

//RestoreShardCommittee ...
func (shardBestState *ShardBestState) RestoreShardCommittee() error {
	shardBestState.ShardCommittee = make([]incognitokey.CommitteePublicKey, 0)
	shardBestState.ShardCommittee = statedb.GetOneShardCommittee(shardBestState.consensusStateDB, shardBestState.ShardID)
	return nil
}

//RestoreShardCommittee ...
func (shardBestState *ShardBestState) RestoreShardValidator() error {
	shardBestState.ShardCommittee = make([]incognitokey.CommitteePublicKey, 0)
	shardBestState.ShardCommittee = statedb.GetOneShardSubstituteValidator(shardBestState.consensusStateDB, shardBestState.ShardID)
	return nil
}

//RestoreShardViewStateFromHash ...
func (shardBestState *ShardBestState) RestoreDetailsShardViewStateFromHash(shardBlock *ShardBlock, blockchain *BlockChain) error {
	err := shardBestState.InitStateRootHash(blockchain.GetShardChainDatabase(byte(shardBlock.GetShardID())), blockchain)
	if err != nil {
		return err
	}
	shardBestState.BestBlock = shardBlock
	shardBestState.BestBeaconHash = shardBlock.Header.BeaconHash
	shardBestState.ShardHeight = shardBlock.Header.Height
	shardBestState.Epoch = shardBlock.Header.Epoch
	shardBestState.BeaconHeight = shardBlock.Header.BeaconHeight
	shardBestState.TotalTxns = 0
	shardBestState.NumTxns = uint64(len(shardBlock.Body.Transactions))
	shardBestState.ConsensusAlgorithm = common.BlsConsensus
	shardBestState.ActiveShards = blockchain.config.ChainParams.ActiveShards
	shardBestState.BlockInterval = blockchain.config.ChainParams.MinShardBlockInterval
	shardBestState.BlockMaxCreateTime = blockchain.config.ChainParams.MaxShardBlockCreation
	shardBestState.MaxShardCommitteeSize = blockchain.config.ChainParams.MaxShardCommitteeSize
	shardBestState.MinShardCommitteeSize = blockchain.config.ChainParams.MinShardCommitteeSize



	return nil


}