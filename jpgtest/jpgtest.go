package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kercre123/vic-go/vcam"
	"github.com/kercre123/vic-go/vjpeg"
)

var (
	width   = 640
	height  = 480
	quality = 100
)

func mjpegStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")

	for {
		jpegData := vjpeg.EncodeToJPEG(vcam.GetFrame(), quality, width, height)
		fmt.Fprintf(w, "--frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(jpegData))
		w.Write(jpegData)
		w.Write([]byte("\r\n"))
		time.Sleep(time.Second / 30)
	}
}

func BeginServer() {
	vjpeg.Init()
	vcam.InitCam(width, height)
	fmt.Println(len(vcam.GetFrame()))
	http.HandleFunc("/stream", mjpegStream)
	fmt.Println("listening at port 7070. go to /stream for mjpeg")
	http.ListenAndServe(":7070", nil)
}

// for debugging
func ToFile() {
	vjpeg.Init()
	vcam.InitCam(width, height)
	frame := vcam.GetFrame()
	fmt.Println(len(frame))
	jpegData := vjpeg.EncodeToJPEG(vcam.GetFrame(), quality, width, height)
	fmt.Println(len(jpegData))
	os.WriteFile("/tmp/test.jpg", jpegData, 0777)
}

func JustDump() {
	vcam.InitCam(1280, 720)
	frame := vcam.GetFrame()
	os.WriteFile("/tmp/test.raw", frame, 0777)
}

func main() {
	JustDump()
}
