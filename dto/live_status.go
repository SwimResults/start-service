package dto

import (
	"time"

	"github.com/swimresults/start-service/model"
	"github.com/swimresults/swimresults-api/go/meetingapi"
)

type LiveStatusType string

const (
	LiveStatusHeat        LiveStatusType = "HEAT"
	LiveStatusBeforeStart LiveStatusType = "BEFORE_START"
	LiveStatusBreak       LiveStatusType = "BREAK"
	LiveStatusDayEnd      LiveStatusType = "DAY_END"
	LiveStatusEnd         LiveStatusType = "END"
	LiveStatusNone        LiveStatusType = "NONE"
)

// LiveStatus is the response of GET /live/meet/:meet_id (schema LiveStatus in swimresults-api specs/start.yaml).
type LiveStatus struct {
	Meeting      string             `json:"meeting"`
	MeetingState string             `json:"meeting_state"`
	ServerTime   time.Time          `json:"server_time"`
	TimeZone     string             `json:"time_zone"`
	Type         LiveStatusType     `json:"type"`
	StartsAt     *time.Time         `json:"starts_at"`
	Message      *model.LiveMessage `json:"message"`
	Current      *model.Heat        `json:"current"`
	Next         *model.Heat        `json:"next"`
	Event        *meetingapi.Event  `json:"event"`
	HeatCount    int                `json:"heat_count"`
	Warnings     []string           `json:"warnings"`
	LastDataAt   *time.Time         `json:"last_data_at"`
	DataDelayed  bool               `json:"data_delayed"`
}

type LiveDataRequest struct {
	Source string `json:"source"`
}

type LiveMessageRequest struct {
	Text string                `json:"text"`
	Mode model.LiveMessageMode `json:"mode"`
}
