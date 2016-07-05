package rgeogo

import (
	"fmt"
)

const (
	interleaveBoost   = []int64{0, 1, 4, 5, 16, 17, 20, 21, 64, 65, 68, 69, 80, 81, 84, 85}
	deinterleaveBoost = [][2]int64{[2]int64{0, 0}, [2]int64{0, 1}, [2]int64{1, 0}, [2]int64{1, 1}, [2]int64{0, 2}, [2]int64{0, 3}, [2]int64{1, 2}, [2]int64{1, 3}, [2]int64{2, 0}, [2]int64{2, 1}, [2]int64{3, 0}, [2]int64{3, 1}, [2]int64{2, 2}, [2]int64{2, 3}, [2]int64{3, 2}, [2]int64{3, 3}}
)

func encodeUInt64(latitude float64, longitude float64) (int64, error) {
	if latitude >= 90.0 || latitude < -90.0 {
		return -1, fmt.Errorf("Latitude must be in the range of (-90.0, 90.0)")
	}

	if longitude < -180 || longitude > 180 {
		return -1, fmt.Errorf("Longitude must be in the range of (-180.0, 180.0)")
	}

	lat := int64(((latitude + 90.0) / 180.0) * (1 << 32))
	lon := int64(((longitude + 180.0) / 360.0) * (1 << 32))

	return interleaveUInt64(lat, lon), nil
}

func interleaveUInt64(lat32 int64, lon32 int64) int64 {
	var intr int64

	for i := 0; i < 8; i++ {
		shift := uint64(28 - i*4)
		intr = (intr << 8) +
			(interleaveboost[(lon32>>shift)%16] << 1) +
			(interleaveboost[(lat32>>shift)%16])
	}

	return intr
}

func decodeUInt64(ui64 uint64) (float64, float64) {
	var lat, lon = deinterleaveUInt64(ui64)

	return float64(180.0*lat)/float64(1<<32) - 90.0, float64(360.0*lon)/float64(1<<32) - 180.0
}

func deinterleaveUInt64(ui64 uint64) (int64, int64) {
	var lat, lon int64

	for i := 0; i < 16; i++ {
		shift := uint64(60 - i*4)

		p := deinterleaveBoost[(ui64>>shift)%16]

		lat = (lat << 2) + p[1]
		lon = (lon << 2) + p[0]
	}

	return lat, lon
}
