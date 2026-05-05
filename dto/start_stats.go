package dto

type StartsByYearAndGenderStatsDto struct {
	Year    int                      `json:"year"`
	Genders []StartsByGenderStatsDto `json:"genders"`
}

type StartsByGenderStatsDto struct {
	Gender string `json:"gender"`
	Amount int    `json:"amount"`
}
