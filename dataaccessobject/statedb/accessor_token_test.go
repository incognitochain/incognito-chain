package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
	"testing"
)

func TestStateDB_StorePrivacyToken(t *testing.T) {
	sDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	tokenIDs := testGenerateTokenIDs(10)
	type args struct {
		stateDB   *StateDB
		tokenID   common.Hash
		name      string
		symbol    string
		tokenType int
		mintable  bool
		amount    uint64
		info      []byte
		txHash    common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test-1",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[0],
				name:      "token-1",
				symbol:    "TK1",
				tokenType: InitToken,
				mintable:  false,
				amount:    10000,
				info:      []byte{},
				txHash:    common.Hash{1},
			},
			wantErr: false,
		},
		{
			name: "test-2",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[1],
				name:      "token-2",
				symbol:    "TK2",
				tokenType: InitToken,
				mintable:  false,
				amount:    1000,
				info:      []byte{},
				txHash:    common.Hash{2},
			},
			wantErr: false,
		},
		{
			name: "test-3",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[2],
				name:      "token-3",
				symbol:    "TK3",
				tokenType: CrossShardToken,
				mintable:  false,
				amount:    202020,
				info:      []byte{},
				txHash:    common.Hash{3},
			},
			wantErr: false,
		},
		{
			name: "test-4",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[3],
				name:      "token-4",
				symbol:    "TK4",
				tokenType: BridgeToken,
				mintable:  false,
				amount:    0,
				info:      []byte{},
				txHash:    common.Hash{4},
			},
			wantErr: false,
		},
		{
			name: "test-5",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[4],
				name:      "token-5",
				symbol:    "TK5",
				tokenType: UnknownToken,
				mintable:  false,
				amount:    82020,
				info:      []byte{},
				txHash:    common.Hash{5},
			},
			wantErr: false,
		},
		{
			name: "test-6",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[4],
				name:      "token-6",
				symbol:    "TK6",
				tokenType: BridgeToken,
				mintable:  false,
				amount:    12345,
				info:      []byte{},
				txHash:    common.Hash{6},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		if err := StorePrivacyToken(sDB, tt.args.tokenID, tt.args.name, tt.args.symbol, tt.args.tokenType, tt.args.mintable, tt.args.amount, tt.args.info, tt.args.txHash); (err != nil) != tt.wantErr {
			t.Errorf("StorePrivacyToken() error = %v, wantErr %v", err, tt.wantErr)
		}
	}
	rootHash, _, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests[:5] {
		key := GenerateTokenObjectKey(tt.args.tokenID)
		tokenState, has, err := sDB.getTokenState(key)
		if err != nil {
			t.Fatal(err)
		}
		if !has || tokenState == nil {
			t.Fatalf("want token %+v but got nothing", tt.args.tokenID)
		}
		if !tokenState.tokenID.IsEqual(&tt.args.tokenID) {
			t.Fatalf("want token %+v but got %+v", tt.args.tokenID, tokenState.tokenID)
		}
		if !tokenState.initTx.IsEqual(&tt.args.txHash) {
			t.Fatalf("want tx hash %+v but got %+v", tt.args.txHash, tokenState.initTx)
		}
		if tokenState.amount != tt.args.amount {
			t.Fatalf("want amount %+v but got %+v", tt.args.amount, tokenState.amount)
		}
		if tokenState.propertyName != tt.args.name {
			t.Fatalf("want name %+v but got %+v", tt.args.name, tokenState.propertyName)
		}
		if tokenState.propertySymbol != tt.args.symbol {
			t.Fatalf("want symbol %+v but got %+v", tt.args.symbol, tokenState.propertySymbol)
		}
		if tokenState.tokenType != tt.args.tokenType {
			t.Fatalf("want token type %+v but got %+v", tt.args.symbol, tokenState.propertySymbol)
		}
	}
	duplicateTest := tests[5]
	key := GenerateTokenObjectKey(duplicateTest.args.tokenID)
	tokenState, has, err := sDB.getTokenState(key)
	if err != nil {
		t.Fatal(err)
	}
	if !has || tokenState == nil {
		t.Fatalf("want token %+v but got nothing", duplicateTest.args.tokenID)
	}
	if !tokenState.tokenID.IsEqual(&duplicateTest.args.tokenID) {
		t.Fatalf("want token %+v but got %+v", duplicateTest.args.tokenID, tokenState.tokenID)
	}
	tt := tests[4]
	if !tokenState.initTx.IsEqual(&tt.args.txHash) {
		t.Fatalf("want tx hash %+v but got %+v", tt.args.txHash, tokenState.initTx)
	}
	if tokenState.amount != tt.args.amount {
		t.Fatalf("want amount %+v but got %+v", tt.args.amount, tokenState.amount)
	}
	if tokenState.propertyName != tt.args.name {
		t.Fatalf("want name %+v but got %+v", tt.args.name, tokenState.propertyName)
	}
	if tokenState.propertySymbol != tt.args.symbol {
		t.Fatalf("want symbol %+v but got %+v", tt.args.symbol, tokenState.propertySymbol)
	}
	if tokenState.tokenType != tt.args.tokenType {
		t.Fatalf("want token type %+v but got %+v", tt.args.symbol, tokenState.propertySymbol)
	}
}

func TestStateDB_StorePrivacyTokenTx(t *testing.T) {
	sDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	tokenIDs := testGenerateTokenIDs(10)
	type args struct {
		stateDB *StateDB
		tokenID common.Hash
		txHash  common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test-1",
			args: args{
				stateDB: sDB,
				tokenID: tokenIDs[0],
				txHash:  common.Hash{1, 1},
			},
			wantErr: false,
		},
		{
			name: "test-2",
			args: args{
				stateDB: sDB,
				tokenID: tokenIDs[0],
				txHash:  common.Hash{1, 2},
			},
			wantErr: false,
		},
		{
			name: "test-3",
			args: args{
				stateDB: sDB,
				tokenID: tokenIDs[0],
				txHash:  common.Hash{1, 3},
			},
			wantErr: false,
		},
		{
			name: "test-4",
			args: args{
				stateDB: sDB,
				tokenID: tokenIDs[0],
				txHash:  common.Hash{1, 4},
			},
			wantErr: false,
		},
		{
			name: "test-5",
			args: args{
				stateDB: sDB,
				tokenID: tokenIDs[0],
				txHash:  common.Hash{1, 5},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		if err := StorePrivacyTokenTx(tt.args.stateDB, tt.args.tokenID, tt.args.txHash); (err != nil) != tt.wantErr {
			t.Errorf("StorePrivacyTokenTx() error = %v, wantErr %v", err, tt.wantErr)
		}
	}
	rootHash, _, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	key := GenerateTokenObjectKey(tokenIDs[0])
	tokenState, has, err := sDB.getTokenState(key)
	if err != nil {
		t.Fatal(err)
	}
	if !has || tokenState == nil {
		t.Fatalf("want token %+v but got nothing", tokenIDs[0])
	}
	if !tokenState.tokenID.IsEqual(&tokenIDs[0]) {
		t.Fatalf("want token %+v but got %+v", tokenIDs[0], tokenState.tokenID)
	}
	txs := sDB.getTokenTxs(tokenIDs[0])
	if len(txs) != len(tests) {
		t.Fatalf("want len %+v but got %+v", len(tests), len(txs)-1)
	}
	for _, wantTx := range tests {
		flag := false
		for _, gotTx := range txs {
			if wantTx.args.txHash.IsEqual(&gotTx) {
				flag = true
				break
			}
		}
		if !flag {
			t.Fatalf("want %+v but got nothing", wantTx)
		}
	}
}

func TestStateDB_ListPrivacyToken(t *testing.T) {
	sDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	tokenIDs := testGenerateTokenIDs(10)
	type args struct {
		stateDB   *StateDB
		tokenID   common.Hash
		name      string
		symbol    string
		tokenType int
		mintable  bool
		amount    uint64
		info      []byte
		txHash    common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test-1",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[0],
				name:      "token-1",
				symbol:    "TK1",
				tokenType: InitToken,
				mintable:  false,
				amount:    10000,
				info:      []byte{},
				txHash:    common.Hash{1},
			},
			wantErr: false,
		},
		{
			name: "test-2",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[1],
				name:      "token-2",
				symbol:    "TK2",
				tokenType: InitToken,
				mintable:  false,
				amount:    1000,
				info:      []byte{},
				txHash:    common.Hash{2},
			},
			wantErr: false,
		},
		{
			name: "test-3",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[2],
				name:      "token-3",
				symbol:    "TK3",
				tokenType: CrossShardToken,
				mintable:  false,
				amount:    202020,
				info:      []byte{},
				txHash:    common.Hash{3},
			},
			wantErr: false,
		},
		{
			name: "test-4",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[3],
				name:      "token-4",
				symbol:    "TK4",
				tokenType: BridgeToken,
				mintable:  false,
				amount:    0,
				info:      []byte{},
				txHash:    common.Hash{4},
			},
			wantErr: false,
		},
		{
			name: "test-5",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[4],
				name:      "token-5",
				symbol:    "TK5",
				tokenType: UnknownToken,
				mintable:  false,
				amount:    82020,
				info:      []byte{},
				txHash:    common.Hash{5},
			},
			wantErr: false,
		},
		{
			name: "test-6",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[4],
				name:      "token-6",
				symbol:    "TK6",
				tokenType: BridgeToken,
				mintable:  false,
				amount:    12345,
				info:      []byte{},
				txHash:    common.Hash{6},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		if err := StorePrivacyToken(sDB, tt.args.tokenID, tt.args.name, tt.args.symbol, tt.args.tokenType, tt.args.mintable, tt.args.amount, tt.args.info, tt.args.txHash); (err != nil) != tt.wantErr {
			t.Errorf("StorePrivacyToken() error = %v, wantErr %v", err, tt.wantErr)
		}
	}
	rootHash, _, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	tokenStates := ListPrivacyToken(sDB)
	for _, tt := range tests[:5] {
		flag := false
		tokenState := NewTokenState()
		for _, tempTokenState := range tokenStates {
			if tempTokenState.tokenID.IsEqual(&tt.args.tokenID) {
				flag = true
				tokenState = tempTokenState
				break
			}
		}
		if flag {
			if !tokenState.initTx.IsEqual(&tt.args.txHash) {
				t.Fatalf("want tx hash %+v but got %+v", tt.args.txHash, tokenState.initTx)
			}
			if tokenState.amount != tt.args.amount {
				t.Fatalf("want amount %+v but got %+v", tt.args.amount, tokenState.amount)
			}
			if tokenState.propertyName != tt.args.name {
				t.Fatalf("want name %+v but got %+v", tt.args.name, tokenState.propertyName)
			}
			if tokenState.propertySymbol != tt.args.symbol {
				t.Fatalf("want symbol %+v but got %+v", tt.args.symbol, tokenState.propertySymbol)
			}
			if tokenState.tokenType != tt.args.tokenType {
				t.Fatalf("want token type %+v but got %+v", tt.args.symbol, tokenState.propertySymbol)
			}
		} else {
			t.Fatalf("want token %+v but not found", tt.args.tokenID)
		}
	}
}

func TestStateDB_GetPrivacyTokenTxs(t *testing.T) {
	sDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	tokenIDs := testGenerateTokenIDs(10)
	type args struct {
		stateDB *StateDB
		tokenID common.Hash
		txHash  common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test-1",
			args: args{
				stateDB: sDB,
				tokenID: tokenIDs[0],
				txHash:  common.Hash{1, 1},
			},
			wantErr: false,
		},
		{
			name: "test-2",
			args: args{
				stateDB: sDB,
				tokenID: tokenIDs[0],
				txHash:  common.Hash{1, 2},
			},
			wantErr: false,
		},
		{
			name: "test-3",
			args: args{
				stateDB: sDB,
				tokenID: tokenIDs[0],
				txHash:  common.Hash{1, 3},
			},
			wantErr: false,
		},
		{
			name: "test-4",
			args: args{
				stateDB: sDB,
				tokenID: tokenIDs[0],
				txHash:  common.Hash{1, 4},
			},
			wantErr: false,
		},
		{
			name: "test-5",
			args: args{
				stateDB: sDB,
				tokenID: tokenIDs[0],
				txHash:  common.Hash{1, 5},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		if err := StorePrivacyTokenTx(tt.args.stateDB, tt.args.tokenID, tt.args.txHash); (err != nil) != tt.wantErr {
			t.Errorf("StorePrivacyTokenTx() error = %v, wantErr %v", err, tt.wantErr)
		}
	}
	rootHash, _, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	txs := GetPrivacyTokenTxs(sDB, tokenIDs[0])
	if len(txs) != len(tests) {
		t.Fatalf("want len %+v but got %+v", len(tests), len(txs))
	}
	for _, wantTx := range tests {
		flag := false
		for _, gotTx := range txs {
			if wantTx.args.txHash.IsEqual(&gotTx) {
				flag = true
				break
			}
		}
		if !flag {
			t.Fatalf("want %+v but got nothing", wantTx)
		}
	}
}

func TestStateDB_PrivacyTokenIDExisted(t *testing.T) {
	sDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	tokenIDs := testGenerateTokenIDs(10)
	type args struct {
		stateDB   *StateDB
		tokenID   common.Hash
		name      string
		symbol    string
		tokenType int
		mintable  bool
		amount    uint64
		info      []byte
		txHash    common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test-1",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[0],
				name:      "token-1",
				symbol:    "TK1",
				tokenType: InitToken,
				mintable:  false,
				amount:    10000,
				info:      []byte{},
				txHash:    common.Hash{1},
			},
			wantErr: false,
		},
		{
			name: "test-2",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[1],
				name:      "token-2",
				symbol:    "TK2",
				tokenType: InitToken,
				mintable:  false,
				amount:    1000,
				info:      []byte{},
				txHash:    common.Hash{2},
			},
			wantErr: false,
		},
		{
			name: "test-3",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[2],
				name:      "token-3",
				symbol:    "TK3",
				tokenType: CrossShardToken,
				mintable:  false,
				amount:    202020,
				info:      []byte{},
				txHash:    common.Hash{3},
			},
			wantErr: false,
		},
		{
			name: "test-4",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[3],
				name:      "token-4",
				symbol:    "TK4",
				tokenType: BridgeToken,
				mintable:  false,
				amount:    0,
				info:      []byte{},
				txHash:    common.Hash{4},
			},
			wantErr: false,
		},
		{
			name: "test-5",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[4],
				name:      "token-5",
				symbol:    "TK5",
				tokenType: UnknownToken,
				mintable:  false,
				amount:    82020,
				info:      []byte{},
				txHash:    common.Hash{5},
			},
			wantErr: false,
		},
		{
			name: "test-6",
			args: args{
				stateDB:   sDB,
				tokenID:   tokenIDs[5],
				name:      "token-6",
				symbol:    "TK6",
				tokenType: BridgeToken,
				mintable:  false,
				amount:    12345,
				info:      []byte{},
				txHash:    common.Hash{6},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		if err := StorePrivacyToken(sDB, tt.args.tokenID, tt.args.name, tt.args.symbol, tt.args.tokenType, tt.args.mintable, tt.args.amount, tt.args.info, tt.args.txHash); (err != nil) != tt.wantErr {
			t.Errorf("StorePrivacyToken() error = %v, wantErr %v", err, tt.wantErr)
		}
	}
	rootHash, _, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	for index, tokenID := range tokenIDs {
		if index <= 5 {
			if !PrivacyTokenIDExisted(sDB, tokenID) {
				t.Fatalf("want token %+v exist", tokenID)
			}
		} else {
			if PrivacyTokenIDExisted(sDB, tokenID) {
				t.Fatalf("DONT want token %+v exist", tokenID)
			}
		}
	}
}
