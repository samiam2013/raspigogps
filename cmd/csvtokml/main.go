package main

// a simple command line tool to convert a given csv file to kml
//  and filtering out possibly bad data (zeros, impossible. etc)
//	in the process

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
)

type gpsRecord struct {
	UnixMicro uint64
	Lat       float64
	Long      float64
}

func main() {
	// get the file argument
	var filepath string
	flag.StringVar(&filepath, "file", "gps.log", "Path to the file to be converted from CSV to KML")
	flag.Parse()
	// open file
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Couldn't open file: %s", err.Error())
	}
	defer f.Close()

	csvR := csv.NewReader(f)
	data, err := csvR.ReadAll()
	if err != nil {
		log.Fatalf("Couldn't read in data from gps log file: %s", err.Error())
	}
	// peel off the header
	data = data[1:]

	gpsDatum := make([]gpsRecord, 0)
	for _, row := range data {
		unixMicroTime, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			log.Fatalf("Could not parse time: %s", err.Error())
		}
		lat, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			log.Fatalf("Could not parse lattitude: %s", err.Error())
		}
		long, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			log.Fatalf("Could not parse longitude: %s", err.Error())
		}
		if lat == 0.0 || lat == long {
			//      log.Print("Zeroes spotted in the data, skipping")
			continue
		}
		gpsDatum = append(gpsDatum, gpsRecord{
			UnixMicro: uint64(unixMicroTime),
			Lat:       lat,
			Long:      long,
		})
	}

	gpsDatum = captureWaypoints(gpsDatum, 10)

	for i, gpsWaypoint := range gpsDatum {
		if i%100 == 0 {
			fmt.Printf("%v\n", gpsWaypoint)
		}
	}

}

func captureWaypoints(data []gpsRecord, secondsInterval uint64) []gpsRecord {
	start := data[0].UnixMicro
	lastWaypointTime := start
	lastWaypointIdx := 0
	previousWPIdx := lastWaypointIdx
	filtered := make([]gpsRecord, 1)
	for i, coord := range data {
		if coord.UnixMicro-(secondsInterval*1_000_000) > lastWaypointTime {
			filtered = append(filtered, coord)
			previousWPIdx = lastWaypointIdx
			lastWaypointTime = coord.UnixMicro
			lastWaypointIdx = i
		} else if turned(data[previousWPIdx], data[lastWaypointIdx], coord) {
			filtered = append(filtered, coord)
			previousWPIdx = lastWaypointIdx
			lastWaypointTime = coord.UnixMicro
			lastWaypointIdx = i
		}
	}
	return filtered
}

func turned(previous, from, to gpsRecord) bool {
	oldAngle := getUnitCirAngle(previous, from)
	newAngle := getUnitCirAngle(from, to)
	turned := math.Abs(oldAngle-newAngle) > 5.0
	if turned {
		fmt.Printf("Turned. old angle %f, new angle %f\n", oldAngle, newAngle)
	}
	return turned
}

// to get the actual heading spin 90 degrees counterclockwise
func getUnitCirAngle(from, to gpsRecord) float64 {
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
