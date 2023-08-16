package vcam

// #cgo LDFLAGS: -L${SRCDIR}/../build -lrobot
// #cgo CFLAGS: -I${SRCDIR}/../include
// #include "librobot.h"
// #include <dlfcn.h>
import "C"
import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"time"
	"unsafe"
)

var Cam_Initted bool

var width int
var height int

// w = desired width, h = desired height
func InitCam(w, h int) {
	width = w
	height = h
	go func() {
		C.cam_init(C.int(width), C.int(height))
	}()
	time.Sleep(time.Second)
	Cam_Initted = true
}

func StopCam() {
	C.stop_cam_stream()
	Cam_Initted = false
}

func GetFrame() []byte {
	if !Cam_Initted {
		fmt.Println("GetFrame(): you must InitCam first")
		return nil
	}
	data := make([]byte, 1382400)
	cSize := C.getFrame((*C.uint8_t)(&data[0]))
	data = data[:cSize]

	return data
}

// slowwwww
func GetFrameAsJPEG() []byte {
	data := GetFrame()
	img := image.NewYCbCr(image.Rect(0, 0, width, height), image.YCbCrSubsampleRatio420)
	copy(img.Y, data[:width*height])
	copy(img.Cb, data[width*height:(width*height)+((width/2)*(height/2))])
	copy(img.Cr, data[(width*height)+((width/2)*(height/2)):])
	imgRGBA := image.NewRGBA(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			col := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}
			imgRGBA.Set(x, y, col)
		}
	}
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, imgRGBA, nil)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func ConvertFrameToRGB565(frame []byte, frameWidth, frameHeight, outputWidth, outputHeight int) []uint16 {
	buffer := make([]uint16, outputWidth*outputHeight)

	C.convert_frame_to_rgb565((*C.uint8_t)(unsafe.Pointer(&frame[0])),
		(*C.uint16_t)(unsafe.Pointer(&buffer[0])), C.int(frameWidth),
		C.int(frameHeight), C.int(outputWidth), C.int(outputHeight))

	return buffer
}
