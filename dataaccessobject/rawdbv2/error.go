package rawdbv2

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	// Beacon
	StoreBeaconBlockError = iota
	StoreBeaconBlockIndexError
	GetIndexOfBeaconBlockError
	HasBeaconBlockError
	GetBeaconBlockByHashError
	GetBeaconBlockByIndexError
	DeleteBeaconBlockError
	StoreBeaconBestStateError
	FetchBeaconBestStateError

	// Shard
	StoreShardBlockError
	StoreShardBlockIndexError
	HasShardBlockError
	GetShardBlockByHashError
	GetShardBlockByIndexError
	DeleteShardBlockError
	StoreCrossShardNextHeightError
	FetchCrossShardNextHeightError
	GetIndexOfBlockError
	StoreShardBestStateError
	StoreFeeEstimatorError
	GetFeeEstimatorError
	// tx
	StoreTransactionIndexError
	GetTransactionByHashError
	StoreTxByPublicKeyError
	GetTxByPublicKeyError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
	HasBeaconBlockError:        {-1000, "Has Beacon Block Error"},
	GetBeaconBlockByHashError:  {-1001, "Get Beacon Block By Hash Error"},
	GetBeaconBlockByIndexError: {-1002, "Get Beacon Block By Index Error"},
	DeleteBeaconBlockError:     {-1003, "Delete Beacon Block Error"},
	StoreBeaconBlockIndexError: {-1004, "Store Beacon Block Index Error"},
	GetIndexOfBeaconBlockError: {-1005, "Get Index Of Beacon Block Error"},
	StoreBeaconBestStateError:  {-1006, "Store Beacon Best State Error"},
	FetchBeaconBestStateError:  {-1007, "Fetch Beacon Best State Error"},

	StoreShardBlockError:           {-2000, "Store Shard Block Error"},
	HasShardBlockError:             {-2001, "Has Shard Block Error"},
	GetShardBlockByHashError:       {-2002, "Get Shard Block By Hash Error"},
	GetShardBlockByIndexError:      {-2003, "Get Shard Block By Index Error"},
	DeleteShardBlockError:          {-2004, "Delete Shard Block Error"},
	StoreCrossShardNextHeightError: {-2005, "Store Cross Shard Next Height Error"},
	FetchCrossShardNextHeightError: {-2006, "Fetch Cross Shard Next Height Error"},
	StoreShardBlockIndexError:      {-2007, "Store Shard Block Index Error"},
	GetIndexOfBlockError:           {-2008, "Get Index Of Shard Block Error"},
	StoreShardBestStateError:       {-2009, "Store Shard Best State Error"},
	StoreFeeEstimatorError:         {-2010, "Store Fee Estimator Error"},
	GetFeeEstimatorError:           {-2011, "Get Fee Estimator Error"},

	StoreTransactionIndexError: {-3000, "Store Transaction Index Error"},
	GetTransactionByHashError:  {-3001, "Get Transaction By Hash Error"},
	StoreTxByPublicKeyError:    {-3002, "Store Tx By PublicKey Error"},
	GetTxByPublicKeyError:      {-3003, "Get Tx By Public Key Error"},
}

type RawdbError struct {
	err     error
	Code    int
	Message string
}

func (e RawdbError) GetErrorCode() int {
	return e.Code
}
func (e RawdbError) GetError() error {
	return e.err
}
func (e RawdbError) GetMessage() string {
	return e.Message
}

func (e RawdbError) Error() string {
	return fmt.Sprintf("%d: %+v", e.Code, e.err)
}

func NewRawdbError(key int, err error, params ...interface{}) *RawdbError {
	return &RawdbError{
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].message, params),
	}
}
