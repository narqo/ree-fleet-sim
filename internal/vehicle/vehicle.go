package vehicle

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/narqo/ree-fleet-sim/internal/geoutil"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type VIN string

func VINFromString(s string) (VIN, error) {
	if s == "" {
		return "", fmt.Errorf("empty vin")
	}
	s = strings.ToUpper(s)
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'A' || c > 'Z') {
			return "", fmt.Errorf("malformed vin %q", s)
		}
	}
	return VIN(s), nil
}

func GenerateVIN() VIN {
	vin, _ := VINFromString(fmt.Sprintf("THE%010dVIN", rand.Int()))
	return vin
}

type Vehicle struct {
	client *FleetStateClient

	VIN VIN
	Lat float64
	Lon float64
}

func NewVehicle(client *FleetStateClient) *Vehicle {
	lat, lon := geoutil.RandLatLon()
	return VehicleInLatLon(client, lat, lon)
}

func VehicleInLatLon(client *FleetStateClient, lat, lon float64) *Vehicle {
	return &Vehicle{
		client: client,

		VIN: GenerateVIN(),
		Lat: lat,
		Lon: lon,
	}
}

func (vc *Vehicle) String() string {
	return fmt.Sprintf("Vehicle %s (%.6f %.6f)", vc.VIN, vc.Lat, vc.Lon)
}

func (vc *Vehicle) MoveNearby(d float64) {
	vc.Lat, vc.Lon = geoutil.RandLatLonNearby(vc.Lat, vc.Lon, d)
}

func (vc *Vehicle) ReportPosition(ctx context.Context) error {
	return vc.client.UpdatePosition(ctx, vc.VIN, vc.Lat, vc.Lon)
}
