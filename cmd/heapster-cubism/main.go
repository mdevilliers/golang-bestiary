package main

import (
	"encoding/csv"
	"fmt"
	// "github.com/davecgh/go-spew/spew"
	cadvisorClient "github.com/google/cadvisor/client"
	info "github.com/google/cadvisor/info/v1"
	"os"
	"time"
)

func main() {

	// TODO : replace with kubeclient
	cluster := []string{"http://172.17.8.102:4194", "http://172.17.8.103:4194"}

	requestChannels := make([]chan *request, len(cluster))

	responseChannel := make(chan *response)
	go func() { fileWriter(responseChannel) }()

	for idx, address := range cluster {

		requestChannel := make(chan *request)
		requestChannels[idx] = requestChannel

		go func(addr string) { monitor(addr, requestChannel, responseChannel) }(address)
	}

	for true {

		now := int32(time.Now().Unix())
		for _, requestChannel := range requestChannels {
			requestChannel <- &request{
				timestamp: now,
			}
		}

		time.Sleep(time.Second * 5)

	}

	var ch chan bool
	<-ch // blocks forever

}

type request struct {
	timestamp int32
}

type response struct {
	memoryStats info.MemoryStats
	cpuStats    info.CpuStats
	identifier  string
	timestamp   int32
}

func monitor(address string, requestChannel chan *request, responseChannel chan *response) {

	client, err := cadvisorClient.NewClient(address)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	containerInfoRequest := info.ContainerInfoRequest{
		NumStats: 1,
	}
	for signal := range requestChannel {

		containerInfo, err := client.ContainerInfo("/", &containerInfoRequest)

		if err != nil {
			fmt.Print(err.Error())
			return
		}

		//	spew.Dump(containerInfo.Stats[0].Memory)
		//	spew.Dump(containerInfo.Stats[0].Cpu)

		responseChannel <- &response{
			memoryStats: containerInfo.Stats[0].Memory,
			cpuStats:    containerInfo.Stats[0].Cpu,
			identifier:  address,
			timestamp:   signal.timestamp,
		}

	}

}

func fileWriter(responseChannel chan *response) {

	csvfile, err := os.Create("data.csv")

	if err != nil {
		fmt.Println("Error creating csv file :", err)
		return
	}

	defer csvfile.Close()

	writer := csv.NewWriter(csvfile)

	writer.Write([]string{"timestamp",
		"id",
		"cpuStats.Usage.Total",
		"memoryStats.Usage",
		"memoryStats.WorkingSet"})

	for {
		select {
		case response := <-responseChannel:

			// fmt.Println("response received - ")
			// spew.Dump(response)

			record := []string{fmt.Sprint(response.timestamp),
				response.identifier,
				fmt.Sprint(response.cpuStats.Usage.Total),
				fmt.Sprint(response.memoryStats.Usage),
				fmt.Sprint(response.memoryStats.WorkingSet)}

			err := writer.Write(record)

			if err != nil {
				fmt.Println("Error writing record :", err)
				return
			}

			writer.Flush()

		}
	}
}
