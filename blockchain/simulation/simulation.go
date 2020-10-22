package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	hooks struct {
		preCreateBlock  func()
		postCreateBlock func(block common.BlockInterface)
		preInsertBlock  func(block common.BlockInterface)
		postInsertBlock func()
	}
}

func main() {
	disableLog(true)
	instance1 := NewSimInstance("test1")
	scnString := `{
		"Action":"GENERATEBLOCKS",
		"Params": [{"ChainID":-1,"Blocks":500,"IsBlocking":true},{"ChainID":0,"Blocks":100,"IsBlocking":true},{"ChainID":1,"Blocks":100,"IsBlocking":true}]
		}`

	// scnString := `{
	// 	"Action":"CHECKBALANCES",
	// 	"Params":[{"PrivateKey":"112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or","IsBlocking":true}]
	// }`
	// scnString := `{
	// 	"Action":"SWITCHTOMANUAL"
	// }`

	scn := ScenerioAction{}
	err := json.Unmarshal([]byte(scnString), &scn)
	if err != nil {
		panic(err)
	}
	instance1.scenerioActions = append(instance1.scenerioActions, scn)
	instance1.run()
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

func NewSimInstance(simName string) *simInstance {
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
	bc := blockchain.BlockChain{}
	cs := mock.Consensus{}
	txpool := mempool.TxPool{}
	temppool := mempool.TxPool{}
	btcrd := mock.BTCRandom{} // use mock for now
	sync := mock.Syncker{
		Bc: &bc,
	}
	server := mock.Server{}
	ps := mock.Pubsub{}
	fees := make(map[byte]*mempool.FeeEstimator)
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

func (sim *simInstance) run() {
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
			createTxs := []GenerateTxParam{}
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
				createTxs = append(createTxs, p)
			}
			err := sim.GenerateTxs(createTxs)
			if err != nil {
				log.Fatalln(err)
			}
		case CREATETXSANDINJECT:
			arrayParams := common.InterfaceSlice(action.Params)
			for _, param := range arrayParams {
				data := CreateTxsAndInjectParam{}
				param := param.(map[string]interface{})
				injectAt := param["InjectAt"].(map[string]float64)
				data.InjectAt.ChainID = int(injectAt["ChainID"])
				data.InjectAt.Height = uint64(injectAt["Height"])
				txs := common.InterfaceSlice(param["Txs"])
				createTxs := []GenerateTxParam{}
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
					createTxs = append(createTxs, p)
				}

				err := sim.CreateAndInjectTx(createTxs, data.InjectAt.ChainID, data.InjectAt.Height)
				if err != nil {
					log.Fatalln(err)
				}
			}
		case CREATESTAKINGTX:
			arrayParams := common.InterfaceSlice(action.Params)
			txs := []string{}
			for _, param := range arrayParams {
				param := param.(map[string]interface{})
				data := CreateStakingTx{
					SenderPrk:   param["SenderPrk"].(string),
					MinerPrk:    param["MinerPrk"].(string),
					RewardAddr:  param["RewardAddr"].(string),
					StakeShard:  param["StakeShard"].(bool),
					AutoRestake: param["AutoRestake"].(bool),
				}
				tx, err := sim.CreateTxStaking(data)
				if err != nil {
					log.Fatalln(err)
				}
				txs = append(txs, tx.Base58CheckData)
			}
			err := sim.InjectTxs(txs)
			if err != nil {
				log.Fatalln(err)
			}
		case CHECKBALANCES:
			arrayParams := common.InterfaceSlice(action.Params)
			for _, param := range arrayParams {
				data := CheckBalanceParam{
					Tokens: make(map[string]uint64),
				}
				param := param.(map[string]interface{})
				data.PrivateKey = param["PrivateKey"].(string)
				data.IsBlocking = param["IsBlocking"].(bool)
				data.Interval = param["Interval"].(int)
				until := param["Until"].(map[string]float64)
				data.Until.ChainID = int(until["ChainID"])
				data.Until.Height = uint64(until["Height"])
				tokens := param["Tokens"].(map[string]interface{})
				for token, amount := range tokens {
					data.Tokens[token] = uint64(amount.(float64))
				}
				if data.IsBlocking {
					err := sim.CheckBalance(data)
					if err != nil {
						log.Fatalln(err)
					}
				} else {
					go func(d CheckBalanceParam) {
						err := sim.CheckBalance(d)
						if err != nil {
							log.Fatalln(err)
						}
					}(data)
				}
			}
		case CHECKBESTSTATES:

		case SWITCHTOMANUAL:
			sim.SwitchToManual()
		}
	}
}

func disableLog(disable bool) {
	disableStdoutLog = disable
}

func (sim *simInstance) GenerateTxs(createTxs []GenerateTxParam) error {
	txsInject := []string{}
	for _, createTxMeta := range createTxs {
		tx, err := sim.CreateTx(createTxMeta.SenderPrK, createTxMeta.Receivers)
		if err != nil {
			return err
		}
		txsInject = append(txsInject, tx.Base58CheckData)
	}
	err := sim.InjectTxs(txsInject)
	if err != nil {
		return err
	}
	return nil
}

func (sim *simInstance) CreateAndInjectTx(createTxs []GenerateTxParam, chainID int, height uint64) error {
	txsInject := []string{}
	for _, createTxMeta := range createTxs {
		tx, err := sim.CreateTx(createTxMeta.SenderPrK, createTxMeta.Receivers)
		if err != nil {
			return err
		}
		txsInject = append(txsInject, tx.Base58CheckData)
	}

	for {
		if chainID == -1 {
			if sim.bc.BeaconChain.GetBestView().GetHeight() >= height {
				err := sim.InjectTxs(txsInject)
				if err != nil {
					return err
				}
				return nil
			}
		} else {
			if sim.bc.ShardChain[byte(chainID)].GetBestView().GetHeight() >= height {
				err := sim.InjectTxs(txsInject)
				if err != nil {
					return err
				}
				return nil
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func (sim *simInstance) CreateTx(senderPrk string, receivers map[string]int) (*jsonresult.CreateTransactionResult, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createtransaction",
		"params":  []interface{}{senderPrk, receivers, 1, 1},
		"id":      1,
	})
	if err != nil {
		return nil, err
	}
	body, err := sendRequest(requestBody)
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

func (sim *simInstance) InjectTxs(txsBase58 []string) error {
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
			err := sim.createAndInsertBlock(chainID, currentTime)
			if err != nil {
				return err
			}
			blockCount++
			prevTimeSlot = common.CalculateTimeSlot(currentTime)
			if blockCount == blocks {
				return nil
			}
		}
		if enable, ok := sim.autoGenerateBlocks[chainID]; !enable && ok {
			return nil
		}
	}
}

func (sim *simInstance) createAndInsertBlock(chainID int, currentTime int64) error {
	newBlock, err := sim.CreateBlock(chainID, currentTime)
	if err != nil {
		return err
	}
	return sim.InsertBlock(newBlock, chainID)
}

func (sim *simInstance) CreateBlock(chainID int, currentTime int64) (common.BlockInterface, error) {
	var block common.BlockInterface

	if sim.hooks.preCreateBlock != nil {
		sim.hooks.preCreateBlock()
	}

	if chainID == -1 {
		newBlock, err := sim.bc.BeaconChain.CreateNewBlock(2, "", 1, currentTime)
		if err != nil {
			return nil, err
		}
		newBlock.(mock.BlockValidation).AddValidationField("test")

		block = newBlock
	} else {
		newBlock, err := sim.bc.ShardChain[byte(chainID)].CreateNewBlock(2, "", 1, currentTime)
		if err != nil {
			return nil, err
		}
		newBlock.(mock.BlockValidation).AddValidationField("test")
		block = newBlock
	}

	if sim.hooks.postCreateBlock != nil {
		sim.hooks.postCreateBlock(block)
	}
	return block, nil
}
func (sim *simInstance) InsertBlock(block common.BlockInterface, chainID int) error {

	if sim.hooks.preInsertBlock != nil {
		sim.hooks.preInsertBlock(block)
	}

	if chainID == -1 {
		err := sim.bc.InsertBeaconBlock(block.(*blockchain.BeaconBlock), true)
		if err != nil {
			return err
		}
	} else {
		err := sim.bc.InsertShardBlock(block.(*blockchain.ShardBlock), true)
		if err != nil {
			return err
		}
	}

	if sim.hooks.postInsertBlock != nil {
		sim.hooks.postInsertBlock()
	}
	return nil
}

func (sim *simInstance) GetBalance(privateKey string) (map[string]uint64, error) {
	tokenList := make(map[string]uint64)
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbalancebyprivatekey",
		"params":  []interface{}{privateKey},
		"id":      1,
	})
	if err != nil {
		return nil, err
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return nil, err
	}
	txResp := struct {
		Result uint64
	}{}
	err = json.Unmarshal(body, &txResp)
	if err != nil {
		return nil, err
	}
	tokenList["PRV"] = txResp.Result

	requestBody2, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getlistprivacycustomtokenbalance",
		"params":  []interface{}{privateKey},
		"id":      1,
	})
	if err != nil {
		return nil, err
	}
	body2, err := sendRequest(requestBody2)
	if err != nil {
		return nil, err
	}
	txResp2 := struct {
		Result jsonresult.ListCustomTokenBalance
	}{}
	err = json.Unmarshal(body2, &txResp2)
	if err != nil {
		return nil, err
	}
	for _, token := range txResp2.Result.ListCustomTokenBalance {
		tokenList[token.Name] = token.Amount
	}
	return tokenList, nil
}

func (sim *simInstance) CheckBalance(data CheckBalanceParam) error {
	tokenList, err := sim.GetBalance(data.PrivateKey)
	if err != nil {
		return err
	}
	amountNotMatch := make(map[string]uint64)
	for token, amount := range data.Tokens {
		if a, ok := tokenList[token]; ok {
			if a != amount {
				amountNotMatch[token] = a
			}
		}
	}
	if len(amountNotMatch) > 0 {
		if data.Until.Height != 0 {
			if data.Until.ChainID == -1 {
				if sim.bc.GetBeaconBestState().GetHeight() <= data.Until.Height {
					return sim.CheckBalance(data)
				}
				return errors.New("token balance not match")
			} else {
				if sim.bc.GetBestStateShard(byte(data.Until.ChainID)).GetHeight() <= data.Until.Height {
					return sim.CheckBalance(data)
				}
				return errors.New("token balance not match")
			}
		}
	}
	return nil
}

func (sim *simInstance) SwitchToManual() error {
	// tempTxs := []string{}
	// tempBlocks := make(map[int]common.BlockInterface)

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">Command: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(strings.ToLower(text), "\n")
		params := strings.Split(text, " ")
		switch params[0] {
		case MANUALEXIT:
			return nil
		case MANUALHELP:

		case MANUALCREATE:
			if len(params) >= 3 {
				//create block -1
				//create tx prv addr amount
				if len(params) >= 3 {
					switch params[1] {
					case "tx":
					case "block":
					}
				}
			}
			fmt.Println("Incorrect command")
		case MANUALINSERT:
			// insert blocks -1
			// insert txs
			if len(params) >= 2 {
				switch params[1] {
				case "tx":
				case "block":
				}
			}
			fmt.Println("Incorrect command")
		case MANUALGENERATE:
			// generate tx prv addr amount
			// generate	block -1 amount
			if len(params) >= 4 {
				switch params[1] {
				case "tx":
				case "block":
				}
			}
			fmt.Println("Incorrect command")
		case MANUALGETBALANCE:
			// balance prk
			if len(params) == 2 {
				balances, err := sim.GetBalance(params[1])
				if err != nil {
					fmt.Println("Err:", err)
					break
				}
				fmt.Println(balances)
			}
			fmt.Println("Incorrect command")
		case MANUALGETBESTSTATE:
			// beststate -1
			if len(params) == 2 {

			}
			fmt.Println("Incorrect command")
		default:
			fmt.Println("Unknown command")
		}
	}
}

func sendRequest(requestBody []byte) ([]byte, error) {
	resp, err := http.Post("http://0.0.0.0:8000", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
