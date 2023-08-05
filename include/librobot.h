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
size_t getFrame(uint8_t** buffer);
void cam_init();
void stop_cam_stream();

#ifdef __cplusplus
}
#endif

#endif // __LIBROBOT_H
