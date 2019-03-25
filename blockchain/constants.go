package blockchain

import "time"

//Network fixed params
const (
	// BlockVersion is the current latest supported block version.
	BlockVersion                = 1
	defaultMaxBlkReqPerPeer     = 60
	defaultMaxBlkReqPerTime     = 600
	defaultBroadcastStateTime   = 2 * time.Second  // in second
	defaultProcessPeerStateTime = 5 * time.Second  // in second
	defaultMaxBlockSyncTime     = 2 * time.Second  // in second
	defaultCacheCleanupTime     = 60 * time.Second // in second

	// Threshold ratio
	ThresholdRatioOfDCBCrisis = 9000
	ThresholdRatioOfGOVCrisis = 9000
)

// CONSTANT for network MAINNET
const (
	// ------------- Mainnet ---------------------------------------------
	Mainnet            = 0x01
	MainetName         = "mainnet"
	MainnetDefaultPort = "9333"

	MainNetShardCommitteeSize  = 1
	MainNetBeaconCommitteeSize = 1
	MainNetActiveShards        = 2

	//board and proposal parameters
	MainnetSalaryPerTx                = 0
	MainnetBasicSalary                = 0
	MainnetInitFundSalary             = 0
	MainnetInitDCBToken               = 0
	MainnetInitGovToken               = 0
	MainnetInitCmBToken               = 0
	MainnetInitBondToken              = 0
	MainnetFeePerTxKb                 = 0
	MainnetGenesisblockPaymentAddress = "1Uv2zzR4LgfX8ToQe8ub3bYcCLk3uDU1sm9U9hiu9EKYXoS77UdikfT9s8d5YjhsTJm61eazsMwk2otFZBYpPHwiMn8z6bKWWJRspsLky"
	// ------------- end Mainnet --------------------------------------
)

var MainnetInitConstant = []string{}

// for beacon
// public key
var PreSelectBeaconNodeMainnetSerializedPubkey = PreSelectBeaconNodeTestnetSerializedPubkey

// For shard
// public key
var PreSelectShardNodeMainnetSerializedPubkey = PreSelectShardNodeTestnetSerializedPubkey

// END CONSTANT for network MAINNET

// CONSTANT for network TESTNET
const (
	Testnet            = 0x02
	TestnetName        = "testnet"
	TestnetDefaultPort = "9444"

	TestNetShardCommitteeSize  = 1
	TestNetBeaconCommitteeSize = 1
	TestNetActiveShards        = 2

	//board and proposal parameters
	TestnetSalaryPerTx                = 10
	TestnetBasicSalary                = 10
	TestnetInitFundSalary             = 1000000
	TestnetInitDCBToken               = 10000
	TestnetInitGovToken               = 10000
	TestnetInitCmBToken               = 10000
	TestnetInitBondToken              = 10000
	TestnetFeePerTxKb                 = 1
	TestnetGenesisBlockPaymentAddress = "1Uv2zzR4LgfX8ToQe8ub3bYcCLk3uDU1sm9U9hiu9EKYXoS77UdikfT9s8d5YjhsTJm61eazsMwk2otFZBYpPHwiMn8z6bKWWJRspsLky"
)

var TestnetInitConstant = []string{`[{"Version":1,"Type":"s","LockTime":1553499428,"Fee":0,"Info":null,"SigPubKey":"AsKCsGYkt3JthzFysVzWHxkESGfEoSRFeWafGB+DZRQA","Sig":"wdz+/yfK83Z6jAx+c8PTSBFF1rBZ8vizHanUp60mprI/rvYSseInpfIDEvGYei8DV8RE558+siktzIMRbrKN/A==","Proof":"11111116WGHqpGKhPnvZ7i2w3heBopZQYdwc4cG7c4H53LZKzjBdafgMwxaXKzLhaeHyPqfuTUVMYYWKUGYwnQ4F7VaMXit6aFEzcg6wDW1HYcWTSKR5mBpJ73XqFJaVg72uts16KZ7xN7Ywi5JQUhmX4C6tTsSHnPhntdzpkReBgYEUX6nwgyGJ2kuoNbXA5i8pVvLkaB9nkhg8HviyvuHphd","PubKeyLastByteSender":0,"Metadata":null}]`}

// for beacon
// public key
var PreSelectBeaconNodeTestnetSerializedPubkey = []string{
	"15NmWBEbc8faj4QxHjBh1ugpkuBC8qaoRAp2mktKiwcKiaQgV8i",
	"16QMc6ARYki7eL3p8cj8T8b54ZAhPrnBcfaTY9CgPBDKEtwcm2u",
	"16S3Db9V2kqmmogfggKAD2bpJjXcveJcdUQmx9S3ewEGQBE3rrv",
	"158haewyeNr4WXGk4Bao2MUonNNaAjSpYeUTJ8JoD4at2AjVS45",
	"177ZHyh2WpZeVcFPJULigwkH6dahem4jDHkXQMZGNgVDijyKjKJ",
	"1671hBGTAT1ui2BQGqpzYyy3pVLPvdDTPEMLfoLix7igUyzG6sE",

	"15QDhbfx5bE7HZPkpmsayKPD13ZM1zfrCJXdbMm8y63LZU8fbG9",
	"15yb9gJC7QH3hTdjxNyV4Mdr7FJkjia5SQ5Bbwb4EHAtBTXSWC1",
	"15MC1D4orEfnJFaEeebVuUmxq2UDAwcogd8efPMDxC2cJW2MFDD",
	"15nAkpuo4BKCPk3VFey81WZaHV1YYV8HW8925eNb24mhTTtHFqV",
	"18E2LKY4PND2tAwcM9hNsZsvK1tUpKUwtLjMVRmjCBURKKD14Du",
	"17AHQTTHCnTDrt1wj3t7rThjEQndrSpk75Pv4xEpAsMw4caD4xa",
	"16XruBoQVSdMwvLzo28cL4SmdQJyBLrZD3Y9ezjAdcxaBp6GXUv",
	"16rsKSNDMGUcr7Xpt5PpkA1KiHkujEcF9BPMN9AW4gS8ZZNBu9R",
	"15QRR2WDLQcpru794E7GFjsJthhJWLhSDo2t9DToizWBAXZkCyi",
	"17k3AdAH1crag9gV4VNT7SLK76zbnc8jMUw9WPmztk7GG5n5YWt",
	"17HZ2km8EHB94EnxC9o4nP14GbzrB4Ci7yULmS7DcooY6eP4wNG",
	"17p8PgmAUpmm6CJqAtw8B5W3hLCdVYiq1W3fZ7v4hy183Pq9xYf",
	"17PKTxtJUtFLb73tu77qPqYNVvYmGMPLPCh9A1ukG3VvWqWSHQL",
	"16eykkm9y72xbrG9pLhdCWm7FVEY77gCrbpaJMgp8Ck7SZXDDkn",
	"188cBXFqnKoTroNAPtwYThbbmEoy1gAZnLCYQPL4q6gEFbz1xCs",
	"164DBoH1SDYY5xnzctuE7KEnut48tbeSTQ27zHXaTv8vvbUc2GF",
	"15rfZautPLY7mxe2MrBCmgpmRCoXfCXe4Ay1mKRKETcxabGQeDq",
	"15PMBJz2vQDPWXmDUZY33SN4uG5BKgsHk2sTxkJo156PJPy28dm",
}

// For shard
// public key
var PreSelectShardNodeTestnetSerializedPubkey = []string{
	"177KNe6pRhi97hD9LqjUvGxLoNeKh9F5oSeh99V6Td2sQcm7qEu", //shard 0
	"16W9eKEqyJqKKDkzxcSAKu4G2b1HvZh9FDRmM3ZyC4tN3MkVx6z", //shard 0
	"17zmxXqnwTK1YE42eNqVJ51mvRaCFoqzm6HogpQQBBt8dWwaUgV", //shard 0
	"17wqq26DuTQ6Hr7ocuMBdeu934rLqSoMyxib4RQdQhLUQ7Le3KD", //shard 0

	"17S44aXG7y9yEmb932MWQrrRT4Rc6pehK2UMC5np84QB2UYZdZM", //shard 1
	"18YNhMumBmeWE8GJJGbW19esqtB22zUiQx73Rwifxkyt1YKCp1s", //shard 1
	"15QYRykFuiFhoU56EAJYFRXn5UWurSuyGiZox9y7rCoSzpKW62H", //shard 1
	"15Mjx8UwK9hG2xFrmyVxco39BSDLidvcy7MBDTivuPBme6JJ24e", //shard 1

	"17E9zkHtf495WBkdo47vDB2AVTLLtSq5QtpFU2X7sQcEgHSLmfB",
	"16zsNt8d4UEtGR5c5gLfW4GvhWe3NXQv9K3tBEfom8FTYKNDeim",
	"173HS3C7RFGJDWH8YwtDvMG1s9tgrHu69DtMxVi9NsunwWnrWjk",
	"16uYeyZyRe3pzpWsyjJvqVGsa65R3A4myCzYYX9qr2Gw6L4YBnu",
	"18eckf9WNsj4hrm9goesUqeXgDzracNbr7m86qpHDLbB3jEC4wt",
	"17bgRBWmoNUCRZFAtmRKo7af98t53AWWmVagB5cwPj26Ri7ipPg",
}

// END CONSTANT for network TESTNET

// -------------- FOR INSTRUCTION --------------
// Action for instruction
const (
	SetAction    = "set"
	InitAction   = "init"
	DeleteAction = "del"
	SwapAction   = "swap"
	RandomAction = "random"
	StakeAction  = "stake"
)

// Key param for instruction
const (
	salaryPerTx = "salaryPerTx"
	basicSalary = "basicSalary"
	salaryFund  = "salaryFund"
	feePerTxKb  = "feePerTxKb"
)

// ---------------------------------------------
