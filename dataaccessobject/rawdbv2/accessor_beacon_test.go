package rawdbv2_test

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"io/ioutil"
	"os"
	"testing"
)

var (
	beaconBlocks       []*blockchain.BeaconBlock
	forkedBeaconBlock1 = blockchain.NewBeaconBlock()
	forkedBeaconBlock2 = blockchain.NewBeaconBlock()
	randomBeaconBlock1 = blockchain.NewBeaconBlock()
	randomBeaconBlock2 = blockchain.NewBeaconBlock()
	max                = 5
	db                 incdb.Database
)
var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_rawdbv2")
	if err != nil {
		panic(err)
	}
	db, err = incdb.Open("leveldb", dbPath)
	if err != nil {
		panic(err)
	}
	for i := 0; i < max; i++ {
		beaconBlock := blockchain.NewBeaconBlock()
		beaconBlock.Header.Height = uint64(i)
		if i != 0 {
			beaconBlock.Header.PreviousBlockHash = beaconBlocks[i-1].Header.Hash()
		}
		beaconBlocks = append(beaconBlocks, beaconBlock)
	}

	forkedBeaconBlock1.Header.Height = 1
	forkedBeaconBlock1.Header.Version = 2
	forkedBeaconBlock2.Header.Height = 2
	forkedBeaconBlock2.Header.Version = 2
	forkedBeaconBlock2.Header.PreviousBlockHash = forkedBeaconBlock1.Header.Hash()

	randomBeaconBlock1.Header.Height = 10001
	randomBeaconBlock2.Header.Height = 10002
	return
}()

func ResetDatabase() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_rawdbv2")
	if err != nil {
		panic(err)
	}
	db, err = incdb.Open("leveldb", dbPath)
	if err != nil {
		panic(err)
	}
}
func TestStoreBeaconBlock(t *testing.T) {
	ResetDatabase()
	for i := 0; i < max; i++ {
		err := rawdbv2.StoreBeaconBlock(db, uint64(i), beaconBlocks[i].Header.Hash(), beaconBlocks[i])
		if err != nil {
			t.Fatal(err)
		}
	}
	err := rawdbv2.StoreBeaconBlock(db, forkedBeaconBlock1.Header.Height, forkedBeaconBlock1.Header.Hash(), forkedBeaconBlock1)
	if err != nil {
		t.Fatal(err)
	}
	err1 := rawdbv2.StoreBeaconBlock(db, forkedBeaconBlock2.Header.Height, forkedBeaconBlock2.Header.Hash(), forkedBeaconBlock2)
	if err1 != nil {
		t.Fatal(err1)
	}
}

func TestHasBeaconBlock(t *testing.T) {
	ResetDatabase()
	for i := 0; i < max; i++ {
		err := rawdbv2.StoreBeaconBlock(db, uint64(i), beaconBlocks[i].Header.Hash(), beaconBlocks[i])
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < max; i++ {
		has, err := rawdbv2.HasBeaconBlock(db, beaconBlocks[i].Header.Hash())
		if !has {
			t.Fatalf("Want block %+v but got nothing", beaconBlocks[i].Header.Hash())
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	has, err := rawdbv2.HasBeaconBlock(db, randomBeaconBlock1.Header.Hash())
	if has {
		t.Fatalf("Want block %+v but got nothing", randomBeaconBlock1.Header.Hash())
	}
	if err != nil {
		t.Fatal(err)
	}
	has1, err := rawdbv2.HasBeaconBlock(db, randomBeaconBlock2.Header.Hash())
	if has1 {
		t.Fatalf("Want block %+v but got nothing", randomBeaconBlock2.Header.Hash())
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestHasBeaconBlockForked(t *testing.T) {
	ResetDatabase()
	for i := 0; i < max; i++ {
		err := rawdbv2.StoreBeaconBlock(db, uint64(i), beaconBlocks[i].Header.Hash(), beaconBlocks[i])
		if err != nil {
			t.Fatal(err)
		}
	}
	err := rawdbv2.StoreBeaconBlock(db, forkedBeaconBlock1.Header.Height, forkedBeaconBlock1.Header.Hash(), forkedBeaconBlock1)
	if err != nil {
		t.Fatal(err)
	}
	err1 := rawdbv2.StoreBeaconBlock(db, forkedBeaconBlock2.Header.Height, forkedBeaconBlock2.Header.Hash(), forkedBeaconBlock2)
	if err1 != nil {
		t.Fatal(err1)
	}
	for i := 0; i < max; i++ {
		has, err := rawdbv2.HasBeaconBlock(db, beaconBlocks[i].Header.Hash())
		if !has {
			t.Fatalf("Want block %+v but got nothing", beaconBlocks[i].Header.Hash())
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	has, err := rawdbv2.HasBeaconBlock(db, forkedBeaconBlock1.Header.Hash())
	if !has {
		t.Fatalf("Want block %+v but got nothing", forkedBeaconBlock1.Header.Hash())
	}
	if err != nil {
		t.Fatal(err)
	}
	has1, err := rawdbv2.HasBeaconBlock(db, forkedBeaconBlock2.Header.Hash())
	if !has1 {
		t.Fatalf("Want block %+v but got nothing", forkedBeaconBlock2.Header.Hash())
	}
	if err != nil {
		t.Fatal(err)
	}
}
