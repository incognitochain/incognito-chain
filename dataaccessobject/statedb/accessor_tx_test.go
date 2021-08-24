package statedb

import (
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
)

//
//func TestStoreAndHasSerialNumbers(t *testing.T) {
//	stateDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
//	if err != nil {
//		t.Fatal(err)
//	}
//	type args struct {
//		stateDB       *StateDB
//		tokenID       common.Hash
//		serialNumbers [][]byte
//		shardID       byte
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//		wantHas bool
//	}{
//		{
//			name: "shard 0",
//			args: args{
//				stateDB:       stateDB,
//				tokenID:       testGenerateTokenIDs(1)[0],
//				serialNumbers: testGenerateSerialNumberList(10),
//				shardID:       0,
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//		{
//			name: "shard 1",
//			args: args{
//				stateDB:       stateDB,
//				tokenID:       testGenerateTokenIDs(1)[0],
//				serialNumbers: testGenerateSerialNumberList(10),
//				shardID:       1,
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//		{
//			name: "shard 2",
//			args: args{
//				stateDB:       stateDB,
//				tokenID:       testGenerateTokenIDs(1)[0],
//				serialNumbers: testGenerateSerialNumberList(10),
//				shardID:       2,
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//		{
//			name: "shard 3",
//			args: args{
//				stateDB:       stateDB,
//				tokenID:       testGenerateTokenIDs(1)[0],
//				serialNumbers: testGenerateSerialNumberList(10),
//				shardID:       3,
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := StoreSerialNumbers(tt.args.stateDB, tt.args.tokenID, tt.args.serialNumbers, tt.args.shardID); (err != nil) != tt.wantErr {
//				t.Errorf("StoreSerialNumbers() error = %v, wantErr %v", err, tt.wantErr)
//			} else {
//				if len(stateDB.GetStateObjectMapForTestOnly()) != 10 && len(stateDB.GetStateObjectPendingMapForTestOnly()) != 10 {
//					t.Errorf("StoreSerialNumbers() must have 10 object")
//				}
//			}
//			tt.args.stateDB.Reset(emptyRoot)
//		})
//	}
//
//	// Actually store
//	for _, tt := range tests {
//		StoreSerialNumbers(tt.args.stateDB, tt.args.tokenID, tt.args.serialNumbers, tt.args.shardID)
//	}
//	rootHash, err := stateDB.Commit(true)
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = stateDB.Database().TrieDB().Commit(rootHash, false)
//	if err != nil {
//		t.Fatal(err)
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			for _, serialNumber := range tt.args.serialNumbers {
//				has, err := HasSerialNumber(tt.args.stateDB, tt.args.tokenID, serialNumber, tt.args.shardID)
//				if err != nil {
//					t.Errorf("HasSerialNumber() error = %v, wantErr %v", err, tt.wantErr)
//				}
//				if !has {
//					t.Errorf("HasSerialNumber() has = %v, wantHas %v", err, tt.wantHas)
//				}
//			}
//			wantSerialNumberM := make(map[string]struct{})
//			keys := []common.Hash{}
//			for _, tempSerialNumber := range tt.args.serialNumbers {
//				serialNumber := base58.Base58Check{}.Encode(tempSerialNumber, common.Base58Version)
//				wantSerialNumberM[serialNumber] = struct{}{}
//				keys = append(keys, GenerateSerialNumberObjectKey(tt.args.tokenID, tt.args.shardID, tempSerialNumber))
//			}
//			_, _, err := tt.args.stateDB.getSerialNumberState(keys[0])
//			if err != nil {
//				t.Fatal(err)
//			}
//
//		})
//	}
//}
//
//func TestStateDB_ListSerialNumber(t *testing.T) {
//	stateDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
//	if err != nil {
//		t.Fatal(err)
//	}
//	wantTokenID := testGenerateTokenIDs(10)
//	wantSerialNumberList := []map[common.Hash]map[string]struct{}{}
//	for _, shardID := range shardIDs {
//		tempWantSerialNumberM := make(map[common.Hash]map[string]struct{})
//		for _, tokenID := range wantTokenID {
//			serialNumbers := testGenerateSerialNumberList(100)
//			err = StoreSerialNumbers(stateDB, tokenID, serialNumbers, shardID)
//			if err != nil {
//				t.Fatal(err)
//			}
//			wantM := make(map[string]struct{})
//			for _, tempSerialNumber := range serialNumbers {
//				serialNumber := base58.Base58Check{}.Encode(tempSerialNumber, common.Base58Version)
//				wantM[serialNumber] = struct{}{}
//			}
//			tempWantSerialNumberM[tokenID] = wantM
//		}
//		wantSerialNumberList = append(wantSerialNumberList, tempWantSerialNumberM)
//	}
//	rootHash, err := stateDB.Commit(true)
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = stateDB.Database().TrieDB().Commit(rootHash, false)
//	if err != nil {
//		t.Fatal(err)
//	}
//	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
//	for index, shardID := range shardIDs {
//		wantSerialNumberMByToken := wantSerialNumberList[index]
//		for tokenID, wantM := range wantSerialNumberMByToken {
//			gotM, err := ListSerialNumber(tempStateDB, tokenID, shardID)
//			if err != nil {
//				t.Fatalf("ListSerialNumber() error = %v, wantErr %v", err, nil)
//			}
//			for k, _ := range wantM {
//				if _, ok := gotM[k]; !ok {
//					t.Fatalf("ListSerialNumber() want %+v but got nothing", k)
//				}
//			}
//		}
//	}
//}
//
//func TestStoreAndHasCommitment(t *testing.T) {
//	stateDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
//	if err != nil {
//		t.Fatal(err)
//	}
//	type args struct {
//		stateDB           *StateDB
//		tokenID           common.Hash
//		commitments       [][]byte
//		commitmentsLength uint64
//		shardID           byte
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//		wantHas bool
//	}{
//		{
//			name: "shard 0",
//			args: args{
//				stateDB:           stateDB,
//				tokenID:           testGenerateTokenIDs(1)[0],
//				commitments:       testGenerateCommitmentList(10),
//				shardID:           0,
//				commitmentsLength: 10,
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//		{
//			name: "shard 1",
//			args: args{
//				stateDB:           stateDB,
//				tokenID:           testGenerateTokenIDs(1)[0],
//				commitments:       testGenerateCommitmentList(10),
//				shardID:           1,
//				commitmentsLength: 10,
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//		{
//			name: "shard 2",
//			args: args{
//				stateDB:           stateDB,
//				tokenID:           testGenerateTokenIDs(1)[0],
//				commitments:       testGenerateCommitmentList(10),
//				shardID:           2,
//				commitmentsLength: 10,
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//		{
//			name: "shard 3",
//			args: args{
//				stateDB:           stateDB,
//				tokenID:           testGenerateTokenIDs(1)[0],
//				commitments:       testGenerateCommitmentList(10),
//				shardID:           3,
//				commitmentsLength: 10,
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := StoreCommitments(tt.args.stateDB, tt.args.tokenID, []byte{}, tt.args.commitments, tt.args.shardID); (err != nil) != tt.wantErr {
//				t.Errorf("StoreCommitments() error = %v, wantErr %v", err, tt.wantErr)
//			} else {
//				if len(stateDB.GetStateObjectMapForTestOnly()) != 21 && len(stateDB.GetStateObjectPendingMapForTestOnly()) != 21 {
//					t.Errorf("StoreCommitments() must have 21 object")
//				}
//			}
//			tt.args.stateDB.Reset(emptyRoot)
//		})
//	}
//
//	// Actually store
//	for _, tt := range tests {
//		err := StoreCommitments(tt.args.stateDB, tt.args.tokenID, []byte{}, tt.args.commitments, tt.args.shardID)
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//	rootHash, err := stateDB.Commit(true)
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = stateDB.Database().TrieDB().Commit(rootHash, false)
//	if err != nil {
//		t.Fatal(err)
//	}
//	for _, tt := range tests {
//		count := 0
//		t.Run(tt.name, func(t *testing.T) {
//			for _, commitment := range tt.args.commitments {
//				has, err := HasCommitment(tt.args.stateDB, tt.args.tokenID, commitment, tt.args.shardID)
//				if err != nil {
//					t.Errorf("HasCommitment() error = %v, wantErr %v", err, tt.wantErr)
//				}
//				if !has {
//					t.Errorf("HasCommitment() has = %v, wantHas %v", has, tt.wantHas)
//				}
//
//				has2, err2 := HasCommitmentIndex(tt.args.stateDB, tt.args.tokenID, uint64(count), tt.args.shardID)
//				if err2 != nil {
//					t.Errorf("HasCommitment() error = %v, wantErr %v", err2, tt.wantErr)
//				}
//				if !has2 {
//					t.Errorf("HasCommitment() has = %v, wantHas %v", has2, tt.wantHas)
//				}
//				gotCIndex, err3 := GetCommitmentIndex(tt.args.stateDB, tt.args.tokenID, commitment, tt.args.shardID)
//				if err3 != nil {
//					t.Errorf("GetCommitmentIndex() error = %v, wantErr %v", err3, tt.wantErr)
//				}
//				if gotCIndex.Uint64() != uint64(count) {
//					t.Errorf("GetCommitmentIndex() want %v, got %v", count, gotCIndex.Uint64())
//				}
//				gotC, err4 := GetCommitmentByIndex(tt.args.stateDB, tt.args.tokenID, uint64(count), tt.args.shardID)
//				if err4 != nil {
//					t.Errorf("GetCommitmentByIndex() error = %v, wantErr %v", err4, tt.wantErr)
//				}
//				if bytes.Compare(gotC, commitment) != 0 {
//					t.Errorf("GetCommitmentByIndex() want %v, got %v", commitment, gotC)
//				}
//				count++
//			}
//			gotCLength, err := GetCommitmentLength(tt.args.stateDB, tt.args.tokenID, tt.args.shardID)
//			if err != nil {
//				t.Errorf("GetCommitmentLength() error = %v, wantErr %v", err, tt.wantErr)
//			}
//			if gotCLength.Uint64() != tt.args.commitmentsLength {
//				t.Errorf("GetCommitmentLength() want %v, got %v", tt.args.commitmentsLength, gotCLength.Uint64())
//			}
//		})
//	}
//}
//
//func TestStateDB_ListCommitment(t *testing.T) {
//	stateDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
//	if err != nil {
//		t.Fatal(err)
//	}
//	wantTokenID := testGenerateTokenIDs(10)
//	wantCommitmentList := []map[common.Hash]map[string]uint64{}
//	wantCommitmentIndexList := []map[common.Hash]map[uint64]string{}
//	for _, shardID := range shardIDs {
//		tempWantCommitmentM := make(map[common.Hash]map[string]uint64)
//		tempWantCommitmentIndexM := make(map[common.Hash]map[uint64]string)
//		for _, tokenID := range wantTokenID {
//			commitments := testGenerateCommitmentList(100)
//			err = StoreCommitments(stateDB, tokenID, []byte{}, commitments, shardID)
//			if err != nil {
//				t.Fatal(err)
//			}
//			wantM := make(map[string]uint64)
//			for index, tempCommitment := range commitments {
//				commitment := base58.Base58Check{}.Encode(tempCommitment, common.Base58Version)
//				wantM[commitment] = uint64(index)
//			}
//			tempWantCommitmentM[tokenID] = wantM
//
//			wantIndexM := make(map[uint64]string)
//			for index, tempCommitment := range commitments {
//				commitment := base58.Base58Check{}.Encode(tempCommitment, common.Base58Version)
//				wantIndexM[uint64(index)] = commitment
//			}
//			tempWantCommitmentIndexM[tokenID] = wantIndexM
//		}
//		wantCommitmentList = append(wantCommitmentList, tempWantCommitmentM)
//		wantCommitmentIndexList = append(wantCommitmentIndexList, tempWantCommitmentIndexM)
//	}
//	rootHash, err := stateDB.Commit(true)
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = stateDB.Database().TrieDB().Commit(rootHash, false)
//	if err != nil {
//		t.Fatal(err)
//	}
//	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
//	for index, shardID := range shardIDs {
//		wantCommitmentMByToken := wantCommitmentList[index]
//		for tokenID, wantM := range wantCommitmentMByToken {
//			gotM, err := ListCommitment(tempStateDB, tokenID, shardID)
//			if err != nil {
//				t.Fatalf("ListCommitment() error = %v, wantErr %v", err, nil)
//			}
//			for k, v := range wantM {
//				if v2, ok := gotM[k]; !ok {
//					t.Fatalf("ListCommitment() want %+v but got nothing", k)
//				} else if v2 != v {
//					t.Fatalf("ListCommitment() want %+v but got %+v", v2, v)
//				}
//			}
//		}
//		wantCommitmentIndexMByToken := wantCommitmentIndexList[index]
//		for tokenID, wantM := range wantCommitmentIndexMByToken {
//			gotM, err := ListCommitmentIndices(tempStateDB, tokenID, shardID)
//			if err != nil {
//				t.Fatalf("ListCommitmentIndices() error = %v, wantErr %v", err, nil)
//			}
//			for k, v := range wantM {
//				if v2, ok := gotM[k]; !ok {
//					t.Fatalf("ListCommitmentIndices() want %+v but got nothing", k)
//				} else if v2 != v {
//					t.Fatalf("ListCommitmentIndices() want %+v but got %+v", v2, v)
//				}
//			}
//		}
//
//	}
//}
//
//func TestStoreAndGetOutputCoin(t *testing.T) {
//	stateDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
//	if err != nil {
//		t.Fatal(err)
//	}
//	type args struct {
//		stateDB     *StateDB
//		tokenID     common.Hash
//		outputCoins [][]byte
//		publicKey   []byte
//		shardID     byte
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//		wantHas bool
//	}{
//		{
//			name: "shard 0",
//			args: args{
//				stateDB:     stateDB,
//				tokenID:     testGenerateTokenIDs(1)[0],
//				outputCoins: testGenerateOutputCoinList(10),
//				shardID:     0,
//				publicKey:   testGeneratePublicKeyList(1)[0],
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//		{
//			name: "shard 1",
//			args: args{
//				stateDB:     stateDB,
//				tokenID:     testGenerateTokenIDs(1)[0],
//				outputCoins: testGenerateOutputCoinList(10),
//				shardID:     1,
//				publicKey:   testGeneratePublicKeyList(1)[0],
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//		{
//			name: "shard 2",
//			args: args{
//				stateDB:     stateDB,
//				tokenID:     testGenerateTokenIDs(1)[0],
//				outputCoins: testGenerateOutputCoinList(10),
//				shardID:     2,
//				publicKey:   testGeneratePublicKeyList(1)[0],
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//		{
//			name: "shard 3",
//			args: args{
//				stateDB:     stateDB,
//				tokenID:     testGenerateTokenIDs(1)[0],
//				outputCoins: testGenerateOutputCoinList(10),
//				shardID:     3,
//				publicKey:   testGeneratePublicKeyList(1)[0],
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := StoreOutputCoins(tt.args.stateDB, tt.args.tokenID, tt.args.publicKey, tt.args.outputCoins, tt.args.shardID); (err != nil) != tt.wantErr {
//				t.Errorf("StoreOutputCoins() error = %v, wantErr %v", err, tt.wantErr)
//			} else {
//				if len(stateDB.GetStateObjectMapForTestOnly()) != 10 && len(stateDB.GetStateObjectPendingMapForTestOnly()) != 10 {
//					t.Errorf("StoreOutputCoins() must have 10 object")
//				}
//			}
//			tt.args.stateDB.Reset(emptyRoot)
//		})
//	}
//
//	// Actually store
//	for _, tt := range tests {
//		err := StoreOutputCoins(tt.args.stateDB, tt.args.tokenID, tt.args.publicKey, tt.args.outputCoins, tt.args.shardID)
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//	rootHash, err := stateDB.Commit(true)
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = stateDB.Database().TrieDB().Commit(rootHash, false)
//	if err != nil {
//		t.Fatal(err)
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			gotOutputCoins, err := GetOutcoinsByPubkey(tt.args.stateDB, tt.args.tokenID, tt.args.publicKey, tt.args.shardID)
//			if err != nil {
//				t.Errorf("GetOutcoinsByPubkey() error = %v, wantErr %v", err, tt.wantErr)
//			}
//			for _, wantOutputCoin := range tt.args.outputCoins {
//				flag := false
//				for _, gotOutputCoin := range gotOutputCoins {
//					if bytes.Compare(wantOutputCoin, gotOutputCoin) == 0 {
//						flag = true
//						break
//					}
//				}
//				if !flag {
//					t.Errorf("GetOutcoinsByPubkey() want = %v, got nothing", wantOutputCoin)
//				}
//			}
//		})
//	}
//}
//
//func TestStoreSNDerivators(t *testing.T) {
//	stateDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
//	if err != nil {
//		t.Fatal(err)
//	}
//	type args struct {
//		stateDB *StateDB
//		tokenID common.Hash
//		snds    [][]byte
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//		wantHas bool
//	}{
//		{
//			name: "token 1",
//			args: args{
//				stateDB: stateDB,
//				tokenID: testGenerateTokenIDs(1)[0],
//				snds:    testGenerateSNDList(1000),
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//		{
//			name: "token 2",
//			args: args{
//				stateDB: stateDB,
//				tokenID: testGenerateTokenIDs(1)[0],
//				snds:    testGenerateSNDList(1000),
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//		{
//			name: "token 3",
//			args: args{
//				stateDB: stateDB,
//				tokenID: testGenerateTokenIDs(1)[0],
//				snds:    testGenerateSNDList(1000),
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//		{
//			name: "token 4",
//			args: args{
//				stateDB: stateDB,
//				tokenID: testGenerateTokenIDs(1)[0],
//				snds:    testGenerateSNDList(1000),
//			},
//			wantErr: false,
//			wantHas: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := StoreSNDerivators(tt.args.stateDB, tt.args.tokenID, tt.args.snds); (err != nil) != tt.wantErr {
//				t.Errorf("StoreSNDerivators() error = %v, wantErr %v", err, tt.wantErr)
//			} else {
//				if len(stateDB.GetStateObjectMapForTestOnly()) != 1000 && len(stateDB.GetStateObjectPendingMapForTestOnly()) != 1000 {
//					t.Errorf("StoreCommitments() must have 1000 object")
//				}
//			}
//			tt.args.stateDB.Reset(emptyRoot)
//		})
//	}
//	// Actually store
//	for _, tt := range tests {
//		err := StoreSNDerivators(stateDB, tt.args.tokenID, tt.args.snds)
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//	rootHash, err := stateDB.Commit(true)
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = stateDB.Database().TrieDB().Commit(rootHash, false)
//	if err != nil {
//		t.Fatal(err)
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			for _, snd := range tt.args.snds {
//				has, err := HasSNDerivator(tt.args.stateDB, tt.args.tokenID, snd)
//				if err != nil {
//					t.Errorf("HasSNDerivator() error = %v, wantErr %v", err, tt.wantErr)
//				}
//				if !has {
//					t.Errorf("HasSNDerivator() has = %v, wantErr %v", has, tt.wantHas)
//				}
//			}
//		})
//	}
//}
//
//func TestStateDB_ListSerialNumberDerivator(t *testing.T) {
//	stateDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
//	if err != nil {
//		t.Fatal(err)
//	}
//	wantTokenID := testGenerateTokenIDs(10)
//	wantSndM := make(map[common.Hash][][]byte)
//	for _, tokenID := range wantTokenID {
//		snds := testGenerateSNDList(1000)
//		err = StoreSNDerivators(stateDB, tokenID, snds)
//		if err != nil {
//			t.Fatal(err)
//		}
//		wantSndM[tokenID] = snds
//	}
//	rootHash, err := stateDB.Commit(true)
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = stateDB.Database().TrieDB().Commit(rootHash, false)
//	if err != nil {
//		t.Fatal(err)
//	}
//	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
//	for tokenID, wantSNDs := range wantSndM {
//		gotSNDs, err := ListSNDerivator(tempStateDB, tokenID)
//		if err != nil {
//			t.Errorf("ListSNDerivator() error = %v, wantErr %v", err, nil)
//		}
//		for _, wantSND := range wantSNDs {
//			flag := false
//			for _, gotSND := range gotSNDs {
//				if bytes.Compare(wantSND, gotSND) == 0 {
//					flag = true
//					break
//				}
//			}
//			if !flag {
//				t.Errorf("ListSNDerivator() want = %v, got nothing", wantSND)
//			}
//		}
//	}
//}
