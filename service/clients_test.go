package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/swimresults/start-service/model"
)

type req struct{ method, uri, serviceKey string }

func stub(t *testing.T, status int, body string) (string, *req) {
	t.Helper()
	r := &req{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, hr *http.Request) {
		r.method, r.uri, r.serviceKey = hr.Method, hr.URL.RequestURI(), hr.Header.Get("X-Swimresults-Service")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/", r // env URLs end with "/"
}

func TestGetAthleteByNameAndYear(t *testing.T) {
	t.Setenv("SR_SERVICE_KEY", "k")
	url, r := stub(t, 200, `{"_id":"650c2e653df25edacb51bd3b","name":"Lilly Kreißig","year":2006}`)
	api, _ := newAthleteAPI(url)
	a, found, err := (&AthleteClient{api}).GetAthleteByNameAndYear("Lilly Kreißig", 2006)
	if err != nil || !found || a.Identifier.Hex() != "650c2e653df25edacb51bd3b" {
		t.Fatalf("a=%+v found=%v err=%v", a, found, err)
	}
	if r.method != http.MethodGet || r.uri != "/athlete/name_year?name=Lilly+Krei%C3%9Fig&year=2006" || r.serviceKey != "k" {
		t.Errorf("request = %s %s key=%q", r.method, r.uri, r.serviceKey)
	}
}

func TestGetAthleteByNameAndYearNotFound(t *testing.T) {
	url, _ := stub(t, 404, `{"message":"no athlete"}`)
	api, _ := newAthleteAPI(url)
	a, found, err := (&AthleteClient{api}).GetAthleteByNameAndYear("X", 2000)
	if err != nil || found || a != nil {
		t.Fatalf("a=%+v found=%v err=%v, want nil,false,nil", a, found, err)
	}
}

func TestGetAthleteByNameAndYearErrors(t *testing.T) {
	url, _ := stub(t, 500, `{"message":"db down"}`)
	api, _ := newAthleteAPI(url)
	if _, _, err := (&AthleteClient{api}).GetAthleteByNameAndYear("X", 2000); err == nil {
		t.Error("500: want error")
	}
	url, _ = stub(t, 200, `{"_id":"not-an-object-id"}`)
	api, _ = newAthleteAPI(url)
	if _, _, err := (&AthleteClient{api}).GetAthleteByNameAndYear("X", 2000); err == nil {
		t.Error("invalid id: want error (the old client failed decoding)")
	}
}

func TestGetTeamByName(t *testing.T) {
	url, r := stub(t, 200, `{"_id":"650c2e653df25edacb51bd3c","name":"SV 1990 Zschopau"}`)
	api, _ := newAthleteAPI(url)
	team, found, err := (&TeamClient{api}).GetTeamByName("SV 1990 Zschopau")
	if err != nil || !found || team.Identifier.Hex() != "650c2e653df25edacb51bd3c" {
		t.Fatalf("team=%+v found=%v err=%v", team, found, err)
	}
	if r.uri != "/team/name?name=SV+1990+Zschopau" {
		t.Errorf("uri = %s", r.uri)
	}
	url, _ = stub(t, 404, `{"message":"no team"}`)
	api, _ = newAthleteAPI(url)
	if team, found, err := (&TeamClient{api}).GetTeamByName("X"); err != nil || found || team != nil {
		t.Fatalf("404: team=%+v found=%v err=%v", team, found, err)
	}
}

func TestEventClient(t *testing.T) {
	url, r := stub(t, 200, `[{"number":1,"final":{"is_final":true}},{"number":2,"final":{}}]`)
	c, _ := newEventClient(url)
	events, err := c.GetEventsByMeetId("IESC19")
	if err != nil || len(*events) != 2 || !(*events)[0].Final.IsFinal || r.uri != "/event/meet/IESC19" {
		t.Fatalf("events=%+v err=%v uri=%s", events, err, r.uri)
	}
	url, r = stub(t, 200, `{"number":7,"distance":100,"relay_distance":"4x50"}`)
	c, _ = newEventClient(url)
	ev, err := c.GetEventByMeetingAndNumber("IESC19", 7)
	if err != nil || ev.Distance != 100 || ev.RelayDistance != "4x50" || r.uri != "/event/meet/IESC19/event/7" {
		t.Fatalf("ev=%+v err=%v uri=%s", ev, err, r.uri)
	}
	url, _ = stub(t, 404, `{"message":"no event"}`)
	c, _ = newEventClient(url)
	if _, err := c.GetEventByMeetingAndNumber("IESC19", 99); err == nil {
		t.Error("404: want error, like before")
	}
}

func TestToObjectID(t *testing.T) {
	if id, err := toObjectID(""); err != nil || !id.IsZero() {
		t.Errorf(`"" -> %v %v`, id, err)
	}
	if _, err := toObjectID("xyz"); err == nil {
		t.Error("invalid hex must fail")
	}
}

// Registration is stored in MongoDB: Address/Contact (now local types) must keep their BSON keys.
func TestRegistrationBSONUnchanged(t *testing.T) {
	r := model.Registration{
		Identifier: primitive.NewObjectID(),
		Address:    model.Address{Street: "Hauptstr.", Number: "1", City: "Zschopau", PostalCode: "09405"},
		Contact:    model.Contact{Name: "A", EMail: "a@example.org", Phone: "1", Fax: "2"},
	}
	b, err := bson.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var doc bson.M
	_ = bson.Unmarshal(b, &doc)
	want := map[string]bson.M{
		"address": {"street": "Hauptstr.", "number": "1", "city": "Zschopau", "postal_code": "09405"},
		"contact": {"name": "A", "email": "a@example.org", "phone": "1", "fax": "2"},
	}
	for k, fields := range want {
		got, _ := doc[k].(bson.M)
		for f, v := range fields {
			if got[f] != v {
				t.Errorf("%s.%s = %v, want %v (doc: %v)", k, f, got[f], v, doc[k])
			}
		}
		if len(got) != len(fields) {
			t.Errorf("%s has keys %v", k, got)
		}
	}
}
