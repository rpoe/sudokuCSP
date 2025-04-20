#!/bin/bash
# integration test
# run computations with Java and Go and compare results
# java -Xint suppresses optimzations of the Java JIT compiler
#
# Ralf Poeppel, 2025-04-12
rm -f *.class &&
    javac Sudoku.java &&
    rm -f Sudoku &&
    go build Sudoku.go &&
    java Sudoku -T1 -non -g one.txt > one.outj &&
    ./Sudoku -g one.txt > one.outg &&
    java Sudoku -T1 -non -g line_separator.txt > line_separator.outj &&
    ./Sudoku -g line_separator.txt > line_separator.outg &&
    java Sudoku -T1 -non -g hardest.txt > hardest.outj &&
    java -Xint Sudoku -T1 -non -g hardest.txt > hardest.outx &&
    ./Sudoku -g hardest.txt > hardest.outg &&
    java Sudoku -T1 -non -g sudoku10k.txt > sudoku10k.outj &&
    java -Xint Sudoku -T1 -non -g sudoku10k.txt > sudoku10k.outx &&
    ./Sudoku -g sudoku10k.txt > sudoku10k.outg &&
    echo "one.txt"
    diff one.outj one.outg
    echo "line_separator.txt"
    diff line_separator.outj line_separator.outg
    echo "hardest.txt"
    diff hardest.outx hardest.outj
    diff hardest.outj hardest.outg
    echo "sudoku10k.txt"
    diff sudoku10k.outx sudoku10k.outj
    diff sudoku10k.outj sudoku10k.outg
