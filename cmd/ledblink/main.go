package logger

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
)

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatalf("Failed to host.Init() for periphio: %s", err.Error())
	}
	white := rpi.P1_7
	blue := rpi.P1_29
	green := rpi.P1_31
	engage := rpi.P1_33
	if err := green.Out(false); err != nil {
		log.Printf("Could not set greenpin low: %s\n", err.Error())
	}
	if err := blue.Out(false); err != nil {
		log.Printf("Could not set blue pin low: %s\n", err.Error())
	}
	if err := white.Out(false); err != nil {
		log.Printf("Could not set white pin low: %s\n", err.Error())
	}

	t := time.NewTicker(1000 * time.Millisecond)
	for l := gpio.Low; ; l = !l {
		fmt.Println("Tick")
		if engage.Read() {
			green.Out(false)
			white.Out(false)
			blue.Out(false)
			<-t.C
			continue
		}
		if err := green.Out(l); err != nil {
			log.Fatalf("Could not write to pin: %s", err.Error())
		}
		time.Sleep(200 * time.Millisecond)
		_ = white.Out(l) // TODO handle error
		time.Sleep(300 * time.Millisecond)
		_ = blue.Out(l) // TODO handle error

		<-t.C
	}
}
