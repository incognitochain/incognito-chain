package rpcservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
)

func (blockService BlockService) GetPdexv3ParamsModifyingRequestStatus(reqTxID string) (*metadataPdexv3.ParamsModifyingRequestStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3ParamsModifyingStatusPrefix(),
		[]byte(reqTxID),
	)
	if err != nil {
		return nil, err
	}

	var status metadataPdexv3.ParamsModifyingRequestStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

// paramSelector helps to wrap transaction creation steps (v2 only)
type paramSelector struct {
	TokenID        common.Hash
	PRV            *bean.CreateRawTxParam
	Token          *bean.CreateRawPrivacyTokenTxParam
	TokenReceivers []*privacy.PaymentInfo
	Metadata       metadataCommon.Metadata
}

// SetTokenID, SetTokenReceivers, SetMetadata add necessary data for token tx creation that are missing from the params struct
func (sel *paramSelector) SetTokenID(id common.Hash)                  { sel.TokenID = id }
func (sel *paramSelector) SetTokenReceivers(r []*privacy.PaymentInfo) { sel.TokenReceivers = r }
func (sel *paramSelector) SetMetadata(md metadataCommon.Metadata)     { sel.Metadata = md }

// PdexTxService extends TxService with wrappers to build TX with cleaner syntax
type PdexTxService struct {
	*TxService
}

func (svc PdexTxService) ReadParamsFrom(raw interface{}, metadataReader interface{}) (*paramSelector, error) {
	var err error
	// token id defaults to PRV
	sel := paramSelector{TokenID: common.PRVCoinID}
	sel.PRV, err = bean.NewCreateRawTxParam(raw)
	if err != nil {
		return nil, err
	}
	sel.Token, err = bean.NewCreateRawPrivacyTokenTxParam(raw)
	if err != nil {
		return nil, err
	}

	arrayParams := common.InterfaceSlice(raw)
	if len(arrayParams) >= 5 {
		var rawMd []byte
		rawMd, err = json.Marshal(arrayParams[4])
		if err != nil {
			return nil, fmt.Errorf("Cannot parse metadata - %v", err)
		}
		err = json.Unmarshal(rawMd, metadataReader)
	}

	return &sel, err
}

func (svc PdexTxService) ValidateTokenIDs(tokenToSell, tokenToBuy *common.Hash) error {
	if tokenToSell == nil || tokenToSell.IsZeroValue() {
		return fmt.Errorf("Invalid TokenToSell %v", tokenToSell)
	}
	if tokenToBuy == nil || tokenToBuy.IsZeroValue() {
		return fmt.Errorf("Invalid TokenToBuy %v", tokenToBuy)
	}
	return nil
}

func (svc PdexTxService) BuildTransaction(
	sel *paramSelector, md metadataCommon.Metadata,
) (metadataCommon.Transaction, *RPCError) {
	if sel.TokenID == common.PRVCoinID {
		return svc.BuildRawTransaction(sel.PRV, md)
	} else {
		return buildTokenTransaction(svc, sel)
	}
}

func (svc PdexTxService) GenerateOTAReceivers(
	tokens []common.Hash, addr privacy.PaymentAddress,
) (map[common.Hash]privacy.OTAReceiver, error) {
	result := map[common.Hash]privacy.OTAReceiver{}
	var err error
	for _, tokenID := range tokens {
		temp := privacy.OTAReceiver{}
		err = temp.FromAddress(addr)
		if err != nil {
			return nil, err
		}
		result[tokenID] = temp
	}
	return result, nil
}

func buildTokenTransaction(svc PdexTxService, sel *paramSelector) (metadataCommon.Transaction, *RPCError) {
	params := sel.Token

	// choose token inputs
	outputTokens, err := svc.BlockChain.TryGetAllOutputCoinsByKeyset(
		params.SenderKeySet, params.ShardIDSender, &sel.TokenID, true,
	)
	if err != nil {
		return nil, NewRPCError(GetOutputCoinError, err)
	}
	outputTokens, err = svc.filterMemPoolOutcoinsToSpent(outputTokens)
	if err != nil {
		return nil, NewRPCError(GetOutputCoinError, err)
	}

	var totalTokenTransferred uint64
	for _, payment := range sel.TokenReceivers {
		totalTokenTransferred += payment.Amount
	}
	candidateOutputTokens, _, _, err := svc.chooseBestOutCoinsToSpent(
		outputTokens, totalTokenTransferred,
	)
	if err != nil {
		return nil, NewRPCError(GetOutputCoinError, err)
	}

	tokenParams := &transaction.TokenParam{
		TokenTxType: int(transaction.CustomTokenTransfer),
		// amount will default to sum of payments
		Amount:     0,
		Receiver:   sel.TokenReceivers,
		Fee:        0,
		TokenInput: candidateOutputTokens,
		PropertyID: sel.TokenID.String(),
	}

	// choose PRV inputs
	inputCoins, realFeePRV, err1 := svc.chooseOutsCoinByKeyset(params.PaymentInfos,
		params.EstimateFeeCoinPerKb, 0, params.SenderKeySet,
		params.ShardIDSender, params.HasPrivacyCoin, nil, tokenParams)
	if err1 != nil {
		return nil, err1
	}
	if len(params.PaymentInfos) == 0 && realFeePRV == 0 {
		params.HasPrivacyCoin = false
	}

	// tx creation params
	txTokenParams := transaction.NewTxTokenParams(&params.SenderKeySet.PrivateKey,
		params.PaymentInfos,
		inputCoins,
		realFeePRV,
		tokenParams,
		svc.BlockChain.GetBestStateShard(params.ShardIDSender).GetCopiedTransactionStateDB(),
		sel.Metadata,
		params.HasPrivacyCoin,
		params.HasPrivacyToken,
		params.ShardIDSender, params.Info,
		svc.BlockChain.BeaconChain.GetFinalViewState().GetBeaconFeatureStateDB(),
	)

	tx := &transaction.TxTokenVersion2{}
	errTx := tx.Init(txTokenParams)
	if errTx != nil {
		return nil, NewRPCError(CreateTxDataError, errTx)
	}
	return tx, nil
}

func (blockService BlockService) GetPdexv3WithdrawalLPFeeStatus(reqTxID string) (*metadataPdexv3.WithdrawalLPFeeStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawalLPFeeStatusPrefix(),
		[]byte(reqTxID),
	)
	if err != nil {
		return nil, err
	}

	var status metadataPdexv3.WithdrawalLPFeeStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func (blockService BlockService) GetPdexv3WithdrawalProtocolFeeStatus(reqTxID string) (*metadataPdexv3.WithdrawalProtocolFeeStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawalProtocolFeeStatusPrefix(),
		[]byte(reqTxID),
	)
	if err != nil {
		return nil, err
	}

	var status metadataPdexv3.WithdrawalProtocolFeeStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func (blockService BlockService) GetPdexv3WithdrawalStakingRewardStatus(reqTxID string) (*metadataPdexv3.WithdrawalStakingRewardStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawalStakingRewardStatusPrefix(),
		[]byte(reqTxID),
	)
	if err != nil {
		return nil, err
	}

	var status metadataPdexv3.WithdrawalStakingRewardStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func (blockService BlockService) GetPdexv3State(
	filterParam map[string]interface{},
	beaconHeight uint64,
) (interface{}, error) {
	beaconBestView := blockService.BlockChain.GetBeaconBestState()
	if beaconHeight == 0 {
		beaconHeight = beaconBestView.BeaconHeight
	}
	if uint64(beaconHeight) < config.Param().PDexParams.Pdexv3BreakPointHeight {
		return nil, NewRPCError(GetPdexv3StateError, fmt.Errorf("pDEX v3 is not available"))
	}

	beaconFeatureStateRootHash, err := blockService.BlockChain.GetBeaconFeatureRootHash(beaconBestView, uint64(beaconHeight))
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, fmt.Errorf("Can't found ConsensusStateRootHash of beacon height %+v, error %+v", beaconHeight, err))
	}
	beaconFeatureStateDB, err := statedb.NewWithPrefixTrie(beaconFeatureStateRootHash, statedb.NewDatabaseAccessWarper(blockService.BlockChain.GetBeaconChainDatabase()))
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}

	beaconBlocks, err := blockService.BlockChain.GetBeaconBlockByHeight(uint64(beaconHeight))
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}
	beaconBlock := beaconBlocks[0]
	beaconTimeStamp := beaconBlock.Header.Timestamp

	var res interface{}
	type FilterParam struct {
		Key       string `json:"Key"`
		Verbosity uint   `json:"Verbosity"`
		ID        string `json:"ID"`
	}
	param := FilterParam{}
	data, err := json.Marshal(filterParam)
	if err != nil {
		return res, NewRPCError(GetPdexv3StateError, err)
	}
	err = json.Unmarshal(data, &param)
	if err != nil {
		return res, NewRPCError(GetPdexv3StateError, err)
	}

	if !reflect.DeepEqual(param, FilterParam{}) {
		switch param.Key {
		case WaitingContributions:
			res, err = getPdexv3WaitingContributions(beaconTimeStamp, beaconFeatureStateDB)
			if err != nil {
				return res, NewRPCError(GetPdexv3StateError, err)
			}

		case PoolPairs:
			res, err = getPdexv3PoolPairs(beaconHeight, beaconTimeStamp, param.Verbosity, beaconFeatureStateDB)
			if err != nil {
				return res, NewRPCError(GetPdexv3StateError, err)
			}

		case PoolPair:
			res, err = getPdexv3PoolPair(beaconHeight, beaconTimeStamp, beaconFeatureStateDB, param.ID)
			if err != nil {
				return res, NewRPCError(GetPdexv3StateError, err)
			}

		case PoolPairShares:
			res, err = getPdexv3PoolPairShares(beaconHeight, beaconTimeStamp, beaconFeatureStateDB, param.ID)
			if err != nil {
				return res, NewRPCError(GetPdexv3StateError, err)
			}

		case PoolPairOrders:
			res, err = getPdexv3PoolPairOrders(beaconHeight, beaconTimeStamp, beaconFeatureStateDB, param.ID)
			if err != nil {
				return res, NewRPCError(GetPdexv3StateError, err)
			}

		case PoolPairOrderRewards:
			res, err = getPdexv3PoolPairOrderReward(beaconHeight, beaconTimeStamp, beaconFeatureStateDB, param.ID)
			if err != nil {
				return res, NewRPCError(GetPdexv3StateError, err)
			}

		case Params:
			res, err = getPdexv3Param(beaconHeight, beaconTimeStamp, beaconFeatureStateDB)
			if err != nil {
				return res, NewRPCError(GetPdexv3StateError, err)
			}

		case StakingPools:
			res, err = getPdexv3StakingPools(beaconHeight, beaconTimeStamp, beaconFeatureStateDB)
			if err != nil {
				return res, NewRPCError(GetPdexv3StateError, err)
			}

		case StakingPool:
			res, err = getPdexv3StakingPool(
				beaconHeight, beaconTimeStamp, beaconFeatureStateDB, param.ID,
			)
			if err != nil {
				return res, NewRPCError(GetPdexv3StateError, err)
			}

		case NftIDs:
			res, err = getPdexv3NftIDs(beaconHeight, beaconTimeStamp, beaconFeatureStateDB)
			if err != nil {
				return res, NewRPCError(GetPdexv3StateError, err)
			}
		case NftIDsCached:
			return beaconBestView.PdeState(pdex.AmplifierVersion).Reader().NftIDs(), nil

		case All:
			res, err = getPdexv3State(beaconHeight, beaconTimeStamp, beaconFeatureStateDB)
			if err != nil {
				return nil, NewRPCError(GetPdexv3StateError, err)
			}

		default:
			return res, NewRPCError(GetPdexv3StateError, errors.New("Can't recognize filter key"))
		}

	} else {
		res, err = getPdexv3State(beaconHeight, beaconTimeStamp, beaconFeatureStateDB)
		if err != nil {
			return nil, NewRPCError(GetPdexv3StateError, err)
		}
	}
	return res, nil
}

func (blockService BlockService) GetPdexv3BlockLPReward(
	pairID string, beaconHeight uint64, stateDB, prevStateDB *statedb.StateDB,
) (map[string]uint64, error) {
	// get accumulated reward to the beacon height
	curLPFeesPerShare, shareAmount, err := getLPFeesPerShare(pairID, beaconHeight, stateDB)
	if err != nil {
		return nil, err
	}
	// get accumulated reward to the previous block of querying beacon height
	oldLPFeesPerShare, _, err := getLPFeesPerShare(pairID, beaconHeight-1, prevStateDB)
	if err != nil {
		oldLPFeesPerShare = map[common.Hash]*big.Int{}
	}

	result := map[string]uint64{}

	for tokenID := range curLPFeesPerShare {
		oldFees, isExisted := oldLPFeesPerShare[tokenID]
		if !isExisted {
			oldFees = big.NewInt(0)
		}
		newFees := curLPFeesPerShare[tokenID]

		reward := new(big.Int).Mul(new(big.Int).Sub(newFees, oldFees), new(big.Int).SetUint64(shareAmount))
		reward = new(big.Int).Div(reward, pdex.BaseLPFeesPerShare)

		if !reward.IsUint64() {
			return nil, fmt.Errorf("Reward of token %v is out of range", tokenID)
		}
		if reward.Uint64() > 0 {
			result[tokenID.String()] = reward.Uint64()
		}
	}

	return result, nil
}

func (blockService BlockService) GetPdexv3BlockStakingReward(
	stakingPoolID string, beaconHeight uint64, stateDB, prevStateDB *statedb.StateDB,
) (map[string]uint64, error) {
	// get accumulated reward to the beacon height
	curRewardsPerShare, shareAmount, err := getStakingRewardsPerShare(stakingPoolID, beaconHeight, stateDB)
	if err != nil {
		return nil, err
	}
	// get accumulated reward to the previous block of querying beacon height
	oldRewardsPerShare, _, err := getStakingRewardsPerShare(stakingPoolID, beaconHeight-1, prevStateDB)
	if err != nil {
		oldRewardsPerShare = map[common.Hash]*big.Int{}
	}

	result := map[string]uint64{}

	for tokenID := range curRewardsPerShare {
		oldFees, isExisted := oldRewardsPerShare[tokenID]
		if !isExisted {
			oldFees = big.NewInt(0)
		}
		newFees := curRewardsPerShare[tokenID]

		reward := new(big.Int).Mul(new(big.Int).Sub(newFees, oldFees), new(big.Int).SetUint64(shareAmount))
		reward = new(big.Int).Div(reward, pdex.BaseLPFeesPerShare)

		if !reward.IsUint64() {
			return nil, fmt.Errorf("Reward of token %v is out of range", tokenID)
		}
		if reward.Uint64() > 0 {
			result[tokenID.String()] = reward.Uint64()
		}
	}

	return result, nil
}

func getPdexv3WaitingContributions(beaconTimeStamp int64, stateDB *statedb.StateDB) (interface{}, error) {
	waitingContributions, err := pdex.InitWaitingContributionsFromDB(stateDB)
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}
	resWaitingContributions := map[string]*rawdbv2.Pdexv3Contribution{}
	for k, v := range waitingContributions {
		temp := new(rawdbv2.Pdexv3Contribution)
		*temp = v
		resWaitingContributions[k] = temp
	}
	res := &jsonresult.Pdexv3State{
		BeaconTimeStamp:      beaconTimeStamp,
		WaitingContributions: &resWaitingContributions,
	}
	return res, nil
}

func getPdexv3State(
	beaconHeight uint64, beaconTimeStamp int64, stateDB *statedb.StateDB,
) (interface{}, error) {
	pDexv3State, err := pdex.InitStateV2FromDBWithoutNftIDs(stateDB, beaconHeight)
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}

	poolPairs := map[string]*pdex.PoolPairState{}
	waitingContributions := map[string]*rawdbv2.Pdexv3Contribution{}
	err = json.Unmarshal(pDexv3State.Reader().WaitingContributions(), &waitingContributions)
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}
	err = json.Unmarshal(pDexv3State.Reader().PoolPairs(), &poolPairs)
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}
	cloneParam := pdex.NewParams()
	*cloneParam = *pDexv3State.Reader().Params()
	nftIDs := pDexv3State.Reader().NftIDs()
	stakingPools := pDexv3State.Reader().StakingPools()

	res := &jsonresult.Pdexv3State{
		BeaconTimeStamp:      beaconTimeStamp,
		Params:               cloneParam,
		PoolPairs:            &poolPairs,
		WaitingContributions: &waitingContributions,
		NftIDs:               &nftIDs,
		StakingPools:         &stakingPools,
	}
	return res, nil
}

func getPdexv3PoolPairs(
	beaconHeight uint64, beaconTimeStamp int64, verbosity uint, stateDB *statedb.StateDB,
) (interface{}, error) {
	switch verbosity {
	case SimpleVerbosity:
		poolPairIDs, err := pdex.InitPoolPairIDsFromDB(stateDB)
		if err != nil {
			return nil, NewRPCError(GetPdexv3StateError, err)
		}
		return poolPairIDs, nil
	case IntermidateVerbosity:
		poolPairStates, err := pdex.InitIntermediatePoolPairStatesFromDB(stateDB)
		if err != nil {
			return nil, NewRPCError(GetPdexv3StateError, err)
		}
		res := &jsonresult.Pdexv3State{
			BeaconTimeStamp: beaconTimeStamp,
			PoolPairs:       &poolPairStates,
		}
		return res, nil
	case FullVerbosity:
		poolPairStates, err := pdex.InitFullPoolPairStatesFromDB(stateDB)
		if err != nil {
			return nil, NewRPCError(GetPdexv3StateError, err)
		}
		res := &jsonresult.Pdexv3State{
			BeaconTimeStamp: beaconTimeStamp,
			PoolPairs:       &poolPairStates,
		}
		return res, nil
	case LiquidityVerbosity:
		poolPairStates, err := pdex.InitLiquidityPoolPairStatesFromDB(stateDB)
		if err != nil {
			return nil, NewRPCError(GetPdexv3StateError, err)
		}
		res := &jsonresult.Pdexv3State{
			BeaconTimeStamp: beaconTimeStamp,
			PoolPairs:       &poolPairStates,
		}
		return res, nil		
	default:
		return nil, NewRPCError(GetPdexv3StateError, errors.New("Can't recognize verbosity"))
	}
}

func getPdexv3Param(
	beaconHeight uint64, beaconTimeStamp int64, stateDB *statedb.StateDB,
) (interface{}, error) {
	param, err := pdex.InitParamFromDB(stateDB)
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}
	res := &jsonresult.Pdexv3State{
		BeaconTimeStamp: beaconTimeStamp,
		Params:          param,
	}
	return res, nil
}

func getPdexv3StakingPools(
	beaconHeight uint64, beaconTimeStamp int64, stateDB *statedb.StateDB,
) (interface{}, error) {
	stakingPools, err := pdex.InitStakingPoolsFromDB(stateDB)
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}
	res := &jsonresult.Pdexv3State{
		BeaconTimeStamp: beaconTimeStamp,
		StakingPools:    &stakingPools,
	}
	return res, nil
}

func getPdexv3NftIDs(
	beaconHeight uint64, beaconTimeStamp int64, stateDB *statedb.StateDB,
) (interface{}, error) {
	nftIDs, err := pdex.InitNftIDsFromDB(stateDB)
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}
	res := &jsonresult.Pdexv3State{
		BeaconTimeStamp: beaconTimeStamp,
		NftIDs:          &nftIDs,
	}
	return res, nil
}

func getPdexv3StakingPool(
	beaconHeight uint64, beaconTimeStamp int64, stateDB *statedb.StateDB, stakingPoolID string,
) (interface{}, error) {
	stakingPoolState, err := pdex.InitStakingPoolFromDB(stateDB, stakingPoolID)
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}
	temp := map[string]*pdex.StakingPoolState{}
	if !reflect.DeepEqual(stakingPoolState, pdex.NewStakingPoolState()) {
		temp[stakingPoolID] = stakingPoolState
	}

	res := &jsonresult.Pdexv3State{
		BeaconTimeStamp: beaconTimeStamp,
		StakingPools:    &temp,
	}
	return res, nil
}

func getPdexv3PoolPair(
	beaconHeight uint64, beaconTimeStamp int64, stateDB *statedb.StateDB, poolPairID string,
) (interface{}, error) {
	poolPairState, err := pdex.InitPoolPair(stateDB, poolPairID)
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}
	res := &jsonresult.Pdexv3State{
		BeaconTimeStamp: beaconTimeStamp,
		PoolPairs: &map[string]*pdex.PoolPairState{
			poolPairID: poolPairState,
		},
	}
	return res, nil
}

func getPdexv3PoolPairShares(
	beaconHeight uint64, beaconTimeStamp int64, stateDB *statedb.StateDB, poolPairID string,
) (interface{}, error) {
	shares, err := pdex.InitPoolPairShares(stateDB, poolPairID)
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}
	temp := map[string]*pdex.PoolPairState{}
	if len(shares) != 0 {
		temp[poolPairID] = pdex.NewPoolPairStateWithValue(
			rawdbv2.Pdexv3PoolPair{}, shares, pdex.Orderbook{}, nil, nil, nil, nil, nil, nil, nil)
	}

	res := &jsonresult.Pdexv3State{
		BeaconTimeStamp: beaconTimeStamp,
		PoolPairs:       &temp,
	}
	return res, nil
}

func getPdexv3PoolPairOrders(
	beaconHeight uint64, beaconTimeStamp int64, stateDB *statedb.StateDB, poolPairID string,
) (interface{}, error) {
	orderBook, err := pdex.InitPoolPairOrders(stateDB, poolPairID)
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}

	res := &jsonresult.Pdexv3State{
		BeaconTimeStamp: beaconTimeStamp,
		PoolPairs: &map[string]*pdex.PoolPairState{
			poolPairID: pdex.NewPoolPairStateWithValue(
				rawdbv2.Pdexv3PoolPair{}, nil, *orderBook, nil, nil, nil, nil, nil, nil, nil),
		},
	}
	return res, nil
}

func getPdexv3PoolPairOrderReward(
	beaconHeight uint64, beaconTimeStamp int64, stateDB *statedb.StateDB, poolPairID string,
) (interface{}, error) {
	orderRewards, err := pdex.InitPoolPairOrderRewards(stateDB, poolPairID)
	if err != nil {
		return nil, NewRPCError(GetPdexv3StateError, err)
	}

	res := &jsonresult.Pdexv3State{
		BeaconTimeStamp: beaconTimeStamp,
		PoolPairs: &map[string]*pdex.PoolPairState{
			poolPairID: pdex.NewPoolPairStateWithValue(
				rawdbv2.Pdexv3PoolPair{}, nil, pdex.Orderbook{}, nil, nil, nil, nil, nil, orderRewards, nil),
		},
	}
	return res, nil
}

func getLPFeesPerShare(
	pairID string, beaconHeight uint64, stateDB *statedb.StateDB,
) (map[common.Hash]*big.Int, uint64, error) {
	// get accumulated reward to the beacon height
	pDexv3State, err := pdex.InitStateFromDB(stateDB, uint64(beaconHeight), pdex.AmplifierVersion)
	if err != nil {
		return nil, 0, err
	}

	poolPairs := make(map[string]*pdex.PoolPairState)
	err = json.Unmarshal(pDexv3State.Reader().PoolPairs(), &poolPairs)
	if err != nil {
		return nil, 0, err
	}

	if _, ok := poolPairs[pairID]; !ok {
		return nil, 0, fmt.Errorf("Pool pair %s not found", pairID)
	}
	pair := poolPairs[pairID]
	pairState := pair.State()

	return pair.LpFeesPerShare(), pairState.ShareAmount(), nil
}

func getStakingRewardsPerShare(
	stakingPoolID string, beaconHeight uint64, stateDB *statedb.StateDB,
) (map[common.Hash]*big.Int, uint64, error) {
	// get accumulated reward to the beacon height
	pDexv3State, err := pdex.InitStateFromDB(stateDB, uint64(beaconHeight), pdex.AmplifierVersion)
	if err != nil {
		return nil, 0, err
	}

	stakingPools := pDexv3State.Reader().StakingPools()

	if _, ok := stakingPools[stakingPoolID]; !ok {
		return nil, 0, fmt.Errorf("Staking pool %s not found", stakingPoolID)
	}

	pool := stakingPools[stakingPoolID].Clone()

	return pool.RewardsPerShare(), pool.Liquidity(), nil
}
