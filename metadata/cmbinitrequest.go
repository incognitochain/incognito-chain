package metadata

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
)

type CMBInitRequest struct {
	MainAccount    privacy.PaymentAddress   // (Offchain) multisig account of CMB, receive deposits
	ReserveAccount privacy.PaymentAddress   // (Offchain) multisig account of CMB, store reserve requirement assets
	Members        []privacy.PaymentAddress // For validating multisig signature

	MetadataBase
}

func NewCMBInitRequest(data map[string]interface{}) *CMBInitRequest {
	keyWalletMainKey, err := wallet.Base58CheckDeserialize(data["MainAccount"].(string))
	if err != nil {
		return nil
	}
	keyWalletReserveKey, err := wallet.Base58CheckDeserialize(data["ReserveAccount"].(string))
	if err != nil {
		return nil
	}
	memberData, ok := data["Members"].([]string)
	if !ok {
		return nil
	}
	members := []privacy.PaymentAddress{}
	for _, m := range memberData {
		keyWalletMemberKey, err := wallet.Base58CheckDeserialize(m)
		if err != nil {
			return nil
		}
		members = append(members, keyWalletMemberKey.KeySet.PaymentAddress)
	}
	result := CMBInitRequest{
		MainAccount:    keyWalletMainKey.KeySet.PaymentAddress,
		ReserveAccount: keyWalletReserveKey.KeySet.PaymentAddress,
		Members:        members,
	}

	result.Type = CMBInitRequestMeta
	return &result
}

func (creq *CMBInitRequest) Hash() *common.Hash {
	record := creq.MainAccount.String()
	record += creq.ReserveAccount.String()
	for _, member := range creq.Members {
		record += member.String()
	}

	// final hash
	record += creq.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (creq *CMBInitRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// TODO(@0xbunyip): check that MainAccount is multisig address and is unique
	// TODO(@0xbunyip); check that ReserveAccount is unique
	return true, nil
}

func (creq *CMBInitRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	// TODO(@0xbunyip)
	return true, true, nil // continue to check for fee
}

func (creq *CMBInitRequest) ValidateMetadataByItself() bool {
	// TODO(@0xbunyip)
	return true
}

func (creq *CMBInitRequest) CalculateSize() uint64 {
	return calculateSize(creq)
}
