package client

import (
	"fmt"
	"testing"
)

func TestStartClient_ImportStart(t *testing.T) {
	var f float64
	f = 5.54
	fmt.Printf("%.f", f)
}

//
//import (
//	"fmt"
//	"github.com/swimresults/start-service/model"
//	"testing"
//	"time"
//)
//
//func TestStartClient_ImportStart(t *testing.T) {
//	client := NewStartClient("http://localhost:8087/")
//
//	start1 := model.Start{
//		Meeting: "IESC19",
//		Event:   5,
//		//AthleteName:     "Simon Meier",
//		//AthleteYear:     1234,
//		//AthleteTeamName: "blub Team",
//		Lane:       3,
//		HeatNumber: 11,
//		Rank:       2,
//		Certified:  true,
//	}
//
//	r, _, e := client.ImportStart(start1)
//	if e != nil {
//		fmt.Println(e)
//	}
//	fmt.Println(r)
//	//fmt.Printf("id: %s, number: %d, start: %s", r.Identifier.String(), r.Number, r.StartEstimation.Format("15:04"))
//}
//
//func TestStartClient_ImportResult(t *testing.T) {
//	client := NewStartClient("http://localhost:8087/")
//
//	start1 := model.Start{
//		Meeting: "IESC19",
//		Event:   5,
//		//AthleteName:     "Simon Meier",
//		//AthleteYear:     1234,
//		//AthleteTeamName: "blub Team",
//		Lane:       3,
//		HeatNumber: 11,
//	}
//
//	result1 := model.Result{
//		Time:       23100000,
//		ResultType: "final",
//		AddedAt:    time.Time{},
//	}
//
//	r, _, e := client.ImportResult(start1, result1)
//	if e != nil {
//		fmt.Println(e)
//	}
//	fmt.Println(r)
//	fmt.Printf("type: %s, time: %f", r.ResultType, r.Time.Seconds())
//}
