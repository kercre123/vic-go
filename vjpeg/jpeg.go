package vjpeg

/*
#cgo CFLAGS: -I../include -I../libjpeg-turbo/include
#cgo LDFLAGS: -L../libjpeg-turbo/lib -L../build -lturbojpeg -ljpeg_interface
#include "libjpeg_interface.h"
*/
import "C"
import "unsafe"

func EncodeToJPEG(yuvData []byte, quality int, width int, height int) []byte {
	var jpegSize C.ulong
	var jpegBuf *C.uchar // Pointer to the allocated buffer

	C.encodeToJPEG((*C.uchar)(&yuvData[0]), C.int(width), C.int(height), C.int(quality), &jpegBuf, &jpegSize)

	// Create a Go slice backed by the C array
	goSlice := C.GoBytes(unsafe.Pointer(jpegBuf), C.int(jpegSize))

	// Free the allocated buffer
	C.free(unsafe.Pointer(jpegBuf))

	return goSlice
}

// debayer first
func RGGB10ToJPEGDownSample(rawData []byte, quality int) []byte {
	width := 1280
	height := 720
	var jpegSize C.ulong
	var jpegBuf *C.uchar

	C.GetFrameAsJPEGDownSampled((*C.uint8_t)(unsafe.Pointer(&rawData[0])), C.int(width), C.int(height), C.int(quality), &jpegBuf, &jpegSize)

	jpegData := C.GoBytes(unsafe.Pointer(jpegBuf), C.int(jpegSize))
	C.free(unsafe.Pointer(jpegBuf))

	return jpegData
}
