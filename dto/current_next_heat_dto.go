package dto

import "github.com/swimresults/start-service/model"

type CurrentNextHeatDto struct {
	Current *model.Heat `json:"current"`
	Next    *model.Heat `json:"next"`
}
