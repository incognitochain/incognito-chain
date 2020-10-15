package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/simulation/mock"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/memcache"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/metadata"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
)

type simInstance struct {
	simName     string
	dbDir       string
	param       *blockchain.Params
	bc          *blockchain.BlockChain
	cs          *mock.Consensus
	txpool      *mempool.TxPool
	temppool    *mempool.TxPool
	btcrd       *mock.BTCRandom
	sync        *mock.Syncker
	server      *mock.Server
	cPendingTxs chan metadata.Transaction
	cRemovedTxs chan metadata.Transaction
	rpcServer   *rpcserver.RpcServer
	cQuit       chan struct{}
}

type simSession struct {
	instance        *simInstance
	scenerioActions []ScenerioAction
}

func main() {
	disableLog(true)
	instance1 := newSimInstance("test1")
	instance1.Run()
	select {}
	// instance1.Stop()
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
	dbLogger.SetLevel(common.LevelTrace)
	blockchainLogger.SetLevel(common.LevelTrace)
	bridgeLogger.SetLevel(common.LevelTrace)
	rpcLogger.SetLevel(common.LevelTrace)
	rpcServiceLogger.SetLevel(common.LevelTrace)
	rpcServiceBridgeLogger.SetLevel(common.LevelTrace)
	transactionLogger.SetLevel(common.LevelTrace)
	privacyLogger.SetLevel(common.LevelTrace)
	mempoolLogger.SetLevel(common.LevelTrace)
	activeNetParams := &blockchain.ChainTest2Param
	cs := mock.Consensus{}
	txpool := mempool.TxPool{}
	temppool := mempool.TxPool{}
	btcrd := mock.BTCRandom{} // use mock for now
	sync := mock.Syncker{}
	server := mock.Server{}
	ps := mock.Pubsub{}
	fees := make(map[byte]*mempool.FeeEstimator)
	bc := blockchain.BlockChain{}
	for i := byte(0); i < byte(activeNetParams.ActiveShards); i++ {
		fees[i] = mempool.NewFeeEstimator(
			mempool.DefaultEstimateFeeMaxRollback,
			mempool.DefaultEstimateFeeMinRegisteredBlocks,
			0)
	}
	cPendingTxs := make(chan metadata.Transaction, 500)
	cRemovedTxs := make(chan metadata.Transaction, 500)
	cQuit := make(chan struct{})
	blockgen, err := blockchain.NewBlockGenerator(&txpool, &bc, &sync, cPendingTxs, cRemovedTxs)
	if err != nil {
		panic(err)
	}

	listenFunc := net.Listen
	listener, err := listenFunc("tcp", "0.0.0.0:8000")
	if err != nil {
		panic(err)
	}
	rpcConfig := rpcserver.RpcServerConfig{
		HttpListenters: []net.Listener{listener},
		RPCMaxClients:  1,
		DisableAuth:    true,
		ChainParams:    activeNetParams,
		BlockChain:     &bc,
		Blockgen:       blockgen,
		TxMemPool:      &txpool,
	}
	rpcServer := &rpcserver.RpcServer{}

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

	txpool.Init(&mempool.Config{
		ConsensusEngine: &cs,
		BlockChain:      &bc,
		DataBase:        db,
		ChainParams:     activeNetParams,
		FeeEstimator:    fees,
		TxLifeTime:      100,
		MaxTx:           1000,
		// DataBaseMempool:   dbmp,
		IsLoadFromMempool: false,
		PersistMempool:    false,
		RelayShards:       nil,
		PubSubManager:     &ps,
	})
	// serverObj.blockChain.AddTxPool(serverObj.memPool)
	txpool.InitChannelMempool(cPendingTxs, cRemovedTxs)

	temppool.Init(&mempool.Config{
		BlockChain:    &bc,
		DataBase:      db,
		ChainParams:   activeNetParams,
		FeeEstimator:  fees,
		MaxTx:         1000,
		PubSubManager: &ps,
	})
	txpool.IsBlockGenStarted = true
	go temppool.Start(cQuit)
	go txpool.Start(cQuit)

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
		rpcServer:   rpcServer,
		cQuit:       cQuit,
	}

	rpcServer.Init(&rpcConfig)

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
	go rpcServer.Start()

	log.Printf("Done sim %v instance\n", simName)
	return sim
}

func (sim *simInstance) Stop() {
	sim.cQuit <- struct{}{}
}

func (sim *simInstance) Run() {
	tx, err := sim.createTx("", nil)
	if err != nil {
		log.Println(err)
	}
	log.Println(tx)
	err = sim.injectTxs([]string{tx.Base58CheckData})
	if err != nil {
		panic(err)
	}
	err = sim.GenerateBlocks(0, 5)
	if err != nil {
		panic(err)
	}
}

func disableLog(disable bool) {
	disableStdoutLog = disable
}

type TxReceiver struct {
	ReceiverPbK string
	Amount      int
}

func (sim *simInstance) createTx(senderPrk string, receivers []TxReceiver) (*jsonresult.CreateTransactionResult, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createtransaction",
		"params": []interface{}{"112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or", map[string]int{
			"12Rtc3sbfHTTSqmS8efnhgb7Rc6ineoQCwJyX63MMRK4HF6JGo51GJp5rk25QfviU7GPjyptT9q3JguQmDEG3uKpPUDEY5CSUJtttfU": 10000,
		}, 1, 1},
		"id": 1,
	})
	if err != nil {
		return nil, err
	}
	resp, err := http.Post("http://0.0.0.0:8000", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	txResp := struct {
		Result jsonresult.CreateTransactionResult
	}{}
	err = json.Unmarshal(body, &txResp)
	if err != nil {
		return nil, err
	}
	return &txResp.Result, nil
}

func (sim *simInstance) injectTxs(txsBase58 []string) error {
	for _, txB58Check := range txsBase58 {
		rawTxBytes, _, err := base58.Base58Check{}.Decode(txB58Check)
		if err != nil {
			return err
		}
		var tx transaction.Tx
		err = json.Unmarshal(rawTxBytes, &tx)
		if err != nil {
			return err
		}
		sim.cPendingTxs <- &tx
	}

	return nil
}

func (sim *simInstance) GenerateBlocks(chainID int, blocks int) error {
	prevTimeSlot := int64(0)
	blockCount := 0
	for {
		currentTime := time.Now().Unix()
		currentTimeSlot := common.CalculateTimeSlot(currentTime)
		newTimeSlot := false
		if prevTimeSlot != currentTimeSlot {
			newTimeSlot = true
		}
		if newTimeSlot {
			if chainID == -1 {
				newBlock, err := sim.bc.BeaconChain.CreateNewBlock(2, "", 1, currentTime)
				if err != nil {
					return err
				}
				newBlock.(mock.BlockValidation).AddValidationField("test")
				err = sim.bc.InsertBeaconBlock(newBlock.(*blockchain.BeaconBlock), true)
				if err != nil {
					return err
				}
				blockCount++
				prevTimeSlot = common.CalculateTimeSlot(currentTime)
			} else {
				newBlock, err := sim.bc.ShardChain[byte(chainID)].CreateNewBlock(2, "", 1, currentTime)
				if err != nil {
					return err
				}
				newBlock.(mock.BlockValidation).AddValidationField("test")
				err = sim.bc.InsertShardBlock(newBlock.(*blockchain.ShardBlock), true)
				if err != nil {
					return err
				}
				blockCount++
				prevTimeSlot = common.CalculateTimeSlot(currentTime)
			}
			if blockCount == blocks {
				break
			}
		}
	}
	return nil
}

func (sim *simInstance) SwitchToManual() error {
	return nil
}

func (sim *simInstance) manualCreateBlock() error {
	return nil
}

func (sim *simInstance) manualInjectBlock(chainID int, block common.BlockInterface) error {
	return nil
}
