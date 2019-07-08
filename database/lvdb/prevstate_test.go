package lvdb

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/syndtr/goleveldb/leveldb"
)

var runScenarioTests = flag.Bool("scenario", false, "Run the integration tests (in addition to the unit tests)")

var _ = func() (_ struct{}) {
	database.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func openTestDB(testFnName string) (*leveldb.DB, error) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_"+testFnName)
	if err != nil {
		log.Fatalf("failed to create temp dir: %+v", err)
		return nil, err
	}
	testDB, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		log.Fatalf("could not open db path: %s, %+v", dbPath, err)
		return nil, err
	}
	return testDB, nil
}

func Test_getPrevPrefix(t *testing.T) {
	type args struct {
		isBeacon bool
		shardID  byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"beacon", args{true, 0}, []byte{112, 114, 101, 118, 66, 101, 97, 45}},
		{"shard255", args{false, 255}, []byte{112, 114, 101, 118, 83, 104, 100, 45, 255, 45}},
		{"shard0", args{false, 0}, []byte{112, 114, 101, 118, 83, 104, 100, 45, 0, 45}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPrevPrefix(tt.args.isBeacon, tt.args.shardID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPrevPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_db_StorePrevBestState(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		val      []byte
		isBeacon bool
		shardID  byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.StorePrevBestState(tt.args.val, tt.args.isBeacon, tt.args.shardID); (err != nil) != tt.wantErr {
				t.Errorf("db.StorePrevBestState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_FetchPrevBestState(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		isBeacon bool
		shardID  byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			got, err := db.FetchPrevBestState(tt.args.isBeacon, tt.args.shardID)
			if (err != nil) != tt.wantErr {
				t.Errorf("db.FetchPrevBestState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("db.FetchPrevBestState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_db_CleanBackup(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		isBeacon bool
		shardID  byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.CleanBackup(tt.args.isBeacon, tt.args.shardID); (err != nil) != tt.wantErr {
				t.Errorf("db.CleanBackup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_BackupCommitmentsOfPubkey(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		tokenID common.Hash
		shardID byte
		pubkey  []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.BackupCommitmentsOfPubkey(tt.args.tokenID, tt.args.shardID, tt.args.pubkey); (err != nil) != tt.wantErr {
				t.Errorf("db.BackupCommitmentsOfPubkey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_RestoreCommitmentsOfPubkey(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		tokenID     common.Hash
		shardID     byte
		pubkey      []byte
		commitments [][]byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.RestoreCommitmentsOfPubkey(tt.args.tokenID, tt.args.shardID, tt.args.pubkey, tt.args.commitments); (err != nil) != tt.wantErr {
				t.Errorf("db.RestoreCommitmentsOfPubkey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_DeleteOutputCoin(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		tokenID       common.Hash
		publicKey     []byte
		outputCoinArr [][]byte
		shardID       byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.DeleteOutputCoin(tt.args.tokenID, tt.args.publicKey, tt.args.outputCoinArr, tt.args.shardID); (err != nil) != tt.wantErr {
				t.Errorf("db.DeleteOutputCoin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_BackupSerialNumbersLen(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		tokenID common.Hash
		shardID byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.BackupSerialNumbersLen(tt.args.tokenID, tt.args.shardID); (err != nil) != tt.wantErr {
				t.Errorf("db.BackupSerialNumbersLen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_RestoreSerialNumber(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		tokenID       common.Hash
		shardID       byte
		serialNumbers [][]byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.RestoreSerialNumber(tt.args.tokenID, tt.args.shardID, tt.args.serialNumbers); (err != nil) != tt.wantErr {
				t.Errorf("db.RestoreSerialNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_DeleteTransactionIndex(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		txId common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.DeleteTransactionIndex(tt.args.txId); (err != nil) != tt.wantErr {
				t.Errorf("db.DeleteTransactionIndex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_DeleteCustomToken(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.DeleteCustomToken(tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.DeleteCustomToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_DeleteCustomTokenTx(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		tokenID     common.Hash
		txIndex     int32
		shardID     byte
		blockHeight uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.DeleteCustomTokenTx(tt.args.tokenID, tt.args.txIndex, tt.args.shardID, tt.args.blockHeight); (err != nil) != tt.wantErr {
				t.Errorf("db.DeleteCustomTokenTx() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_DeletePrivacyCustomToken(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.DeletePrivacyCustomToken(tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.DeletePrivacyCustomToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_DeletePrivacyCustomTokenTx(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		tokenID     common.Hash
		txIndex     int32
		shardID     byte
		blockHeight uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.DeletePrivacyCustomTokenTx(tt.args.tokenID, tt.args.txIndex, tt.args.shardID, tt.args.blockHeight); (err != nil) != tt.wantErr {
				t.Errorf("db.DeletePrivacyCustomTokenTx() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_DeletePrivacyCustomTokenCrossShard(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.DeletePrivacyCustomTokenCrossShard(tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.DeletePrivacyCustomTokenCrossShard() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_RestoreCrossShardNextHeights(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		fromShard byte
		toShard   byte
		curHeight uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.RestoreCrossShardNextHeights(tt.args.fromShard, tt.args.toShard, tt.args.curHeight); (err != nil) != tt.wantErr {
				t.Errorf("db.RestoreCrossShardNextHeights() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_DeleteCommitteeByEpoch(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		blkEpoch uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.DeleteCommitteeByEpoch(tt.args.blkEpoch); (err != nil) != tt.wantErr {
				t.Errorf("db.DeleteCommitteeByEpoch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_DeleteAcceptedShardToBeacon(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		shardID      byte
		shardBlkHash common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.DeleteAcceptedShardToBeacon(tt.args.shardID, tt.args.shardBlkHash); (err != nil) != tt.wantErr {
				t.Errorf("db.DeleteAcceptedShardToBeacon() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_DeleteIncomingCrossShard(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		shardID      byte
		crossShardID byte
		crossBlkHash common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.DeleteIncomingCrossShard(tt.args.shardID, tt.args.crossShardID, tt.args.crossBlkHash); (err != nil) != tt.wantErr {
				t.Errorf("db.DeleteIncomingCrossShard() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_BackupBridgedTokenByTokenID(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.BackupBridgedTokenByTokenID(tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.BackupBridgedTokenByTokenID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_RestoreBridgedTokenByTokenID(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"NotScenarioTest", fields{}, args{common.Hash{}}, false},
		{"ScenarioTest", fields{}, args{common.Hash{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB, errDB := openTestDB(tt.name)
			if errDB != nil {
				t.Errorf("db.RestoreBridgedTokenByTokenID() error = %v, wantErr %v", errDB, tt.wantErr)
			}
			db := &db{
				lvdb: testDB,
			}
			if *runScenarioTests && (tt.name == "NotScenarioTest") {
				t.Skip(tt.name)
			} else {
				if tt.name != "NotScenarioTest" {
					t.Skipf("%v", tt.name)
				}
				err := db.BackupBridgedTokenByTokenID(common.Hash{})
				if err != nil {
					t.Fatalf("db.BackupBridgedTokenByTokenID() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err := db.RestoreBridgedTokenByTokenID(tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.RestoreBridgedTokenByTokenID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_BackupShardRewardRequest(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		epoch   uint64
		shardID byte
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"NotScenarioTest", fields{}, args{1, 0, common.Hash{}}, false},
		{"ScenarioTest", fields{}, args{1, 0, common.Hash{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB, errDB := openTestDB(tt.name)
			if errDB != nil {
				t.Errorf("db.BackupShardRewardRequest() error = %v, wantErr %v", errDB, tt.wantErr)
			}
			db := &db{
				lvdb: testDB,
			}
			if *runScenarioTests && (tt.name == "NotScenarioTest") {
				t.Skip(tt.name)
			} else {
				if tt.name != "NotScenarioTest" {
					t.Skipf("%v", tt.name)
				}
			}
			if err := db.BackupShardRewardRequest(tt.args.epoch, tt.args.shardID, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.BackupShardRewardRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_BackupCommitteeReward(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		committeeAddress []byte
		tokenID          common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"NotScenarioTest", fields{}, args{[]byte{}, common.Hash{}}, false},
		{"ScenarioTest", fields{}, args{[]byte{}, common.Hash{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB, errDB := openTestDB(tt.name)
			if errDB != nil {
				t.Errorf("db.BackupCommitteeReward() error = %v, wantErr %v", errDB, tt.wantErr)
			}
			db := &db{
				lvdb: testDB,
			}
			if *runScenarioTests && (tt.name == "NotScenarioTest") {
				t.Skip(tt.name)
			} else {
				if tt.name != "NotScenarioTest" {
					t.Skipf("%v", tt.name)
				}
			}
			if err := db.BackupCommitteeReward(tt.args.committeeAddress, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.BackupCommitteeReward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_RestoreShardRewardRequest(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		epoch   uint64
		shardID byte
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"NotScenarioTest", fields{}, args{1, 0, common.Hash{}}, false},
		{"ScenarioTest", fields{}, args{1, 0, common.Hash{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB, errDB := openTestDB(tt.name)
			if errDB != nil {
				t.Errorf("db.RestoreShardRewardRequest() error = %v, wantErr %v", errDB, tt.wantErr)
			}
			db := &db{
				lvdb: testDB,
			}
			if *runScenarioTests && (tt.name == "NotScenarioTest") {
				t.Skip(tt.name)
			} else {
				if tt.name != "NotScenarioTest" {
					t.Skipf("%v", tt.name)
				}
				err := db.BackupShardRewardRequest(1, 0, common.Hash{})
				if err != nil {
					t.Errorf("db.BackupShardRewardRequest() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err := db.RestoreShardRewardRequest(tt.args.epoch, tt.args.shardID, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.RestoreShardRewardRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_RestoreCommitteeReward(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		committeeAddress []byte
		tokenID          common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"NotScenarioTest", fields{}, args{[]byte{}, common.Hash{}}, false},
		{"ScenarioTest", fields{}, args{[]byte{}, common.Hash{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB, errDB := openTestDB(tt.name)
			if errDB != nil {
				t.Errorf("db.RestoreCommitteeReward() error = %v, wantErr %v", errDB, tt.wantErr)
			}
			db := &db{
				lvdb: testDB,
			}
			if *runScenarioTests && (tt.name == "NotScenarioTest") {
				t.Skip(tt.name)
			} else {
				if tt.name != "NotScenarioTest" {
					t.Skipf("%v", tt.name)
				}
				err := db.BackupCommitteeReward([]byte{}, common.Hash{})
				if err != nil {
					t.Errorf("db.BackupCommitteeReward() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			err := db.RestoreCommitteeReward(tt.args.committeeAddress, tt.args.tokenID)

			if (err != nil) != tt.wantErr {
				t.Errorf("db.RestoreCommitteeReward() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}
