package sudoku

import (
	"math/rand"
	"testing"

	ch "gopkg.in/check.v1"
)

// Set up environment for tests.
// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { ch.TestingT(t) }

type SudokuSuite struct {
}

var _ = ch.Suite(&SudokuSuite{})

// Verify an empty grid is correct.
func (s *SudokuSuite) TestEmptyGrid(c *ch.C) {

	grid := EmptyGrid()

	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			// Ensure all values are zero
			c.Assert(grid[i][j], ch.Equals, int8(0))
		}
	}
}

// Verify merging grids is correct.
func (s *SudokuSuite) TestMergeGrids(c *ch.C) {

	intoGrid := EmptyGrid()
	fromGrid := EmptyGrid()
	expectedGrid := EmptyGrid()

	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			intoGrid[i][j] = int8(rand.Intn(GridSize + 1))
			fromGrid[i][j] = int8(rand.Intn(GridSize + 1))
			if fromGrid[i][j] != 0 {
				expectedGrid[i][j] = fromGrid[i][j]
			} else {
				expectedGrid[i][j] = intoGrid[i][j]
			}
		}
	}

	MergeGrids(intoGrid, fromGrid)

	c.Assert(intoGrid, ch.DeepEquals, expectedGrid)
}

// Verify dropping inputs from a full grid is correct.
func (s *SudokuSuite) TestGridDropUntil(c *ch.C) {

	grid := EmptyGrid()

	// Run it multiple times.
	for n := 0; n < 50; n++ {
		for i := 0; i < GridSize; i++ {
			for j := 0; j < GridSize; j++ {
				//Insert non-zero value
				grid[i][j] = int8(rand.Intn(GridSize) + 1)
			}
		}
		// select a value between 0 and grid size
		expectedInputs := rand.Intn(GridSize * GridSize)

		GridDropUntil(grid, expectedInputs)

		inputs := 0
		// count the non-zero values
		for i := 0; i < GridSize; i++ {
			for j := 0; j < GridSize; j++ {
				if grid[i][j] != 0 {
					inputs++
				}
			}
		}

		c.Assert(inputs, ch.Equals, expectedInputs)
	}
}

// Verify selecting a random input from a grid is correct.
func (s *SudokuSuite) TestSelectRandomInput(c *ch.C) {

	grid := EmptyGrid()

	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			grid[i][j] = int8(rand.Intn(GridSize + 1))
		}
	}

	// Run multiple times.
	for n := 0; n < 50; n++ {

		input := SelectRandomInput(grid)

		c.Assert(input, ch.NotNil)
		// Check the input exists in the grid
		c.Assert(grid[input.X][input.Y], ch.Equals, input.N)
	}

	grid2 := EmptyGrid()
	input := SelectRandomInput(grid2)
	// Expect no input if the grid is empty
	c.Assert(input, ch.IsNil)
}

// Verify grid validation is correct. TODO - Randomise these grids.
func (s *SudokuSuite) TestValidateGrid(c *ch.C) {

	badGrid := [][]int8{
		{5, 3, 4, 6, 7, 8, 9, 1},
		{6, 7, 2, 1, 9, 5, 3, 4, 8},
		{1, 9, 8, 3, 4, 2, 5, 6, 7},
		{8, 5, 9, 7, 6, 1, 4, 2, 3},
		{4, 2, 6, 8, 5, 3, 7, 9, 1},
		{7},
		{9, 6, 1, 5, 3, 7, 2, 8, 4},
		{2, 8, 7, 4, 1, 9, 6, 3, 5},
		{3, 4, 5, 2, 8, 6, 1, 7}}
	validGrid := [][]int8{
		{5, 3, 4, 6, 7, 8, 9, 1, 2},
		{6, 7, 2, 1, 9, 5, 3, 4, 8},
		{1, 9, 8, 3, 4, 2, 5, 6, 7},
		{8, 5, 9, 7, 6, 1, 4, 2, 3},
		{4, 2, 6, 8, 5, 3, 7, 9, 1},
		{7, 1, 3, 9, 2, 4, 8, 5, 6},
		{9, 6, 1, 5, 3, 7, 2, 8, 4},
		{2, 8, 7, 4, 1, 9, 6, 3, 5},
		{3, 4, 5, 2, 8, 6, 1, 7, 9}}
	invalidGrid := [][]int8{
		{4, 2, 6, 8, 5, 3, 7, 9, 1},
		{6, 7, 2, 1, 9, 5, 3, 4, 8},
		{8, 5, 9, 7, 6, 1, 4, 2, 3},
		{5, 3, 4, 6, 7, 8, 9, 1, 2},
		{7, 1, 3, 9, 2, 4, 8, 5, 6},
		{9, 6, 1, 5, 3, 7, 2, 8, 4},
		{2, 8, 7, 4, 1, 9, 6, 3, 5},
		{1, 9, 8, 3, 4, 2, 5, 6, 7},
		{3, 4, 5, 2, 8, 6, 1, 7, 9}}

	err := ValidateGrid(badGrid)
	// Expecting bad grid to return error
	c.Assert(err, ch.NotNil)

	err = ValidateGrid(validGrid)
	// Expecting valid grid to not return error
	c.Assert(err, ch.IsNil)

	err = ValidateGrid(invalidGrid)
	// Expecting invalid grid to return error
	c.Assert(err, ch.NotNil)
}

// Test creation of row permutation is correct.
func (s *SudokuSuite) TestRowPerm(c *ch.C) {

	row := make([]int8, GridSize)

	RowPerm(row)

	// Check that every value form 1..N occurs once
	seen := make([]bool, GridSize)

	for _, v := range row {
		c.Assert(v, ch.Not(ch.Equals), 0)
		c.Assert(seen[v-1], ch.Equals, false)
		seen[v-1] = true
	}
}
