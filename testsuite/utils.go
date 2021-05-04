package devframework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"io"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/incognitochain/incognito-chain/blockchain"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"

	"github.com/incognitochain/incognito-chain/testsuite/account"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

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
		filepath.Join(dataFolder, btcDataFolderName),
		relayingChainParams[btcRelayingChainID],
		relayingChainGenesisBlkHeight[btcRelayingChainID],
	)
}

func getBNBRelayingChainState(bnbRelayingChainID string, dataFolder string) (*bnbrelaying.BNBChainState, error) {
	bnbChainState := new(bnbrelaying.BNBChainState)
	err := bnbChainState.LoadBNBChainState(
		filepath.Join(dataFolder, "bnbrelayingv3"),
		bnbRelayingChainID,
	)
	if err != nil {
		log.Printf("Error getBNBRelayingChainState: %v\n", err)
		return nil, err
	}
	return bnbChainState, nil
}

func createGenesisTx(accounts []account.Account) []string {
	transactions := []string{}
	db, err := incdb.Open("leveldb", "/tmp/"+time.Now().UTC().String())
	if err != nil {
		fmt.Print("could not open connection to leveldb")
		fmt.Print(err)
		panic(err)
	}
	stateDB, _ := statedb.NewWithPrefixTrie(common.EmptyRoot, statedb.NewDatabaseAccessWarper(db))
	initPRV := int(1000000000000 * math.Pow(10, 9))
	for _, account := range accounts {
		txs := initSalaryTx(strconv.Itoa(initPRV), account.PrivateKey, stateDB)
		transactions = append(transactions, txs[0])
	}
	return transactions
}

func initSalaryTx(amount string, privateKey string, stateDB *statedb.StateDB) []string {
	var initTxs []string
	var initAmount, _ = strconv.Atoi(amount) // amount init
	testUserkeyList := []string{
		privateKey,
	}
	for _, val := range testUserkeyList {

		testUserKey, _ := wallet.Base58CheckDeserialize(val)
		testSalaryTX := transaction.Tx{}
		testSalaryTX.InitTxSalary(uint64(initAmount), &testUserKey.KeySet.PaymentAddress, &testUserKey.KeySet.PrivateKey,
			stateDB,
			nil,
		)
		initTx, _ := json.Marshal(testSalaryTX)
		initTxs = append(initTxs, string(initTx))
	}
	return initTxs
}

type JsonRequest struct {
	Jsonrpc string      `json:"Jsonrpc"`
	Method  string      `json:"Method"`
	Params  interface{} `json:"Params"`
	Id      interface{} `json:"Id"`
}

func makeRPCDownloadRequest(address string, method string, w io.Writer, params ...interface{}) error {
	request := JsonRequest{
		Jsonrpc: "1.0",
		Method:  method,
		Params:  params,
		Id:      "1",
	}
	requestBytes, err := json.Marshal(&request)
	if err != nil {
		return err
	}
	fmt.Println(string(requestBytes))
	resp, err := http.Post(address, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		fmt.Println(err)
		return err
	}
	n, err := io.Copy(w, resp.Body)
	fmt.Println(n, err)
	if err != nil {
		return err
	}
	return nil
}

func (sim *NodeEngine) SendPRV(args ...interface{}) (string, error) {
	var sender string
	var receivers = make(map[string]uint64)
	for i, arg := range args {
		if i == 0 {
			sender = arg.(account.Account).PrivateKey
		} else {
			switch arg.(type) {
			default:
				if i%2 == 1 {
					amount, ok := args[i+1].(float64)
					if !ok {
						amountF64 := args[i+1].(float64)
						amount = amountF64
					}
					receivers[arg.(account.Account).PaymentAddress] = uint64(amount)
				}
			}
		}
	}
	res, err := sim.RPC.API_SendTxPRV(sender, receivers, -1, true)
	if err != nil {
		fmt.Println(err)
		sim.Pause()
	}
	return res.TxID, nil
}

func (sim *NodeEngine) ShowBalance(acc account.Account) map[string]uint64 {
	res, err := sim.RPC.API_GetBalance(acc)
	if err != nil {
		fmt.Println(err)
	}
	return res
}

func GenerateCommitteeIndex(nCommittee int) []int {
	res := []int{}
	for i := 0; i < nCommittee; i++ {
		res = append(res, i)
	}
	return res
}

func (sim *NodeEngine) TrackAccount(acc account.Account) (int, int) {
	chain := sim.GetBlockchain()
	shardWaitingList, _ := incognitokey.CommitteeKeyListToString(chain.BeaconChain.GetShardsWaitingList())
	shardPendingList := make(map[int][]string)
	shardCommitteeList := make(map[int][]string)
	for sid := 0; sid < chain.GetActiveShardNumber(); sid++ {
		shardPendingList[sid], _ = incognitokey.CommitteeKeyListToString(chain.BeaconChain.GetShardsPendingList()[sid])
		shardCommitteeList[sid], _ = incognitokey.CommitteeKeyListToString(chain.BeaconChain.GetShardsCommitteeList()[sid])
	}
	//fmt.Println("shardWaitingList", shardWaitingList)
	//fmt.Println("shardPendingList", shardPendingList)
	//fmt.Println("shardCommitteeList", shardCommitteeList)
	for sid := 0; sid < chain.GetActiveShardNumber(); sid++ {
		if common.IndexOfStr(acc.SelfCommitteePubkey, shardPendingList[sid]) > -1 {
			return 1, common.IndexOfStr(acc.SelfCommitteePubkey, shardPendingList[sid])
		}
		if common.IndexOfStr(acc.SelfCommitteePubkey, shardCommitteeList[sid]) > -1 {
			return 2, common.IndexOfStr(acc.SelfCommitteePubkey, shardCommitteeList[sid])
		}
	}

	if common.IndexOfStr(acc.SelfCommitteePubkey, shardWaitingList) > -1 {
		return 0, common.IndexOfStr(acc.SelfCommitteePubkey, shardWaitingList)
	}
	return -1, -1
}
