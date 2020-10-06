package geoutil

import (
	"math"
	"math/rand"
)

const (
	rad = math.Pi / 180
	// Earth radius in km
	earthKm = 6371
)

// Distance calculates haversine distance between two Lat, Lon coordinates.
// Refer to https://stackoverflow.com/questions/365826/calculate-distance-between-2-gps-coordinates
func Distance(lat0, lon0, lat1, lon1 float64) float64 {
	// convert degrees to radians
	dlat := (lat1 - lat0) * rad
	dlon := (lon1 - lon0) * rad
	lat0 = lat0 * rad
	lat1 = lat1 * rad
	a := 0.5 - math.Cos(dlat)/2 + math.Cos(lat0)*math.Cos(lat1)*(1-math.Cos(dlon))/2
	return 2 * earthKm * math.Asin(math.Sqrt(a))
}

// RandLatLon returns a random pair Lat, Lon
func RandLatLon() (lat, lon float64) {
	lat = rand.Float64()*180 - 90
	lon = rand.Float64()*360 - 180
	return
}

// RandLatLonNearby returns a random pair Lat, Lon nearby the point lat0, lon0, within distance in meters.
// Refer to https://gis.stackexchange.com/questions/25877/generating-random-locations-nearby
func RandLatLonNearby(lat0, lon0, distance float64) (lat1, lon1 float64) {
	r := distance / 111300 // convert meters to degrees
	w := r * math.Sqrt(rand.Float64())
	t := 2 * math.Pi * rand.Float64()
	x := w * math.Cos(t)
	y := w * math.Sin(t)
	lat1 = lat0 + y
	lon1 = lon0 + x/math.Cos(lat0)
	return
}
