var sudokuApp = angular.module('sudokuApp', ['ui.bootstrap']);

var SUBGRID_SIZE = 3;
var GRID_SIZE = SUBGRID_SIZE * SUBGRID_SIZE;
var DEFAULT_INPUTS = 32;
var UPDATE_TIME_MILLIS = 750;

sudokuApp.controller('SudokuCtrl', function ($http, $scope, $timeout) {

    var updateGrid = function(newGrid) {
        var i,j;
        for (i = 0; i < GRID_SIZE; i++) {
            for(j = 0; j < GRID_SIZE; j++) {
                if (newGrid[i][j] === 0) {
                    $scope.grid[i][j] = null;
                } else {
                    $scope.grid[i][j] = newGrid[i][j];
                }
            }
        }
        // Run the changed grid logic.
        $scope.gridChange();
    };

    var initialise = function() {
        // Initialise an empty grid.
        var n;
        $scope.grid = new Array(GRID_SIZE);
        for (n = 0; n < GRID_SIZE; n++) {
            $scope.grid[n] = new Array(GRID_SIZE);
        }
        console.log($scope.grid);

        // Retrieve saved state if it exists.
        $http.get('/grid').
            success(function(data, status) {
                if (status === 200) {
                    updateGrid(data.grid);
                } else {
                     $scope.resetGrid();
                }
        }).
        error(function(data) {
            console.warn(data);
        });

        // Set the default inputs
        $scope.startWith = DEFAULT_INPUTS;
    };

    var visualiseUpdate = function(k) {

        $scope.isUpdated[k] = true;

        $timeout(function() {
            delete($scope.isUpdated[k]);
        }, UPDATE_TIME_MILLIS);
    };

    var saveGrid = function() {
        // Save the grid state, return the number of solutions to the grid.
        $http.post('/grid', {grid: $scope.grid}).
            success(function(data) {
                $scope.solutions = data.solutions;
                   
        }).
            error(function(data) {
                console.warn(data);
            });
        
    };

    var validateGrid = function() {
        // Reset the invalid cells.
        $scope.invalidCells = {}; 

        // check each number appears only once in each row and column
        var rowCheck = new Array(GRID_SIZE)
        var colCheck = new Array(GRID_SIZE)
        var subgridCheck = new Array(GRID_SIZE)
        var x
        for (x = 0; x < GRID_SIZE; x++) {
            rowCheck[x] = new Array(GRID_SIZE)
            colCheck[x] = new Array(GRID_SIZE)
            subgridCheck[x] = new Array(GRID_SIZE)
        }
        var i, j, n, seen, subGridIndex, nIndex;
        for (i = 0; i < GRID_SIZE; i++) {
            for (j = 0; j < GRID_SIZE; j++) {
                n = $scope.grid[i][j];
                if (n !==  null) { 
                    if (n instanceof String) {
                        n = parseInt(n,10);
                    }

                    // Decrement value to get index in checker array.
                    nIndex = n-1;

                    // Check that we have yet to see this value in this row.
                    seen = rowCheck[i][nIndex];
                    if (seen) {
                        $scope.invalidCells[i * GRID_SIZE +j] = true;
                    }
                    rowCheck[i][nIndex] = true;

                    // Check that we have yet to see this value in this col.
                    seen = colCheck[j][nIndex];
                    if (seen) {
                        $scope.invalidCells[i * GRID_SIZE +j] = true;
                    }
                    colCheck[j][nIndex] = true;

                    // Check that we have yet to see this value in this subgrid.
                    subGridIndex = SUBGRID_SIZE * Math.floor(i/SUBGRID_SIZE) + Math.floor(j/SUBGRID_SIZE); 
                    seen = subgridCheck[subGridIndex][nIndex];
                    if (seen) {
                        $scope.invalidCells[i * GRID_SIZE +j] = true;
                    }
                    subgridCheck[subGridIndex][nIndex] = true;
                }
            }
        }
        return;
    };

    var fullGrid = function() {
        var i,j;
        for (i = 0; i < GRID_SIZE; i++) {
            for (j = 0; j < GRID_SIZE; j++) {
                if ($scope.grid[i][j] === null) {
                    return false;
                }
            }
        }

        return true;
    };


    $scope.randomGrid = function() {

        // Get how many inputs are desired.
        var sw = ($scope.startWith !== null && $scope.startWith !== undefined) ? $scope.startWith : DEFAULT_INPUTS;

        // Retrieve a random grid from the server.
        $http.get('/randomgrid', {params: { startWith: sw}}).
            success(function(data) {
                updateGrid(data.grid);
        }).
        error(function(data) {
            console.warn(data);
        });
    };

    $scope.resetGrid = function() {
        var i,j;
        // Remove all inputs.
        for (i = 0; i < GRID_SIZE; i++) {
            for (j = 0; j < GRID_SIZE; j++) {
                $scope.grid[i][j] = null;
            }
        }
        // Run grid logic.
        $scope.gridChange();
    };

    $scope.completeGrid = function() {
        // Get a full solution to the problem from the server.
        $http.post('/solutions', {grid: $scope.grid}).
            success(function(data) {
                $scope.solutions = data.solutions;

                var i,j;
                for (i = 0; i < GRID_SIZE; i++) {
                    for (j = 0; j < GRID_SIZE; j++) {
                        if ($scope.grid[i][j] !== data.grid[i][j]) {
                            $scope.grid[i][j] = data.grid[i][j];
                            visualiseUpdate(i * GRID_SIZE + j);
                        }
                    }
                }
                $scope.gridChange();
                   
        }).
            error(function(data) {
                console.warn(data);
            });
        
    };

    $scope.hintGrid = function() {
        // Get an input hint from the server.
        $http.post('/solutions/hint', {grid: $scope.grid, hint: true}).
            success(function(data, status) {
                if (status === 200) {

                    // Insert the input.
                    $scope.grid[data.X][data.Y] = data.N;

                    // Show the input visually to the user.
                    var k = data.X * GRID_SIZE + data.Y;
                    visualiseUpdate(k);

                    // Run the grid changed logic.
                    $scope.gridChange();
                }
        }).
            error(function(data) {
                console.warn(data);
            });
        
    };

    $scope.isUpdated = {}; 
    $scope.invalidCells = {}; 
    $scope.isInvalid = function (i , j) {
        return $scope.invalidCells[i * GRID_SIZE  + j];
    };

    $scope.isUpdated = function (i , j) {
        return $scope.isUpdated[i * GRID_SIZE + j];
    };

    $scope.gridChange = function() {
        validateGrid();
        // Save the grid if there are no invalid inputs.
        if (angular.equals({}, $scope.invalidCells)) {
            saveGrid();
        }
    };

    $scope.solutionsText = function(sols) {
        if (sols > 10000) {
            return "Solutions remaining: Loads";
        } 
        if (sols > 0) {
            if (sols === 1 && fullGrid() ) {
                return "WINNING SOLUTION!"
            }
            return "Solutions remaining: " + sols;
        }
        return "No possible solutions!"
    };

    // Setup the grid
    initialise();
});
