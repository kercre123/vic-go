package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kercre123/vic-go/vbody"
	"github.com/kercre123/vic-go/vcam"
	"github.com/kercre123/vic-go/vjpeg"
)

type MotorRequest struct {
	Motors struct {
		Head   int `json:"head"`
		Lift   int `json:"lift"`
		Rwheel int `json:"rwheel"`
		Lwheel int `json:"lwheel"`
	} `json:"motors"`
}

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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func GetMotorValue(in int) int16 {
	var out int
	switch {
	case in == 0:
		out = 0
	case in == 1:
		out = 150
	case in == 2:
		out = 230
	case in == -1:
		out = -150
	case in == -2:
		out = -230
	}
	return int16(out)
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := conn.WriteMessage(messageType, p); err != nil {
			fmt.Println(err)
			return
		}
		var req MotorRequest
		json.Unmarshal([]byte(p), &req)
		vbody.Set_Motors(GetMotorValue(req.Motors.Rwheel), GetMotorValue(req.Motors.Lwheel), GetMotorValue(req.Motors.Lift), GetMotorValue(req.Motors.Head))
	}
}

func BeginServer() {
	vbody.Init_Spine()
	fmt.Println("Lowering head and lift")
	vbody.Set_Motors(0, 0, -100, -100)
	time.Sleep(time.Second * 2)
	vbody.Set_Motors(0, 0, 0, 0)
	vjpeg.Init()
	vcam.InitCam(width, height)
	fmt.Println(len(vcam.GetFrame()))
	http.HandleFunc("/ws", handler)
	http.Handle("/", http.FileServer(http.Dir("./webroot")))
	http.HandleFunc("/stream", mjpegStream)
	fmt.Println("listening at port 8888")
	http.ListenAndServe(":8888", nil)
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

func main() {
	BeginServer()
}
