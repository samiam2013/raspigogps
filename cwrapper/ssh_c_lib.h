#ifndef _CINTERFACE_H
#define _CINTERFACE_H
int lcd_init(char *filename, int addr);
int lcd_close();
int lcd_move(int x, int y);
int lcd_clear();
int lcd_printc(const char chr, const int cursor);
int lcd_printmap(const char map[64][16]);
#endif

