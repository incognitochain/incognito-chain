package rpcserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/consensus_v2/blsbft"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/patrickmn/go-cache"

	"github.com/incognitochain/incognito-chain/dataaccessobject/stats"

	"github.com/incognitochain/incognito-chain/consensus_v2"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/pkg/errors"
)

type txs struct {
	Txs []string `json:"Txs"`
}

func (httpServer *HttpServer) handleTestHttpServer(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return nil, nil
}

/*
For testing and benchmark only
*/
type CountResult struct {
	Success int
	Fail    int
}

func (httpServer *HttpServer) handleUnlockMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	httpServer.config.TxMemPool.SendTransactionToBlockGen()
	return nil, nil
}

func (httpServer *HttpServer) handleGetConsensusInfoV3(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	_, ok := httpServer.config.ConsensusEngine.(*consensus_v2.Engine)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("consensus engine not found, got "+reflect.TypeOf(httpServer.config.ConsensusEngine).String()))
	}

	arr := []interface{}{}
	/*for chainID, bftactor := range engine.BFTProcess() {*/
	//bftactorV2, ok := bftactor.(temp)
	//if !ok {
	//continue
	//}
	//m := map[string]interface{}{
	//"ChainID":              chainID,
	//"VoteHistory":          bftactorV3.GetVoteHistory(),
	//"ReceiveBlockByHash":   bftactorV3.GetReceiveBlockByHash(),
	//"ReceiveBlockByHeight": bftactorV3.GetReceiveBlockByHeight(),
	//}
	//arr = append(arr, m)
	/*}*/

	return arr, nil
}

func (httpServer *HttpServer) handleGetAutoEnableFeatureConfig(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return config.Param().AutoEnableFeature, nil
}

func (httpServer *HttpServer) handleSetAutoEnableFeatureConfig(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	if config.Config().Network() == config.MainnetNetwork {
		return nil, nil
	}
	arrayParams := common.InterfaceSlice(params)
	jsonStr := arrayParams[0].(string)

	v := map[string]config.AutoEnableFeature{}
	err := json.Unmarshal([]byte(jsonStr), &v)
	if err != nil {
		return nil, rpcservice.NewRPCError(-1, err)
	}
	config.Param().AutoEnableFeature = v
	httpServer.GetBlockchain().SendFeatureStat()
	return nil, nil
}

func (httpServer *HttpServer) handleSendFinishSync(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	miningKeyStr := arrayParams[0].(string)
	cpk := arrayParams[1].(string)
	sid := arrayParams[2].(float64)
	miningKey, err := consensus_v2.GetMiningKeyFromPrivateSeed(miningKeyStr)
	if err != nil {
		return nil, rpcservice.NewRPCError(-1, err)
	}
	finishedSyncValidators := []string{}
	finishedSyncSignatures := [][]byte{}
	signature, err := miningKey.BriSignData([]byte(wire.CmdMsgFinishSync))
	if err != nil {
		return nil, rpcservice.NewRPCError(-1, err)
	}
	finishedSyncSignatures = append(finishedSyncSignatures, signature)
	finishedSyncValidators = append(finishedSyncValidators, cpk)

	msg := wire.NewMessageFinishSync(finishedSyncValidators, finishedSyncSignatures, byte(sid))
	if err := httpServer.config.Server.PushMessageToShard(msg, common.BeaconChainSyncID); err != nil {
		return nil, rpcservice.NewRPCError(-1, err)
	}
	return nil, nil
}

func (httpServer *HttpServer) handleGetAutoStakingByHeight(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	height := int(arrayParams[0].(float64))
	beaconConsensusStateRootHash, err := httpServer.blockService.BlockChain.GetBeaconConsensusRootHash(httpServer.blockService.BlockChain.GetBeaconBestState(), uint64(height))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	// beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(beaconConsensusStateRootHash, statedb.NewDatabaseAccessWarper(httpServer.blockService.BlockChain.GetBeaconChainDatabase()))
	// if err != nil {
	// 	return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	// }
	// _, newAutoStaking := statedb.GetRewardReceiverAndAutoStaking(beaconConsensusStateDB, httpServer.blockService.BlockChain.GetShardIDs())
	newAutoStaking := map[string]bool{}
	return []interface{}{beaconConsensusStateRootHash, newAutoStaking}, nil
}

func (httpServer *HttpServer) handleGetTotalBlockInEpoch(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(-1, errors.New("Invalid number of params, accept only 1 value"))
	}
	epoch := uint64(arrayParams[0].(float64))

	res := make(map[byte]*stats.NumberOfBlockInOneEpochStats)
	for i := 0; i < httpServer.config.BlockChain.BeaconChain.GetActiveShardNumber(); i++ {
		shardID := byte(i)
		numberOfBlockInOneEpochStats, err := stats.GetShardEpochBPV3Stats(httpServer.config.BlockChain.GetShardChainDatabase(shardID), shardID, epoch)
		if err != nil {
			return nil, rpcservice.NewRPCError(-1, err)
		}
		res[shardID] = numberOfBlockInOneEpochStats
	}
	return res, nil
}

func (httpServer *HttpServer) handleGetDetailBlocksOfEpoch(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	shardID := byte(arrayParams[0].(float64))
	epoch := uint64(arrayParams[1].(float64))
	if len(arrayParams) != 2 {
		return nil, rpcservice.NewRPCError(-1, errors.New("Invalid number of params, accept only 2 value"))
	}
	res, err := stats.GetShardHeightBPV3Stats(httpServer.config.BlockChain.GetShardChainDatabase(shardID), shardID, epoch)
	if err != nil {
		return nil, rpcservice.NewRPCError(-1, err)
	}
	return res, nil
}

func (httpServer *HttpServer) handleGetCommitteeState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	height := uint64(arrayParams[0].(float64))
	tempHash := arrayParams[1].(string)

	var beaconConsensusStateRootHash = &blockchain.BeaconRootHash{}
	var err1 error = nil

	currentValidator := make(map[int][]incognitokey.CommitteePublicKey)
	substituteValidator := make(map[int][]incognitokey.CommitteePublicKey)
	nextEpochShardCandidate := []incognitokey.CommitteePublicKey{}
	currentEpochShardCandidate := []incognitokey.CommitteePublicKey{}
	nextEpochBeaconCandidate := []incognitokey.CommitteePublicKey{}
	currentEpochBeaconCandidate := []incognitokey.CommitteePublicKey{}
	syncingValidators := make(map[byte][]incognitokey.CommitteePublicKey)
	rewardReceivers := make(map[string]key.PaymentAddress)
	autoStaking := make(map[string]bool)
	stakingTx := map[string]common.Hash{}
	delegateList := map[string]string{}

	if height == 0 && tempHash == "" {
		bState := httpServer.GetBlockchain().GetBeaconBestState()
		beaconConsensusStateRootHash, err1 = blockchain.GetBeaconRootsHashByBlockHash(
			httpServer.config.BlockChain.GetBeaconChainDatabase(),
			bState.BestBlockHash,
		)
		if err1 != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
		}
		cs := bState.GetCommitteeState()
		currentEpochShardCandidate = cs.GetCandidateShardWaitingForCurrentRandom()
		currentEpochBeaconCandidate = cs.GetCandidateBeaconWaitingForCurrentRandom()
		nextEpochShardCandidate = cs.GetCandidateBeaconWaitingForCurrentRandom()
		nextEpochBeaconCandidate = cs.GetCandidateBeaconWaitingForNextRandom()
		syncingValidators = cs.GetSyncingValidators()
		rewardReceivers = cs.GetRewardReceiver()
		autoStaking = cs.GetAutoStaking()
		delegateList = cs.GetDelegate()
		for i, v := range cs.GetShardCommittee() {
			currentValidator[int(i)] = v
		}
		currentValidator[-1] = cs.GetBeaconCommittee()
		for i, v := range cs.GetShardSubstitute() {
			substituteValidator[int(i)] = v
		}
		substituteValidator[-1] = cs.GetBeaconSubstitute()
		stakingTx = cs.GetStakingTx()
	} else {
		if height == 0 || tempHash != "" {
			hash, err := common.Hash{}.NewHashFromStr(tempHash)
			if err != nil {
				return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
			}
			beaconConsensusStateRootHash, err1 = blockchain.GetBeaconRootsHashByBlockHash(
				httpServer.config.BlockChain.GetBeaconChainDatabase(),
				*hash,
			)
			if err1 != nil {
				return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
			}
		} else {
			beaconConsensusStateRootHash, err1 = httpServer.config.BlockChain.GetBeaconRootsHashFromBlockHeight(
				height,
			)
			if err1 != nil {
				return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
			}
		}
		shardIDs := []int{-1}
		shardIDs = append(shardIDs, httpServer.config.BlockChain.GetShardIDs()...)
		stateDB, err2 := statedb.NewWithPrefixTrie(beaconConsensusStateRootHash.ConsensusStateDBRootHash,
			statedb.NewDatabaseAccessWarper(httpServer.config.BlockChain.GetBeaconChainDatabase()))
		if err2 != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
		}

		currentValidator, substituteValidator, nextEpochShardCandidate, currentEpochShardCandidate, nextEpochBeaconCandidate, currentEpochBeaconCandidate, syncingValidators, rewardReceivers, autoStaking, stakingTx, delegateList =
			statedb.GetAllCandidateSubstituteCommittee(stateDB, shardIDs)
	}

	currentValidatorStr := make(map[int][]string)
	for shardID, v := range currentValidator {
		tempV, _ := incognitokey.CommitteeKeyListToString(v)
		currentValidatorStr[shardID] = tempV
	}
	substituteValidatorStr := make(map[int][]string)
	for shardID, v := range substituteValidator {
		tempV, _ := incognitokey.CommitteeKeyListToString(v)
		substituteValidatorStr[shardID] = tempV
	}
	syncingValidatorsStr := make(map[int][]string)
	for shardID, v := range syncingValidators {
		tempV, _ := incognitokey.CommitteeKeyListToString(v)
		syncingValidatorsStr[int(shardID)] = tempV
	}
	nextEpochShardCandidateStr, _ := incognitokey.CommitteeKeyListToString(nextEpochShardCandidate)
	currentEpochShardCandidateStr, _ := incognitokey.CommitteeKeyListToString(currentEpochShardCandidate)
	tempStakingTx := make(map[string]string)
	for k, v := range stakingTx {
		tempStakingTx[k] = v.String()
	}
	tempRewardReceiver := make(map[string]string)
	for k, v := range rewardReceivers {
		wl := wallet.KeyWallet{}
		wl.KeySet.PaymentAddress = v
		paymentAddress := wl.Base58CheckSerialize(wallet.PaymentAddressType)
		tempRewardReceiver[k] = paymentAddress
	}
	beaconWaitingStr, _ := incognitokey.CommitteeKeyListToString(
		append(currentEpochBeaconCandidate, nextEpochBeaconCandidate...),
	)
	syncingValidatorsStr[-1] = beaconWaitingStr

	shardStakerInfos := map[string]*statedb.ShardStakerInfo{}
	beaconStakerInfos := map[string]*statedb.BeaconStakerInfo{}

	return &jsonresult.CommiteeState{
		Root:              beaconConsensusStateRootHash.ConsensusStateDBRootHash.String(),
		Committee:         currentValidatorStr,
		Substitute:        substituteValidatorStr,
		NextCandidate:     nextEpochShardCandidateStr,
		CurrentCandidate:  currentEpochShardCandidateStr,
		RewardReceivers:   tempRewardReceiver,
		AutoStaking:       autoStaking,
		StakingTx:         tempStakingTx,
		Syncing:           syncingValidatorsStr,
		DelegateList:      delegateList,
		ShardStakerInfos:  shardStakerInfos,
		BeaconStakerInfos: beaconStakerInfos,
	}, nil
}

func (httpServer *HttpServer) handleGetShardCommitteeFromBeaconHash(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	shardID := uint64(arrayParams[0].(float64))
	tempHash := arrayParams[1].(string)
	hash, err := common.Hash{}.NewHashFromStr(tempHash)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	committees, err := httpServer.GetBlockchain().GetShardCommitteeFromBeaconHash(*hash, byte(shardID))
	return committees, nil

}

func (httpServer *HttpServer) handleGetCommitteeStateByShard(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	shardID := uint64(arrayParams[0].(float64))
	tempHash := arrayParams[1].(string)

	hash, err := common.Hash{}.NewHashFromStr(tempHash)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	shardRootHash, err := blockchain.GetShardRootsHashByBlockHash(
		httpServer.config.BlockChain.GetShardChainDatabase(byte(shardID)),
		byte(shardID),
		*hash,
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	stateDB, err := statedb.NewWithPrefixTrie(shardRootHash.ConsensusStateDBRootHash,
		statedb.NewDatabaseAccessWarper(httpServer.config.BlockChain.GetShardChainDatabase(byte(shardID))))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	committees := statedb.GetOneShardCommittee(stateDB, byte(shardID))
	resCommittees := make([]string, len(committees))
	for i := 0; i < len(resCommittees); i++ {
		key, _ := committees[i].ToBase58()
		resCommittees[i] = key
	}
	substitutes := statedb.GetOneShardSubstituteValidator(stateDB, byte(shardID))
	resSubstitutes := make([]string, len(substitutes))
	for i := 0; i < len(resSubstitutes); i++ {
		key, _ := substitutes[i].ToBase58()
		resSubstitutes[i] = key
	}

	return map[string]interface{}{
		"root":       shardRootHash.ConsensusStateDBRootHash,
		"shardID":    shardID,
		"committee":  resCommittees,
		"substitute": resSubstitutes,
	}, nil
}

func (httpServer *HttpServer) handleGetSlashingCommittee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Number Of Params"))
	}
	epoch := uint64(arrayParams[0].(float64))
	if epoch < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Epoch Value"))
	}
	beaconBestState := httpServer.blockService.BlockChain.GetBeaconBestState()
	if epoch >= beaconBestState.Epoch {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Epoch Value"+
			"expect epoch from %+v to %+v", 1, beaconBestState.Epoch-1))
	}
	slashingCommittee := statedb.GetSlashingCommittee(beaconBestState.GetBeaconSlashStateDB(), epoch)
	return slashingCommittee, nil
}

func (httpServer *HttpServer) handleGetSlashingCommitteeDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Number Of Params"))
	}
	epoch := uint64(arrayParams[0].(float64))
	if epoch < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Epoch Value"))
	}
	beaconBestState := httpServer.blockService.BlockChain.GetBeaconBestState()
	if epoch >= beaconBestState.Epoch {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Epoch Value"+
			"expect epoch from %+v to %+v", 1, beaconBestState.Epoch-1))
	}
	slashingCommittees := statedb.GetSlashingCommittee(beaconBestState.GetBeaconSlashStateDB(), epoch)
	slashingCommitteeDetail := make(map[byte][]incognitokey.CommitteeKeyString)
	for shardID, slashingCommittee := range slashingCommittees {
		res, err := incognitokey.CommitteeBase58KeyListToStruct(slashingCommittee)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
		slashingCommitteeDetail[shardID] = incognitokey.CommitteeKeyListToStringList(res)
	}
	return slashingCommitteeDetail, nil
}

func (httpServer *HttpServer) handleGetRewardAmountByEpoch(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 2 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("want length %+v but got %+v", 2, len(arrayParams)))
	}
	tempShardID, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid ShardID Value"))
	}
	tempEpoch, ok := arrayParams[1].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Epoch Value"))
	}
	shardID := byte(tempShardID)
	epoch := uint64(tempEpoch)
	rewardStateDB := httpServer.config.BlockChain.GetBeaconBestState().GetBeaconRewardStateDB()
	amount, err := statedb.GetRewardOfShardByEpoch(rewardStateDB, epoch, shardID, common.PRVCoinID)
	return amount, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
}

func (httpServer *HttpServer) handleGetFinalityProof(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 2 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("want length %+v but got %+v", 2, len(arrayParams)))
	}
	tempShardID, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid ShardID Value"))
	}
	tempHash, ok := arrayParams[1].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Epoch Value"))
	}
	shardID := byte(tempShardID)
	hash := common.Hash{}.NewHashFromStr2(tempHash)
	shardBlock, m, err := httpServer.config.BlockChain.ShardChain[shardID].GetFinalityProof(hash)
	return map[string]interface{}{
		"Block": shardBlock,
		"Data":  m,
	}, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
}

func (httpServer *HttpServer) handleSetConsensusRule(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	if config.Param().Net == config.MainnetNet {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidRequestError, fmt.Errorf("Cannot execute on mainnet"))
	}

	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("want length %+v but got %+v", 1, len(arrayParams)))
	}

	param, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("invalid flag Value"))
	}

	voteRule := param["vote_rule"]
	preVoteRule := param["prevote_rule"]
	createRule := param["create_rule"]
	handleVoteRule := param["handle_vote_rule"]
	handleProposeRule := param["handle_propose_rule"]
	insertRule := param["insert_rule"]
	validatorRule := param["validator_rule"]

	if voteRule != nil && voteRule.(string) != "" {
		blsbft.ActorRuleBuilderContext.VoteRule = voteRule.(string)
	}
	if preVoteRule != nil && preVoteRule.(string) != "" {
		blsbft.ActorRuleBuilderContext.PreVoteRule = preVoteRule.(string)
	}
	if createRule != nil && createRule.(string) != "" {
		blsbft.ActorRuleBuilderContext.CreateRule = createRule.(string)
	}
	if handleVoteRule != nil && handleVoteRule.(string) != "" {
		blsbft.ActorRuleBuilderContext.HandleVoteRule = handleVoteRule.(string)
	}
	if handleProposeRule != nil && handleProposeRule.(string) != "" {
		blsbft.ActorRuleBuilderContext.HandleProposeRule = handleProposeRule.(string)
	}
	if insertRule != nil && insertRule.(string) != "" {
		blsbft.ActorRuleBuilderContext.InsertRule = insertRule.(string)
	}
	if validatorRule != nil && validatorRule.(string) != "" {
		blsbft.ActorRuleBuilderContext.ValidatorRule = validatorRule.(string)
	}

	return map[string]interface{}{
		"vote_rule":           blsbft.ActorRuleBuilderContext.VoteRule,
		"create_rule":         blsbft.ActorRuleBuilderContext.CreateRule,
		"handle_vote_rule":    blsbft.ActorRuleBuilderContext.HandleVoteRule,
		"handle_propose_rule": blsbft.ActorRuleBuilderContext.HandleProposeRule,
		"insert_rule":         blsbft.ActorRuleBuilderContext.InsertRule,
		"validator_rule":      blsbft.ActorRuleBuilderContext.ValidatorRule,
		"lemma2_height":       blsbft.ActorRuleBuilderContext.Lemma2Height,
	}, nil
}

func (httpServer *HttpServer) handleGetByzantineDetectorInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return blsbft.ByzantineDetectorObject.GetByzantineDetectorInfo(), nil
}

func (httpServer *HttpServer) handleGetByzantineBlackList(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	res := blsbft.ByzantineDetectorObject.GetByzantineDetectorInfo()
	return res["BlackList"], nil
}

func (httpServer *HttpServer) handleRemoveByzantineDetector(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("want length %+v but got %+v", 1, len(arrayParams)))
	}
	validator, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid validator type"))
	}
	err := blsbft.ByzantineDetectorObject.RemoveBlackListValidator(validator)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	return "Delete Black list validator " + validator, nil
}

func (httpServer *HttpServer) handleGetConsensusRule(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return blsbft.ActorRuleBuilderContext, nil
}

func (httpServer *HttpServer) handleGetConsensusData(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("want length %+v but got %+v", 1, len(arrayParams)))
	}
	tempChainID, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid chain id type"))
	}
	chainID := int(tempChainID)
	voteHistory, err := blsbft.InitVoteHistory(chainID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	proposeHistory, err := blsbft.InitProposeHistory(chainID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	receiveBlockByHash, err := blsbft.InitReceiveBlockByHash(chainID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return map[string]interface{}{
		"voteHistory":        voteHistory,
		"proposeHistory":     proposeHistory,
		"receiveBlockByHash": receiveBlockByHash,
	}, nil
}

func (httpServer *HttpServer) handleGetProposerIndex(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("want length %+v but got %+v", 2, len(arrayParams)))
	}
	tempShardID, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid ShardID Value"))
	}

	sChain := httpServer.blockService.BlockChain.ShardChain[byte(tempShardID)]

	shardBestState := sChain.GetBestState()
	tempCommittee, committeIndex := blsbft.GetProposerByTimeSlotFromCommitteeList(shardBestState.CalculateTimeSlot(time.Now().Unix()), shardBestState.GetShardCommittee(), shardBestState.GetProposerLength())
	committee, _ := tempCommittee.ToBase58()

	return map[string]interface{}{
		"Proposer":      committee,
		"ProposerIndex": committeIndex,
	}, nil
}

func (httpServer *HttpServer) handleGetAndSendTxsFromFile(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	shardIDParam := int(arrayParams[0].(float64))
	Logger.log.Critical(arrayParams)
	txType := arrayParams[1].(string)
	isSent := arrayParams[2].(bool)
	interval := int64(arrayParams[3].(float64))
	Logger.log.Criticalf("Interval between transactions %+v \n", interval)
	datadir := "./bin/"
	filename := ""
	success := 0
	fail := 0
	switch txType {
	case "privacy3000_1":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-privacy-3000.1.json"
	case "privacy3000_2":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-privacy-3000.2.json"
	case "privacy3000_3":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-privacy-3000.3.json"
	case "noprivacy9000":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-9000.json"
	case "noprivacy10000_2":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-10000.2.json"
	case "noprivacy10000_3":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-10000.3.json"
	case "noprivacy10000_4":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-10000.4.json"
	case "noprivacy":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-5000.json"
	case "privacy":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-privacy-5000.json"
	case "cstoken":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-cstoken-5000.json"
	case "cstokenprivacy":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-cstokenprivacy-5000.json"
	default:
		return CountResult{}, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("Can't find file"))
	}

	Logger.log.Critical("Getting Transactions from file: ", datadir+filename)
	file, err := ioutil.ReadFile(datadir + filename)
	if err != nil {
		Logger.log.Error("Fail to get Transactions from file: ", err)
	}
	data := txs{}
	count := 0
	_ = json.Unmarshal([]byte(file), &data)
	Logger.log.Criticalf("Get %+v Transactions from file \n", len(data.Txs))
	intervalDuration := time.Duration(interval) * time.Millisecond
	for index, txBase58Data := range data.Txs {
		<-time.Tick(intervalDuration)
		Logger.log.Critical("Number of Transaction: ", index)
		//<-time.Tick(50*time.Millisecond)
		rawTxBytes, _, err := base58.Base58Check{}.Decode(txBase58Data)
		if err != nil {
			fail++
			continue
		}
		switch txType {
		case "cstokenprivacy":
			{
				tx, err := transaction.NewTransactionTokenFromJsonBytes(rawTxBytes)
				if err != nil {
					fail++
					continue
				}
				if !isSent {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
					if err != nil {
						fail++
						continue
					} else {
						success++
						continue
					}
				} else {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
					//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
					if err != nil {
						fail++
						continue
					}
					txMsg, err := wire.MakeEmptyMessage(wire.CmdPrivacyCustomToken)
					if err != nil {
						fail++
						continue
					}
					txMsg.(*wire.MessageTxPrivacyToken).Transaction = tx
					err = httpServer.config.Server.PushMessageToAll(txMsg)
					if err != nil {
						fail++
						continue
					}
				}
				if err == nil {
					count++
					success++
				}
			}
		default:
			tx, err := transaction.NewTransactionFromJsonBytes(rawTxBytes)
			if err != nil {
				fail++
				continue
			}
			if !isSent {
				_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
				if err != nil {
					fail++
					continue
				} else {
					success++
					continue
				}
			} else {
				_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
				//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
				if err != nil {
					fail++
					continue
				}
				txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
				if err != nil {
					fail++
					continue
				}
				txMsg.(*wire.MessageTx).Transaction = tx
				err = httpServer.config.Server.PushMessageToAll(txMsg)
				if err != nil {
					fail++
					continue
				}
			}
		}
		if err == nil {
			count++
			success++
		}
	}
	return CountResult{Success: success, Fail: fail}, nil
}

func (httpServer *HttpServer) handleGetAndSendTxsFromFileV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	Logger.log.Critical(arrayParams)
	shardIDParam := int(arrayParams[0].(float64))
	txType := arrayParams[1].(string)
	isSent := arrayParams[2].(bool)
	interval := int64(arrayParams[3].(float64))
	Logger.log.Criticalf("Interval between transactions %+v \n", interval)
	datadir := "./utility/"
	Txs := []string{}
	filename := ""
	filenames := []string{}
	success := 0
	fail := 0
	count := 0
	switch txType {
	case "noprivacy":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-9000.json"
		filenames = append(filenames, filename)
		for i := 2; i <= 3; i++ {
			filename := "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-10000." + fmt.Sprintf("%d", i) + ".json"
			filenames = append(filenames, filename)
		}
	default:
		return CountResult{}, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("Can't find file"))
	}
	for _, filename := range filenames {
		Logger.log.Critical("Getting Transactions from file: ", datadir+filename)
		file, err := ioutil.ReadFile(datadir + filename)
		if err != nil {
			Logger.log.Error("Fail to get Transactions from file: ", err)
			continue
		}
		data := txs{}
		_ = json.Unmarshal([]byte(file), &data)
		Logger.log.Criticalf("Get %+v Transactions from file %+v \n", len(data.Txs), filename)
		Txs = append(Txs, data.Txs...)
	}
	intervalDuration := time.Duration(interval) * time.Millisecond
	for index, txBase58Data := range Txs {
		<-time.Tick(intervalDuration)
		Logger.log.Critical("Number of Transaction: ", index)
		//<-time.Tick(50*time.Millisecond)
		rawTxBytes, _, err := base58.Base58Check{}.Decode(txBase58Data)
		if err != nil {
			fail++
			continue
		}
		switch txType {
		case "cstokenprivacy":
			{
				tx, err := transaction.NewTransactionTokenFromJsonBytes(rawTxBytes)
				if err != nil {
					fail++
					continue
				}
				if !isSent {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
					if err != nil {
						fail++
						continue
					} else {
						success++
						continue
					}
				} else {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
					//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
					if err != nil {
						fail++
						continue
					}
					txMsg, err := wire.MakeEmptyMessage(wire.CmdPrivacyCustomToken)
					if err != nil {
						fail++
						continue
					}
					txMsg.(*wire.MessageTxPrivacyToken).Transaction = tx
					err = httpServer.config.Server.PushMessageToAll(txMsg)
					if err != nil {
						fail++
						continue
					}
				}
				if err == nil {
					count++
					success++
				}
			}
		default:
			tx, err := transaction.NewTransactionFromJsonBytes(rawTxBytes)
			if err != nil {
				fail++
				continue
			}
			if !isSent {
				_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
				if err != nil {
					fail++
					continue
				} else {
					success++
					continue
				}
			} else {
				_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
				//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
				if err != nil {
					fail++
					continue
				}
				txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
				if err != nil {
					fail++
					continue
				}
				txMsg.(*wire.MessageTx).Transaction = tx
				err = httpServer.config.Server.PushMessageToAll(txMsg)
				if err != nil {
					fail++
					continue
				}
			}
		}
		if err == nil {
			count++
			success++
		}
	}
	return CountResult{Success: success, Fail: fail}, nil
}

func (httpServer *HttpServer) handleConvertPaymentAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("a payment address is needed to proceed"))
	}
	address, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("the payment address should be a string"))
	}

	convertedAddress, err := wallet.GetPaymentAddressV1(address, false)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	return convertedAddress, nil
}

// func (httpServer *HttpServer) handleTestBuildDoubleSpendTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
// 	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
// 	if errNewParam != nil {
// 		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
// 	}

// 	txs, err := httpServer.txService.TestBuildDoubleSpendingTransaction(createRawTxParam, nil)
// 	if err != nil {
// 		// return hex for a new tx
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	// tx := data.(jsonresult.CreateTransactionResult)
// 	// base58CheckData := tx.Base58CheckData
// 	// newParam := make([]interface{}, 0)
// 	// newParam = append(newParam, base58CheckData)
// 	// sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
// 	// if err != nil {
// 	// 	return nil, rpcservice.NewRPCError(rpcservice.SendTxDataError, err)
// 	// }
// 	// result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, tx.ShardID)
// 	return result, nil
// }

// func (httpServer *HttpServer) handleTestBuildDuplicateInputTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
// 	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
// 	if errNewParam != nil {
// 		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
// 	}

// 	txs, err := httpServer.txService.TestBuildDuplicateInputTransaction(createRawTxParam, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	return result, nil
// }

// func (httpServer *HttpServer) handleTestBuildOutGtInTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
// 	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
// 	if errNewParam != nil {
// 		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
// 	}

// 	txs, err := httpServer.txService.TestBuildOutGtInTransaction(createRawTxParam, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	return result, nil
// }

// func (httpServer *HttpServer) handleTestBuildReceiverExistsTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
// 	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
// 	if errNewParam != nil {
// 		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
// 	}

// 	txs, err := httpServer.txService.TestBuildReceiverExistsTransaction(createRawTxParam, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	return result, nil
// }

// func (httpServer *HttpServer) handleTestBuildDoubleSpendTokenTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
// 	// createRawTxParam, errNewParam := bean.NewCreateRawPrivacyTokenTxParam(params)
// 	// if errNewParam != nil {
// 	// 	return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
// 	// }

// 	txs, err := httpServer.txService.TestBuildDoubleSpendingTokenTransaction(params, nil)
// 	if err != nil {
// 		// return hex for a new tx
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	return result, nil
// }

// func (httpServer *HttpServer) handleTestBuildDuplicateInputTokenTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

// 	txs, err := httpServer.txService.TestBuildDuplicateInputTokenTransaction(params, nil)
// 	if err != nil {
// 		// return hex for a new tx
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	return result, nil
// }

// func (httpServer *HttpServer) handleTestBuildReceiverExistsTokenTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

// 	txs, err := httpServer.txService.TestBuildReceiverExistsTokenTransaction(params, nil)
// 	if err != nil {
// 		// return hex for a new tx
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	return result, nil
// }

func (httpServer *HttpServer) handleStallInsert(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	if config.Config().Network() != "mainnet" {
		arrayParams := common.InterfaceSlice(params)
		chainID := int(arrayParams[0].(float64))
		httpServer.config.Syncker.ShardSyncProcess[chainID].Stall()
	}
	return nil, nil
}

func (httpServer *HttpServer) handleResetCache(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	n := int(arrayParams[0].(float64))
	switch n {
	case 1:
		incognitokey.GetMiningKeyBase58Cache, _ = lru.New(2000)
		log.Println("reset GetMiningKeyBase58Cache")
	case 2:
		incognitokey.ToBase58Cache, _ = lru.New(2000)
		log.Println("reset ToBase58Cache")
	case 3:
		blsmultisig.Cacher = cache.New(4*time.Hour, 4*time.Hour)
		log.Println("reset blsmultisig.Cacher")
	default:
	}
	return "ok", nil
}

func (httpServer *HttpServer) handleTestValidate(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	log.Println(arrayParams)
	aggSigHash, err := base64.StdEncoding.DecodeString(arrayParams[0].(string))
	if err != nil {
		return "Cannot decode aggsig", nil
	}

	blkHash := common.Hash{}.NewHashFromStr2(arrayParams[1].(string))
	valIdx := make([]int, len(arrayParams[2].([]interface{})))
	for id, x := range arrayParams[2].([]interface{}) {
		valIdx[id] = int(x.(float64))
	}
	committeeFromBlockHash := common.Hash{}.NewHashFromStr2(arrayParams[3].(string))
	chainID := int(arrayParams[4].(float64))
	proposerIndex := int(arrayParams[5].(float64))
	blockVersion := int(arrayParams[6].(float64))
	var committee []incognitokey.CommitteePublicKey
	if chainID == -1 {
		committee = httpServer.GetBlockchain().BeaconChain.GetCommittee()
	} else {
		committee, _ = httpServer.GetBlockchain().GetShardCommitteeFromBeaconHash(committeeFromBlockHash, byte(chainID))
		committee = httpServer.GetBlockchain().ShardChain[chainID].GetSigningCommittees(proposerIndex, committee, blockVersion)
	}

	committeeStr, _ := incognitokey.CommitteeKeyListToString(committee)
	committeeBLSKeys := []blsmultisig.PublicKey{}
	for _, member := range committee {
		committeeBLSKeys = append(committeeBLSKeys, member.MiningPubKey[common.BlsConsensus])
	}

	log.Printf("Sig %+v\n BlockHash %+v \n Valindex %+v \n Committee %+v \n", aggSigHash, blkHash.GetBytes(), valIdx, committeeBLSKeys)

	if ok, err := blsmultisig.Verify(aggSigHash, blkHash.GetBytes(), valIdx, committeeBLSKeys); !ok {
		return fmt.Sprintf("Invalid Signature: aggSig %+v blkHash: %+v, valIdx %+v, committeeStr %+v %+v", aggSigHash, blkHash.String(), valIdx, committeeStr, err), nil
	}

	return "ok", nil
}
