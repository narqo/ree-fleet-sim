package geoutil

import (
	"fmt"
	"testing"
)

func TestDistance(t *testing.T) {
	cases := []struct {
		Lat0, Lon0 float64
		Lat1, Lon1 float64
		Want       float64
	}{
		{
			0, 0,
			0, 0,
			0,
		},
		{
			51.5, 0,
			38.8, -77.1,
			5918.185064088764,
		},
		{
			38.8, -77.1,
			51.5, 0,
			5918.185064088764,
		},
		{
			52.518898, 13.401797, // Berlin, Cathedral
			52.520645, 13.409779, // Berlin, Fernsehturm
			0.5739420180673658,
		},
	}

	for n, tc := range cases {
		t.Run(fmt.Sprintf("case=%d", n), func(t *testing.T) {
			d := Distance(tc.Lat0, tc.Lon0, tc.Lat1, tc.Lon1)
			if tc.Want != d {
				t.Fatalf("want %v got %v", tc.Want, d)
			}
		})
	}
}
