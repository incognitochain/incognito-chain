package lvdb

import (
	"errors"

	"github.com/ninjadotorg/cash-prototype/database"
)

func init() {
	driver := database.Driver{
		DbType: "lvdb",
		Open:   openDriver,
	}
	if err := database.RegisterDriver(driver); err != nil {
		panic("failed to register db driver")
	}
}

func openDriver(args ...interface{}) (database.DB, error) {
	if len(args) != 1 {
		return nil, errors.New("invalid arguments")
	}
	dbPath, ok := args[0].(string)
	if !ok {
		return nil, errors.New("expected db path")
	}
	return open(dbPath)
}
