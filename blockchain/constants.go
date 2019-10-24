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
	// SHARD_BLOCK_VERSION is the current latest supported block version.
	VERSION                      = 1
	RANDOM_NUMBER                = 3
	SHARD_BLOCK_VERSION          = 1
	BEACON_BLOCK_VERSION         = 1
	DefaultMaxBlkReqPerPeer      = 600
	DefaultMaxBlkReqPerTime      = 1200
	MinCommitteeSize             = 3                // min size to run bft
	DefaultBroadcastStateTime    = 6 * time.Second  // in second
	DefaultStateUpdateTime       = 8 * time.Second  // in second
	DefaultMaxBlockSyncTime      = 1 * time.Second  // in second
	DefaultCacheCleanupTime      = 30 * time.Second // in second
	WorkerNumber                 = 5
	MAX_S2B_BLOCK                = 30
	MAX_BEACON_BLOCK             = 5
	LowerBoundPercentForIncDAO   = 3
	UpperBoundPercentForIncDAO   = 10
	GetValidBlock                = 20
	TestRandom                   = true
	NumberOfFixedBlockValidators = 22
)

// CONSTANT for network MAINNET
const (
	// ------------- Mainnet ---------------------------------------------
	Mainnet                 = 0x01
	MainetName              = "mainnet"
	MainnetDefaultPort      = "9333"
	MainnetGenesisBlockTime = "2019-10-31T00:00:00.000Z"
	MainnetEpoch            = 350
	MainnetRandomTime       = 175
	MainnetOffset           = 4
	MainnetSwapOffset       = 4
	MainnetAssignOffset     = 8

	MainNetShardCommitteeSize     = 32
	MainNetMinShardCommitteeSize  = 4
	MainNetBeaconCommitteeSize    = 32
	MainNetMinBeaconCommitteeSize = 4
	MainNetActiveShards           = 8
	MainNetStakingAmountShard     = 1750000000000 // 1750 PRV = 1750 * 10^9 nano PRV

	MainnetMinBeaconBlkInterval = 40 * time.Second //second
	MainnetMaxBeaconBlkCreation = 10 * time.Second //second
	MainnetMinShardBlkInterval  = 40 * time.Second //second
	MainnetMaxShardBlkCreation  = 10 * time.Second //second

	//board and proposal parameters
	MainnetBasicReward                      = 1386666000 //1.386666 PRV
	MainnetRewardHalflife                   = 3155760    //1 year, reduce 12.5% per year
	MainETHContractAddressStr               = "0x10e492e6383DfE37d0d0B7B86015AE0876e88663"
	MainnetIncognitoDAOAddress              = "" // community fund
	MainnetCentralizedWebsitePaymentAddress = "12Rvjw6J3FWY3YZ1eDZ5uTy6DTPjFeLhCK7SXgppjivg9ShX2RRq3s8pdoapnH8AMoqvUSqZm1Gqzw7rrKsNzRJwSK2kWbWf1ogy885"
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
	Testnet                 = 0x16
	TestnetName             = "testnet"
	TestnetDefaultPort      = "9444"
	TestnetGenesisBlockTime = "2019-10-21T00:00:20.000Z"
	TestnetEpoch            = 100
	TestnetRandomTime       = 50
	TestnetOffset           = 4
	TestnetSwapOffset       = 4
	TestnetAssignOffset     = 8

	TestNetShardCommitteeSize     = 32
	TestNetMinShardCommitteeSize  = 23
	TestNetBeaconCommitteeSize    = 4
	TestNetMinBeaconCommitteeSize = 4
	TestNetActiveShards           = 8
	TestNetStakingAmountShard     = 1750000000000 // 1750 PRV = 1750 * 10^9 nano PRV

	TestNetMinBeaconBlkInterval = 10 * time.Second //second
	TestNetMaxBeaconBlkCreation = 8 * time.Second  //second
	TestNetMinShardBlkInterval  = 10 * time.Second //second
	TestNetMaxShardBlkCreation  = 6 * time.Second  //second

	//board and proposal parameters
	TestnetBasicReward                      = 400000000                                    //40 mili PRV
	TestnetRewardHalflife                   = 3155760                                      //1 year, reduce 12.5% per year
	TestnetETHContractAddressStr            = "0x904836fb12c4A8eafCfFe805F1C561cC2940932a" // v35 - kovan, devnet, for branch dev/issue339
	TestnetIncognitoDAOAddress              = "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci"
	TestnetCentralizedWebsitePaymentAddress = "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci"
)

// for beacon
// public key
var PreSelectBeaconNodeTestnetSerializedPubkey = []string{}
var PreSelectBeaconNodeTestnetSerializedPaymentAddress = []string{}
var PreSelectShardNodeTestnetSerializedPubkey = []string{}
var PreSelectShardNodeTestnetSerializedPaymentAddress = []string{}

func init() {
	if len(os.Args) > 0 && (strings.Contains(os.Args[0], "test") || strings.Contains(os.Args[0], "Test")) {
		return
	}
	var keyData []byte
	var err error
	var IsTestNet = true
	if IsTestNet {
		keyData, err = ioutil.ReadFile("keylist.json")
		if err != nil {
			panic(err)
		}
	} else {
		keyData, err = ioutil.ReadFile("keylist-mainnet.json")
		if err != nil {
			panic(err)
		}
	}

	type AccountKey struct {
		PrivateKey     string
		PaymentAddress string
		// PubKey     string
		CommitteePublicKey string
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

	if IsTestNet {
		for i := 0; i < TestNetMinBeaconCommitteeSize; i++ {
			PreSelectBeaconNodeTestnetSerializedPubkey = append(PreSelectBeaconNodeTestnetSerializedPubkey, keylist.Beacon[i].CommitteePublicKey)
			PreSelectBeaconNodeTestnetSerializedPaymentAddress = append(PreSelectBeaconNodeTestnetSerializedPaymentAddress, keylist.Beacon[i].PaymentAddress)
		}

		for i := 0; i < TestNetActiveShards; i++ {
			for j := 0; j < TestNetMinShardCommitteeSize; j++ {
				PreSelectShardNodeTestnetSerializedPubkey = append(PreSelectShardNodeTestnetSerializedPubkey, keylist.Shard[i][j].CommitteePublicKey)
				PreSelectShardNodeTestnetSerializedPaymentAddress = append(PreSelectShardNodeTestnetSerializedPaymentAddress, keylist.Shard[i][j].PaymentAddress)
			}
		}
	} else {
		for i := 0; i < MainNetMinBeaconCommitteeSize; i++ {
			PreSelectBeaconNodeTestnetSerializedPubkey = append(PreSelectBeaconNodeTestnetSerializedPubkey, keylist.Beacon[i].CommitteePublicKey)
			PreSelectBeaconNodeTestnetSerializedPaymentAddress = append(PreSelectBeaconNodeTestnetSerializedPaymentAddress, keylist.Beacon[i].PaymentAddress)
		}

		for i := 0; i < MainNetActiveShards; i++ {
			for j := 0; j < MainNetMinShardCommitteeSize; j++ {
				PreSelectShardNodeTestnetSerializedPubkey = append(PreSelectShardNodeTestnetSerializedPubkey, keylist.Shard[i][j].CommitteePublicKey)
				PreSelectShardNodeTestnetSerializedPaymentAddress = append(PreSelectShardNodeTestnetSerializedPaymentAddress, keylist.Shard[i][j].PaymentAddress)
			}
		}
	}
}

// For shard
// public key

// END CONSTANT for network TESTNET

// -------------- FOR INSTRUCTION --------------
// Action for instruction
const (
	SetAction     = "set"
	SwapAction    = "swap"
	RandomAction  = "random"
	StakeAction   = "stake"
	AssignAction  = "assign"
	StopAutoStake = "stopautostake"
)

// ---------------------------------------------
var TestnetInitPRV = []string{
	`{
		"Version":1,
		"Type":"s",
		"LockTime":1570159128,
		"Fee":0,
		"Info":null,
		"SigPubKey":"5xVSzcZpA3uHmBO5ejENk13iayexILopySACdieLugA=",
		"Sig":"oMJPBLxKgTnfQhMgfvvH68ed0UTuTfl3ofOoWgk8dgvfhovgvued9HH4dXz60rY32H4Y4c85Zd8bSXSnvNhZAA==",
		"Proof":"AAAAAAAAAbAAriDnFVLNxmkDe4eYE7l6MQ2TXeJrJ7EguinJIAJ2J4u6ACARCc1/qyLEePe1zSthzmRSqf2VNOlo036JwtDgbNg24yAb6hGuk1tRBVMO4ruHaNEasY09ZiBc4iuK/dpDSyNTCCABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACDMoX8yNBbY68SO44umD1CMfz/r0T4YiXhDgDgT6+k4BgdY0V4XYoAAAAAAAAAAAAA=",
		"PubKeyLastByteSender":0,
		"Metadata":null
	}`,
}
var IntegrationTestInitPRV = []string{
	`{"Version":1,"Type":"s","LockTime":1571731249,"Fee":0,"Info":null,"SigPubKey":"Lhv3KVghjL8k/2/ytRvrKgCNIEbAE5Pk/cDs8tvJ6gA=","Sig":"FNyMr0FCI/dSjZQRBgWKm0YmVj56oBNmkm153TEiTQbU4JWWcwYtniXXUfaT1r3POIfYQdhZXq/qW5+GVu5uDA==","Proof":"AAAAAAAAAbAAriAuG/cpWCGMvyT/b/K1G+sqAI0gRsATk+T9wOzy28nqACDp/9qCrAy6kJ0DnjtDNBs0dAAqO4vzPnfF1YGWpgM1/CCsxrZUKrLqlCH3RWruqDgKQfCrH8BF7JbMIW4EOyc8DyABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACBEz/l9o3i63Hok6N2sDfKWQZviNfBvDBHdjIg3PKscAwcDjX6kxoAAAAAAAAAAAAA=","PubKeyLastByteSender":0,"Metadata":null}`,
	`{"Version":1,"Type":"s","LockTime":1571731249,"Fee":0,"Info":null,"SigPubKey":"0vbTakF/PEd8VFP+yDc/qI5Az2yF53EvUB7ybNPFkwA=","Sig":"UR7iP8LDBFfeFRFevq/mGsiCve9Xxk2aVa9vPqO5DQLcObqcVIA6zP0/ctMST/RYxnEPKPl4NPr1nOr+dBJSAg==","Proof":"AAAAAAAAAbAAriDS9tNqQX88R3xUU/7INz+ojkDPbIXncS9QHvJs08WTACDhMuKM5EQsPDgCd7IricPaQS6N9uMM3gN4PZRe1yXwYCCBAEWU8c2Gt0kJe7t4UZBQEcEZXxuQ9ndVQKWbNqLeCSABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACDWy9/ILNqq+MpLFPg0v1trqCR64z6Gw2WrX1laMaDhAQcDjX6kxoAAAAAAAAAAAAA=","PubKeyLastByteSender":0,"Metadata":null}`,
	`{"Version":1,"Type":"s","LockTime":1571731249,"Fee":0,"Info":null,"SigPubKey":"JYZc0+zwvihljeRdfyuC6hOcN/irbiLn8dIes2BxlAA=","Sig":"4XcBCMtdO3kiOpJOfEgpvArVvfISEZzOFDnS4ptkkg7CEYVvZobJscGpSSLJUG8OqETWTUBy10pKeiP0vatuBw==","Proof":"AAAAAAAAAbAAriAlhlzT7PC+KGWN5F1/K4LqE5w3+KtuIufx0h6zYHGUACBWs5rUOtje/lry+HC7cFlFvrLBkUrDEgcitZdXda5ARiDrevG4H9W3wLOyuU3SAAtlnjoYvWZouUCct2LrsJkODCABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACDfhCqzv4vaWgUTWF8Tla1U7hHrmTM/oj0KeWKT4GWNAAcDjX6kxoAAAAAAAAAAAAA=","PubKeyLastByteSender":0,"Metadata":null}`,
	`{"Version":1,"Type":"s","LockTime":1571731249,"Fee":0,"Info":null,"SigPubKey":"NO5G5g4ppdudWnsU1cmEIsHErlV53B3gpHXeEpAogAE=","Sig":"jXkLYClmHg8PJGiEsrwc3eHLrx3ltjVOeRJtZKutEQWiF1omE7hbfvVFrLF066Ezbh26cgwylhuyT3oHXF1BAA==","Proof":"AAAAAAAAAbAAriA07kbmDiml251aexTVyYQiwcSuVXncHeCkdd4SkCiAASDM5SDr3N2e+9uxCv4xXGSUhm4BnS+c10zIcpSdEcEhuSCQ/fZMC5DFydGQ2YSa+t6ati94w2NbY/a98/r7YD+fCSABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACBcP2W++8InwVZ0d2LuazW/3LHQf289EmJNTyerC0GaDAcDjX6kxoAAAAAAAAAAAAA=","PubKeyLastByteSender":1,"Metadata":null}`,
	`{"Version":1,"Type":"s","LockTime":1571731249,"Fee":0,"Info":null,"SigPubKey":"iCyiBFOex1ZENgUORIgI7n1dnHU0SG42M5FLUJbqNgE=","Sig":"G8dgLjoWHAX8EsDu5hXlFa27Al1YpSxpBV+4luR6GQB+d1XC1iX18b6cHav1RtUGyieJaUAzVLVzH9f5fO6tBQ==","Proof":"AAAAAAAAAbAAriCILKIEU57HVkQ2BQ5EiAjufV2cdTRIbjYzkUtQluo2ASA+BMHwZ0sseyitECq/9uYTqsi2TCbG42KUN832niu6wCAzddFyWlXTPzQT3zwiuPg6SCZu51ZXuVXlqQNGhywbAiABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAr265F7yxwbvsC9kz5yMDIQPxWwrX1rXyAyPxJ5+eFBgcDjX6kxoAAAAAAAAAAAAA=","PubKeyLastByteSender":1,"Metadata":null}`,
	`{"Version":1,"Type":"s","LockTime":1571731249,"Fee":0,"Info":null,"SigPubKey":"2SRglEGb200vgeuGC8sT1VhYdEn5aWNgoYvB19NBWwE=","Sig":"FItaAX0rg57zf542eln8OAtIn0YR08EZQCJabFipqwcQerMkJ3Ay2Py4TMV7yItqBHect+TJcpxFcrhIaXAgBg==","Proof":"AAAAAAAAAbAAriDZJGCUQZvbTS+B64YLyxPVWFh0SflpY2Chi8HX00FbASBKG0ZrdSXBGqWnuLBrF+/sk1qvtx+4K+r/s4Z61//C3iDfoUbeXyDyIJ34GqLfGSQYqnQZ6cODxId3UgKjd9SECSABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACBnqg/K6lHPnX8k3+LZ6CXayYUgD54kon4rHQlDP6s3BQcDjX6kxoAAAAAAAAAAAAA=","PubKeyLastByteSender":1,"Metadata":null}`,
}
