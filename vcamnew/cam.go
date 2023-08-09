package main

/*
 #cgo LDFLAGS: -L${SRCDIR}/../build -lrobot -lanki-camera
 #cgo CFLAGS: -I${SRCDIR}/../include -I../anki/platform/camera/vicos/camera_client/ -I../anki
 #include "camera_client.h"
void mysleep(unsigned int microseconds) {
	usleep(microseconds);
}
*/
import "C"
import (
	"fmt"
	"os"
	"sync"
	"time"
	"unsafe"
)

var camera *C.struct_anki_camera_handle
var frameBuffer []byte
var frameBufferMutex sync.Mutex
var stopLooping bool
var readyForFrames bool

func InitCam() error {
	readyForFrames = false
	fmt.Println("Initializing Camera...")
	rc := C.camera_init(&camera)
	if rc != 0 {
		return fmt.Errorf("failed to initialize camera")
	}

	C.mysleep(100000)
	rc = C.camera_start(camera)
	if rc != 0 {
		return fmt.Errorf("failed to start camera")
	}

	stopLooping = false
	go func() {
		for C.camera_status(camera) != C.ANKI_CAMERA_STATUS_RUNNING && !stopLooping {
			C.mysleep(30000)
		}
		go func() {
			C.mysleep(100000)
			readyForFrames = true
		}()
		var r C.int
		for C.camera_status(camera) == C.ANKI_CAMERA_STATUS_RUNNING && !stopLooping {
			C.mysleep(60000)
			var frame *C.anki_camera_frame_t
			r = C.camera_frame_acquire(camera, 0, &frame)
			if r != 0 {
				continue
			}

			frameSize := int(frame.height) * int(frame.bytes_per_row)
			frameData := C.GoBytes(unsafe.Pointer(&frame.data), C.int(frameSize))

			frameBufferMutex.Lock()
			frameBuffer = make([]byte, frameSize)
			copy(frameBuffer, frameData)
			frameBufferMutex.Unlock()

			_ = C.camera_frame_release(camera, frame.frame_id)
		}
	}()

	return nil
}

func GetFrame() []byte {
	frameBufferMutex.Lock()
	defer frameBufferMutex.Unlock()
	return frameBuffer
}

func StopCam() error {
	readyForFrames = false
	stopLooping = true
	fmt.Println("Stopping Camera...")
	rc := C.camera_stop(camera)
	if rc != 0 {
		return fmt.Errorf("failed to stop camera")
	}

	rc = C.camera_release(camera)
	if rc != 0 {
		return fmt.Errorf("failed to release camera")
	}

	return nil
}

func main() {
	err := InitCam()
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		if !readyForFrames {
			time.Sleep(time.Millisecond * 50)
		} else {
			break
		}
	}

	frame := GetFrame()
	fmt.Println("Received Frame:", fmt.Sprint(len(frame)))
	os.WriteFile("/tmp/test.raw", frame, 0777)

	err = StopCam()
	if err != nil {
		fmt.Println(err)
	}
}
