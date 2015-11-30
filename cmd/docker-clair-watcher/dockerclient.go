package main

import (
	"github.com/fsouza/go-dockerclient"
	"log"
	"os/exec"
	"sort"
	"strings"
)

type DockerClient struct {
	client *docker.Client
}

func NewDockerClient(endpoint string) (*DockerClient, error) {

	client, err := docker.NewClient(endpoint)

	if err != nil {
		return nil, err
	}

	return &DockerClient{
		client: client,
	}, nil
}

func (c *DockerClient) SignalOnImagePull(imageNameChannel chan string) {

	apiEvents := make(chan *docker.APIEvents)

	// TODO : add cancellation signal
	go func() {

		for {
			select {
			case event := <-apiEvents:
				// spew.Dump(event)

				if event.Status == "pull" {
					imageNameChannel <- event.ID
				}
			}
		}

	}()

	c.client.AddEventListener(apiEvents)
}

func (c *DockerClient) RunningContainers() ([]string, error) {

	imgs, err := c.client.ListImages(docker.ListImagesOptions{})

	if err != nil {
		return nil, err
	}

	toReturn := []string{}
	for _, img := range imgs {

		// predicate on named images
		if len(img.RepoTags) != 0 && !strings.Contains(img.RepoTags[0], "<none>") {

			toReturn = append(toReturn, img.ID)
		}
	}

	return toReturn, nil

}

type ImageHistoryByCreated []docker.ImageHistory

func (a ImageHistoryByCreated) Len() int      { return len(a) }
func (a ImageHistoryByCreated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ImageHistoryByCreated) Less(i, j int) bool {
	return a[i].Created < a[j].Created
}

func (c *DockerClient) ImageHistory(imageName string) ([]string, error) {

	history, err := c.client.ImageHistory(imageName)
	sortable := ImageHistoryByCreated(history)
	sort.Sort(sortable)
	if err != nil {
		return nil, err
	}

	toReturn := []string{}
	for _, h := range sortable {
		toReturn = append(toReturn, h.ID)
	}

	return toReturn, nil
}

func (c *DockerClient) ExportAndUnTarImage(path string, imageIdentifier string) error {

	extract := exec.Command("tar", "xf", "-", "-C"+path)

	pipe, err := extract.StdinPipe()
	if err != nil {
		return err
	}

	err = extract.Start()
	if err != nil {
		return err
	}

	err = c.client.ExportImage(docker.ExportImageOptions{OutputStream: pipe, Name: imageIdentifier})

	if err != nil {
		return err
	}

	err = pipe.Close()
	if err != nil {
		return err
	}

	err = extract.Wait()
	if err != nil {
		return err
	}
	return nil
}
