package dlx

// An implementation of the 'Dancing Links' algorithm (DLX).
// http://www-cs-faculty.stanford.edu/~uno/papers/dancing-color.ps.gz

import (
	"math"
	"math/rand"
)

const foundLimit = 10000

// DataObject is the key data structure used by the Dancing Links implementation
// to quickly search the problem space.
type DataObject struct {
	Left   *DataObject
	Right  *DataObject
	Up     *DataObject
	Down   *DataObject
	Column *DataObject
	//String representation of this input
	Input string
	// Column Data
	Name string
	//Lets assume 2^32-1 is sufficient
	Size uint32
	Root *DataObject
}

// Row is a doubly linked list of DataObjects traversable by moving left or
// right.
type Row *DataObject

// Column is a doubly linked list of DataObjects traversable by moving up or down.
type Column *DataObject

// Root is the entry point in to the Column row.
type Root Column

// AppendToColumn appends a data object to the column.
func AppendToColumn(column Column, data *DataObject) {
	data.Down = column
	data.Up = column.Up
	column.Up = data
	data.Up.Down = data
	data.Column = column
	if data != column {
		column.Size++
	}
}

// AppendToRow appends a data object to the row.
func AppendToRow(row Row, data *DataObject) {
	data.Right = row
	data.Left = row.Left
	row.Left = data
	data.Left.Right = data
}

// InitialiseColumnHeader inserts a column into the list of headers.
func InitialiseColumnHeader(header *DataObject, root Row, name string) {
	AppendToColumn(header, header)
	AppendToRow(root, header)
	header.Size = 0
	header.Name = name
	header.Root = root
}

// Select the first column in the list of headers.
func simpleSelect(r Root) Column {
	return r.Right
}

// Select the column with the smallest size from list of headers.
func minimalSelect(r Root) Column {
	minSize := uint32(math.MaxUint32)
	column := r.Right
	for j := r.Right; j != r; j = j.Right {
		if j.Size < minSize {
			column = j
			minSize = j.Size
		}
	}
	return column
}

// Cover hides a column from the headers list and hide all row inputs that are
// in this column.
func Cover(c Column) {
	c.Right.Left = c.Left
	c.Left.Right = c.Right
	for i := c.Down; i != c; i = i.Down {
		for j := i.Right; j != i; j = j.Right {
			j.Down.Up = j.Up
			j.Up.Down = j.Down
			j.Column.Size--
		}
	}
}

// Uncover reverses a covering, make possible by the doubly linked lists.
func Uncover(c Column) {
	for i := c.Up; i != c; i = i.Up {
		for j := i.Left; j != i; j = j.Left {
			j.Column.Size++
			j.Down.Up = j
			j.Up.Down = j
		}
	}
	c.Right.Left = c
	c.Left.Right = c
}

// State is the interface for storing a problem modelled using DLX.
type State interface {
	// Returns the number of solutions found in the search and a sample solution
	// from the search (if one exists). Must return > 0 if
	// a solution exists.
	Search() (uint32, interface{})
	// Get the root
	GetRoot() Root
	// store the results of a solution. It is up to the implementation to decide
	// what this does.
	StoreSolution(path []Row, output interface{})
}

// SearchDLX performs the DLX search algorithm. We have an additionally
// augmented the search by keeping track of the number of solutions found so far.
func SearchDLX(e State, path []Row, foundSoFar uint32, output interface{}) uint32 {
	if e.GetRoot().Right == e.GetRoot() {
		if rand.Intn(int(foundSoFar+1)) == 0 {
			if output != nil {
				// Store the solution.
				e.StoreSolution(path, output)
			}
		}
		return 1
	}
	// Use minimal select to speed up search
	c := minimalSelect(e.GetRoot())
	// if c.Size == 0, we have a constraint that can not be satisfied
	if c.Size == 0 {
		return 0
	}
	found := uint32(0)
	Cover(c)
	for r := c.Down; r != c; r = r.Down {
		// Append input to search path.
		path = append(path, r)
		for j := r.Right; j != r; j = j.Right {
			Cover(j.Column)
		}
		found += SearchDLX(e, path, found+foundSoFar, output)
		for j := r.Left; j != r; j = j.Left {
			Uncover(j.Column)
		}
		// Reslice to remove input from path.
		path = path[:len(path)-1]
		// Stop once we have found as many solutions as we want.
		// TODO - Make this configurable in the search.
		if found > foundLimit {
			break
		}
	}
	Uncover(c)
	return found
}
