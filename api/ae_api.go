package api

import (
	"encoding/json"
	"errors"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/fofanov/sudoku-appengine/persistence"
	"github.com/fofanov/sudoku-appengine/sudoku"
	"github.com/gorilla/context"

	"appengine"
	"appengine/user"
)

type State struct {
	Grid sudoku.Grid `json:"grid"` // State of the game
}

type Solutions struct {
	Count uint32      `json:"solutions"`      // Number of possible solutions
	Grid  sudoku.Grid `json:"grid,omitempty"` // A sample grid solution
}

type contextKey int

// Used for storing the grid in the request context.
const gridKey contextKey = 0

type SudokuAeApi struct {
	store persistence.DatastoreGrid
}

// Write out the response as JSON.
func writeResult(w http.ResponseWriter, r interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// Read the current game state from the request body.
// Expected to be JSON encoded.
func readState(r *http.Request) (*State, error) {

	state := &State{}
	if err := json.NewDecoder(r.Body).Decode(state); err != nil {
		return nil, err
	}

	return state, nil
}

// Read the number starting inputs required in state.
// Expected as query parameter.
func readStartingInputs(r *http.Request) (int, error) {
	startWith := r.URL.Query().Get("startWith")
	if startWith == "" {
		return 0, errors.New("Missing parameter 'startWith'")
	}
	return strconv.Atoi(startWith)
}

// Wrapper for handlers that require a user to be authenticated.
func withAuthRequiredHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		u := user.Current(c)

		if u != nil {
			fn(w, r)
			return
		}

		c.Warningf("Unauthorised request to URL: %v", r.URL)
		http.Error(w, "Unauthorised", http.StatusUnauthorized)
	}

}

// Wrapper for handlers which expect a valid grid state to be sent in requst.
func withValidGridHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		state, err := readState(r)
		if err != nil {
			appengine.NewContext(r).Errorf(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := sudoku.ValidateGrid(state.Grid); err != nil {
			appengine.NewContext(r).Warningf("Received bad grid in request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		context.Set(r, gridKey, state.Grid)
		// We're using the mux package that does calls Clear, but better not expect
		// that depency
		defer context.Clear(r)

		fn(w, r)

	}
}

// Redirect logged out users to login. Redirect logged in users to app.
func (s *SudokuAeApi) loginHandler(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		url, _ := user.LoginURL(c, "/")
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	t, err := template.ParseFiles("index.html")
	if err != nil {
		c.Errorf("Unable to find template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	t.Execute(w, u)
}

// Logout users. Redirect to the login page.
func (s *SudokuAeApi) logoutHandler(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	u := user.Current(c)
	if u != nil {
		url, _ := user.LogoutURL(c, "/")
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// Retrieve the previously stored grid state for a user.
func (s *SudokuAeApi) getGridHandler(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	grid, err := s.store.LoadGrid(c)
	if err != nil {
		c.Errorf("Unable to load grid state: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if grid == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := writeResult(w, &State{Grid: grid}); err != nil {
		c.Errorf(err.Error())
	}

}

// Store the grid state for a user.
func (s *SudokuAeApi) postGridHandler(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	grid, ok := context.Get(r, gridKey).(sudoku.Grid)
	if !ok {
		//Should never happen, but we'll handle it anyway
		c.Errorf("Unable to find grid state from context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Look for solutions to this grid whilst we save it.
	solutions := make(chan uint32)
	go func() {
		sudokuState := sudoku.NewSudokuState(grid)

		// Now search for solutions
		sols, _ := sudokuState.Search()

		solutions <- sols
	}()

	if err := s.store.SaveGrid(c, grid); err != nil {
		c.Errorf("Unable to save grid state: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// This heavily relies on a solution search that always terminates (and for the sake
	// of responsiveness, terminates quickly).
	// If not we may start leaking resources.
	if err := writeResult(w, &Solutions{Count: <-solutions}); err != nil {
		c.Errorf(err.Error())
	}
}

// Give the client the number of solutions (and possible solution grid) to their
// current grid state.
func (s *SudokuAeApi) solutionsHandler(w http.ResponseWriter, r *http.Request) {

	grid, ok := context.Get(r, gridKey).(sudoku.Grid)
	if !ok {
		//Should never happen, but we'll handle it anyway
		appengine.NewContext(r).Errorf("Unable to find grid state from context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sudokuState := sudoku.NewSudokuState(grid)

	// Now search for solutions
	solutions, remainderGrid := sudokuState.Search()
	sudoku.MergeGrids(grid, remainderGrid.(sudoku.Grid))

	if err := writeResult(w, &Solutions{
		Count: solutions,
		Grid:  grid,
	}); err != nil {
		appengine.NewContext(r).Errorf(err.Error())
	}

}

// Give the client a hint as to what to input next, given their current grid
// state.
// TODO - merge with the solutions handler.
func (s *SudokuAeApi) solutionsHintHandler(w http.ResponseWriter, r *http.Request) {

	grid, ok := context.Get(r, gridKey).(sudoku.Grid)
	if !ok {
		//Should never happen, but we'll handle it anyway
		appengine.NewContext(r).Errorf("Unable to find grid state from context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Now search for solutions
	sudokuState := sudoku.NewSudokuState(grid)
	_, remainderGrid := sudokuState.Search()

	m := sudoku.SelectRandomInput(remainderGrid.(sudoku.Grid))
	if m == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := writeResult(w, m); err != nil {
		appengine.NewContext(r).Errorf(err.Error())
	}
}

// Return the client a random grid with the desired number of starting inputs.
func (s *SudokuAeApi) randomGridHandler(w http.ResponseWriter, r *http.Request) {

	startingInputs, err := readStartingInputs(r)
	if err != nil {
		appengine.NewContext(r).Warningf("Unable to get number of starting inputs from request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	randomGrid := sudoku.EmptyGrid()
	// Randomise the one of the rows to improve randomness of whole grid solution.
	// This is because we cut short the search at 10000 solutions.
	sudoku.RowPerm(randomGrid[rand.Intn(sudoku.GridSize)])

	sudokuState := sudoku.NewSudokuState(randomGrid)
	_, remainderGrid := sudokuState.Search()
	sudoku.MergeGrids(randomGrid, remainderGrid.(sudoku.Grid))

	sudoku.GridDropUntil(randomGrid, startingInputs)

	if err := writeResult(w, &State{Grid: randomGrid}); err != nil {
		appengine.NewContext(r).Errorf(err.Error())
	}
}

// Create an instance of the sudoku API to run on GAE.
func NewSudokuAeApi(d persistence.DatastoreGrid) *SudokuAeApi {
	return &SudokuAeApi{
		store: d,
	}
}
