package notification

// Client for user-service notifications, built on github.com/swimresults/swimresults-api/go/userapi.
// Same endpoints and semantics as the user-service client package it replaces.

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/swimresults/swimresults-api/go/userapi"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const requestTimeout = 30 * time.Second

// nil if SR_START_USER_URL is not configured
var notificationClient *NotificationClient

type NotificationClient struct{ c *userapi.ClientWithResponses }

func Init() {
	userServiceUrl := os.Getenv("SR_START_USER_URL")
	if userServiceUrl == "" {
		return
	}
	c, err := userapi.NewClientWithResponses(strings.TrimRight(userServiceUrl, "/"),
		userapi.WithHTTPClient(&http.Client{Timeout: requestTimeout}),
		userapi.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			if key := os.Getenv("SR_SERVICE_KEY"); key != "" {
				req.Header.Set("X-Swimresults-Service", key)
			}
			return nil
		}))
	if err != nil {
		fmt.Printf("invalid SR_START_USER_URL %q: %v\n", userServiceUrl, err)
		return
	}
	notificationClient = &NotificationClient{c}
}

func (n *NotificationClient) SendNotificationForMeetingAndAthlete(meeting string, athleteId primitive.ObjectID, subtitle string, message string, messageType string, interruptionLevel string) (*userapi.NotificationsResponseDto, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	res, err := n.c.SendNotificationForMeetingAndAthleteWithResponse(ctx, meeting, athleteId.Hex(), userapi.MeetingNotificationRequestDto{
		Subtitle:          subtitle,
		Message:           message,
		MessageType:       messageType,
		InterruptionLevel: interruptionLevel,
	})
	if err != nil {
		return nil, err
	}
	if res.JSON200 == nil {
		return nil, fmt.Errorf("notification request for meeting and athlete returned: %d", res.StatusCode())
	}
	return res.JSON200, nil
}

// SendMeetingBroadcastNotification posts body (any JSON) as a live activity broadcast for the meeting.
func (n *NotificationClient) SendMeetingBroadcastNotification(meeting string, body interface{}) (*userapi.BroadcastResponseDto, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	res, err := n.c.SendMeetingBroadcastWithResponse(ctx, meeting, body)
	if err != nil {
		return nil, err
	}
	// the handler answers with the APNs status code (spec: default response); success is 200, like before
	if res.StatusCode() != http.StatusOK || res.JSONDefault == nil {
		return nil, fmt.Errorf("meeting broadcast request returned: %d", res.StatusCode())
	}
	return res.JSONDefault, nil
}
