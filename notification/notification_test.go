package notification

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func stubUserService(t *testing.T, status int, body string) (*map[string]any, *string) {
	t.Helper()
	got, path := &map[string]any{}, new(string)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*path = r.Method + " " + r.URL.Path + " key=" + r.Header.Get("X-Swimresults-Service")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, got)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("SR_START_USER_URL", srv.URL+"/")
	t.Setenv("SR_SERVICE_KEY", "k")
	Init()
	return got, path
}

func TestSendNotificationForMeetingAndAthlete(t *testing.T) {
	body, path := stubUserService(t, 200, `{"success":1}`)
	id, _ := primitive.ObjectIDFromHex("650c2e653df25edacb51bd3b")
	if _, err := notificationClient.SendNotificationForMeetingAndAthlete("IESC19", id, "Wettkampf 3", "Du startest gleich", "athlete", "time-sensitive"); err != nil {
		t.Fatal(err)
	}
	if *path != "POST /notification/meet/IESC19/athlete/650c2e653df25edacb51bd3b key=k" {
		t.Errorf("request = %s", *path)
	}
	want := map[string]any{"subtitle": "Wettkampf 3", "message": "Du startest gleich", "message_type": "athlete", "interruption_level": "time-sensitive"}
	for k, v := range want {
		if (*body)[k] != v {
			t.Errorf("body[%s] = %v, want %v", k, (*body)[k], v)
		}
	}
}

func TestSendNotificationError(t *testing.T) {
	stubUserService(t, 500, `{"message":"apns down"}`)
	if _, err := notificationClient.SendNotificationForMeetingAndAthlete("M", primitive.NewObjectID(), "", "", "athlete", "active"); err == nil {
		t.Error("500: want error")
	}
}

func TestSendMeetingBroadcast(t *testing.T) {
	body, path := stubUserService(t, 200, `{"apns_request_id":"r","apns_unique_id":"u","body":""}`)
	res, err := notificationClient.SendMeetingBroadcastNotification("IESC19", MeetingBroadcastData{Status: MeetingBroadcastStatusData{Event: 3, Heat: 2, Delay: 5}})
	if err != nil || res.ApnsRequestId != "r" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if *path != "POST /notification/broadcast/meeting/IESC19 key=k" {
		t.Errorf("request = %s", *path)
	}
	status, _ := (*body)["status"].(map[string]any)
	if status["event"] != float64(3) || status["heat"] != float64(2) || status["delay"] != float64(5) || status["over"] != false {
		t.Errorf("body = %v", *body)
	}

	stubUserService(t, 410, `{"apns_request_id":"r"}`) // APNs status passed through
	if _, err := notificationClient.SendMeetingBroadcastNotification("IESC19", MeetingBroadcastData{}); err == nil {
		t.Error("non-200: want error, like before")
	}
}
