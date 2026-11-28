package service

// Live ticker: pushes the live status to the iOS Live Activity channel of each live meeting (via user-service).
// It runs every liveTickInterval for meetings in a live state with a push channel, and right after heat and
// message changes. A broadcast is only sent when its relevant part changed (liveBroadcastKey).

import (
	"fmt"
	"sync"
	"time"

	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/notification"
)

const liveTickInterval = 20 * time.Second

// liveBroadcastKey is the part of a live status that Live Activities show. Data freshness is left out on
// purpose: it changes all the time and isn't shown there.
type liveBroadcastKey struct {
	Type        dto.LiveStatusType
	StartsAt    time.Time // to the minute; apps show minutes
	Message     string
	MessageMode string
	Event, Heat int
	HasWarnings bool
}

func liveBroadcastKeyOf(s dto.LiveStatus) liveBroadcastKey {
	k := liveBroadcastKey{Type: s.Type, HasWarnings: len(s.Warnings) > 0}
	if s.StartsAt != nil {
		k.StartsAt = s.StartsAt.UTC().Truncate(time.Minute)
	}
	if s.Message != nil {
		k.Message, k.MessageMode = s.Message.Text, string(s.Message.Mode)
	}
	if s.Current != nil {
		k.Event, k.Heat = s.Current.Event, s.Current.Number
	}
	return k
}

// liveBroadcaster remembers the last broadcast per meeting (memory only: after a restart each live meeting
// gets one broadcast).
type liveBroadcaster struct {
	mu   sync.Mutex
	last map[string]liveBroadcastKey
	send func(meeting string, s dto.LiveStatus) error
}

func newLiveBroadcaster(send func(meeting string, s dto.LiveStatus) error) *liveBroadcaster {
	return &liveBroadcaster{last: map[string]liveBroadcastKey{}, send: send}
}

// update sends the status if its broadcast-relevant part changed. A failed send is retried on the next run.
func (b *liveBroadcaster) update(meeting string, s dto.LiveStatus) (bool, error) {
	key := liveBroadcastKeyOf(s)
	b.mu.Lock()
	prev, ok := b.last[meeting]
	b.mu.Unlock()
	if ok && prev == key {
		return false, nil
	}
	if err := b.send(meeting, s); err != nil {
		return false, err
	}
	b.mu.Lock()
	b.last[meeting] = key
	b.mu.Unlock()
	return true, nil
}

var liveTrigger = make(chan string, 64)

// triggerLiveBroadcast asks the ticker to check a meeting now (non-blocking; dropped if the queue is full,
// the next tick catches up).
func triggerLiveBroadcast(meeting string) {
	select {
	case liveTrigger <- meeting:
	default:
	}
}

// liveHeatsChanged is called after a heat of the meeting changed.
func liveHeatsChanged(meeting string) {
	invalidateLiveHeats(meeting)
	triggerLiveBroadcast(meeting)
}

func broadcastLive(b *liveBroadcaster, meeting string) {
	s, err := GetLiveStatus(meeting)
	if err != nil {
		fmt.Printf("live broadcast %s: status: %v\n", meeting, err)
		return
	}
	sent, err := b.update(meeting, s)
	if err != nil {
		fmt.Printf("live broadcast %s: %v\n", meeting, err)
	} else if sent {
		fmt.Printf("live broadcast %s: %s\n", meeting, s.Type)
	}
}

// StartLiveTicker starts the broadcast loop; it needs user-service and meeting-service.
func StartLiveTicker() {
	if !notification.Enabled() || eventClient == nil {
		fmt.Println("live ticker not started: SR_START_USER_URL and SR_START_MEETING_URL are needed")
		return
	}
	b := newLiveBroadcaster(notification.BroadcastLiveStatus)
	go func() {
		ticker := time.NewTicker(liveTickInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for _, meeting := range getLiveBroadcastMeetings() {
					broadcastLive(b, meeting)
				}
			case meeting := <-liveTrigger:
				if isLiveBroadcastMeeting(getMeetingCached(meeting)) {
					broadcastLive(b, meeting)
				}
			}
		}
	}()
}
