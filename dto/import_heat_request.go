package dto

import "github.com/swimresults/start-service/model"

type ImportHeatRequestDto struct {
	Heat model.Heat `json:"heat"`
}
