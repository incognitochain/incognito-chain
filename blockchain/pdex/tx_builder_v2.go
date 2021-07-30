package pdex

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"

	"github.com/incognitochain/incognito-chain/privacy"
)

type TxBuilderV2 struct {
}

func (txBuilder *TxBuilderV2) Build(
	metaType int,
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
	featureStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	var err error

	switch metaType {
	case metadataCommon.Pdexv3MintPDEXGenesisMeta:
		if len(inst) == 4 {
			tx, err = buildMintingPDEXTokenGensis(
				inst[2],
				inst[3],
				producerPrivateKey,
				shardID,
				transactionStateDB,
				featureStateDB,
			)
		} else {
			return tx, fmt.Errorf("Length of instruction is invalid expect %v but get %v", 4, len(inst))
		}
	}

	return tx, err
}

func buildMintingPDEXTokenGensis(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
	featureStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	if instStatus != metadataPdexv3.RequestAcceptedChainStatus {
		return nil, fmt.Errorf("Pdex v3 mint PDEX token genesis: Not support status %v", instStatus)
	}

	contentBytes := []byte(contentStr)
	var instContent metadataPdexv3.MintPDEXGenesisContent
	err := json.Unmarshal(contentBytes, &instContent)
	if err != nil {
		Logger.log.Errorf("[buildMintingPDEXTokenGensis]: an error occured while unmarshaling instruction content: %+v", err)
		return nil, nil
	}

	if instContent.ShardID != shardID {
		Logger.log.Errorf("[buildMintingPDEXTokenGensis]: ShardID unexpected expect %v, but got %+v", shardID, instContent.ShardID)
		return nil, nil
	}

	meta := metadataPdexv3.NewPdexv3MintPDEXGenesisResponse(
		metadataCommon.Pdexv3MintPDEXGenesisMeta,
		instContent.MintingPaymentAddress,
		instContent.MintingAmount,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(instContent.MintingPaymentAddress)
	if err != nil {
		Logger.log.Errorf("[buildMintingPDEXTokenGensis]: an error occured while deserializing minting address string: %+v", err)
		return nil, nil
	}
	// in case the returned currency is privacy custom token
	receiver := &privacy.PaymentInfo{
		Amount:         instContent.MintingAmount,
		PaymentAddress: keyWallet.KeySet.PaymentAddress,
	}

	tokenID := common.PDEXCoinID
	txParam := transaction.TxSalaryOutputParams{Amount: receiver.Amount, ReceiverAddress: &receiver.PaymentAddress, TokenID: &tokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, makeMD)
}
