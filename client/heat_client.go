package client

import (
	"encoding/json"
	"fmt"
	"github.com/swimresults/service-core/client"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
	"net/http"
	"strconv"
)

type HeatClient struct {
	apiUrl string
}

func NewHeatClient(url string) *HeatClient {
	return &HeatClient{apiUrl: url}
}

func (c *HeatClient) ImportHeat(heat model.Heat) (*model.Heat, bool, error) {
	request := dto.ImportHeatRequestDto{
		Heat: heat,
	}

	res, err := client.Post(c.apiUrl, "heat/import", request, nil)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()

	newHeat := &model.Heat{}
	err = json.NewDecoder(res.Body).Decode(newHeat)
	if err != nil {
		return nil, false, err
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("import request returned: %d", res.StatusCode)
	}
	return newHeat, res.StatusCode == http.StatusCreated, nil
}

func (c *HeatClient) SetHeatStart(meeting string, event int, number int) (*model.Heat, error) {
	res, err := client.Post(c.apiUrl, "heat/meet/"+meeting+"/event/"+strconv.Itoa(event)+"/heat/"+strconv.Itoa(number)+"/start", nil, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	newHeat := &model.Heat{}
	err = json.NewDecoder(res.Body).Decode(newHeat)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("set heat start request returned: %d", res.StatusCode)
	}
	return newHeat, nil
}
