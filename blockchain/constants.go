package blockchain

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

//Network fixed params
const (
	// BlockVersion is the current latest supported block version.
	BlockVersion            = 1
	defaultMaxBlkReqPerPeer = 600
	defaultMaxBlkReqPerTime = 1200

	defaultBroadcastStateTime = 2 * time.Second  // in second
	defaultStateUpdateTime    = 3 * time.Second  // in second
	defaultMaxBlockSyncTime   = 1 * time.Second  // in second
	defaultCacheCleanupTime   = 30 * time.Second // in second
	workerNum                 = 5
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
	MainNetStakingAmountShard  = 175000

	//board and proposal parameters
	MainnetBasicReward                = 50      //50 mili PRV
	MainnetRewardHalflife             = 6307200 //1 year, reduce 10% per year
	MainnetGenesisblockPaymentAddress = "1Uv2zzR4LgfX8ToQe8ub3bYcCLk3uDU1sm9U9hiu9EKYXoS77UdikfT9s8d5YjhsTJm61eazsMwk2otFZBYpPHwiMn8z6bKWWJRspsLky"
	// ------------- end Mainnet --------------------------------------
)

// VARIABLE for mainnet
var (
	MainnetInitConstant = []string{}
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

	TestNetShardCommitteeSize  = 3
	TestNetBeaconCommitteeSize = 3
	TestNetActiveShards        = 1
	TestNetStakingAmountShard  = 175000

	//board and proposal parameters
	TestnetBasicReward                = 50      //50 mili PRV
	TestnetRewardHalflife             = 6307200 //1 year, reduce 10% per year
	TestnetGenesisBlockPaymentAddress = "1Uv46Pu4pqBvxCcPw7MXhHfiAD5Rmi2xgEE7XB6eQurFAt4vSYvfyGn3uMMB1xnXDq9nRTPeiAZv5gRFCBDroRNsXJF1sxPSjNQtivuHk"
)

// for beacon
// public key
var PreSelectBeaconNodeTestnetSerializedPubkey = []string{}
var PreSelectShardNodeTestnetSerializedPubkey = []string{}

func init() {
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

	for i := 0; i < TestNetBeaconCommitteeSize; i++ {
		PreSelectBeaconNodeTestnetSerializedPubkey = append(PreSelectBeaconNodeTestnetSerializedPubkey, keylist.Beacon[i].PubKey)
	}

	for i := 0; i < TestNetActiveShards; i++ {
		for j := 0; j < TestNetShardCommitteeSize; j++ {
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
	DeleteAction = "del"
	SwapAction   = "swap"
	RandomAction = "random"
	StakeAction  = "stake"
)

// ---------------------------------------------
var TestnetInitConstant = []string{
	`{  
   "Version":1,
   "Type":"s",
   "LockTime":1557799670,
   "Fee":0,
   "Info":null,
   "SigPubKey":"AsKCsGYkt3JthzFysVzWHxkESGfEoSRFeWafGB+DZRQA",
   "Sig":"P1wbiDpmn2PK9G3FNILqu3JrU5E4ekfrnOz9X7Dd9HRHwp+YDFiEAMLicj7mhKcp3RCR+SWsWOaFxenbrmzXdA==",
   "Proof":"1111111RMhr5Bpy8zSZm7bQEnJcEEbeMSYh6wX9LdwSBjWhESroPN9mvBuwapr4DfKH26bQm9Eu8jtKR3saFoseZj46YQbf7iyuiA6JhAKmekHK1ds4qtFw1ipFzhLYNvp4MYXEupErvQGvZ6bvd9sxDxbwrEFJuV7i8QnHMLftsAAwDAEpDr8MkuxwDXAr5rEjoo9h6SDBHo4c1X6VRBT2GSe3",
   "PubKeyLastByteSender":0,
   "Metadata":null
	}`,
}
