package dto

type LivestreamDto struct {
	Event  LivestreamEventDto   `json:"event,omitempty"`
	Heat   LivestreamHeatDto    `json:"heat,omitempty"`
	Starts []LivestreamStartDto `json:"starts,omitempty"`
}

type LivestreamEventDto struct {
	Number        int    `json:"number,omitempty"`
	Distance      int    `json:"distance,omitempty"`
	RelayDistance string `json:"relay_distance,omitempty"`
	Gender        string `json:"gender,omitempty"`
	Style         string `json:"style,omitempty"`
	Final         bool   `json:"final,omitempty"`
	Part          int    `json:"part,omitempty"`
}

type LivestreamHeatDto struct {
	Number int `json:"number,omitempty"`
	Max    int `json:"max,omitempty"`
}

type LivestreamStartDto struct {
	Lane         int `json:"lane,omitempty"`
	Time         int `json:"time,omitempty"`
	Registration int `json:"registration,omitempty"`
	Distance     int `json:"distance,omitempty"`
}

type LivestreamHeatStateDto struct {
	State string `json:"state,omitempty"`
}
