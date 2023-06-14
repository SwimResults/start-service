package client

import (
	"encoding/json"
	"fmt"
	"github.com/swimresults/service-core/client"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
	"net/http"
	"time"
)

type HeatClient struct {
	apiUrl string
}

func NewHeatClient(url string) *HeatClient {
	return &HeatClient{apiUrl: url}
}

func (c *HeatClient) ImportHeat(meeting string, event int, number int, estimatedStart time.Time) (*model.Heat, bool, error) {
	request := dto.ImportHeatRequestDto{
		Heat: model.Heat{
			Meeting:         meeting,
			Event:           event,
			Number:          number,
			StartEstimation: estimatedStart,
		},
	}

	res, err := client.Post(c.apiUrl, "heat/import", request)
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
