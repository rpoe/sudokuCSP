# Readme

This project is a port of the sudoku solver written in Java bei Peter Norvig to Go.
The Java source from [Sudoku.java](https://github.com/norvig/pytudes/blob/main/ipynb/Sudoku.java)
has been copied to this repository. This will enable easy repetition of the results.

Test input files have been copied too. Links see below.

The copied files are with MIT License as this project.

## Changes

These are the changes to the algotithm in the Go version.:
1. Variable names are only slightly modified to follow the Go conventions.
2. Comments are enhanced. 
3. Not all options of the Java implementation are ported. These options are not implemented:
  - -T<n> number of threads,
  - -t print thread stats,
  - -n constraint naked pair,
  - -u run unit tests.
4. Error handling of functions is as usual in Go with returning an error.
5. Fatal errors are handled with panic.

## Build and Execution Environment

Laptop Lenovo Thinkpad E560, Prozessor Intel Core i7-6500U x 4, Ubuntu 24.04.2 LTS24, Java openjdk version "21.0.6", Go 1.24.2.

## Build

- Java Sudoku.java
- Go build Sudoku.go

## Test

### Unit Tests

Go unit test are postponed for a later commit.

### Integration Tests

Integration tests are run by running the Java program and the Go program on the same
input file with grid output. The output is then compared with unix diff. Test
are considered passed when the grid outputs are same.

Run strings are:
- Java Sudoku -T1 -non -g \<file\>,
- ./Sudoku -g  \<file\>.

Test are executed with the scripts:
1. ./test_err_input.sh
2. ./test_all.sh

Please be patient when executing test_all.sh. It takes some minutes.

Used text files, with source link, are:
- [one.txt](https://github.com/norvig/pytudes/blob/main/ipynb/one.txt)
- [hardest.txt](https://github.com/norvig/pytudes/blob/main/ipynb/hardest.txt)
- [sudoku10k.txt](https://github.com/norvig/pytudes/blob/main/ipynb/sudoku10k.txt)

### Performance Tests

The program computes the average execution time and prints the result by default.

Initial tests showed a faster execution of Java compared to Go. Root cause is the
Java Just in Time Compiler. If an input file has more than 1 Sudoku the compiler can
optimize the compiled code. With the option -Xint the optimization of compiler is switched
of. The runstrings used for the performance tests are:
- Java -Xint Sudoku -T1 -non  \<file\>,
- Java Sudoku -T1 -non -g \<file\>,
- ./Sudoku -g \<file\>.

Test are run 5 times and the smallest value is taken.

Test results, time in µsec per grid:

File            | # Puzzles  | &nbsp; Java -Xint  |          Java  |            Go
:---------------|-----------:|-------------------:|---------------:|---------------:
one.txt         |         1  |            9485,5  | &nbsp; 4740,6  | &nbsp; 1653.0
hardest.txt     |        10  |            1083,2  |         493,4  |         213.2
sudoku10k.txt   |     10000  |            2173,2  |         205,2  |         281.5

Results are as expected. Go is much faster as Java without optimization.
Due to the nature of the program Java speeds up on many executions of the same code.

