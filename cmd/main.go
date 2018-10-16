package main

import (
	"fmt"
	"time"
	"sync"
	"os"
	"strconv"
)

func ParallelMultiply(A, B *SquareMatrix) *SquareMatrix {
	var wg sync.WaitGroup

	C := &SquareMatrix{size : A.size}
	C.data = make([]int, A.size * A.size)

	for row := uint(0); row < A.size; row++ {
		for col := uint(0); col < A.size; col++ {
			go mulRowCol(row, col, A, B, C, wg)
		}
	}
	wg.Wait()

	return C
}

func mulRowCol(row, col uint, A, B, C *SquareMatrix, wg sync.WaitGroup) {
	defer wg.Done()
	wg.Add(1)

	value := 0

	for k := uint(0); k < A.size; k++ {
		//go mul(row, col, k, A, B, C, &value, other_wg, mux)
		value += A.Get(row, k) * B.Get(k, col)
	}
	C.Set(row, col, value)
}

func mul(row, col, k uint, A, B, C *SquareMatrix, value *int, wg sync.WaitGroup, mux sync.Mutex) {
	wg.Add(1)
	defer wg.Done()

	x := A.Get(row, k) * B.Get(k, col)

	//mux.Lock()
	(*value) += x
	//mux.Unlock()
}

func SerialMultiply(A, B *SquareMatrix) *SquareMatrix {
	var C *SquareMatrix = &SquareMatrix{size: A.size}
	C.data = make([]int, A.size * A.size)

	for row := uint(0); row < A.size; row++ {
		for col := uint(0); col < A.size; col++ {
			C.Set(row, col, 0)
			for k := uint(0); k < A.size; k++ {
				value := A.Get(row, k) * B.Get(k, col)
				C.Set(row, col, C.Get(row, col) + value)
			}
		}
	}

	return C
}

func TimeIt(mul func(*SquareMatrix, *SquareMatrix) *SquareMatrix, A, B *SquareMatrix, c chan time.Duration) {
	start := time.Now()
	mul(A,B)
	end := time.Now()

	c <- end.Sub(start)
	close(c)
}

func _main() {
	size, err := strconv.ParseUint(os.Args[1], 10, 64)
	if err != nil {
		fmt.Println("Invalid size for matrix")
	}

	A := UnitMatrix(uint(size))
	B := UnitMatrix(uint(size))

	p_channel := make(chan time.Duration, 1)
	s_channel := make(chan time.Duration, 1)

	go TimeIt(ParallelMultiply, A, B, p_channel)
	go TimeIt(SerialMultiply, A, B, s_channel)

	gotP, gotS := false, false
	for {
		select {
		case elapsed, ok := <-p_channel:
			if ok {
				gotP = true
				fmt.Println("Time in parallel:", elapsed)
			}
		case elapsed, ok := <-s_channel:
			if ok {
				gotS = true
				fmt.Println("Time in serial:", elapsed)
			}
		}
		if gotP && gotS { break }
	}
}
