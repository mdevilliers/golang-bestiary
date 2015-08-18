package main

import (
	c "crypto/rand"
	"fmt"
	"net/http"
	"time"
)

func ChunkedResponse(writer http.ResponseWriter, req *http.Request) {

	msg := rand_str(1024 * 5) // IE need 4 * 1024 to function
	flusher, _ := writer.(http.Flusher)

	writer.Header().Set("Content-Type", "text/html")
	writer.WriteHeader(200)

	flusher.Flush()

	for true {

		fmt.Fprintf(writer, "data: %s\n\n", msg)
		flusher.Flush()

		time.Sleep(time.Second * 2)
	}
}

func main() {
	http.HandleFunc("/", ChunkedResponse)
	http.ListenAndServe(":9999", nil)

}

func rand_str(size int) []byte {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, size)
	c.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return bytes
}
