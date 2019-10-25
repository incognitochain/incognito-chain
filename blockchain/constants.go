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
	VERSION                    = 1
	RANDOM_NUMBER              = 3
	SHARD_BLOCK_VERSION        = 1
	BEACON_BLOCK_VERSION       = 1
	DefaultMaxBlkReqPerPeer    = 600
	DefaultMaxBlkReqPerTime    = 1200
	MinCommitteeSize           = 3                // min size to run bft
	DefaultBroadcastStateTime  = 6 * time.Second  // in second
	DefaultStateUpdateTime     = 8 * time.Second  // in second
	DefaultMaxBlockSyncTime    = 1 * time.Second  // in second
	DefaultCacheCleanupTime    = 30 * time.Second // in second
	WorkerNumber               = 5
	MAX_S2B_BLOCK              = 30
	MAX_BEACON_BLOCK           = 5
	LowerBoundPercentForIncDAO = 3
	UpperBoundPercentForIncDAO = 10
	GetValidBlock              = 20
	TestRandom                 = true
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
	MainNetMinShardCommitteeSize  = 22
	MainNetBeaconCommitteeSize    = 32
	MainNetMinBeaconCommitteeSize = 7
	MainNetActiveShards           = 8
	MainNetStakingAmountShard     = 1750000000000 // 1750 PRV = 1750 * 10^9 nano PRV

	MainnetMinBeaconBlkInterval = 40 * time.Second //second
	MainnetMaxBeaconBlkCreation = 10 * time.Second //second
	MainnetMinShardBlkInterval  = 40 * time.Second //second
	MainnetMaxShardBlkCreation  = 10 * time.Second //second

	//board and proposal parameters
	MainnetBasicReward                      = 1386666000 //1.386666 PRV
	MainnetRewardHalflife                   = 3155760    //1 year, reduce 12.5% per year
	MainETHContractAddressStr               = ""
	MainnetIncognitoDAOAddress              = "12S32fSyF4h8VxFHt4HfHvU1m9KHvBQsab5zp4TpQctmMdWuveXFH9KYWNemo7DRKvaBEvMgqm4XAuq1a1R4cNk2kfUfvXR3DdxCho3" // community fund
	MainnetCentralizedWebsitePaymentAddress = "12Rvjw6J3FWY3YZ1eDZ5uTy6DTPjFeLhCK7SXgppjivg9ShX2RRq3s8pdoapnH8AMoqvUSqZm1Gqzw7rrKsNzRJwSK2kWbWf1ogy885"
	// ------------- end Mainnet --------------------------------------
)

// VARIABLE for mainnet
var PreSelectBeaconNodeMainnetSerializedPubkey = []string{}
var PreSelectBeaconNodeMainnetSerializedPaymentAddress = []string{}
var PreSelectShardNodeMainnetSerializedPubkey = []string{}
var PreSelectShardNodeMainnetSerializedPaymentAddress = []string{}
var MainnetInitPRV = []string{`
	{
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
	}
	`}

// END CONSTANT for network MAINNET

// CONSTANT for network TESTNET
const (
	Testnet                 = 0x16
	TestnetName             = "testnet"
	TestnetDefaultPort      = "9444"
	TestnetGenesisBlockTime = "2019-10-21T00:00:20.000Z"
	TestnetEpoch            = 100
	TestnetRandomTime       = 50
	TestnetOffset           = 1
	TestnetSwapOffset       = 1
	TestnetAssignOffset     = 2

	TestNetShardCommitteeSize     = 16
	TestNetMinShardCommitteeSize  = 4
	TestNetBeaconCommitteeSize    = 4
	TestNetMinBeaconCommitteeSize = 4
	TestNetActiveShards           = 8
	TestNetStakingAmountShard     = 1750000000000 // 1750 PRV = 1750 * 10^9 nano PRV

	TestNetMinBeaconBlkInterval = 10 * time.Second //second
	TestNetMaxBeaconBlkCreation = 8 * time.Second  //second, timeout is 25
	TestNetMinShardBlkInterval  = 10 * time.Second //second
	TestNetMaxShardBlkCreation  = 6 * time.Second  //second, timeout is 25

	//board and proposal parameters
	TestnetBasicReward                      = 400000000 //40 mili PRV
	TestnetRewardHalflife                   = 3155760   //1 year, reduce 12.5% per year
	TestnetETHContractAddressStr            = "0x717B5F3667A21a0b5e09A8d0E8648C1D525503C4"
	TestnetIncognitoDAOAddress              = "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci" // community fund
	TestnetCentralizedWebsitePaymentAddress = "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci"
)

// VARIABLE for testnet
var PreSelectBeaconNodeTestnetSerializedPubkey = []string{}
var PreSelectBeaconNodeTestnetSerializedPaymentAddress = []string{}
var PreSelectShardNodeTestnetSerializedPubkey = []string{}
var PreSelectShardNodeTestnetSerializedPaymentAddress = []string{}
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

func init() {
	if len(os.Args) > 0 && (strings.Contains(os.Args[0], "test") || strings.Contains(os.Args[0], "Test")) {
		return
	}
	var keyData []byte
	var err error

	keyData, err = ioutil.ReadFile("keylist.json")
	if err != nil {
		panic(err)
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

	var IsTestNet = true
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
			PreSelectBeaconNodeMainnetSerializedPubkey = append(PreSelectBeaconNodeMainnetSerializedPubkey, keylist.Beacon[i].CommitteePublicKey)
			PreSelectBeaconNodeMainnetSerializedPaymentAddress = append(PreSelectBeaconNodeMainnetSerializedPaymentAddress, keylist.Beacon[i].PaymentAddress)
		}

		for i := 0; i < MainNetActiveShards; i++ {
			for j := 0; j < MainNetMinShardCommitteeSize; j++ {
				PreSelectShardNodeMainnetSerializedPubkey = append(PreSelectShardNodeMainnetSerializedPubkey, keylist.Shard[i][j].CommitteePublicKey)
				PreSelectShardNodeMainnetSerializedPaymentAddress = append(PreSelectShardNodeMainnetSerializedPaymentAddress, keylist.Shard[i][j].PaymentAddress)
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

var IntegrationTestInitPRV = []string{`{"Version":1,"Type":"s","LockTime":1564213226,"Fee":0,"Info":null,"SigPubKey":"A6zmFqIlTKgsV23Qk9jz2roo3VhisVy5Flg6EGuOKaQA","Sig":"f+JDTKpO7+veF6DVYobNp6l0l6rAYxCZjYCNRrsFN0lx7aOMOwXhZK0OGrKiDLfqSIMX7CXr9ProBz7TIx3yqg==","Proof":"1111111dP9RnNnGCD9afUsg4bvrBHNWfjZijttFU2bkFYLYFGqCoK6i6RCeSEk2NUmv7p8B4kyhi1qaoMjvYCotjhDogGiuYrEqUT4NQLXatq2xqkfxgX8DURcv9xCgrgqVceQ2DrBR5NcgbMQHHBnW1xV3Dte2kmq837EeufP3KoQpz3m5N3oN6x1UssfWSeHAuw4t2dUinKDTe7SgRnFFhfF59dvy","PubKeyLastByteSender":0,"Metadata":null}`,
	`{"Version":1,"Type":"s","LockTime":1564213226,"Fee":0,"Info":null,"SigPubKey":"As3StzeOJhR5qheXo9stChC6WqQJChZNqmPqdgNOFtkA","Sig":"ccWpvPZjitORv6+9WOWv7K5e8purHA4sX7mfBNE9m9YYFyPJ2awx5+1iHuWKD7BH9oum64XCiLYtW9iihVGlDw==","Proof":"1111111dP9RnNmZen93jhEW3eXaKkne72tbWVGtcdfAEfnbdf7fPDQmwYaTve2a9MBA56HHWXzXCbDxx79KCrtrArUqQKnxgun69qQpCjDZhaBdpKNZAAvYf7uBHrnxpm7qxRA4XLGSKbuLS6mBtrCUFPnit9BDbSAu9ZxQsPnr7XPPyHdbBofrBzFLqf2zTPMrqCAZqBqapA5AMtd8J8yknUHX6hWJ","PubKeyLastByteSender":0,"Metadata":null}`,
	`{"Version":1,"Type":"s","LockTime":1564502136,"Fee":0,"Info":null,"SigPubKey":"AmusT4yw6LoipXRBH10JL7D1I9B2jwN5gVsQA6SexgoB","Sig":"1aZeIjgrFhe9P16J9vd0V4pCOemknsJ/Ljy/a0fhqimyZL+YUpo+Q+rD0T2Tan9e8StbXQi944VD4EItqYhuWw==","Proof":"1111111dP9RnNmFpBcsd8WSQtTxPB9QfuMN8YS39CkSCi7zvR9pRxSNgVgXADCBjkCdMDH9K9VC3SQ1DstvsTSGuJ1XkjHghTWtMbGEeedBai4f4DjByeLzStJagXtuwQAxoia7Gowg7rutuJVLThVEHFDNVjdgmy8h7NCYZrL4YQiy3QLqeLqKwzoBULxEW2rw62HM2xsFjCsk7twTJCpHW1kc9ThT","PubKeyLastByteSender":1,"Metadata":null}`}
