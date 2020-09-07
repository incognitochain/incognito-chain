package metadata_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/metadata/mocks"
	"github.com/incognitochain/incognito-chain/trie"
)

var (
	wrarperDB statedb.DatabaseAccessWarper
	diskDB    incdb.Database
)

var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_metadata")
	if err != nil {
		panic(err)
	}
	diskDB, _ = incdb.Open("leveldb", dbPath)
	wrarperDB = statedb.NewDatabaseAccessWarper(diskDB)
	emptyStateDB, _ = statedb.NewWithPrefixTrie(common.EmptyRoot, wrarperDB)
	validCommitteePublicKeyStructs, _ = incognitokey.CommitteeBase58KeyListToStruct(validCommitteePublicKeys)
	metadata.Logger.Init(common.NewBackend(nil).Logger("test", true))
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

var (
	emptyStateDB     *statedb.StateDB
	validPrivateKeys = []string{
		"112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
		"112t8rrEW3NPNgU8xzbeqE7cr4WTT8JvyaQqSZyczA5hBJVvpQMTBVqNfcCdzhvquWCHH11jHihZtgyJqbdWPhWYbmmsw5aV29WSXBEsgbVX",
	}
	validCommitteePublicKeys = []string{
		"121VhftSAygpEJZ6i9jGk4fj81FpWVTwe3wWDzRZjzdjaQXk9QtGbwNWNwjt3p8zi3p2LRug8m78TDeq4LCAiQT2shDLSrK9sSHBX4DrNgnqsRbkEazrnWapvs7F5CMTPj5kT859WHJV26Wm1P8hwHXpxLwbeMM9n2kJXznTgRJGzdBZ4iY2CTF28s7ADyknqcBJ1RBfEUT9GVeixKC3AKDAna2QqQfdcdFiJaps5PixjJznk7CcTgcYgfPcnysdUgRuygAcbDikvw35KF9jzmeTZWZtbXhbXePhyPP8MuaGwDY75hCiDn1iDEvNHBGMqKJtENq8mfkQTW9GrGu2kkDBmNsmDVannjsbxUuoHU9MT5hYftTcsvyVi4s2S73JbGDNnWD7e3cVwXF8rgYGMFNyYBm3qWB3jobBkGwTPNh5Tpb7",
		"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
	}
	validCommitteePublicKeyStructs = []incognitokey.CommitteePublicKey{}
	validPaymentAddresses          = []string{
		"12S42qYc9pzsfWoxPZ21sVihEHJxYfNzEp1SXNnxvr7CGYMHNWX12ZaQkzcwvTYKAnhiVsDWwSqz5jFo6xuwzXZmz7QX1TnJaWnwEyX",
		"12RrjUWjyCNPXoCChrpEVLxucs3WEw9KyFxzP3UrdRzped2UouDzBM9gNugySqt4RpmgkqL1H7xxE8PfNmDwAatnSXPUVdNomBK1yYC",
	}
	validPublicKeys = []string{
		"12buoC8Nmh8WbPhSAiF1SSNB8AuxTu3QbX3sSUydqod4y9ws3e3",
		"1HXXH7MxWGQgg2QZP854WjDYtebEiKDwPjJzFqBpTUJ447GEG2",
	}
	validPrivateSeeds = []string{
		"129pZpqYqYAA8wTAeDKuVwRthoBjNLUFm8FnLwUTkXddUqwShN9",
		"12JqKehM24bfSkfv3FKGtzFw4seoJSJbbgAqaYtX3w6DjVuH8mb",
	}
	invalidCommitteePublicKeys = []string{"121VhftSAygpEJZ6i9jGk4fj81FpWVTwe3wWDzRZjzdjaQXk9QtGbwNWNwjt3p8zi3p2LRug8m78TDeq4LCAiQT2shDLSrK9sSHBX4DrNgnqsRbkEazrnWapvs7F5CMTPj5kT859WHJV26Wm1P8hwHXpxLwbeMM9n2kJXznTgRJGzdBZ4iY2CTF28s7ADyknqcBJ1RBfEUT9GVeixKC3AKDAnaeixKC3AdFiJaps5PixjJznk7CcTgcYgfPcnysdUgRuygAcbDikvw35KF9jzmeTZWZtbXhbXePhyPP8MuaGwDY75hCiDn1iDEvNHBGMqKJtENq8mfkQTW9GrGu2kkDBmNsmDVannjsbxUuoHU9MT5hYftTcsvyVi4s2S73JbGDNnWD7e3cVwXF8rgYGMFNyYBm3qWB3jobBkGwTPNh5Tpb7",
		"1hm766APBSXcyDbNbPLbb65Hm2DkK35RJp1cwYx95mFExK3VAkE9qfzDJLTKTMiKbscm4zns5QuDpGS4yc5Hi994G1BVVE2hdLgoNJbvxXdbmsRdrwVCENVYJhYk2k1kci7b8ysb9nFXW8fUEJNsBtfQjtXQY7pEqngbwpEFuF45Kj8skjDriKp2Sc9TjxnPw4478dN4h4XYojPaiSo3sJpqJWDfcZ68DqSWuUAud5REAqeBT3sUiyJCpnfZ9Lp2Uk7M7Pc9CeuTZBVfV3M669zpPdErUgWf7VDYe5wujvcMLhqqjvJRe5WREYLjVni1H1d4qhcuzdbPdW8BC4b7xY2qRSBtiFav8tJt7iSdycTeTTsaYN1"}
	invalidPaymentAddresses = []string{
		"12S42qYc9pzsfWoxPZ21sVih7CGYMHNWX12SXNnxvr7CGYMHNWX12ZaQkzcwvTYKAnhiVsDWwSqz5jFo6xuwzXZmz7QX1TnJaWnwEyX"}
)

// TODO: @lam
// TESTCASE
// 1. RETURN FALSE: NOT PASS CONDITION check StakingType
// 2. RETURN TRUE: PASS CONDITION check StakingType
func TestNewStakingMetadata(t *testing.T) {
	type args struct {
		stakingType                  int
		funderPaymentAddress         string
		rewardReceiverPaymentAddress string
		stakingAmountShard           uint64
		committeePublicKey           string
		autoReStaking                bool
	}
	tests := []struct {
		name    string
		args    args
		want    *metadata.StakingMetadata
		wantErr bool
	}{
		{
			name: "check StakingType error case",
			args: args{
				stakingType:                  65,
				funderPaymentAddress:         validPaymentAddresses[0],
				rewardReceiverPaymentAddress: validPaymentAddresses[0],
				stakingAmountShard:           1750000000000,
				committeePublicKey:           validCommitteePublicKeys[0],
				autoReStaking:                false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "check StakingType success case",
			args: args{
				stakingType:                  63,
				funderPaymentAddress:         validPaymentAddresses[0],
				rewardReceiverPaymentAddress: validPaymentAddresses[0],
				stakingAmountShard:           1750000000000,
				committeePublicKey:           validCommitteePublicKeys[0],
				autoReStaking:                false,
			},
			want:    &metadata.StakingMetadata{metadata.MetadataBase{63}, "12S42qYc9pzsfWoxPZ21sVihEHJxYfNzEp1SXNnxvr7CGYMHNWX12ZaQkzcwvTYKAnhiVsDWwSqz5jFo6xuwzXZmz7QX1TnJaWnwEyX", "12S42qYc9pzsfWoxPZ21sVihEHJxYfNzEp1SXNnxvr7CGYMHNWX12ZaQkzcwvTYKAnhiVsDWwSqz5jFo6xuwzXZmz7QX1TnJaWnwEyX", 1750000000000, false, "121VhftSAygpEJZ6i9jGk4fj81FpWVTwe3wWDzRZjzdjaQXk9QtGbwNWNwjt3p8zi3p2LRug8m78TDeq4LCAiQT2shDLSrK9sSHBX4DrNgnqsRbkEazrnWapvs7F5CMTPj5kT859WHJV26Wm1P8hwHXpxLwbeMM9n2kJXznTgRJGzdBZ4iY2CTF28s7ADyknqcBJ1RBfEUT9GVeixKC3AKDAna2QqQfdcdFiJaps5PixjJznk7CcTgcYgfPcnysdUgRuygAcbDikvw35KF9jzmeTZWZtbXhbXePhyPP8MuaGwDY75hCiDn1iDEvNHBGMqKJtENq8mfkQTW9GrGu2kkDBmNsmDVannjsbxUuoHU9MT5hYftTcsvyVi4s2S73JbGDNnWD7e3cVwXF8rgYGMFNyYBm3qWB3jobBkGwTPNh5Tpb7"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := metadata.NewStakingMetadata(tt.args.stakingType, tt.args.funderPaymentAddress, tt.args.rewardReceiverPaymentAddress, tt.args.stakingAmountShard, tt.args.committeePublicKey, tt.args.autoReStaking)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStakingMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStakingMetadata() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TODO: @lam
// TESTCASE
// 1. RETURN FALSE: NOT PASS CONDITION check Base58CheckDeserialize
// 2. RETURN FALSE: NOT PASS CONDITION check IsInBase58ShortFormat
// 3. RETURN FALSE: NOT PASS CONDITION check CommitteePublicKey.FromString
// 4. RETURN FALSE: NOT PASS CONDITION check CommitteePublicKey.CheckSanityData
// 5. RETURN TRUE: PASS ALL CONDITION
func TestStakingMetadata_ValidateMetadataByItself(t *testing.T) {
	type fields struct {
		MetadataBase                 metadata.MetadataBase
		FunderPaymentAddress         string
		RewardReceiverPaymentAddress string
		StakingAmountShard           uint64
		AutoReStaking                bool
		CommitteePublicKey           string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "check Base58CheckDeserialize error case",
			fields: fields{
				MetadataBase:                 metadata.MetadataBase{63},
				FunderPaymentAddress:         validPaymentAddresses[0],
				RewardReceiverPaymentAddress: invalidPaymentAddresses[0],
				StakingAmountShard:           1750000000000,
				AutoReStaking:                false,
				CommitteePublicKey:           validCommitteePublicKeys[0],
			},
			want: false,
		},
		{
			name: "check IsInBase58ShortFormat error case",
			fields: fields{
				MetadataBase:                 metadata.MetadataBase{63},
				FunderPaymentAddress:         validPaymentAddresses[0],
				RewardReceiverPaymentAddress: validPaymentAddresses[0],
				StakingAmountShard:           1750000000000,
				AutoReStaking:                false,
				CommitteePublicKey:           invalidCommitteePublicKeys[0],
			},
			want: false,
		},
		{
			name: "check CommitteePublicKey.FromString error case",
			fields: fields{
				MetadataBase:                 metadata.MetadataBase{63},
				FunderPaymentAddress:         validPaymentAddresses[0],
				RewardReceiverPaymentAddress: validPaymentAddresses[0],
				StakingAmountShard:           1750000000000,
				AutoReStaking:                false,
				CommitteePublicKey:           invalidCommitteePublicKeys[0],
			},
			want: false,
		},
		{
			name: "check CommitteePublicKey.CheckSanityData error case",
			fields: fields{
				MetadataBase:                 metadata.MetadataBase{63},
				FunderPaymentAddress:         validPaymentAddresses[0],
				RewardReceiverPaymentAddress: validPaymentAddresses[0],
				StakingAmountShard:           1750000000000,
				AutoReStaking:                false,
				CommitteePublicKey:           invalidCommitteePublicKeys[1],
			},
			want: false,
		},
		{
			name: "happy case",
			fields: fields{
				MetadataBase:                 metadata.MetadataBase{63},
				FunderPaymentAddress:         validPaymentAddresses[0],
				RewardReceiverPaymentAddress: validPaymentAddresses[0],
				StakingAmountShard:           1750000000000,
				AutoReStaking:                false,
				CommitteePublicKey:           validCommitteePublicKeys[0],
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &metadata.StakingMetadata{
				MetadataBase:                 tt.fields.MetadataBase,
				FunderPaymentAddress:         tt.fields.FunderPaymentAddress,
				RewardReceiverPaymentAddress: tt.fields.RewardReceiverPaymentAddress,
				StakingAmountShard:           tt.fields.StakingAmountShard,
				AutoReStaking:                tt.fields.AutoReStaking,
				CommitteePublicKey:           tt.fields.CommitteePublicKey,
			}
			if got := sm.ValidateMetadataByItself(); got != tt.want {
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
// 4. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check wallet.bcr.GetStakingAmountShard() && Stake Shard
// 5. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check wallet.bcr.GetStakingAmountShard() * 3 && Stake Beacon
// 6. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check wallet.Base58CheckDeserialize(rewardReceiverPaymentAddress)
// 7. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check wallet.Base58CheckDeserialize(funderPaymentAddress)
// 8. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check CommitteePublicKey.FromString
// 9. RETURN TRUE,TRUE,NO-ERROR : PASS ALL CONDITION
func TestStakingMetadata_ValidateSanityData(t *testing.T) {
	type fields struct {
		MetadataBase                 metadata.MetadataBase
		FunderPaymentAddress         string
		RewardReceiverPaymentAddress string
		StakingAmountShard           uint64
		AutoReStaking                bool
		CommitteePublicKey           string
	}
	type args struct {
		bcr          metadata.BlockchainRetriever
		txr          metadata.Transaction
		beaconHeight uint64
	}

	txIsPrivacyError := &mocks.Transaction{}
	txIsPrivacyError.On("IsPrivacy").Return(true)

	txGetUniqueReceiverError := &mocks.Transaction{}
	txGetUniqueReceiverError.On("IsPrivacy").Return(false)
	txGetUniqueReceiverError.On("GetUniqueReceiver").Return(false, []byte{}, uint64(0))

	bcrBase58CheckDeserializeError := &mocks.BlockchainRetriever{}
	bcrBase58CheckDeserializeError.On("GetBurningAddress", uint64(0)).Return("15pABFiJVeh9D5uiipQxBdSVibGGbdAVipQxBdxkmDqAJaoG1EdFKHBrNfs")
	txBase58CheckDeserializeError := &mocks.Transaction{}
	txBase58CheckDeserializeError.On("IsPrivacy").Return(false)
	txBase58CheckDeserializeError.On("GetUniqueReceiver").Return(true, []byte{}, uint64(0))

	bcrGetStakingAmountShardError := &mocks.BlockchainRetriever{}
	bcrGetStakingAmountShardError.On("GetBurningAddress", uint64(0)).Return("15pABFiJVeh9D5uiipQxBdSVibGGbdAVipQxBdxkmDqAJaoG1EdFKHBrNfs")
	txGetStakingAmountShardError := &mocks.Transaction{}
	txGetStakingAmountShardError.On("IsPrivacy").Return(false)
	txGetStakingAmountShardError.On("GetUniqueReceiver").Return(true, []byte{}, uint64(0))

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		want1   bool
		want2   error
		wantErr bool
	}{
		{
			name:   "check txr.IsPrivacy error case",
			fields: fields{},
			args: args{
				txr: txIsPrivacyError,
			},
			want:    false,
			want1:   false,
			want2:   errors.New("staking Transaction Is No Privacy Transaction"),
			wantErr: true,
		},
		{
			name:   "check txr.GetUniqueReceiver error case",
			fields: fields{},
			args: args{
				txr: txGetUniqueReceiverError,
			},
			want:    false,
			want1:   false,
			want2:   errors.New("staking Transaction Should Have 1 Output Amount crossponding to 1 Receiver"),
			wantErr: true,
		},
		{
			name:   "check wallet.Base58CheckDeserialize error case",
			fields: fields{},
			args: args{
				txr: txBase58CheckDeserializeError,
				bcr: bcrBase58CheckDeserializeError,
			},
			want:    false,
			want1:   false,
			want2:   errors.New("burning address is invalid"),
			wantErr: true,
		},
		{
			name: "check wallet.bcr.GetStakingAmountShard() && Stake Shard error case",
			fields: fields{
				MetadataBase:       metadata.MetadataBase{63},
				StakingAmountShard: 1650000000000,
			},
			args: args{
				txr: txGetStakingAmountShardError,
				bcr: bcrGetStakingAmountShardError,
			},
			want:    false,
			want1:   false,
			want2:   errors.New("invalid Stake Shard Amount"),
			wantErr: true,
		},
		{
			name: "check wallet.bcr.GetStakingAmountShard() * 3 && Stake Beacon error case",
			fields: fields{
				MetadataBase:       metadata.MetadataBase{64},
				StakingAmountShard: 1750000000000,
			},
			args: args{
				txr: txGetStakingAmountShardError,
				bcr: bcrGetStakingAmountShardError},
			want:    false,
			want1:   false,
			want2:   errors.New("invalid Stake Beacon Amount"),
			wantErr: true,
		},
		{
			name:    "check wallet.Base58CheckDeserialize(funderPaymentAddress) error case",
			fields:  fields{},
			args:    args{},
			want:    false,
			want1:   false,
			want2:   errors.New("Invalid Funder Payment Address, Failed to Deserialized Into Key Wallet"),
			wantErr: true,
		},
		{
			name:    "check CommitteePublicKey.FromString error case",
			fields:  fields{},
			args:    args{},
			want:    false,
			want1:   false,
			want2:   nil,
			wantErr: true,
		},
		{
			name:    "happy case",
			fields:  fields{},
			args:    args{},
			want:    true,
			want1:   true,
			want2:   nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stakingMetadata := metadata.StakingMetadata{
				MetadataBase:                 tt.fields.MetadataBase,
				FunderPaymentAddress:         tt.fields.FunderPaymentAddress,
				RewardReceiverPaymentAddress: tt.fields.RewardReceiverPaymentAddress,
				StakingAmountShard:           tt.fields.StakingAmountShard,
				AutoReStaking:                tt.fields.AutoReStaking,
				CommitteePublicKey:           tt.fields.CommitteePublicKey,
			}
			got, got1, err := stakingMetadata.ValidateSanityData(tt.args.bcr, tt.args.txr, tt.args.beaconHeight)
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
			if err.Error() != tt.want2.Error() {
				t.Errorf("ValidateSanityData() err = %v, want %v", err, tt.want2)
			}
		})
	}
}

// TODO: @lam
// TESTCASE
// 1. RETURN FALSE,ERROR: NOT PASS CONDITION check GetAllCommitteeValidatorCandidate
// 2. RETURN FALSE,ERROR: NOT PASS CONDITION check incognitokey.CommitteeBase58KeyListToStruct
// 3. RETURN FALSE,ERROR: len(tempStaker) == 0 after filter with
// 4. RETURN TRUE,NO-ERROR: len(tempStaker) == 1 after filter
func TestStakingMetadata_ValidateTxWithBlockChain(t *testing.T) {
	SC := make(map[byte][]incognitokey.CommitteePublicKey)
	SPV := make(map[byte][]incognitokey.CommitteePublicKey)
	happyCaseBlockChainRetriever := &mocks.BlockchainRetriever{}
	happyCaseBlockChainRetriever.On("GetAllCommitteeValidatorCandidate").
		Return(SC, SPV, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{},
			nil)
	stakeAlreadyBlockChainRetriever := &mocks.BlockchainRetriever{}
	stakeAlreadyBlockChainRetriever.On("GetAllCommitteeValidatorCandidate").
		Return(SC, SPV, []incognitokey.CommitteePublicKey{validCommitteePublicKeyStructs[0]}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{},
			nil)
	getCommitteeErrorBlockChainRetriever := &mocks.BlockchainRetriever{}
	getCommitteeErrorBlockChainRetriever.On("GetAllCommitteeValidatorCandidate").
		Return(SC, SPV, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{},
			errors.New("get committee error"))

	type fields struct {
		MetadataBase                 metadata.MetadataBase
		FunderPaymentAddress         string
		RewardReceiverPaymentAddress string
		StakingAmountShard           uint64
		AutoReStaking                bool
		CommitteePublicKey           string
	}
	type args struct {
		txr     metadata.Transaction
		bcr     metadata.BlockchainRetriever
		b       byte
		stateDB *statedb.StateDB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "happy case",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					metadata.ShardStakingMeta,
				},
				FunderPaymentAddress:         validPaymentAddresses[0],
				RewardReceiverPaymentAddress: validPaymentAddresses[0],
				StakingAmountShard:           1750000000000,
				AutoReStaking:                false,
				CommitteePublicKey:           validCommitteePublicKeys[0],
			},
			args: args{
				txr:     &mocks.Transaction{},
				bcr:     happyCaseBlockChainRetriever,
				b:       0,
				stateDB: emptyStateDB,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "stake already error case",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					metadata.ShardStakingMeta,
				},
				FunderPaymentAddress:         validPaymentAddresses[0],
				RewardReceiverPaymentAddress: validPaymentAddresses[0],
				StakingAmountShard:           1750000000000,
				AutoReStaking:                false,
				CommitteePublicKey:           validCommitteePublicKeys[0],
			},
			args: args{
				txr:     &mocks.Transaction{},
				bcr:     stakeAlreadyBlockChainRetriever,
				b:       0,
				stateDB: emptyStateDB,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "get committee error case",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					metadata.ShardStakingMeta,
				},
				FunderPaymentAddress:         validPaymentAddresses[0],
				RewardReceiverPaymentAddress: validPaymentAddresses[0],
				StakingAmountShard:           1750000000000,
				AutoReStaking:                false,
				CommitteePublicKey:           validCommitteePublicKeys[0],
			},
			args: args{
				txr:     &mocks.Transaction{},
				bcr:     getCommitteeErrorBlockChainRetriever,
				b:       0,
				stateDB: emptyStateDB,
			},
			want:    false,
			wantErr: true,
		},

		{
			name: "CommitteeBase58KeyListToStruct error case",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					metadata.ShardStakingMeta,
				},
				FunderPaymentAddress:         validPaymentAddresses[0],
				RewardReceiverPaymentAddress: validPaymentAddresses[0],
				StakingAmountShard:           1750000000000,
				AutoReStaking:                false,
				CommitteePublicKey:           invalidCommitteePublicKeys[0],
			},
			args: args{
				txr:     &mocks.Transaction{},
				bcr:     happyCaseBlockChainRetriever,
				b:       0,
				stateDB: emptyStateDB,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stakingMetadata := metadata.StakingMetadata{
				MetadataBase:                 tt.fields.MetadataBase,
				FunderPaymentAddress:         tt.fields.FunderPaymentAddress,
				RewardReceiverPaymentAddress: tt.fields.RewardReceiverPaymentAddress,
				StakingAmountShard:           tt.fields.StakingAmountShard,
				AutoReStaking:                tt.fields.AutoReStaking,
				CommitteePublicKey:           tt.fields.CommitteePublicKey,
			}
			got, err := stakingMetadata.ValidateTxWithBlockChain(tt.args.txr, tt.args.bcr, tt.args.b, tt.args.stateDB)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTxWithBlockChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateTxWithBlockChain() got = %v, want %v", got, tt.want)
			}
			fmt.Println(err)
		})
	}
}
