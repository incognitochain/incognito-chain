package config

import (
	"time"
)

var MainnetParam = &param{
	Name: "mainnet",
	Net:  0x01,
	GenesisParam: &genesisParam{
		SelectBeaconNodeSerializedPubkeyV2:         map[uint64][]string{},
		SelectBeaconNodeSerializedPaymentAddressV2: map[uint64][]string{},
		SelectShardNodeSerializedPubkeyV2:          map[uint64][]string{},
		SelectShardNodeSerializedPaymentAddressV2:  map[uint64][]string{},
		FeePerTxKb:         0,
		ConsensusAlgorithm: "bls",
		BlockTimestamp:     "2019-10-29T00:00:00.000Z",
		TxStake:            "d0e731f55fa6c49f602807a6686a7ac769de4e04882bb5eaf8f4fe209f46535d",
	},
	CommitteeSize: committeeSize{
		MaxShardCommitteeSize:            32,
		MinShardCommitteeSize:            22,
		MaxBeaconCommitteeSize:           32,
		MinBeaconCommitteeSize:           7,
		InitShardCommitteeSize:           22,
		InitBeaconCommitteeSize:          7,
		ShardCommitteeSizeKeyListV2:      22,
		BeaconCommitteeSizeKeyListV2:     7,
		NumberOfFixedShardBlockValidator: 22,
	},
	BlockTime: blockTime{
		MinShardBlockInterval:  40 * time.Second,
		MaxShardBlockCreation:  10 * time.Second,
		MinBeaconBlockInterval: 40 * time.Second,
		MaxBeaconBlockCreation: 10 * time.Second,
	},
	StakingAmountShard: 1750000000000,
	ActiveShards:       8,
	BasicReward:        1386666000,
	EpochParam: epochParam{
		NumberOfBlockInEpoch:   350,
		NumberOfBlockInEpochV2: 1e9,
		EpochV2BreakPoint:      1e9,
		RandomTime:             175,
		RandomTimeV2:           1e9,
	},
	EthContractAddressStr:            "0x43D037A562099A4C2c95b1E2120cc43054450629",
	BscContractAddressStr:            "",
	IncognitoDAOAddress:              "12S32fSyF4h8VxFHt4HfHvU1m9KHvBQsab5zp4TpQctmMdWuveXFH9KYWNemo7DRKvaBEvMgqm4XAuq1a1R4cNk2kfUfvXR3DdxCho3",
	CentralizedWebsitePaymentAddress: "12Rvjw6J3FWY3YZ1eDZ5uTy6DTPjFeLhCK7SXgppjivg9ShX2RRq3s8pdoapnH8AMoqvUSqZm1Gqzw7rrKsNzRJwSK2kWbWf1ogy885",
	SwapCommitteeParam: swapCommitteeParam{
		Offset:       4,
		SwapOffset:   4,
		AssignOffset: 8,
	},
	ConsensusParam: consensusParam{
		ConsensusV2Epoch:          3071,
		StakingFlowV2Height:       1207793,
		EnableSlashingHeight:      1000000000000,
		Timeslot:                  40,
		EpochBreakPointSwapNewKey: []uint64{1917},
	},
	BeaconHeightBreakPointBurnAddr: 150500,
	ReplaceStakingTxHeight:         559380,
	ETHRemoveBridgeSigEpoch:        1973,
	BCHeightBreakPointNewZKP:       934858,
	EnableFeatureFlags:             map[int]uint64{0: 1, 1: 0},
	BCHeightBreakPointPortalV3:     10000000,
	TxPoolVersion:                  0,
	GethParam: gethParam{
		Host:     "https://eth-fullnode.incognito.org",
		Protocol: "",
		Port:     "",
	},
	BSCParam: bscParam{
		Host: "https://bsc-dataseed.binance.org",
	},
	IsBackup: false,
}

var Testnet1Param = &param{
	Name: "testnet-1",
	Net:  0x16,
	GenesisParam: &genesisParam{
		SelectBeaconNodeSerializedPubkeyV2:         map[uint64][]string{},
		SelectBeaconNodeSerializedPaymentAddressV2: map[uint64][]string{},
		SelectShardNodeSerializedPubkeyV2:          map[uint64][]string{},
		SelectShardNodeSerializedPaymentAddressV2:  map[uint64][]string{},
		FeePerTxKb:         0,
		ConsensusAlgorithm: "bls",
		BlockTimestamp:     "2020-08-11T00:00:00.000Z",
		TxStake:            "d0e731f55fa6c49f602807a6686a7ac769de4e04882bb5eaf8f4fe209f46535d",
	},
	CommitteeSize: committeeSize{
		MaxShardCommitteeSize:            32,
		MinShardCommitteeSize:            4,
		MaxBeaconCommitteeSize:           4,
		MinBeaconCommitteeSize:           4,
		InitShardCommitteeSize:           4,
		InitBeaconCommitteeSize:          4,
		ShardCommitteeSizeKeyListV2:      4,
		BeaconCommitteeSizeKeyListV2:     4,
		NumberOfFixedShardBlockValidator: 4,
	},
	BlockTime: blockTime{
		MinShardBlockInterval:  10 * time.Second,
		MaxShardBlockCreation:  8 * time.Second,
		MinBeaconBlockInterval: 10 * time.Second,
		MaxBeaconBlockCreation: 6 * time.Second,
	},
	StakingAmountShard: 1750000000000,
	ActiveShards:       8,
	BasicReward:        400000000,
	EpochParam: epochParam{
		NumberOfBlockInEpoch:   100,
		NumberOfBlockInEpochV2: 1e9,
		EpochV2BreakPoint:      1e9,
		RandomTime:             50,
		RandomTimeV2:           1e9,
	},
	EthContractAddressStr:            "0xE0D5e7217c6C4bc475404b26d763fAD3F14D2b86",
	BscContractAddressStr:            "0x1ce57B254DC2DBB41e1aeA296Dc7dBD6fb549241",
	IncognitoDAOAddress:              "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
	CentralizedWebsitePaymentAddress: "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
	SwapCommitteeParam: swapCommitteeParam{
		Offset:       1,
		SwapOffset:   1,
		AssignOffset: 2,
	},
	ConsensusParam: consensusParam{
		ConsensusV2Epoch:          15290,
		StakingFlowV2Height:       2051863,
		EnableSlashingHeight:      2087789,
		Timeslot:                  10,
		EpochBreakPointSwapNewKey: []uint64{1280},
	},
	BeaconHeightBreakPointBurnAddr: 1,
	ReplaceStakingTxHeight:         1,
	ETHRemoveBridgeSigEpoch:        2085,
	BCHeightBreakPointNewZKP:       1148608,
	EnableFeatureFlags:             map[int]uint64{0: 1, 1: 0},
	BCHeightBreakPointPortalV3:     1328816,
	TxPoolVersion:                  0,
	GethParam: gethParam{
		Host:     "kovan.infura.io/v3/1138a1e99b154b10bae5c382ad894361",
		Protocol: "https",
		Port:     "",
	},
	BSCParam: bscParam{
		Host: "https://data-seed-prebsc-2-s1.binance.org:8545",
	},
	IsBackup: false,
}

var Tesnet2Param = &param{
	Name: "testnet-2",
	Net:  0x32,
	GenesisParam: &genesisParam{
		SelectBeaconNodeSerializedPubkeyV2:         map[uint64][]string{},
		SelectBeaconNodeSerializedPaymentAddressV2: map[uint64][]string{},
		SelectShardNodeSerializedPubkeyV2:          map[uint64][]string{},
		SelectShardNodeSerializedPaymentAddressV2:  map[uint64][]string{},
		FeePerTxKb:         0,
		ConsensusAlgorithm: "bls",
		BlockTimestamp:     "2020-08-11T00:00:00.000Z",
		TxStake:            "d0e731f55fa6c49f602807a6686a7ac769de4e04882bb5eaf8f4fe209f46535d",
	},
	CommitteeSize: committeeSize{
		MaxShardCommitteeSize:            32,
		MinShardCommitteeSize:            4,
		MaxBeaconCommitteeSize:           4,
		MinBeaconCommitteeSize:           4,
		InitShardCommitteeSize:           4,
		InitBeaconCommitteeSize:          4,
		ShardCommitteeSizeKeyListV2:      4,
		BeaconCommitteeSizeKeyListV2:     4,
		NumberOfFixedShardBlockValidator: 4,
	},
	BlockTime: blockTime{
		MinShardBlockInterval:  10 * time.Second,
		MaxShardBlockCreation:  8 * time.Second,
		MinBeaconBlockInterval: 10 * time.Second,
		MaxBeaconBlockCreation: 6 * time.Second,
	},
	StakingAmountShard: 1750000000000,
	ActiveShards:       8,
	BasicReward:        400000000,
	EpochParam: epochParam{
		NumberOfBlockInEpoch:   100,
		NumberOfBlockInEpochV2: 1e9,
		EpochV2BreakPoint:      1e9,
		RandomTime:             50,
		RandomTimeV2:           1e9,
	},
	EthContractAddressStr:            "0x2f6F03F1b43Eab22f7952bd617A24AB46E970dF7",
	BscContractAddressStr:            "0x2f6F03F1b43Eab22f7952bd617A24AB46E970dF7",
	IncognitoDAOAddress:              "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
	CentralizedWebsitePaymentAddress: "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
	SwapCommitteeParam: swapCommitteeParam{
		Offset:       1,
		SwapOffset:   1,
		AssignOffset: 2,
	},
	ConsensusParam: consensusParam{
		ConsensusV2Epoch:          15290,
		StakingFlowV2Height:       2051863,
		EnableSlashingHeight:      2087789,
		Timeslot:                  10,
		EpochBreakPointSwapNewKey: []uint64{1280},
	},
	BeaconHeightBreakPointBurnAddr: 1,
	ReplaceStakingTxHeight:         1,
	ETHRemoveBridgeSigEpoch:        2085,
	BCHeightBreakPointNewZKP:       1148608,
	EnableFeatureFlags:             map[int]uint64{0: 1, 1: 0},
	BCHeightBreakPointPortalV3:     1328816,
	TxPoolVersion:                  0,
	GethParam: gethParam{
		Host:     "kovan.infura.io/v3/1138a1e99b154b10bae5c382ad894361",
		Protocol: "https",
		Port:     "",
	},
	BSCParam: bscParam{
		Host: "https://data-seed-prebsc-2-s1.binance.org:8545",
	},
	IsBackup: false,
}

var LocalParam = &param{
	Name: "local",
	Net:  0x32,
	GenesisParam: &genesisParam{
		SelectBeaconNodeSerializedPubkeyV2:         map[uint64][]string{},
		SelectBeaconNodeSerializedPaymentAddressV2: map[uint64][]string{},
		SelectShardNodeSerializedPubkeyV2:          map[uint64][]string{},
		SelectShardNodeSerializedPaymentAddressV2:  map[uint64][]string{},
		FeePerTxKb:         0,
		ConsensusAlgorithm: "bls",
		BlockTimestamp:     "2020-08-11T00:00:00.000Z",
		TxStake:            "d0e731f55fa6c49f602807a6686a7ac769de4e04882bb5eaf8f4fe209f46535d",
	},
	CommitteeSize: committeeSize{
		MaxShardCommitteeSize:            32,
		MinShardCommitteeSize:            4,
		MaxBeaconCommitteeSize:           4,
		MinBeaconCommitteeSize:           4,
		InitShardCommitteeSize:           4,
		InitBeaconCommitteeSize:          4,
		ShardCommitteeSizeKeyListV2:      4,
		BeaconCommitteeSizeKeyListV2:     4,
		NumberOfFixedShardBlockValidator: 4,
	},
	BlockTime: blockTime{
		MinShardBlockInterval:  10 * time.Second,
		MaxShardBlockCreation:  8 * time.Second,
		MinBeaconBlockInterval: 10 * time.Second,
		MaxBeaconBlockCreation: 6 * time.Second,
	},
	StakingAmountShard: 1750000000000,
	ActiveShards:       8,
	BasicReward:        400000000,
	EpochParam: epochParam{
		NumberOfBlockInEpoch:   100,
		NumberOfBlockInEpochV2: 1e9,
		EpochV2BreakPoint:      1e9,
		RandomTime:             50,
		RandomTimeV2:           1e9,
	},
	EthContractAddressStr:            "0x2f6F03F1b43Eab22f7952bd617A24AB46E970dF7",
	BscContractAddressStr:            "0x2f6F03F1b43Eab22f7952bd617A24AB46E970dF7",
	IncognitoDAOAddress:              "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
	CentralizedWebsitePaymentAddress: "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
	SwapCommitteeParam: swapCommitteeParam{
		Offset:       1,
		SwapOffset:   1,
		AssignOffset: 2,
	},
	ConsensusParam: consensusParam{
		ConsensusV2Epoch:          15290,
		StakingFlowV2Height:       2051863,
		EnableSlashingHeight:      2087789,
		EnableSlashingHeightV2:    1e9,
		Timeslot:                  10,
		EpochBreakPointSwapNewKey: []uint64{1280},
	},
	BeaconHeightBreakPointBurnAddr: 1,
	ReplaceStakingTxHeight:         1,
	ETHRemoveBridgeSigEpoch:        2085,
	BCHeightBreakPointNewZKP:       1148608,
	EnableFeatureFlags:             map[int]uint64{0: 1, 1: 0},
	BCHeightBreakPointPortalV3:     1328816,
	TxPoolVersion:                  0,
	GethParam: gethParam{
		Host:     "kovan.infura.io/v3/1138a1e99b154b10bae5c382ad894361",
		Protocol: "https",
		Port:     "",
	},
	BSCParam: bscParam{
		Host: "https://data-seed-prebsc-2-s1.binance.org:8545",
	},
	IsBackup: false,
}

var LocalDCSParam = &param{
	Name: "local-dcs",
	Net:  0x32,
	GenesisParam: &genesisParam{
		SelectBeaconNodeSerializedPubkeyV2:         map[uint64][]string{},
		SelectBeaconNodeSerializedPaymentAddressV2: map[uint64][]string{},
		SelectShardNodeSerializedPubkeyV2:          map[uint64][]string{},
		SelectShardNodeSerializedPaymentAddressV2:  map[uint64][]string{},
		FeePerTxKb:         0,
		ConsensusAlgorithm: "bls",
		BlockTimestamp:     "2020-08-11T00:00:00.000Z",
		TxStake:            "d0e731f55fa6c49f602807a6686a7ac769de4e04882bb5eaf8f4fe209f46535d",
	},
	CommitteeSize: committeeSize{
		MaxShardCommitteeSize:            12,
		MinShardCommitteeSize:            4,
		MaxBeaconCommitteeSize:           7,
		MinBeaconCommitteeSize:           4,
		InitShardCommitteeSize:           4,
		InitBeaconCommitteeSize:          4,
		ShardCommitteeSizeKeyListV2:      4,
		BeaconCommitteeSizeKeyListV2:     4,
		NumberOfFixedShardBlockValidator: 4,
	},
	BlockTime: blockTime{
		MinShardBlockInterval:  10 * time.Second,
		MaxShardBlockCreation:  8 * time.Second,
		MinBeaconBlockInterval: 10 * time.Second,
		MaxBeaconBlockCreation: 6 * time.Second,
	},
	StakingAmountShard: 1750000000000,
	ActiveShards:       8,
	BasicReward:        400000000,
	EpochParam: epochParam{
		NumberOfBlockInEpoch:   50,
		NumberOfBlockInEpochV2: 1e9,
		EpochV2BreakPoint:      1e9,
		RandomTime:             25,
		RandomTimeV2:           1e9,
	},
	EthContractAddressStr:            "0x2f6F03F1b43Eab22f7952bd617A24AB46E970dF7",
	BscContractAddressStr:            "0x2f6F03F1b43Eab22f7952bd617A24AB46E970dF7",
	IncognitoDAOAddress:              "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
	CentralizedWebsitePaymentAddress: "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
	SwapCommitteeParam: swapCommitteeParam{
		Offset:       1,
		SwapOffset:   1,
		AssignOffset: 2,
	},
	ConsensusParam: consensusParam{
		ConsensusV2Epoch:          1,
		StakingFlowV2Height:       1,
		EnableSlashingHeight:      1,
		EnableSlashingHeightV2:    210,
		Timeslot:                  10,
		EpochBreakPointSwapNewKey: []uint64{1280},
	},
	BeaconHeightBreakPointBurnAddr: 1,
	ReplaceStakingTxHeight:         1,
	ETHRemoveBridgeSigEpoch:        2085,
	BCHeightBreakPointNewZKP:       1148608,
	EnableFeatureFlags:             map[int]uint64{0: 1, 1: 0},
	BCHeightBreakPointPortalV3:     1328816,
	TxPoolVersion:                  0,
	GethParam: gethParam{
		Host:     "kovan.infura.io/v3/1138a1e99b154b10bae5c382ad894361",
		Protocol: "https",
		Port:     "",
	},
	BSCParam: bscParam{
		Host: "https://data-seed-prebsc-2-s1.binance.org:8545",
	},
	IsBackup: false,
}

func NewDefaultParam(network string) *param {
	var p *param
	initTx := new(initTx)
	switch network {
	case "mainnet":
		p = MainnetParam
		p.LoadKey(MainnetKeylist, Mainnetv2Keylist) //if there is keylist file in config folder, this default keylist will be not used
		initTx.load(MainnetInitTx)                  //if there is init_tx file in config folder, this default init_tx  will be not used
		p.GenesisParam.InitialIncognito = initTx.InitialIncognito
	case "testnet-1":
		p = Testnet1Param
		p.LoadKey(Testnet2Keylist, Testnet2v2Keylist)
		initTx.load(Testnet1InitTx)
		p.GenesisParam.InitialIncognito = initTx.InitialIncognito
	case "testnet-2":
		p = Tesnet2Param
		p.LoadKey(Testnet2Keylist, Testnet2v2Keylist)
		initTx.load(Testnet2InitTx)
		p.GenesisParam.InitialIncognito = initTx.InitialIncognito
	case "local":
		p = LocalParam
		p.LoadKey(LocalKeylist, Localv2Keylist)
		initTx.load(LocalInitTx)
		p.GenesisParam.InitialIncognito = initTx.InitialIncognito
	case "local-dcs":
		p = LocalDCSParam
		p.LoadKey(LocalKeylist, Localv2Keylist)
		initTx.load(LocalDCSInitTx)
		p.GenesisParam.InitialIncognito = initTx.InitialIncognito
	default:
		panic("Cannot recognize network")
	}
	return p
}