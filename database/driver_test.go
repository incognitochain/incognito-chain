package database_test

import (
	"testing"

	"github.com/ninjadotorg/constant/database"
	_ "github.com/ninjadotorg/constant/database/lvdb"
)

func TestLevelDBDriver(t *testing.T) {
	_, err := database.Open("leveldb", ".")
	if err != nil {
		t.Fatalf("database.Open %+v", err)
	}
}
