package notification

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/swimresults/start-service/dto"
)

// LiveBroadcastMessageMaxLength keeps the broadcast well below the 4 KB APNs payload limit.
const LiveBroadcastMessageMaxLength = 200

// MeetingBroadcastData is the content-state of the iOS Live Activity. Installed app versions decode
// `status` with Codable, so it must stay unchanged; `live` is additive.
type MeetingBroadcastData struct {
	Status MeetingBroadcastStatusData `json:"status"`
	Live   *MeetingBroadcastLiveData  `json:"live,omitempty"`
}

type MeetingBroadcastStatusData struct {
	Event int  `json:"event"`
	Heat  int  `json:"heat"`
	Over  bool `json:"over"`
	Delay int  `json:"delay"`
}

type MeetingBroadcastLiveData struct {
	Type        dto.LiveStatusType `json:"type"`
	StartsAt    *time.Time         `json:"starts_at,omitempty"`
	Message     string             `json:"message,omitempty"`
	MessageMode string             `json:"message_mode,omitempty"`
	Warnings    int                `json:"warnings"`
}

// Enabled reports whether user-service is configured, i.e. broadcasts can be sent.
func Enabled() bool { return notificationClient != nil }

func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max-1]) + "…"
}

// LiveBroadcastData builds the Live Activity content-state from a live status.
func LiveBroadcastData(s dto.LiveStatus) MeetingBroadcastData {
	var status MeetingBroadcastStatusData
	if s.Current != nil {
		status.Event = s.Current.Event
		status.Heat = s.Current.Number
		if !s.Current.StartAt.IsZero() {
			// same sign as the former heat start broadcast: negative when the heat started late
			status.Delay = int(s.Current.StartEstimation.Sub(s.Current.StartAt).Seconds())
		}
	}
	status.Over = s.Type == dto.LiveStatusEnd

	live := &MeetingBroadcastLiveData{Type: s.Type, StartsAt: s.StartsAt, Warnings: len(s.Warnings)}
	if s.Message != nil {
		live.Message = truncateRunes(s.Message.Text, LiveBroadcastMessageMaxLength)
		live.MessageMode = string(s.Message.Mode)
	}
	return MeetingBroadcastData{Status: status, Live: live}
}

// BroadcastLiveStatus sends the live status to the meeting's Live Activity channel (via user-service).
func BroadcastLiveStatus(meeting string, s dto.LiveStatus) error {
	if notificationClient == nil {
		return errors.New("user-service not configured")
	}
	_, err := notificationClient.SendMeetingBroadcastNotification(meeting, LiveBroadcastData(s))
	return err
}
