package tx_ver2

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

func (txToken *TxToken) LoadCommitment(*statedb.StateDB) error {
	return nil
}

func (txToken TxToken) ValidateDoubleSpendWithBlockChain(*statedb.StateDB) (bool, error) {
	return true, nil
}

func (txToken TxToken) ValidateSanityDataByItSelf() (bool, error) {
	return true, nil
}

func (txToken *TxToken) ValidateSanityDataWithBlockchain(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, uint64) (bool, error) {
	return true, nil
}

func (txToken *TxToken) ValidateTxCorrectness() (bool, error) {
	return true, nil
}

func (txToken *TxToken) VerifySigTx() (bool, error) {
	return true, nil
}

func (txToken *TxToken) initEnv() metadata.ValidationEnviroment {

	valEnv := tx_generic.DefaultValEnv()
	// if txCustomTokenPrivacy.IsSalaryTx() {
	valEnv = tx_generic.WithAct(valEnv, common.TxActTranfer)
	// }
	if txToken.IsPrivacy() {
		valEnv = tx_generic.WithPrivacy(valEnv)
	} else {
		valEnv = tx_generic.WithNoPrivacy(valEnv)
	}

	valEnv = tx_generic.WithType(valEnv, txToken.GetType())
	sID := common.GetShardIDFromLastByte(txToken.GetSenderAddrLastByte())
	valEnv = tx_generic.WithShardID(valEnv, int(sID))
	txToken.SetValidationEnv(valEnv)
	txNormalValEnv := valEnv.Clone()
	if txToken.GetTxTokenData().Type == utils.CustomTokenInit {
		txNormalValEnv = tx_generic.WithAct(txNormalValEnv, common.TxActInit)
	} else {
		txNormalValEnv = tx_generic.WithAct(txNormalValEnv, common.TxActTranfer)
	}
	if txToken.GetTxTokenData().TxNormal.IsPrivacy() {
		txNormalValEnv = tx_generic.WithPrivacy(txNormalValEnv)
	} else {
		txNormalValEnv = tx_generic.WithNoPrivacy(txNormalValEnv)
	}
	txToken.GetTxTokenData().TxNormal.SetValidationEnv(txNormalValEnv)
	return valEnv
}

func (txToken *TxToken) GetValidationEnv() metadata.ValidationEnviroment {
	return txToken.Tx.GetValidationEnv()
}

func (txToken *TxToken) SetValidationEnv(valEnv metadata.ValidationEnviroment) {
	txToken.Tx.SetValidationEnv(valEnv)
}
