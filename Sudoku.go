// Solve Sudoku Puzzles
// Ralf Poeppel, 2025
//
// Port of the program Sudoku.java by Peter Norvig 2021 to Go.
// The source of the Java version can be found at:
// https://github.com/norvig/pytudes/blob/main/ipynb/Sudoku.java
// Some parts are changed to be more idiomatic Go.
// The options -T, -t, -n and -u are removed.
//
// Most of the comments are taken from the Java version.
//
// There are two representations of puzzles that we will use:
// 1. A gridstring is 81 chars, with characters '0' or '.' for blank and '1' to '9' for digits.
// 2. A puzzle grid is an int[81] with a digit d (1-9) represented by the integer (1 << (d - 1));
//    that is, a bit pattern that has a single 1 bit representing the digit.
//    A blank is represented by the OR of all the digits 1-9, meaning that any digit is possible.
//    While solving the puzzle, some of these digits are eliminated, leaving fewer possibilities.
//    The puzzle is solved when every square has only a single possibility.
//
// Search for a solution with `search`:
//  - Fill an empty square with a guessed digit and do constraint propagation.
//  - If the guess is consistent, search deeper; if not, try a different guess for the square.
//  - If all guesses fail, back up to the previous level.
//  - In selecting an empty square, we pick one that has the minimum number of possible digits.
//  - To be able to back up, we need to keep the grid from the previous recursive level.
//    But we only need to keep one grid for each level, so to save garbage collection,
//    we pre-allocate one grid per level (there are 81 levels) in a `gridpool`.
// Do constraint propagation with `arcConsistent`, `dualConsistent`.
//

package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strconv"
	"strings"
	"time"
)

//////////////////////////////// main; command line options //////////////////////////////

const usage = "" +
	"usage: Sudoku -(no)[fghnprstuv] | -[RT]<number> | <filename> ...\n" +
	"E.g., -v turns verify flag on, -nov turns it off. -R requires a number. The options:\n\n" +
	"  -f(ile)    Print summary stats for each file (default on)\n" +
	"  -g(rid)    Print each puzzle grid and solution grid (default off)\n" +
	"  -h(elp)    Print this usage message\n" +
	"  -p(uzzle)  Print summary stats for each puzzle (default off)\n" +
	"  -r(everse) Solve the reverse of each puzzle as well as each puzzle itself (default off)\n" +
	"  -s(earch)  Run search (default on, but some puzzles can be solved with CSP methods alone)\n" +
	"  -v(erify)  Verify each solution is valid (default on)\n" +
	"  -R<number> Repeat each puzzle <number> times (default 1)\n" +
	"  <filename> Solve all puzzles in filename, which has one puzzle per line"

//////////////////////////////// Globals ////////////////////////////////

// global variables for options
var printFileStats = true    // -f
var printGrid = false        // -g
var printPuzzleStats = false // -p
var reversePuzzle = false    // -r
var runSearch = true         // -s
var verifySolution = true    // -v
var repeat = 1               // -R

var backtracks = 0 // count total backtracks

//////////////////////////////// Constants ////////////////////////////////

const N = 9 // Number of cells on a side of grid.
const ALL_DIGITS = 0b111111111

var DIGITS = [...]int{1 << 0, 1 << 1, 1 << 2, 1 << 3, 1 << 4, 1 << 5, 1 << 6, 1 << 7, 1 << 8}
var ROWS = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
var COLS = ROWS
var SQUARES [N * N]int
var BLOCKS = [][]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}}
var ALL_UNITS [3 * N][]int
var UNITS [N * N][3][N]int
var PEERS [N * N][20]int
var NUM_DIGITS [ALL_DIGITS + 1]int
var HIGHEST_DIGIT [ALL_DIGITS + 1]int

// init do initialization of other 'constant' global variables
func init() {
	// Initialize SQUARES to be the numbers from 0 to N*N
	for i := range N * N {
		SQUARES[i] = i
	}

	// Initialize ALL_UNITS to be an array of the 27 units: rows, columns, and blocks
	i := 0
	for _, r := range ROWS {
		ALL_UNITS[i] = cross([]int{r}, COLS)
		i++
	}
	for _, c := range COLS {
		ALL_UNITS[i] = cross(ROWS, []int{c})
		i++
	}
	for _, rb := range BLOCKS {
		for _, cb := range BLOCKS {
			ALL_UNITS[i] = cross(rb, cb)
			i++
		}
	}
	// debug fmt.Println(ALL_UNITS)

	// Initialize each UNITS[s] to be an array of the 3 units for square s.
	for _, s := range SQUARES {
		i = 0
		for _, u := range ALL_UNITS {
			if member(s, u) {
				UNITS[s][i] = [9]int(u)
				i++
			}
		}
	}

	// Initialize each PEERS[s] to be an array of the 20 squares that are peers of square s.
	for _, s := range SQUARES {
		i = 0
		for _, u := range UNITS[s] {
			for _, s2 := range u {
				if s2 != s && !member(s2, PEERS[s][:i]) {
					PEERS[s][i] = s2
					i++
				}
			}
		}
	}

	// Initialize NUM_DIGITS[val] to be the number of 1 bits in the bitset val
	// and HIGHEST_DIGIT[val] to the highest bit set in the bitset val
	for val := 0; val <= ALL_DIGITS; val++ {
		uval := uint(val)
		NUM_DIGITS[val] = bits.OnesCount(uval)
		HIGHEST_DIGIT[val] = bits.Len(uval)
	}
}

//////////////////////////////// Main ////////////////////////////////

// main parse command line args and solve puzzles in files.
func main() {
	for _, arg := range os.Args[1:] {
		argrs := []rune(arg)
		if argrs[0] != '-' {
			err := solveFile(arg)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			option := argrs[1]
			value := true
			if strings.HasPrefix(arg, "-no") {
				option = argrs[3]
				value = false
			}
			var err error
			switch option {
			case 'f':
				printFileStats = value
			case 'g':
				printGrid = value
			case 'h':
				fmt.Println(usage)
				return
			case 'p':
				printPuzzleStats = value
			case 'r':
				reversePuzzle = value
			case 's':
				runSearch = value
			case 'v':
				verifySolution = value
			case 'R':
				repeat, err = strconv.Atoi(arg[2:])
			default:
				fmt.Println("Unrecognized option: " + arg + "\n" + usage)
			}
			if err != nil {
				fmt.Println("No numeric value: " + arg + "\n" + usage)
			}
		}
	}
}

//////////////////////////////// Handling Lists of Puzzles ////////////////////////////////

// solveFile  Solve all the puzzles in a file. Report timing statistics.
func solveFile(filename string) (err error) {
	grids, err := readFile(filename)
	// debug fmt.Println("solveFile grids", grids)
	if err != nil {
		return err
	}
	startFileTime := time.Now()
	solveList(grids)
	if printFileStats {
		printStats(len(grids)*repeat, startFileTime, filename)
	}
	return nil
}

// solveList solve a list of puzzles in a single thread.
// repeat -R<number> times; print each puzzle's stats if -p; print grid if -g; verify if -v.
func solveList(grids [][]int) {
	puzzle := make([]int, N*N)          // Used to save a copy of the original grid
	gridpool := make([][N * N]int, N*N) // Reuse grids during the search
	for g, grid := range grids {
		copy(puzzle, grid)
		for i := 0; i < repeat; i++ {
			var startTime time.Time
			if printPuzzleStats {
				startTime = time.Now()
			}
			solution := initialize(grid) // All the real work is a on these lines.
			// debug fmt.Println("solveList", solution)
			if runSearch {
				solution = search(solution, gridpool, 0)
			}
			puzzleNo := "Puzzle " + strconv.Itoa(g+1)
			if printPuzzleStats {
				printStats(1, startTime, puzzleNo)
			}
			if i == 0 && (printGrid || (verifySolution && !verify(solution, puzzle))) {
				printGrids(puzzleNo, grid, solution)
			}
		}
	}
}

//////////////////////////////// Utility functions ////////////////////////////////

// cross Return an array of all squares in the intersection of these rows and cols
func cross(rows, cols []int) []int {
	result := make([]int, len(rows)*len(cols))
	i := 0
	for _, r := range rows {
		for _, c := range cols {
			result[i] = N*r + c
			i++
		}
	}
	// debug fmt.Println("cross", rows, cols, result)
	return result
}

// member return true iff item is an element of array.
func member(item int, array []int) bool {
	// debug fmt.Println("member", item, array)
	for i := 0; i < len(array); i++ {
		if array[i] == item {
			return true
		}
	}
	return false
}

// reverse returns its argument string reversed rune-wise left to right.
// From https://github.com/golang/example/blob/master/hello/reverse/reverse.go.
// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file on the source site.
func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

//////////////////////////////// Search algorithm ////////////////////////////////

// search for a solution to grid. If there is an unfilled square, select one
// and try--that is, search recursively--every possible digit for the square.
// return false if no solution found.
func search(grid []int, gridpool [][N * N]int, level int) []int {
	if grid == nil {
		return nil
	}
	s := selectSquare(grid)
	if s == -1 {
		return grid // No squares to select means we are done!
	}
	for _, d := range DIGITS {
		// For each possible digit d that could fill square s, try it
		if (d & grid[s]) > 0 {
			// Copy grid's contents into gridpool[level], and use that at the next level
			copy(gridpool[level][:], grid)
			result := search(fill(gridpool[level][:], s, d), gridpool, level+1)
			if result != nil {
				return result
			}
			backtracks += 1
		}
	}
	return nil
}

// verify that grid is a solution to the puzzle.
func verify(grid []int, puzzle []int) bool {
	if grid == nil {
		return false
	}
	// Check that all squares have a single digit, and
	// no filled square in the puzzle was changed in the solution.
	for _, s := range SQUARES {
		if (NUM_DIGITS[grid[s]] != 1) || (NUM_DIGITS[puzzle[s]] == 1 && grid[s] != puzzle[s]) {
			return false
		}
	}
	// Check that each unit is a permutation of digits
	for _, u := range ALL_UNITS {
		unit_digits := 0 // All the digits in a unit.
		for s := range u {
			unit_digits |= grid[s]
		}
		if unit_digits != ALL_DIGITS {
			return false
		}
	}
	return true
}

// selectSquare choose an unfilled square with the minimum number of possible values.
// If all squares are filled, return -1 (which means the puzzle is complete).
func selectSquare(grid []int) int {
	square := -1
	mint := N + 1
	for _, s := range SQUARES {
		c := NUM_DIGITS[grid[s]]
		if c == 2 {
			return s // Can't get fewer than 2 possible digits
		} else if c > 1 && c < mint {
			square = s
			mint = c
		}
	}
	return square
}

// fill grid[s] = d. If this leads to contradiction, return nil.
// grid is a slice, gots modified.
func fill(grid []int, s, d int) []int {
	if grid == nil || grid[s]&d == 0 {
		grid = nil
		return nil // d not possible for grid[s]
	}
	grid[s] = d
	for _, p := range PEERS[s] {
		if !eliminate(grid, p, d) {
			grid = nil
			return nil // If we can't eliminate d from all peers of s, then fail
		}
	}
	return grid
}

// Eliminate digit d as a possibility for grid[s].
// Run the 3 constraint propagation routines.
// If constraint propagation detects a contradiction, return false.
// Attention: elements of grid are modified, size stays the same.
func eliminate(grid []int, s, d int) bool {
	// debug fmt.Println(">eliminate", grid, s, d)
	if grid[s]&d == 0 {
		return true // d already eliminated from grid[s]
	}
	grid[s] -= d
	// debug fmt.Println(" eliminate", grid, s, d)
	ret := arcConsistent(grid, s) && dualConsistent(grid, s, d)
	// debug fmt.Println("<eliminate", grid, s, d)
	return ret
}

//////////////////////////////// Constraint Propagation ////////////////////////////////

// arcConsistent check if square s is consistent: that is, it has multiple possible values,
// or it has one possible value which we can consistently fill.
func arcConsistent(grid []int, s int) bool {
	// debug fmt.Println("arcConsistent", grid, s, grid[s])
	count := NUM_DIGITS[grid[s]]
	return count >= 2 || (count == 1 && (fill(grid, s, grid[s]) != nil))
}

// dualConsistent after we eliminate d from possibilities for grid[s],
// check each unit of s and make sure there is some position in the unit where d can go.
// If there is only one possible place for d, fill it with d.
func dualConsistent(grid []int, s, d int) bool {
	for _, u := range UNITS[s] {
		dPlaces := 0 // The number of possible places for d within unit u
		dplace := -1 // Try to find a place in the unit where d can go
		for _, s2 := range u {
			if grid[s2]&d > 0 { // s2 is a possible place for d
				dPlaces++
				if dPlaces > 1 {
					break
				}
				dplace = s2
			}
		}
		if dPlaces == 0 || (dPlaces == 1 && (fill(grid, dplace, d) == nil)) {
			return false
		}
	}
	return true
}

//////////////////////////////// Input ////////////////////////////////

// readFile reads one puzzle per file line and returns a List of puzzle grids.
func readFile(filename string) (grids [][]int, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return grids, err
	}
	defer f.Close()

	grids = make([][]int, 0, 1000)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		gridString := scanner.Text()
		if err := scanner.Err(); err != nil {
			return grids, err
		}
		grid, err := parseGrid(gridString)
		if err != nil {
			return nil, err
		}
		grids = append(grids, grid)
		if reversePuzzle {
			gridString = reverse(gridString)
			grid, err = parseGrid(gridString)
			if err != nil {
				return nil, err
			}
		}
	}

	return
}

// parseGrid parse a gridstring into a puzzle grid: an int[] with values DIGITS[0-9] or ALL_DIGITS.
func parseGrid(gridString string) (grid []int, err error) {
	n2 := N * N
	grid = make([]int, n2)
	gridRunes := []rune(gridString)
	s := 0
	d := 0
	for i, r := range gridRunes {
		if s == n2 {
			// debug fmt.Println(d, "s == n2", s)
			break // Prevent array index out of bounds
		}
		if '1' <= r && r <= '9' {
			// debug fmt.Println(d, "number", r, string(r))
			c, _ := strconv.Atoi(string(r)) // Atoi works only on '1' to '9'
			grid[s] = DIGITS[c-1]           // A single-bit set to represent a digit
			s++
		} else if r == '0' || r == '.' {
			// debug fmt.Println(d, "o .", r, string(r))
			grid[s] = ALL_DIGITS // Any digit is possible
			s++
		} else {
			// debug fmt.Println(d, "skip")
		}
		d = i
	}
	d++
	if s < n2 {
		return nil, errors.New(fmt.Sprintf("Line '%v'\n has %v digits, want %v digits.", gridString, s, n2))
	} else if strings.ContainsAny(string(gridRunes[d:]), ".0123456789") {
		// debug fmt.Println(gridRunes[d:], string(gridRunes[d:]), d)
		return nil, errors.New(fmt.Sprintf("Line '%v'\n has more than %v digits.", gridString, n2))
	}
	// debug fmt.Println("parseGrid", grid)
	return grid, nil
}

// initialize a grid from a puzzle.
// First initialize every square in the new grid to ALL_DIGITS, meaning any value is possible.
// Then, call `fill` on the puzzle's filled squares to initiate constraint propagation.
// grid can be nil.
func initialize(puzzle []int) (grid []int) {
	grid = make([]int, N*N)
	for i := range grid {
		grid[i] = ALL_DIGITS
	}
	for s := range SQUARES {
		if puzzle[s] != ALL_DIGITS {
			fill(grid, s, puzzle[s])
		}
	}
	return grid
}

// ////////////////////////////////// Output //////////////////////////////////
var headerPrinted = false

// printStats print stats on puzzles solved, average time, frequency, threads used, and name.
func printStats(nGrids int, startTime time.Time, name string) {
	t := time.Now()
	elapsed := t.Sub(startTime)
	usecs := float64(elapsed.Microseconds())
	ngrd := float64(nGrids)
	bcktrcks := float64(backtracks) / ngrd
	line := fmt.Sprintf("%7d %6.1f %7.3f %10.1f %s",
		nGrids, usecs/ngrd, 1000*ngrd/usecs, bcktrcks, name)
	if !headerPrinted {
		fmt.Println("Puzzles   μsec     kHz Backtracks Name\n" +
			"======= ====== ======= ========== ====")
		headerPrinted = true
	}
	fmt.Println(line)
	backtracks = 0
}

// printGrids print the original puzzle grid and the solution grid.
func printGrids(name string, puzzle []int, solution []int) {
	bar := "------+-------+------"
	gap := "      " // Space between the puzzle grid and solution grid
	if solution == nil {
		solution = make([]int, N*N)
	}
	solfail := "FAILED:"
	if verify(solution, puzzle) {
		solfail = "Solution:"
	}
	fmt.Printf("\n%-22s%s%s\n", name+":", gap, solfail)
	for r := 0; r < N; r++ {
		fmt.Println(rowString(puzzle, r) + gap + rowString(solution, r))
		if r == 2 || r == 5 {
			fmt.Println(bar + gap + " " + bar)
		}
	}
}

// rowString return a string representing a row of this puzzle.
func rowString(grid []int, r int) string {
	row := ""
	for s := r * 9; s < (r+1)*9; s++ {
		if NUM_DIGITS[grid[s]] == 9 {
			row += "."
		} else {
			if NUM_DIGITS[grid[s]] != 1 {
				row += "?"
			} else {
				row += strconv.Itoa(bits.TrailingZeros(uint(grid[s])) + 1)
			}
		}
		if s%9 == 2 || s%9 == 5 {
			row += " | "
		} else {
			row += " "
		}
	}
	return row
}
