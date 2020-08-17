package metadata_test

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"reflect"
	"testing"
)

// TODO: @lam
// TESTCASE
// 1. RETURN FALSE: NOT PASS CONDITION check StakingType
// 2. RETURN TRUE: PASS CONDITION check StakingType
func TestNewStopAutoStakingMetadata(t *testing.T) {
	type args struct {
		stopStakingType    int
		committeePublicKey string
	}
	tests := []struct {
		name    string
		args    args
		want    *metadata.StopAutoStakingMetadata
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := metadata.NewStopAutoStakingMetadata(tt.args.stopStakingType, tt.args.committeePublicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStopAutoStakingMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStopAutoStakingMetadata() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TODO: @lam
// TESTCASE
// 1. RETURN FALSE: NOT PASS CONDITION check Base58CheckDeserialize()
// 2. RETURN FALSE: NOT PASS CONDITION check CheckSanityData()
// 3. RETURN FALSE: NOT PASS CONDITION check (stopAutoStakingMetadata.Type != StopAutoStakingMeta)
// 3. RETURN TRUE: PASS ALL CONDITION
func TestStopAutoStakingMetadata_ValidateMetadataByItself(t *testing.T) {
	type fields struct {
		MetadataBase       metadata.MetadataBase
		CommitteePublicKey string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stopAutoStakingMetadata := &metadata.StopAutoStakingMetadata{
				MetadataBase:       tt.fields.MetadataBase,
				CommitteePublicKey: tt.fields.CommitteePublicKey,
			}
			if got := stopAutoStakingMetadata.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TODO: @lam
// TESTCASE
// 1. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check txr.IsPrivacy
// 2. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check txr.GetUniqueReceiver
// 3. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check wallet.Base58CheckDeserialize
// 4. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check stopAutoStakingMetadata.Type != StopAutoStakingMeta && amount != StopAutoStakingAmount
// 5. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check CommitteePublicKey.FromString() && CommitteePublicKey.CheckSanityData()
// 6. RETURN TRUE,TRUE,NO-ERROR: PASS ALL CONDITION
func TestStopAutoStakingMetadata_ValidateSanityData(t *testing.T) {
	type fields struct {
		MetadataBase       metadata.MetadataBase
		CommitteePublicKey string
	}
	type args struct {
		bcr          metadata.BlockchainRetriever
		txr          metadata.Transaction
		beaconHeight uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		want1   bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stopAutoStakingMetadata := metadata.StopAutoStakingMetadata{
				MetadataBase:       tt.fields.MetadataBase,
				CommitteePublicKey: tt.fields.CommitteePublicKey,
			}
			got, got1, err := stopAutoStakingMetadata.ValidateSanityData(tt.args.bcr, tt.args.txr, tt.args.beaconHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSanityData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateSanityData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ValidateSanityData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

// TODO: @lam
// TESTCASE
// 1. RETURN FALSE,ERROR: NOT PASS CONDITION check GetAllCommitteeValidatorCandidateFlattenListFromDatabase
// 2. RETURN FALSE,ERROR: NOT PASS CONDITION check (common.IndexOfStr(requestedPublicKey, committees) > -1)
// 3. RETURN FALSE,ERROR: NOT PASS CONDITION check stakingTx[requestedPublicKey]
// 4. RETURN FALSE,ERROR: NOT PASS CONDITION check common.Hash{}.NewHashFromStr()
// 5. RETURN FALSE,ERROR: NOT PASS CONDITION check bcr.GetTransactionByHash()
// 6. RETURN FALSE,ERROR: NOT PASS CONDITION check bytes.Equal(stakingTx.GetSender(), txr.GetSender())
// 7. RETURN FALSE,ERROR: NOT PASS CONDITION check autoStakingList[stopStakingMetadata.CommitteePublicKey]
// 8. RETURN FALSE,ERROR: NOT PASS CONDITION check !isAutoStaking
// 9. RETURN TRUE,NO-ERROR: PASS ALL CONDITION
func TestStopAutoStakingMetadata_ValidateTxWithBlockChain(t *testing.T) {
	type fields struct {
		MetadataBase       metadata.MetadataBase
		CommitteePublicKey string
	}
	type args struct {
		txr     metadata.Transaction
		bcr     metadata.BlockchainRetriever
		shardID byte
		stateDB *statedb.StateDB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stopAutoStakingMetadata := metadata.StopAutoStakingMetadata{
				MetadataBase:       tt.fields.MetadataBase,
				CommitteePublicKey: tt.fields.CommitteePublicKey,
			}
			got, err := stopAutoStakingMetadata.ValidateTxWithBlockChain(tt.args.txr, tt.args.bcr, tt.args.shardID, tt.args.stateDB)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTxWithBlockChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateTxWithBlockChain() got = %v, want %v", got, tt.want)
			}
		})
	}
}
