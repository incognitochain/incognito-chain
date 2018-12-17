package blockchain

/*
Params defines a network by its params. These params may be used by Applications
to differentiate network as well as addresses and keys for one network
from those intended for use on another network
*/
type ParamsNew struct {
	// Name defines a human-readable identifier for the network.
	Name string

	// Net defines the magic bytes used to identify the network.
	Net uint32

	// DefaultPort defines the default peer-to-peer port for the network.
	DefaultPort string
	ShardsNum   int //max 256 shards
	// GenesisBlock defines the first block of the chain.
	GenesisBlockBeacon *BlockV2

	// GenesisBlock defines the first block of the chain.
	GenesisBlockShard *BlockV2
}

var preSelectBeaconNodeTestnetSerializedPubkey = []string{
	"1Uv23eJExNd3rumtpviranhEaKHdfccksyPucLMyaDePhJPNSuKKyfhSYDKPojPJjsvCN9fKWMRksEkGBJZiV2QyWTmM2asD5RdtRHeNA",
	"1Uv4BGuq4FmYqTMEffqhVhYpFPmo6kNBeRiCtSqZCJfduoxixftH67fxzyDgSKL45STpxzciHsNSf7eP3iDQzVZhPFzHknXtpSxgsq5bu",
	"1Uv3vqr76QGCVMRLbZ9vgDZU7mTUHLZcFv7Dy1pHxiswYXnXaABEhXrzjTWDbspeN4UQjYfUbsyQBRWvvjd8X1NytXwuM4bzP2mdWsUGg",
}

var preSelectBeaconNodeTestnet = []string{
	"124sf2tJ4K6iVD6PS4dZzs3BNYuYmHmup3Q9MfhorDrJ6aiSr46",
	"1WG3ys2tsZKpAYV7UEMirmALrMe7wDijnZfTp2Nnd9Ei6upGhc",
	"12K2poTdqzStNZjKdvYzdTBihhigTRWimHWVd7nZ5wRjEPVEZ8n",
}

var preSelectShardNodeTestnetSerializedPubkey = []string{
	"1Uv3Df2nCnm8MgzCyScbVzCKiBCJG1DbdXKN8sayDX3KwaMDEH1Vw2LmxqZzNfqxoz6jooBVj4M87GgtMfAnErNLYbjLzUpUyM4W7tNCC",
	"1Uv4CPh9RxmZK4kYzAWYDgnRJCjnZ41kqvXiETqfhYsRGKH9NidXTygWUug3nisX9NpfV96Jf2GsxT49Fa2htjQjEQezqLoE5Gg3USzRo",
	"1Uv3dihoiDhShGNn9RxayCkStfhhtrZ9tCn3mJZQmr3nQ8HhtfQn3tqctVm98knugikoxuMtqXMPf5icgR5nd34JMjGA1VAAkCAa7Pj8b",
	"1Uv4ZqdY7bSrhMH3m9ZANbxpzG5aNwkkRCrmsiL5ZGxk7T65tRajy3WWDaeQAzesosEvoMgFxYd6uUnomUih1qtbbM17HuxHYMtnyV29W",
	"1Uv3Zdz1Z5VLzkjWH4ZJDEVNq6TVWFDWVhNvsUToMvsQ8sGN8ipcW2NZ5RHGGX7RN6aBTpxd5YMePbjPPaw41Ysp37yZBkUGRiQm1RQCS",
	"1Uv2zx8j8qXwPLHUPsF5184e9wzvb37FcesmNyFBG5JLapfpfcdxuXcofLdfJzAsGFyLVAWzthRVgPd2tZv95vkz4yoHXtA3Pp4koQNMA",
	"1Uv4UB5sySevxsfdJS4YGnxULbjy9xGw8UNrZwNxcRUHtoK8eo3VLidgaxiCSFNjRGpJCURsrsZSzMqTKHcCPsTDSLuNsvq7rDK3TsSrW",
	"1Uv3zvswZeABdLmQhdVNLqmAC4YCsuyGwRDJk6WPWDjm9W4EXJvkkEfXu9nT5CTbVYt18fuLcEe1rFSaPjEvcBxTT7xBu9QTv6Jvp3swH",
	"1Uv4TtbLhvhQ7tPiB977A4aCJWzQP3DvoiyVqUi4fTKzRCGZcoAUrAEcf7h5EJmpFDTmyPKnf337r213bHyJWChg9G96K3AB498iiLpSz",
	"1Uv3wH9dEgDVPReUSzKu7XU5MQtCVapkhc7gz7fwD6TR4sypknLkm5Tq49BHRUs8D6pzFpaJwuiSCwMNXKhHMtoCkiRdTfqz6bDYMzA83",
	"1Uv2HTQ838zFvvwmJTz89wzCHL4WfLcdbqZsJcGx73GAE2zGZrJMwGh6QMPuqHJr3tyoRGCrgdkESHTppTpDHgeSBBrVwJTER2NYgtR1h",
	"1Uv3QxZ2P7vMB7A42mdYoNP25U7uAnAKUpGuZgvQUnNH5xF9ajfuAxGp3b1Ysw7rYLiVrWrj94bf4aiDDWyeQSLnWfqTWAJa7cmv4sQSm",
	"1Uv2YytznHgwQGZbHkFxpExzit89RXEMGMCmiGz9huYEQvTB3rye58zCgXYj1fxMzq9378JZhiwYnTxbfM3g9JD1Vu7dprM6jFnSHEDgq",
	"1Uv3ZNQwW4w8teiH1UEUypb3MtGdiqenu39QHXecL6r7hNj4NV2VTQL4dRj4H4KVqZoC72nrNDddysMLPHYvH8SD7d8KthNtiRZiYyV2R",
	"1Uv445DRACkpxGzDVt1rT81s4dEgCkYkZK2tguNWFjj6aVzfYWkwFgrN3SQZXEGUrR1iE9jaejmSkfcUEHu9vAB2djDEUvqSA53Se7GsR",
	"1Uv3aNz1LEitNTyQjtfKhofM3WfEZffbHSacMvA5JroZi7dDAXyUJVUsUquNz5UuwkCkDRwTrGbgEJXXiuZRctKAqhZztrTVPMYEGk5in",
	"1Uv28rhQaNhh4fJNcRKLkCbrcRLd7W21WV1pmWBMQ4FTBSLcrL8Zk3K8D92VDK86EPWdHcf37RsZh1k9oWWejNxXgrPTgk6PcGwiu8qow",
}

var preSelectShardNodeTestnet = []string{
	"12VGen58VjKC8cT3hGhSohdb8n4kz3huXka9UNcYFbUzGdgnXKZ",
	"12nVJxbZnexTmkbqcs9huztH9kN4DBCbjZewHgoyH6kHsLnf9uE",
	"12TZJQbucHA97TJNVtp8xud2BUbrzt1Mgq8Kif1BEdf51BVPFwR",
	"112hmH8nGFpJoqbevB7pmXGqyHenzxuP67tSyh4jfGqr5PbC4yNQ",
	"12ixtJSwVqvLrB4x14ux9c3h2DyUgdfvyjt5XooHkxh6vbcZomW",
	"1cizgU9GeDuEiH7GddwnV2YhPBB3aD1DMir3dynDQahjwQyqTk",
	"17EMNk6W3QpgmjxdtCaZAYmG7sBqN4XxC9bo6YfnAu587ASGv9g",
	"1Jd94JYrqLGLUV6wEa43gdsDGc6JGcy2hYbsNptRuSS3iPz24e",
	"1Q7P7QZGfJSrzC3US1Eqw2iPYDX5rqEG2T8ADsjrML5cQbSaU8",
	"12mZfvHfV5h92TTF45EQgsKU7SkLNRZXLUf6WGLf24EcKfU5Xb6",
	"1n7Zch76tzjdQVLpJxeBmPkimBTWbFmQkSsDsvGAE7GMyUYmuh",
	"17V5TXkUr12JvDrChUQ1kHaQPVFUoVCGGQji9qphTS8asVJBwdF", // me
	"1YX8vFm8zkQEyHLMRSdr8LG4TS7Ua1xq7pWp8dzsbWkDZjsoZY",
	"12ts69QMg83g2v8tutoFPxaKbbxPzpSCCQ12k6XTtDxHzr4d46S",
	"1AH2pPWpF9TjmMaaAUT26WgfSJw31EhdyssHUecxKCmCzZGMB3",
	"12obfKTP2yTtQVx3mcHk2pKBZBoZEeyjmmcfA7SgtNwCFhHKLrB",
	"12k5BfodMQLMDZXmKNwd9gj7eqek3WQqmwYxyj37HBtJpMx1djR",
}

var icoParamsTestnetNew = IcoParams{
	InitialPaymentAddress: TestnetGenesisBlockPaymentAddress,
	InitFundSalary:        TestnetInitFundSalary,
	InitialBondToken:      TestnetInitBondToken,
	InitialCMBToken:       TestnetInitCmBToken,
	InitialDCBToken:       TestnetInitDCBToken,
	InitialGOVToken:       TestnetInitGovToken,
}

var TestNetParamsNew = ParamsNew{
	Name:        TestnetName,
	Net:         Testnet,
	DefaultPort: TestnetDefaultPort,
	ShardsNum:   3,
	// blockChain parameters
	GenesisBlockBeacon: BeaconBlockGenerator{}.CreateBeaconGenesisBlock(1, preSelectBeaconNodeTestnetSerializedPubkey, preSelectShardNodeTestnetSerializedPubkey, icoParamsTestnetNew, 1000, 1000),
}

type IcoParams struct {
	InitialPaymentAddress string
	InitFundSalary        uint64
	InitialDCBToken       uint64
	InitialCMBToken       uint64
	InitialGOVToken       uint64
	InitialBondToken      uint64
	InitialVoteDCBToken   uint64
	InitialVoteGOVToken   uint64
}
