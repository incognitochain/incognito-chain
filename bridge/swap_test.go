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

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/bridge/incognito_proxy"
	"github.com/incognitochain/incognito-chain/bridge/vault"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

// Test calling
const max_path = 4
const comm_size = 1 << max_path
const pubkey_length = comm_size * max_path
const inst_size = 1 << max_path
const inst_length = 150

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

func keccak256(b ...[]byte) [32]byte {
	h := crypto.Keccak256(b...)
	r := [32]byte{}
	copy(r[:], h)
	return r
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

type decodedProof struct {
	instruction []byte

	beaconInstPath         [max_path][32]byte
	beaconInstPathIsLeft   [max_path]bool
	beaconInstPathLen      *big.Int
	beaconInstRoot         [32]byte
	beaconBlkData          [32]byte
	beaconBlkHash          [32]byte
	beaconSignerPubkeys    []byte
	beaconSignerCount      *big.Int
	beaconSignerSig        [32]byte
	beaconSignerPaths      [pubkey_length][32]byte
	beaconSignerPathIsLeft [pubkey_length]bool
	beaconSignerPathLen    *big.Int

	bridgeInstPath         [max_path][32]byte
	bridgeInstPathIsLeft   [max_path]bool
	bridgeInstPathLen      *big.Int
	bridgeInstRoot         [32]byte
	bridgeBlkData          [32]byte
	bridgeBlkHash          [32]byte
	bridgeSignerPubkeys    []byte
	bridgeSignerCount      *big.Int
	bridgeSignerSig        [32]byte
	bridgeSignerPaths      [pubkey_length][32]byte
	bridgeSignerPathIsLeft [pubkey_length]bool
	bridgeSignerPathLen    *big.Int
}

type getProofResult struct {
	Result jsonresult.GetInstructionProof
	Error  string
	Id     int
}

func decodeProof(r *getProofResult) *decodedProof {
	inst := decode(r.Result.Instruction)
	fmt.Printf("inst: %d %x\n", len(inst), inst)

	beaconInstRoot := decode32(r.Result.BeaconInstRoot)
	beaconInstPath := [max_path][32]byte{}
	beaconInstPathIsLeft := [max_path]bool{}
	for i, path := range r.Result.BeaconInstPath {
		beaconInstPath[i] = decode32(path)
		beaconInstPathIsLeft[i] = r.Result.BeaconInstPathIsLeft[i]
	}
	beaconInstPathLen := big.NewInt(int64(len(r.Result.BeaconInstPath)))
	fmt.Printf("beaconInstRoot: %x\n", beaconInstRoot)

	beaconBlkData := toByte32(decode(r.Result.BeaconBlkData))
	beaconBlkHash := toByte32(decode(r.Result.BeaconBlkHash))
	fmt.Printf("expected beaconBlkHash: %x\n", keccak256(beaconBlkData[:], beaconInstRoot[:]))
	fmt.Printf("beaconBlkHash: %x\n\n", beaconBlkHash)

	beaconSignerPubkeys := []byte{}
	for _, signer := range r.Result.BeaconSignerPubkeys {
		beaconSignerPubkeys = append(beaconSignerPubkeys, decode(signer)...)
	}
	beaconSignerCount := big.NewInt(int64(len(r.Result.BeaconSignerPubkeys)))

	beaconSignerSig := toByte32(decode(r.Result.BeaconSignerSig))
	beaconSignerPaths := [pubkey_length][32]byte{}
	beaconSignerPathIsLeft := [pubkey_length]bool{}
	for i, fullPath := range r.Result.BeaconSignerPaths {
		for j, node := range fullPath {
			k := i*len(fullPath) + j
			beaconSignerPaths[k] = decode32(node)
			beaconSignerPathIsLeft[k] = r.Result.BeaconSignerPathIsLeft[i][j]
		}
	}
	beaconSignerPathLen := big.NewInt(int64(len(r.Result.BeaconSignerPaths[0])))

	// For bridge
	bridgeInstRoot := decode32(r.Result.BridgeInstRoot)
	bridgeInstPath := [max_path][32]byte{}
	bridgeInstPathIsLeft := [max_path]bool{}
	for i, path := range r.Result.BridgeInstPath {
		bridgeInstPath[i] = decode32(path)
		bridgeInstPathIsLeft[i] = r.Result.BridgeInstPathIsLeft[i]
	}
	bridgeInstPathLen := big.NewInt(int64(len(r.Result.BridgeInstPath)))
	// fmt.Printf("bridgeInstRoot: %x\n", bridgeInstRoot)

	bridgeBlkData := toByte32(decode(r.Result.BridgeBlkData))
	bridgeBlkHash := toByte32(decode(r.Result.BridgeBlkHash))
	// fmt.Printf("bridgeBlkHash: %x\n", bridgeBlkHash)

	bridgeSignerPubkeys := []byte{}
	for _, signer := range r.Result.BridgeSignerPubkeys {
		bridgeSignerPubkeys = append(bridgeSignerPubkeys, decode(signer)...)
	}
	bridgeSignerCount := big.NewInt(int64(len(r.Result.BridgeSignerPubkeys)))

	bridgeSignerSig := toByte32(decode(r.Result.BridgeSignerSig))
	bridgeSignerPaths := [pubkey_length][32]byte{}
	bridgeSignerPathIsLeft := [pubkey_length]bool{}
	for i, fullPath := range r.Result.BridgeSignerPaths {
		for j, node := range fullPath {
			k := i*len(fullPath) + j
			bridgeSignerPaths[k] = decode32(node)
			bridgeSignerPathIsLeft[k] = r.Result.BridgeSignerPathIsLeft[i][j]
		}
	}
	bridgeSignerPathLen := big.NewInt(int64(len(r.Result.BridgeSignerPaths[0])))
	return &decodedProof{
		instruction: inst,

		beaconInstPath:         beaconInstPath,
		beaconInstPathIsLeft:   beaconInstPathIsLeft,
		beaconInstPathLen:      beaconInstPathLen,
		beaconInstRoot:         beaconInstRoot,
		beaconBlkData:          beaconBlkData,
		beaconBlkHash:          beaconBlkHash,
		beaconSignerPubkeys:    beaconSignerPubkeys,
		beaconSignerCount:      beaconSignerCount,
		beaconSignerSig:        beaconSignerSig,
		beaconSignerPaths:      beaconSignerPaths,
		beaconSignerPathIsLeft: beaconSignerPathIsLeft,
		beaconSignerPathLen:    beaconSignerPathLen,

		bridgeInstPath:         bridgeInstPath,
		bridgeInstPathIsLeft:   bridgeInstPathIsLeft,
		bridgeInstPathLen:      bridgeInstPathLen,
		bridgeInstRoot:         bridgeInstRoot,
		bridgeBlkData:          bridgeBlkData,
		bridgeBlkHash:          bridgeBlkHash,
		bridgeSignerPubkeys:    bridgeSignerPubkeys,
		bridgeSignerCount:      bridgeSignerCount,
		bridgeSignerSig:        bridgeSignerSig,
		bridgeSignerPaths:      bridgeSignerPaths,
		bridgeSignerPathIsLeft: bridgeSignerPathIsLeft,
		bridgeSignerPathLen:    bridgeSignerPathLen,
	}
}

func TestSwapBeacon(t *testing.T) {
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

func TestBurn(t *testing.T) {
	body := getBurnProof()
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
		t.Fatalf("Fail to deloy contract: %v\n", err)
	}

	oldBalance, newBalance, err := deposit(p, 100000)
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

	withdrawer := common.HexToAddress("0xe722D8b71DCC0152D47D2438556a45D3357d631f")
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

func getBurnProof() string {
	url := "http://127.0.0.1:9338"

	txID := "af58af3ea4cf3df3b6b3108663647c70529b7cb1b65023ec597f80ce1f8bee70"
	payload := strings.NewReader(fmt.Sprintf("{\n    \"id\": 1,\n    \"jsonrpc\": \"1.0\",\n    \"method\": \"getburnproof\",\n    \"params\": [\n    \t\"%s\"\n    ]\n}", txID))

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

func TestExtract(t *testing.T) {
	p, err := setupWithCommittee()
	if err != nil {
		t.Fatalf("Fail to deloy contract: %v\n", err)
	}

	a, _ := hex.DecodeString("000000000000000000000000e722D8b71DCC0152D47D2438556a45D3357d631f")
	addr, err := p.vault.TestExtract(nil, a)
	fmt.Printf("addr: %x\n", addr)
}

func TestCallFunc(t *testing.T) {
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

func toByte32(s []byte) [32]byte {
	a := [32]byte{}
	copy(a[:], s)
	return a
}

func decode(s string) []byte {
	d, _ := hex.DecodeString(s)
	return d
}

func decode32(s string) [32]byte {
	return toByte32(decode(s))
}
