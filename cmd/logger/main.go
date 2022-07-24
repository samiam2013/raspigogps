package main

import (
	"fmt"
	"os"
	"time"

	"github.com/samiam2013/trakembox/common/gps"
)

func main() {
	recordChan := gps.StartSerial("/dev/ttyACM0", 9600)

	// start a goroutine to read the gps data or exit if there's an error
	for {
		gr := <-recordChan
		fmt.Printf("%v, %v, %v\n", time.Now().UnixNano(), gr.Lat, gr.Long)
	}
}

func makeSaver(filename string) (func(string) error, func(), error) {
	// open or create a file
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}
	if stat, err := f.Stat(); err != nil {
		return nil, nil, err
	} else if stat.Size() == 0 {
		f.WriteString("Unix Micro Time, Lattitude, Longitude\n")
	}

	return func(data string) error {
		_, err := f.WriteString(data)
		if err != nil {
			return err
		}
		return nil
	}, func() { f.Close() }, nil
}
