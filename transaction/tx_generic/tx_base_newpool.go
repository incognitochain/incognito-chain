package tx_generic //nolint:revive

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (tx *TxBase) initEnv() metadata.ValidationEnviroment {
	valEnv := DefaultValEnv()
	if tx.IsSalaryTx() {
		valEnv = WithAct(valEnv, common.TxActInit)
	}
	if tx.IsPrivacy() {
		valEnv = WithPrivacy(valEnv)
	} else {
		valEnv = WithNoPrivacy(valEnv)
	}
	valEnv = WithType(valEnv, tx.GetType())
	sID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	valEnv = WithShardID(valEnv, int(sID))
	tx.SetValidationEnv(valEnv)
	return valEnv
}

func (tx *TxBase) GetValidationEnv() metadata.ValidationEnviroment {
	return tx.valEnv
}

func (tx *TxBase) SetValidationEnv(vEnv metadata.ValidationEnviroment) {
	if vE, ok := vEnv.(*ValidationEnv); ok {
		tx.valEnv = vE
	} else {
		valEnv := DefaultValEnv()
		if tx.IsPrivacy() {
			valEnv = WithPrivacy(valEnv)
		} else {
			valEnv = WithNoPrivacy(valEnv)
		}
		valEnv = WithType(valEnv, tx.GetType())
		tx.valEnv = valEnv
	}
}
