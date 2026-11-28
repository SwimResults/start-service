package model

import "time"

type LiveMessageMode string

const (
	// LiveMessageModeReplace: the live bar shows the text instead of the computed status.
	LiveMessageModeReplace LiveMessageMode = "REPLACE"
	// LiveMessageModeInfo: the text is shown in addition (live screen banner).
	LiveMessageModeInfo LiveMessageMode = "INFO"
)

// LiveMessage is a manual live message set by an organiser or admin.
type LiveMessage struct {
	Text      string          `json:"text" bson:"text"`
	Mode      LiveMessageMode `json:"mode" bson:"mode"`
	UpdatedAt time.Time       `json:"updated_at" bson:"updated_at"`
	UpdatedBy string          `json:"-" bson:"updated_by,omitempty"`
}

// Live is the live state of one meeting (collection "live", one document per meeting).
type Live struct {
	Meeting        string       `json:"meeting" bson:"meeting"`
	LastDataAt     time.Time    `json:"last_data_at" bson:"last_data_at,omitempty"`
	LastDataSource string       `json:"last_data_source" bson:"last_data_source,omitempty"`
	Message        *LiveMessage `json:"message" bson:"message,omitempty"`
}
