package model

import "github.com/incognitochain/incognito-chain/incognitokey"
import "github.com/incognitochain/incognito-chain/appservices/data"

type BeaconState struct {
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
	AutoStaking                            	[]data.CommitteeKeySetAutoStake                `json:"AutoStaking"`
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
}
