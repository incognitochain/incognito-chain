package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/simulation/mock"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/memcache"
	"github.com/incognitochain/incognito-chain/metadata"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
)

type simInstance struct {
	simName     string
	dbDir       string
	param       *blockchain.Params
	bc          *blockchain.BlockChain
	cs          *mock.Consensus
	txpool      *mock.TxPool
	temppool    *mock.TxPool
	btcrd       *mock.BTCRandom
	sync        *mock.Syncker
	server      *mock.Server
	cPendingTxs chan metadata.Transaction
	cRemovedTxs chan metadata.Transaction
	cQuit       chan struct{}
}

func main() {
	instance1 := newSimInstance("test1")
	instance1.Run()
	instance1.Stop()
}

func getBTCRelayingChain(btcRelayingChainID string, btcDataFolderName string, dataFolder string) (*btcrelaying.BlockChain, error) {
	relayingChainParams := map[string]*chaincfg.Params{
		blockchain.TestnetBTCChainID:  btcrelaying.GetTestNet3Params(),
		blockchain.Testnet2BTCChainID: btcrelaying.GetTestNet3ParamsForInc2(),
		blockchain.MainnetBTCChainID:  btcrelaying.GetMainNetParams(),
	}
	relayingChainGenesisBlkHeight := map[string]int32{
		blockchain.TestnetBTCChainID:  int32(1833130),
		blockchain.Testnet2BTCChainID: int32(1833130),
		blockchain.MainnetBTCChainID:  int32(634140),
	}
	return btcrelaying.GetChainV2(
		filepath.Join("./"+dataFolder, btcDataFolderName),
		relayingChainParams[btcRelayingChainID],
		relayingChainGenesisBlkHeight[btcRelayingChainID],
	)
}

func getBNBRelayingChainState(bnbRelayingChainID string, dataFolder string) (*bnbrelaying.BNBChainState, error) {
	bnbChainState := new(bnbrelaying.BNBChainState)
	err := bnbChainState.LoadBNBChainState(
		filepath.Join("./"+dataFolder, "bnbrelayingv3"),
		bnbRelayingChainID,
	)
	if err != nil {
		log.Printf("Error getBNBRelayingChainState: %v\n", err)
		return nil, err
	}
	return bnbChainState, nil
}

func newSimInstance(simName string) *simInstance {
	log.Printf("Creating sim %v instance...\n", simName)
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	initLogRotator(filepath.Join(path, simName+"/"+simName+".log"))
	activeNetParams := &blockchain.ChainTest2Param
	cs := mock.Consensus{}
	txpool := mock.TxPool{}
	temppool := mock.TxPool{}
	btcrd := mock.BTCRandom{} // use mock for now
	sync := mock.Syncker{}
	server := mock.Server{}
	ps := mock.Pubsub{}
	fees := make(map[byte]blockchain.FeeEstimator)
	bc := blockchain.BlockChain{}
	for i := byte(0); i < byte(activeNetParams.ActiveShards); i++ {
		fees[i] = &mock.Fee{}
	}
	cPendingTxs := make(chan metadata.Transaction, 500)
	cRemovedTxs := make(chan metadata.Transaction, 500)
	cQuit := make(chan struct{})
	blockgen, err := blockchain.NewBlockGenerator(&txpool, &bc, &sync, cPendingTxs, cRemovedTxs)
	if err != nil {
		panic(err)
	}

	db, err := incdb.OpenMultipleDB("leveldb", filepath.Join("./"+simName, "database"))
	// Create db and use it.
	if err != nil {
		panic(err)
	}

	btcChain, err := getBTCRelayingChain(activeNetParams.BTCRelayingHeaderChainID, "btcchain", simName)
	if err != nil {
		panic(err)
	}
	bnbChainState, err := getBNBRelayingChainState(activeNetParams.BNBRelayingHeaderChainID, simName)
	if err != nil {
		panic(err)
	}

	err = bc.Init(&blockchain.Config{
		BTCChain:        btcChain,
		BNBChainState:   bnbChainState,
		ChainParams:     activeNetParams,
		DataBase:        db,
		MemCache:        memcache.New(),
		BlockGen:        blockgen,
		TxPool:          &txpool,
		TempTxPool:      &temppool,
		Server:          &server,
		Syncker:         &sync,
		PubSubManager:   &ps,
		FeeEstimator:    make(map[byte]blockchain.FeeEstimator),
		RandomClient:    &btcrd,
		ConsensusEngine: &cs,
		GenesisParams:   blockchain.GenesisParam,
	})
	if err != nil {
		panic(err)
	}
	bc.InitChannelBlockchain(cRemovedTxs)

	sim := &simInstance{
		simName:     simName,
		param:       activeNetParams,
		bc:          &bc,
		cs:          &cs,
		txpool:      &txpool,
		temppool:    &temppool,
		btcrd:       &btcrd,
		sync:        &sync,
		server:      &server,
		cPendingTxs: cPendingTxs,
		cRemovedTxs: cRemovedTxs,
		cQuit:       cQuit,
	}

	go func() {
		for {
			select {
			case <-cQuit:
				return
			case <-cRemovedTxs:
			}

		}
	}()
	go blockgen.Start(cQuit)

	log.Printf("Done sim %v instance\n", simName)
	return sim
}

func (sim *simInstance) Stop() {
	sim.cQuit <- struct{}{}
}

func (sim *simInstance) Run() {
	prevTimeSlot := int64(0)
	for {
		currentTime := time.Now().Unix()
		currentTimeSlot := common.CalculateTimeSlot(currentTime)
		newTimeSlot := false
		if prevTimeSlot != currentTimeSlot {
			newTimeSlot = true
		}
		if newTimeSlot {
			newBlock, err := sim.bc.ShardChain[0].CreateNewBlock(2, "", 1, currentTime)
			if err != nil {
				panic(err)
			}
			newBlock.(mock.BlockValidation).AddValidationField("test")
			err = sim.bc.InsertShardBlock(newBlock.(*blockchain.ShardBlock), true)
			if err != nil {
				panic(err)
			}
			prevTimeSlot = common.CalculateTimeSlot(currentTime)
			if newBlock.GetHeight() == 5 {
				break
			}
		}
	}

}
