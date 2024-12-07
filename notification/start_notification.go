package notification

import (
	"fmt"
	"github.com/swimresults/start-service/model"
	"time"
)

func SendStartNotificationForAthlete(start model.Start) {
	minutes := start.Heat.StartDelayEstimation.Sub(time.Now()).Minutes()
	fmt.Printf("notifying athlete for event: %d, heat: %d, lane: %d, athlete: %s in %f minutes\n", start.Event, start.HeatNumber, start.Lane, start.AthleteName, minutes)
	go func() {
		response, err := notificationClient.SendNotificationForMeetingAndAthlete(
			serviceKey,
			start.Meeting,
			start.Athlete,
			fmt.Sprintf("Wettkampf %d", start.Event),
			fmt.Sprintf("Du startest in ca. %f Minuten in Lauf %d auf Bahn %d.", minutes, start.HeatNumber, start.Lane),
			"athlete",
			"time-sensitive",
		)

		if err != nil {
			fmt.Printf("notify failed: %e\n", err)
		}
		println(response.Body)
		println(response.ApnsId)
	}()
}

func SendStartNotificationForFavourite(start model.Start) {
	minutes := start.Heat.StartDelayEstimation.Sub(time.Now()).Minutes()
	fmt.Printf("notifying favourite for event: %d, heat: %d, lane: %d, athlete: %s in %f minutes\n", start.Event, start.HeatNumber, start.Lane, start.AthleteName, minutes)
	go notificationClient.SendNotificationForMeetingAndAthlete(
		serviceKey,
		start.Meeting,
		start.Athlete,
		fmt.Sprintf("Wettkampf %d", start.Event),
		fmt.Sprintf("%s startet in ca. %f Minuten: Wettkampf in Lauf %d auf Bahn %d.", start.AthleteName, minutes, start.HeatNumber, start.Lane),
		"athlete",
		"time-sensitive",
	)
}
