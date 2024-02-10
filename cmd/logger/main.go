package main

import (
	"fmt"
	"os"

	"github.com/samiam2013/raspigogps/common/gps"
)

func main() {
	recordChan := gps.StartSerial("/dev/ttyACM0", 9600)

	// start a goroutine to read the gps data or exit if there's an error
	for {
		gr := <-recordChan
		fmt.Printf("%+v\n", gr)
	}
}
