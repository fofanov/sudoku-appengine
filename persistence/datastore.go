package persistence

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"

	"github.com/fofanov/sudoku-appengine/sudoku"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

const sudokuGridEntity = "SudokuGrid"

// DatastoreGrid is the interface for storing and loading a sudoku grid.
type DatastoreGrid interface {
	SaveGrid(c appengine.Context, g sudoku.Grid) error
	LoadGrid(c appengine.Context) (sudoku.Grid, error)
}

type binaryDatastoreGrid struct{}

// NewBinaryDatastoreGrid returns a datastore that saves to and loads from
// sudoku grids in a binary encoding.
func NewBinaryDatastoreGrid() DatastoreGrid {
	return &binaryDatastoreGrid{}
}

// Save a sudoku grid as binary stream in the AE datastore.
func (d *binaryDatastoreGrid) SaveGrid(c appengine.Context, g sudoku.Grid) error {

	u := user.Current(c)
	if u == nil {
		return errors.New("No user")
	}

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(&g)
	if err != nil {
		log.Println(err)
		return err
	}

	key := datastore.NewKey(c, sudokuGridEntity, u.ID, 0, nil)
	if _, err = datastore.Put(c, key, &struct{ BS []byte }{BS: buf.Bytes()}); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// Load a sudoku grid from binary stream in the AE datastore.
func (d *binaryDatastoreGrid) LoadGrid(c appengine.Context) (sudoku.Grid, error) {

	u := user.Current(c)
	if u == nil {
		return nil, errors.New("No user")
	}

	key := datastore.NewKey(c, sudokuGridEntity, u.ID, 0, nil)

	bs := &struct{ BS []byte }{}
	if err := datastore.Get(c, key, bs); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, nil
		}
		return nil, err
	}

	grid := sudoku.EmptyGrid()
	if err := gob.NewDecoder(bytes.NewReader(bs.BS)).Decode(&grid); err != nil {
		log.Println(err)
		return nil, err
	}

	return grid, nil
}
