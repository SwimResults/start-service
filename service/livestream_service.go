package service

import (
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
)

func GetLivestreamData(meeting string) (*dto.LivestreamDto, error) {
	heat, err := GetCurrentHeat(meeting)
	if err != nil {
		return nil, err
	}

	heatAmount, err := GetHeatsAmountByMeetingAndEvent(meeting, heat.Event)
	if err != nil {
		return nil, err
	}

	starts, err := GetStartsByMeetingAndEventAndHeat(meeting, heat.Event, heat.Number)
	if err != nil {
		return nil, err
	}

	event, err := eventClient.GetEventByMeetingAndNumber(meeting, heat.Event)
	if err != nil {
		return nil, err
	}

	livestreamEvent := dto.LivestreamEventDto{
		Number:        event.Number,
		Distance:      event.Distance,
		RelayDistance: event.RelayDistance,
		Gender:        event.Gender,
		Style:         event.Style.Name,
		Final:         event.Final.IsFinal,
		Part:          event.Part.Number,
	}

	livestreamHeat := dto.LivestreamHeatDto{
		Number: heat.Number,
		Max:    heatAmount,
	}

	var livestreamStarts []dto.LivestreamStartDto

	for _, start := range starts {
		var mostRecentResult model.Result
		var registration model.Result

		if len(start.Results) > 0 {
			mostRecentResult = start.Results[len(start.Results)-1]
			registration = start.Results[0]
		}

		livestreamStart := dto.LivestreamStartDto{
			Lane:         start.Lane,
			Time:         int(mostRecentResult.Time.Milliseconds()),
			Registration: int(registration.Time.Milliseconds()),
			Distance:     mostRecentResult.LapMeters,
		}

		livestreamStarts = append(livestreamStarts, livestreamStart)
	}

	livestreamData := dto.LivestreamDto{
		Event:  livestreamEvent,
		Heat:   livestreamHeat,
		Starts: livestreamStarts,
	}

	return &livestreamData, nil
}

func GetLivestreamHeatState(meeting string) (*dto.LivestreamHeatStateDto, error) {
	heat, err := GetCurrentHeat(meeting)
	if err != nil {
		return nil, err
	}

	state := ""
	if heat.FinishedAt.IsZero() {
		state = "running"
	} else {
		state = "finished"
	}

	heatState := dto.LivestreamHeatStateDto{
		State: state,
	}
	return &heatState, nil
}
