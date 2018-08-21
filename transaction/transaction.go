package transaction

import (
	"github.com/ninjadotorg/cash-prototype/common"
)

type Transaction interface {
    Hash() (*common.Hash)
    ValidateTransaction() (bool)
}