package lookup

import (
	"bytes"
	"io"
	"os"

	"github.com/jamessanford/geoplay/data"
	"github.com/tidwall/buntdb"
)

// DB contains the underlying buntdb for Geo lookups,
// and in the future may not even be buntdb.
type DB struct {
	db *buntdb.DB
}

const memoryFile = ":memory:"

func populateFromFileOrBuiltin(db *buntdb.DB, locationFile string) error {
	var r io.Reader

	if _, err := os.Stat(locationFile); os.IsNotExist(err) {
		a, err := data.Asset("locations.json")
		if err != nil {
			return err
		}
		r = bytes.NewReader(a)
	} else {
		f, err := os.Open(locationFile)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}

	return create(db, r)
}

// TryOpenDB opens dbFile if it exists, otherwise it reads locationFile.
// If locationFile also does not exist, a builtin dataset is used.
//
// This complexity is so that the command line server example can be run
// with no required files or flags.
//
// Caller must call Close() on DB when done.
func TryOpenDB(dbFile, locationFile string) (*DB, error) {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		dbFile = memoryFile // use in-mem db
	}
	db, err := buntdb.Open(dbFile)
	if err == nil && dbFile == memoryFile {
		err = populateFromFileOrBuiltin(db, locationFile)
	}
	if err == nil {
		err = hasKeys(db)
	}
	if err == nil {
		err = index(db)
	}
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// Close closes the underlying DB.  Must call when done with your *DB.
func (lu *DB) Close() error {
	return lu.db.Close()
}
