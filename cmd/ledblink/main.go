package main

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

func main() {
	host.Init()
	white := gpioreg.ByName("7")
	green := gpioreg.ByName("29")
	blue := gpioreg.ByName("31")

	t := time.NewTicker(500 * time.Millisecond)
	for l := gpio.Low; ; l = !l {
		fmt.Println("Tick")
		white.Out(l)
		green.Out(!l)
		blue.Out(!l)
		<-t.C
	}
}
