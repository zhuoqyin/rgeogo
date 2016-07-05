package rgeogo

import (
	"fmt"
	"math"
)

const (
	EarthRadius = 6378.1 //km

	degreesToRadians = math.Pi / 180.0
)

var rgeoObject = &rgeo{
	M: make(map[int64]geo),
	I: int64Slice{},
}

func RGeocode(lat, lon float64, sample int) *geo {
	if lat >= 90.0 || lat < -90.0 || lon < -180 || lon > 180 {
		return nil
	}

	// get desired candidate points
	index := float64(rgeoObject.I.Search(encodeUInt64(lat, lon)))

	var candidates []int64
	for i := math.Max(0, index-float64(sample)/2); i < math.Min(float64(len(rgeoObject.I)), index+float64(sample)/2); i++ {
		candidates = append(candidates, rgeoObject.I[int(i)])
	}

	var result int64
	var minDist = 4.0 //  range of arccos is [0,pi], set this to 4 to represent +inf
	for _, candidate := range candidates {
		latC, lonC := decodeUInt64(uint64(candidate))

		dist := DistanceOnUnitSphere(lat, lon, latC, lonC)
		if dist < minDist {
			result = candidate
			minDist = dist
		}
	}

	if g, ok := rgeoObject.M[result]; ok && DistanceOnUnitSphere(lat, lon, g.PostalCode.Lat, g.PostalCode.Lon)*Earth_Radius < 50 {
		return &g
	} else {
		return nil
	}
}

func DistanceOnUnitSphere(lat1, long1, lat2, long2 float64) float64 {
	// phi = 90 - latitude
	phi1 := (90.0 - lat1) * degreesToRadians
	phi2 := (90.0 - lat2) * degreesToRadians

	// theta = longitude
	theta1 := long1 * degreesToRadians
	theta2 := long2 * degreesToRadians

	// Compute spherical distance from spherical coordinates.

	// For two locations in spherical coordinates
	// (1, theta, phi) and (1, theta', phi')
	// cosine( arc length ) =
	//  sin phi sin phi' cos(theta-theta') + cos phi cos phi'
	// distance = rho * arc length

	cos := (math.Sin(phi1)*math.Sin(phi2)*math.Cos(theta1-theta2) + math.Cos(phi1)*math.Cos(phi2))

	return math.Acos(cos)
}
