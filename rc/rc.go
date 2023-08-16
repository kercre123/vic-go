package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kercre123/vic-go/vbody"
	vcam "github.com/kercre123/vic-go/vcamnew"
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
	// resolution determined by which function you use
	// vjpeg.RGGB10ToJPEGDownSample does 640x360
	// (soon to be) vjpeg.RGGB10ToJPEGBilinear will do 1280x720
	quality = 50
)

func mjpegStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")

	for {
		jpegData := vjpeg.RGGB10ToJPEGDownSample(vcam.GetFrame(), quality)
		//jpegData := vjpeg.EncodeToJPEG(vcam.GetFrame(), quality, 1280, 720)
		_, err := fmt.Fprintf(w, "--frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(jpegData))
		if err != nil {
			fmt.Println("stopping mjpeg stream: " + err.Error())
			break
		}
		_, err = w.Write(jpegData)
		if err != nil {
			fmt.Println("stopping mjpeg stream: " + err.Error())
			break
		}
		_, err = w.Write([]byte("\r\n"))
		if err != nil {
			fmt.Println("stopping mjpeg stream: " + err.Error())
			break
		}
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

func exposuretest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/set_cam" {
		expF := r.FormValue("exp")
		gainF := r.FormValue("gain")
		exp, _ := strconv.Atoi(expF)
		gain, _ := strconv.ParseFloat(gainF, 64)
		vcam.SetExposure(uint16(exp), gain)
		fmt.Fprint(w, "success!!!!")
	} else if r.URL.Path == "/exptest" {
		val, val2 := vcam.AutoExposure(vcam.GetFrame())
		time.Sleep(time.Millisecond * 100)
		fmt.Println(val, val2)
		fmt.Fprint(w, "")
	}
}

func BeginServer() {
	vbody.Init_Spine()
	fmt.Println("Lowering head and lift")
	vbody.Set_Motors(0, 0, -100, -100)
	time.Sleep(time.Second * 2)
	vbody.Set_Motors(0, 0, 0, 0)
	vcam.InitCam()
	http.HandleFunc("/ws", handler)
	http.Handle("/", http.FileServer(http.Dir("./webroot")))
	http.HandleFunc("/set_cam", exposuretest)
	http.HandleFunc("/exptest", exposuretest)
	http.HandleFunc("/stream", mjpegStream)
	fmt.Println("listening at port 8888")
	http.ListenAndServe(":8888", nil)
}

func main() {
	BeginServer()
}
