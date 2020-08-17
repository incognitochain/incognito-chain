package metadata_test

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/metadata/mocks"
	"github.com/incognitochain/incognito-chain/trie"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
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
)

// TODO: @lam
// @TESTCASE
// @1. RETURN FALSE: NOT PASS CONDITION check StakingType
// @2. RETURN TRUE: PASS CONDITION check StakingType
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
		// TODO: Add test cases.
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
// @TESTCASE
// @1. RETURN FALSE: NOT PASS CONDITION check Base58CheckDeserialize
// @2. RETURN FALSE: NOT PASS CONDITION check IsInBase58ShortFormat
// @3. RETURN FALSE: NOT PASS CONDITION check CommitteePublicKey.FromString
// @4. RETURN FALSE: NOT PASS CONDITION check CommitteePublicKey.CheckSanityData
// @5. RETURN TRUE: PASS ALL CONDITION
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
		// TODO: Add test cases.
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
// @TESTCASE
// @1. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check txr.IsPrivacy
// @2. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check txr.GetUniqueReceiver
// @3. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check wallet.Base58CheckDeserialize
// @4. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check wallet.bcr.GetStakingAmountShard() && Stake Shard
// @5. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check wallet.bcr.GetStakingAmountShard() * 3 && Stake Beacon
// @6. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check wallet.Base58CheckDeserialize(rewardReceiverPaymentAddress)
// @7. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check wallet.Base58CheckDeserialize(funderPaymentAddress)
// @8. RETURN FALSE,FALSE,ERROR: NOT PASS CONDITION check CommitteePublicKey.FromString
// @9. RETURN TRUE,TRUE,NO-ERROR : PASS ALL CONDITION
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
		})
	}
}

// TODO: @lam
// @TESTCASE
// @1. RETURN FALSE,ERROR: NOT PASS CONDITION check GetAllCommitteeValidatorCandidate
// @2. RETURN FALSE,ERROR: NOT PASS CONDITION check incognitokey.CommitteeBase58KeyListToStruct
// @3. RETURN FALSE,ERROR: len(tempStaker) == 0 after filter with
// @4. RETURN TRUE,NO-ERROR: len(tempStaker) == 1 after filter
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
		})
	}
}
