package vehicle

import (
	"fmt"
	"testing"
)

func TestVINFromString(t *testing.T) {
	cases := []struct {
		in      string
		wantVIN VIN
		wantErr bool
	}{
		{
			in:      "THE1VIN",
			wantVIN: "THE1VIN",
		},
		{
			in:      "the1vin",
			wantVIN: "THE1VIN",
		},
		{
			in:      "",
			wantErr: true,
		},
		{
			in:      "THE-VIN",
			wantErr: true,
		},
		{
			in:      "THE1VIN_",
			wantErr: true,
		},
		{
			in:      "THE1VIN.html",
			wantErr: true,
		},
	}

	for n, tc := range cases {
		t.Run(fmt.Sprintf("case=%d", n), func(t *testing.T) {
			vin, err := VINFromString(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("want err %v got %v", tc.wantErr, err)
			}
			if vin != tc.wantVIN {
				t.Fatalf("want vin %v got %v", tc.wantVIN, vin)
			}
		})
	}
}
