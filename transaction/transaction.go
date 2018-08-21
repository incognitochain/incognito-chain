package transaction

import (
	"github.com/ninjadotorg/money-prototype/common"
)

type Transaction interface {
    Hash() (*common.Hash)
    ValidateTransaction() (bool)
    GetType() (string)
}