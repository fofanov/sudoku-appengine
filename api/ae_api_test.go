package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"appengine"
	"appengine/aetest"
	"appengine/user"

	"github.com/fofanov/sudoku-appengine/sudoku"
	"github.com/gorilla/context"
	ch "gopkg.in/check.v1"
)

// Set up environment for tests.
// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { ch.TestingT(t) }

type ApiSuite struct {
	con  aetest.Context
	inst aetest.Instance
}

var _ = ch.Suite(&ApiSuite{})

// Set up appengine local services.
func (a *ApiSuite) SetUpSuite(c *ch.C) {
	ins, err := aetest.NewInstance(nil)
	c.Assert(err, ch.IsNil)
	a.inst = ins
}

// Tear down appengine local services.
func (a *ApiSuite) TearDownSuite(c *ch.C) {
	a.inst.Close()
}

// Trivial implemention of interface that keeps grid in memory.
type testDatastore struct {
	grid sudoku.Grid
}

func (t *testDatastore) SaveGrid(c appengine.Context, g sudoku.Grid) error {
	t.grid = g
	return nil
}

func (t *testDatastore) LoadGrid(c appengine.Context) (sudoku.Grid, error) {
	return t.grid, nil
}

// Verify that handlers that require authentication are reached when user is
// known.
func (a *ApiSuite) TestWithAuthRequiredSuccess(c *ch.C) {
	r, err := a.inst.NewRequest("GET", "/requires/login", nil)
	if err != nil {
		c.Fatal(err)
	}

	// Set the user
	aetest.Login(&user.User{ID: "testuser",
		Email: "testuser@company.com"}, r)

	w := httptest.NewRecorder()

	fn := withAuthRequiredHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("foo", "bar")
		w.WriteHeader(http.StatusOK)
	})

	fn(w, r)

	c.Assert(w.Code, ch.Equals, http.StatusOK)
	// Expecting header to be set in response
	c.Assert(w.Header().Get("foo"), ch.Equals, "bar")
}

// Verify that handlers that require authentication are not reached when the
// user is not known.
func (a *ApiSuite) TestWithAuthRequiredFailure(c *ch.C) {
	r, err := a.inst.NewRequest("GET", "/requires/login", nil)
	if err != nil {
		c.Fatal(err)
	}

	// Unset the user
	aetest.Login(&user.User{}, r)

	w := httptest.NewRecorder()

	fn := withAuthRequiredHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("bar", "foo")
		w.WriteHeader(http.StatusOK)
	})

	fn(w, r)

	c.Assert(w.Code, ch.Equals, http.StatusUnauthorized)
	c.Assert(w.Header().Get("bar"), ch.Equals, "")
}

// Verify users are redirected to the app when they are already logged in.
func (a *ApiSuite) TestLoginHandlerLoggedIn(c *ch.C) {
	r, err := a.inst.NewRequest("GET", "/", nil)
	if err != nil {
		c.Fatal(err)
	}

	aetest.Login(&user.User{ID: "testuser",
		Email: "testuser@company.com"}, r)

	w := httptest.NewRecorder()

	NewSudokuAeApi(nil).loginHandler(w, r)

	c.Assert(w.Code, ch.Equals, http.StatusFound)
	// Redirect logged in user to app
	c.Assert(w.Header().Get("Location"), ch.Equals, "/app/index.html")
}

// Verify users are redirected to the appengine login page when they are logged out.
func (a *ApiSuite) TestLoginHandlerLoggedOut(c *ch.C) {
	r, err := a.inst.NewRequest("GET", "/", nil)
	if err != nil {
		c.Fatal(err)
	}

	w := httptest.NewRecorder()

	NewSudokuAeApi(nil).loginHandler(w, r)

	c.Assert(w.Code, ch.Equals, http.StatusFound)
	// Redirect user to log in page
	c.Assert(w.Header().Get("Location"), ch.Equals, "/_ah/login?continue=/")
}

// Verify users are redirected to the appengine logout when they are already logged in.
func (a *ApiSuite) TestLogoutHandlerLoggedIn(c *ch.C) {
	r, err := a.inst.NewRequest("GET", "/", nil)
	if err != nil {
		c.Fatal(err)
	}

	aetest.Login(&user.User{ID: "testuser",
		Email: "testuser@company.com"}, r)

	w := httptest.NewRecorder()

	NewSudokuAeApi(nil).loginHandler(w, r)

	c.Assert(w.Code, ch.Equals, http.StatusFound)
	// Redirect logged in user to app
	c.Assert(w.Header().Get("Location"), ch.Equals, "/_ah/login?continue=")
}

// Verify users are redirected to login when they are logged out.
func (a *ApiSuite) TestLogoutHandlerLoggedOut(c *ch.C) {
	r, err := a.inst.NewRequest("GET", "/", nil)
	if err != nil {
		c.Fatal(err)
	}

	w := httptest.NewRecorder()

	NewSudokuAeApi(nil).loginHandler(w, r)

	c.Assert(w.Code, ch.Equals, http.StatusFound)
	// Redirect user to log in page
	c.Assert(w.Header().Get("Location"), ch.Equals, "/")
}

// Verify that we are able to return a saved grid in the response (if one
// exists).
func (a *ApiSuite) TestGetGrid(c *ch.C) {
	r, err := a.inst.NewRequest("GET", "/grid", nil)
	if err != nil {
		c.Fatal(err)
	}

	// Set the user
	aetest.Login(&user.User{ID: "testuser",
		Email: "testuser@company.com"}, r)

	w := httptest.NewRecorder()
	testStore := &testDatastore{}
	testAeApi := NewSudokuAeApi(testStore)

	testAeApi.getGridHandler(w, r)
	// Expecting no grid data
	c.Assert(w.Code, ch.Equals, http.StatusNoContent)

	testStore.grid = sudoku.GridPerm()
	w2 := httptest.NewRecorder()

	testAeApi.getGridHandler(w2, r)
	// Expecting grid data
	c.Assert(w2.Code, ch.Equals, http.StatusOK)

	var responseState State
	// Response should be valid JSON
	err = json.Unmarshal(w2.Body.Bytes(), &responseState)
	c.Assert(err, ch.IsNil)

	// Stored grid and response grid should be equal
	c.Assert(responseState.Grid, ch.DeepEquals, testStore.grid)
}

// Verify handlers that require a grid reject requests that are missing a grid.
func (a *ApiSuite) TestPostGridMissingGrid(c *ch.C) {
	// Send request with a missing grid
	var missingGrid = &struct{ MissingGrid string }{MissingGrid: "Not a grid"}

	bs, err := json.Marshal(missingGrid)
	if err != nil {
		c.Fatal(err)
	}

	r, err := a.inst.NewRequest("POST", "/grid/required", bytes.NewReader(bs))
	if err != nil {
		c.Fatal(err)
	}

	fn := withValidGridHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("grid", "valid")
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()

	fn(w, r)

	// Expecting bad request response
	c.Assert(w.Code, ch.Equals, http.StatusBadRequest)
	// Expecting header not to be set
	c.Assert(w.Header().Get("grid"), ch.Equals, "")
}

// Verify handlers that require a grid reject requests that send a bad grid.
func (a *ApiSuite) TestPostGridBadGrid(c *ch.C) {
	// Send request with bad grid (not the right length)
	var requestState State
	requestState.Grid = [][]int8{{1, 2, 3, 4}, {1, 3, 2}, {2, 1, 3}}

	bs, err := json.Marshal(&requestState)
	if err != nil {
		c.Fatal(err)
	}

	r, err := a.inst.NewRequest("POST", "/grid/required", bytes.NewReader(bs))
	if err != nil {
		c.Fatal(err)
	}

	fn := withValidGridHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("grid", "valid")
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()

	fn(w, r)

	// Expecting bad request response
	c.Assert(w.Code, ch.Equals, http.StatusBadRequest)
	// Expecting header not to be set
	c.Assert(w.Header().Get("grid"), ch.Equals, "")
}

// Verify handlers that require a grid reject requests that send an invalid grid.
func (a *ApiSuite) TestWithInvalidGrid(c *ch.C) {
	// Send request with invalid grid (not valid sudoku state)
	var requestState State
	requestState.Grid = sudoku.GridPerm()
	// Invalidate in the off-chance it was actually valid
	requestState.Grid[0][0] = requestState.Grid[0][sudoku.GridSize-1]

	bs, err := json.Marshal(&requestState)
	if err != nil {
		c.Fatal(err)
	}

	r, err := a.inst.NewRequest("POST", "/grid/required", bytes.NewReader(bs))
	if err != nil {
		c.Fatal(err)
	}

	fn := withValidGridHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("grid", "valid")
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()

	fn(w, r)

	// Expecting bad request response
	c.Assert(w.Code, ch.Equals, http.StatusBadRequest)
	// Expecting header not to be set
	c.Assert(w.Header().Get("grid"), ch.Equals, "")
}

// Verify that we can store a valid grid and return some solution data in the
// response.
func (a *ApiSuite) TestPostGrid(c *ch.C) {
	grid := sudoku.EmptyGrid()
	// Add a few inputs
	sudoku.RowPerm(grid[rand.Intn(sudoku.GridSize)])

	r, err := a.inst.NewRequest("POST", "/grid", nil)
	if err != nil {
		c.Fatal(err)
	}

	aetest.Login(&user.User{ID: "testuser",
		Email: "testuser@company.com"}, r)

	context.Set(r, gridKey, grid)
	// Manually clear
	defer context.Clear(r)

	testStore := &testDatastore{}
	w := httptest.NewRecorder()

	NewSudokuAeApi(testStore).postGridHandler(w, r)

	// Expecting the request grid to the equal to the stored grid
	c.Assert(grid, ch.DeepEquals, testStore.grid)

	// Expecting possible solutions in response
	c.Assert(w.Code, ch.Equals, http.StatusOK)

	var sols Solutions
	err = json.Unmarshal(w.Body.Bytes(), &sols)
	// Response should be valid JSON
	c.Assert(err, ch.IsNil)

	// Response should contain the number of solutions to the grid
	if sols.Count <= 0 {
		c.Fatal("Expecting number of solutions in response")
	}
}

// Verify that we can return solution data to given a valid grid state.
func (a *ApiSuite) TestSolutionsHandler(c *ch.C) {
	grid := sudoku.EmptyGrid()
	// Add a few inputs
	sudoku.RowPerm(grid[rand.Intn(sudoku.GridSize)])

	r, err := a.inst.NewRequest("POST", "/solutions", nil)
	if err != nil {
		c.Fatal(err)
	}

	context.Set(r, gridKey, grid)
	// Manually clear
	defer context.Clear(r)

	w := httptest.NewRecorder()

	NewSudokuAeApi(nil).solutionsHandler(w, r)

	// Expecting possible solutions in response
	c.Assert(w.Code, ch.Equals, http.StatusOK)

	var sols Solutions
	err = json.Unmarshal(w.Body.Bytes(), &sols)
	// Response should be valid JSON
	c.Assert(err, ch.IsNil)

	// Response should contain the number of solutions to the grid
	if sols.Count <= 0 {
		c.Fatal("Expecting number of solutions in response")
	}

	// Grid should be valid
	err = sudoku.ValidateGrid(sols.Grid)
	c.Assert(err, ch.IsNil)

	// Response grid should be full
	for i := 0; i < sudoku.GridSize; i++ {
		for j := 0; j < sudoku.GridSize; j++ {
			if sols.Grid[i][j] == 0 {
				c.Error("Empty cell at position %d,%d", i, j)
			}
		}
	}
}

// Verify that we can return an input hint for a given valid grid state.
func (a *ApiSuite) TestSolutionsHintHandlerNoHint(c *ch.C) {

	// Get a full grid
	grid := sudoku.EmptyGrid()
	ss := sudoku.NewSudokuState(grid)
	_, fullGrid := ss.Search()

	r, err := a.inst.NewRequest("POST", "/solutions/hint", nil)
	if err != nil {
		c.Fatal(err)
	}

	context.Set(r, gridKey, fullGrid)
	// Manually clear
	defer context.Clear(r)

	w := httptest.NewRecorder()

	NewSudokuAeApi(nil).solutionsHintHandler(w, r)

	// Expecting possible solutions in response
	c.Assert(w.Code, ch.Equals, http.StatusNoContent)
}

// Verify that we can return a random valid grid with the specified number of
// starting inputs.
func (a *ApiSuite) TestRandomGridHandler(c *ch.C) {

	startWith := rand.Intn(sudoku.GridSize * sudoku.GridSize)

	r, err := a.inst.NewRequest("GET", fmt.Sprintf("/randomgrid?startWith=%d", startWith), nil)
	if err != nil {
		c.Fatal(err)
	}

	w := httptest.NewRecorder()

	NewSudokuAeApi(nil).randomGridHandler(w, r)

	// Expecting data response
	c.Assert(w.Code, ch.Equals, http.StatusOK)

	var sols Solutions
	err = json.Unmarshal(w.Body.Bytes(), &sols)
	// Response should be valid JSON
	c.Assert(err, ch.IsNil)

	// Grid should be valid
	err = sudoku.ValidateGrid(sols.Grid)
	c.Assert(err, ch.IsNil)

	// Response grid should contain only the number of inputs we asked for
	count := 0
	for i := 0; i < sudoku.GridSize; i++ {
		for j := 0; j < sudoku.GridSize; j++ {
			if sols.Grid[i][j] != 0 {
				count++
			}
		}
	}
	c.Assert(startWith, ch.Equals, count)
}

// Verify that we the request is rejected if the starting inputs is not
// requested.
func (a *ApiSuite) TestRandomGridHandlerMissingParam(c *ch.C) {

	r, err := a.inst.NewRequest("GET", "/randomgrid", nil)
	if err != nil {
		c.Fatal(err)
	}

	w := httptest.NewRecorder()

	NewSudokuAeApi(nil).randomGridHandler(w, r)

	c.Assert(w.Code, ch.Equals, http.StatusBadRequest)
}
