package service

// Clients for athlete- and meeting-service, built on the generated packages from
// github.com/swimresults/swimresults-api. They keep the method set and semantics of the per-service
// client packages they replace (same endpoints, same "found" flag and 404 handling).

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/swimresults/swimresults-api/go/athleteapi"
	"github.com/swimresults/swimresults-api/go/meetingapi"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const requestTimeout = 30 * time.Second

var httpClient = &http.Client{Timeout: requestTimeout}

// serviceKey authenticates start-service against the other services, like the service-core client did.
func serviceKey(_ context.Context, req *http.Request) error {
	if key := os.Getenv("SR_SERVICE_KEY"); key != "" {
		req.Header.Set("X-Swimresults-Service", key)
	}
	return nil
}

// baseURL accepts the existing env values ("http://athlete-service:8080/").
func baseURL(u string) string { return strings.TrimRight(u, "/") }

func reqCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), requestTimeout)
}

// toObjectID converts an ObjectId from another service. "" means unset (NilObjectID), like before.
func toObjectID(id string) (primitive.ObjectID, error) {
	if id == "" {
		return primitive.NilObjectID, nil
	}
	return primitive.ObjectIDFromHex(id)
}

// ---------------------------------------------------------------- athlete-service

type AthleteClient struct {
	c *athleteapi.ClientWithResponses
}
type TeamClient struct {
	c *athleteapi.ClientWithResponses
}

func newAthleteAPI(url string) (*athleteapi.ClientWithResponses, error) {
	return athleteapi.NewClientWithResponses(baseURL(url),
		athleteapi.WithHTTPClient(httpClient), athleteapi.WithRequestEditorFn(serviceKey))
}

// ForeignAthlete is the part of an athlete-service athlete that start-service uses.
type ForeignAthlete struct {
	Identifier primitive.ObjectID
}

// ForeignTeam is the part of an athlete-service team that start-service uses.
type ForeignTeam struct {
	Identifier primitive.ObjectID
}

// GetAthleteByNameAndYear returns (athlete, true) or (nil, false) if not found (404).
func (a *AthleteClient) GetAthleteByNameAndYear(name string, year int) (*ForeignAthlete, bool, error) {
	ctx, cancel := reqCtx()
	defer cancel()
	res, err := a.c.GetAthleteByNameAndYearWithResponse(ctx, &athleteapi.GetAthleteByNameAndYearParams{Name: &name, Year: &year})
	if err != nil {
		return nil, false, err
	}
	switch {
	case res.StatusCode() == http.StatusNotFound:
		return nil, false, nil
	case res.JSON200 == nil:
		return nil, false, fmt.Errorf("GetAthleteByNameAndYear received error: %d", res.StatusCode())
	}
	id, err := toObjectID(res.JSON200.Identifier)
	if err != nil {
		return nil, false, fmt.Errorf("GetAthleteByNameAndYear: invalid athlete id %q: %w", res.JSON200.Identifier, err)
	}
	return &ForeignAthlete{Identifier: id}, true, nil
}

// GetTeamByName returns (team, true) or (nil, false) if not found (404).
func (t *TeamClient) GetTeamByName(name string) (*ForeignTeam, bool, error) {
	ctx, cancel := reqCtx()
	defer cancel()
	res, err := t.c.GetTeamByNameWithResponse(ctx, &athleteapi.GetTeamByNameParams{Name: &name})
	if err != nil {
		return nil, false, err
	}
	switch {
	case res.StatusCode() == http.StatusNotFound:
		return nil, false, nil
	case res.JSON200 == nil:
		return nil, false, fmt.Errorf("GetTeamByName received error: %d", res.StatusCode())
	}
	id, err := toObjectID(res.JSON200.Identifier)
	if err != nil {
		return nil, false, fmt.Errorf("GetTeamByName: invalid team id %q: %w", res.JSON200.Identifier, err)
	}
	return &ForeignTeam{Identifier: id}, true, nil
}

// ---------------------------------------------------------------- meeting-service

type EventClient struct {
	c *meetingapi.ClientWithResponses
}

func newEventClient(url string) (*EventClient, error) {
	c, err := meetingapi.NewClientWithResponses(baseURL(url),
		meetingapi.WithHTTPClient(httpClient), meetingapi.WithRequestEditorFn(serviceKey))
	return &EventClient{c}, err
}

func (e *EventClient) GetEventsByMeetId(meeting string) (*[]meetingapi.Event, error) {
	ctx, cancel := reqCtx()
	defer cancel()
	res, err := e.c.GetEventsByMeetIdWithResponse(ctx, meeting)
	if err != nil {
		return nil, err
	}
	if res.JSON200 == nil {
		return nil, fmt.Errorf("GetEventsByMeetId received error: %d", res.StatusCode())
	}
	return res.JSON200, nil
}

func (e *EventClient) GetEventByMeetingAndNumber(meeting string, number int) (*meetingapi.Event, error) {
	ctx, cancel := reqCtx()
	defer cancel()
	res, err := e.c.GetEventByMeetingAndNumberWithResponse(ctx, meeting, number)
	if err != nil {
		return nil, err
	}
	if res.JSON200 == nil {
		return nil, fmt.Errorf("GetEventByMeetingAndNumber received error: %d", res.StatusCode())
	}
	return res.JSON200, nil
}

func (e *EventClient) GetMeetingByMeetId(meeting string) (*meetingapi.Meeting, error) {
	ctx, cancel := reqCtx()
	defer cancel()
	res, err := e.c.GetMeetingByMeetIdWithResponse(ctx, meeting)
	if err != nil {
		return nil, err
	}
	if res.JSON200 == nil {
		return nil, fmt.Errorf("GetMeetingByMeetId received error: %d", res.StatusCode())
	}
	return res.JSON200, nil
}

// ---------------------------------------------------------------- caches for the live status

const (
	meetingCacheTTL = 30 * time.Second // warnings and state reach the live status within this time
	eventCacheTTL   = 60 * time.Second
)

var meetingCache = newTTLCache[string, *meetingapi.Meeting]()

type eventKey struct {
	meeting string
	number  int
}

var eventCache = newTTLCache[eventKey, *meetingapi.Event]()

// getMeetingCached returns the meeting, or nil if meeting-service is not configured or unavailable.
func getMeetingCached(meeting string) *meetingapi.Meeting {
	if eventClient == nil {
		return nil
	}
	m, err := meetingCache.get(meeting, meetingCacheTTL, func() (*meetingapi.Meeting, error) {
		return eventClient.GetMeetingByMeetId(meeting)
	})
	if err != nil {
		fmt.Printf("live: meeting %s: %v\n", meeting, err)
		return nil
	}
	return m
}

// getEventCached returns the event, or nil if meeting-service is not configured or unavailable.
func getEventCached(meeting string, number int) *meetingapi.Event {
	if eventClient == nil {
		return nil
	}
	ev, err := eventCache.get(eventKey{meeting, number}, eventCacheTTL, func() (*meetingapi.Event, error) {
		return eventClient.GetEventByMeetingAndNumber(meeting, number)
	})
	if err != nil {
		fmt.Printf("live: event %s/%d: %v\n", meeting, number, err)
		return nil
	}
	return ev
}
