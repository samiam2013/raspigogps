#include <stdint.h>
#include <errno.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <linux/i2c-dev.h>
#include <sys/ioctl.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include "font8x8/font8x8.h"

static int file;
const char init_[] = {0x0, 0xAE, 0x8D, 0x14, 0xAF}; // sequence to turn on the LCD ?

int init_i2c(const char *filename, int addr) {
    if ((file = open(filename, O_RDWR)) < 0) {
        printf("Failed to open the bus.");
        return -1;
    }

    if (ioctl(file, I2C_SLAVE, addr) < 0) {
        printf("Failed to acquire bus access and/or talk to slave.\n");
        return -1;
    }
    return 0;
}

int I2C_Write(const char *buf, size_t sz) {
    if (write(file, buf, sz) != sz) {
        printf("Failed to write to the i2c bus.\n");
        return -1;
    }
    return 0;
}

int lcd_init(char *filename, int addr) {
    if (init_i2c(filename, addr)) {
        return -1;
    }
    return I2C_Write(init_, 5); // sizeof init_ / sizeof init_[0]
}

int lcd_close() {
    return close(file);
}

int lcd_move(uint8_t x, uint8_t y) {
    char ret[4] = {0};
    x <<= 3;
    ret[1] = 0xb0 | y;
    ret[2] = x & 0xf;
    ret[3] = 0x10 | (x >> 4);
    return I2C_Write(ret, 4);
}

int lcd_clear() {
    char dat[9] = {0x40, 0};
    for (uint8_t r = 0; r < 8; r++) {
        if (lcd_move(0, r))
            return -1;
        for (uint8_t c = 0; c < 16; c++) {
            if (I2C_Write(dat, 9)) {
                return -1;
            }
        }
    }
    return lcd_move(0, 0);
}

int lcd_printc(const char chr, const int cursor) {
    if (lcd_move(cursor & 0xf, cursor >> 4)) {
        return -1;
    }
    char tmp[9] = {0x40, 0};
    for (size_t b = 0; b < 8; b++) {
        for (size_t bt = 0; bt < 8; bt++) {
            char bit = (font8x8_basic[chr & 0x7f][b] >> bt) & 1;
            tmp[bt + 1] |= bit << b;
        }
    }
    return I2C_Write(tmp, 9);
}

int lcd_printmap(const char map[64][16]) {
    for (uint8_t r = 0; r < 8; r++) {
        if (lcd_move(0, r)) {
            return -1;
        }
        for (uint8_t c = 0; c < 16; c++) {
            char dat[9] = {0x40, 0};
            for (size_t b = 0; b < 8; b++) {
                for (size_t bt = 0; bt < 8; bt++) {
                    char bit = (map[b + (r * 8)][c] >> bt) & 1;
                    dat[bt + 1] |= bit << b;
                }
            }
            if (I2C_Write(dat, 9)) {
                return -1;
            }
        }
    }
    return 0;
}
