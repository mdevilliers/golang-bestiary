package main

import (

	"fmt"

	"github.com/davecgh/go-spew/spew"
	"k8s.io/kubernetes/pkg/api"
	k8 "k8s.io/kubernetes/pkg/client"
	"k8s.io/kubernetes/pkg/labels"
)

func main() {
	fmt.Printf("kubernetes-haproxy-bridge\n")

	ns := api.NamespaceDefault

	config := k8.Config{
		Host: "http://172.17.8.101:8080",
		// Username: "test",
		// Password: "password",
	}
	client, err := k8.New(&config)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	selector := labels.Set{"external/public": "true"}.AsSelector()
	//selector := labels.Everything()
	servicesList, err := client.Services(ns).List(selector)

	if err != nil {
		fmt.Print(err.Error())
		return
	}
	for _, service := range servicesList.Items {
		//		fmt.Printf("%v \n", service.ObjectMeta.Name)
		spew.Dump(service)
		spew.Dump(service.ObjectMeta.Annotations)

	}
	//fmt.Printf("", client.Pods("default").List(labels.Everything(), fields.Everything()));

	//construct state
	//format the config file
	// reload ha proxy

}
