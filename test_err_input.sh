#!/bin/bash
# integration test
# run computations with Java and Go and compare results
#
# Ralf Poeppel, 2025-04-12
rm -f *.class &&
    rm -f *out*
    javac Sudoku.java &&
    rm -f Sudoku &&
    go build Sudoku.go &&
    echo "line_to_long.txt"
    ./Sudoku -g line_to_long.txt > line_to_long.outg &&
    # Java returns on exception not zero
    java Sudoku -T1 -non -g line_to_long.txt > line_to_long.outj
    cat line_to_long.outg
    echo
    echo "line_separator.txt"
    java Sudoku -T1 -non -g line_separator.txt > line_separator.outj &&
    ./Sudoku -g line_separator.txt > line_separator.outg &&
    diff --unified line_separator.outj line_separator.outg
    echo
    echo "line_to_short.txt"
    java Sudoku -T1 -non -g line_to_short.txt > line_to_short.outj &&
    ./Sudoku -g line_to_short.txt > line_to_short.outg &&
    diff --unified line_to_short.outj line_to_short.outg
    echo
    echo "one.txt"
    java Sudoku -T1 -non -g one.txt > one.outj &&
    ./Sudoku -g one.txt > one.outg &&
    diff --unified one.outj one.outg
