package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ClairClient struct {
	endpoint string
}

func NewClairClient(endpoint string) *ClairClient {
	return &ClairClient{
		endpoint: endpoint,
	}
}

func (c *ClairClient) AnalyzeLayer(uri, layerID, parentLayerID string) error {

	postLayerURI := "/v1/layers"

	payload := struct{ ID, Path, ParentID string }{ID: layerID, Path: uri, ParentID: parentLayerID}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", c.endpoint+postLayerURI, bytes.NewBuffer(jsonPayload))
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
