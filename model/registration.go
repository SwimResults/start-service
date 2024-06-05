package model

import (
	"github.com/swimresults/athlete-service/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Registration struct {
	Identifier           primitive.ObjectID    `json:"_id,omitempty" bson:"_id,omitempty"`
	Meeting              string                `json:"meeting" bson:"meeting"`
	CreatorUserId        primitive.ObjectID    `json:"creator_user_id,omitempty" bson:"creator_user_id,omitempty"`
	Address              model.Address         `json:"address,omitempty" bson:"address,omitempty"`
	Contact              model.Contact         `json:"contact,omitempty" bson:"contact,omitempty"`
	AthleteRegistrations []AthleteRegistration `json:"athlete_registrations" bson:"athlete_registrations"`
	AddedAt              time.Time             `json:"added_at,omitempty" bson:"added_at,omitempty"`
	UpdatedAt            time.Time             `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type AthleteRegistration struct {
	AthleteId          primitive.ObjectID  `json:"athlete_id,omitempty" bson:"athlete_id,omitempty"`
	AthleteFirstName   string              `json:"athlete_first_name,omitempty" bson:"athlete_first_name,omitempty"`
	AthleteLastName    string              `json:"athlete_last_name,omitempty" bson:"athlete_last_name,omitempty"`
	AthleteYear        int                 `json:"athlete_year,omitempty" bson:"athlete_year,omitempty"`
	StartRegistrations []StartRegistration `json:"start_registrations" bson:"start_registrations"`
}

type StartRegistration struct {
	Event            int           `json:"event" bson:"event"`
	RegistrationTime time.Duration `json:"registration_time" bson:"registration_time"`
}
