package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"testing"
)

func TestNewRebuildInfo(t *testing.T) {
	h1 := common.Hash{}.NewHashFromStr2("123")
	h2 := common.Hash{}.NewHashFromStr2("456")
	info1 := NewRebuildInfo(common.STATEDB_BATCH_COMMIT_MODE, h1, h2, 1800, 900)
	fmt.Println(info1.ToBytes())

	info2 := NewEmptyRebuildInfo(0)
	info2.FromBytes(info1.ToBytes())

	if info1.mode != info2.mode {
		t.Fatal("!= mode")
	}

	if info1.rebuildRootHash.String() != info2.rebuildRootHash.String() {
		t.Fatal("!= rebuild")
	}

	if info1.pivotRootHash.String() != info2.pivotRootHash.String() {
		t.Fatal("!= pivotroot")
	}

	if info1.rebuildFFIndex != info2.rebuildFFIndex {
		t.Fatal("!= rebuildFFIndex")
	}
	if info1.pivotFFIndex != info2.pivotFFIndex {
		t.Fatal("!= pivotffindex")
	}
	fmt.Printf("%+v", info2)
}
