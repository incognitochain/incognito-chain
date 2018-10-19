package database_test

import (
	"testing"

	"github.com/ninjadotorg/cash/database"
	_ "github.com/ninjadotorg/cash/database/lvdb"
)

func TestLevelDBDriver(t *testing.T) {
	_, err := database.Open("leveldb", ".")
	if err != nil {
		t.Fatalf("database.Open %+v", err)
	}
}
