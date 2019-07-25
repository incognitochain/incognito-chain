package blockchain

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

//Network fixed params
const (
	// BlockVersion is the current latest supported block version.
	BlockVersion              = 1
	DefaultMaxBlkReqPerPeer   = 600
	DefaultMaxBlkReqPerTime   = 1200
	MinCommitteeSize          = 3                // min size to run bft
	DefaultBroadcastStateTime = 2 * time.Second  // in second
	DefaultStateUpdateTime    = 3 * time.Second  // in second
	DefaultMaxBlockSyncTime   = 1 * time.Second  // in second
	DefaultCacheCleanupTime   = 30 * time.Second // in second
	WorkerNumber              = 5
)

// CONSTANT for network MAINNET
const (
	// ------------- Mainnet ---------------------------------------------
	Mainnet            = 0x01
	MainetName         = "mainnet"
	MainnetDefaultPort = "9333"

	MainNetShardCommitteeSize  = 3
	MainNetBeaconCommitteeSize = 3
	MainNetActiveShards        = 2
	MainNetStakingAmountShard  = 1750000000000 // 1750 PRV = 1750 * 10^9 nano PRV

	//board and proposal parameters
	MainnetBasicReward                = 400000000 //40 mili PRV
	MainnetRewardHalflife             = 3155760   //1 year, reduce 12.5% per year
	MainnetGenesisblockPaymentAddress = "1Uv2zzR4LgfX8ToQe8ub3bYcCLk3uDU1sm9U9hiu9EKYXoS77UdikfT9s8d5YjhsTJm61eazsMwk2otFZBYpPHwiMn8z6bKWWJRspsLky"
	// ------------- end Mainnet --------------------------------------
)

// VARIABLE for mainnet
var (
	MainnetInitPRV = []string{}
	// for beacon
	// public key
	PreSelectBeaconNodeMainnetSerializedPubkey = PreSelectBeaconNodeTestnetSerializedPubkey
	// For shard
	// public key
	PreSelectShardNodeMainnetSerializedPubkey = PreSelectShardNodeTestnetSerializedPubkey
	MaxTxsInBlock                             = 600
	MaxTxsProcessTimeInBlockCreation          = float64(0.85)
	TxsAverageProcessTime                     = int64(5000) // count in nano second ~= 5 mili seconds
	DefaultTxsAverageProcessTime              = int64(5000) // count in nano second
)

// END CONSTANT for network MAINNET

// CONSTANT for network TESTNET
const (
	Testnet            = 0x16
	TestnetName        = "testnet"
	TestnetDefaultPort = "9444"

	TestNetShardCommitteeSize     = 16
	TestNetMinShardCommitteeSize  = 4
	TestNetBeaconCommitteeSize    = 4
	TestNetMinBeaconCommitteeSize = 4
	TestNetActiveShards           = 8
	TestNetStakingAmountShard     = 1750000000000 // 1750 PRV = 1750 * 10^9 nano PRV

	//board and proposal parameters
	TestnetBasicReward                = 400000000 //40 mili PRV
	TestnetRewardHalflife             = 3155760   //1 year, reduce 12.5% per year
	TestnetGenesisBlockPaymentAddress = "1Uv46Pu4pqBvxCcPw7MXhHfiAD5Rmi2xgEE7XB6eQurFAt4vSYvfyGn3uMMB1xnXDq9nRTPeiAZv5gRFCBDroRNsXJF1sxPSjNQtivuHk"
)

// for beacon
// public key
var PreSelectBeaconNodeTestnetSerializedPubkey = []string{}
var PreSelectShardNodeTestnetSerializedPubkey = []string{}

func init() {
	if len(os.Args) > 0 && (strings.Contains(os.Args[0], "test") || strings.Contains(os.Args[0], "Test")) {
		return
	}
	keyData, err := ioutil.ReadFile("keylist.json")
	if err != nil {
		panic(err)
	}

	type AccountKey struct {
		PrivateKey string
		PaymentAdd string
		PubKey     string
	}

	type KeyList struct {
		Shard  map[int][]AccountKey
		Beacon []AccountKey
	}

	keylist := KeyList{}

	err = json.Unmarshal(keyData, &keylist)
	if err != nil {
		panic(err)
	}

	for i := 0; i < TestNetMinBeaconCommitteeSize; i++ {
		PreSelectBeaconNodeTestnetSerializedPubkey = append(PreSelectBeaconNodeTestnetSerializedPubkey, keylist.Beacon[i].PubKey)
	}

	for i := 0; i < TestNetActiveShards; i++ {
		for j := 0; j < TestNetMinShardCommitteeSize; j++ {
			PreSelectShardNodeTestnetSerializedPubkey = append(PreSelectShardNodeTestnetSerializedPubkey, keylist.Shard[i][j].PubKey)
		}
	}
}

// For shard
// public key

// END CONSTANT for network TESTNET

// -------------- FOR INSTRUCTION --------------
// Action for instruction
const (
	SetAction    = "set"
	SwapAction   = "swap"
	RandomAction = "random"
	StakeAction  = "stake"
	AssignAction = "assign"
)

// ---------------------------------------------
var TestnetInitPRV = []string{
	`{
  "Version": 1,
  "Type": "s",
  "LockTime": 1563438751,
  "Fee": 0,
  "Info": null,
  "SigPubKey": "AsKCsGYkt3JthzFysVzWHxkESGfEoSRFeWafGB+DZRQA",
  "Sig": "OA3DSbUjZt28zPtTRdbHRvwI8CfZvLeVpsBggHnDMusfpkGmE3MgkmTuhqh9/rOwlEgB1ULgU3yxmdYRSUQpOA==",
  "Proof": "1111111dP9RnNmXbXtb5GKjmThj1fuurPVnBJjr5Nw15gvMRyNfy8QdqGFnPrYmeQe5NpYwgRvx7hRsgDaYGwZmM8rNGBszCM5CGyTcFsHUP95AqhTzZFugrmRU3EFt8TnfM3LktX13eD9ep7V51Ww2UcQ2PewVLz3VwktfUAvmZ3tbPWtQoQLmSFmZ4z7A47gkk7q6WjjRDLtfUbF1yj6CcswkKwMN",
  "PubKeyLastByteSender": 0,
  "Metadata": null
}`,
}
