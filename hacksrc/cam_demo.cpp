#include "cam.h"
#include "utils.h"
#include "librobot.h"
#include <fcntl.h> // for open()
#include <unistd.h> // for write() and close()
#include <dlfcn.h>
#include <pthread.h>
#include <stdint.h>
#include <arm_neon.h>

// Define mutex and buffer globally
pthread_mutex_t frameMutex = PTHREAD_MUTEX_INITIALIZER;
uint8_t* frameBuffer = NULL;
size_t frameBufferSize = 0;
mm_cam_lib_t lib;


int stop_cam;
bool writeFrame;

#define ION_IOC_INV_CACHES 0xC0144D01

int test(const char* t, int base, void* offset, int expected)
{
    int v = (int)offset - base;
    if (v != expected)
    {
        printf("%s: expected %d but got %d\n", t, expected, v);
        return 1;
    }
    return 0;
}

int testSz(const char* t, size_t sz, int expected)
{
    if (sz != expected)
    {
        printf("%s: expected %d but got %d\n", t, expected, sz);
        return 1;
    }
    return 0;
}

int validate()
{
    static mm_camera_lib_handle x;
    long base = (long)&x;
    if (testSz("cam_stream_info_t size", sizeof(cam_stream_info_t), 932)) return -1;
    if (testSz("mm_camera_buf_def_t size", sizeof(mm_camera_buf_def_t), 528)) return -1;
    if (testSz("mm_camera_app_meminfo_t size", sizeof(mm_camera_app_meminfo_t), 20)) return -1;
    if (testSz("mm_camera_app_buf_t size", sizeof(mm_camera_app_buf_t), 548)) return -1;
    if (testSz("mm_camera_lib_handle size", sizeof(mm_camera_lib_handle), 480320)) return -1;
    if (test("test_obj.channels[0].streams[0].offset", base, &x.test_obj.channels[0].streams[0].offset, 88)) return -1;
    if (test("test_obj.channels[0].streams[0].s_info_buf.buf.stream_type", base, &x.test_obj.channels[0].streams[0].s_info_buf.buf.stream_type, 13516)) return -1;
    if (test("test_obj.channels[0].streams[0].s_info_buf.buf.buf_type", base, &x.test_obj.channels[0].streams[0].s_info_buf.buf.buf_type, 13517)) return -1;
    if (test("test_obj.channels[0].streams[0].s_info_buf.buf.buf_idx", base, &x.test_obj.channels[0].streams[0].s_info_buf.buf.buf_idx, 13520)) return -1;

    if (test("test_obj.cap_buf.buf", base, &x.test_obj.cap_buf.buf, 392672)) return -1;

    if (test("test_obj.tune_data", base, &x.test_obj.tune_data, 395649)) return -1;
    if (test("tsctrl", base, &x.tsctrl, 480296)) return -1;
  
    return 0;
}

void mm_app_dump_frame(mm_camera_buf_def_t *frame,
                       const char *name,
                       const char *ext,
                       uint32_t frame_idx)
{
    char file_name[FILENAME_MAX];
    int file_fd;
    int i;
    int offset = 0;
    if ( frame != NULL) {
        for (i = 0; i < frame->planes_buf.num_planes; i++) {
            printf("%s: saving file from address: %p, data offset: %d, length: %d\n",  
                __func__, frame->buffer, frame->planes_buf.planes[i].data_offset, frame->planes_buf.planes[i].length);
            snprintf(file_name, sizeof(file_name), "/hack/%s_%04d_%02d.%s", name, frame_idx, i, ext);
            file_fd = open(file_name, O_RDWR | O_CREAT, 0777);
            if (file_fd < 0) {
                printf("%s: ERROR. cannot open file %s \n", __func__, file_name);
            } else {
                write(file_fd, (uint8_t *)frame->buffer + offset, frame->planes_buf.planes[i].length);
                offset += (int)frame->planes_buf.planes[i].length;
                close(file_fd);
                printf("dump %s\n", file_name);
            }
        }
    }
}

void mm_app_dump_frame2(mm_camera_buf_def_t *frame,
                       const char *name,
                       const char *ext,
                       uint32_t frame_idx)
{
    char file_name[FILENAME_MAX];
    int file_fd;
    int i;
    int offset = 0;
    if ( frame != NULL) {     
        snprintf(file_name, sizeof(file_name), "/hack/%s_%04d.%s", name, frame_idx, ext);
        file_fd = open(file_name, O_RDWR | O_CREAT, 0777);
        if (file_fd < 0) {
            printf("%s: ERROR. cannot open file %s \n", __func__, file_name);
        } else {
            for (i = 0; i < frame->planes_buf.num_planes; i++) {
                printf("%s: saving file from address: %p, data offset: %d, length: %d\n",  
                    __func__, frame->buffer, frame->planes_buf.planes[i].data_offset, frame->planes_buf.planes[i].length);
            
                write(file_fd, (uint8_t *)frame->buffer + offset, frame->planes_buf.planes[i].length);
                offset += (int)frame->planes_buf.planes[i].length;
            }
            close(file_fd);
            printf("dump %s\n", file_name);
        }
    }
}

unsigned short rgb888_rgb565(uint32_t sourceColor)
{
	uint32_t red = (sourceColor & 0x00FF0000) >> 16;
	uint32_t green = (sourceColor & 0x0000FF00) >> 8;
	uint32_t blue =  sourceColor & 0x000000FF;
	return (red >> 3 << 11) + (green >> 2 << 5) + (blue >> 3);
}

uint8_t clamp(uint8_t v, uint8_t minValue, uint8_t maxValue) {
    if (v < minValue) return minValue;
    else if (v > maxValue) return maxValue;
    else return v;
}
uint32_t yuv2rgb(uint8_t yValue, uint8_t uValue, uint8_t vValue) {
    uint8_t r = yValue + (1.370705 * (vValue-128));
    uint8_t g = yValue - (0.698001 * (vValue-128)) - (0.337633 * (uValue-128));
    uint8_t b = yValue + (1.732446 * (uValue-128));
    r = clamp(r, 0, 255);
    g = clamp(g, 0, 255);
    b = clamp(b, 0, 255);
    return  (r >> 3 << 11) + (g >> 2 << 5) + (b >> 3);
}

//https://android.googlesource.com/platform/frameworks/av/+/0e4e5a8/media/libstagefright/yuv/YUVImage.cpp

void mm_camera_app_notify_cb2(mm_camera_super_buf_t *bufs, void *user_data)
{
    mm_cam_lib_t *pme = (mm_cam_lib_t *) user_data;
    mm_camera_buf_def_t *frame = bufs->bufs[0];

    int i;
    int offset = 0;
    size_t totalLength = 0;

    for (i = 0; i < frame->planes_buf.num_planes; i++) {
        totalLength += frame->planes_buf.planes[i].length;
    }

    pthread_mutex_lock(&frameMutex);

    if (frameBuffer != NULL) {
        free(frameBuffer);
    }

    frameBuffer = (uint8_t*) malloc(totalLength);
    if (frameBuffer == NULL) {
        pthread_mutex_unlock(&frameMutex);
        printf("FRAMEBUFFER IS EMPTY");
        return;
    }

    frameBufferSize = totalLength;

    for (i = 0; i < frame->planes_buf.num_planes; i++) {
        memcpy(frameBuffer + offset, (uint8_t *)frame->buffer + offset, frame->planes_buf.planes[i].length);
        offset += (int)frame->planes_buf.planes[i].length;
    }

    pthread_mutex_unlock(&frameMutex);

    pme->handle.test_obj.cam->ops->qbuf(bufs->camera_handle, bufs->ch_id, frame);
    pme->mm_app_cache_ops((mm_camera_app_meminfo_t *)frame->mem_info, ION_IOC_INV_CACHES);

    mm_camera_app_done();
}

void convert_frame_to_rgb565(uint8_t* frame, uint16_t* buffer, int frameWidth, int frameHeight, int outputWidth, int outputHeight)
{
    int uOffset = frameWidth * frameHeight;
    int vOffset = uOffset + 1;
    int dx = frameWidth / outputWidth;
    int dy = frameHeight / outputHeight;

    for(int y = 0; y < frameHeight; y += dy) {
        for(int x = 0; x < frameWidth; x += dx) {
            int uvOffset = (y >> 1) * (frameWidth >> 1) + (x >> 1);
            int cY = frame[y * frameWidth + x];
            int cU = frame[uOffset + 2 * uvOffset];
            int cV = frame[vOffset + 2 * uvOffset];
            int i = x / dx;
            int j = y / dy;
            if (i < outputWidth && j < outputHeight) {
                buffer[j * outputWidth + i] = yuv2rgb(cY, cU, cV);
            }
        }
    }
}

size_t getFrame(uint8_t* goBuffer) {
    size_t size;

    pthread_mutex_lock(&frameMutex);
    
    if (frameBuffer != NULL && goBuffer != NULL) {
        memcpy(goBuffer, frameBuffer, frameBufferSize);
        size = frameBufferSize;
    } else {
        size = 0;
    }

    pthread_mutex_unlock(&frameMutex);

    return size;
}

void stop_cam_stream() {
    stop_cam = 1;
    mm_camera_app_stop(&lib);
}

void cam_init(int newWidth, int newHeight)
{

    //printf("%s: camera test\n", __func__);
    if (!validate()) printf("%s: validation - OK\n", __func__);
    else return;

    GPIO cenPin, sbyPin;
    gpioInit(&cenPin, 83, 1, 1); //poweron
    gpioInit(&sbyPin, 94, 1, 0); //unpause
    //printf("%s: gpio initialized\n", __func__);

    if (!lib_init(&lib))
    {
        int rc = mm_camera_app_start(&lib, 0, mm_camera_app_notify_cb2, newWidth, newHeight);
        if (!rc)
        {
            stop_cam = 0;
            //printf("%s: mm_camera_app_start: ok\n", __func__);
            //mm_camera_app_wait();
            while(true) {
                if (stop_cam == 1) {
                    break;
                }
            }
            //printf("%s: mm_camera_app_wait: done\n", __func__);
        }
    }
}

inline int clamp(int value, int min, int max) {
    return value < min ? min : (value > max ? max : value);
}

void convert_resize_yuv_to_rgba(uint8_t* data, uint8_t* rgba_data, int width, int height, int newWidth, int newHeight) {
    int size = width * height;
    int chromaSize = size / 4;

    uint8_t* Y = data;
    uint8_t* Cb = data + size;
    uint8_t* Cr = Cb + chromaSize;

    float x_ratio = (float)width / (float)newWidth;
    float y_ratio = (float)height / (float)newHeight;

    for(int y = 0; y < newHeight; y++) {
        int srcY = (int)(y * y_ratio);
        for(int x = 0; x < newWidth; x++) {
            int srcX = (int)(x * x_ratio);

            uint8_t yy = Y[srcY * width + srcX];
            uint8_t cb = Cb[(srcY / 2) * (width / 2) + (srcX / 2)];
            uint8_t cr = Cr[(srcY / 2) * (width / 2) + (srcX / 2)];

            int r = yy + 1.402 * (cr - 128);
            int g = yy - 0.344136 * (cb - 128) - 0.714136 * (cr - 128);
            int b = yy + 1.772 * (cb - 128);


            int destIdx = 4 * (y * newWidth + x);

            rgba_data[destIdx] = (uint8_t)clamp(r, 0, 255);
            rgba_data[destIdx + 1] = (uint8_t)clamp(g, 0, 255);
            rgba_data[destIdx + 2] = (uint8_t)clamp(b, 0, 255);
            rgba_data[destIdx + 3] = 255;
        }
    }
}


