package blockchain

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	eCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/incognitochain/incognito-chain/blockchain/portalContract/delegator"
	"github.com/incognitochain/incognito-chain/blockchain/portalContract/erc20"
	incognitoproxy "github.com/incognitochain/incognito-chain/blockchain/portalContract/incognitoProxy"
	"github.com/incognitochain/incognito-chain/blockchain/portalContract/portalv3"
	"github.com/incognitochain/incognito-chain/common"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type PortalV3TestSuite struct {
	suite.Suite
	currentPortalStateForProducer CurrentPortalState
	currentPortalStateForProcess  CurrentPortalState
	sdb                           *statedb.StateDB
	portalParams                  PortalParams
	blockChain                    *BlockChain
	p                             *Platform
}

type TestCaseCustodianDepositV3 struct {
	remoteAddress       map[string]string
	depositAmount       uint64
	blockHash           eCommon.Hash
	txIndex             uint
	proofStrs           []string
	custodianIncAddress string
}

type committees struct {
	beacons     []eCommon.Address
	beaconPrivs [][]byte
}

type contracts struct {
	delegatorAddr eCommon.Address
	portalV3Ins   *portalv3.Portalv3
	portalv3      eCommon.Address
	inc           *incognitoproxy.Incognitoproxy
	incAddr       eCommon.Address
	token         *erc20.Erc20
	tokenAddr     eCommon.Address
}

type Platform struct {
	*contracts
	sim        *backends.SimulatedBackend
	auth       *bind.TransactOpts
	genesisAcc *account
}

type account struct {
	PrivateKey *ecdsa.PrivateKey
	Address    eCommon.Address
}

type Receipt struct {
	Result *types.Receipt `json:"result"`
}

func (s *PortalV3TestSuite) SetupTest() {

	fmt.Println("Initializing genesis account...")
	s.p.genesisAcc = loadAccount()
	s.p.auth = bind.NewKeyedTransactor(s.p.genesisAcc.PrivateKey)
	c := getFixedCommittee()
	err := setup(c.beacons, s.p)
	if err != nil {
		panic(err)
	}

	dbPath, err := ioutil.TempDir(os.TempDir(), "portal_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	stateDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)

	s.sdb = stateDB

	finalExchangeRate := statedb.NewFinalExchangeRatesStateWithValue(
		map[string]statedb.FinalExchangeRatesDetail{
			common.PRVIDStr:       {Amount: 1000000},
			common.PortalBNBIDStr: {Amount: 20000000},
			common.PortalBTCIDStr: {Amount: 10000000000},
		})
	s.currentPortalStateForProducer = CurrentPortalState{
		CustodianPoolState:         map[string]*statedb.CustodianState{},
		WaitingPortingRequests:     map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:      map[string]*statedb.RedeemRequest{},
		MatchedRedeemRequests:      map[string]*statedb.RedeemRequest{},
		FinalExchangeRatesState:    finalExchangeRate,
		LiquidationPool:            map[string]*statedb.LiquidationPool{},
		LockedCollateralForRewards: new(statedb.LockedCollateralState),
		ExchangeRatesRequests:      map[string]*metadata.ExchangeRatesRequestStatus{},
	}
	s.currentPortalStateForProcess = CurrentPortalState{
		CustodianPoolState:         map[string]*statedb.CustodianState{},
		WaitingPortingRequests:     map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:      map[string]*statedb.RedeemRequest{},
		MatchedRedeemRequests:      map[string]*statedb.RedeemRequest{},
		FinalExchangeRatesState:    finalExchangeRate,
		LiquidationPool:            map[string]*statedb.LiquidationPool{},
		LockedCollateralForRewards: new(statedb.LockedCollateralState),
		ExchangeRatesRequests:      map[string]*metadata.ExchangeRatesRequestStatus{},
	}
	s.portalParams = PortalParams{
		TimeOutCustodianReturnPubToken:       24 * time.Hour,
		TimeOutWaitingPortingRequest:         24 * time.Hour,
		TimeOutWaitingRedeemRequest:          15 * time.Minute,
		MaxPercentLiquidatedCollateralAmount: 120,
		MaxPercentCustodianRewards:           10,
		MinPercentCustodianRewards:           1,
		MinLockCollateralAmountInEpoch:       5000 * 1e9, // 5000 prv
		MinPercentLockedCollateral:           200,
		TP120:                                120,
		TP130:                                130,
		MinPercentPortingFee:                 0.01,
		MinPercentRedeemFee:                  0.01,
	}
	s.blockChain = &BlockChain{
		config: Config{
			ChainParams: &Params{
				MinBeaconBlockInterval:      40 * time.Second,
				Epoch:                       100,
				PortalETHContractAddressStr: s.p.delegatorAddr.String(),
			},
		},
	}
}

func (s *PortalV3TestSuite) TestCustodianDepositCollateralV3() {
	fmt.Println("Running TestCustodianDepositCollateralV3 - beacon height 1000 ...")
	bc := s.blockChain
	pm := NewPortalManager()
	beaconHeight := uint64(1000)
	shardID := byte(0)
	s.p.sim.Commit()
	//newMatchedRedeemReqIDs := []string{}
	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}

	// build test cases
	testcases := []TestCaseCustodianDepositV3{
		// valid
		{
			custodianIncAddress: "custodianIncAddress1",
			remoteAddress: map[string]string{
				common.PortalBNBIDStr: "bnbAddress1",
				common.PortalBTCIDStr: "btcAddress1",
			},
			depositAmount: 5 * 1e18,
		},
		// custodian deposit more with new remote addresses
		// expect don't change to new remote addresses,
		// custodian is able to update new remote addresses when total collaterals is empty
		{
			custodianIncAddress: "custodianIncAddress1",
			remoteAddress: map[string]string{
				common.PortalBNBIDStr: "bnbAddress2",
				common.PortalBTCIDStr: "btcAddress2",
			},
			depositAmount: 2 * 1e18,
		},
		// new custodian supply only bnb address
		{
			custodianIncAddress: "custodianIncAddress2",
			remoteAddress: map[string]string{
				common.PortalBNBIDStr: "bnbAddress2",
			},
			depositAmount: 1 * 1e18,
		},
		// new custodian supply only btc address
		{
			custodianIncAddress: "custodianIncAddress3",
			remoteAddress: map[string]string{
				common.PortalBTCIDStr: "btcAddress3",
			},
			depositAmount: 10 * 1e18,
		},
	}

	for _, testcase := range testcases {
		tx, err := s.p.portalV3Ins.Deposit(s.p.auth, testcase.custodianIncAddress)
		s.p.sim.Commit()
		s.Equal(nil, err)
		_, blockHash, txIndex, proofStrs, err := getETHDepositProof(tx.Hash(), s.p.sim)
		s.Equal(nil, err)
		testcase.blockHash = blockHash
		testcase.txIndex = txIndex
		testcase.proofStrs = proofStrs
	}

	// build actions from testcases
	insts := buildCustodianDepositActionsFromTcsV3(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight, insts, s.sdb, &s.currentPortalStateForProducer, s.portalParams, shardID, pm)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, updatingInfoByTokenID)

	// check results
	s.Equal(4, len(newInsts))
	s.Equal(nil, err)

	custodianKey1 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress1").String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress2").String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress3").String()

	custodian1 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress1", 7000*1e9, 7000*1e9,
		map[string]uint64{}, map[string]uint64{},
		map[string]string{
			common.PortalBNBIDStr: "bnbAddress1",
			common.PortalBTCIDStr: "btcAddress1",
		}, map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})

	custodian2 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress2", 1000*1e9, 1000*1e9,
		map[string]uint64{}, map[string]uint64{},
		map[string]string{
			common.PortalBNBIDStr: "bnbAddress2",
		}, map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})

	custodian3 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress3", 10000*1e9, 10000*1e9,
		map[string]uint64{}, map[string]uint64{},
		map[string]string{
			common.PortalBTCIDStr: "btcAddress3",
		}, map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})

	s.Equal(3, len(s.currentPortalStateForProducer.CustodianPoolState))
	s.Equal(custodian1, s.currentPortalStateForProducer.CustodianPoolState[custodianKey1])
	s.Equal(custodian2, s.currentPortalStateForProducer.CustodianPoolState[custodianKey2])
	s.Equal(custodian3, s.currentPortalStateForProducer.CustodianPoolState[custodianKey3])

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

func buildCustodianDepositActionsFromTcsV3(tcs []TestCaseCustodianDepositV3, shardID byte) [][]string {
	insts := [][]string{}

	for _, tc := range tcs {
		inst := buildPortalCustodianDepositActionV3(tc.remoteAddress, tc.blockHash, tc.txIndex, tc.proofStrs, shardID)
		insts = append(insts, inst)
	}

	return insts
}

func buildPortalCustodianDepositActionV3(
	remoteAddress map[string]string,
	blockHash eCommon.Hash,
	txIndex uint,
	proofStrs []string,
	shardID byte,
) []string {
	data := metadata.PortalCustodianDepositV3{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalCustodianDepositMetaV3,
		},
		RemoteAddresses: remoteAddress,
		BlockHash:       blockHash,
		TxIndex:         txIndex,
		ProofStrs:       proofStrs,
	}

	actionContent := metadata.PortalCustodianDepositActionV3{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalCustodianDepositMetaV3), actionContentBase64Str}
}

func getFixedCommittee() *committees {
	beaconCommPrivs := []string{
		"aad53b70ad9ed01b75238533dd6b395f4d300427da0165aafbd42ea7a606601f",
		"ca71365ceddfa8e0813cf184463bd48f0b62c9d7d5825cf95263847628816e82",
		"1e4d2244506211200640567630e3951abadbc2154cf772e4f0d2ff0770290c7c",
		"c7146b500240ed7aac9445e2532ae8bf6fc7108f6ea89fde5eebdf2fb6cefa5a",
	}
	beaconComm := []string{
		"0xD7d93b7fa42b60b6076f3017fCA99b69257A912D",
		"0xf25ee30cfed2d2768C51A6Eb6787890C1c364cA4",
		"0x0D8c517557f3edE116988DD7EC0bAF83b96fe0Cb",
		"0xc225fcd5CE8Ad42863182Ab71acb6abD9C4ddCbE",
	}
	beaconPrivs := make([][]byte, len(beaconCommPrivs))
	for i, p := range beaconCommPrivs {
		priv, _ := hex.DecodeString(p)
		beaconPrivs[i] = priv
	}

	beacons := toAddresses(beaconComm)
	return &committees{
		beacons:     beacons,
		beaconPrivs: beaconPrivs,
	}
}

func toAddresses(beaconComm []string) []eCommon.Address {
	beacons := make([]eCommon.Address, len(beaconComm))
	for i, p := range beaconComm {
		beacons[i] = eCommon.HexToAddress(p)
	}

	return beacons
}

func loadAccount() *account {
	key, err := crypto.LoadECDSA("genesisKey.hex")
	if err != nil {
		return newAccount()
	}
	return &account{
		PrivateKey: key,
		Address:    crypto.PubkeyToAddress(key.PublicKey),
	}
}

func newAccount() *account {
	key, _ := crypto.GenerateKey()
	return &account{
		PrivateKey: key,
		Address:    crypto.PubkeyToAddress(key.PublicKey),
	}
}

func setup(
	beaconComm []eCommon.Address,
	p *Platform,
	accs ...eCommon.Address,
) error {
	alloc := make(core.GenesisAlloc)
	balance, _ := big.NewInt(1).SetString("1000000000000000000000000000000", 10) // 1E30 wei
	alloc[p.auth.From] = core.GenesisAccount{Balance: balance}
	for _, acc := range accs {
		alloc[acc] = core.GenesisAccount{Balance: balance}
	}
	p.sim = backends.NewSimulatedBackend(alloc, 8000000)
	var err error
	var tx *types.Transaction
	_ = tx

	// ERC20: always deploy first so its address is fixed
	p.tokenAddr, tx, p.token, err = erc20.DeployErc20(p.auth, p.sim, "MyErc20", "ERC", big.NewInt(8), big.NewInt(int64(1e18)))
	if err != nil {
		return fmt.Errorf("failed to deploy ERC20 contract: %v", err)
	}
	// fmt.Printf("token addr: %s\n", p.tokenAddr.Hex())
	p.sim.Commit()
	// IncognitoProxy
	admin := p.auth.From
	p.incAddr, tx, p.inc, err = incognitoproxy.DeployIncognitoproxy(p.auth, p.sim, admin, beaconComm)
	if err != nil {
		return fmt.Errorf("failed to deploy IncognitoProxy contract: %v", err)
	}
	p.sim.Commit()

	p.portalv3, tx, _, err = portalv3.DeployPortalv3(p.auth, p.sim)
	if err != nil {
		return fmt.Errorf("failed to deploy Portal contract: %v", err)
	}
	p.sim.Commit()

	// Portal
	p.delegatorAddr, _, _, err = delegator.DeployDelegator(p.auth, p.sim, p.auth.From, p.portalv3, p.incAddr)
	if err != nil {
		return err
	}
	p.sim.Commit()

	p.portalV3Ins, err = portalv3.NewPortalv3(p.delegatorAddr, p.sim)
	if err != nil {
		return fmt.Errorf("failed to assgin portal contract to delegator address: %v", err)
	}

	return nil
}

func getETHDepositProof(
	txHash eCommon.Hash,
	sim *backends.SimulatedBackend,
) (*big.Int, eCommon.Hash, uint, []string, error) {
	// Get tx content
	txContent, err := sim.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		fmt.Println("Cannot get transaction by hash : ", err)
		return nil, eCommon.Hash{}, 0, nil, err
	}

	blockHeader := sim.Blockchain().GetBlockByNumber(txContent.BlockNumber.Uint64())
	if blockHeader == nil {
		fmt.Println("Cannot get block by height")
		return nil, eCommon.Hash{}, 0, nil, err
	}

	// Constructing the receipt trie (source: go-ethereum/core/types/derive_sha.go)
	keybuf := new(bytes.Buffer)
	receiptTrie := new(trie.Trie)
	for i, tx := range blockHeader.Transactions() {
		siblingReceipt, err := sim.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			return nil, eCommon.Hash{}, 0, nil, err
		}
		keybuf.Reset()
		rlp.Encode(keybuf, uint(i))
		encodedReceipt, err := rlp.EncodeToBytes(siblingReceipt)
		if err != nil {
			return nil, eCommon.Hash{}, 0, nil, err
		}
		receiptTrie.Update(keybuf.Bytes(), encodedReceipt)
	}

	// Constructing the proof for the current receipt (source: go-ethereum/trie/proof.go)
	proof := light.NewNodeSet()
	keybuf.Reset()
	rlp.Encode(keybuf, txContent.TransactionIndex)
	err = receiptTrie.Prove(keybuf.Bytes(), 0, proof)
	if err != nil {
		return nil, eCommon.Hash{}, 0, nil, err
	}

	nodeList := proof.NodeList()
	encNodeList := make([]string, 0)
	for _, node := range nodeList {
		str := base64.StdEncoding.EncodeToString(node)
		encNodeList = append(encNodeList, str)
	}

	return txContent.BlockNumber, txContent.BlockHash, txContent.TransactionIndex, encNodeList, nil
}

func TestPortalSuiteV3(t *testing.T) {
	suite.Run(t, new(PortalV3TestSuite))
}
