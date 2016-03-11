// This is to play with basic concurrency concepts in golang

package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

func dataSource(dataChannel chan float64, sourceNum int) {

	const outputMin, outputMax float64 = -100, 100
	const delayMaxMs int32 = 1000;

	for {
		// simulate the data generation delay
		delay := time.Millisecond * time.Duration(rand.Int31n(delayMaxMs))
		time.Sleep(delay)
		// generate the data
		data := outputMin + (outputMax - outputMin) * rand.Float64()
		fmt.Printf("Source %v puts data %v into the source channel after a delay of %v\n", sourceNum, data, delay)
		dataChannel <- data
	}
}

type Task struct {
	source, intermediate, result float64
	error string
}

func processingStage(inputChannel chan Task, outputChannel chan Task, stageNumber int) {
	for {
		task := <-inputChannel;
		if task.error == "" {
			switch stageNumber {
			case 0:
				if task.source >= 0.0 {
					task.intermediate = math.Sqrt( task.source )
				} else {
					task.error = "Negative Sqrt argument"
				}
			case 1:
				task.result = task.intermediate * task.intermediate;
			case 2:
				if task.result != task.source {
					task.error = "Source and result mismatch"
				}
			}
		}
		outputChannel <- task;
	}
}

func dataProcessor(inputChannel chan float64, outputChannel chan Task, processorNumber int) {

	const delayMaxMs int32 = 1000;

	const numberOfPipelineStages = 3

	// create the processing pipeline
	pipelineInput := make(chan Task)
	pipelineOutput := pipelineInput
	for stage := 0; stage < numberOfPipelineStages; stage++ {
		pipelineIntermediate := make(chan Task)
		go processingStage(pipelineOutput, pipelineIntermediate, stage)
		pipelineOutput = pipelineIntermediate
	}

	// forwart pipeline output to processor output
	go func() {
		for {
			data := <-pipelineOutput
			fmt.Printf("Processor %v obtained output data %v from the pipeline\n",
				processorNumber, data)
			outputChannel <- data
		}
	}()

	for {
		inputData := <-inputChannel
		fmt.Printf("Processor %v received input data %v\n", processorNumber, inputData)
		// simulate a processing delay
		delay := time.Millisecond * time.Duration(rand.Int31n(delayMaxMs))
		time.Sleep(delay)
		fmt.Printf("Processor %v puts input data %v into the pipeline after processing delay of %v\n",
			processorNumber, inputData, delay)
		pipelineInput <- Task{source: inputData}
	}

}

var channelStats []int64

func balancerStats(channelToIncrement int) {
	for len(channelStats) < channelToIncrement + 1 { channelStats = append(channelStats, 0); }
	channelStats[channelToIncrement]++
	sum := channelStats[0]
	for i := 1; i < len(channelStats); i++ { sum += channelStats[i]; }
	fmt.Printf("Load balancer stats: ");
	for _, v := range channelStats {
		fmt.Printf("%d%% ", (v * 100) / sum)
	}
	fmt.Println();
}

func loadBalancer(inputChannel chan float64, processingChannels []chan float64) {
	outputsNumber := len(processingChannels)
	for {
		inputData := <-inputChannel
		fmt.Printf("Load balancer received %v data\n", inputData)
		// choose one of the free input channels
		freeChannelIndices := make([]int, 0, outputsNumber)
		for i :=0; i < outputsNumber; i++ {
			if len(processingChannels[i]) < cap(processingChannels[i]) {
				freeChannelIndices = append(freeChannelIndices, i)
			}
		}
		freeChannels := len(freeChannelIndices)
		fmt.Printf("Load balancer found %v free channels\n", freeChannels)
		if freeChannels > 0 {
			selectedChannel := freeChannelIndices[rand.Intn(freeChannels)]
			fmt.Printf("Load balancer is dispatching the data %v to the processing channel %v\n",
				inputData, selectedChannel)
			processingChannels[selectedChannel] <- inputData
			balancerStats(selectedChannel); // update and print balancer stats
		} else {
			// no free channel found. drop the data, report an error
			fmt.Printf("Load balancer found no free processing channel. Dropping the data %v\n", inputData)
		}
	}
}

func main() {

	const inputChannelsNumber int = 6
	const processingChannelsNumber int = 5
	const processingChannelCapacity int = 3
	
	fmt.Println("Let's the concurrency experiment begin...")

	inputData := make(chan float64);
	balancerData := make(chan float64);
	outputData := make(chan Task);
	
	processingChannels := make([]chan float64, processingChannelsNumber)
	for i := 0; i < processingChannelsNumber; i++ {
		processingChannels[i] = make(chan float64, processingChannelCapacity)
	}

	// start data processors
	for i := 0; i < processingChannelsNumber; i++ {
		go dataProcessor(processingChannels[i], outputData, i)
	}
	
	// start the load balancer
	go loadBalancer(balancerData, processingChannels)
	
	// start input data sources
	for i := 0; i < inputChannelsNumber; i++ {
		go dataSource(inputData, i)
	}

	// receive input data and pass them to load balancer
	go func() {
		for {
			data := <-inputData
			fmt.Printf("Main routine received the data %v from sources and passed it to the load balancer.\n",
				data)
			balancerData <- data
		}
	}()
	
	// receive output data (this is the main thread's main loop)
	for {
		data := <-outputData;
		fmt.Printf("Main routine received the output data %v\n", data)
	}

}
