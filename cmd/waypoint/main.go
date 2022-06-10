package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adrianmo/go-nmea"
	"github.com/sirupsen/logrus"
	"github.com/tarm/serial" // TODO can this be replace by periphio?
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
)

func main() {
	if getProcessOwner() != "root" {
		log.Fatalf("Must be run as root. user given '%s'", getProcessOwner())
	}

	if _, err := host.Init(); err != nil {
		log.Fatalf("Failed to host.Init() for periphio: %s", err.Error())
	}

	// turn off white
	white := rpi.P1_7
	if err := white.Out(gpio.Low); err != nil {
		log.Fatalf("Failed to turn off white led: %s", err.Error())
	}
	// go func(led gpio.PinIO) {
	// 	outVal := gpio.High
	// 	for {
	// 		time.Sleep(time.Millisecond * 500)
	// 		outVal = !outVal
	// 		led.Out(outVal)
	// 	}
	// }(white)

	// overall objective:
	// using an engage button (side button)
	// when the button is pressed get the gps waypoint and print it with the time
	// and flash the number of the waypoint
	button := rpi.P1_33
	buttonLed := rpi.P1_29

	// Set it as input, with an internal pull down resistor:
	if err := button.In(gpio.PullDown, gpio.BothEdges); err != nil {
		logrus.WithError(err).Fatal("Could not set main button pull down and mode.")
	}

	sideButton := rpi.P1_35
	sideButtonLed := rpi.P1_31

	type engageFlag struct {
		mu   sync.Mutex
		flag bool
	}
	engage := engageFlag{}
	engage.mu.Lock()
	engage.flag = false
	engage.mu.Unlock()

	if err := sideButton.In(gpio.PullDown, gpio.BothEdges); err != nil {
		logrus.WithError(err).Fatal("Could not set side button pull down and mode.")
	}
	sideButtonLed.Out(gpio.Low)

	go func(e *engageFlag) {
		everyOther := true
		for {
			sideButton.WaitForEdge(-1)
			everyOther = !everyOther
			if everyOther {
				continue
			}
			e.mu.Lock()
			e.flag = !e.flag
			e.mu.Unlock()
			if e.flag {
				sideButtonLed.Out(gpio.High)
			} else {
				sideButtonLed.Out(gpio.Low)
			}
			// time delay
			time.Sleep(time.Second * 1)
		}
	}(&engage)

	gps := NewGPS()
	timeout := time.Second * 10
	waypointCount := 0
	lastWPTime := time.Now().Add(-1 * timeout)
	for {
		// if the current time is before the timeout period for the last waypoint
		if lastWPTime.Add(timeout).Before(time.Now()) {
			// sleep for the amount of time left
			time.Sleep(time.Until(lastWPTime.Add(timeout)))
		}
		button.WaitForEdge(-1)
		engage.mu.Lock()
		engaged := engage.flag
		engage.mu.Unlock()
		if !engaged {
			logrus.Error("Not engaged!")
			time.Sleep(time.Second * 10)
			continue
		}
		button.Read()

		waypointCount++
		if waypointCount%2 == 0 {
			continue
		}
		w, err := gps.GetWaypoint()
		if err != nil {
			logrus.WithError(err).Error("Couldn't get waypoint.")
		}
		lastWPTime = time.Now()
		actualCount := ((waypointCount + 1) / 2)
		fmt.Printf("%d,%f,%f,%d\n", w.UnixMicroTime, w.Latitude, w.Longitude, actualCount)
		// fmt.Printf("Waypoint count: %d\n", actualCount)
		inverse, err := time.ParseDuration(fmt.Sprintf("%fs", 1/math.Log10(float64(actualCount)*33)))
		if err != nil {
			logrus.WithError(err).Fatal("Failed to parse inverse duration")
		}
		for i := 0; i < actualCount; i++ {
			buttonLed.Out(gpio.High)
			time.Sleep(inverse)
			buttonLed.Out(gpio.Low)
			time.Sleep(inverse)
		}
	}
	gps.Close()
}

type GPS struct {
	Port      *serial.Port
	Waypoints []Waypoint
}

func NewGPS() GPS {
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
	return GPS{
		Port:      port,
		Waypoints: make([]Waypoint, 0),
	}
}

func (g *GPS) GetWaypoint() (Waypoint, error) {
	buf := make([]byte, 1024)
	retries := 3
	for {
		_, err := g.Port.Read(buf)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				continue
			}
			return Waypoint{}, err
		}
		waypoint, err := g.Parse(string(buf))
		if err != nil {
			if retries > 0 {
				retries--
				continue
			}
			return Waypoint{}, err
		}
		g.Waypoints = append(g.Waypoints, waypoint)
		return waypoint, nil
	}
}

func (g *GPS) Parse(data string) (Waypoint, error) {
	data = strings.Trim(data, "\x00")
	data = strings.TrimRight(data, "\r\n")
	sentences := strings.Split(data, "\r\n")

	for i := range sentences {
		if len(sentences[i]) == 0 || sentences[i][0] != '$' {
			continue
		}
		s, err := nmea.Parse(sentences[i])
		if err != nil {
			return Waypoint{}, err
		}
		if s.DataType() == nmea.TypeGLL {
			m := s.(nmea.GLL)
			// fmt.Println("lat:", m.Latitude, "lon:", m.Longitude)
			return Waypoint{
				Latitude:      m.Latitude,
				Longitude:     m.Longitude,
				UnixMicroTime: time.Now().UnixMicro(),
			}, nil
		}
	}
	return Waypoint{}, fmt.Errorf("could not parse data '%s'", data)
}

func (g *GPS) Close() error {
	return g.Port.Close()
}

type Waypoint struct {
	Longitude     float64
	Latitude      float64
	UnixMicroTime int64
}

func getProcessOwner() string {
	stdout, err := exec.Command("ps", "-o", "user=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return strings.Trim(string(stdout), "\n")
}
