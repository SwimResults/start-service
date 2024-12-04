package notification

import (
	"fmt"
	"strconv"
)

func BroadcastHeatStart(meeting string, event int, heat int, delay int) {
	b := `
		{
			"status": {
				"event": ` + strconv.Itoa(event) + `,
				"heat": ` + strconv.Itoa(heat) + `,
				"over": false,
				"delay": ` + strconv.Itoa(delay) + `
			}
		}
	`
	notification, err := notificationClient.SendMeetingBroadcastNotification(serviceKey, meeting, b)
	if err != nil {
		fmt.Printf("failed sending meeting broadcast notification: %s\n", err.Error())
		return
	}

	println(notification)
}
