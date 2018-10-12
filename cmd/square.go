package main

import (
	"strconv"
	//"sync"
)

type SquareMatrix struct {
	data []int
	size uint
}

func (m *SquareMatrix) Get(row, column uint) int {
	return m.data[row * m.size + column]
}

func (m * SquareMatrix) Set(row, column uint, value int) {

	m.data[row * m.size + column] = value
}

func UnitMatrix(size uint) *SquareMatrix {
	var matrix *SquareMatrix = &SquareMatrix {
		data: make([]int, size * size),
		size: size,
	}
	for i := uint(0); i < size * size; i++ {
		matrix.data[i] = 1
	}
	return matrix
}

func (matrix *SquareMatrix) String() (s string) {
	s = ""
	for i := uint(0); i < matrix.size; i++ {
		for j:= uint(0); j < matrix.size; j++ {
			s += strconv.Itoa(matrix.Get(i, j))
			s += " "
		}
		s = s[:len(s) - 2]
		s += "\n"
	}
	return
}
