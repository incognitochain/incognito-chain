package statedb

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

func TestStoreAndGetBeaconCommittee(t *testing.T) {
	number := 20
	beaconCommittees := committeePublicKeys[:number]
	beaconCommitteesStruct, _ := incognitokey.CommitteeBase58KeyListToStruct(beaconCommittees)
	rewardReceiver := make(map[string]string)
	autoStaking := make(map[string]bool)
	for index, beaconCommittee := range beaconCommitteesStruct {
		incPublicKey := beaconCommittee.GetIncKeyBase58()
		paymentAddress := receiverPaymentAddresses[index]
		rewardReceiver[incPublicKey] = paymentAddress
		autoStaking[beaconCommittees[index]] = true
	}
	sDB, _ := NewWithPrefixTrie(emptyRoot, wrarperDB)
	err := StoreBeaconCommittee(sDB, beaconCommitteesStruct)
	if err != nil {
		t.Fatal(err)
	}
	rootHash, _ := sDB.Commit(true)
	err1 := sDB.Database().TrieDB().Commit(rootHash, false)
	if err1 != nil {
		t.Fatal(err1)
	}
	gotBeaconCommitteeStruct := GetBeaconCommittee(sDB)
	if len(gotBeaconCommitteeStruct) != number {
		t.Fatalf("expect number of committee %+v, but got %+v", len(gotBeaconCommitteeStruct), number)
	}
	for index, wantC := range beaconCommitteesStruct {
		if !wantC.IsEqual(gotBeaconCommitteeStruct[index]) {
			t.Fatalf("want %+v, got %+v", wantC, gotBeaconCommitteeStruct[index])
		}
	}
}

func TestStoreAndGetShardCommittee(t *testing.T) {
	number := 20
	shardID := byte(0)
	shardCommittees := committeePublicKeys[:number]
	shardCommitteesStruct, _ := incognitokey.CommitteeBase58KeyListToStruct(shardCommittees)
	rewardReceiver := make(map[string]string)
	autoStaking := make(map[string]bool)
	for index, beaconCommittee := range shardCommitteesStruct {
		incPublicKey := beaconCommittee.GetIncKeyBase58()
		paymentAddress := receiverPaymentAddresses[index]
		rewardReceiver[incPublicKey] = paymentAddress
		autoStaking[shardCommittees[index]] = true
	}
	sDB, _ := NewWithPrefixTrie(emptyRoot, wrarperDB)
	err := StoreOneShardCommittee(sDB, shardID, shardCommitteesStruct)
	if err != nil {
		t.Fatal(err)
	}
	rootHash, _ := sDB.Commit(true)
	err1 := sDB.Database().TrieDB().Commit(rootHash, false)
	if err1 != nil {
		t.Fatal(err1)
	}
	gotShardCommitteeStruct := GetOneShardCommittee(sDB, shardID)
	if len(gotShardCommitteeStruct) != number {
		t.Fatalf("expect number of committee %+v, but got %+v", len(gotShardCommitteeStruct), number)
	}
	for index, wantC := range shardCommitteesStruct {
		if !wantC.IsEqual(gotShardCommitteeStruct[index]) {
			t.Fatalf("want %+v, got %+v", wantC, gotShardCommitteeStruct[index])
		}
	}
}

func TestDeleteOneShardCommittee(t *testing.T) {
	number := 20
	split := 10
	shardID := byte(0)
	shardCommittees := committeePublicKeys[:number]
	shardCommitteesStruct, _ := incognitokey.CommitteeBase58KeyListToStruct(shardCommittees)
	rewardReceiver := make(map[string]string)
	autoStaking := make(map[string]bool)
	for index, beaconCommittee := range shardCommitteesStruct {
		incPublicKey := beaconCommittee.GetIncKeyBase58()
		paymentAddress := receiverPaymentAddresses[index]
		rewardReceiver[incPublicKey] = paymentAddress
		autoStaking[shardCommittees[index]] = true
	}
	sDB, _ := NewWithPrefixTrie(emptyRoot, wrarperDB)
	err := StoreOneShardCommittee(sDB, shardID, shardCommitteesStruct)
	if err != nil {
		t.Fatal(err)
	}
	rootHash, _ := sDB.Commit(true)
	err1 := sDB.Database().TrieDB().Commit(rootHash, false)
	if err1 != nil {
		t.Fatal(err1)
	}
	deletedShardCommittee := shardCommitteesStruct[:split]
	remainedShardCommittee := shardCommitteesStruct[split:]
	err2 := DeleteOneShardCommittee(sDB, shardID, deletedShardCommittee)
	if err2 != nil {
		t.Fatal(err2)
	}
	rootHash2, _ := sDB.Commit(true)
	err3 := sDB.Database().TrieDB().Commit(rootHash2, false)
	if err3 != nil {
		t.Fatal(err3)
	}
	gotShardCommitteeStruct := GetOneShardCommittee(sDB, shardID)
	if len(gotShardCommitteeStruct) != len(remainedShardCommittee) {
		t.Fatalf("expect number of committee %+v, but got %+v", len(gotShardCommitteeStruct), len(remainedShardCommittee))
	}
	for index, wantC := range remainedShardCommittee {
		if !wantC.IsEqual(gotShardCommitteeStruct[index]) {
			t.Fatalf("want %+v, got %+v", wantC, gotShardCommitteeStruct[index])
		}
	}
	for _, wantC := range deletedShardCommittee {
		for _, gotC := range gotShardCommitteeStruct {
			if wantC.IsEqual(gotC) {
				t.Fatalf("want %+v, got %+v", wantC, gotC)
			}
		}
	}
}

func TestDeleteBeaconCommittee(t *testing.T) {
	number := 20
	split := 10
	beaconCommittees := committeePublicKeys[:number]
	beaconCommitteesStruct, _ := incognitokey.CommitteeBase58KeyListToStruct(beaconCommittees)
	rewardReceiver := make(map[string]string)
	autoStaking := make(map[string]bool)
	for index, beaconCommittee := range beaconCommitteesStruct {
		incPublicKey := beaconCommittee.GetIncKeyBase58()
		paymentAddress := receiverPaymentAddresses[index]
		rewardReceiver[incPublicKey] = paymentAddress
		autoStaking[beaconCommittees[index]] = true
	}
	sDB, _ := NewWithPrefixTrie(emptyRoot, wrarperDB)
	err := StoreBeaconCommittee(sDB, beaconCommitteesStruct)
	if err != nil {
		t.Fatal(err)
	}
	rootHash, _ := sDB.Commit(true)
	err1 := sDB.Database().TrieDB().Commit(rootHash, false)
	if err1 != nil {
		t.Fatal(err1)
	}
	deletedBeaconCommittee := beaconCommitteesStruct[:split]
	remainedBeaconCommittee := beaconCommitteesStruct[split:]
	err2 := DeleteBeaconCommittee(sDB, deletedBeaconCommittee)
	if err2 != nil {
		t.Fatal(err2)
	}
	rootHash2, _ := sDB.Commit(true)
	err3 := sDB.Database().TrieDB().Commit(rootHash2, false)
	if err3 != nil {
		t.Fatal(err3)
	}
	gotBeaconCommitteeStruct := GetBeaconCommittee(sDB)
	if len(gotBeaconCommitteeStruct) != len(remainedBeaconCommittee) {
		t.Fatalf("expect number of committee %+v, but got %+v", len(gotBeaconCommitteeStruct), len(remainedBeaconCommittee))
	}
	for index, wantC := range remainedBeaconCommittee {
		if !wantC.IsEqual(gotBeaconCommitteeStruct[index]) {
			t.Fatalf("want %+v, got %+v", wantC, gotBeaconCommitteeStruct[index])
		}
	}
	for _, wantC := range deletedBeaconCommittee {
		for _, gotC := range gotBeaconCommitteeStruct {
			if wantC.IsEqual(gotC) {
				t.Fatalf("want %+v, got %+v", wantC, gotC)
			}
		}
	}
}

func TestStoreAndGetStakerInfo(t *testing.T) {
	number := 20
	shardID := byte(0)
	shardCommittees := committeePublicKeys[:number]
	shardCommitteesStruct, _ := incognitokey.CommitteeBase58KeyListToStruct(shardCommittees)
	rewardReceiver := make(map[string]privacy.PaymentAddress)
	autoStaking := make(map[string]bool)
	stakingTx := make(map[string]common.Hash)
	for index, beaconCommittee := range shardCommitteesStruct {
		incPublicKey := beaconCommittee.GetIncKeyBase58()
		paymentAddress := receiverPaymentAddresses[index]
		wl, err := wallet.Base58CheckDeserialize(paymentAddress)
		if err != nil {
			t.Fatal(err)
		}
		rewardReceiver[incPublicKey] = wl.KeySet.PaymentAddress
		autoStaking[shardCommittees[index]] = true
		stakingTx[shardCommittees[index]] = common.HashH([]byte{0})
	}
	sDB, _ := NewWithPrefixTrie(emptyRoot, wrarperDB)
	err := StoreOneShardCommittee(sDB, shardID, shardCommitteesStruct)
	if err != nil {
		t.Fatal(err)
	}
	err = StoreStakerInfo(sDB, shardCommitteesStruct, rewardReceiver, autoStaking, stakingTx)
	if err != nil {
		t.Fatal(err)
	}
	rootHash, _ := sDB.Commit(true)
	err1 := sDB.Database().TrieDB().Commit(rootHash, false)
	if err1 != nil {
		t.Fatal(err1)
	}
	gotShardCommitteeStruct := GetOneShardCommittee(sDB, shardID)
	if len(gotShardCommitteeStruct) != number {
		t.Fatalf("expect number of committee %+v, but got %+v", len(gotShardCommitteeStruct), number)
	}
	for index, wantC := range shardCommitteesStruct {
		if !wantC.IsEqual(gotShardCommitteeStruct[index]) {
			t.Fatalf("want %+v, got %+v", wantC, gotShardCommitteeStruct[index])
		}
		cString, _ := wantC.ToBase58()
		s, ok, err := GetStakerInfo(sDB, cString)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("Can not get committee info")
		}
		if s == nil {
			t.Fatal("wtf")
		}
	}
}

func TestStoreSlashingCommittee(t *testing.T) {
	sDB, _ := NewWithPrefixTrie(emptyRoot, wrarperDB)
	m1 := map[byte][]string{
		0: []string{committeePublicKeys[0], committeePublicKeys[1], committeePublicKeys[2]},
	}
	m2 := map[byte][]string{
		0: []string{committeePublicKeys[0], committeePublicKeys[1], committeePublicKeys[2]},
		1: []string{committeePublicKeys[3], committeePublicKeys[4], committeePublicKeys[5]},
	}
	m3 := map[byte][]string{
		0: []string{committeePublicKeys[6], committeePublicKeys[8], committeePublicKeys[10]},
		1: []string{committeePublicKeys[7], committeePublicKeys[9], committeePublicKeys[11]},
	}
	m4 := map[byte][]string{
		0: []string{},
		1: []string{},
		2: []string{},
	}
	type args struct {
		stateDB            *StateDB
		epoch              uint64
		slashingCommittees map[byte][]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "store 1 shard",
			args: args{
				stateDB:            sDB,
				epoch:              2,
				slashingCommittees: m1,
			},
			wantErr: false,
		},
		{
			name: "store 2 shard",
			args: args{
				stateDB:            sDB,
				epoch:              3,
				slashingCommittees: m2,
			},
			wantErr: false,
		},
		{
			name: "store 2 shard",
			args: args{
				stateDB:            sDB,
				epoch:              4,
				slashingCommittees: m3,
			},
			wantErr: false,
		},
		{
			name: "store 2 shard, no slashing committee",
			args: args{
				stateDB:            sDB,
				epoch:              5,
				slashingCommittees: m4,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := StoreSlashingCommittee(tt.args.stateDB, tt.args.epoch, tt.args.slashingCommittees); (err != nil) != tt.wantErr {
				t.Errorf("StoreSlashingCommittee() error = %v, wantErr %v", err, tt.wantErr)
			}
			rootHash, err := tt.args.stateDB.Commit(true)
			if err != nil {
				t.Errorf("tt.args.stateDB.Commit() error = %v, wantErr %v", err, nil)
			}
			err1 := tt.args.stateDB.Database().TrieDB().Commit(rootHash, false)
			if err1 != nil {
				t.Errorf("tt.args.stateDB.Commit() error = %v, wantErr %v", err1, nil)
			}
		})
	}

	got1 := GetSlashingCommittee(sDB, 2)
	got2 := GetSlashingCommittee(sDB, 3)
	got3 := GetSlashingCommittee(sDB, 4)
	got4 := GetSlashingCommittee(sDB, 5)
	if !reflect.DeepEqual(got1, m1) {
		t.Fatalf("epoch %+v, want %+v, got %+v", 2, m1, got1)
	}
	if !reflect.DeepEqual(got2, m2) {
		t.Fatalf("epoch %+v, want %+v, got %+v", 3, m2, got2)
	}
	if !reflect.DeepEqual(got3, m3) {
		t.Fatalf("epoch %+v, want %+v, got %+v", 4, m3, got3)
	}
	if !reflect.DeepEqual(got4, m4) {
		t.Fatalf("epoch %+v, want %+v, got %+v", 4, m4, got4)
	}
}
