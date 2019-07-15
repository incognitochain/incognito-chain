package database

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func Test_RegisterDriver(t *testing.T) {
	driver := Driver{
		DbType: "leveldb",
		Open: func(args ...interface{}) (databaseInterface DatabaseInterface, e error) {
			return nil, nil
		},
	}
	err := RegisterDriver(driver)
	assert.Equal(t, nil, err)
	err = RegisterDriver(driver)
	assert.NotEqual(t, nil, err)
}

func Test_Open(t *testing.T) {
	_, err := Open("leveldb", filepath.Join("./", "test"))
	assert.Equal(t, nil, err)
	_, err = Open("leveldb1", filepath.Join("./", "test"))
	assert.NotEqual(t, nil, err)
}
