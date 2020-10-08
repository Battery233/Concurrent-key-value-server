// Tests for SquarerImpl. Students should write their code in this file.

package p0partB

import (
	"fmt"
	"testing"
	"time"
)

const (
	testInputs    = 20
	timeoutMillis = 5000
)

func TestBasicCorrectness(t *testing.T) {
	fmt.Println("Running TestBasicCorrectness.")
	input := make(chan int)
	sq := SquarerImpl{}
	squares := sq.Initialize(input)
	go func() {
		input <- 2
	}()
	timeoutChan := time.After(time.Duration(timeoutMillis) * time.Millisecond)
	select {
	case <-timeoutChan:
		t.Error("Test timed out.")
	case result := <-squares:
		if result != 4 {
			t.Error("Error, got result", result, ", expected 4 (=2^2).")
		}
	}
}

func TestYourFirstGoTest(t *testing.T) {
	input := make(chan int)
	sq := SquarerImpl{}
	squares := sq.Initialize(input)

	// add 20 integers to the channel
	go inputKNumbers(input, testInputs)

	for i := 0; i < testInputs; i++ {
		//validate the results
		timeoutChan := time.After(time.Duration(timeoutMillis) * time.Millisecond)
		select {
		case <-timeoutChan:
			t.Error("Test timed out.")
		case result := <-squares:
			if result != i*i {
				t.Error("Unexpected result!")
			}
		}
	}

	//close the sq and input one more int to the channel
	sq.Close()
	go inputKNumbers(input, 1)

	timeoutChan := time.After(time.Duration(timeoutMillis) * time.Millisecond)
	select {
	case <-timeoutChan:
		return
	case <-squares:
		t.Error("Error, no result should be returned after sq is closed.")
	}
}

func inputKNumbers(input chan int, k int) {
	for i := 0; i < k; i++ {
		input <- i
	}
}
