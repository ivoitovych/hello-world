// This is to play with basic concurrency concepts in golang

package main

import (
	"fmt"
	"math/rand"
	"time"
)

func dataSource(dataChannel chan float64, sourceNum int) {

	const outputMin, outputMax float64 = -100, 100
	const delayMaxMs int32 = 100;

	for {
		// simulate the data generation delay
		delay := time.Millisecond * time.Duration(rand.Int31n(delayMaxMs))
		time.Sleep(delay)
		// generate the data
		number := outputMin + (outputMax - outputMin) * rand.Float64()
		fmt.Printf("Source %v puts data %v into the source channel after a delay of %v\n", sourceNum, number, delay)
		dataChannel <- number
	}
}

func dataProcessor(inputChannel chan float64, outputChannel chan float64, processorNumber int) {

	const delayMaxMs int32 = 1000;

	for {
		inputData := <-inputChannel
		fmt.Printf("Processor %v received input data %v\n", processorNumber, inputData)
		// simulate a processing delay
		delay := time.Millisecond * time.Duration(rand.Int31n(delayMaxMs))
		time.Sleep(delay)
		// process the data
		outputData := inputData// TODO: Do the actual processing
		outputChannel <- outputData
		fmt.Printf("Processor %v turned input data %v into output data %v with processing delay of %v\n", processorNumber, inputData, outputData, delay)
	}

}

func loadBalancer(inputChannel chan float64, processingChannels []chan float64) {
	const channelCapacityMax int = 2 // to simulate some channel capacity limit
	outputsNumber := len(processingChannels)
	for {
		inputData := <-inputChannel
		fmt.Printf("Load balancer received %v data\n", inputData)
		// choose one of the free input channels
		freeChannelIndices := make([]int, 0, outputsNumber)
		for i :=0; i < outputsNumber; i++ {
			if len(processingChannels[i]) <= channelCapacityMax {
				freeChannelIndices = append(freeChannelIndices, i)
			}
		}
		freeChannels := len(freeChannelIndices)
		fmt.Printf("Load balancer found %v free channels\n", freeChannels)
		if freeChannels > 0 {
			selectedChannel := rand.Intn(freeChannels)
			fmt.Printf("Dispatching data %v to processing channel %v\n", inputData, selectedChannel)
			processingChannels[selectedChannel] <- inputData
		} else {
			// no free channel found. drop the data, report an error
			fmt.Printf("Load balancer found no free processing channel. Dropping the data %v\n", inputData);
		}
	}
}

func main() {

	const inputChannelsNumber int = 6
	const processingChannelsNumber int = 5
	
	fmt.Println("Let's the concurrency experiment begin...")

	processingChannels := make([]chan float64, processingChannelsNumber)
	for i := 0; i < processingChannelsNumber; i++ {
		processingChannels[i] = make(chan float64);
	}

	inputData := make(chan float64);
	balancerData := make(chan float64);
	outputData := make(chan float64);
	
	// start data processors
	for i := 0; i < processingChannelsNumber; i++ {
		go dataProcessor(processingChannels[i], outputData, i)
	}
	
	// start the load balancer
	go loadBalancer(balancerData, processingChannels)
	//go loadBalancer(inputData, processingChannels)
	
	// start input data sources
	for i := 0; i < inputChannelsNumber; i++ {
		go dataSource(inputData, i)
	}

	// receive input data and pass them to load balancer
	go func() {
		for {
			data := <-inputData
			fmt.Printf("Data %v is received from data sources and passed to the load balancer.\n", data)
			balancerData <- data
		}
	}()
	
	// receive output data (this is the main thread's main loop)
	for {
		data := <-outputData;
		fmt.Printf("Output data is received: %v\n", data);
	}

}
