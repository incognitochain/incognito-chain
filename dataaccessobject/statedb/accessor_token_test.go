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
				tokenID:   testGenerateTokenIDs(1)[0],
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
				tokenID:   testGenerateTokenIDs(1)[0],
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
				tokenID:   testGenerateTokenIDs(1)[0],
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
				tokenID:   testGenerateTokenIDs(1)[0],
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
				tokenID:   testGenerateTokenIDs(1)[0],
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := StorePrivacyToken(tt.args.stateDB, tt.args.tokenID, tt.args.name, tt.args.symbol, tt.args.tokenType, tt.args.mintable, tt.args.amount, tt.args.info, tt.args.txHash); (err != nil) != tt.wantErr {
				t.Errorf("StorePrivacyToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
