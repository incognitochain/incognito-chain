package database
import "fmt"

type DBBackendType string

const (
	LevelDBBackend DBBackendType = "leveldb"
)

type dbCreator func(name string, dir string, cache int, handles int) (DB, error)

var backends = map[DBBackendType]dbCreator{}

func registerDBCreator(backend DBBackendType, creator dbCreator, force bool) {
	_, ok := backends[backend]
	if !force && ok {
		return
	}
	backends[backend] = creator
}

func NewDB(name string, backend DBBackendType, dir string, cache int, handles int) DB {
	db, err := backends[backend](name, dir, cache, handles)
	if err != nil {
		panic(fmt.Sprintf("Error initializing DB: %v", err))
	}
	return db
}
