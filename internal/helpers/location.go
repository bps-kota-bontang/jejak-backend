package helpers

import (
	"fmt"
	"math"
)

func CanonicalIDFromCoordinates(lat, lon float64) string {
	return fmt.Sprintf("%.3f,%.3f", lat, lon)
}

func RoundCoordinate(v float64) float64 {
	return math.Round(v*1000) / 1000
}

func HaversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusMeters = 6371000.0

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)
	lat1Rad := toRadians(lat1)
	lat2Rad := toRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}

func toRadians(deg float64) float64 {
	return deg * math.Pi / 180
}
