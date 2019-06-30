package bridge

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/bridge/incognito_proxy"
	"github.com/incognitochain/incognito-chain/bridge/vault"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/ethrelaying/accounts/abi/bind"
	"github.com/incognitochain/incognito-chain/ethrelaying/accounts/abi/bind/backends"
	"github.com/incognitochain/incognito-chain/ethrelaying/common"
	"github.com/incognitochain/incognito-chain/ethrelaying/core"
	"github.com/incognitochain/incognito-chain/ethrelaying/core/types"
	"github.com/incognitochain/incognito-chain/ethrelaying/crypto"
)

var genesisKey *ecdsa.PrivateKey
var auth *bind.TransactOpts

type Platform struct {
	inc       *incognito_proxy.IncognitoProxy
	incAddr   common.Address
	vault     *vault.Vault
	vaultAddr common.Address
	sim       *backends.SimulatedBackend
}

func (p *Platform) getBalance(addr common.Address) *big.Int {
	bal, _ := p.sim.BalanceAt(context.Background(), addr, nil)
	return bal
}

func init() {
	fmt.Println("Initializing...")
	genesisKey, _ = crypto.GenerateKey()
	auth = bind.NewKeyedTransactor(genesisKey)
}

func setup(beaconCommRoot, bridgeCommRoot [32]byte) (*Platform, error) {
	alloc := make(core.GenesisAlloc)
	balance := big.NewInt(123000000000000000)
	alloc[auth.From] = core.GenesisAccount{Balance: balance}
	sim := backends.NewSimulatedBackend(alloc, 10000000)
	p := &Platform{sim: sim}

	incognitoAddr, tx, inc, err := incognito_proxy.DeployIncognitoProxy(auth, sim, beaconCommRoot, bridgeCommRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy IncognitoProxy contract: %v", err)
	}
	sim.Commit()

	p.printReceipt(tx)
	p.inc = inc
	p.incAddr = incognitoAddr
	fmt.Printf("deployed bridge, addr: %x\n", incognitoAddr)

	vaultAddr, tx, vault, err := vault.DeployVault(auth, sim, incognitoAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Vault contract: %v", err)
	}
	sim.Commit()

	p.printReceipt(tx)
	p.vault = vault
	p.vaultAddr = vaultAddr
	fmt.Printf("deployed bridge, addr: %x\n", vaultAddr)
	return p, nil
}

func setupWithCommittee() (*Platform, error) {
	url := "http://127.0.0.1:9334"

	payload := strings.NewReader("{\n    \"id\": 1,\n    \"jsonrpc\": \"1.0\",\n    \"method\": \"getbeaconbeststate\",\n    \"params\": []\n}")

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Host", "127.0.0.1:9334")
	req.Header.Add("accept-encoding", "gzip, deflate")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("cache-control", "no-cache")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	type beaconBestStateResult struct {
		BeaconCommittee []string
		ShardCommittee  map[string][]string
	}

	type getBeaconBestStateResult struct {
		Result beaconBestStateResult
		Error  string
		Id     int
	}

	r := getBeaconBestStateResult{}
	err = json.Unmarshal([]byte(body), &r)
	if err != nil {
		return nil, err
	}

	// Genesis committee
	beaconOldFlat := [][]byte{}
	for i, val := range r.Result.BeaconCommittee {
		pk, _, _ := base58.Base58Check{}.Decode(val)
		fmt.Printf("pk[%d]: %x %d\n", i, pk, len(pk))
		fmt.Printf("hash(pk[%d]): %x\n", i, keccak256(pk))
		beaconOldFlat = append(beaconOldFlat, pk)
	}
	beaconOldRoot := toByte32(blockchain.GetKeccak256MerkleRoot(beaconOldFlat))
	fmt.Printf("beaconOldRoot: %x\n", beaconOldRoot[:])

	bridgeOldFlat := [][]byte{}
	for _, val := range r.Result.ShardCommittee["1"] {
		pk, _, _ := base58.Base58Check{}.Decode(val)
		bridgeOldFlat = append(bridgeOldFlat, pk)
	}
	bridgeOldRoot := toByte32(blockchain.GetKeccak256MerkleRoot(bridgeOldFlat))
	fmt.Printf("bridgeOldRoot: %x\n", bridgeOldRoot[:])

	return setup(beaconOldRoot, bridgeOldRoot)
}

type account struct {
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
}

func newAccount() *account {
	key, _ := crypto.GenerateKey()
	return &account{
		PrivateKey: key,
		Address:    crypto.PubkeyToAddress(key.PublicKey),
	}
}

func (p *Platform) printReceipt(tx *types.Transaction) {
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	receipt, err := p.sim.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		fmt.Println("receipt err:", err)
	}
	fmt.Printf("tx gas used: %v\n", receipt.CumulativeGasUsed)

	if len(receipt.Logs) == 0 {
		fmt.Println("empty log")
		return
	}

	for i, log := range receipt.Logs {
		var data interface{}
		data = log.Data

		format := "%+v"
		switch log.Topics[0].Hex() {
		case "0x8b1126c8e4087477c3efd9e3785935b29c778491c70e249de774345f7ca9b7f9":
			format = "%s"
		case "0xb42152598f9b870207037767fd41b627a327c9434c796b2ee501d68acec68d1b":
			format = "%x"
		case "0x6c8f06ff564112a969115be5f33d4a0f87ba918c9c9bc3090fe631968e818be4":
			format = "%t"
			data = log.Data[len(log.Data)-1] > 0
		case "0x8e2fc7b10a4f77a18c553db9a8f8c24d9e379da2557cb61ad4cc513a2f992cbd":
			format = "%s"
			data = big.NewInt(int64(0)).SetBytes(log.Data)
		case "0x0ac6e167e94338a282ec23bdd86f338fc787bd67f48b3ade098144aac3fcd86e":
			format = "%x"
			data = log.Data[12:]
		}

		fmt.Printf(fmt.Sprintf("logs[%%d]: %s\n", format), i, data)
		// for _, topic := range log.Topics {
		// 	fmt.Printf("topic: %x\n", topic)
		// }
	}
}

func getBridgeSwapProof() string {
	url := "http://127.0.0.1:9338"

	block := 32
	payload := strings.NewReader(fmt.Sprintf("{\n    \"id\": 1,\n    \"jsonrpc\": \"1.0\",\n    \"method\": \"getbridgeswapproof\",\n    \"params\": [\n    \t%d\n    ]\n}", block))

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Host", "127.0.0.1:9338")
	req.Header.Add("accept-encoding", "gzip, deflate")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("cache-control", "no-cache")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("err:", err)
		return ""
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	//fmt.Println(string(body))
	return string(body)
}

func getBeaconSwapProof() string {
	url := "http://127.0.0.1:9338"

	block := 51
	payload := strings.NewReader(fmt.Sprintf("{\n    \"id\": 1,\n    \"jsonrpc\": \"1.0\",\n    \"method\": \"getbeaconswapproof\",\n    \"params\": [\n    \t%d\n    ]\n}", block))

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Host", "127.0.0.1:9338")
	req.Header.Add("accept-encoding", "gzip, deflate")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("cache-control", "no-cache")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("err:", err)
		return ""
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	//fmt.Println(string(body))
	return string(body)
}

func TestSimulatedSwapBeacon(t *testing.T) {
	// body := getBeaconSwapProof()
	body := getBridgeSwapProof()
	if len(body) < 1 {
		return
	}

	r := getProofResult{}
	err := json.Unmarshal([]byte(body), &r)
	if err != nil {
		t.Error(err)
	}
	proof := decodeProof(&r)
	_ = proof

	p, err := setupWithCommittee()
	if err != nil {
		t.Fatal(err)
	}
	_ = p

	auth.GasLimit = 6000000
	tx, err := p.inc.SwapCommittee(
		auth,
		proof.instruction,

		proof.beaconInstPath,
		proof.beaconInstPathIsLeft,
		proof.beaconInstPathLen,
		proof.beaconInstRoot,
		proof.beaconBlkData,
		proof.beaconBlkHash,
		proof.beaconSignerPubkeys,
		proof.beaconSignerCount,
		proof.beaconSignerSig,
		proof.beaconSignerPaths,
		proof.beaconSignerPathIsLeft,
		proof.beaconSignerPathLen,

		proof.bridgeInstPath,
		proof.bridgeInstPathIsLeft,
		proof.bridgeInstPathLen,
		proof.bridgeInstRoot,
		proof.bridgeBlkData,
		proof.bridgeBlkHash,
		proof.bridgeSignerPubkeys,
		proof.bridgeSignerCount,
		proof.bridgeSignerSig,
		proof.bridgeSignerPaths,
		proof.bridgeSignerPathIsLeft,
		proof.bridgeSignerPathLen,
	)
	if err != nil {
		fmt.Println("err:", err)
	}
	p.sim.Commit()
	p.printReceipt(tx)
}

func TestSimulatedBurn(t *testing.T) {
	proof, err := getAndDecodeBurnProof("")
	if err != nil {
		t.Fatal(err)
	}

	p, err := setupWithCommittee()
	if err != nil {
		t.Fatalf("Fail to deloy contract: %v\n", err)
	}

	oldBalance, newBalance, err := deposit(p, int64(500000000000))
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("deposit to vault: %d -> %d\n", oldBalance, newBalance)

	// tokenID := proof.instruction[3:35]
	// to := proof.instruction[35:67]
	// amount := big.NewInt(0).SetBytes(proof.instruction[67:99])
	// fmt.Printf("tokenID: %x\n", tokenID)
	// fmt.Printf("to: %x\n", to)
	// fmt.Printf("amount: %s\n", amount.String())

	withdrawer := common.HexToAddress("0x0FFBd68F130809BcA7b32D9536c8339E9A844620")
	fmt.Printf("withdrawer init balance: %d\n", p.getBalance(withdrawer))

	auth.GasLimit = 6000000
	tx, err := p.vault.Withdraw(
		auth,
		proof.instruction,

		proof.beaconInstPath,
		proof.beaconInstPathIsLeft,
		proof.beaconInstPathLen,
		proof.beaconInstRoot,
		proof.beaconBlkData,
		proof.beaconBlkHash,
		proof.beaconSignerPubkeys,
		proof.beaconSignerCount,
		proof.beaconSignerSig,
		proof.beaconSignerPaths,
		proof.beaconSignerPathIsLeft,
		proof.beaconSignerPathLen,

		proof.bridgeInstPath,
		proof.bridgeInstPathIsLeft,
		proof.bridgeInstPathLen,
		proof.bridgeInstRoot,
		proof.bridgeBlkData,
		proof.bridgeBlkHash,
		proof.bridgeSignerPubkeys,
		proof.bridgeSignerCount,
		proof.bridgeSignerSig,
		proof.bridgeSignerPaths,
		proof.bridgeSignerPathIsLeft,
		proof.bridgeSignerPathLen,
	)
	if err != nil {
		fmt.Println("err:", err)
	}
	p.sim.Commit()
	p.printReceipt(tx)

	fmt.Printf("withdrawer new balance: %d\n", p.getBalance(withdrawer))
}

func deposit(p *Platform, amount int64) (*big.Int, *big.Int, error) {
	initBalance := p.getBalance(p.vaultAddr)
	auth := bind.NewKeyedTransactor(genesisKey)
	auth.Value = big.NewInt(amount)
	_, err := p.vault.Deposit(auth, "")
	if err != nil {
		return nil, nil, err
	}
	p.sim.Commit()
	newBalance := p.getBalance(p.vaultAddr)
	return initBalance, newBalance, nil
}

func TestSimulatedExtract(t *testing.T) {
	p, err := setupWithCommittee()
	if err != nil {
		t.Fatalf("Fail to deloy contract: %v\n", err)
	}

	a, _ := hex.DecodeString("000000000000000000000000e722D8b71DCC0152D47D2438556a45D3357d631f")
	addr, wei, err := p.vault.TestExtract(nil, a)
	fmt.Printf("addr: %x\n", addr)
	fmt.Printf("wei: %d\n", wei)
}

func TestSimulatedCallFunc(t *testing.T) {
	p, err := setupWithCommittee()
	if err != nil {
		t.Fatalf("Fail to deloy contract: %v\n", err)
	}

	v := [32]byte{}
	b := big.NewInt(135790246810123).Bytes()
	copy(v[32-len(b):], b)
	tx, err := p.inc.NotifyPls(auth, v)
	if err != nil {
		fmt.Println("err:", err)
	}
	p.sim.Commit()
	p.printReceipt(tx)
}
