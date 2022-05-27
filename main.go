package main

import (
	"fmt"
	"strings"

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
	// start a goroutine to read the gps data or exit if there's an error
	go func(gpsData chan string, errC chan error, exit chan bool) {
		for {
			select {
			case err := <-errC:
				logrus.WithError(err).Fatal("Error reading serial port")
			case data := <-gpsData:
				fmt.Println("--BEGIN--\n" + data + "\n--END--")
			}
		}
	}(gpsData, errC, exit)
	<-exit // wait for the exit signal from the running go routine (it's not coming)
	// display the gps data to a nokia 5510 display
	// check the location and a switch for an audible alarm
	// log the gps data to a file

}
