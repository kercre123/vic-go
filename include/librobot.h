#ifndef __LIBROBOT_H
#define __LIBROBOT_H

#include <stdint.h>
#include "spine.h"

#ifdef __cplusplus
extern "C" {
#endif

int spine_full_init();
void close_spine();
void spine_full_update(uint32_t seq, int16_t* motors_data, uint32_t* leds_data);
spine_dataframe_t iterate();
size_t getFrame(uint8_t* goBuffer);
void init_lcd();
void set_pixels(uint16_t *pixels);
void cam_init(int newWidth, int newHeight);
void stop_cam_stream();
void convert_frame_to_rgb565(uint8_t* frame, uint16_t* buffer, int frameWidth, int frameHeight, int outputWidth, int outputHeight);
void debayerRGGB(uint8_t* rawData, int width, int height, uint8_t* outputData);


#ifdef __cplusplus
}
#endif

#endif // __LIBROBOT_H
