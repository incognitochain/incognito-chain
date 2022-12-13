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
	"sort"
	"strconv"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/common/base58"
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

const (
	SHARD_WAIT   = "s_waiting"
	SHARD_PEND   = "s_pending"
	SHARD_SYNC   = "s_syncing"
	SHARD_VALS   = "s_committee"
	SHARD_SLASH  = "s_slashed"
	SHARD_NORMAL = "s_normal"
)

type AccountInfo struct {
	Acc              *account.Account
	Name             string
	Queue            string
	CountInCommittee uint64
	CID              int
}

type StakerInfo struct {
	Name            string
	Delegate        string
	StakingAmount   uint64
	HasCredit       bool
	RewardPRV       uint64
	RewardPRVLocked uint64
	InCommittee     int
	AutoStake       bool
	Balance         uint64
	Chain           string
}

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
					amount, ok := args[i+1].(uint64)
					if !ok {
						amountF64 := args[i+1].(float64)
						amount = uint64(amountF64)
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

func (sim *NodeEngine) ShowAccountsInfo(infos map[string]*AccountInfo) {
	fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", sim.GetBlockchain().BeaconChain.CurrentHeight(), sim.GetBlockchain().BeaconChain.GetEpoch())
	infosByPool := map[string]map[int][]*AccountInfo{}
	poolNames := []string{SHARD_NORMAL, SHARD_WAIT, SHARD_SYNC, SHARD_PEND, SHARD_VALS, SHARD_SLASH}
	for _, name := range poolNames {
		infosByPool[name] = map[int][]*AccountInfo{}
		for _, info := range infos {
			if info.Queue == name {
				infosByPool[name][info.CID] = append(infosByPool[name][info.CID], info)
			}
		}
	}
	for _, name := range poolNames {
		fmt.Printf("List %v:\n", name)
		infoByCIDs := infosByPool[name]
		switch name {
		case SHARD_SYNC, SHARD_PEND, SHARD_VALS, SHARD_SLASH:
			for cID := 0; cID < sim.bc.GetBeaconBestState().ActiveShards; cID++ {
				fmt.Printf("\tcID %v: ", cID)
				for _, v := range infoByCIDs[cID] {
					fmt.Printf(" %v ", v.Name)
				}
				fmt.Println()
			}
		case SHARD_NORMAL, SHARD_WAIT:
			fmt.Printf("\t")
			for _, v := range infoByCIDs[-2] {
				fmt.Printf(" %v ", v.Name)
			}
			fmt.Println()
		}
	}
}

func (sim *NodeEngine) GetAccountPosition(accounts []account.Account, bcView *blockchain.BeaconBestState) map[string]*AccountInfo {
	chain := sim.GetBlockchain()
	pkMap := make(map[string]*AccountInfo)
	for _, acc := range accounts {
		x := acc
		pkMap[acc.SelfCommitteePubkey] = &AccountInfo{&x, acc.Name, SHARD_NORMAL, 0, -2}
	}
	shardWaitingList, _ := incognitokey.CommitteeKeyListToString(chain.BeaconChain.GetShardsWaitingList())
	tmp := ""
	for _, pk := range shardWaitingList {
		if pkMap[pk] != nil && pkMap[pk].Name != "" {
			tmp += pkMap[pk].Name + " "
			pkMap[pk].Queue = SHARD_WAIT
			pkMap[pk].CID = -2
		} else {
			tmp += "@@ "
		}
	}

	shardPendingList := make(map[int][]string)
	shardCommitteeList := make(map[int][]string)
	shardSyncingList := make(map[int][]string)
	shardSlashingList := sim.bc.GetBeaconBestState().GetAllCurrentSlashingCommittee()

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
				pkMap[pk].Queue = SHARD_SYNC
				pkMap[pk].CID = sid
			} else {
				tmp += "@@ "
			}
		}

		tmp = ""
		for _, pk := range shardPendingList[sid] {
			if pkMap[pk] != nil && pkMap[pk].Name != "" {
				tmp += pkMap[pk].Name + " "
				pkMap[pk].Queue = SHARD_PEND
				pkMap[pk].CID = sid
			} else {
				tmp += "@@ "
			}
		}
	}

	for sid := 0; sid < chain.GetActiveShardNumber(); sid++ {
		tmp = ""
		for _, pk := range shardCommitteeList[sid] {
			if pkMap[pk] != nil && pkMap[pk].Name != "" {
				tmp += pkMap[pk].Name + " "
				pkMap[pk].Queue = SHARD_VALS
				pkMap[pk].CID = sid
			} else {
				tmp += "@@ "
			}
		}
	}

	for sid := 0; sid < chain.GetActiveShardNumber(); sid++ {
		tmp = ""
		for _, pk := range shardSlashingList[byte(sid)] {
			if pkMap[pk] != nil && pkMap[pk].Name != "" {
				tmp += pkMap[pk].Name + " "
				pkMap[pk].Queue = SHARD_SLASH
				pkMap[pk].CID = sid
			} else {
				tmp += "@@ "
			}
		}
	}
	return pkMap
}

func (sim *NodeEngine) ShowAccountStakeInfo(accounts []account.Account) {
	chain := sim.GetBlockchain()
	type AccountInfo struct {
		Name            string
		Delegate        string
		HasCredit       bool
		RewardPRV       uint64
		RewardPRVLocked uint64
		InCommittee     int
		AutoStake       bool
		Balance         uint64
	}
	allReward, err := sim.RPC.Client.GetAllRewardAmount()
	if err != nil {
		panic(err)
	}
	pkMap := make(map[string]*AccountInfo)
	for _, acc := range accounts {
		pkMap[acc.SelfCommitteePubkey] = &AccountInfo{acc.Name, "", false, 0, 0, 0, false, 0}
	}
	fmt.Println()
	bBestState := chain.GetBeaconBestState()
	// bC := bBestState.GetBeaconCommittee()
	// bCStr, _ := incognitokey.CommitteeKeyListToString(bC)

	for _, acc := range accounts {
		pkMap[acc.SelfCommitteePubkey].RewardPRV = allReward[acc.PublicKey][common.PRVIDStr]
		stakerInfo, ok, _ := bBestState.GetStakerInfo(acc.SelfCommitteePubkey)
		if ok {
			// delegate := stakerInfo.Delegate()
			// pkMap[acc.SelfCommitteePubkey].InCommittee = stakerInfo.ActiveTimesInCommittee()
			// pkMap[acc.SelfCommitteePubkey].Delegate = delegate
			pkMap[acc.SelfCommitteePubkey].HasCredit = stakerInfo.HasCredit()
			pkMap[acc.SelfCommitteePubkey].AutoStake = stakerInfo.AutoStaking()
			// for idx, bPK := range bCStr {
			// 	if bPK == delegate {
			// 		pkMap[acc.SelfCommitteePubkey].Delegate = fmt.Sprintf("Beacon %+v", idx)
			// 	}
			// }
		}
		stakerInfo2, ok, _ := bBestState.GetBeaconStakerInfo(acc.SelfCommitteePubkey)
		if ok {
			pkMap[acc.SelfCommitteePubkey].AutoStake = stakerInfo2.AutoStaking()
			pkMap[acc.SelfCommitteePubkey].InCommittee = int(stakerInfo2.ActiveTimesInCommittee())
		}

		balanceMap, err := sim.RPC.API_GetBalance(acc)
		if err != nil {
			panic(err)
		}
		pkMap[acc.SelfCommitteePubkey].Balance = balanceMap[common.PRVCoinName]
	}
	for _, acc := range accounts {
		stakerInfo := pkMap[acc.SelfCommitteePubkey]
		fmt.Printf("Acc: %v , Has credit %v, In committee %v, AutoStake %+v Balance %v\n", stakerInfo.Name, stakerInfo.HasCredit, stakerInfo.InCommittee, stakerInfo.AutoStake, stakerInfo.Balance)
	}
}

func (sim *NodeEngine) GetStakerInfo(accounts []account.Account) map[string]*StakerInfo {
	chain := sim.GetBlockchain()
	allReward, err := sim.RPC.Client.GetAllRewardAmount()
	if err != nil {
		panic(err)
	}
	pkMap := make(map[string]*StakerInfo)
	for _, acc := range accounts {
		pkMap[acc.SelfCommitteePubkey] = &StakerInfo{acc.Name, "unknown", 0, false, 0, 0, 0, false, 0, "unknown"}
	}
	bBestState := chain.GetBeaconBestState()

	bC := bBestState.GetBeaconCommittee()
	bCStr, _ := incognitokey.CommitteeKeyListToString(bC)

	for _, acc := range accounts {
		pkMap[acc.SelfCommitteePubkey].RewardPRV = allReward[acc.PublicKey][common.PRVIDStr]
		pkMap[acc.SelfCommitteePubkey].StakingAmount = 0
		stakerInfo, ok, _ := bBestState.GetStakerInfo(acc.SelfCommitteePubkey)
		if ok {
			delegate := stakerInfo.Delegate()
			pkMap[acc.SelfCommitteePubkey].Delegate = delegate
			pkMap[acc.SelfCommitteePubkey].HasCredit = stakerInfo.HasCredit()
			pkMap[acc.SelfCommitteePubkey].StakingAmount = 1750000000000
			pkMap[acc.SelfCommitteePubkey].AutoStake = stakerInfo.AutoStaking()
			for idx, bPK := range bCStr {
				if bPK == delegate {
					pkMap[acc.SelfCommitteePubkey].Delegate = fmt.Sprintf("Beacon %+v", idx)
				}
			}
			pkMap[acc.SelfCommitteePubkey].Chain = "Shard"
		}
		stakerInfo2, ok, _ := bBestState.GetBeaconStakerInfo(acc.SelfCommitteePubkey)
		if ok {
			pkMap[acc.SelfCommitteePubkey].InCommittee = int(stakerInfo2.ActiveTimesInCommittee())
			pkMap[acc.SelfCommitteePubkey].StakingAmount += stakerInfo2.StakingAmount()
		}

		balanceMap, err := sim.RPC.API_GetBalance(acc)
		if err != nil {
			panic(err)
		}
		pkMap[acc.SelfCommitteePubkey].Balance = balanceMap[common.PRVCoinName]
	}
	for _, acc := range accounts {
		stakerInfo := pkMap[acc.SelfCommitteePubkey]
		k := acc.SelfCommitteePubkey
		fmt.Printf("Acc: %v, %v, In committee %v, AutoStake %+v Balance %v, sAmount %v, Has credit %v, Reward %v\n", stakerInfo.Name, k[len(k)-5:], stakerInfo.InCommittee, stakerInfo.AutoStake, stakerInfo.Balance, stakerInfo.StakingAmount, stakerInfo.HasCredit, stakerInfo.RewardPRV)
	}
	return pkMap
}

var rep []uint64
var rew []uint64

func (sim *NodeEngine) ShowBeaconCandidateInfo(accounts, acc []account.Account, epoch uint64) {
	chain := sim.GetBlockchain()
	type CandidateInfo struct {
		Name                     string
		CurrentDelegators        int
		Reputation               uint
		Performance              uint
		CurrentDelegatorsDetails []string
		RewardPRV                uint64
		StakingAmount            uint64
		StakingTxs               []string
	}

	pkStakerMap := make(map[string]string)
	pkCandidateMap := map[string]CandidateInfo{}
	for _, acc := range accounts {
		pkStakerMap[acc.SelfCommitteePubkey] = acc.Name
	}
	bBestState := chain.GetBeaconBestState()
	bCState := bBestState.GetCommitteeState()
	dState := bCState.GetDelegateState()
	keys := []string{}
	allReward, err := sim.RPC.Client.GetAllRewardAmount()
	if err != nil {
		panic(err)
	}
	keys = []string{}
	rep := map[string]uint64{}
	per := map[string]uint64{}
	bcV4, ok := bCState.(*committeestate.BeaconCommitteeStateV4)
	if ok {
		fmt.Println("aaaaaaaaaaaaa")
		rep = bcV4.Reputation
		per = bcV4.Performance
	}
	bcCommitteeStruct := bBestState.GetBeaconCommittee()
	bcWaitingStruct := bBestState.GetCandidateBeaconWaiting()
	bcSubstituteStruct := bBestState.GetBeaconPendingValidator()
	bcList := []string{}
	bcCommitteeString, _ := incognitokey.CommitteeKeyListToString(bcCommitteeStruct)
	bcList = append(bcList, bcCommitteeString...)
	bcWaitingString, _ := incognitokey.CommitteeKeyListToString(bcWaitingStruct)
	bcList = append(bcList, bcWaitingString...)
	bcSubsString, _ := incognitokey.CommitteeKeyListToString(bcSubstituteStruct)
	bcList = append(bcList, bcSubsString...)
	for index, b := range bcCommitteeString {
		bCPk := base58.Base58Check{}.Encode(bcCommitteeStruct[index].GetNormalKey(), 0)
		pkCandidateMap[b] = CandidateInfo{
			Name:                     fmt.Sprintf("Beacon %v", index),
			CurrentDelegators:        0,
			CurrentDelegatorsDetails: []string{},
			Reputation:               uint(rep[b]),
			Performance:              uint(per[b]),
			RewardPRV:                allReward[bCPk][common.PRVIDStr],
		}

		keys = append(keys, b)

		if info, ok := dState[b]; ok {
			detailsList := []string{}
			for _, pk := range info.GetCurrentDelegatorsList() {
				for _, acc := range accounts {
					if pk == acc.SelfCommitteePubkey {
						detailsList = append(detailsList, acc.Name)
						break
					}
				}
			}
			infor, ok := pkCandidateMap[b]
			if !ok {
				infor = CandidateInfo{
					Name:                     fmt.Sprintf("Beacon %v", index),
					CurrentDelegators:        0,
					CurrentDelegatorsDetails: []string{},
					Reputation:               uint(bcV4.Reputation[b]),
					Performance:              uint(bcV4.Performance[b]),
					RewardPRV:                allReward[bCPk][common.PRVIDStr],
				}
			}
			infor.CurrentDelegators = info.CurrentDelegators
			infor.CurrentDelegatorsDetails = detailsList
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		if info2, has, err := statedb.GetBeaconStakerInfo(bBestState.GetBeaconConsensusStateDB(), k); (!has) || (err != nil) {
			fmt.Printf("Can not get beacon staker infor of key %v, err %v\n", k[len(k)-5:], err)
		} else {
			infor := pkCandidateMap[k]
			infor.StakingAmount = info2.StakingAmount()
			for _, v := range info2.TxStakingIDs() {
				infor.StakingTxs = append(infor.StakingTxs, v.String())
			}
			pkCandidateMap[k] = infor
		}
		cInfo := pkCandidateMap[k]
		fmt.Printf("%v\tRep:%v\tPer:%v\tDelegators: %v\tReward %+v\tDetails: %+v\tStakingAmount %+v Staking Txs %+v\n",
			cInfo.Name, cInfo.Reputation, cInfo.Performance, cInfo.CurrentDelegators, cInfo.RewardPRV, cInfo.CurrentDelegatorsDetails, cInfo.StakingAmount, cInfo.StakingTxs)
	}
	for id, value := range bcCommitteeString {
		if len(value) > 5 {
			bcCommitteeString[id] = value[len(value)-5:]
		}
	}
	for id, value := range bcWaitingString {
		if len(value) > 5 {
			bcWaitingString[id] = value[len(value)-5:]
		}
	}
	for id, value := range bcSubsString {
		if len(value) > 5 {
			bcSubsString[id] = value[len(value)-5:]
		}
	}

	fmt.Printf("List beacon waiting: %+v\n", bcWaitingString)
	fmt.Printf("List beacon pending: %+v\n", bcSubsString)
	fmt.Printf("List beacon committee: %+v\n", bcCommitteeString)
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

func (node *NodeEngine) PreparePRVForTest(
	sender account.Account,
	receivers []account.Account,
	amounts []uint64,
) (txIDs []string, errs []error) {
	node.ShowBalance(node.GenesisAccount)
	done := map[int]interface{}{}
	maxTries := 5
	txIDs = make([]string, len(receivers))
	errs = make([]error, len(receivers))
	for {
		for id, acc := range receivers {
			if _, ok := done[id]; ok {
				continue
			}
			node.RPC.API_SubmitKey(acc.PrivateKey)
			txid, err := node.SendPRV(node.GenesisAccount, acc, amounts[id])
			if err == nil {
				done[id] = nil
			}
			txIDs[id] = txid
			errs[id] = err
			node.GenerateBlock().NextRound()
		}
		maxTries--
		if (len(done) == len(receivers)) || (maxTries == 0) {
			break
		}
		fmt.Printf("Retry one more times %v\n", maxTries)
		node.Pause()
	}
	for idx, txID := range txIDs {
		fmt.Printf("Send PRV to account %v, got txID %v, err %v\n", receivers[idx].Name, txID, errs[idx])
	}
	return txIDs, errs
}

func (node *NodeEngine) StakeNewShards(
	stakers []account.Account,
	delegates []string,
	autoStakings []bool,
) (txIDs []string, errs []error) {
	node.ShowBalance(node.GenesisAccount)
	done := map[int]interface{}{}
	maxTries := 5
	txIDs = make([]string, len(stakers))
	errs = make([]error, len(stakers))
	for {
		for id, acc := range stakers {
			if _, ok := done[id]; ok {
				continue
			}
			txid, err := node.RPC.StakeNew(acc, delegates[id], autoStakings[id])
			if err == nil {
				done[id] = nil
				txIDs[id] = txid.TxID
			}
			errs[id] = err
		}
		maxTries--
		node.GenerateBlock().NextRound()
		if (len(done) == len(stakers)) || (maxTries == 0) {
			break
		}
		fmt.Printf("Try one more times %v", maxTries)
		node.Pause()
	}
	for idx, txID := range txIDs {
		fmt.Printf("stake for account %v, got txID %v, err %v\n", stakers[idx].Name, txID, errs[idx])
	}
	return txIDs, errs
}

func (node *NodeEngine) StakeNewBeacons(
	stakers []account.Account,
) (txIDs []string, errs []error) {
	node.ShowBalance(node.GenesisAccount)
	done := map[int]interface{}{}
	maxTries := 5
	txIDs = make([]string, len(stakers))
	errs = make([]error, len(stakers))
	for {
		for id, acc := range stakers {
			if _, ok := done[id]; ok {
				continue
			}
			txid, err := node.RPC.StakeNewBeacon(acc)
			if err == nil {
				done[id] = nil
				txIDs[id] = txid.TxID
			}
			errs[id] = err
		}
		maxTries--
		node.GenerateBlock().NextRound()
		if (len(done) == len(stakers)) || (maxTries == 0) {
			break
		}
	}
	for idx, txID := range txIDs {
		fmt.Printf("stake beacon for account %v, got txID %v, err %v\n", stakers[idx].Name, txID, errs[idx])
	}
	return txIDs, errs
}
