package client

import (
	"fmt"
	"testing"
	"time"
)

func TestHeatClient_ImportHeat(t *testing.T) {
	client := NewHeatClient("http://localhost:8087/")
	tm, _ := time.Parse("2006-01-02 15:04", "2019-12-06 13:01")
	r, _, e := client.ImportHeat("IESC19", 13, 3, tm)
	if e != nil {
		fmt.Printf(e.Error())
	}
	fmt.Printf("id: %s, number: %d, start: %s", r.Identifier.String(), r.Number, r.StartEstimation.Format("15:04"))
}
