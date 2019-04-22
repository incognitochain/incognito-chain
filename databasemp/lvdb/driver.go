package lvdb

import (
	"github.com/constant-money/constant-chain/databasemp"
	"github.com/pkg/errors"
)

func init() {
	driver := databasemp.Driver{
		DbType: "leveldbmempool",
		Open:   openDriver,
	}
	if err := databasemp.RegisterDriver(driver); err != nil {
		panic("failed to register db driver")
	}
}

func openDriver(args ...interface{}) (databasemp.DatabaseInterface, error) {
	if len(args) != 1 {
		return nil, errors.New("invalid arguments")
	}
	dbPath, ok := args[0].(string)
	if !ok {
		return nil, errors.New("expected db path")
	}
	return open(dbPath)
}

