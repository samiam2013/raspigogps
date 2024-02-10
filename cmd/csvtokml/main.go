package main

// a simple command line tool to convert a given csv file to kml
//  and filtering out possibly bad data (zeros, impossible. etc)
//	in the process

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/samiam2013/raspigogps/common/gps"
)

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

	gpsDatum := make([]gps.GPSRecord, 0)
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
		gpsDatum = append(gpsDatum, gps.GPSRecord{
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

func captureWaypoints(data []gps.GPSRecord, secondsInterval uint64) []gps.GPSRecord {
	start := data[0].UnixMicro
	lastWaypointTime := start
	lastWaypointIdx := 0
	previousWPIdx := lastWaypointIdx
	filtered := make([]gps.GPSRecord, 1)
	for i, coord := range data {
		if coord.UnixMicro-(secondsInterval*1_000_000) > lastWaypointTime {
			filtered = append(filtered, coord)
			previousWPIdx = lastWaypointIdx
			lastWaypointTime = coord.UnixMicro
			lastWaypointIdx = i
		} else if data[previousWPIdx].Turned(data[lastWaypointIdx], coord) {
			filtered = append(filtered, coord)
			previousWPIdx = lastWaypointIdx
			lastWaypointTime = coord.UnixMicro
			lastWaypointIdx = i
		}
	}
	return filtered
}
