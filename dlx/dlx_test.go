package dlx

import (
	"math/rand"
	"testing"

	ch "gopkg.in/check.v1"
)

// Set up environment for tests.
// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { ch.TestingT(t) }

type DlxSuite struct {
}

var _ = ch.Suite(&DlxSuite{})

// Verify that appending to a row is correct.
func (d *DlxSuite) TestAppendToRow(c *ch.C) {

	// Create row
	row := &DataObject{Input: "TestRow"}

	AppendToRow(row, row)
	// Can set up doubly linked row with just 1 object
	c.Assert(row.Left, ch.Equals, row)
	c.Assert(row.Right, ch.Equals, row)

	obsRow := []*DataObject{&DataObject{Input: "Test1"}, &DataObject{Input: "Test2"}, &DataObject{Input: "Test3"},
		&DataObject{Input: "Test4"}, &DataObject{Input: "Test5"}, &DataObject{Input: "Test6"}, &DataObject{Input: "Test7"}}

	for _, obj := range obsRow {
		AppendToRow(row, obj)
	}

	current := row
	prev := row.Left
	// Run down row and check that the objs were added correctly
	for _, obj := range obsRow {
		current = current.Right
		prev = prev.Right
		c.Assert(obj, ch.Equals, current)
		c.Assert(current.Left, ch.Equals, prev)
	}

	// Should be back at the start of the row
	c.Assert(current.Right, ch.Equals, row)
}

// Verify that appending to a column is correct.
func (d *DlxSuite) TestAppendToColumn(c *ch.C) {

	// Create column
	col := &DataObject{Input: "TestCol"}

	AppendToColumn(col, col)
	// Can set up doubly linked column with just 1 object
	c.Assert(col.Up, ch.Equals, col)
	c.Assert(col.Down, ch.Equals, col)

	obsCol := []*DataObject{&DataObject{Input: "Test1"}, &DataObject{Input: "Test2"}, &DataObject{Input: "Test3"},
		&DataObject{Input: "Test4"}, &DataObject{Input: "Test5"}, &DataObject{Input: "Test6"}, &DataObject{Input: "Test7"}}

	for _, obj := range obsCol {
		AppendToColumn(col, obj)
	}

	current := col
	prev := col.Up
	// Run down column and check that the objs were added correctly
	for _, obj := range obsCol {
		current = current.Down
		prev = prev.Down
		c.Assert(obj, ch.Equals, current)
		c.Assert(current.Up, ch.Equals, prev)
	}

	// Should be back at the start of the row
	c.Assert(current.Down, ch.Equals, col)
	// Assert the column header count is correct
	c.Assert(col.Size, ch.Equals, uint32(len(obsCol)))
}

// Verify that initialising a column header is correct.
func (d *DlxSuite) TestInitialiseColumnHeader(c *ch.C) {

	// Create root
	root := &DataObject{Input: "TestRoot"}
	InitialiseColumnHeader(root, root, "TestRoot")

	// Can set up doubly linked  header with just one column
	c.Assert(root.Right, ch.Equals, root)
	c.Assert(root.Left, ch.Equals, root)
	c.Assert(root.Down, ch.Equals, root)
	c.Assert(root.Up, ch.Equals, root)
	c.Assert(root.Root, ch.Equals, root)
	c.Assert(root.Size, ch.Equals, uint32(0))

	obsHead := []*DataObject{&DataObject{Input: "Test1"}, &DataObject{Input: "Test2"}, &DataObject{Input: "Test3"},
		&DataObject{Input: "Test4"}, &DataObject{Input: "Test5"}, &DataObject{Input: "Test6"}, &DataObject{Input: "Test7"}}

	for _, obj := range obsHead {
		InitialiseColumnHeader(obj, root, obj.Name)
	}

	current := root
	prev := root.Left
	// Run through headers and check that the objs were added correctly
	for _, obj := range obsHead {
		current = current.Right
		prev = prev.Right
		c.Assert(obj, ch.Equals, current)
		c.Assert(current.Up, ch.Equals, current)
		c.Assert(current.Down, ch.Equals, current)
		c.Assert(current.Left, ch.Equals, prev)
		c.Assert(current.Size, ch.Equals, uint32(0))
		c.Assert(current.Root, ch.Equals, root)
	}

	// Should be back at the root
	c.Assert(current.Right, ch.Equals, root)
}

func (d *DlxSuite) TestSimpleSelect(c *ch.C) {

	// Create root
	root := &DataObject{Input: "TestRoot"}
	InitialiseColumnHeader(root, root, "TestRoot")

	obsHead := []Column{&DataObject{Input: "Test1"}, &DataObject{Input: "Test2"}, &DataObject{Input: "Test3"},
		&DataObject{Input: "Test4"}, &DataObject{Input: "Test5"}, &DataObject{Input: "Test6"}, &DataObject{Input: "Test7"}}

	for _, obj := range obsHead {
		InitialiseColumnHeader(obj, root, obj.Name)
	}

	col := simpleSelect(root)

	// Should be the first column in the header list
	c.Assert(col, ch.Equals, obsHead[0])
}

// Verify that selecting the minimal column is correct.
func (d *DlxSuite) TestMinimalSelect(c *ch.C) {

	// Create root
	root := &DataObject{Input: "TestRoot"}
	InitialiseColumnHeader(root, root, "TestRoot")

	obsHead := []Column{&DataObject{Input: "Test1"}, &DataObject{Input: "Test2"}, &DataObject{Input: "Test3"},
		&DataObject{Input: "Test4"}, &DataObject{Input: "Test5"}, &DataObject{Input: "Test6"}, &DataObject{Input: "Test7"}}

	for _, obj := range obsHead {
		InitialiseColumnHeader(obj, root, obj.Name)
	}

	// Set some random sizes
	var expected Column = nil
	min := 100
	for _, o := range obsHead {
		v := rand.Intn(100)
		o.Size = uint32(v)
		if v < min {
			min = v
			expected = o
		}
	}

	col := minimalSelect(root)

	// Should be the column with the smallest size
	c.Assert(col, ch.Equals, expected)
}

// Verify that the mechanics of covering and uncovering work as expected.
func (d *DlxSuite) TestCoverUncover(c *ch.C) {

	// Create root
	root := &DataObject{Input: "TestRoot"}
	InitialiseColumnHeader(root, root, "TestRoot")

	obsHead := []*DataObject{&DataObject{Input: "Test1"}, &DataObject{Input: "Test2"}, &DataObject{Input: "Test3"},
		&DataObject{Input: "Test4"}, &DataObject{Input: "Test5"}, &DataObject{Input: "Test6"}, &DataObject{Input: "Test7"}}

	// Add column headers
	for _, obj := range obsHead {
		InitialiseColumnHeader(obj, root, obj.Name)
	}

	// Create an array of inputs and hook them into the headers
	var inputs [10][3]DataObject
	for i := 0; i < 10; i++ {
		var used [7]bool
		for j := 0; j < 3; j++ {
			AppendToRow(&inputs[i][0], &inputs[i][j])
			// Append to a random column not chosen already in this row
			k := rand.Intn(7)
			for ; used[k]; k = rand.Intn(7) {
			}
			used[k] = true
			AppendToColumn(obsHead[k], &inputs[i][j])
		}
	}

	// Store the column sizes of each header for future verification purposes
	var expectedHeaderSize [7]uint32
	for i := 0; i < 7; i++ {
		expectedHeaderSize[i] = obsHead[i].Size
	}

	// Randomly choose a column, cover them, then uncover and check back to
	// same state

	for i := 0; i < 5; i++ {
		col := rand.Intn(7)
		// Cover column
		Cover(obsHead[col])
		for r := obsHead[col].Down; r != obsHead[col]; r = r.Down {
			// Cover rest of input
			for j := r.Right; j != r; j = j.Right {
				Cover(j.Column)
			}
			// Now uncover in reverse order
			for j := r.Left; j != r; j = j.Left {
				Uncover(j.Column)
			}
		}
		Uncover(obsHead[col])

		current := root
		//Verify that everything that no columns are missing and that they have
		//the expected size
		for i, obj := range obsHead {
			current = current.Right
			c.Assert(obj, ch.Equals, current)
			c.Assert(obj.Size, ch.Equals, expectedHeaderSize[i])
		}
	}
}
