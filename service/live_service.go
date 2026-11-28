package service

// Live status: one server-side status for the live bar, the live screens and the live notifications of all apps.
// The status type is decided by heat times only (see ComputeLiveStatus); clients translate and format it.

import (
	"errors"
	"fmt"
	"strings"
	"time"
	_ "time/tzdata" // the status compares calendar days in Europe/Berlin; don't depend on the host's zoneinfo
	"unicode/utf8"

	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
	"github.com/swimresults/swimresults-api/go/meetingapi"
)

const (
	// LiveTimeZone is the zone all day comparisons are made in, and that clients format times in.
	LiveTimeZone = "Europe/Berlin"
	// LiveFinishGrace: a finished heat stays the current heat for this long before a status message may show.
	LiveFinishGrace = 1 * time.Minute
	// LiveBreakGap: last finish → next estimated start longer than this (same day) is a break.
	LiveBreakGap = 30 * time.Minute
	// LiveDataDelay: no timing data for longer than this while heats are running marks the data as delayed.
	LiveDataDelay = 3 * time.Minute
	// LiveMessageMaxLength is the maximum length of a manual message, in characters.
	LiveMessageMaxLength = 500
)

var liveLocation = mustLoadLocation(LiveTimeZone)

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.In(liveLocation).Date()
	by, bm, bd := b.In(liveLocation).Date()
	return ay == by && am == bm && ad == bd
}

// nextStart is the estimated start of a heat: with delay if known, else as planned.
func nextStart(h *model.Heat) *time.Time {
	t := h.StartDelayEstimation
	if t.IsZero() {
		t = h.StartEstimation
	}
	if t.IsZero() {
		return nil
	}
	return &t
}

// ComputeLiveStatus derives the live status from heat times (pure; no I/O). meeting may be nil when
// meeting-service is unavailable. Event and HeatCount are filled in by the caller.
func ComputeLiveStatus(now time.Time, meetingId string, meeting *meetingapi.Meeting, current, next *model.Heat, live model.Live) dto.LiveStatus {
	s := dto.LiveStatus{
		Meeting:    meetingId,
		ServerTime: now,
		TimeZone:   LiveTimeZone,
		Current:    current,
		Next:       next,
		Warnings:   []string{},
	}
	if meeting != nil {
		s.MeetingState = meeting.State
		if meeting.Warnings != nil {
			s.Warnings = meeting.Warnings
		}
	}
	if live.Message != nil && live.Message.Text != "" {
		s.Message = live.Message
	}
	if !live.LastDataAt.IsZero() {
		t := live.LastDataAt
		s.LastDataAt = &t
	}

	s.Type = liveStatusType(now, current, next)
	switch s.Type {
	case dto.LiveStatusBeforeStart, dto.LiveStatusBreak, dto.LiveStatusDayEnd:
		s.StartsAt = nextStart(next)
	}

	s.DataDelayed = s.Type == dto.LiveStatusHeat &&
		meeting != nil && meeting.State == "RUNNING" && meeting.Data.HasLivetiming &&
		s.LastDataAt != nil && now.Sub(*s.LastDataAt) > LiveDataDelay

	return s
}

func liveStatusType(now time.Time, current, next *model.Heat) dto.LiveStatusType {
	if current == nil {
		if next != nil {
			return dto.LiveStatusBeforeStart
		}
		return dto.LiveStatusNone
	}
	if current.FinishedAt.IsZero() || now.Sub(current.FinishedAt) <= LiveFinishGrace {
		return dto.LiveStatusHeat
	}
	if next == nil {
		return dto.LiveStatusEnd
	}
	if sameDay(current.StartEstimation, next.StartEstimation) {
		if start := nextStart(next); start != nil && start.Sub(current.FinishedAt) > LiveBreakGap {
			return dto.LiveStatusBreak
		}
		return dto.LiveStatusHeat
	}
	if sameDay(now, current.StartEstimation) {
		return dto.LiveStatusDayEnd
	}
	return dto.LiveStatusBeforeStart
}

// liveHeats is the heat part of a live status; cached briefly because every client polls it.
type liveHeats struct {
	current, next *model.Heat
	event         *meetingapi.Event
	heatCount     int
}

const liveHeatsTTL = 3 * time.Second

var liveHeatsCache = newTTLCache[string, liveHeats]()

func loadLiveHeats(meeting string) (liveHeats, error) {
	cn, err := GetCurrentAndNextHeat(meeting)
	if err != nil {
		return liveHeats{}, err
	}
	h := liveHeats{current: cn.Current, next: cn.Next}
	if h.current != nil && h.current.Event != 0 {
		h.event = getEventCached(meeting, h.current.Event)
		if h.heatCount, err = GetHeatsAmountByMeetingAndEvent(meeting, h.current.Event); err != nil {
			return liveHeats{}, err
		}
	}
	return h, nil
}

// GetLiveStatus returns the live status of a meeting at the current time.
func GetLiveStatus(meeting string) (dto.LiveStatus, error) {
	heats, err := liveHeatsCache.get(meeting, liveHeatsTTL, func() (liveHeats, error) { return loadLiveHeats(meeting) })
	if err != nil {
		return dto.LiveStatus{}, err
	}
	s := ComputeLiveStatus(time.Now().In(liveLocation), meeting, getMeetingCached(meeting), heats.current, heats.next, getLive(meeting))
	s.Event = heats.event
	s.HeatCount = heats.heatCount
	return s, nil
}

// invalidateLiveHeats drops the cached heats after a heat changed, so the next status read is current.
func invalidateLiveHeats(meeting string) {
	liveHeatsCache.delete(meeting)
}

// ValidateLiveMessage trims the text and checks text and mode of a manual message.
func ValidateLiveMessage(text string, mode model.LiveMessageMode) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("text must not be empty")
	}
	if n := utf8.RuneCountInString(text); n > LiveMessageMaxLength {
		return "", fmt.Errorf("text is too long (%d > %d characters)", n, LiveMessageMaxLength)
	}
	if mode != model.LiveMessageModeReplace && mode != model.LiveMessageModeInfo {
		return "", fmt.Errorf("mode must be %s or %s", model.LiveMessageModeReplace, model.LiveMessageModeInfo)
	}
	return text, nil
}
