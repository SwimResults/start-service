package service

import (
	"fmt"
	"github.com/swimresults/start-service/model"
	"time"
)

func StartNotificationMainLoop() {
	ticker := time.NewTicker(time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// search for heats that are
				//  - closer than 15 minutes to now() (start_delay_estimation)
				//  - not in the list of heats that has been notified about
				fmt.Printf("heat notification loop run startet:\n")

				heats, _ := GetHeatsWithStartTodayAndNoNotification()

				now := time.Now()
				then := now.Add(time.Minute * 15)

				fmt.Printf("found: %d heats to notify about\n", len(heats))

				var starts []model.Start
				for _, heat := range heats {
					if !(heat.StartDelayEstimation.After(now) && heat.StartDelayEstimation.Before(then)) {
						continue
					}

					heat, err := UpdateHeatNotifiedState(heat.Identifier, true)
					if err != nil {
						fmt.Printf("error updating heat m: %s, e: %d, h: %d\n", heat.Meeting, heat.Event, heat.Number)
						continue
					}

					heatStarts, _ := GetStartsByMeetingAndEventAndHeat(heat.Meeting, heat.Event, heat.Number)
					starts = append(starts, heatStarts...)
				}

				// notify own athlete
				for _, start := range starts {
					fmt.Printf("Dein Start in 15 Minuten: Wettkampf %d, Lauf: %d, Bahn: %d \n", start.Event, start.HeatNumber, start.Lane)
				}

				// notify favourites athlete
				for _, start := range starts {
					fmt.Printf("Dein Favorit '%s' startet in 15 Minuten: Wettkampf %d, Lauf: %d, Bahn: %d \n", start.AthleteName, start.Event, start.HeatNumber, start.Lane)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
