package gps

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/adrianmo/go-nmea"
	"github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

type GPSRecord struct {
	UnixMicro uint64
	Lat       float64
	Long      float64
	Alt       float64
	Speed     float64
	Heading   float64
	NumSats   int64
	TimeStr   string
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

	var gr GPSRecord
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
			gr.Lat = m.Latitude
			gr.Long = m.Longitude
			gr.TimeStr = m.Time.String()
		} else if s.DataType() == nmea.TypeGGA {
			// fmt.Println("alt:", s.(nmea.GGA).Altitude, "sats:", s.(nmea.GGA).NumSatellites)
			gr.Alt = s.(nmea.GGA).Altitude * 3.28084 // convert to feet
			gr.NumSats = s.(nmea.GGA).NumSatellites
		} else if s.DataType() == nmea.TypeVTG {
			// fmt.Println("speed:", s.(nmea.VTG).GroundSpeedKPH, "heading:", s.(nmea.VTG).TrueTrack)
			gr.Speed = s.(nmea.VTG).GroundSpeedKPH / 1.852 // convert to mph
			gr.Heading = s.(nmea.VTG).TrueTrack
		}
	}
	if gr.Lat == 0.0 || gr.Long == 0.0 {
		return GPSRecord{}, fmt.Errorf("no lat/long")
	}
	gr.UnixMicro = uint64(time.Now().UnixNano() / 1000)
	return gr, nil
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
		logrus.Error("Error opening serial port: ", err)
		return nil
	}
	grC := make(chan GPSRecord)
	go func(grC chan GPSRecord) {
		for {
			buf := make([]byte, 1024)
			delayedBuf := ""
			nRead, err := port.Read(buf)
			if err != nil && !strings.Contains(err.Error(), "EOF") {
				logrus.WithError(err).Error("Error reading serial port: ")
				continue
			}
			if nRead > 0 {
				delayedBuf += string(buf[:nRead])
				// clear the buffer
				buf = make([]byte, 1024)
				time.Sleep(time.Millisecond * 100) // wait for the next read
				nRead, _ = port.Read(buf)          // assuming if there was no error before there is none now :D
				// append the new data to the delayed buffer
				delayedBuf += string(buf[:nRead])
				logrus.Info("Read from serial port: ", delayedBuf)

				gr, err := Parse(delayedBuf)
				if err != nil {
					logrus.WithError(err).Error("Error parsing data from serial port: ")
					continue
				}
				grC <- gr
			}
		}
	}(grC)
	return grC
}
