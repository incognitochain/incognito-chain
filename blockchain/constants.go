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
	BlockVersion            = 1
	DefaultMaxBlkReqPerPeer = 600
	DefaultMaxBlkReqPerTime = 1200

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
	MainNetStakingAmountShard  = 175000

	//board and proposal parameters
	MainnetBasicReward                = 50      //50 mili PRV
	MainnetRewardHalflife             = 6307200 //1 year, reduce 10% per year
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
//var PreSelectBeaconNodeTestnetSerializedPubkey = []string{}
var PreSelectBeaconNodeTestnetSerializedPubkey = []string{
	"17YiepCpN6tMwD91MGSXbEBjLBjeUysVcQJm87kCSHfXA2bWM46",
	"16jWkWY5xRqZDkwDQaA32uvnakS27YyyWqeWtap6a6ELEc2Vbwi",
	"15gXn3sEVp41Kdquy696ftT5bi3fRRbExFT7p5ZftEQusaJGi3y",
}

//var PreSelectShardNodeTestnetSerializedPubkey = []string{}
var PreSelectShardNodeTestnetSerializedPubkey = []string{
	// Committee of shard 0
	"183GBqPhSfcEFZP7MQFTnuLVuX2PRkd5HFA3qkqkLN4STghvxpw", //shard 0
	"15ezEJs61P8qq6F8Zrhbcd2RpuqrtDWtzPheJWiEM6ct1sWjFTi", //shard 0
	"15s3gAPBWfWshH4uZLwL1W4wWYamFqRs2GUMKY8Qmdhet3a6s7Z", //shard 0

	// Committee of shard 1
	"18BxVp5W9LwXMzeKxauAA6BC1HvUCsC6XAij2Ng1K99ghb9G6hX", //shard 1
	"16H5t5ezMF16S5j5ZEyHP3N4nBBcsppg5bRfU5Ft8N1VZYBQu38", //shard 1
	"156qsnqcYWPUb8PLbdowV4TtUhS8kuEboABfHgVeh4MguoPwqVj", //shard 1
}

func init() {
	return

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
var TestnetInitPRV = []string{
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
