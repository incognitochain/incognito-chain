package bridge

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

var genesisKey *ecdsa.PrivateKey
var auth *bind.TransactOpts

type Platform struct {
	c   *Bridge
	sim *backends.SimulatedBackend
}

func (p *Platform) BalanceOf(addr common.Address) *big.Int {
	b, _ := p.c.BalanceOf(nil, addr)
	return b
}

func init() {
	fmt.Println("Initializing...")
	genesisKey, _ = crypto.GenerateKey()
	auth = bind.NewKeyedTransactor(genesisKey)
}

func setup() (*Platform, error) {
	alloc := make(core.GenesisAlloc)
	balance := big.NewInt(123000000000000000)
	alloc[auth.From] = core.GenesisAccount{Balance: balance}
	sim := backends.NewSimulatedBackend(alloc, 6000000)

	name := "myERC20"
	symbol := "@"
	decimals := big.NewInt(12)
	totalSupply := big.NewInt(1000000000)
	_, _, c, err := DeployBridge(auth, sim, name, symbol, decimals, totalSupply)
	if err != nil {
		return nil, err
	}
	sim.Commit()

	ts, _ := c.TotalSupply(nil)
	fmt.Println("deployed, totalSupply:", ts)
	return &Platform{c, sim}, nil
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

func TestTransfer(t *testing.T) {
	p, err := setup()
	if err != nil {
		t.Fatalf("Fail to deloy contract: %v\n", err)
	}

	a := newAccount()
	to := a.Address
	value := big.NewInt(1000)

	if _, err = p.c.Transfer(auth, to, value); err != nil {
		t.Error(err)
	}

	p.sim.Commit()

	e := big.NewInt(1000)
	if v := p.BalanceOf(a.Address); v.Cmp(e) != 0 {
		t.Errorf("transfer failed, new balance = %v, expected %v", v, e)
	}
	fmt.Println(p.BalanceOf(auth.From))

	// Test calling swapBeacon
	const comm_height = 1
	inst_length := 100
	beacon_length := 1000
	bridge_length := 1000
	newComRoot := [32]byte{}
	inst := make([]byte, inst_length)
	for i := 0; i < inst_length; i++ {
		inst[i] = byte(1)
	}
	beaconInstPath := [comm_height][32]byte{}
	beaconPathIsLeft := [comm_height]bool{}
	beaconInstRoot := [32]byte{}
	beaconBlkData := make([]byte, beacon_length)
	beaconBlkHash := [32]byte{}
	beaconSignerPubkeys := [comm_height][32]byte{}
	beaconSignerSig := [32]byte{}
	beaconSignerPaths := [comm_height][32]byte{}
	bridgeInstPath := [comm_height][32]byte{}
	bridgePathIsLeft := [comm_height]bool{}
	bridgeInstRoot := [32]byte{}
	bridgeBlkData := make([]byte, bridge_length)
	bridgeBlkHash := [32]byte{0}
	bridgeSignerPubkeys := [comm_height][32]byte{}
	bridgeSignerSig := [32]byte{0}
	bridgeSignerPaths := [comm_height][32]byte{}
	if _, err = p.c.SwapBeacon(
		auth,
		newComRoot,
		inst,
		beaconInstPath,
		beaconPathIsLeft,
		beaconInstRoot,
		beaconBlkData,
		beaconBlkHash,
		beaconSignerPubkeys,
		beaconSignerSig,
		beaconSignerPaths,
		bridgeInstPath,
		bridgePathIsLeft,
		bridgeInstRoot,
		bridgeBlkData,
		bridgeBlkHash,
		bridgeSignerPubkeys,
		bridgeSignerSig,
		bridgeSignerPaths,
	); err != nil {
		fmt.Println("err:", err)
	}

	x, y := p.c.Get(nil)
	fmt.Println(x, y)

	x2, y2 := p.c.ParseSwapBeaconInst(nil, nil)
	fmt.Println(x2, y2)

	x3, _ := p.c.GetHash(nil, inst)
	fmt.Printf("%x\n", x3)

	x4 := crypto.Keccak256(inst[:])
	fmt.Printf("%x\n", x4)

	leaf := [32]byte{1, 2, 3}
	right := [32]byte{}
	comb := append(leaf[:], right[:]...)
	path := [1][32]byte{right}
	left := [1]bool{false}
	r := crypto.Keccak256(comb)
	root := [32]byte{}
	copy(root[:], r)
	leaf = [32]byte{1, 2, 3}
	x5, _ := p.c.InMerkleTree(nil, leaf, root, path, left)
	fmt.Println(x5)
}

func TestMyErc20(t *testing.T) {
	// <setup code>
	//t.Run("A=1", func(t *testing.T) { ... })
	//t.Run("A=2", func(t *testing.T) { ... })
	//t.Run("B=1", func(t *testing.T) { ... })
	// <tear-down code>
}
