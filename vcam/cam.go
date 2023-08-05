package vcam

// #cgo LDFLAGS: -L${SRCDIR}/.. -lrobot
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
	"os"
	"time"
	"unsafe"
)

func InitCam() {
	go func() {
		C.cam_init()
	}()
	time.Sleep(time.Second * 2)
}

func StopCam() {
	C.stop_cam_stream()
}

func GetFrame() []byte {
	var cData *C.uint8_t
	cSize := C.getFrame(&cData)

	if cData == nil {
		fmt.Println("cdata is nil")
		return nil
	}

	// Convert C array to Go slice
	length := int(cSize)
	slice := (*[1 << 30]byte)(unsafe.Pointer(cData))[:length:length]

	// Create copy of data to preserve it after C memory is freed
	data := make([]byte, length)
	copy(data, slice)
	return data
}

func GetFrameAsJPEG() []byte {
	width := 1280
	height := 720
	var cData *C.uint8_t
	cSize := C.getFrame(&cData)

	if cData == nil {
		fmt.Println("cdata is nil")
		return nil
	}

	// Convert C array to Go slice
	length := int(cSize)
	slice := (*[1 << 30]byte)(unsafe.Pointer(cData))[:length:length]

	// Create copy of data to preserve it after C memory is freed
	data := make([]byte, length)
	copy(data, slice)
	img := image.NewYCbCr(image.Rect(0, 0, width, height), image.YCbCrSubsampleRatio420)
	copy(img.Y, data[:width*height])
	copy(img.Cb, data[width*height:(width*height)+((width/2)*(height/2))])
	copy(img.Cr, data[(width*height)+((width/2)*(height/2)):])

	// Convert the YCbCr image to RGB since JPEG requires RGB
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

func SaveAsJPEG(data []byte, width int, height int, path string) {
	// Construct image from the YUV420p data
	img := image.NewYCbCr(image.Rect(0, 0, width, height), image.YCbCrSubsampleRatio420)
	copy(img.Y, data[:width*height])
	copy(img.Cb, data[width*height:(width*height)+((width/2)*(height/2))])
	copy(img.Cr, data[(width*height)+((width/2)*(height/2)):])

	// Convert the YCbCr image to RGB since JPEG requires RGB
	imgRGBA := image.NewRGBA(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			col := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}
			imgRGBA.Set(x, y, col)
		}
	}

	// Create file
	outFile, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	// Write image to file
	err = jpeg.Encode(outFile, imgRGBA, nil)
	if err != nil {
		panic(err)
	}
}
