#include <cstdio>
#include <stdlib.h>
#include <turbojpeg.h>
#include "libjpeg_interface.h"
#include <vector>

void* tjInstance = NULL;

void loadLibTurboJpeg() {
    tjInstance = tjInitCompress();
    if (!tjInstance) {
        fprintf(stderr, "Error initializing TurboJPEG: %s\n", tjGetErrorStr());
        exit(1);
    }
}

void encodeToJPEG(unsigned char* yuvData, int width, int height, int quality, unsigned char** jpegBuf, unsigned long* jpegSize) {
    if (!tjInstance) {
        fprintf(stderr, "tjInstance does not exist\n");
        return;
    }

    // Confirming that width and height are divisible by 2
    if (width % 2 != 0 || height % 2 != 0) {
        fprintf(stderr, "Width and height must be divisible by 2 for TJSAMP_420\n");
        return;
    }

    // Using TJCS_YCbCr as the color space, padding the width
    if (tjCompressFromYUV(tjInstance, yuvData, width, 1, height, TJSAMP_420, jpegBuf, jpegSize, quality, TJFLAG_FASTDCT) != 0) {
        fprintf(stderr, "Error with TJ compression: %s\n", tjGetErrorStr());
    }
}


int clamp(int val, int min, int max) {
    return (val < min) ? min : (val > max) ? max : val;
}