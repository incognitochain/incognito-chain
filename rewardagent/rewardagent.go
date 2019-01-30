package rewardagent

import (
	"github.com/ninjadotorg/constant/blockchain"
)

type RewardAgent struct {
	msgChan chan interface{}
	quit    chan struct{}

	config *RewardAgentConfig
}

type RewardAgentConfig struct {
	BlockChain *blockchain.BlockChain
}

func (rewardAgent RewardAgent) Init(cfg *RewardAgentConfig) (*RewardAgent, error) {
	rewardAgent.config = cfg
	rewardAgent.quit = make(chan struct{})
	rewardAgent.msgChan = make(chan interface{})
	return &rewardAgent, nil
}

func (self *RewardAgent) GetBasicSalary(shardID byte) uint64 {
	// TODO: 0xbahamoot
	// return self.config.BlockChain.BestState.Beacon.BestBlock.Header.(*blockchain.BlockHeaderBeacon).GOVConstitution.GOVParams.BasicSalary
	return 1000
}

func (self *RewardAgent) GetSalaryPerTx(shardID byte) uint64 {
	// TODO: 0xbahamoot
	// return self.config.BlockChain.BestState.Beacon.BestBlock.Header.(*blockchain.BlockHeaderBeacon).GOVConstitution.GOVParams.SalaryPerTx
	return 1000
}

// func getMedians(agentDataPoints []*blockchain.AgentDataPoint) (
// 	float64, float64, float64,
// ) {
// 	agentDataPointsLen := len(agentDataPoints)
// 	if agentDataPointsLen == 0 {
// 		return 0, 0, 0
// 	}
// 	var sumOfCoins float64 = 0
// 	var sumOfBonds float64 = 0
// 	var sumOfTaxs float64 = 0
// 	for _, dataPoint := range agentDataPoints {
// 		sumOfCoins += dataPoint.NumOfCoins
// 		sumOfBonds += dataPoint.NumOfBonds
// 		sumOfTaxs += dataPoint.Tax
// 	}
// 	return float64(sumOfCoins / float64(agentDataPointsLen)), float64(sumOfBonds / float64(agentDataPointsLen)), float64(sumOfTaxs / float64(agentDataPointsLen))
// }

// func calculateReward(
// 	agentDataPoints map[string]*blockchain.AgentDataPoint,
// 	feeMap map[string]uint64,
// ) map[string]uint64 {
// 	if len(agentDataPoints) < NUMBER_OF_MAKING_DECISION_AGENTS {
// 		return map[string]uint64{
// 			"coins": DEFAULT_COINS + feeMap[common.AssetTypeCoin],
// 			"bonds": DEFAULT_BONDS + feeMap[common.AssetTypeBond],
// 		}
// 	}

// 	// group actions by their purpose (ie. issuing or contracting)
// 	issuingCoinsActions := []*blockchain.AgentDataPoint{}
// 	contractingCoinsActions := []*blockchain.AgentDataPoint{}
// 	for _, dataPoint := range agentDataPoints {
// 		if (dataPoint.NumOfCoins > 0 && dataPoint.NumOfBonds > 0) || (dataPoint.NumOfCoins > 0 && dataPoint.Tax > 0) {
// 			continue
// 		}
// 		if dataPoint.NumOfCoins > 0 {
// 			issuingCoinsActions = append(issuingCoinsActions, dataPoint)
// 			continue
// 		}
// 		contractingCoinsActions = append(contractingCoinsActions, dataPoint)
// 	}
// 	if math.Max(float64(len(issuingCoinsActions)), float64(len(contractingCoinsActions))) < (math.Floor(float64(len(agentDataPoints)/2)) + 1) {
// 		return map[string]uint64{
// 			"coins": DEFAULT_COINS + feeMap[common.AssetTypeCoin],
// 			"bonds": DEFAULT_BONDS + feeMap[common.AssetTypeBond],
// 		}
// 	}

// 	if len(issuingCoinsActions) == len(contractingCoinsActions) {
// 		return map[string]uint64{
// 			"coins": DEFAULT_COINS + feeMap[common.AssetTypeCoin],
// 			"bonds": DEFAULT_BONDS + feeMap[common.AssetTypeBond],
// 		}
// 	}

// 	if len(issuingCoinsActions) < len(contractingCoinsActions) {
// 		_, medianBond, medianTax := getMedians(contractingCoinsActions)
// 		coins := uint64(math.Floor((100 - medianTax) * 0.01 * float64(feeMap[common.AssetTypeCoin])))
// 		burnedCoins := feeMap[common.AssetTypeCoin] - coins
// 		bonds := uint64(math.Floor(medianBond)) + feeMap[common.AssetTypeBond] + burnedCoins
// 		return map[string]uint64{
// 			"coins":       coins,
// 			"bonds":       bonds,
// 			"burnedCoins": burnedCoins,
// 		}
// 	}
// 	// issuing coins
// 	medianCoin, _, _ := getMedians(issuingCoinsActions)

// 	return map[string]uint64{
// 		"coins": uint64(math.Floor(medianCoin)) + feeMap[common.AssetTypeCoin],
// 		"bonds": feeMap[common.AssetTypeBond],
// 	}
// }
