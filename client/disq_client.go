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

type DisqualificationClient struct {
	apiUrl string
}

func NewDisqualificationClient(url string) *DisqualificationClient {
	return &DisqualificationClient{apiUrl: url}
}

func (c *DisqualificationClient) ImportDisqualification(start model.Start, reason string, timeOfAnnouncement time.Time) (*model.Disqualification, bool, error) {
	request := dto.ImportDisqualificationRequestDto{
		Start: start,
		Disqualification: model.Disqualification{
			Reason:           reason,
			AnnouncementTime: timeOfAnnouncement,
		},
	}

	res, err := client.Post(c.apiUrl, "disqualification/import", request)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()

	newDisqualification := &model.Disqualification{}
	err = json.NewDecoder(res.Body).Decode(newDisqualification)
	if err != nil {
		return nil, false, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("import start request returned: %d", res.StatusCode)
	}
	return newDisqualification, true, nil
}
