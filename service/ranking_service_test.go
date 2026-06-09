package service

import (
	"github.com/go-playground/assert/v2"
	"github.com/swimresults/start-service/model"

	"testing"
)

func TestSetAgesForRanking(t *testing.T) {
	group := model.Ranking{
		MinAge: "2004",
		MaxAge: "2002",
		Ages:   nil,
		IsYear: true,
	}

	SetAgesForRanking(&group)

	assert.Equal(t, []int{2002, 2003, 2004}, group.Ages)
}

func TestSetAgesForRankingWrongOrder(t *testing.T) {
	group := model.Ranking{
		MinAge: "2002",
		MaxAge: "2004",
		Ages:   nil,
		IsYear: true,
	}

	SetAgesForRanking(&group)

	assert.Equal(t, []int{2002, 2003, 2004}, group.Ages)
}

func TestSetAgesForRankingOlder(t *testing.T) {
	group := model.Ranking{
		MinAge: "-1",
		MaxAge: "2002",
		Ages:   nil,
		IsYear: true,
	}

	var list []int
	for i := 2002; i <= 2100; i++ {
		list = append(list, i)
	}

	SetAgesForRanking(&group)

	assert.Equal(t, list, group.Ages)
}

func TestSetAgesForRankingYounger(t *testing.T) {
	group := model.Ranking{
		MinAge: "2002",
		MaxAge: "-1",
		Ages:   nil,
		IsYear: true,
	}

	var list []int
	for i := 1900; i <= 2002; i++ {
		list = append(list, i)
	}

	SetAgesForRanking(&group)

	assert.Equal(t, list, group.Ages)
}
