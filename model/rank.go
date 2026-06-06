package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Rank struct {
	Identifier primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`               // automatically
	Rank       int                `json:"rank,omitempty" bson:"rank,omitempty"`             // PDF + DSV
	Ranking    Ranking            `json:"ranking,omitempty" bson:"-"`                       // LENEX + DSV
	RankingId  primitive.ObjectID `json:"-" bson:"ranking_id,omitempty"`                    // automatically
	AddedAt    time.Time          `json:"added_at,omitempty" bson:"added_at,omitempty"`     // automatically
	UpdatedAt  time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"` // automatically
}
