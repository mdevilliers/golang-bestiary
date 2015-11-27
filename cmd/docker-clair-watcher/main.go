package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/fsouza/go-dockerclient"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {

	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)

	if err != nil {
		log.Fatal("Unable to create docker client")
	}

	path, err := ioutil.TempDir("", "docker-clair-watcher")

	if err != nil {
		log.Fatal("Unable to create working folder")
	}

	httpPort := 33301
	myAddress := "127.0.0.1"
	layerServingUrl := myAddress + ":" + strconv.Itoa(httpPort)

	go StartDownloadServer(client, layerServingUrl, path)

	// get initial list of images
	imgs, _ := client.ListImages(docker.ListImagesOptions{All: false})

	log.Println("Total images listed:", len(imgs))
	for _, img := range imgs {

		// predicate on named images
		if len(img.RepoTags) != 0 && !strings.Contains(img.RepoTags[0], "<none>") {

			// get image history
			history, err := GetImageHistory(client, img.ID)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			// spew.Dump(history)

			// send to clair for analysis
			uri := "http://" + layerServingUrl + "/" + history[0] + "/layer.tar"
			err = AnalyzeLayer(uri, history[0], "")

			if err != nil {
				log.Println("Error analyzing layer : ", err)
				continue
			}

			for i := 1; i < len(history); i++ {
				uri := "http://" + layerServingUrl + "/" + history[i] + "/layer.tar"
				err = AnalyzeLayer(uri, history[i], history[i-1])

				if err != nil {
					log.Println("Error analyzing layer : ", err)
					continue
				}

			}
		}
	}

	apiEvents := make(chan *docker.APIEvents)

	go func() {

		for {
			select {
			case event := <-apiEvents:
				spew.Dump(event)
			}
		}

	}()

	client.AddEventListener(apiEvents)

	// start watch for new images
	var ch chan bool
	<-ch // blocks forever
}

func StartDownloadServer(client *docker.Client, listenAddress string, path string) {

	// where clair comes from
	allowedHost := "127.0.0.1"

	err := http.ListenAndServe(listenAddress, restrictedFileServer(client, path, allowedHost))
	if err != nil {
		log.Fatalln("Unable to start image download server.")
	}
}

func restrictedFileServer(client *docker.Client, path, allowedHost string) http.Handler {
	fc := func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request for :", r.RequestURI, " from", r.Host)

		if strings.Contains(r.RemoteAddr, allowedHost) {
			spew.Dump(r.RequestURI)

			// "/511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158/layer.tar"
			bits := strings.Split(r.RequestURI, "/")

			err := ExportAndUnTarImageOnDemand(client, path, bits[1])

			if err != nil {
				log.Println("Error untaring image: ", err.Error())
			}
			//TODO - delete file
			http.FileServer(http.Dir(path)).ServeHTTP(w, r)

			go func() {
				time.Sleep(5 * time.Second)
				removePath := path + "/" + bits[1]

				os.RemoveAll(removePath)
				log.Println("Removed temp file : ", removePath)
			}()

			return
		}
		log.Println("403 : ", r.RequestURI)
		w.WriteHeader(403)
	}
	return http.HandlerFunc(fc)
}

func AnalyzeLayer(uri, layerID, parentLayerID string) error {

	endpoint := "http://127.0.0.1:6060"
	postLayerURI := "/v1/layers"

	payload := struct{ ID, Path, ParentID string }{ID: layerID, Path: uri, ParentID: parentLayerID}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", endpoint+postLayerURI, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 201 {
		body, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("Got response %d with message %s", response.StatusCode, string(body))
	}

	return nil
}

func GetImageHistory(client *docker.Client, imageName string) ([]string, error) {

	history, err := client.ImageHistory(imageName)

	if err != nil {
		return nil, err
	}

	toReturn := []string{}
	for _, h := range history {
		toReturn = append(toReturn, h.ID)
	}

	return toReturn, nil
}

func ExportAndUnTarImageOnDemand(client *docker.Client, path string, imageIdentifier string) error {

	extract := exec.Command("tar", "xf", "-", "-C"+path)

	pipe, err := extract.StdinPipe()
	if err != nil {
		return err
	}

	err = extract.Start()
	if err != nil {
		return err
	}

	err = client.ExportImage(docker.ExportImageOptions{OutputStream: pipe, Name: imageIdentifier})

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
