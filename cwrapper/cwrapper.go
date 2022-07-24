package cwrapper

// #cgo CFLAGS: -g -Wall
// #include <stdlib.h>
// #include "ssh_c_lib.h"
import "C"
import (
	"fmt"
	"unsafe"
)

// LCD has a fixed width and height and adjustable left and right margins
type LCD struct {
	i2cPath    string
	i2cAddr    int
	cols       int
	rows       int
	leftMargin int
}

// NewLCD returns a new LCD struct with default values
func NewLCD(i2cPath string, addr int) LCD {
	return LCD{
		i2cPath:    i2cPath,
		i2cAddr:    addr,
		cols:       16,
		rows:       8,
		leftMargin: 1,
	}
}

// LCDInit initializes the LCD
func (l LCD) LCDInit() error {
	i2cPath := C.CString(l.i2cPath)
	defer C.free(unsafe.Pointer(i2cPath))
	if C.lcd_init(i2cPath, C.int(l.i2cAddr)) != 0 {
		return fmt.Errorf("could not initiale LCD, non-zero return code from C.lcd_init()")
	}
	return l.Clear()
}

// PrintAtPos prints a value at a position using the C.lcd_printc function
func (l LCD) PrintAtRowCol(value rune, row, col int) error {
	var position int
	if row > l.rows {
		return fmt.Errorf("row value %d is greater than LCD rows %d", row, l.rows)
	} else if col > (l.cols - l.leftMargin) {
		return fmt.Errorf("col value %d is greater than LCD cols %d minus left margin (%d)",
			col, l.cols, l.leftMargin)
	} else {
		// compute the actual position
		position = (row-1)*l.cols + col - l.leftMargin
	}

	if C.lcd_printc(C.char(value), C.int(position)) != 0 {
		return fmt.Errorf("error printing value %c at position %d", value, position)
	}
	return nil
}

// Clear empties the screen with C.lcd_clear()
func (l LCD) Clear() error {
	if C.lcd_clear() != 0 {
		return fmt.Errorf("call to C.lcd_clear returned non-zero code")
	}
	return nil
}

func (l LCD) Close() error {
	if C.lcd_close() != 0 {
		return fmt.Errorf("call to C.lcd_close returned non-zero code")
	}
	return nil
}
