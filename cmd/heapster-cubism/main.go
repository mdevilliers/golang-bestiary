package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	cadvisorClient "github.com/google/cadvisor/client"
	info "github.com/google/cadvisor/info/v1"
)

func main() {

	// TODO : replace with kubeclient
	//cluster := string[] {}

	client, err := cadvisorClient.NewClient("http://172.17.8.102:4194")

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	request := info.ContainerInfoRequest{
		NumStats: 1,
	}
	containerInfo, err := client.ContainerInfo("/", &request)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	spew.Dump(containerInfo.Stats[0].Memory)
	spew.Dump(containerInfo.Stats[0].Cpu)

	// for each node
	// start a go routine
	// passing in
	// 	a return channel of the correct type to return
	// 	a command channel to send go get a result
	// write each complete row to a csv

}
