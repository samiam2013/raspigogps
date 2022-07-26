package main

import (
	"fmt"
	"time"

	"github.com/samiam2013/trakembox/common/gps"
	"github.com/samiam2013/trakembox/cwrapper"
)

func main() {
	recordChan := gps.StartSerial("/dev/ttyACM1", 9600)

	lcd := cwrapper.NewLCD("/dev/i2c-1", 0x3c)
	lcd.LCDInit()
	lcd.Clear()

	// start a goroutine to read the gps data or exit if there's an error
	latestUpdate := time.Now()
	for {
		gr := <-recordChan
		fmt.Printf("%+v\n", gr)
		if time.Since(latestUpdate) > time.Second {
			lcd.Clear()
			latestUpdate = time.Now()
			lat := fmt.Sprintf(" %3.6f", gr.Lat)
			for i := 1; i < len(lat)+1; i++ {
				lcd.PrintAtRowCol(rune(lat[i-1]), 1, i)
			}
			long := fmt.Sprintf(" %3.6f", gr.Long)
			for i := 1; i < len(long)+1; i++ {
				lcd.PrintAtRowCol(rune(long[i-1]), 2, i)
			}
			spd := fmt.Sprintf("  speed %3.1f", gr.Speed)
			for i := 0; i < len(spd); i++ {
				lcd.PrintAtRowCol(rune(spd[i]), 4, i)
			}
			alt := fmt.Sprintf("  alt %.1f", gr.Alt)
			for i := 0; i < len(alt); i++ {
				lcd.PrintAtRowCol(rune(alt[i]), 6, i)
			}
			hdg := fmt.Sprintf("  hdg %.1f", gr.Heading)
			for i := 0; i < len(hdg); i++ {
				lcd.PrintAtRowCol(rune(hdg[i]), 7, i)
			}
			sats := fmt.Sprintf("  sats %d", gr.NumSats)
			for i := 0; i < len(sats); i++ {
				lcd.PrintAtRowCol(rune(sats[i]), 8, i)
			}

		}
	}
	lcd.Close()
}
