package notification

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
)

func TestLiveBroadcastDataKeepsStatus(t *testing.T) {
	planned := time.Date(2026, 11, 28, 10, 0, 0, 0, time.UTC)
	starts := time.Date(2026, 11, 28, 13, 0, 0, 0, time.UTC)
	s := dto.LiveStatus{
		Type:     dto.LiveStatusBreak,
		StartsAt: &starts,
		Current:  &model.Heat{Event: 5, Number: 3, StartEstimation: planned, StartAt: planned.Add(2 * time.Minute)},
		Message:  &model.LiveMessage{Text: "Siegerehrung", Mode: model.LiveMessageModeInfo},
		Warnings: []string{"a", "b"},
	}

	b, err := json.Marshal(LiveBroadcastData(s))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	// installed iOS versions decode exactly these fields
	st := got["status"]
	if st["event"] != float64(5) || st["heat"] != float64(3) || st["over"] != false || st["delay"] != float64(-120) {
		t.Errorf("status = %v", st)
	}
	live := got["live"]
	if live["type"] != "BREAK" || live["starts_at"] != "2026-11-28T13:00:00Z" || live["message"] != "Siegerehrung" ||
		live["message_mode"] != "INFO" || live["warnings"] != float64(2) {
		t.Errorf("live = %v", live)
	}
}

func TestLiveBroadcastDataEndAndEmpty(t *testing.T) {
	d := LiveBroadcastData(dto.LiveStatus{Type: dto.LiveStatusEnd, Current: &model.Heat{Event: 40, Number: 2}})
	if !d.Status.Over || d.Status.Event != 40 || d.Status.Heat != 2 || d.Status.Delay != 0 {
		t.Errorf("end: %+v", d.Status)
	}

	b, _ := json.Marshal(LiveBroadcastData(dto.LiveStatus{Type: dto.LiveStatusNone}))
	if string(b) != `{"status":{"event":0,"heat":0,"over":false,"delay":0},"live":{"type":"NONE","warnings":0}}` {
		t.Errorf("none: %s", b)
	}
}

func TestLiveBroadcastMessageTruncated(t *testing.T) {
	long := strings.Repeat("ü", 500) // 2 bytes each
	d := LiveBroadcastData(dto.LiveStatus{Type: dto.LiveStatusHeat,
		Message: &model.LiveMessage{Text: long, Mode: model.LiveMessageModeReplace}})
	if n := utf8.RuneCountInString(d.Live.Message); n != LiveBroadcastMessageMaxLength || !strings.HasSuffix(d.Live.Message, "…") {
		t.Errorf("message has %d characters", n)
	}
	if b, _ := json.Marshal(d); len(b) > 1024 {
		t.Errorf("payload is %d bytes", len(b))
	}

	short := LiveBroadcastData(dto.LiveStatus{Message: &model.LiveMessage{Text: "kurz"}})
	if short.Live.Message != "kurz" {
		t.Errorf("short message changed: %q", short.Live.Message)
	}
}

func TestBroadcastLiveStatus(t *testing.T) {
	notificationClient = nil
	if err := BroadcastLiveStatus("M", dto.LiveStatus{}); err == nil {
		t.Error("not configured: want error")
	}

	body, path := stubUserService(t, 200, `{"apns_request_id":"r","apns_unique_id":"u","body":""}`)
	if err := BroadcastLiveStatus("IESC26", dto.LiveStatus{Type: dto.LiveStatusEnd}); err != nil {
		t.Fatal(err)
	}
	if *path != "POST /notification/broadcast/meeting/IESC26 key=k" {
		t.Errorf("request = %s", *path)
	}
	if (*body)["status"] == nil || (*body)["live"] == nil {
		t.Errorf("body = %v", *body)
	}
}
