package notification

import (
	"fmt"
)

type MeetingBroadcastData struct {
	Status MeetingBroadcastStatusData `json:"status"`
}

type MeetingBroadcastStatusData struct {
	Event int  `json:"event"`
	Heat  int  `json:"heat"`
	Over  bool `json:"over"`
	Delay int  `json:"delay"`
}

func BroadcastHeatStart(meeting string, event int, heat int, delay int) {
	status := MeetingBroadcastStatusData{
		Event: event,
		Heat:  heat,
		Over:  false,
		Delay: delay,
	}

	data := MeetingBroadcastData{status}

	notification, err := notificationClient.SendMeetingBroadcastNotification(serviceKey, meeting, data)
	if err != nil {
		fmt.Printf("failed sending meeting broadcast notification: %s\n", err.Error())
		return
	}

	println(notification)
}
