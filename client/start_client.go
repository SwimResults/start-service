package client

import (
	"encoding/json"
	"fmt"
	"github.com/swimresults/service-core/client"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
	"net/http"
)

type StartClient struct {
	apiUrl string
}

func NewStartClient(url string) *StartClient {
	return &StartClient{apiUrl: url}
}

func (c *StartClient) Actuator() (string, error) {
	res, err := client.Get(c.apiUrl, "actuator", nil, nil)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	// TODO: read string response

	if res.StatusCode == http.StatusOK {
		return "OPERATIONAL", nil
	}
	return "OFFLINE", nil
}

func (c *StartClient) ImportStart(start model.Start) (*model.Start, bool, error) {
	request := dto.ImportStartRequestDto{
		Start: start,
	}

	res, err := client.Post(c.apiUrl, "start/import", request, nil)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()

	newStart := &model.Start{}
	err = json.NewDecoder(res.Body).Decode(newStart)
	if err != nil {
		return nil, false, err
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("import start request returned: %d", res.StatusCode)
	}
	return newStart, res.StatusCode == http.StatusCreated, nil
}

func (c *StartClient) ImportResult(start model.Start, result model.Result) (*model.Result, bool, error) {
	request := dto.ImportResultRequestDto{
		Start:  start,
		Result: result,
	}

	res, err := client.Post(c.apiUrl, "result/import", request, nil)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()

	newResult := &model.Result{}
	err = json.NewDecoder(res.Body).Decode(newResult)
	if err != nil {
		return nil, false, err
	}

	if res.StatusCode != http.StatusCreated {
		return nil, false, fmt.Errorf("import result request returned: %d", res.StatusCode)
	}
	return newResult, res.StatusCode == http.StatusCreated, nil
}
