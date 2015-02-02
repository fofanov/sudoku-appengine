package persistence

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/fofanov/sudoku-appengine/sudoku"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

const SUDOKU_GRID_ENTITY = "SudokuGrid"

type DatastoreGrid interface {
	SaveGrid(c appengine.Context, g sudoku.Grid) error
	LoadGrid(c appengine.Context) (sudoku.Grid, error)
}

//TODO Store data in binary format
type jsonDatastoreGrid struct{}

func NewJsonDatastoreGrid() DatastoreGrid {
	return &jsonDatastoreGrid{}
}

// Save a sudoku grid as JSON in the AE datastore.
func (d *jsonDatastoreGrid) SaveGrid(c appengine.Context, g sudoku.Grid) error {

	u := user.Current(c)
	if u == nil {
		return errors.New("No user")
	}

	bs, err := json.Marshal(&g)
	if err != nil {
		log.Println(err)
		return err
	}

	key := datastore.NewKey(c, SUDOKU_GRID_ENTITY, u.ID, 0, nil)
	if _, err = datastore.Put(c, key, &struct{ BS []byte }{BS: bs}); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// Load a sudoku grid from JSON in the AE datastore.
func (d *jsonDatastoreGrid) LoadGrid(c appengine.Context) (sudoku.Grid, error) {

	u := user.Current(c)
	if u == nil {
		return nil, errors.New("No user")
	}

	key := datastore.NewKey(c, SUDOKU_GRID_ENTITY, u.ID, 0, nil)

	bs := &struct{ BS []byte }{}
	if err := datastore.Get(c, key, bs); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, nil
		}
		return nil, err
	}

	grid := sudoku.EmptyGrid()
	if err := json.Unmarshal(bs.BS, &grid); err != nil {
		log.Println(err)
		return nil, err
	}

	return grid, nil
}
