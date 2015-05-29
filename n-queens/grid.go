package nqueens

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fofanov/sudoku-appengine/dlx"
)

const (
	// Number of constraint types (cell,row,column,subgrid).
	constraintTypes = 4
)

// Grid represents The N-Queens grid. We use int8 as Go's JSON encoding
// confuses it for bytes.
type Grid [][]int8

// EmptyGrid creates an empty N-Queens grid. (0 denotes no input)
func EmptyGrid(n int) Grid {
	rows := make([][]int8, n)
	for i := range rows {
		rows[i] = make([]int8, n)
	}
	return rows
}

// Inputs represents all possible inputs into the n-queens grid.
type Inputs [][][constraintTypes]dlx.DataObject

type nqueensDlxState struct {
	size   int
	inputs *[][][constraintTypes]dlx.DataObject
	root   dlx.Root
}

// Search for solutions to the current state of the grid.
func (s *nqueensDlxState) Search() (uint32, interface{}) {

	solutionGrid := EmptyGrid(s.size)
	sols := dlx.SearchDLX(s, make([]dlx.Row, 0), 0, solutionGrid)
	return sols, solutionGrid
}

// Store a winning grid found in the search.
func (s *nqueensDlxState) StoreSolution(path []dlx.Row, output interface{}) {

	grid := output.(Grid)
	for _, row := range grid {
		for i := range row {
			row[i] = 0
		}
	}
	//TODO Should we assume this type assertion is correct?
	for _, t := range path {
		v := strings.Fields(t.Input)
		v0, _ := strconv.Atoi(v[0])
		v1, _ := strconv.Atoi(v[1])
		grid[v1][v0] = int8(v0 + 1)
	}
	fmt.Println("_________________")
	for _, row := range grid {
		fmt.Println(row)
	}
}

func (s *nqueensDlxState) GetRoot() dlx.Root {
	return s.root
}

func initialiseNQueensDLXState(n int) *nqueensDlxState {

	root := &dlx.DataObject{}
	dlx.InitialiseColumnHeader(root, root, "root")

	// Columns = Constraints
	// Constraint types:
	// *Required in each solution*
	// Constraint1 = Every horizontal must contain exactly one queen => n columns
	// Constraint2 = Every vertical must contain exactly one queen => n columns
	// *Not required in each solution*
	// Constraint3 = Every left diagonal must contain exactly one queen => (2*n - 1) columns
	// Constraint4 = Every right diagonal must contain exactly one queen => (2*n - 1) columns

	// Create 4 * 81 column headers

	horizontalHeaders := make([]dlx.DataObject, n)
	for x := 0; x < n; x++ {
		dlx.InitialiseColumnHeader(&horizontalHeaders[x], root, fmt.Sprintf("horizontal %d", x))
	}

	verticalHeaders := make([]dlx.DataObject, n)
	for x := 0; x < n; x++ {
		dlx.InitialiseColumnHeader(&verticalHeaders[x], root, fmt.Sprintf("vertical %d", x))
	}

	// The diagonal constrains are not requirements of a solution. Hence we
	// don't link them into the root header row.
	leftDiagonalHeaders := make([]dlx.DataObject, 2*n-1)
	for x := 0; x < 2*n-1; x++ {
		dlx.InitialiseColumnHeader(&leftDiagonalHeaders[x], &leftDiagonalHeaders[0], fmt.Sprintf("left diagonal %d", x))
	}

	rightDiagonalHeaders := make([]dlx.DataObject, 2*n-1)
	for x := 0; x < 2*n-1; x++ {
		dlx.InitialiseColumnHeader(&rightDiagonalHeaders[x], &rightDiagonalHeaders[0], fmt.Sprintf("right diagonal %d", x))
	}

	// Rows = Possible inputs.
	// Inputs = a queen cell x,y => n*n rows.
	// Every row fulfills 4 constraints, one of each type.
	// Therefore we need to create 4 * n * n Data Objects to make exact
	// cover.
	// Input = which horizontal, vertical, left/right diagonal constraint it fills => which
	//   implies queen in cell x,y.
	// Left - Right links the four constraints of the input
	//    e.g. x,y horzontal <=> diagonal <=> left diag <=> right diag.
	// Up - Down links => linked to the constraint it fulfills
	//   e.g. (subgrid 3, value 4).

	getLeftDiagonal := func(x, y, n int) int {
		nX, nY := 0, 0
		if x <= y {
			nX, nY = 0, y-x
		} else {
			nX, nY = x-y, 0
		}

		if nX == 0 {
			return n - (nY + 1)
		}

		return n + nX - 1
	}

	getRightDiagonal := func(x, y, n int) int {
		nX, nY := x, y
		for nX < n-1 && nY > 0 {
			nX++
			nY--
		}

		if nX == n-1 {
			return n - (nY + 1)
		}

		return 2*n - (nX + 2)
	}

	rows := make([][][constraintTypes]dlx.DataObject, n)
	for i := range rows {
		rows[i] = make([][constraintTypes]dlx.DataObject, n)
	}
	for x := 0; x < n; x++ {
		for y := 0; y < n; y++ {
			//First link up the constraints
			for c := 0; c < constraintTypes; c++ {
				dlx.AppendToRow(&rows[x][y][0], &rows[x][y][c])
				rows[x][y][c].Input = fmt.Sprintf("%d %d", x, y)
			}
			// Link to horizontal constraint
			dlx.AppendToColumn(&horizontalHeaders[x], &rows[x][y][0])
			// Link to vertical constraint
			dlx.AppendToColumn(&verticalHeaders[y], &rows[x][y][1])
			// Link to left diag constraint
			lDiag := getLeftDiagonal(x, y, n)
			dlx.AppendToColumn(&leftDiagonalHeaders[lDiag], &rows[x][y][2])
			// Link to right diag  constraint
			rDiag := getRightDiagonal(x, y, n)
			dlx.AppendToColumn(&rightDiagonalHeaders[rDiag], &rows[x][y][3])
		}
	}

	return &nqueensDlxState{
		size:   n,
		root:   root,
		inputs: &rows,
	}
}

// NewNQueensState is a DLX state representation of the n queens problem.
func NewNQueensState(n int) dlx.State {
	return initialiseNQueensDLXState(n)
}
