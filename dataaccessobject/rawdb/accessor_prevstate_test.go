package rawdb

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"reflect"
	"testing"
)

func TestBackupBridgedTokenByTokenID(t *testing.T) {
	type args struct {
		db      incdb.Database
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BackupBridgedTokenByTokenID(tt.args.db, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("BackupBridgedTokenByTokenID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackupCommitmentsOfPublicKey(t *testing.T) {
	type args struct {
		db      incdb.Database
		tokenID common.Hash
		shardID byte
		pubkey  []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BackupCommitmentsOfPublicKey(tt.args.db, tt.args.tokenID, tt.args.shardID, tt.args.pubkey); (err != nil) != tt.wantErr {
				t.Errorf("BackupCommitmentsOfPublicKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackupCommitteeReward(t *testing.T) {
	type args struct {
		db               incdb.Database
		committeeAddress []byte
		tokenID          common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BackupCommitteeReward(tt.args.db, tt.args.committeeAddress, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("BackupCommitteeReward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackupSerialNumbersLen(t *testing.T) {
	type args struct {
		db      incdb.Database
		tokenID common.Hash
		shardID byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BackupSerialNumbersLen(tt.args.db, tt.args.tokenID, tt.args.shardID); (err != nil) != tt.wantErr {
				t.Errorf("BackupSerialNumbersLen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackupShardRewardRequest(t *testing.T) {
	type args struct {
		db      incdb.Database
		epoch   uint64
		shardID byte
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BackupShardRewardRequest(tt.args.db, tt.args.epoch, tt.args.shardID, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("BackupShardRewardRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCleanBackup(t *testing.T) {
	type args struct {
		db       incdb.Database
		isBeacon bool
		shardID  byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CleanBackup(tt.args.db, tt.args.isBeacon, tt.args.shardID); (err != nil) != tt.wantErr {
				t.Errorf("CleanBackup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteAcceptedShardToBeacon(t *testing.T) {
	type args struct {
		db           incdb.Database
		shardID      byte
		shardBlkHash common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteAcceptedShardToBeacon(tt.args.db, tt.args.shardID, tt.args.shardBlkHash); (err != nil) != tt.wantErr {
				t.Errorf("DeleteAcceptedShardToBeacon() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteCommitteeByHeight(t *testing.T) {
	type args struct {
		db       incdb.Database
		blkEpoch uint64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteCommitteeByHeight(tt.args.db, tt.args.blkEpoch); (err != nil) != tt.wantErr {
				t.Errorf("DeleteCommitteeByHeight() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteIncomingCrossShard(t *testing.T) {
	type args struct {
		db           incdb.Database
		shardID      byte
		crossShardID byte
		crossBlkHash common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteIncomingCrossShard(tt.args.db, tt.args.shardID, tt.args.crossShardID, tt.args.crossBlkHash); (err != nil) != tt.wantErr {
				t.Errorf("DeleteIncomingCrossShard() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteOutputCoin(t *testing.T) {
	type args struct {
		db            incdb.Database
		tokenID       common.Hash
		publicKey     []byte
		outputCoinArr [][]byte
		shardID       byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteOutputCoin(tt.args.db, tt.args.tokenID, tt.args.publicKey, tt.args.outputCoinArr, tt.args.shardID); (err != nil) != tt.wantErr {
				t.Errorf("DeleteOutputCoin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeletePrivacyToken(t *testing.T) {
	type args struct {
		db      incdb.Database
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeletePrivacyToken(tt.args.db, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("DeletePrivacyToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeletePrivacyTokenCrossShard(t *testing.T) {
	type args struct {
		db      incdb.Database
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeletePrivacyTokenCrossShard(tt.args.db, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("DeletePrivacyTokenCrossShard() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeletePrivacyTokenTx(t *testing.T) {
	type args struct {
		db          incdb.Database
		tokenID     common.Hash
		txIndex     int32
		shardID     byte
		blockHeight uint64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeletePrivacyTokenTx(tt.args.db, tt.args.tokenID, tt.args.txIndex, tt.args.shardID, tt.args.blockHeight); (err != nil) != tt.wantErr {
				t.Errorf("DeletePrivacyTokenTx() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteTransactionIndex(t *testing.T) {
	type args struct {
		db   incdb.Database
		txId common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteTransactionIndex(tt.args.db, tt.args.txId); (err != nil) != tt.wantErr {
				t.Errorf("DeleteTransactionIndex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFetchPrevBestState(t *testing.T) {
	type args struct {
		db       incdb.Database
		isBeacon bool
		shardID  byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchPrevBestState(tt.args.db, tt.args.isBeacon, tt.args.shardID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchPrevBestState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchPrevBestState() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRestoreBridgedTokenByTokenID(t *testing.T) {
	type args struct {
		db      incdb.Database
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RestoreBridgedTokenByTokenID(tt.args.db, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("RestoreBridgedTokenByTokenID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRestoreCommitmentsOfPubkey(t *testing.T) {
	type args struct {
		db          incdb.Database
		tokenID     common.Hash
		shardID     byte
		pubkey      []byte
		commitments [][]byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RestoreCommitmentsOfPubkey(tt.args.db, tt.args.tokenID, tt.args.shardID, tt.args.pubkey, tt.args.commitments); (err != nil) != tt.wantErr {
				t.Errorf("RestoreCommitmentsOfPubkey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRestoreCommitteeReward(t *testing.T) {
	type args struct {
		db               incdb.Database
		committeeAddress []byte
		tokenID          common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RestoreCommitteeReward(tt.args.db, tt.args.committeeAddress, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("RestoreCommitteeReward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRestoreCrossShardNextHeights(t *testing.T) {
	type args struct {
		db        incdb.Database
		fromShard byte
		toShard   byte
		curHeight uint64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RestoreCrossShardNextHeights(tt.args.db, tt.args.fromShard, tt.args.toShard, tt.args.curHeight); (err != nil) != tt.wantErr {
				t.Errorf("RestoreCrossShardNextHeights() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRestoreSerialNumber(t *testing.T) {
	type args struct {
		db            incdb.Database
		tokenID       common.Hash
		shardID       byte
		serialNumbers [][]byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RestoreSerialNumber(tt.args.db, tt.args.tokenID, tt.args.shardID, tt.args.serialNumbers); (err != nil) != tt.wantErr {
				t.Errorf("RestoreSerialNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRestoreShardRewardRequest(t *testing.T) {
	type args struct {
		db      incdb.Database
		epoch   uint64
		shardID byte
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RestoreShardRewardRequest(tt.args.db, tt.args.epoch, tt.args.shardID, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("RestoreShardRewardRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorePrevBestState(t *testing.T) {
	type args struct {
		db       incdb.Database
		val      []byte
		isBeacon bool
		shardID  byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := StorePrevBestState(tt.args.db, tt.args.val, tt.args.isBeacon, tt.args.shardID); (err != nil) != tt.wantErr {
				t.Errorf("StorePrevBestState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
