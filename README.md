# Simple Sudoku 

A simple Sudoku app written in Go & Javascript. Designed to run on [Google App Engine](https://cloud.google.com/appengine/docs).

[Demo](http://sudoku-simple.appspot.com)

## Design

The Sudoku game is played via the Javascript client. This is augmented with
server-side persistence (login-required) and analysis of the current game state.
The game state is saved in App Engines's [Datastore](https://cloud.google.com/appengine/docs/go/datastore/). The server represents the
game state as an [exact cover problem](http://en.wikipedia.org/wiki/Exact_cover), and provides possible solutions and hints
by using the ['Dancing
Links'](http://www-cs-faculty.stanford.edu/~uno/papers/dancing-color.ps.gz) algorithm to search the state space.

## Running locally
The [Go Appengine SDK](https://cloud.google.com/appengine/docs/go/) is required.

Start the server
`goapp serve api`

Point your browser to `http://localhost:8080` 

