package vjpeg

/*
#cgo CFLAGS: -I../include -I../libjpeg-turbo/include
#cgo LDFLAGS: -L../libjpeg-turbo/lib -L../build -lturbojpeg -ljpeg_interface
#include "libjpeg_interface.h"
*/
import "C"
import "unsafe"

func Init() {
	C.loadLibTurboJpeg()
}

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
