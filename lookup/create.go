package lookup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/tidwall/buntdb"
)

type jsonLocation struct {
	Lat  float64 `json:"latitude"`
	Lon  float64 `json:"longitude"`
	Name string  `json:"name"`
}

func create(db *buntdb.DB, r io.Reader) error {
	dec := json.NewDecoder(r)
	_, err := dec.Token()
	if err != nil {
		return err
	}

	err = db.Update(func(tx *buntdb.Tx) error {
		return tx.DeleteAll()
	})
	if err != nil {
		return err
	}

	err = db.Update(func(tx *buntdb.Tx) error {
		for dec.More() {
			var l jsonLocation
			err = dec.Decode(&l)
			if err != nil {
				return err
			}
			_, _, err = tx.Set(
				fmt.Sprintf("store:%s:pos", l.Name),
				fmt.Sprintf("[%f %f]", l.Lon, l.Lat),
				nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func index(db *buntdb.DB) error {
	return db.Update(func(tx *buntdb.Tx) error {
		return tx.CreateSpatialIndex("store",
			"store:*:pos",
			buntdb.IndexRect)
	})
}

func hasKeys(db *buntdb.DB) error {
	return db.View(func(tx *buntdb.Tx) error {
		size, err := tx.Len()
		if err == nil && size < 1 {
			return fmt.Errorf("no keys in db, try 'geoplay -create'?")
		}
		return err
	})
}

// CreateBuntDB is called by the CLI tool -create.
// This removes any existing keys before doing the import.
func CreateBuntDB(dbFile, locationFile string) error {
	f, err := os.Open(locationFile)
	if err != nil {
		return err
	}
	defer f.Close()

	db, err := buntdb.Open(dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	err = create(db, f)
	if err != nil {
		return err
	}

	err = index(db)
	return err
}
