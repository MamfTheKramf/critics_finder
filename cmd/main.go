package main

import (
	"fmt"
	"math"
)

func worker(channel chan<- bool, vals []int, id int) {
	for _, val := range vals {
		channel <- val%2 == 0
	}

	fmt.Printf("Worker %d done...\n", id)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	channel := make(chan bool, 1)

	vals := make([]int, 1234567)
	for i := 0; i < len(vals); i++ {
		vals[i] = i
	}
	workers := 6

	stepSize := int(math.Ceil(float64(len(vals)) / float64(workers)))
	fmt.Printf("StepSize: %d\n", stepSize)

	sum := 0

	for idx := 0; idx < workers; idx++ {
		lower := idx * stepSize
		upper := min(lower+stepSize, len(vals))
		fmt.Printf("Worker %d: [%d:%d]\n", idx, lower, upper)

		sum += upper - lower

		go worker(channel, vals[lower:upper], idx)
	}
	fmt.Printf("Sum: %d\n", sum)

	odds := 0
	evens := 0

	for i := 0; i < len(vals); i++ {
		res := <-channel
		if res {
			evens += 1
		} else {
			odds += 1
		}
	}

	close(channel)

	fmt.Printf("Evens: %d\nOdds: %d\n", evens, odds)

}
