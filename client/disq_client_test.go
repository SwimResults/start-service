package client

//import (
//	"fmt"
//	"github.com/swimresults/start-service/model"
//	"testing"
//	"time"
//)
//
//func TestDisqualificationClient_ImportDisqualification(t *testing.T) {
//	client := NewDisqualificationClient("http://localhost:8087/")
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
//	r, _, e := client.ImportDisqualification(start1, "Guter Grund", time.Now())
//	if e != nil {
//		fmt.Println(e)
//	}
//	fmt.Println(r)
//	//fmt.Printf("id: %s, number: %d, start: %s", r.Identifier.String(), r.Number, r.StartEstimation.Format("15:04"))
//}
