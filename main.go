package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/adrianmo/go-nmea"
	"github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

func main() {
	// look for the gps dongle and open it
	config := &serial.Config{
		Name:        "/dev/ttyACM0",
		Baud:        9600,
		ReadTimeout: 1,
		Size:        8,
	}
	// open a serial port to the gps dongle
	port, err := serial.OpenPort(config)
	// if it's not found, exit with an error
	if err != nil {
		logrus.WithError(err).Fatal("Could not open serial port")
	}
	defer port.Close()

	// start a channel to send the gps data to monitoring goroutines
	gpsData := make(chan string)
	errC := make(chan error)
	go func(gpsData chan string, errC chan error) {

		buf := make([]byte, 1024)
		for {
			_, err := port.Read(buf)
			if err != nil {
				if strings.Contains(err.Error(), "EOF") {
					continue
				}
				errC <- err
				return
			}
			gpsData <- string(buf)
		}
	}(gpsData, errC)

	exit := make(chan bool)
	saver, closer, err := makeSaver("gps.log")
	defer closer()
	if err != nil {
		logrus.WithError(err).Fatal("Could make saver function")
	}
	// start a goroutine to read the gps data or exit if there's an error
	go func(gpsData chan string, errC chan error, exit chan bool) {
		for {
			select {
			case err := <-errC:
				logrus.WithError(err).Fatal("Error reading serial port")
			case data := <-gpsData:
				display(data, errC, exit)
				parsed, err := parse(data)
				if err != nil {
					logrus.WithError(err).Error("Error monitoring NMEA GPS data")
					continue
				}
				err = saver(parsed.String())
				if err != nil {
					logrus.WithError(err).Fatal("Error saving GPS data")
				}

			}
		}
	}(gpsData, errC, exit)
	<-exit // wait for the exit signal from the running go routine (it's not coming)
}

func display(data string, errC chan error, exit chan bool) {
	fmt.Println("--BEGIN--\n" + data + "\n--END--")
	// display the gps data to a nokia 5510 display
}

type gpsData struct {
	latitude  float64
	longitude float64
}

func (g gpsData) String() string {
	return fmt.Sprintf("time: %v, latitude: %f, longitude: %f\n", time.Now(), g.latitude, g.longitude)
}

func parse(data string) (gpsData, error) {
	data = strings.Trim(data, "\x00")
	data = strings.TrimRight(data, "\r\n")
	sentences := strings.Split(data, "\r\n")

	for i := range sentences {
		if len(sentences[i]) == 0 || sentences[i][0] != '$' {
			continue
		}
		s, err := nmea.Parse(sentences[i])
		if err != nil {
			return gpsData{}, err
		}
		if s.DataType() == nmea.TypeGLL {
			m := s.(nmea.GLL)
			fmt.Println("lat:", m.Latitude, "lon:", m.Longitude)
			return gpsData{
				latitude:  m.Latitude,
				longitude: m.Longitude,
			}, nil
		}
	}
	return gpsData{}, nil
}

func makeSaver(filename string) (func(string) error, func(), error) {
	// open or create a file
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}
	return func(data string) error {
		_, err := f.WriteString(data)
		if err != nil {
			return err
		}
		return nil
	}, func() { f.Close() }, nil
}
