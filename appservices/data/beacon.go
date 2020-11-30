package data

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type CommitteeKeySetAutoStake struct {
	IncPubKey    							string						`json:"IncPubKey"`
	MiningPubKey 							map[string]string			`json:"MiningPubKey"`
	IsAutoStake  							bool						`json:"IsAutoStake"`
}

type PDEShareState struct {
	Token1ID           string
	Token2ID           string
	ContributorAddress string
	Amount             uint64
}

type PDETradingFeeState struct {
	Token1ID           string
	Token2ID           string
	ContributorAddress string
	Amount             uint64
}

type PDEPoolPairState struct {
	Token1ID        string
	Token1PoolValue uint64
	Token2ID        string
	Token2PoolValue uint64
}

type WaitingPDEContributionState struct {
	PairID             string
	ContributorAddress string
	TokenID            string
	Amount             uint64
	TXReqID            common.Hash
}

type CustodianState struct {
	IncognitoAddress       string
	TotalCollateral        uint64            // prv
	FreeCollateral         uint64            // prv
	HoldingPubTokens       map[string]uint64 // tokenID : amount
	LockedAmountCollateral map[string]uint64 // tokenID : amount
	RemoteAddresses        map[string]string // tokenID : remote address
	RewardAmount           map[string]uint64 // tokenID : amount
}

type WaitingPortingRequest struct {
	UniquePortingID string
	TokenID         string
	PorterAddress   string
	Amount          uint64
	Custodians      []statedb.MatchingPortingCustodianDetail
	PortingFee      uint64
	BeaconHeight    uint64
	TXReqID         common.Hash
}

type FinalExchangeRatesState struct {
	Rates map[string]statedb.FinalExchangeRatesDetail
}

type RedeemRequest struct {
	UniqueRedeemID        string
	TokenID               string
	RedeemerAddress       string
	RedeemerRemoteAddress string
	RedeemAmount          uint64
	Custodians            []MatchingRedeemCustodianDetail
	RedeemFee             uint64
	BeaconHeight          uint64
	TXReqID               common.Hash
}

type MatchingRedeemCustodianDetail struct {
	IncAddress    string
	RemoteAddress string
	Amount        uint64
}

type LockedCollateralState struct {
	TotalLockedCollateralForRewards uint64
	LockedCollateralDetail          map[string]uint64 // custodianAddress : amount
}

type Beacon struct {
	ShardID									int							`json:"ShardID"`
	BlockHash 								string					 	`json:"BlockHash"`
	PreviousBlockHash 						string					 	`json:"PreviousBlockHash"`
	BestShardHash 							map[byte]string			 	`json:"BestShardHash"`
	BestShardHeight     					map[byte]uint64          	`json:"BestShardHeight"`
	Epoch									uint64					 					`json:"Epoch"`
	Height									uint64					 					`json:"Height"`
	ProposerIndex							int                                         `json:"ProposerIndex"`
	BeaconCommittee                        	[]incognitokey.CommitteeKeyString         `json:"BeaconCommittee"`
	BeaconPendingValidator                 	[]incognitokey.CommitteeKeyString          `json:"BeaconPendingValidator"`
	CandidateBeaconWaitingForCurrentRandom 	[]incognitokey.CommitteeKeyString          `json:"CandidateBeaconWaitingForCurrentRandom"`
	CandidateShardWaitingForCurrentRandom  	[]incognitokey.CommitteeKeyString           `json:"CandidateShardWaitingForCurrentRandom"` // snapshot shard candidate list, waiting to be shuffled in this current epoch
	CandidateBeaconWaitingForNextRandom    	[]incognitokey.CommitteeKeyString         `json:"CandidateBeaconWaitingForNextRandom"`
	CandidateShardWaitingForNextRandom     	[]incognitokey.CommitteeKeyString          `json:"CandidateShardWaitingForNextRandom"` // shard candidate list, waiting to be shuffled in next epoch
	ShardCommittee                         	map[byte][]incognitokey.CommitteeKeyString `json:"ShardCommittee"`        // current committee and validator of all shard
	ShardPendingValidator                  	map[byte][]incognitokey.CommitteeKeyString  `json:"ShardPendingValidator"` // pending candidate waiting for swap to get in committee of all shard
	AutoStaking                            	[]CommitteeKeySetAutoStake                `json:"AutoStaking"`
	CurrentRandomNumber                    	int64                                     `json:"CurrentRandomNumber"`
	CurrentRandomTimeStamp                 	int64                                      `json:"CurrentRandomTimeStamp"` // random timestamp for this epoch
	MaxBeaconCommitteeSize                 	int                                        `json:"MaxBeaconCommitteeSize"`
	MinBeaconCommitteeSize                 	int                                        `json:"MinBeaconCommitteeSize"`
	MaxShardCommitteeSize                  	int                                         `json:"MaxShardCommitteeSize"`
	MinShardCommitteeSize                  	int                                         `json:"MinShardCommitteeSize"`
	ActiveShards                           	int                                          `json:"ActiveShards"`
	LastCrossShardState                    	map[byte]map[byte]uint64                  `json:"LastCrossShardState"`
	Time                					int64                                        `json:"Time"`
	ConsensusAlgorithm                     	string                      `json:"ConsensusAlgorithm"`
	ShardConsensusAlgorithm                	map[byte]string             `json:"ShardConsensusAlgorithm"`
	Instruction								[][]string				 	`json:"Instruction"`
	BridgeToken								[]*rawdbv2.BridgeTokenInfo	`json:"BridgeToken"`
	WaitingPDEContributionState				[]WaitingPDEContributionState	`json:"WaitingPDEContribution"`
	PDEPoolPair								[]PDEPoolPairState				`json:"PDEPoolPair"`
	PDEShare								[]PDEShareState					`json:"PDEShare"`
	PDETradingFee							[]PDETradingFeeState			`json:"PDETradingFee"`
	Custodian 								[]CustodianState				`json:"Custodian"`
	WaitingPortingRequest					[]WaitingPortingRequest		`json:"WaitingPortingRequest"` //key uniquePortingID
	FinalExchangeRates						FinalExchangeRatesState		`json:"FinalExchangeRates"`
	WaitingRedeemRequest					[]RedeemRequest					`json:"WaitingRedeemRequest"`
	MatchedRedeemRequest					[]RedeemRequest					`json:"MatchedRedeemRequest"`
	LockedCollateralState					LockedCollateralState			`json:"LockedCollateralState"`

}

func NewBeaconFromBeaconState(data *blockchain.BeaconBestState) *Beacon {
	result := &Beacon{
		ShardID:				256, //fake ShardID for beacon for load balancing
		BlockHash:          	data.BestBlockHash.String(),
		PreviousBlockHash:  	data.PreviousBestBlockHash.String(),
		BestShardHash: 			newBestShardHashFromBeaconState(data.BestShardHash),
		BestShardHeight: 		newBestShardHeightFromBeaconState(data.BestShardHeight),
		Epoch:                  data.Epoch,
		Height:         		data.BeaconHeight,
		ProposerIndex:	    	data.BeaconProposerIndex,
		BeaconCommittee:		incognitokey.CommitteeKeyListToStringList(data.BeaconCommittee),
		BeaconPendingValidator: incognitokey.CommitteeKeyListToStringList(data.BeaconPendingValidator),
		CandidateBeaconWaitingForCurrentRandom: incognitokey.CommitteeKeyListToStringList(data.CandidateBeaconWaitingForCurrentRandom),
		CandidateShardWaitingForCurrentRandom: incognitokey.CommitteeKeyListToStringList(data.CandidateShardWaitingForCurrentRandom),
		CandidateBeaconWaitingForNextRandom: incognitokey.CommitteeKeyListToStringList(data.CandidateBeaconWaitingForNextRandom),
		CandidateShardWaitingForNextRandom: incognitokey.CommitteeKeyListToStringList(data.CandidateShardWaitingForNextRandom),
		ShardCommittee: 		newShardCommitteeFromBeaconState(data.ShardCommittee),
		ShardPendingValidator: 	newShardPendingValidatorFromBeaconState(data.ShardPendingValidator),
		AutoStaking: 			newAutoStackingFromBeaconState(data.AutoStaking),
		CurrentRandomNumber:    data.CurrentRandomNumber,
		CurrentRandomTimeStamp: data.CurrentRandomTimeStamp,
		MaxShardCommitteeSize:   data.MaxShardCommitteeSize,
		MinShardCommitteeSize:   data.MinShardCommitteeSize,
		MaxBeaconCommitteeSize:  data.MaxBeaconCommitteeSize,
		MinBeaconCommitteeSize:  data.MinBeaconCommitteeSize,
		ActiveShards:            data.ActiveShards,
		LastCrossShardState:     newLastCrossShardFromBeaconState(data.LastCrossShardState),
		Time:                    data.GetBlockTime(),
		ConsensusAlgorithm:      data.ConsensusAlgorithm,
		ShardConsensusAlgorithm: newShardConsensusAlgorithmFromBeaconState(data.ShardConsensusAlgorithm),
		Instruction:             newInstructionFromBeaconState(data.BestBlock.GetInstructions()),
		BridgeToken:             getBridgeTokenInfoFromBeaconState(data.GetBeaconFeatureStateDB()),
		WaitingPDEContributionState: getWaitingPDEContributionStateFromBeaconState(data.GetBeaconFeatureStateDB()),
		PDEPoolPair: getPDEPoolPairFromBeaconState(data.GetBeaconFeatureStateDB()),
		PDEShare: getPDEShareFromBeaconState(data.GetBeaconFeatureStateDB()),
		PDETradingFee: getDETradingFeeFromBeaconState(data.GetBeaconFeatureStateDB()),
		Custodian: getCustodianStateFromBeaconState(data.GetBeaconFeatureStateDB()),
		WaitingPortingRequest: getWaitingPortingRequestFromBeaconState(data.GetBeaconFeatureStateDB()),
		WaitingRedeemRequest: getAllWaitingRedeemRequestFromBeaconState(data.GetBeaconFeatureStateDB()),
		FinalExchangeRates: getFinalExchangeRatesFromBeaconState(data.GetBeaconFeatureStateDB()),
		MatchedRedeemRequest: getAllMatchedRedeemRequestFromBeaconState(data.GetBeaconFeatureStateDB()),
		LockedCollateralState: getLockedCollateralStateFromBeaconState(data.GetBeaconFeatureStateDB()),
	}
	return result
}

func newShardCommitteeFromBeaconState(committee map[byte][]incognitokey.CommitteePublicKey ) map[byte][]incognitokey.CommitteeKeyString {
	shardCommittee := make(map[byte][]incognitokey.CommitteeKeyString)
	for shardID, v := range committee {
		shardCommittee[shardID] = make([]incognitokey.CommitteeKeyString, len(v))
		tempV := incognitokey.CommitteeKeyListToStringList(v)
		copy(shardCommittee[shardID], tempV)
	}
	return shardCommittee
}

func newShardPendingValidatorFromBeaconState(pendingValidator map[byte][]incognitokey.CommitteePublicKey)  map[byte][]incognitokey.CommitteeKeyString {
	shardPendingValidator := make(map[byte][]incognitokey.CommitteeKeyString)
	for shardID, v := range pendingValidator {
		shardPendingValidator[shardID] = make([]incognitokey.CommitteeKeyString, len(v))
		tempV := incognitokey.CommitteeKeyListToStringList(v)
		copy(shardPendingValidator[shardID], tempV)
	}
	return shardPendingValidator
}

func newLastCrossShardFromBeaconState (crossShardState map[byte]map[byte]uint64 ) map[byte]map[byte]uint64 {
	lastCrossShardState := make(map[byte]map[byte]uint64)
	for k1, v1 := range crossShardState {
		lastCrossShardState[k1] = make(map[byte]uint64)
		for k2, v2 := range v1 {
			lastCrossShardState[k1][k2] = v2
		}
	}
	return lastCrossShardState
}

func newAutoStackingFromBeaconState(staking *blockchain.MapStringBool) []CommitteeKeySetAutoStake {
	autoStacking := make([]CommitteeKeySetAutoStake,0)
	for k, v := range staking.GetMap() {
		var keySet incognitokey.CommitteePublicKey
		keySet.FromString(k)
		var keyMap CommitteeKeySetAutoStake
		keyMap.IncPubKey = keySet.GetIncKeyBase58()
		keyMap.MiningPubKey = make(map[string]string)
		for keyType := range keySet.MiningPubKey {
			keyMap.MiningPubKey[keyType] = keySet.GetMiningKeyBase58(keyType)
		}
		keyMap.IsAutoStake = v
		autoStacking = append(autoStacking, keyMap)
	}
	return autoStacking
}

func newBestShardHashFromBeaconState(shardHash map[byte]common.Hash ) map[byte]string{
	bestShardHash := make(map[byte]string)
	for sharID, v := range shardHash {
		bestShardHash[sharID] = v.String()
	}
	return bestShardHash
}

func newBestShardHeightFromBeaconState(shardHeight map[byte]uint64) map[byte]uint64 {
	bestShardHeight := make(map[byte]uint64)
	for shardID, v := range shardHeight {
		bestShardHeight[shardID] = v
	}
	return bestShardHeight
}

func newShardConsensusAlgorithmFromBeaconState(consensusAlgorithm  map[byte]string) map[byte]string  {
	shardConsensusAlgorithm := make(map[byte]string)
	for shardID, v := range consensusAlgorithm {
		shardConsensusAlgorithm[shardID] = v
	}
	return shardConsensusAlgorithm
}

func newInstructionFromBeaconState(instructions [][]string) [][]string{
		dest := make ([][]string, 0, len(instructions))
		for _, inst := range instructions {
			i := make([]string, len(inst))
			copy(i, inst)
			dest = append(dest, i)
		}
		return dest
}

func getBridgeTokenInfoFromBeaconState(bridgeStateDB *statedb.StateDB)  []*rawdbv2.BridgeTokenInfo{
	allBridgeTokens := []*rawdbv2.BridgeTokenInfo{}
	allBridgeTokensBytes, err := statedb.GetAllBridgeTokens(bridgeStateDB)
	if err != nil {
		return allBridgeTokens
	}
	json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
	return allBridgeTokens
}

func getWaitingPDEContributionStateFromBeaconState(PDEStateDB *statedb.StateDB) []WaitingPDEContributionState {
	waitingPDEContributionState := PDEStateDB.GetAllWaitingPDEContributionState()
	if len(waitingPDEContributionState) == 0 {
		return []WaitingPDEContributionState{}
	}

	newWaitingPDEContributionState := make([]WaitingPDEContributionState, 0, len(waitingPDEContributionState))
	for _, waiting := range waitingPDEContributionState {
		newWaitingPDEContributionState = append(newWaitingPDEContributionState, WaitingPDEContributionState{
			PairID:             waiting.PairID(),
			ContributorAddress: waiting.ContributorAddress(),
			TokenID:            waiting.TokenID(),
			Amount:             waiting.Amount(),
			TXReqID:            waiting.TxReqID(),
		})
	}
	return newWaitingPDEContributionState
}

func getPDEPoolPairFromBeaconState(PDEStateDB *statedb.StateDB) []PDEPoolPairState {
	pdePoolPairState := PDEStateDB.GetAllPDEPoolPairState()
	if len(pdePoolPairState) == 0 {
		return []PDEPoolPairState{}
	}

	newPDEPoolPairState := make([]PDEPoolPairState, 0, len(pdePoolPairState))
	for _, pool := range pdePoolPairState {
		newPDEPoolPairState = append(newPDEPoolPairState, PDEPoolPairState{
			Token1ID:        pool.Token1ID(),
			Token1PoolValue: pool.Token1PoolValue(),
			Token2ID:        pool.Token2ID(),
			Token2PoolValue: pool.Token2PoolValue(),
		})
	}
	return newPDEPoolPairState
}

func getPDEShareFromBeaconState(PDEStateDB *statedb.StateDB) []PDEShareState {
	pdeShareState := PDEStateDB.GetAllPDEShareState()
	if len(pdeShareState) == 0 {
		return []PDEShareState{}
	}

	newPDEShareState:= make([]PDEShareState, 0, len(pdeShareState))
	for _, share := range pdeShareState {
		newPDEShareState = append(newPDEShareState, PDEShareState{
			Token1ID:           share.Token1ID(),
			Token2ID:           share.Token2ID(),
			ContributorAddress: share.ContributorAddress(),
			Amount:             share.Amount(),
		})
	}
	return newPDEShareState
}

func getDETradingFeeFromBeaconState(PDEStateDB *statedb.StateDB) []PDETradingFeeState {
	pdeTradingFee := PDEStateDB.GetAllPDETradingFeeState()
	if len(pdeTradingFee) == 0 {
		return []PDETradingFeeState{}
	}

	newPDETradingFee := make([]PDETradingFeeState, 0, len(pdeTradingFee))
	for _, fee := range pdeTradingFee {
		newPDETradingFee = append(newPDETradingFee, PDETradingFeeState{
			Token1ID:           fee.Token1ID(),
			Token2ID:           fee.Token2ID(),
			ContributorAddress: fee.ContributorAddress(),
			Amount:             fee.Amount(),
		})
	}
	return newPDETradingFee
}

func getCustodianStateFromBeaconState(PortalStateDB *statedb.StateDB) []CustodianState {
	custodianPool := PortalStateDB.GetAllCustodianStatePool()
	if len(custodianPool) == 0 {
		return []CustodianState{}
	}
	custodians  := make([]CustodianState,0, len(custodianPool))
	for _, custodian := range custodianPool {
			custodians = append(custodians, CustodianState{
				IncognitoAddress:       custodian.GetIncognitoAddress(),
				TotalCollateral:        custodian.GetTotalCollateral(),
				FreeCollateral:         custodian.GetFreeCollateral(),
				HoldingPubTokens:       custodian.GetHoldingPublicTokens(),
				LockedAmountCollateral: custodian.GetLockedAmountCollateral(),
				RemoteAddresses:        custodian.GetRemoteAddresses(),
				RewardAmount:           custodian.GetRewardAmount(),
			})
	}
	return custodians
}

func getWaitingPortingRequestFromBeaconState(PortalStateDB *statedb.StateDB) []WaitingPortingRequest {
	waitingPortalRequestMap := PortalStateDB.GetWaitingPortingRequests()
	if len(waitingPortalRequestMap) == 0 {
		return []WaitingPortingRequest{}
	}
	waitingPortalRequest := make([]WaitingPortingRequest, 0, len(waitingPortalRequestMap))
	for _, waiting := range waitingPortalRequestMap {
		waitingPortalRequest = append(waitingPortalRequest, WaitingPortingRequest{
			UniquePortingID: waiting.UniquePortingID(),
			TokenID:         waiting.TokenID(),
			PorterAddress:   waiting.PorterAddress(),
			Amount:          waiting.Amount(),
			Custodians:      getCustodianOfWaitingPortingRequest(waiting.Custodians()),
			PortingFee:      waiting.PortingFee(),
			BeaconHeight:    waiting.BeaconHeight(),
			TXReqID:         common.Hash{},
		})
	}
	return waitingPortalRequest
}

func getCustodianOfWaitingPortingRequest(custodians      []*statedb.MatchingPortingCustodianDetail) []statedb.MatchingPortingCustodianDetail {
	custo :=  []statedb.MatchingPortingCustodianDetail{}
	for _, custodian := range custodians {
		custo = append(custo, statedb.MatchingPortingCustodianDetail{
			IncAddress:             custodian.IncAddress,
			RemoteAddress:          custodian.RemoteAddress,
			Amount:                 custodian.Amount,
			LockedAmountCollateral: custodian.LockedAmountCollateral,
		})
	}
	return custo
}

func getAllWaitingRedeemRequestFromBeaconState (PortalStateDB *statedb.StateDB) []RedeemRequest {
	allWaitingRedeemRequest := PortalStateDB.GetAllWaitingRedeemRequest()
	if len(allWaitingRedeemRequest) == 0 {
		return []RedeemRequest{}
	}
	waitingRedeemRequest := make([]RedeemRequest, 0, len(allWaitingRedeemRequest))
	for _, redeem := range allWaitingRedeemRequest {
		waitingRedeemRequest = append(waitingRedeemRequest, RedeemRequest{
			UniqueRedeemID:        redeem.GetUniqueRedeemID(),
			TokenID:               redeem.GetTokenID(),
			RedeemerAddress:       redeem.GetRedeemerAddress(),
			RedeemerRemoteAddress: redeem.GetRedeemerRemoteAddress(),
			RedeemAmount:          redeem.GetRedeemAmount(),
			Custodians:            getCustodianOfRedeemRequest(redeem.GetCustodians()),
			RedeemFee:             redeem.GetRedeemFee(),
			BeaconHeight:          redeem.GetBeaconHeight(),
			TXReqID:               redeem.GetTxReqID(),
		})
	}
	return waitingRedeemRequest
}

func getCustodianOfRedeemRequest(custodians  []*statedb.MatchingRedeemCustodianDetail) []MatchingRedeemCustodianDetail {
	custo :=  []MatchingRedeemCustodianDetail{}
	for _, custodian := range custodians {
		custo = append(custo, MatchingRedeemCustodianDetail{
			IncAddress:    custodian.GetIncognitoAddress(),
			RemoteAddress: custodian.GetRemoteAddress(),
			Amount:        custodian.GetAmount(),
		})
	}
	return custo
}


func getAllMatchedRedeemRequestFromBeaconState (PortalStateDB *statedb.StateDB) []RedeemRequest {
	allMatchedRedeemRequest := PortalStateDB.GetAllMatchedRedeemRequest()
	if len(allMatchedRedeemRequest) == 0 {
		return []RedeemRequest{}
	}
	matchedRedeemRequest := make([]RedeemRequest, 0, len(allMatchedRedeemRequest))
	for _, redeem := range allMatchedRedeemRequest {
		matchedRedeemRequest = append(matchedRedeemRequest, RedeemRequest{
			UniqueRedeemID:        redeem.GetUniqueRedeemID(),
			TokenID:               redeem.GetTokenID(),
			RedeemerAddress:       redeem.GetRedeemerAddress(),
			RedeemerRemoteAddress: redeem.GetRedeemerRemoteAddress(),
			RedeemAmount:          redeem.GetRedeemAmount(),
			Custodians:            getCustodianOfRedeemRequest(redeem.GetCustodians()),
			RedeemFee:             redeem.GetRedeemFee(),
			BeaconHeight:          redeem.GetBeaconHeight(),
			TXReqID:               redeem.GetTxReqID(),
		})
	}
	return matchedRedeemRequest
}

func getFinalExchangeRatesFromBeaconState(PortalStateDB *statedb.StateDB) FinalExchangeRatesState {
	finalExchangeRatesState, err := statedb.GetFinalExchangeRatesState(PortalStateDB)
	if err != nil  {
		return FinalExchangeRatesState{Rates: make(map[string]statedb.FinalExchangeRatesDetail)}
	}
	details := make( map[string]statedb.FinalExchangeRatesDetail)
	if len(finalExchangeRatesState.Rates()) > 0 {
		for key, val := range finalExchangeRatesState.Rates() {
			details[key] = val
		}
	}
	return FinalExchangeRatesState{Rates: details}
}


func getLockedCollateralStateFromBeaconState(PortalStateDB *statedb.StateDB) LockedCollateralState {
	lockedCollateralState, err := statedb.GetLockedCollateralStateByBeaconHeight(PortalStateDB)
	if err != nil {
		return LockedCollateralState{
			TotalLockedCollateralForRewards: 0,
			LockedCollateralDetail:          make(map[string]uint64),
		}
	}

	details := make(map[string]uint64 )

	if len(lockedCollateralState.GetLockedCollateralDetail()) > 0 {
		for key, val := range lockedCollateralState.GetLockedCollateralDetail() {
			details[key] = val
		}
	}
	return  LockedCollateralState{
		TotalLockedCollateralForRewards: lockedCollateralState.GetTotalLockedCollateralForRewards(),
		LockedCollateralDetail:          details,
	}
}