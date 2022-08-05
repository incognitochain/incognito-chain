package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/pruner"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
	"math/rand"
)

func Test_Prune() {
	cfg := testsuite.Config{
		DataDir: "./data/",
		Network: testsuite.ID_TESTNET2,
		ResetDB: false,
	}

	node := testsuite.InitChainParam(cfg, func() {
		config.Param().ActiveShards = 2
		config.Param().BCHeightBreakPointNewZKP = 1
		config.Param().BCHeightBreakPointPrivacyV2 = 2
		config.Param().BeaconHeightBreakPointBurnAddr = 1
		config.Param().ConsensusParam.EnableSlashingHeightV2 = 2
		config.Param().ConsensusParam.StakingFlowV2Height = 5
		config.Param().ConsensusParam.AssignRuleV3Height = 10
		config.Param().ConsensusParam.StakingFlowV3Height = 15
		config.Param().CommitteeSize.MaxShardCommitteeSize = 16
		config.Param().CommitteeSize.MinShardCommitteeSize = 4
		config.Param().CommitteeSize.NumberOfFixedShardBlockValidator = 4
		config.Param().ConsensusParam.ConsensusV2Epoch = 1
		config.Param().EpochParam.NumberOfBlockInEpoch = 20
		config.Param().EpochParam.RandomTime = 10
		config.Param().ConsensusParam.EpochBreakPointSwapNewKey = []uint64{1e9}
		config.Config().LimitFee = 0
		config.Param().PDexParams.Pdexv3BreakPointHeight = 1e9
		config.Param().TxPoolVersion = 0
		config.Config().StateBloomSize = 2000
	}, func(node *testsuite.NodeEngine) {
		db := make(map[int]incdb.Database)
		db[0] = node.GetBlockchain().GetShardChainDatabase(0)
		db[1] = node.GetBlockchain().GetShardChainDatabase(1)

		p := pruner.NewPrunerWithValue(db, make(map[byte]byte))
		p.Prune()
		node.Pause()
	})

	genRandomString := func(strLen int) string {
		characters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
		res := ""
		for i := 0; i < strLen; i++ {
			u := string(characters[rand.Int()%len(characters)])
			res = res + u
		}
		return res
	}

	sb := node.GetBlockchain().GetChain(0).(*blockchain.ShardChain).GetBestView().(*blockchain.ShardBestState)
	txDB := sb.GetCopiedTransactionStateDB()
	lastRoot := common.Hash{}

	for i := 0; i < 1000; i++ {
		for j := 0; j < 100; j++ {
			r := genRandomString(1000)
			err := txDB.SetStateObject(statedb.TestObjectType, common.HashH([]byte(r)), []byte(r))
			if err != nil {
				panic(err)
			}
		}
		var err error
		lastRoot, err = txDB.Commit(true)
		if err != nil {
			panic(err)
		}
		err = txDB.Database().TrieDB().Commit(lastRoot, false)
		if err != nil {
			panic(err)
		}
		txDB.ClearObjects()
		fmt.Println("create i", i, lastRoot.String())
	}
	sb.SetTransactonDB(lastRoot, txDB)

	//send PRV
	for i := 0; i < 5; i++ {
		acc := node.NewAccountFromShard(0)
		node.RPC.API_SubmitKey(acc.PrivateKey)
		node.SendPRV(node.GenesisAccount, acc, float64(100))
		node.GenerateBlock().NextRound()
		node.GenerateBlock().NextRound()
		fmt.Println("send PRV ", acc.Name)
		node.GenerateBlock().NextRound()
	}
	db := make(map[int]incdb.Database)
	db[0] = node.GetBlockchain().GetShardChainDatabase(0)
	db[1] = node.GetBlockchain().GetShardChainDatabase(1)

	p := pruner.NewPrunerWithValue(db, make(map[byte]byte))
	p.Prune()

	sb = node.GetBlockchain().GetChain(0).(*blockchain.ShardChain).GetBestView().(*blockchain.ShardBestState)
	txDB = sb.GetCopiedTransactionStateDB()
	lastRoot = common.Hash{}

	for i := 0; i < 100; i++ {
		for j := 0; j < 10; j++ {
			r := genRandomString(1000)
			err := txDB.SetStateObject(statedb.TestObjectType, common.HashH([]byte(r)), []byte(r))
			if err != nil {
				panic(err)
			}
		}
		var err error
		lastRoot, err = txDB.Commit(true)
		if err != nil {
			panic(err)
		}
		err = txDB.Database().TrieDB().Commit(lastRoot, false)
		if err != nil {
			panic(err)
		}
		txDB.ClearObjects()
		fmt.Println("create i", i, lastRoot.String())
	}
	sb.SetTransactonDB(lastRoot, txDB)

	//send PRV
	for i := 0; i < 5; i++ {
		acc := node.NewAccountFromShard(0)
		node.RPC.API_SubmitKey(acc.PrivateKey)
		node.SendPRV(node.GenesisAccount, acc, float64(100))
		node.GenerateBlock().NextRound()
		node.GenerateBlock().NextRound()
		fmt.Println("send PRV ", acc.Name)
		node.GenerateBlock().NextRound()
	}

}
