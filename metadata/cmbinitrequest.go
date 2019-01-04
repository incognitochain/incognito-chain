package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/wallet"
)

type CMBInitRequest struct {
	MainAccount    privacy.PaymentAddress   // (Offchain) multisig account of CMB, receive deposits
	ReserveAccount privacy.PaymentAddress   // (Offchain) multisig account of CMB, store reserve requirement assets
	Members        []privacy.PaymentAddress // For validating multisig signature

	MetadataBase
}

func NewCMBInitRequest(data map[string]interface{}) *CMBInitRequest {
	mainKey, err := wallet.Base58CheckDeserialize(data["MainAccount"].(string))
	if err != nil {
		return nil
	}
	reserveKey, err := wallet.Base58CheckDeserialize(data["ReserveAccount"].(string))
	if err != nil {
		return nil
	}
	memberData, ok := data["Members"].([]string)
	if !ok {
		return nil
	}
	members := []privacy.PaymentAddress{}
	for _, m := range memberData {
		memberKey, err := wallet.Base58CheckDeserialize(m)
		if err != nil {
			return nil
		}
		members = append(members, memberKey.KeySet.PaymentAddress)
	}
	result := CMBInitRequest{
		MainAccount:    mainKey.KeySet.PaymentAddress,
		ReserveAccount: reserveKey.KeySet.PaymentAddress,
		Members:        members,
	}

	result.Type = CMBInitRequestMeta
	return &result
}

func (creq *CMBInitRequest) Hash() *common.Hash {
	record := string(creq.MainAccount.Bytes())
	record += string(creq.ReserveAccount.Bytes())
	for _, member := range creq.Members {
		record += string(member.Bytes())
	}

	// final hash
	record += string(creq.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (creq *CMBInitRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// TODO(@0xbunyip): check that MainAccount is multisig address and is unique
	// TODO(@0xbunyip); check that ReserveAccount is unique
	return true, nil
}

func (creq *CMBInitRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return true, true, nil // continue to check for fee
}

func (creq *CMBInitRequest) ValidateMetadataByItself() bool {
	return true
}
