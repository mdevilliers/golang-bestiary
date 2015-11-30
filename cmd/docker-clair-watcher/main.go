package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type dockerClairWatcher struct {
	dockerClient *DockerClient
	clairClient  *ClairClient
	endpoint     string
}

func main() {

	dockerEndpoint := "unix:///var/run/docker.sock"
	dockerClient, err := NewDockerClient(dockerEndpoint)

	if err != nil {
		log.Fatal("Unable to create docker client")
	}

	clairEndpoint := "http://127.0.0.1:6060"
	clairClient := NewClairClient(clairEndpoint)

	path, err := ioutil.TempDir("", "docker-clair-watcher")

	if err != nil {
		log.Fatal("Unable to create working folder")
	}

	log.Println("Watcher dir : ", path)

	myPort := 33301
	myAddress := "127.0.0.1"
	allowedHost := "127.0.0.1"

	layerServingUrl := myAddress + ":" + strconv.Itoa(myPort)

	go startDownloadServer(layerServingUrl, fileServer(dockerClient, path, allowedHost))

	watcher := &dockerClairWatcher{
		clairClient:  clairClient,
		dockerClient: dockerClient,
		endpoint:     layerServingUrl,
	}

	imgChannel := make(chan string)

	go watcher.doAnalysis(imgChannel)

	// get initial list of images
	imgs, err := dockerClient.RunningContainers()

	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Total images listed:", len(imgs))

	for _, img := range imgs {
		imgChannel <- img
	}

	// start watching for new images
	go dockerClient.SignalOnImagePull(imgChannel)

	var ch chan bool
	<-ch // blocks forever
}

func (w *dockerClairWatcher) doAnalysis(imgChannel chan string) {

	for {
		select {

		case img := <-imgChannel:
			log.Println(img)
			history, err := w.dockerClient.ImageHistory(img)
			if err != nil {
				log.Println(err.Error())
				continue
			}

			// send top layer to clair for analysis
			uri := "http://" + w.endpoint + "/" + history[0] + "/layer.tar"
			err = w.clairClient.AnalyzeLayer(uri, history[0], "")

			if err != nil {
				log.Println("Error analyzing :", img, " layer : ", history[0], err)
				continue
			}

			// iterate through other layers
			for i := 1; i < len(history); i++ {

				uri := "http://" + w.endpoint + "/" + history[i] + "/layer.tar"
				err = w.clairClient.AnalyzeLayer(uri, history[i], history[i-1])

				if err != nil {
					log.Println("Error analyzing :", img, " parent :", history[i-1], "layer :", history[i], err)
					continue
				}
			}
		}
	}
}

func startDownloadServer(listenAddress string, handler http.Handler) {

	err := http.ListenAndServe(listenAddress, handler)
	if err != nil {
		log.Fatalln("Unable to start docker image download server.")
	}
}

func fileServer(client *DockerClient, path, allowedHost string) http.Handler {
	fc := func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request for :", r.RequestURI, " from", r.Host)

		if strings.Contains(r.RemoteAddr, allowedHost) {

			// "/511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158/layer.tar"
			bits := strings.Split(r.RequestURI, "/")

			err := client.ExportAndUnTarImage(path, bits[1])

			if err != nil {
				log.Println("Error untaring image: ", err.Error())
			}

			// clean up process
			go func(mypath string, uri string) {
				time.Sleep(10 * time.Second)
				removePath := mypath + "/" + uri

				os.RemoveAll(removePath)
				log.Println("Removed temp file : ", removePath)
			}(path, bits[1])

			http.FileServer(http.Dir(path)).ServeHTTP(w, r)

			return
		}
		log.Println("403 : ", r.RequestURI)
		w.WriteHeader(403)
	}
	return http.HandlerFunc(fc)
}
