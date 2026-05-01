package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/swimresults/service-core/client"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
)

type RankingClient struct {
	apiUrl string
}

func NewRankingClient(url string) *RankingClient {
	return &RankingClient{apiUrl: url}
}

func (c *RankingClient) ImportRanking(ranking model.Ranking) (*model.Ranking, bool, error) {
	request := dto.ImportRankingRequestDto{
		Ranking: ranking,
	}

	res, err := client.Post(c.apiUrl, "ranking/import", request, nil)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()

	newRanking := &model.Ranking{}
	err = json.NewDecoder(res.Body).Decode(newRanking)
	if err != nil {
		return nil, false, err
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("import age group request returned: %d", res.StatusCode)
	}
	return newRanking, res.StatusCode == http.StatusCreated, nil
}

func (c *RankingClient) GetRankingsForMeetingAndEvent(meeting string, number int) (*[]model.Ranking, error) {
	fmt.Printf("request '%s'\n", c.apiUrl+"/ranking/meet/"+meeting+"/event/"+strconv.Itoa(number))

	res, err := client.Get(c.apiUrl, "ranking/meet/"+meeting+"/event/"+strconv.Itoa(number), nil, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetRankingsForMeetingAndEvent received error: %d\n", res.StatusCode)
	}

	rankings := &[]model.Ranking{}
	err = json.NewDecoder(res.Body).Decode(rankings)
	if err != nil {
		return nil, err
	}

	return rankings, nil
}
