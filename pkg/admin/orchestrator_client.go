package admin

import (
	"encoding/json"
	"fmt"
	"hexa/pkg/hawk_support"
	"io"
	"log"
	"net/http"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
}

type orchestratorClient struct {
	client HTTPClient
	key string
}

func NewOrchestratorClient(client HTTPClient, key string) Client {
	return &orchestratorClient{client, key}
}

func (c orchestratorClient) Health(url string) (string, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		log.Println(err)
		return "{\"status\": \"fail\"}", err
	}
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

type applicationList struct {
	Applications []application `json:"applications"`
}

type application struct {
	Name string `json:"name"`
}

func (c orchestratorClient) Applications(url string) (applications []Application, err error) {
	resp, err := hawk_support.HawkGet(c.client, "anId", c.key, url)
	if err != nil {
		log.Println(err)
		return applications, err
	}

	var jsonResponse applicationList
	body := resp.Body
	err = json.NewDecoder(body).Decode(&jsonResponse)
	if err != nil {
		fmt.Printf("unable to parse customer json: %s\n", err.Error())
		return applications, err
	}

	for _, app := range jsonResponse.Applications {
		applications = append(applications, Application{app.Name})
	}

	return applications, nil
}
