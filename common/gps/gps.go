package gps

import (
	"fmt"
	"math"
	"strings"

	"github.com/adrianmo/go-nmea"
	"github.com/tarm/serial"
)

type GPSRecord struct {
	UnixMicro uint64
	Lat       float64
	Long      float64
	Alt       float64
	Speed     float64
	Heading   float64
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

func Parse(data string) (GPSRecord, error) {
	data = strings.Trim(data, "\x00")
	data = strings.TrimRight(data, "\r\n")
	sentences := strings.Split(data, "\r\n")

	for i := range sentences {
		if len(sentences[i]) == 0 || sentences[i][0] != '$' {
			continue
		}
		s, err := nmea.Parse(sentences[i])
		if err != nil {
			// TODO can these be waited to finish/appended/prepended and parsed?
			// return GPSRecord{}, err
			continue
		}
		if s.DataType() == nmea.TypeGLL {
			m := s.(nmea.GLL)
			// fmt.Println("lat:", m.Latitude, "lon:", m.Longitude)
			return GPSRecord{
				Lat:  m.Latitude,
				Long: m.Longitude,
			}, nil
		} else if s.DataType() == nmea.TypeGGA {
			fmt.Println("alt:", s.(nmea.GGA).Altitude, "sats:", s.(nmea.GGA).NumSatellites)
		} else if s.DataType() == nmea.TypeVTG {
			fmt.Println("speed:", s.(nmea.VTG).GroundSpeedKPH, "heading:", s.(nmea.VTG).TrueTrack)
		} else {
			//fmt.Println("unknown sentence:", sentences[i])
		}
	}
	return GPSRecord{}, fmt.Errorf("no GLL sentence found")
}

type SerialRead func() (GPSRecord, error)
type SerialClose func() error

func StartSerial(serialPortPath string, baudrate int) chan GPSRecord {

	// look for the gps dongle and open it
	config := &serial.Config{
		Name:        serialPortPath,
		Baud:        baudrate,
		ReadTimeout: 1,
		Size:        8,
	}
	// open a serial port to the gps dongle
	port, err := serial.OpenPort(config) // TODO figure out how to close this
	// if it's not found, exit with an error
	if err != nil {
		//logrus.Error("Error opening serial port: ", err)
		return nil
	}
	grC := make(chan GPSRecord)
	go func(grC chan GPSRecord) {
		for {
			buf := make([]byte, 1024)
			nRead, err := port.Read(buf)
			if err != nil && !strings.Contains(err.Error(), "EOF") {
				// logrus.WithError(err).Error("Error reading serial port: ")
				continue
			}
			if nRead > 0 {
				// logrus.Info("Read from serial port: ", string(buf[:nRead]))
				gr, err := Parse(string(buf[:nRead]))
				if err != nil {
					// logrus.WithError(err).Error("Error parsing data from serial port: ")
					continue
				}
				grC <- gr
			}
		}
	}(grC)
	return grC
}
