package blockchain

import (
	"github.com/incognitochain/incognito-chain/metadata"
	"reflect"
	"testing"
)

// TODO: @lam
// TESTCASE
// 1. RETURN 1 STAKING INSTRUCTION, NO-ERROR
//	INPUT: 3 Transaction with Valid Shard Stake Metadata (no duplicate committee publickey)
// 2. RETURN 1 STOP AUTO STAKE INSTRUCTION, NO-ERROR
//	INPUT: 3 Transaction with Valid Stop Auto Stake Metadata (no duplicate committee publickey)
// 3. COMBINE 2 cases above
func TestCreateShardInstructionsFromTransactionAndInstruction(t *testing.T) {
	type args struct {
		transactions []metadata.Transaction
		bc           *BlockChain
		shardID      byte
	}
	tests := []struct {
		name             string
		args             args
		wantInstructions [][]string
		wantErr          bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInstructions, err := CreateShardInstructionsFromTransactionAndInstruction(tt.args.transactions, tt.args.bc, tt.args.shardID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateShardInstructionsFromTransactionAndInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotInstructions, tt.wantInstructions) {
				t.Errorf("CreateShardInstructionsFromTransactionAndInstruction() gotInstructions = %v, want %v", gotInstructions, tt.wantInstructions)
			}
		})
	}
}
