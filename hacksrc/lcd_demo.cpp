#include "utils.h"
#include "lcd.h"
#include <fstream>
#include <sstream>
#include <string>

#define BUF_SZ 184*96
uint16_t buff[BUF_SZ];

#define LINE_SIZE 184
#define LINES 96

void tail_kernel_log(std::string& log) {
    std::ifstream file("/var/log/kern.log", std::ios_base::in | std::ios_base::ate);
    if(file.is_open()) {
        file.seekg(-1, std::ios_base::end); // go to one spot before the EOF

        // Keep going up in the file.
        for(int i = 0; i < LINES; i++) {
            while(file.peek() != '\n') {
                file.seekg(-2, std::ios_base::cur);
            }
            file.seekg(-2, std::ios_base::cur); // skip newline
        }
        std::getline(file, log);
    }
}

void lcd_demo()
{
    int spi = lcd_init();

    uint16_t colors[2][5] = {
        {0xF800, 0x7E0, 0x1F, 0xFFFF, 0},
        {0x1F, 0xF800, 0x7E0, 0xFFFF, 0}
    };

    int c=0;
    int i=0;
    for(c=0;c<5;c++)
    {
        std::string log;
        tail_kernel_log(log);
        std::istringstream logstream(log);
        std::string line;
        int line_num = 0;

        while(std::getline(logstream, line)) {
            for(int j=0; j<LINE_SIZE && j<line.size(); j++)
                buff[line_num*LINE_SIZE + j] = line[j]<LINE_SIZE/2 ? colors[0][c] : colors[1][c];
            line_num++;
        }

        lcd_write_data(spi, (char *)buff, BUF_SZ*2);
        delay(250);
    }
}
