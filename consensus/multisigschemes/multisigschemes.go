package multisigschemes

import (
	"github.com/incognitochain/incognito-chain/common"
)

type MultiSigsSchemeInterface interface {
	Prepare(data interface{}) error
	ValidateAggSig(dataHash []byte, validationField string, validatorPubkeyList []string) error
	ValidateSingleSig(dataHash *common.Hash, sig string, pubkey string) error
	SignData(privKey string, dataHash []byte) string
	AggregateSig() string
}
