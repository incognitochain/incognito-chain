package metadata_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommonMocks "github.com/incognitochain/incognito-chain/metadata/common/mocks"
	coinMocks "github.com/incognitochain/incognito-chain/privacy/coin/mocks"
)

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
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{metadata.StopAutoStakingMeta},
				},
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

func TestStopAutoStakingMetadata_ValidateMetadataByItself(t *testing.T) {
	type fields struct {
		MetadataBaseWithSignature metadata.MetadataBaseWithSignature
		CommitteePublicKey        string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Base58CheckDeserialize error",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "blah",
			},
		},
		{
			name: "CheckSanityData error",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "1hm766APBSXcyDbNbPLbb65Hm2DkK35RJp1cwYx95mFExK3VAkE9qfzDJLTKTMiKbscm4zns5QuDpGS4yc5Hi994G1BVVE2hdLgoNJbvxXdbmsRdrwVCENVYJhYk2k1kci7b8ysb9nFXW8fUEJNsBtfQjtXQY7pEqngbwpEFuF45Kj8skjDriKp2Sc9TjxnPw4478dN4h4XYojPaiSo3sJpqJWDfcZ68DqSWuUAud5REAqeBT3sUiyJCpnfZ9Lp2Uk7M7Pc9CeuTZBVfV3M669zpPdErUgWf7VDYe5wujvcMLhqqjvJRe5WREYLjVni1H1d4qhcuzdbPdW8BC4b7xY2qRSBtiFav8tJt7iSdycTeTTsaYN1",
			},
		},
		{
			name: "stopAutoStakingMetadata error",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.ShardStakingMeta,
					},
				},
				CommitteePublicKey: "blah"},
		},
		{
			name: "happy case",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stopAutoStakingMetadata := &metadata.StopAutoStakingMetadata{
				MetadataBaseWithSignature: tt.fields.MetadataBaseWithSignature,
				CommitteePublicKey:        tt.fields.CommitteePublicKey,
			}
			if got := stopAutoStakingMetadata.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStopAutoStakingMetadata_ValidateSanityData(t *testing.T) {
	type fields struct {
		MetadataBaseWithSignature metadata.MetadataBaseWithSignature
		CommitteePublicKey        string
	}
	type args struct {
		chainRetriever  metadata.ChainRetriever
		shardRetriever  metadata.ShardViewRetriever
		beaconRetriever metadata.BeaconViewRetriever
		beaconHeight    uint64
		tx              metadata.Transaction
	}

	burnCoin := &coinMocks.Coin{}
	burnCoin.On("GetValue").Return(uint64(0))

	txIsPrivacyError := &metadataCommonMocks.Transaction{}
	txIsPrivacyError.On("IsPrivacy").Return(true)
	txIsPrivacyError.On("GetTxBurnData").Return(true, burnCoin, &common.PRVCoinID, nil)

	txGetUniqueReceiverError := &metadataCommonMocks.Transaction{}
	txGetUniqueReceiverError.On("IsPrivacy").Return(false)
	txGetUniqueReceiverError.On("GetUniqueReceiver").Return(false, []byte{}, uint64(0))
	txGetUniqueReceiverError.On("GetTxBurnData").Return(true, burnCoin, &common.PRVCoinID, nil)

	chainBase58CheckDeserializeError := &metadataCommonMocks.ChainRetriever{}
	chainBase58CheckDeserializeError.On("GetBurningAddress", uint64(0)).Return("15pABFiJVeh9D5uiipQxBdSVibGGbdAVipQxBdxkmDqAJaoG1EdFKHBrNfs")
	txBase58CheckDeserializeError := &metadataCommonMocks.Transaction{}
	txBase58CheckDeserializeError.On("IsPrivacy").Return(false)
	txBase58CheckDeserializeError.On("GetUniqueReceiver").Return(true, []byte{}, uint64(0))
	txBase58CheckDeserializeError.On("GetTxBurnData").Return(true, burnCoin, &common.PRVCoinID, nil)

	chainBurningAddressPublicKeyError := &metadataCommonMocks.ChainRetriever{}
	chainBurningAddressPublicKeyError.On("GetBurningAddress", uint64(0)).Return("12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA")
	txBurningAddressPublicKeyError := &metadataCommonMocks.Transaction{}
	txBurningAddressPublicKeyError.On("IsPrivacy").Return(false)
	txBurningAddressPublicKeyError.On("GetUniqueReceiver").Return(true, []byte{99, 183, 246, 161, 68, 172, 228, 222, 153, 9, 172, 39, 208, 245, 167, 79, 11, 2, 114, 65, 241, 69, 85, 40, 193, 104, 199, 79, 70, 4, 53, 0}, uint64(0))
	txBurningAddressPublicKeyError.On("GetTxBurnData").Return(true, burnCoin, &common.PRVCoinID, nil)

	chainStopAutoStakingMetadataError := &metadataCommonMocks.ChainRetriever{}
	chainStopAutoStakingMetadataError.On("GetBurningAddress", uint64(0)).Return("12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA")
	txStopAutoStakingMetadataError := &metadataCommonMocks.Transaction{}
	txStopAutoStakingMetadataError.On("IsPrivacy").Return(false)
	txStopAutoStakingMetadataError.On("GetUniqueReceiver").Return(true, []byte{127, 76, 149, 36, 97, 166, 59, 24, 204, 39, 108, 209, 42, 199, 106, 173, 88, 95, 221, 184, 142, 215, 198, 51, 10, 150, 125, 89, 73, 86, 24, 0}, uint64(0))
	txStopAutoStakingMetadataError.On("GetTxBurnData").Return(true, burnCoin, &common.PRVCoinID, nil)

	txStopAutoStakingMetadataError1 := &metadataCommonMocks.Transaction{}
	txStopAutoStakingMetadataError1.On("IsPrivacy").Return(false)
	txStopAutoStakingMetadataError1.On("GetUniqueReceiver").Return(true, []byte{127, 76, 149, 36, 97, 166, 59, 24, 204, 39, 108, 209, 42, 199, 106, 173, 88, 95, 221, 184, 142, 215, 198, 51, 10, 150, 125, 89, 73, 86, 24, 0}, uint64(1))
	txStopAutoStakingMetadataError1.On("GetTxBurnData").Return(true, burnCoin, &common.PRVCoinID, nil)

	chainHappyCase := &metadataCommonMocks.ChainRetriever{}
	chainHappyCase.On("GetBurningAddress", uint64(0)).Return("15pABFiJVeh9D5uiQEhQX4SVibGGbdAVipQxBdxkmDqAJaoG1EdFKHBrNfs")
	txHappyCase := &metadataCommonMocks.Transaction{}
	txHappyCase.On("IsPrivacy").Return(false)
	txHappyCase.On("GetUniqueReceiver").Return(true, []byte{99, 183, 246, 161, 68, 172, 228, 222, 153, 9, 172, 39, 208, 245, 167, 79, 11, 2, 114, 65, 241, 69, 85, 40, 193, 104, 199, 79, 70, 4, 53, 0}, uint64(0))
	txHappyCase.On("GetTxBurnData").Return(true, burnCoin, &common.PRVCoinID, nil)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		want1   bool
		wantErr bool
	}{
		{
			name: "tx.IsPrivacy error",
			args: args{
				tx: txIsPrivacyError,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name:   "check tx.GetUniqueReceiver error case",
			fields: fields{},
			args: args{
				tx: txGetUniqueReceiverError,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name:   "check wallet.Base58CheckDeserialize error case",
			fields: fields{},
			args: args{
				tx:             txBase58CheckDeserializeError,
				chainRetriever: chainBase58CheckDeserializeError,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "stopAutoStakingMetadata check burning address error",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
			},
			args: args{
				tx:             txBurningAddressPublicKeyError,
				chainRetriever: chainBurningAddressPublicKeyError,
			},
			wantErr: true,
		},
		{
			name: "stopAutoStakingMetadata type error",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.ShardStakingMeta,
					},
				},
			},
			args: args{
				tx:             txStopAutoStakingMetadataError,
				chainRetriever: chainStopAutoStakingMetadataError,
			},
			wantErr: true,
		},
		{
			name: "stopAutoStakingMetadata amount error",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
			},
			args: args{
				tx:             txStopAutoStakingMetadataError1,
				chainRetriever: chainStopAutoStakingMetadataError,
			},
			wantErr: true,
		},
		{
			name: "CommitteePublicKey.FromString error",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6Afv9xkLDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				chainRetriever: chainHappyCase,
				tx:             txHappyCase,
			},
			wantErr: true,
		},
		{
			name: "CommitteePublicKey.CheckSanityData() error",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "1hm766APBSXcyDbNbPLbb65Hm2DkK35RJp1cwYx95mFExK3VAkE9qfzDJLTKTMiKbscm4zns5QuDpGS4yc5Hi994G1BVVE2hdLgoNJbvxXdbmsRdrwVCENVYJhYk2k1kci7b8ysb9nFXW8fUEJNsBtfQjtXQY7pEqngbwpEFuF45Kj8skjDriKp2Sc9TjxnPw4478dN4h4XYojPaiSo3sJpqJWDfcZ68DqSWuUAud5REAqeBT3sUiyJCpnfZ9Lp2Uk7M7Pc9CeuTZBVfV3M669zpPdErUgWf7VDYe5wujvcMLhqqjvJRe5WREYLjVni1H1d4qhcuzdbPdW8BC4b7xY2qRSBtiFav8tJt7iSdycTeTTsaYN1",
			},
			args: args{
				chainRetriever: chainHappyCase,
				tx:             txHappyCase,
			},
			wantErr: true,
		},
		{
			name: "happy case",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				chainRetriever: chainHappyCase,
				tx:             txHappyCase,
			},
			want:    true,
			want1:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stopAutoStakingMetadata := metadata.StopAutoStakingMetadata{
				MetadataBaseWithSignature: tt.fields.MetadataBaseWithSignature,
				CommitteePublicKey:        tt.fields.CommitteePublicKey,
			}
			got, got1, err := stopAutoStakingMetadata.ValidateSanityData(tt.args.chainRetriever, tt.args.shardRetriever, tt.args.beaconRetriever, tt.args.beaconHeight, tt.args.tx)
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

func TestStopAutoStakingMetadata_ValidateTxWithBlockChain(t *testing.T) {
	type fields struct {
		MetadataBaseWithSignature metadata.MetadataBaseWithSignature
		CommitteePublicKey        string
	}
	type args struct {
		tx                  metadata.Transaction
		chainRetriever      metadata.ChainRetriever
		shardViewRetriever  metadata.ShardViewRetriever
		beaconViewRetriever metadata.BeaconViewRetriever
		shardID             byte
		stateDB             *statedb.StateDB
	}

	var stopStakeTx1Meta metadata.Metadata
	stopStakeTx1Meta = &metadata.StopAutoStakingMetadata{
		MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
			MetadataBase: metadata.MetadataBase{
				metadata.StopAutoStakingMeta,
			},
		},
		CommitteePublicKey: validCommitteePublicKeys[1],
	}
	stopStakeTx1 := &metadataCommonMocks.Transaction{}
	stopStakeTx1.On("GetMetadata").Return(stopStakeTx1Meta)
	stopStakeTx1.On("GetMetadataType").Return(metadata.StopAutoStakingMeta)
	stopStakeTx1.On("GetSender").Return([]byte("12buoC8Nmh8WbPhSAiF1SSNB8AuxTu3QbX3sSUydqod4y9ws3e3"))

	chain1 := &metadataCommonMocks.ChainRetriever{}
	beacon1 := &metadataCommonMocks.BeaconViewRetriever{}
	shard1 := &metadataCommonMocks.ShardViewRetriever{}
	beacon1.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{}, errors.New("get error"))

	var stopStakeTx2Meta metadata.Metadata
	stopStakeTx2Meta = &metadata.StopAutoStakingMetadata{
		MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
			MetadataBase: metadata.MetadataBase{
				metadata.StopAutoStakingMeta,
			},
		},
		CommitteePublicKey: validCommitteePublicKeys[2],
	}
	stopStakeTx2 := &metadataCommonMocks.Transaction{}
	stopStakeTx2.On("GetMetadata").Return(stopStakeTx2Meta)
	stopStakeTx2.On("GetMetadataType").Return(metadata.StopAutoStakingMeta)
	stopStakeTx2.On("GetSender").Return([]byte("12buoC8Nmh8WbPhSAiF1SSNB8AuxTu3QbX3sSUydqod4y9ws3e3"))

	chain2 := &metadataCommonMocks.ChainRetriever{}
	beacon2 := &metadataCommonMocks.BeaconViewRetriever{}
	shard2 := &metadataCommonMocks.ShardViewRetriever{}
	beacon2.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{""}, nil)
	emptyMap := make(map[string]string)
	shard2.On("GetShardStakingTx").Return(emptyMap)

	chain21 := &metadataCommonMocks.ChainRetriever{}
	beacon21 := &metadataCommonMocks.BeaconViewRetriever{}
	shard21 := &metadataCommonMocks.ShardViewRetriever{}
	beacon21.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil)
	beacon21.On("GetStakerInfo", "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc").Return(
		&statedb.ShardStakerInfo{},
		false,
		errors.New("error"),
	)
	chain21.On("GetShardStakingTx", byte(0), uint64(1)).Return(emptyMap, errors.New("get staking tx error"))
	shard21.On("GetShardID").Return(byte(0))
	shard21.On("GetBeaconHeight").Return(uint64(1))
	beacon22 := &metadataCommonMocks.BeaconViewRetriever{}
	beacon22.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil)
	beacon22.On("GetStakerInfo", "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc").Return(
		&statedb.ShardStakerInfo{},
		false,
		nil,
	)

	stakingTxHash, _ := common.Hash{}.NewHashFromStr("9648b1f460d853d878a3b7ab7a926acab5f45c726de0610221f78a95f333c6dc")
	stakerInfo1 := &statedb.ShardStakerInfo{}
	stakerInfo1.SetAutoStaking(true)
	stakerInfo1.SetTxStakingID(*stakingTxHash)
	stakerInfo2 := &statedb.ShardStakerInfo{}
	stakerInfo2.SetAutoStaking(false)
	stakerInfo2.SetTxStakingID(*stakingTxHash)

	chain3 := &metadataCommonMocks.ChainRetriever{}
	beacon3 := &metadataCommonMocks.BeaconViewRetriever{}
	shard3 := &metadataCommonMocks.ShardViewRetriever{}
	beacon3.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil)
	beacon3.On("GetStakerInfo", "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc").Return(
		stakerInfo1,
		true,
		nil,
	)
	stakingMap := make(map[string]string)
	stakingMap["121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"] = "asdasdasdas"
	chain3.On("GetShardStakingTx", byte(0), uint64(1)).Return(stakingMap, nil)
	shard3.On("GetShardID").Return(byte(0))
	shard3.On("GetBeaconHeight").Return(uint64(1))

	chain4 := &metadataCommonMocks.ChainRetriever{}
	beacon4 := &metadataCommonMocks.BeaconViewRetriever{}
	shard4 := &metadataCommonMocks.ShardViewRetriever{}
	beacon4.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil)
	beacon4.On("GetStakerInfo", "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc").Return(
		stakerInfo1,
		true,
		nil,
	)
	stakingMap1 := make(map[string]string)
	stakingMap1["121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"] = "9648b1f460d853d878a3b7ab7a926acab5f45c726de0610221f78a95f333c6dc"
	chain4.On("GetShardStakingTx", byte(0), uint64(1)).Return(stakingMap1, nil)
	shard4.On("GetShardID").Return(byte(0))
	shard4.On("GetBeaconHeight").Return(uint64(1))
	stakingTxHash2, _ := common.Hash{}.NewHashFromStr("9648b1f460d853d878a3b7ab7a926acab5f45c726de0610221f78a95f333c6dc")
	chain4.On("GetTransactionByHash", *stakingTxHash2).Return(byte(0), nil, uint64(0), int(0), nil, errors.New("error"))

	stopStakeTx3 := &metadataCommonMocks.Transaction{}
	stopStakeTx3.On("GetMetadata").Return(stopStakeTx2Meta)
	stopStakeTx3.On("GetMetadataType").Return(metadata.StopAutoStakingMeta)
	stopStakeTx3.On("GetSender").Return([]byte("12buoC8Nmh8WbPhSAiF1SSNB8AuxZu3QbX3sSUydqod4y9ws3e3"))

	var stakeTxMeta metadata.Metadata
	stakeTxMeta = &metadata.StakingMetadata{
		MetadataBase: metadata.MetadataBase{
			metadata.ShardStakingMeta,
		},
		FunderPaymentAddress: validCommitteePublicKeys[2],
	}

	stakeTx := &metadataCommonMocks.Transaction{}
	stakeTx.On("GetMetadata").Return(stakeTxMeta)

	chain5 := &metadataCommonMocks.ChainRetriever{}
	beacon5 := &metadataCommonMocks.BeaconViewRetriever{}
	shard5 := &metadataCommonMocks.ShardViewRetriever{}
	beacon5.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil)
	beacon5.On("GetStakerInfo", "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc").Return(
		stakerInfo1,
		true,
		nil,
	)
	chain5.On("GetShardStakingTx", byte(0), uint64(1)).Return(stakingMap1, nil)
	shard5.On("GetShardID").Return(byte(0))
	shard5.On("GetBeaconHeight").Return(uint64(1))
	chain5.On("GetTransactionByHash", *stakingTxHash).Return(byte(0), nil, uint64(0), int(0), stakeTx, nil)
	autoStakingList := make(map[string]bool)
	beacon5.On("GetAutoStakingList").Return(autoStakingList)

	stopStakeTx4 := &metadataCommonMocks.Transaction{}
	stopStakeTx4.On("GetSender").Return([]byte("12buoC8Nmh8WbPhSAiF1SSNB8AuxTu3QbX3sSUydqod4y9ws3e3"))
	chain6 := &metadataCommonMocks.ChainRetriever{}
	beacon6 := &metadataCommonMocks.BeaconViewRetriever{}
	shard6 := &metadataCommonMocks.ShardViewRetriever{}
	beacon6.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").
		Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil).
		Once()
	beacon6.On("GetStakerInfo", "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc").Return(
		stakerInfo1,
		true,
		nil,
	)
	chain6.On("GetShardStakingTx", byte(0), uint64(1)).Return(stakingMap1, nil)
	shard6.On("GetShardID").Return(byte(0))
	shard6.On("GetBeaconHeight").Return(uint64(1))
	chain6.On("GetTransactionByHash", *stakingTxHash).Return(byte(0), nil, uint64(0), int(0), stakeTx, nil)
	autoStakingList1 := make(map[string]bool)
	beacon6.On("GetAutoStakingList").Return(autoStakingList1)

	chain7 := &metadataCommonMocks.ChainRetriever{}
	beacon7 := &metadataCommonMocks.BeaconViewRetriever{}
	shard7 := &metadataCommonMocks.ShardViewRetriever{}
	beacon7.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").
		Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil).
		Once()
	beacon7.On("GetStakerInfo", "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc").Return(
		stakerInfo1,
		true,
		nil,
	)
	chain7.On("GetShardStakingTx", byte(0), uint64(1)).Return(stakingMap1, nil)
	shard7.On("GetShardID").Return(byte(0))
	shard7.On("GetBeaconHeight").Return(uint64(1))
	chain7.On("GetTransactionByHash", *stakingTxHash).Return(byte(0), nil, uint64(0), int(0), stakeTx, nil)
	autoStakingList2 := make(map[string]bool)
	autoStakingList2["121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"] = false
	beacon7.On("GetAutoStakingList").Return(autoStakingList2)

	chain8 := &metadataCommonMocks.ChainRetriever{}
	beacon8 := &metadataCommonMocks.BeaconViewRetriever{}
	shard8 := &metadataCommonMocks.ShardViewRetriever{}
	beacon8.On("GetAllCommitteeValidatorCandidateFlattenListFromDatabase").
		Return([]string{"121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"}, nil).
		Once()
	beacon8.On("GetStakerInfo", "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc").Return(
		stakerInfo1,
		true,
		nil,
	)
	chain8.On("GetShardStakingTx", byte(0), uint64(1)).Return(stakingMap1, nil)
	shard8.On("GetShardID").Return(byte(0))
	shard8.On("GetBeaconHeight").Return(uint64(1))
	chain8.On("GetTransactionByHash", *stakingTxHash).Return(byte(0), nil, uint64(0), int(0), stakeTx, nil)
	autoStakingList3 := make(map[string]bool)
	autoStakingList3["121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc"] = true
	beacon8.On("GetAutoStakingList").Return(autoStakingList3)

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
				tx:                  stopStakeTx1,
				chainRetriever:      chain1,
				beaconViewRetriever: beacon1,
				shardViewRetriever:  shard1,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "check common.IndexOfStr error case",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "",
			},
			args: args{
				tx:                  stopStakeTx2,
				chainRetriever:      chain2,
				beaconViewRetriever: beacon2,
				shardViewRetriever:  shard2,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Get Staker Info error",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				tx:                  stopStakeTx1,
				chainRetriever:      chain21,
				beaconViewRetriever: beacon21,
				shardViewRetriever:  shard21,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Not have staker info",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				tx:                  stopStakeTx1,
				chainRetriever:      chain21,
				beaconViewRetriever: beacon22,
				shardViewRetriever:  shard21,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "check chainRetriever.GetTransactionByHash() error case",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				tx:                  stopStakeTx1,
				chainRetriever:      chain4,
				beaconViewRetriever: beacon4,
				shardViewRetriever:  shard4,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "check bytes.Equal(stakingTx.GetSender(), tx.GetSender()) error case",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				tx:                  stopStakeTx1,
				chainRetriever:      chain5,
				beaconViewRetriever: beacon5,
				shardViewRetriever:  shard5,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "check autoStakingList[stopStakingMetadata.CommitteePublicKey] error case",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				tx:                  stopStakeTx1,
				chainRetriever:      chain6,
				beaconViewRetriever: beacon6,
				shardViewRetriever:  shard6,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "check !isAutoStaking error case",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				tx:                  stopStakeTx1,
				chainRetriever:      chain7,
				beaconViewRetriever: beacon7,
				shardViewRetriever:  shard7,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "happy case",
			fields: fields{
				MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
					MetadataBase: metadata.MetadataBase{
						metadata.StopAutoStakingMeta,
					},
				},
				CommitteePublicKey: "121VhftSAygpEJZ6i9jGkCFHRkD4yhxxccAqVjQTWR9gy7skM1KcNf3uGLpX1NvojmHqs9bWwsPfvyBmer39YNBPwBHpgXg1Qku4EDhtUBZnGw2PZGMF7DMCrYa27GNS97uA9WC5z55YuCDA4WsnKfoEEuCFDNUN3iSCeUyrQ4SF5smx9CwBYX6AWAMAvNDPKf4tCuc7Wiafv9xkLKuHSFr7jaxBfg4rdaxtwXzR5eMpFDDpiXz6hQmdcee8xSXQRKceiafg9RMiuqLxDzx9tmLKvBD5TJq4G76LB3rrVmsYwMo1fY4RZLpiYn6AstAfca5EVnMeexueSAE5sam3Lsq8mq5poJfsW6KXzAbsmFPSsSjhmQ4wGhSXoKSap331gBMuuy7KtmVwQAPpwuFPo9hi7RBgrrn1ssdCdjYSwE226Ekc",
			},
			args: args{
				tx:                  stopStakeTx1,
				chainRetriever:      chain8,
				beaconViewRetriever: beacon8,
				shardViewRetriever:  shard8,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stopAutoStakingMetadata := metadata.StopAutoStakingMetadata{
				MetadataBaseWithSignature: tt.fields.MetadataBaseWithSignature,
				CommitteePublicKey:        tt.fields.CommitteePublicKey,
			}
			got, err := stopAutoStakingMetadata.ValidateTxWithBlockChain(tt.args.tx, tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.shardID, tt.args.stateDB)
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
