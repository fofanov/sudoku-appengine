package persistence

import (
	"testing"

	"appengine/aetest"
	"appengine/user"

	"github.com/fofanov/sudoku-appengine/sudoku"
	ch "gopkg.in/check.v1"
)

// Set up environment for tests.
// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { ch.TestingT(t) }

type PersistenceSuite struct {
	con aetest.Context
}

var _ = ch.Suite(&PersistenceSuite{})

// Set up appengine local services.
func (p *PersistenceSuite) SetUpSuite(c *ch.C) {
	con, err := aetest.NewContext(nil)
	c.Assert(err, ch.IsNil)
	p.con = con
}

// Tear down appengine local services.
func (p *PersistenceSuite) TearDownSuite(c *ch.C) {
	p.con.Close()
}

// Verify that the sudoku grid is round-trippable.
func (p *PersistenceSuite) TestPersistenceRoundTrip(c *ch.C) {
	p.con.Login(&user.User{ID: "testuser",
		Email: "testuser@company.com"})

	store := NewBinaryDatastoreGrid()
	// Store a grid (we don't care if it is valid)
	grid := sudoku.GridPerm()

	err := store.SaveGrid(p.con, grid)
	c.Assert(err, ch.IsNil)

	rtGrid, err := store.LoadGrid(p.con)
	c.Assert(err, ch.IsNil)

	c.Assert(grid, ch.DeepEquals, rtGrid)
}

// Verify that saving a empty grid is correct.
func (p *PersistenceSuite) TestSaveNilGrid(c *ch.C) {
	p.con.Login(&user.User{ID: "testuser",
		Email: "testuser@company.com"})

	store := NewBinaryDatastoreGrid()

	err := store.SaveGrid(p.con, nil)
	c.Assert(err, ch.IsNil)
}

// Verify that saving a grid fails when the user is missing from the context.
func (p *PersistenceSuite) TestFailOnMissingUser(c *ch.C) {
	p.con.Login(&user.User{})

	grid := sudoku.EmptyGrid()
	store := NewBinaryDatastoreGrid()

	err := store.SaveGrid(p.con, grid)
	c.Assert(err, ch.NotNil) // Expected error due to missing user

	_, err = store.LoadGrid(p.con)
	c.Assert(err, ch.NotNil) // Expected error due to missing user

}

// Verify that no grid is returned when there is no grid to load.
func (p *PersistenceSuite) TestNoGridToLoad(c *ch.C) {
	p.con.Login(&user.User{ID: "testuser",
		Email: "testuser@company.com"})

	store := NewBinaryDatastoreGrid()

	rtGrid, err := store.LoadGrid(p.con)

	c.Assert(err, ch.IsNil)
	c.Assert(rtGrid, ch.IsNil) // Expected no grid as we have not saved one.

}
