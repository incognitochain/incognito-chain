package database_test

import (
	"testing"

	"github.com/ninjadotorg/cash-prototype/database"
	_ "github.com/ninjadotorg/cash-prototype/database/lvdb"
)

func TestLevelDBDriver(t *testing.T) {
	_, err := database.Open("leveldb", ".")
	if err != nil {
		t.Fatalf("database.Open %+v", err)
	}
}
