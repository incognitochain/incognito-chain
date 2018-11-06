package lvdb

import (
	"errors"

	"github.com/ninjadotorg/constant/database"
)

func init() {
	driver := database.Driver{
		DbType: "leveldb",
		Open:   openDriver,
	}
	if err := database.RegisterDriver(driver); err != nil {
		panic("failed to register db driver")
	}
}

func openDriver(args ...interface{}) (database.DatabaseInterface, error) {
	if len(args) != 1 {
		return nil, errors.New("invalid arguments")
	}
	dbPath, ok := args[0].(string)
	if !ok {
		return nil, errors.New("expected db path")
	}
	return open(dbPath)
}
