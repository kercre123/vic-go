package vcamnew

/*
 #cgo LDFLAGS: -L${SRCDIR}/../build -lrobot -lanki-camera
 #cgo CFLAGS: -I${SRCDIR}/../include -I../anki/platform/camera/vicos/camera_client/ -I../anki
 #include "camera_client.h"
 #include "librobot.h"
*/
import "C"
import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"
)

var camera *C.struct_anki_camera_handle
var frameBuffer []byte
var frameBufferMutex sync.Mutex
var stopLooping bool
var readyForFrames bool

func sleep(ms int) {
	time.Sleep(time.Millisecond * time.Duration(ms))
}

func InitCam() error {
	file, err := os.ReadFile("/lib/systemd/system/mm-anki-camera.service")
	if err != nil {
		panic("mm-anki-camera service doesn't exist?")
	}
	if !strings.Contains(string(file), "mm-anki-camera -r 1") {
		fmt.Println("Adding -r 1 to mm-anki-camera service file for faster framerate...")
		original := `ExecStart=/usr/bin/mm-anki-camera $MM_ANKI_CAMERA_OPTS`
		replacement := `ExecStart=/usr/bin/mm-anki-camera -r 1 $MM_ANKI_CAMERA_OPTS`
		err := exec.Command("sudo", "sed", "-i", fmt.Sprintf("s|%s|%s|g", original, replacement), "/lib/systemd/system/mm-anki-camera.service").Run()
		if err != nil {
			fmt.Println("failed to execute sed command:", err)
		}
		exec.Command("/bin/bash", "-c", "systemctl daemon-reload").Run()
		exec.Command("/bin/bash", "-c", "systemctl restart mm-anki-camera").Run()
		sleep(100)
		fmt.Println("success")
	}
	readyForFrames = false
	fmt.Println("Initializing Camera...")

	rc := C.camera_init(&camera)
	if rc != 0 {
		return fmt.Errorf("failed to initialize camera camera_init()")
	}

	sleep(1000)
	rc = C.camera_start(camera)
	if rc != 0 {
		return fmt.Errorf("failed to start camera camera_start()")
	}

	stopLooping = false
	go func() {
		// wait for camera to be ready
		for C.camera_status(camera) != C.ANKI_CAMERA_STATUS_RUNNING && !stopLooping {
			sleep(30)
		}
		var r C.int

		// frame-buffer-fill loop
		for C.camera_status(camera) == C.ANKI_CAMERA_STATUS_RUNNING && !stopLooping {
			sleep(30)
			var frame *C.anki_camera_frame_t

			r = C.camera_frame_acquire(camera, 0, &frame)
			if r != 0 {
				continue
			}

			frameSize := int(frame.height) * int(frame.bytes_per_row)
			frameData := C.GoBytes(unsafe.Pointer(&frame.data), C.int(frameSize))

			//fmt.Println(int(frame.width), int(frame.height), int(frame.bits_per_pixel), int(frame.bytes_per_row))

			frameBufferMutex.Lock()
			frameBuffer = make([]byte, frameSize)
			copy(frameBuffer, frameData)
			frameBufferMutex.Unlock()
			C.camera_frame_release(camera, frame.frame_id)
			if !readyForFrames && len(frameData) > 0 {
				readyForFrames = true
			}
		}
	}()

	// setup auto exposure
	go func() {
		for {
			if !readyForFrames {
				time.Sleep(time.Millisecond * 50)
			} else {
				break
			}
		}
		time.Sleep(time.Second * 1)
		for !stopLooping {
			sleep(200)
			AutoExposure(GetFrame())
		}
	}()

	// only return when frames are ready
	for {
		if !readyForFrames {
			time.Sleep(time.Millisecond * 50)
		} else {
			break
		}
	}
	return nil
}

func GetFrame() []byte {
	if !readyForFrames {
		fmt.Println("err: GetFrame() called before camera init")
		os.Exit(1)
	}
	frameBufferMutex.Lock()
	defer frameBufferMutex.Unlock()
	return frameBuffer
}

func StopCam() error {
	if !readyForFrames {
		return fmt.Errorf("camera already stopped")
	}
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

func TestDump() {
	err := InitCam()
	if err != nil {
		fmt.Println(err)
		return
	}
	time.Sleep(time.Second * 2)
	frame := GetFrame()

	fmt.Println("Received Frame, size:", len(frame))

	os.WriteFile("/tmp/test.raw", frame, 0777)

	err = StopCam()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Stopped camera. Exiting")
}

func SetExposure(ms uint16, gain float64) {
	if !readyForFrames {
		fmt.Println("must init camera before setting exposure")
		return
	}
	C.camera_set_exposure(camera, C.uint16_t(ms), C.float(gain))
}

func AutoExposure(rawData []byte) (uint16, float64) {
	width := 1280
	height := 720
	// debayering, go style
	rgbImage := make([][][]uint8, height/2)
	for i := range rgbImage {
		rgbImage[i] = make([][]uint8, width/2)
		for j := range rgbImage[i] {
			rgbImage[i][j] = make([]uint8, 3)
		}
	}

	for y := 0; y < height; y += 2 {
		for x := 0; x < width; x += 2 {
			idxRaw := (y*width + x) / 4 * 5
			idxY := y / 2
			idxX := x / 2

			r := (uint16(rawData[idxRaw+0]) << 2) | ((uint16(rawData[idxRaw+4]) >> 6) & 0x03)
			g1 := (uint16(rawData[idxRaw+1]) << 2) | ((uint16(rawData[idxRaw+4]) >> 4) & 0x03)
			g2 := (uint16(rawData[idxRaw+2]) << 2) | ((uint16(rawData[idxRaw+4]) >> 2) & 0x03)
			b := (uint16(rawData[idxRaw+3]) << 2) | ((uint16(rawData[idxRaw+4]) >> 0) & 0x03)

			g := (g1 + g2) >> 1

			rgbImage[idxY][idxX][0] = uint8(r >> 2)
			rgbImage[idxY][idxX][1] = uint8(g >> 2)
			rgbImage[idxY][idxX][2] = uint8(b >> 2)
		}
	}

	// brightest spot alg
	brightnessValues := make([]float64, 0, (height/2)*(width/2))
	sumBrightness := 0.0
	for _, row := range rgbImage {
		for _, pixel := range row {
			brightness := float64(pixel[0]+pixel[1]+pixel[2]) / 3.0
			brightnessValues = append(brightnessValues, brightness)
			sumBrightness += brightness
		}
	}

	sort.Float64s(brightnessValues)
	percentile95 := brightnessValues[int(0.95*float64(len(brightnessValues)))]

	meanBrightness := sumBrightness / float64(len(brightnessValues))

	targetBrightness := 130.0

	referenceBrightness := 0.7*percentile95 + 0.3*meanBrightness

	exposureMsFloat := 100 * (targetBrightness - referenceBrightness) / targetBrightness
	exposureMs := uint16(math.Max(1, math.Min(100, exposureMsFloat))) // 100 seems to be effective max

	gain := 4 * (targetBrightness - referenceBrightness) / targetBrightness
	gain = math.Max(0, math.Min(5, gain))
	SetExposure(exposureMs, gain)

	return exposureMs, gain
}

/*

a bilinear debayering algorithm which actually works!!!!

func unpackRaw10Bilinear(rawData []byte) []uint16 {
	unpackedData := make([]uint16, len(rawData)*8/10)
	for i := 0; i < len(rawData)/5*5; i += 5 {
		unpackedData[i/5*4+0] = uint16(rawData[i+0])<<2 | uint16(rawData[i+4]>>6)&0x03
		unpackedData[i/5*4+1] = uint16(rawData[i+1])<<2 | uint16(rawData[i+4]>>4)&0x03
		unpackedData[i/5*4+2] = uint16(rawData[i+2])<<2 | uint16(rawData[i+4]>>2)&0x03
		unpackedData[i/5*4+3] = uint16(rawData[i+3])<<2 | uint16(rawData[i+4]>>0)&0x03
	}
	return unpackedData
}

func debayerRGGBBilinear(imag []uint16, width, height int) *image.NRGBA {
	rgbImage := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			r, g, b := uint16(0), uint16(0), uint16(0)
			if y%2 == 0 && x%2 == 0 {
				r = imag[idx]
			} else if y%2 == 1 && x%2 == 1 {
				b = imag[idx]
			} else {
				g = imag[idx]
			}
			rgbImage.SetNRGBA(x, y, color.NRGBA{R: uint8(r >> 2), G: uint8(g >> 2), B: uint8(b >> 2), A: 255})
		}
	}

	// Interpolating the missing values for each color channel
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			c := rgbImage.NRGBAAt(x, y)
			if c.R == 0 {
				c.R = uint8((uint16(rgbImage.NRGBAAt(x-1, y).R) + uint16(rgbImage.NRGBAAt(x+1, y).R) + uint16(rgbImage.NRGBAAt(x, y-1).R) + uint16(rgbImage.NRGBAAt(x, y+1).R)) / 4)
			}
			if c.G == 0 {
				c.G = uint8((uint16(rgbImage.NRGBAAt(x-1, y).G) + uint16(rgbImage.NRGBAAt(x+1, y).G) + uint16(rgbImage.NRGBAAt(x, y-1).G) + uint16(rgbImage.NRGBAAt(x, y+1).G)) / 4)
			}
			if c.B == 0 {
				c.B = uint8((uint16(rgbImage.NRGBAAt(x-1, y).B) + uint16(rgbImage.NRGBAAt(x+1, y).B) + uint16(rgbImage.NRGBAAt(x, y-1).B) + uint16(rgbImage.NRGBAAt(x, y+1).B)) / 4)
			}
			rgbImage.SetNRGBA(x, y, c)
		}
	}

	return rgbImage
}

------------------ more basic algorithm VVV does 640x360. now implemented in JPEG library in c
func unpackRAW10(rawData []byte, width, height int) [][]uint16 {
	unpackedData := make([][]uint16, height)
	for i := range unpackedData {
		unpackedData[i] = make([]uint16, width)
	}
	for i := 0; i < len(rawData)/5*5; i += 5 {
		row := (i / 5 * 4) / width
		col := (i / 5 * 4) % width
		unpackedData[row][col+0] = uint16(rawData[i+0])<<2 | uint16(rawData[i+4]>>6)&0x03
		unpackedData[row][col+1] = uint16(rawData[i+1])<<2 | uint16(rawData[i+4]>>4)&0x03
		unpackedData[row][col+2] = uint16(rawData[i+2])<<2 | uint16(rawData[i+4]>>2)&0x03
		unpackedData[row][col+3] = uint16(rawData[i+3])<<2 | uint16(rawData[i+4]>>0)&0x03
	}
	return unpackedData
}

// debayers the unpacked data RGGB
func debayerRGGB(unpackedData [][]uint16, width, height int) *image.RGBA {
	rgbImage := image.NewRGBA(image.Rect(0, 0, width/2, height/2))
	for y := 0; y < height; y += 2 {
		for x := 0; x < width; x += 2 {
			r := unpackedData[y][x]
			g := (unpackedData[y][x+1] + unpackedData[y+1][x]) / 2
			b := unpackedData[y+1][x+1]
			rgbImage.Set(x/2, y/2, color.RGBA{uint8(r >> 2), uint8(g >> 2), uint8(b >> 2), 255})
		}
	}
	return rgbImage
}

*/
