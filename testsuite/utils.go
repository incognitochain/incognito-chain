package devframework

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/portal"

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
		portal.TestnetBTCChainID:  btcrelaying.GetTestNet3Params(),
		portal.Testnet2BTCChainID: btcrelaying.GetTestNet3ParamsForInc2(),
		portal.MainnetBTCChainID:  btcrelaying.GetMainNetParams(),
	}
	relayingChainGenesisBlkHeight := map[string]int32{
		portal.TestnetBTCChainID:  int32(1896910),
		portal.Testnet2BTCChainID: int32(1863675),
		portal.MainnetBTCChainID:  int32(634140),
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

func createGenesisTx(accounts []account.Account) []config.InitialIncognito {
	transactions := []config.InitialIncognito{}
	db, err := incdb.Open("leveldb", "/tmp/"+time.Now().UTC().String())
	if err != nil {
		fmt.Print("could not open connection to leveldb")
		fmt.Print(err)
		panic(err)
	}
	stateDB, _ := statedb.NewWithPrefixTrie(common.EmptyRoot, statedb.NewDatabaseAccessWarper(db))
	initPRV := 1e18
	for _, account := range accounts {
		txs := initSalaryTx(strconv.Itoa(int(initPRV)), account.PrivateKey, stateDB)
		transactions = append(transactions, txs[0])
	}
	return transactions
}

func initSalaryTx(amount string, privateKey string, stateDB *statedb.StateDB) []config.InitialIncognito {
	var initTxs = []config.InitialIncognito{}
	var initAmount, _ = strconv.Atoi(amount) // amount init
	testUserkeyList := []string{
		privateKey,
	}
	for _, val := range testUserkeyList {
		testUserKey, _ := wallet.Base58CheckDeserialize(val)
		testSalaryTX := transaction.TxVersion1{}
		testSalaryTX.InitTxSalary(uint64(initAmount), &testUserKey.KeySet.PaymentAddress, &testUserKey.KeySet.PrivateKey,
			stateDB,
			nil,
		)

		proofByte := b64.StdEncoding.EncodeToString([]byte(testSalaryTX.Proof.Bytes()))
		sig := b64.StdEncoding.EncodeToString([]byte(testSalaryTX.Sig))
		sigPub := b64.StdEncoding.EncodeToString([]byte(testSalaryTX.SigPubKey))
		initTx := config.InitialIncognito{
			Version:              int(testSalaryTX.Version),
			Type:                 testSalaryTX.Type,
			LockTime:             uint64(testSalaryTX.LockTime),
			Fee:                  int(testSalaryTX.Fee),
			Info:                 string(testSalaryTX.Info),
			SigPubKey:            string(sigPub),
			Sig:                  string(sig),
			Proof:                string(proofByte),
			PubKeyLastByteSender: int(testSalaryTX.PubKeyLastByteSender),
			Metadata:             nil,
		}
		initTxs = append(initTxs, initTx)
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

func (sim *NodeEngine) SendPRVToMultiAccs(args ...interface{}) (string, error) {
	var sender string
	var receivers = make(map[string]uint64)
	argsNew := args[0].([]interface{})

	for i, arg := range argsNew {
		if i == 0 {
			sender = arg.(account.Account).PrivateKey
		} else {
			switch arg.(type) {
			default:
				if i%2 == 1 {
					amount, ok := argsNew[i+1].(float64)
					if !ok {
						amountF64 := argsNew[i+1].(float64)
						amount = amountF64
					}
					receivers[arg.(account.Account).PaymentAddress] = uint64(amount)
				}
			}
		}
	}
	res, err := sim.RPC.API_SendTxPRV(sender, receivers, -1, true)
	x, e := sim.RPC.Client.GetBalanceByPrivateKey(sender)
	fmt.Printf("Balance of account %v is %+v, err %+v\n", sender, x, e)
	if err != nil {
		fmt.Println(err)
		sim.Pause()
	}
	return res.TxID, nil
}

func (sim *NodeEngine) ShowBalance(acc account.Account) map[string]uint64 {
	res, err := sim.RPC.API_GetBalance(acc)
	fmt.Println(res, err)
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

func (sim *NodeEngine) ShowAccountPosition(accounts []account.Account) {
	chain := sim.GetBlockchain()
	type AccountInfo struct {
		Name  string
		Queue string
	}
	pkMap := make(map[string]*AccountInfo)
	for _, acc := range accounts {
		pkMap[acc.SelfCommitteePubkey] = &AccountInfo{acc.Name, ""}
	}
	shardWaitingList, _ := incognitokey.CommitteeKeyListToString(chain.BeaconChain.GetShardsWaitingList())
	tmp := ""
	for _, pk := range shardWaitingList {
		if pkMap[pk] != nil && pkMap[pk].Name != "" {
			tmp += pkMap[pk].Name + " "
			pkMap[pk].Queue = "waiting"
		} else {
			tmp += "@@ "
		}
	}
	fmt.Printf("WaitingList: %v\n", tmp)

	shardPendingList := make(map[int][]string)
	shardCommitteeList := make(map[int][]string)
	shardSyncingList := make(map[int][]string)
	for sid := 0; sid < chain.GetActiveShardNumber(); sid++ {
		shardPendingList[sid], _ = incognitokey.CommitteeKeyListToString(chain.BeaconChain.GetShardsPendingList()[common.BlsConsensus][common.GetShardChainKey(byte(sid))])
		shardCommitteeList[sid], _ = incognitokey.CommitteeKeyListToString(chain.BeaconChain.GetAllCommittees()[common.BlsConsensus][common.GetShardChainKey(byte(sid))])
		shardSyncingList[sid], _ = incognitokey.CommitteeKeyListToString(chain.BeaconChain.GetSyncingValidators()[byte(sid)])
	}

	for sid := 0; sid < chain.GetActiveShardNumber(); sid++ {
		tmp = ""
		for _, pk := range shardSyncingList[sid] {
			if pkMap[pk] != nil && pkMap[pk].Name != "" {
				tmp += pkMap[pk].Name + " "
				pkMap[pk].Queue = "syncing"
			} else {
				tmp += "@@ "
			}
		}
		fmt.Printf("Synching Shard %v: %v\n", sid, tmp)

		tmp = ""
		for _, pk := range shardPendingList[sid] {
			if pkMap[pk] != nil && pkMap[pk].Name != "" {
				tmp += pkMap[pk].Name + " "
				pkMap[pk].Queue = "pending"
			} else {
				tmp += "@@ "
			}
		}
		fmt.Printf("Pending Shard %v: %v\n", sid, tmp)
	}

	for sid := 0; sid < chain.GetActiveShardNumber(); sid++ {
		tmp = ""
		for _, pk := range shardCommitteeList[sid] {
			if pkMap[pk] != nil && pkMap[pk].Name != "" {
				tmp += pkMap[pk].Name + " "
				pkMap[pk].Queue = "committee"
			} else {
				tmp += "@@ "
			}
		}
		fmt.Printf("Committee Shard %v: %v\n", sid, tmp)
	}

	tmp = ""
	for _, acc := range accounts {
		if pkMap[acc.SelfCommitteePubkey].Queue == "" {
			tmp += acc.Name + " "
		}
	}
	fmt.Printf("Unstake: %v\n", tmp)
}

func (sim *NodeEngine) TrackAccount(acc account.Account) (int, int) {
	chain := sim.GetBlockchain()
	shardWaitingList, _ := incognitokey.CommitteeKeyListToString(chain.BeaconChain.GetShardsWaitingList())
	shardPendingList := make(map[int][]string)
	shardCommitteeList := make(map[int][]string)
	for sid := 0; sid < chain.GetActiveShardNumber(); sid++ {
		shardPendingList[sid], _ = incognitokey.CommitteeKeyListToString(chain.BeaconChain.GetShardsPendingList()[common.BlsConsensus][common.GetShardChainKey(byte(sid))])
		shardCommitteeList[sid], _ = incognitokey.CommitteeKeyListToString(chain.BeaconChain.GetAllCommittees()[common.BlsConsensus][common.GetShardChainKey(byte(sid))])
	}
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

func (sim *NodeEngine) PrintChainInfo(chainIDs []int) {
	chain := sim.GetBlockchain()
	beacon := chain.BeaconChain
	for _, cid := range chainIDs {
		if cid == -1 {
			fmt.Printf("Beacon chain height %v hash %v epoch %v\n", beacon.CurrentHeight(), beacon.GetBestViewHash(), beacon.GetEpoch())
		} else {
			shard := chain.GetChain(cid).(*blockchain.ShardChain)
			fmt.Printf("Shard chain %v height %v hash %v beacon height %v committeefromblock %v\n", shard.GetShardID(), shard.CurrentHeight(), shard.GetBestViewHash(), shard.GetBestState().BeaconHeight, shard.GetBestState().CommitteeFromBlock().String())
		}
	}
}

func (node *NodeEngine) GenerateFork2Branch(chainID int, foo func()) (multiview.View, multiview.View) {
	var multiView0 multiview.MultiView
	if chainID == -1 {
		multiView0 = node.GetBlockchain().BeaconChain.CloneMultiView()
	} else {
		multiView0 = node.GetBlockchain().ShardChain[chainID].CloneMultiView()
	}
	foo()
	node.GenerateBlock().NextRound()
	node.NextRound()
	view1 := multiView0.GetBestView()
	node.GenerateBlock().NextRound()
	view2 := multiView0.GetBestView()
	node.NextRound()
	if chainID == -1 {
		node.GetBlockchain().GetChain(chainID).(*blockchain.BeaconChain).SetMultiView(multiView0)
	} else {
		node.GetBlockchain().GetChain(chainID).(*blockchain.ShardChain).SetMultiView(multiView0)
	}

	node.EmptyPool()
	node.GenerateBlock().NextRound()
	node.NextRound()
	foo()
	node.GenerateBlock().NextRound()
	node.NextRound()
	view4 := multiView0.GetBestView()

	if chainID == -1 {
		node.GetBlockchain().GetChain(chainID).(*blockchain.BeaconChain).InsertBlock(view1.GetBlock(), true)
		node.GetBlockchain().GetChain(chainID).(*blockchain.BeaconChain).InsertBlock(view2.GetBlock(), true)
	} else {
		node.GetBlockchain().GetChain(chainID).(*blockchain.ShardChain).InsertBlock(view1.GetBlock(), true)
		node.GetBlockchain().GetChain(chainID).(*blockchain.ShardChain).InsertBlock(view2.GetBlock(), true)
	}
	return view2, view4
}

func (node *NodeEngine) PrintAccountNameFromCPK(committee []incognitokey.CommitteePublicKey) {
	pkMap := make(map[string]string)
	for _, acc := range node.GetAllAccounts() {
		pkMap[acc.SelfCommitteePubkey] = acc.Name
	}
	committeeStr, _ := incognitokey.CommitteeKeyListToString(committee)
	tmp := ""
	for _, pk := range committeeStr {
		if pkMap[pk] != "" {
			tmp += pkMap[pk] + " "
		} else {
			tmp += "@@ "
		}
	}
	fmt.Println(tmp)
}
