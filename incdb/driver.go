package incdb

import "github.com/pkg/errors"

// Driver defines a structure for backend drivers to use when they registered
// themselves as a backend which implements the Database interface.
type Driver struct {
	DbType string
	Open   func(args ...interface{}) (Database, error)
}

var drivers = make(map[string]*Driver)

// RegisterDriver registers the driver d.
func RegisterDriver(d Driver) error {
	if _, exists := drivers[d.DbType]; exists {
		return NewDatabaseError(DriverExistErr, errors.Errorf("Driver %s is already registered", d.DbType))
	}
	drivers[d.DbType] = &d
	return nil
}

// Open opens the db connection.
func Open(typ string, args ...interface{}) (Database, error) {
	d, exists := drivers[typ]
	if !exists {
		return nil, NewDatabaseError(DriverNotRegisterErr, errors.Errorf("Driver %s is not registered", typ))
	}
	return d.Open(args...)
}
