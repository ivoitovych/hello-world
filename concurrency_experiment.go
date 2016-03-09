
// This is to play with basic concurrency concepts in golang

package main

import (
	"fmt"
	"math/rand"
	"time"
)

func dataSource(dataChannel chan float64, sourceNum int) {

	const outputMin, outputMax float64 = -100, 100

	for {
		delay := time.Millisecond * time.Duration(rand.Int31n(1000))
		time.Sleep(delay)
		number := outputMin + (outputMax - outputMin) * rand.Float64()
		fmt.Printf("Source %v puts a number %v into the data channel after a delay of %v\n", sourceNum, number, delay);
		dataChannel <- number;
	}
}

func main() {

    fmt.Println("Hi! Let's the concurrency experiment begin...")

	inputData := make(chan float64);
	
	for i := 1; i<6; i++ {
		go dataSource(inputData, i)
	}

	for {
		number := <-inputData;
		fmt.Printf("Number %v is received\n", number);
	}

}
