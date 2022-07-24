package gps

import (
	"fmt"
	"math"
)

type GPSRecord struct {
	UnixMicro uint64
	Lat       float64
	Long      float64
}

// to get the actual heading spin 90 degrees counterclockwise
func getUnitCirAngle(from, to GPSRecord) float64 {
	// handle the edge case of heading directly west or east
	if to.Long == from.Long {
		if to.Lat == from.Lat {
			return 0.0
		} else if to.Lat > from.Lat {
			return 90.0 // straight north
		} else {
			return 270.0 // straight south
		}
	} else if to.Long > from.Long {
		// if the to longitude is higher, it's headed right (east)
		if to.Lat > from.Lat {
			// if the lattitude is higher, it's headed up (north)
			slope := (to.Lat - from.Lat) / (from.Long - to.Long)
			angle := ((math.Atan(slope) / (math.Pi * 2)) * 360.0) * -1.0
			fmt.Printf("angle radians: %f; degrees: %f;\n", math.Atan(slope), angle)
			return angle
		} else {
			// it's headed down (south)
			slope := (from.Lat - to.Lat) / (from.Long - to.Long)
			angle := ((math.Atan(slope) / (math.Pi * 2)) * 360.0) + 360
			fmt.Printf("angle: %f degrees\n", angle)
			return angle
		}
	} else {
		// if the longitude is lower, it's headed left (west)
		if to.Lat > from.Lat {
			// if the lattitude is higher, it's headed up (north)
			slope := (to.Lat - from.Lat) / (to.Long - from.Long)
			angle := 180 + ((math.Atan(slope) / (math.Pi * 2)) * 360.0)
			fmt.Printf("angle radians: %f; degrees: %f;\n", math.Atan(slope), angle)
			return angle
		} else {
			// it's headed down (south)
			slope := (from.Lat - to.Lat) / (to.Long - from.Long)
			angle := 180 + (((math.Atan(slope) / (math.Pi * 2)) * 360.0) * -1.0)
			fmt.Printf("angle: %f degrees\n", angle)
			return angle
		}
	}
}

func (g GPSRecord) Turned(from, to GPSRecord) bool {
	oldAngle := getUnitCirAngle(g, from)
	newAngle := getUnitCirAngle(from, to)
	turned := math.Abs(oldAngle-newAngle) > 5.0
	if turned {
		fmt.Printf("Turned. old angle %f, new angle %f\n", oldAngle, newAngle)
	}
	return turned
}
