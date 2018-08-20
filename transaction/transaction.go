package transaction

import (
	"github.com/internet-cash/prototype/common"
)

type Transaction interface {
    Hash() (*common.Hash)
    ValidateTransaction() (bool)
}