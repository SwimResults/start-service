package model

// Address and Contact used to come from github.com/swimresults/athlete-service/model (v0.3.0).
// They are stored inside registrations, so the json and bson tags must stay exactly as they were.

type Address struct {
	Street     string `json:"street,omitempty" bson:"street,omitempty"`
	Number     string `json:"number,omitempty" bson:"number,omitempty"`
	City       string `json:"city,omitempty" bson:"city,omitempty"`
	PostalCode string `json:"postal_code,omitempty" bson:"postal_code,omitempty"`
}

type Contact struct {
	Name  string `json:"name,omitempty" bson:"name,omitempty"`
	EMail string `json:"email,omitempty" bson:"email,omitempty"`
	Phone string `json:"phone,omitempty" bson:"phone,omitempty"`
	Fax   string `json:"fax,omitempty" bson:"fax,omitempty"`
}
