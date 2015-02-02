package api

import (
	"net/http"

	"github.com/fofanov/sudoku-appengine/persistence"
	"github.com/gorilla/mux"
)

func init() {

	store := persistence.NewJsonDatastoreGrid()
	api := NewSudokuAeApi(store)

	r := mux.NewRouter()
	r.HandleFunc("/", api.loginHandler).Methods("GET")
	r.HandleFunc("/grid", withAuthRequiredHandler(api.getGridHandler)).Methods("GET")
	r.HandleFunc("/grid", withAuthRequiredHandler(withValidGridHandler(api.postGridHandler))).Methods("POST")
	r.HandleFunc("/randomgrid", api.randomGridHandler).Methods("GET")
	r.HandleFunc("/solutions", withValidGridHandler(api.solutionsHandler)).Methods("POST")
	r.HandleFunc("/solutions/hint", withValidGridHandler(api.solutionsHintHandler)).Methods("POST")

	http.Handle("/", r)
}
