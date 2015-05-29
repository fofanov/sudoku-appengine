package sudoku

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/fofanov/sudoku-appengine/dlx"
)

const (
	// Number of constraint types (cell,row,column,subgrid).
	constraintTypes = 4
	// Size of subgrid.
	sqN = 3
	// GridSize is the number of values in the grid
	// Also the size of the grid.
	GridSize = sqN * sqN
)

// Inputs represents all possible inputs into the Sudoku grid.
type Inputs [GridSize][GridSize][GridSize][constraintTypes]dlx.DataObject

// Grid represents the sudoku grid. We use int8 as Go's JSON encoding confuses
// it for bytes. Valid input values are [1..N]. 0 indicates no input.
type Grid [][]int8

type sudokuDLXState struct {
	inputs *Inputs
	root   dlx.Root
}

// Input represents a number at a position in the Sudoku grid
type Input struct {
	X uint8
	Y uint8
	N int8
}

// EmptyGrid creates an empty grid. (0 denotes no input)
func EmptyGrid() Grid {
	rows := make([][]int8, GridSize)
	for i := range rows {
		rows[i] = make([]int8, GridSize)
	}
	return rows
}

// MergeGrids Merges one grid into another.
func MergeGrids(result Grid, partial Grid) {
	for i := 0; i < len(result); i++ {
		for j := 0; j < len(result[i]); j++ {
			if partial[i][j] != 0 {
				result[i][j] = partial[i][j]
			}
		}
	}
}

// GridDropUntil drop inputs from a full grid until the required number remain.
func GridDropUntil(grid Grid, startingInputs int) {

	// Randomly remove values
	n := GridSize * GridSize
	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			if rand.Intn(n) >= startingInputs {
				grid[i][j] = 0
			} else {
				startingInputs--
			}
			n--
		}
	}
}

// SelectRandomInput returns a valid input from a given grid.
func SelectRandomInput(grid Grid) *Input {
	c := 0
	var m *Input
	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			if grid[i][j] != 0 {
				c++
				if rand.Intn(c) == 0 {
					m = &Input{
						X: uint8(i),
						Y: uint8(j),
						N: grid[i][j],
					}
				}
			}
		}
	}
	return m
}

// ValidateGrid checks that the grid is in a valid state.
func ValidateGrid(grid Grid) error {
	if len(grid) != GridSize {
		return errors.New("Grid does not have correct number of rows")
	}

	for i := 0; i < GridSize; i++ {
		if len(grid[i]) != GridSize {
			return fmt.Errorf("Row %d is not the correct length", i)
		}
	}
	var seenRow [GridSize][GridSize]bool
	var seenCol [GridSize][GridSize]bool
	var seenSub [GridSize][GridSize]bool

	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			n := grid[i][j]
			if n < 0 || n > GridSize {
				return fmt.Errorf("Invalid value %d at position %d,%d", n, i, j)
			}
			if n > 0 {
				// Decrement n when storing in/looking up validation arrays
				if seenRow[i][n-1] {
					return fmt.Errorf("Duplicate value %d in row %d", n, i)
				}
				seenRow[i][n-1] = true
				if seenCol[j][n-1] {
					return fmt.Errorf("Duplicate value %d in col %d", n, j)
				}
				seenCol[j][n-1] = true
				sub := sqN*(i/sqN) + j/sqN
				if seenSub[sub][n-1] {
					return fmt.Errorf("Duplicate value %d in sub-grid %d", n, j)
				}
				seenSub[sub][n-1] = true
			}
		}
	}

	return nil
}

// RowPerm creates a random row permutation of a sudoku grid.
func RowPerm(row []int8) {

	for i := 0; i < len(row); i++ {
		j := rand.Intn(i + 1)
		row[i] = row[j]
		row[j] = int8(i + 1)
	}
}

// GridPerm creates a random (not necessarily valid Sudoku) grid.
func GridPerm() Grid {
	grid := EmptyGrid()
	for i := 0; i < len(grid); i++ {
		RowPerm(grid[i])
	}

	return grid
}

// Add an input to the Dlx state by covering the constraints the input
// sastisfies.
func addInput(s *sudokuDLXState, input *Input) {

	//Find the corresponding row
	r := &s.inputs[input.X][input.Y][input.N-1][0]

	// cover the column constraints that r satisfies
	dlx.Cover(r.Column)
	for j := r.Right; j != r; j = j.Right {
		dlx.Cover(j.Column)
	}
}

// Remove an input from the Dlx state by uncovering the constraints the input
// sastisfies.
// TODO - Remove dead code
func removeInput(s *sudokuDLXState, input *Input) {
	//Find the corresponding row
	r := &s.inputs[input.X][input.Y][input.N-1][0]

	// cover the column constraints that r satisfies
	for j := r.Left; j != r; j = j.Left {
		dlx.Uncover(j.Column)
	}
	dlx.Uncover(r.Column)

}

// Search for solutions to the current state of the grid.
func (s *sudokuDLXState) Search() (uint32, interface{}) {

	solutionGrid := EmptyGrid()
	sols := dlx.SearchDLX(s, make([]dlx.Row, 0), 0, solutionGrid)
	return sols, solutionGrid
}

// Store a winning grid found in the search.
func (s *sudokuDLXState) StoreSolution(path []dlx.Row, output interface{}) {

	grid := output.(Grid)
	//TODO Should we assume this type assertion is correct?
	for _, t := range path {
		v := strings.Fields(t.Input)
		v0, _ := strconv.Atoi(v[0])
		v1, _ := strconv.Atoi(v[1])
		v2, _ := strconv.Atoi(v[2])
		grid[v0][v1] = int8(v2 + 1)
	}
}

func (s *sudokuDLXState) GetRoot() dlx.Root {
	return s.root
}

// Formalise a sudoku state into an exact cover problem solvable by DLX.
func initialiseSudokuDLXState() *sudokuDLXState {

	root := &dlx.DataObject{}
	dlx.InitialiseColumnHeader(root, root, "root")

	// Columns = Constraints
	// Constraint types:
	// Constraint1 = Every cell must contain one (and only one) number => 9*9 =
	//81 columns
	// Constraint2 = Every row must contain [1-9] exactly once =>  9*9 = 81 columns
	// Constraint3 = Every column must contain [1-9] exactly once => 9*9 = 81 columns
	// Constraint4 = Every subgrid must contain [1-9] exactly once => 9*9 = 81 columns

	// Create 4 * 81 column headers
	var cellHeaders [GridSize][GridSize]dlx.DataObject
	for x := 0; x < GridSize; x++ {
		for y := 0; y < GridSize; y++ {
			dlx.InitialiseColumnHeader(&cellHeaders[x][y], root, fmt.Sprintf("cell %d %d", x, y))
		}
	}

	var rowHeaders [GridSize][GridSize]dlx.DataObject
	for x := 0; x < GridSize; x++ {
		for y := 0; y < GridSize; y++ {
			dlx.InitialiseColumnHeader(&rowHeaders[x][y], root, fmt.Sprintf("row %d %d", x, y))
		}
	}

	var columnHeaders [GridSize][GridSize]dlx.DataObject
	for x := 0; x < GridSize; x++ {
		for y := 0; y < GridSize; y++ {
			dlx.InitialiseColumnHeader(&columnHeaders[x][y], root, fmt.Sprintf("column %d %d", x, y))
		}
	}

	var subgridHeaders [GridSize][GridSize]dlx.DataObject
	for x := 0; x < GridSize; x++ {
		for y := 0; y < GridSize; y++ {
			dlx.InitialiseColumnHeader(&subgridHeaders[x][y], root, fmt.Sprintf("subgrid %d %d", x, y))
		}
	}

	// Rows = Possible inputs.
	// Inputs = number n in cell x,y => 9*9*9 = 729 rows.
	// Every row fulfills 4 constraints, one of each type.
	// Therefore we need to create 4 * 9 * 9 * 9 Data Objects to make exact
	// cover.
	// Input = which cell/row/column/subgrid constraint it fills => which
	//   implies n in cell x,y.
	// Left - Right links the four constraints of the input
	//    e.g. x,y cell <=> row <=> column <=> subgrid for number n.
	// Up - Down links => linked to the constraint it fulfills
	//   e.g. (subgrid 3, value 4).

	var rows Inputs
	for x := 0; x < GridSize; x++ {
		for y := 0; y < GridSize; y++ {
			for n := 0; n < GridSize; n++ {
				//First link up the constraints
				for c := 0; c < constraintTypes; c++ {
					dlx.AppendToRow(&rows[x][y][n][0], &rows[x][y][n][c])
					rows[x][y][n][c].Input = fmt.Sprintf("%d %d %d", x, y, n)
				}
				// Link to cell constraint
				dlx.AppendToColumn(&cellHeaders[x][y], &rows[x][y][n][0])
				// Link to row constraint
				dlx.AppendToColumn(&rowHeaders[x][n], &rows[x][y][n][1])
				// Link to column constraint
				dlx.AppendToColumn(&columnHeaders[y][n], &rows[x][y][n][2])
				// Link to subgrid constraint
				subGrid := sqN*(x/sqN) + (y / sqN)
				dlx.AppendToColumn(&subgridHeaders[subGrid][n], &rows[x][y][n][3])
			}
		}
	}

	return &sudokuDLXState{
		root:   root,
		inputs: &rows,
	}
}

// NewSudokuState sets up a Dlx state for a sudoku grid, priming the state with
// any initial grid inputs.
func NewSudokuState(grid Grid) dlx.State {
	ss := initialiseSudokuDLXState()

	for i := 0; i < len(grid); i++ {
		for j := 0; j < len(grid[i]); j++ {
			if grid[i][j] != 0 {
				addInput(ss, &Input{X: uint8(i), Y: uint8(j), N: grid[i][j]})
			}
		}
	}

	return ss
}
