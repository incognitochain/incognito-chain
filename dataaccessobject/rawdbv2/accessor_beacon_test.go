package rawdbv2_test

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"io/ioutil"
	"os"
)

var (
	beaconBlocks       []*types.BeaconBlock
	forkedBeaconBlock1 = types.NewBeaconBlock()
	forkedBeaconBlock2 = types.NewBeaconBlock()
	randomBeaconBlock1 = types.NewBeaconBlock()
	randomBeaconBlock2 = types.NewBeaconBlock()
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
		beaconBlock := types.NewBeaconBlock()
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

func resetDatabase() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_rawdbv2")
	if err != nil {
		panic(err)
	}
	db, err = incdb.Open("leveldb", dbPath)
	if err != nil {
		panic(err)
	}
}
