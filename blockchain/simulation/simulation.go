package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
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

	scenerioActions    []ScenerioAction
	autoGenerateBlocks map[int]bool
}

func main() {
	disableLog(true)
	instance1 := newSimInstance("test1")

	scnString := `{
		"Action":"GENERATEBLOCKS",
		"Params": [{"ChainID":0,"Blocks":100,"IsBlocking":true},{"ChainID":1,"Blocks":100,"IsBlocking":true}]
		}`
	scn := ScenerioAction{}
	err := json.Unmarshal([]byte(scnString), &scn)
	if err != nil {
		panic(err)
	}
	instance1.scenerioActions = append(instance1.scenerioActions, scn)
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

		autoGenerateBlocks: make(map[int]bool),
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
	sim.rpcServer.Stop()
	sim.cQuit <- struct{}{}
}

func (sim *simInstance) Run() {
	for _, action := range sim.scenerioActions {
		switch action.Action {
		case GENERATEBLOCKS:
			arrayParams := common.InterfaceSlice(action.Params)
			var wg sync.WaitGroup
			for _, param := range arrayParams {
				param := param.(map[string]interface{})
				p := GenerateBlocksParam{
					ChainID:    int(param["ChainID"].(float64)),
					Blocks:     int(param["Blocks"].(float64)),
					IsBlocking: param["IsBlocking"].(bool),
				}
				if p.IsBlocking {
					wg.Add(1)
				}
				go func(data GenerateBlocksParam) {
					if enable, ok := sim.autoGenerateBlocks[p.ChainID]; enable && ok {
						log.Fatalln(errors.New("Can't generate blocks if autogenerate is enable"))
						return
					}
					err := sim.GenerateBlocks(data.ChainID, data.Blocks)
					if err != nil {
						log.Fatalln(err)
					}
					if p.IsBlocking {
						wg.Done()
					}
				}(p)
			}
			wg.Wait()
		case AUTOGENERATEBLOCKS:
			arrayParams := common.InterfaceSlice(action.Params)
			for _, param := range arrayParams {
				param := param.(map[string]interface{})
				p := AutoGenerateBlocks{
					ChainID: int(param["ChainID"].(float64)),
					Enable:  param["Enable"].(bool),
				}
				if p.Enable {
					if enable, ok := sim.autoGenerateBlocks[p.ChainID]; !enable || !ok {
						go func(chainID int) {
							err := sim.GenerateBlocks(chainID, 0)
							if err != nil {
								log.Fatalln(err)
							}
						}(p.ChainID)
					}
				}
				sim.autoGenerateBlocks[p.ChainID] = p.Enable
			}
		case GENERATETXS:
			arrayParams := common.InterfaceSlice(action.Params)
			for _, param := range arrayParams {
				param := param.(map[string]interface{})
				receivers := param["Receivers"].(map[string]interface{})
				p := GenerateTxParam{
					SenderPrK: param["SenderPrk"].(string),
					Receivers: make(map[string]int),
				}
				for receiver, amount := range receivers {
					p.Receivers[receiver] = int(amount.(float64))
				}
				err := sim.generateTxs(p.SenderPrK, p.Receivers)
				if err != nil {
					log.Fatalln(err)
				}
			}
		case CREATETXSANDINJECT:
			arrayParams := common.InterfaceSlice(action.Params)
			for _, param := range arrayParams {
				data := CreateTxsAndInjectParam{
					InjectAt: make(map[int]int),
				}
				param := param.(map[string]interface{})
				injectAt := param["InjectAt"].(map[float64]float64)
				if len(injectAt) > 1 {
					log.Fatalln("")
					return
				}
				for i1, i2 := range injectAt {
					data.InjectAt[int(i1)] = int(i2)
				}

				txs := common.InterfaceSlice(param["Txs"])
				for _, tx := range txs {
					txParam := tx.(map[string]interface{})
					receivers := txParam["Receivers"].(map[string]interface{})
					p := GenerateTxParam{
						SenderPrK: txParam["SenderPrk"].(string),
						Receivers: make(map[string]int),
					}
					for receiver, amount := range receivers {
						p.Receivers[receiver] = int(amount.(float64))
					}
					err := sim.generateTxs(p.SenderPrK, p.Receivers)
					if err != nil {
						log.Fatalln(err)
					}
				}

			}
		case CHECKBALANCES:

		case CHECKBESTSTATES:

		case SWITCHTOMANUAL:

		}
	}
}

func disableLog(disable bool) {
	disableStdoutLog = disable
}

type TxReceiver struct {
	ReceiverPbK string
	Amount      int
}

func (sim *simInstance) generateTxs(senderPrk string, receivers map[string]int) error {
	tx, err := sim.createTx(senderPrk, receivers)
	if err != nil {
		return err
	}
	err = sim.injectTxs([]string{tx.Base58CheckData})
	if err != nil {
		return err
	}
	return nil
}

func (sim *simInstance) createAndInjectTx(senderPrk string, receivers map[string]int) error {
	return nil
}

func (sim *simInstance) createTx(senderPrk string, receivers map[string]int) (*jsonresult.CreateTransactionResult, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createtransaction",
		"params":  []interface{}{senderPrk, receivers, 1, 1},
		"id":      1,
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
			err := sim.createAndInsertBlock(chainID, blocks, currentTime)
			if err != nil {
				return err
			}
			blockCount++
			prevTimeSlot = common.CalculateTimeSlot(currentTime)
		}
		if enable, ok := sim.autoGenerateBlocks[chainID]; !enable && ok {
			return nil
		}
	}
}

func (sim *simInstance) createAndInsertBlock(chainID int, blocks int, currentTime int64) error {
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
