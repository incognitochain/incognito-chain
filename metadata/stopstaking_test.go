package metadata_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/metadata/mocks"
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
		{
			name: "check StakingType error",
			args: args{
				stopStakingType: metadata.BeaconStakingMeta,
			},
			wantErr: true,
		},
		{
			name: "check StakingType success",
			args: args{
				stopStakingType:    metadata.StopAutoStakingMeta,
				committeePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			want: &metadata.StopAutoStakingMetadata{
				MetadataBase:       metadata.MetadataBase{metadata.StopAutoStakingMeta},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			wantErr: false,
		},
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
		{
			name: "Base58CheckDeserialize error",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					metadata.StopAutoStakingMeta,
				},
				CommitteePublicKey: "blah",
			},
		},
		{
			name: "CheckSanityData error",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					metadata.StopAutoStakingMeta,
				},
				CommitteePublicKey: "1hm766APBSXcyDbNbPLbb65Hm2DkK35RJp1cwYx95mFExK3VAkE9qfzDJLTKTMiKbscm4zns5QuDpGS4yc5Hi994G1BVVE2hdLgoNJbvxXdbmsRdrwVCENVYJhYk2k1kci7b8ysb9nFXW8fUEJNsBtfQjtXQY7pEqngbwpEFuF45Kj8skjDriKp2Sc9TjxnPw4478dN4h4XYojPaiSo3sJpqJWDfcZ68DqSWuUAud5REAqeBT3sUiyJCpnfZ9Lp2Uk7M7Pc9CeuTZBVfV3M669zpPdErUgWf7VDYe5wujvcMLhqqjvJRe5WREYLjVni1H1d4qhcuzdbPdW8BC4b7xY2qRSBtiFav8tJt7iSdycTeTTsaYN1",
			},
		},
		{
			name: "stopAutoStakingMetadata error",
			fields: fields{MetadataBase: metadata.MetadataBase{
				metadata.ShardStakingMeta,
			},
				CommitteePublicKey: "blah"},
		},
		{
			name: "happy case",
			fields: fields{
				MetadataBase: metadata.MetadataBase{
					metadata.StopAutoStakingMeta,
				},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			want: true,
		},
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

	txStopAutoStakingMetadataError := &mocks.Transaction{}
	txStopAutoStakingMetadataError.On("IsPrivacy").Return(false)
	txStopAutoStakingMetadataError.On("GetUniqueReceiver").Return(false, []byte{}, uint64(0))

	bcrCommitteePublicKeyError := &mocks.BlockchainRetriever{}
	bcrCommitteePublicKeyError.On("GetBurningAddress", uint64(0)).Return("15pABFiJVeh9D5uiQEhQX4SVibGGbdAVipQxBdxkmDqAJaoG1EdFKHBrNfs")
	txCommitteePublicKeyError := &mocks.Transaction{}
	txCommitteePublicKeyError.On("IsPrivacy").Return(false)
	txCommitteePublicKeyError.On("GetUniqueReceiver").Return(true, []byte{99, 183, 246, 161, 68, 172, 228, 222, 153, 9, 172, 39, 208, 245, 167, 79, 11, 2, 114, 65, 241, 69, 85, 40, 193, 104, 199, 79, 70, 4, 53, 0}, uint64(0))

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		want1   bool
		wantErr bool
	}{
		{
			name: "txr.IsPrivacy error",
			args: args{
				txr: txIsPrivacyError,
			},
			want:    false,
			want1:   false,
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
			wantErr: true,
		},
		{
			name: "stopAutoStakingMetadata error",
			fields: fields{
				MetadataBase: metadata.MetadataBase{metadata.ShardStakingMeta},
			},
			args: args{
				txr: txStopAutoStakingMetadataError,
			},
			wantErr: true,
		},
		{
			name: "CommitteePublicKey.FromString error",
			fields: fields{
				MetadataBase:       metadata.MetadataBase{metadata.StopAutoStakingMeta},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6Afv9xkLDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				bcr: bcrCommitteePublicKeyError,
				txr: txCommitteePublicKeyError,
			},
			wantErr: true,
		},
		{
			name: "happy case",
			fields: fields{
				MetadataBase:       metadata.MetadataBase{metadata.StopAutoStakingMeta},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				bcr: bcrCommitteePublicKeyError,
				txr: txCommitteePublicKeyError,
			},
			want:    true,
			want1:   true,
			wantErr: false,
		},
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

	bcr1 := &mocks.BlockchainRetriever{}
	bcr1.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{}, errors.New("get error"))

	stopStakeTx1 := &mocks.Transaction{}
	var stopStakeTx1Meta metadata.Metadata
	stopStakeTx1Meta = &metadata.StopAutoStakingMetadata{
		MetadataBase: metadata.MetadataBase{
			metadata.StopAutoStakingMeta,
		},
		CommitteePublicKey: validCommitteePublicKeys[1],
	}
	stopStakeTx1.On("GetMetadata").Return(stopStakeTx1Meta)
	stopStakeTx1.On("GetMetadataType").Return(metadata.StopAutoStakingMeta)
	stopStakeTx1.On("GetSender").Return([]byte("12buoC8Nmh8WbPhSAiF1SSNB8AuxTu3QbX3sSUydqod4y9ws3e3"))

	bcr2 := &mocks.BlockchainRetriever{}
	bcr2.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil)
	emptyMap := make(map[string]string)
	bcr2.On("GetStakingTx", byte(0)).Return(emptyMap)

	bcr3 := &mocks.BlockchainRetriever{}
	bcr3.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil)
	stakingMap := make(map[string]string)
	stakingMap["121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"] = "asdasdasdas"
	bcr3.On("GetStakingTx", byte(0)).Return(stakingMap)

	bcr4 := &mocks.BlockchainRetriever{}
	bcr4.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil)
	stakingMap1 := make(map[string]string)
	stakingMap1["121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"] = "9648b1f460d853d878a3b7ab7a926acab5f45c726de0610221f78a95f333c6dc"
	bcr4.On("GetStakingTx", byte(0)).Return(stakingMap)
	bcr4.On("GetTransactionByHash").Return(0, nil, 0, nil, errors.New("error"))

	stopStakeTx2 := &mocks.Transaction{}
	stopStakeTx2.On("GetSender").Return([]byte{})
	bcr5 := &mocks.BlockchainRetriever{}
	bcr5.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil)
	bcr5.On("GetStakingTx", byte(0)).Return(stakingMap1)
	stakingTxHash, _ := common.Hash{}.NewHashFromStr("9648b1f460d853d878a3b7ab7a926acab5f45c726de0610221f78a95f333c6dc")
	bcr5.On("GetTransactionByHash", *stakingTxHash).Return(byte(0), nil, int(0), stopStakeTx2, nil)

	stopStakeTx3 := &mocks.Transaction{}
	stopStakeTx3.On("GetSender").Return([]byte("12buoC8Nmh8WbPhSAiF1SSNB8AuxTu3QbX3sSUydqod4y9ws3e3"))
	bcr6 := &mocks.BlockchainRetriever{}
	bcr6.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil)
	bcr6.On("GetStakingTx", byte(0)).Return(stakingMap1)
	bcr6.On("GetTransactionByHash", *stakingTxHash).Return(byte(0), nil, int(0), stopStakeTx3, nil)
	autoStakingList := make(map[string]bool)
	bcr6.On("GetAutoStakingList").Return(autoStakingList)

	bcr7 := &mocks.BlockchainRetriever{}
	bcr7.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil)
	bcr7.On("GetStakingTx", byte(0)).Return(stakingMap1)
	bcr7.On("GetTransactionByHash", *stakingTxHash).Return(byte(0), nil, int(0), stopStakeTx3, nil)
	autoStakingList1 := make(map[string]bool)
	autoStakingList1["121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"] = false
	bcr7.On("GetAutoStakingList").Return(autoStakingList1)

	bcr8 := &mocks.BlockchainRetriever{}
	bcr8.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil)
	bcr8.On("GetStakingTx", byte(0)).Return(stakingMap1)
	bcr8.On("GetTransactionByHash", *stakingTxHash).Return(byte(0), nil, int(0), stopStakeTx3, nil)
	autoStakingList2 := make(map[string]bool)
	autoStakingList2["121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"] = true
	bcr8.On("GetAutoStakingList").Return(autoStakingList2)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:   "check GetAllCommitteeValidatorCandidateFlattenListFromDatabase error case",
			fields: fields{},
			args: args{
				txr: stopStakeTx1,
				bcr: bcr1,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "check common.IndexOfStr error case",
			fields: fields{
				MetadataBase:       metadata.MetadataBase{metadata.StopAutoStakingMeta},
				CommitteePublicKey: "",
			},
			args: args{
				txr: stopStakeTx1,
				bcr: bcr2,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "check stakingTx[requestedPublicKey] error case",
			fields: fields{
				MetadataBase:       metadata.MetadataBase{metadata.StopAutoStakingMeta},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				txr: stopStakeTx1,
				bcr: bcr2,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "check common.Hash{}.NewHashFromStr() error case",
			fields: fields{
				MetadataBase:       metadata.MetadataBase{metadata.StopAutoStakingMeta},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				txr: stopStakeTx1,
				bcr: bcr3,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "check bcr.GetTransactionByHash() error case",
			fields: fields{
				MetadataBase:       metadata.MetadataBase{metadata.StopAutoStakingMeta},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				txr: stopStakeTx1,
				bcr: bcr4,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "check bytes.Equal(stakingTx.GetSender(), txr.GetSender()) error case",
			fields: fields{
				MetadataBase:       metadata.MetadataBase{metadata.StopAutoStakingMeta},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				txr: stopStakeTx1,
				bcr: bcr5,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "check autoStakingList[stopStakingMetadata.CommitteePublicKey] error case",
			fields: fields{
				MetadataBase:       metadata.MetadataBase{metadata.StopAutoStakingMeta},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				txr: stopStakeTx1,
				bcr: bcr6,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "check !isAutoStaking error case",
			fields: fields{
				MetadataBase:       metadata.MetadataBase{metadata.StopAutoStakingMeta},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				txr: stopStakeTx1,
				bcr: bcr7,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "happy case",
			fields: fields{
				MetadataBase:       metadata.MetadataBase{metadata.StopAutoStakingMeta},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				txr: stopStakeTx1,
				bcr: bcr8,
			},
			want:    true,
			wantErr: false,
		},
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
