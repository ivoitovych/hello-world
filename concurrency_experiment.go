
// This is to play with basic concurrency concepts in golang

package main

import (
	"fmt"
	"math/rand"
	"time"
)

func dataSource(dataChannel chan float64, sourceNum int) {

	const outputMin, outputMax float64 = -100, 100
	const delayMaxMs int32 = 1000;

	for {
		delay := time.Millisecond * time.Duration(rand.Int31n(delayMaxMs))
		time.Sleep(delay)
		number := outputMin + (outputMax - outputMin) * rand.Float64()
		fmt.Printf("Source %v puts a number %v into the data channel after a delay of %v\n", sourceNum, number, delay);
		dataChannel <- number;
	}
}

func main() {

	const inputChannelsNumber int = 6

    fmt.Println("Hi! Let's the concurrency experiment begin...")

	inputData := make(chan float64);
	
	for i := 0; i < inputChannelsNumber; i++ {
		go dataSource(inputData, i)
	}

	for {
		number := <-inputData;
		fmt.Printf("Number %v is received\n", number);
	}

}
