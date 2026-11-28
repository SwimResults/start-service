package service

import (
	"errors"
	"testing"
	"time"

	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
	"github.com/swimresults/swimresults-api/go/meetingapi"
)

func TestLiveBroadcasterChangeDetection(t *testing.T) {
	var sent []dto.LiveStatus
	fail := false
	b := newLiveBroadcaster(func(_ string, s dto.LiveStatus) error {
		if fail {
			return errors.New("user-service down")
		}
		sent = append(sent, s)
		return nil
	})
	now := berlin(2026, 11, 28, 10, 3)
	cur := &model.Heat{Event: 5, Number: 3}
	base := dto.LiveStatus{Type: dto.LiveStatusHeat, Current: cur, ServerTime: now}

	step := func(name string, s dto.LiveStatus, wantSent bool) {
		t.Helper()
		before := len(sent)
		ok, err := b.update("M", s)
		if ok != wantSent || (len(sent) > before) != wantSent {
			t.Errorf("%s: sent=%v err=%v, want sent=%v", name, ok, err, wantSent)
		}
	}

	step("first status", base, true)
	step("nothing changed", base, false)

	fresh := base
	fresh.ServerTime = now.Add(time.Minute)
	last := now
	fresh.LastDataAt = &last
	fresh.DataDelayed = true
	step("only time and data freshness changed", fresh, false)

	nextHeat := base
	nextHeat.Current = &model.Heat{Event: 5, Number: 4}
	step("next heat", nextHeat, true)

	warn := nextHeat
	warn.Warnings = []string{"x"}
	step("warnings appear", warn, true)
	warn2 := warn
	warn2.Warnings = []string{"x", "y"}
	step("more warnings, still has warnings", warn2, false)

	msg := warn2
	msg.Message = &model.LiveMessage{Text: "Pause", Mode: model.LiveMessageModeInfo}
	step("message set", msg, true)
	msgMode := msg
	msgMode.Message = &model.LiveMessage{Text: "Pause", Mode: model.LiveMessageModeReplace}
	step("message mode changed", msgMode, true)

	brk := msgMode
	brk.Type = dto.LiveStatusBreak
	t1 := berlin(2026, 11, 28, 14, 0)
	brk.StartsAt = &t1
	step("break", brk, true)
	brkSec := brk
	t2 := t1.Add(20 * time.Second)
	brkSec.StartsAt = &t2
	step("starts_at moved by seconds", brkSec, false)
	brkMin := brk
	t3 := t1.Add(time.Minute)
	brkMin.StartsAt = &t3
	step("starts_at moved by a minute", brkMin, true)

	fail = true
	end := brkMin
	end.Type = dto.LiveStatusEnd
	if ok, err := b.update("M", end); ok || err == nil {
		t.Errorf("failed send: sent=%v err=%v", ok, err)
	}
	fail = false
	step("retry after failure", end, true)

	if ok, _ := b.update("OTHER", end); !ok {
		t.Error("meetings are tracked separately")
	}
}

func TestIsLiveBroadcastMeeting(t *testing.T) {
	m := func(state, channel string) *meetingapi.Meeting {
		return &meetingapi.Meeting{State: state, Data: meetingapi.MeetingData{PushNotificationChannel: channel}}
	}
	for _, c := range []struct {
		m    *meetingapi.Meeting
		want bool
	}{
		{nil, false},
		{m("RUNNING", "ch"), true},
		{m("BREAK", "ch"), true},
		{m("PAUSE", "ch"), true},
		{m("RUNNING", ""), false},
		{m("FINISHED", "ch"), false},
		{m("OPENED", "ch"), false},
	} {
		if got := isLiveBroadcastMeeting(c.m); got != c.want {
			t.Errorf("%+v: got %v", c.m, got)
		}
	}
}

func TestTriggerLiveBroadcastDoesNotBlock(t *testing.T) {
	for i := 0; i < cap(liveTrigger)+10; i++ {
		triggerLiveBroadcast("M")
	}
	for len(liveTrigger) > 0 {
		<-liveTrigger
	}
}
