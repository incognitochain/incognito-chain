package manager

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

// func SamplingStakingTxs() *statedb.StateDB {

// }

func TestShardManager_BuildTransactionsFromInstructions(t *testing.T) {
	type fields struct {
		ShardDB map[byte]incdb.Database
	}
	type args struct {
		inses              []instruction.Instruction
		txStateDB          *statedb.StateDB
		producerPrivateKey *privacy.PrivateKey
		shardID            byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []metadata.Transaction
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sM := &ShardManager{
				ShardDB: tt.fields.ShardDB,
			}
			if got := sM.BuildTransactionsFromInstructions(tt.args.inses, tt.args.txStateDB, tt.args.producerPrivateKey, tt.args.shardID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShardManager.BuildTransactionsFromInstructions() = %v, want %v", got, tt.want)
			}
		})
	}
}
