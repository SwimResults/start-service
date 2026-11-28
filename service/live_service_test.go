package service

import (
	"strings"
	"testing"
	"time"

	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
	"github.com/swimresults/swimresults-api/go/meetingapi"
)

// berlin builds a time in Europe/Berlin.
func berlin(y int, m time.Month, d, hh, mm int) time.Time {
	return time.Date(y, m, d, hh, mm, 0, 0, liveLocation)
}

func heat(event, number int, planned, estimated, startAt, finishedAt time.Time) *model.Heat {
	return &model.Heat{Meeting: "M", Event: event, Number: number,
		StartEstimation: planned, StartDelayEstimation: estimated, StartAt: startAt, FinishedAt: finishedAt}
}

var zero time.Time

func TestComputeLiveStatusType(t *testing.T) {
	day1 := func(hh, mm int) time.Time { return berlin(2026, 11, 28, hh, mm) }
	day2 := func(hh, mm int) time.Time { return berlin(2026, 11, 29, hh, mm) }

	running := heat(5, 3, day1(10, 0), day1(10, 2), day1(10, 2), zero)
	finished := heat(5, 3, day1(10, 0), day1(10, 2), day1(10, 2), day1(10, 5))
	nextSoon := heat(5, 4, day1(10, 6), day1(10, 8), zero, zero)

	tests := []struct {
		name          string
		now           time.Time
		current, next *model.Heat
		want          dto.LiveStatusType
		wantStartsAt  time.Time // zero: expect nil
	}{
		{"no heats", day1(9, 0), nil, nil, dto.LiveStatusNone, zero},
		{"before first heat", day1(9, 0), nil, heat(1, 1, day1(10, 0), day1(10, 0), zero, zero), dto.LiveStatusBeforeStart, day1(10, 0)},
		{"before first heat, no estimation with delay", day1(9, 0), nil, heat(1, 1, day1(10, 0), zero, zero, zero), dto.LiveStatusBeforeStart, day1(10, 0)},
		{"heat running", day1(10, 3), running, nextSoon, dto.LiveStatusHeat, zero},
		{"finished within grace, no next", day1(10, 6), finished, nil, dto.LiveStatusHeat, zero},
		{"finished exactly at grace", day1(10, 5).Add(LiveFinishGrace), finished, nil, dto.LiveStatusHeat, zero},
		{"finished after grace, no next", day1(10, 7), finished, nil, dto.LiveStatusEnd, zero},
		{"short gap is no break", day1(10, 7), finished, nextSoon, dto.LiveStatusHeat, zero},
		{"gap just below break", day1(10, 7), finished, heat(6, 1, day1(10, 30), day1(10, 34), zero, zero), dto.LiveStatusHeat, zero},
		{"gap exactly break gap", day1(10, 7), finished, heat(6, 1, day1(10, 30), day1(10, 35), zero, zero), dto.LiveStatusHeat, zero},
		{"gap just above break", day1(10, 7), finished, heat(6, 1, day1(10, 30), day1(10, 36), zero, zero), dto.LiveStatusBreak, day1(10, 36)},
		{"lunch break", day1(10, 30), finished, heat(6, 1, day1(14, 0), day1(14, 0), zero, zero), dto.LiveStatusBreak, day1(14, 0)},
		{"break uses estimation with delay", day1(10, 30), finished, heat(6, 1, day1(10, 20), day1(10, 50), zero, zero), dto.LiveStatusBreak, day1(10, 50)},
		{"day end, next tomorrow", day1(18, 0), finished, heat(10, 1, day2(9, 0), day2(9, 0), zero, zero), dto.LiveStatusDayEnd, day2(9, 0)},
		{"next day morning, before first start", day2(7, 0), finished, heat(10, 1, day2(9, 0), day2(9, 0), zero, zero), dto.LiveStatusBeforeStart, day2(9, 0)},
		{"day end, next a week later", day1(18, 0), finished, heat(10, 1, berlin(2026, 12, 5, 9, 0), zero, zero, zero), dto.LiveStatusDayEnd, berlin(2026, 12, 5, 9, 0)},
		// regression: the web app compared only the day of month, so 28 Nov and 28 Dec looked like the same day
		{"next heat a month later on the same day of month", day1(18, 0), finished,
			heat(10, 1, berlin(2026, 12, 28, 9, 0), zero, zero, zero), dto.LiveStatusDayEnd, berlin(2026, 12, 28, 9, 0)},
		// late evening heat: 23:30 Berlin is 22:30 UTC; next at 00:30 Berlin is the next calendar day in Berlin
		{"across midnight in Berlin", day2(0, 10), heat(5, 3, day1(23, 30), zero, day1(23, 30), day1(23, 35)),
			heat(5, 4, day2(0, 30), zero, zero, zero), dto.LiveStatusBeforeStart, day2(0, 30)},
		{"day end decided in Berlin, not UTC", day1(23, 50), heat(5, 3, day1(23, 30), zero, day1(23, 30), day1(23, 35)),
			heat(5, 4, day2(0, 30), zero, zero, zero), dto.LiveStatusDayEnd, day2(0, 30)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ComputeLiveStatus(tt.now, "M", nil, tt.current, tt.next, model.Live{Meeting: "M"})
			if s.Type != tt.want {
				t.Errorf("type = %s, want %s", s.Type, tt.want)
			}
			switch {
			case tt.wantStartsAt.IsZero() && s.StartsAt != nil:
				t.Errorf("starts_at = %v, want nil", *s.StartsAt)
			case !tt.wantStartsAt.IsZero() && (s.StartsAt == nil || !s.StartsAt.Equal(tt.wantStartsAt)):
				t.Errorf("starts_at = %v, want %v", s.StartsAt, tt.wantStartsAt)
			}
		})
	}
}

func TestComputeLiveStatusMeetingAndMessage(t *testing.T) {
	now := berlin(2026, 11, 28, 10, 3)
	cur := heat(5, 3, berlin(2026, 11, 28, 10, 0), zero, now, zero)

	s := ComputeLiveStatus(now, "M", nil, cur, nil, model.Live{})
	if s.MeetingState != "" || s.Warnings == nil || len(s.Warnings) != 0 || s.Message != nil || s.LastDataAt != nil {
		t.Errorf("without meeting/live: %+v", s)
	}
	if s.TimeZone != "Europe/Berlin" || !s.ServerTime.Equal(now) || s.Meeting != "M" || s.Current != cur {
		t.Errorf("base fields: %+v", s)
	}

	m := &meetingapi.Meeting{State: "BREAK", Warnings: []string{"Nur Bahn 1"}}
	msg := &model.LiveMessage{Text: "Siegerehrung", Mode: model.LiveMessageModeReplace}
	s = ComputeLiveStatus(now, "M", m, cur, nil, model.Live{Message: msg})
	if s.MeetingState != "BREAK" || len(s.Warnings) != 1 || s.Message != msg {
		t.Errorf("with meeting/message: %+v", s)
	}
	if s.Type != dto.LiveStatusHeat {
		t.Errorf("meeting state BREAK must not change type: %s", s.Type)
	}

	s = ComputeLiveStatus(now, "M", nil, cur, nil, model.Live{Message: &model.LiveMessage{Text: ""}})
	if s.Message != nil {
		t.Error("empty message must not be returned")
	}
}

func TestComputeLiveStatusDataDelayed(t *testing.T) {
	now := berlin(2026, 11, 28, 10, 10)
	running := heat(5, 3, berlin(2026, 11, 28, 10, 0), zero, berlin(2026, 11, 28, 10, 0), zero)
	finished := heat(5, 3, berlin(2026, 11, 28, 10, 0), zero, berlin(2026, 11, 28, 10, 0), berlin(2026, 11, 28, 10, 2))
	lunch := heat(6, 1, berlin(2026, 11, 28, 14, 0), zero, zero, zero)
	live := func(ago time.Duration) model.Live { return model.Live{LastDataAt: now.Add(-ago)} }
	meeting := func(state string, livetiming bool) *meetingapi.Meeting {
		return &meetingapi.Meeting{State: state, Data: meetingapi.MeetingData{HasLivetiming: livetiming}}
	}

	tests := []struct {
		name    string
		meeting *meetingapi.Meeting
		current *model.Heat
		next    *model.Heat
		live    model.Live
		want    bool
		typ     dto.LiveStatusType
	}{
		{"fresh data", meeting("RUNNING", true), running, nil, live(time.Minute), false, dto.LiveStatusHeat},
		{"exactly at threshold", meeting("RUNNING", true), running, nil, live(LiveDataDelay), false, dto.LiveStatusHeat},
		{"delayed", meeting("RUNNING", true), running, nil, live(LiveDataDelay + time.Second), true, dto.LiveStatusHeat},
		{"never received", meeting("RUNNING", true), running, nil, model.Live{}, false, dto.LiveStatusHeat},
		{"no live timing", meeting("RUNNING", false), running, nil, live(time.Hour), false, dto.LiveStatusHeat},
		{"meeting not running", meeting("BREAK", true), running, nil, live(time.Hour), false, dto.LiveStatusHeat},
		{"meeting unknown", nil, running, nil, live(time.Hour), false, dto.LiveStatusHeat},
		{"break: not delayed, type stays", meeting("RUNNING", true), finished, lunch, live(time.Hour), false, dto.LiveStatusBreak},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ComputeLiveStatus(now, "M", tt.meeting, tt.current, tt.next, tt.live)
			if s.DataDelayed != tt.want || s.Type != tt.typ {
				t.Errorf("data_delayed = %v type = %s, want %v %s", s.DataDelayed, s.Type, tt.want, tt.typ)
			}
		})
	}
}

func TestValidateLiveMessage(t *testing.T) {
	if text, err := ValidateLiveMessage("  Pause bis 14 Uhr \n", model.LiveMessageModeInfo); err != nil || text != "Pause bis 14 Uhr" {
		t.Errorf("valid: %q %v", text, err)
	}
	if _, err := ValidateLiveMessage(strings.Repeat("ä", LiveMessageMaxLength), model.LiveMessageModeReplace); err != nil {
		t.Errorf("max length in characters must pass: %v", err)
	}
	for name, c := range map[string]struct {
		text string
		mode model.LiveMessageMode
	}{
		"empty":     {"   ", model.LiveMessageModeInfo},
		"too long":  {strings.Repeat("a", LiveMessageMaxLength+1), model.LiveMessageModeInfo},
		"bad mode":  {"x", "BANNER"},
		"no mode":   {"x", ""},
		"lowercase": {"x", "info"},
	} {
		if _, err := ValidateLiveMessage(c.text, c.mode); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

func TestLiveStoreWithoutDatabase(t *testing.T) {
	liveCollection = nil
	t.Cleanup(func() { liveMu.Lock(); delete(liveEntries, "STORE"); liveMu.Unlock() })

	if l := getLive("STORE"); !l.LastDataAt.IsZero() || l.Message != nil || l.Meeting != "STORE" {
		t.Fatalf("initial: %+v", l)
	}
	before := time.Now()
	RecordLiveData("STORE", "easywk")
	l := getLive("STORE")
	if l.LastDataAt.Before(before) || l.LastDataSource != "easywk" {
		t.Errorf("after data: %+v", l)
	}

	if err := SetLiveMessage("STORE", "Hallo", model.LiveMessageModeInfo, "user-1"); err != nil {
		t.Fatal(err)
	}
	l = getLive("STORE")
	if l.Message == nil || l.Message.Text != "Hallo" || l.Message.UpdatedBy != "user-1" || l.Message.UpdatedAt.IsZero() {
		t.Fatalf("message: %+v", l.Message)
	}
	l.Message.Text = "changed" // getLive returns a copy
	if getLive("STORE").Message.Text != "Hallo" {
		t.Error("getLive must not expose the stored message")
	}

	if err := ClearLiveMessage("STORE"); err != nil {
		t.Fatal(err)
	}
	if getLive("STORE").Message != nil {
		t.Error("message not cleared")
	}
}

func TestTTLCache(t *testing.T) {
	c := newTTLCache[string, int]()
	now := time.Unix(0, 0)
	c.now = func() time.Time { return now }
	calls := 0
	load := func(v int, err error) func() (int, error) {
		return func() (int, error) { calls++; return v, err }
	}

	if v, _ := c.get("a", time.Minute, load(1, nil)); v != 1 || calls != 1 {
		t.Fatalf("first: v=%d calls=%d", v, calls)
	}
	if v, _ := c.get("a", time.Minute, load(2, nil)); v != 1 || calls != 1 {
		t.Errorf("cached: v=%d calls=%d", v, calls)
	}
	now = now.Add(time.Minute)
	if v, _ := c.get("a", time.Minute, load(2, nil)); v != 2 || calls != 2 {
		t.Errorf("expired: v=%d calls=%d", v, calls)
	}
	now = now.Add(time.Minute)
	if v, err := c.get("a", time.Minute, load(0, errTest)); v != 2 || err != nil {
		t.Errorf("stale on error: v=%d err=%v", v, err)
	}
	if _, err := c.get("b", time.Minute, load(0, errTest)); err == nil {
		t.Error("miss with error: want error")
	}
	c.delete("a")
	if v, _ := c.get("a", time.Minute, load(3, nil)); v != 3 {
		t.Errorf("after delete: v=%d", v)
	}
}

type testError string

func (e testError) Error() string { return string(e) }

const errTest = testError("load failed")
