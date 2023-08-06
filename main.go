package main

import (
	"fmt"
	"time"

	"github.com/inancgumus/screen"
	"github.com/kercre123/vic-go/vbody"
)

func main() {
	TestBodyRXTXInteractive()
}

func TestBodyRXTXInteractive() {
	fmt.Println("Initiating spine")
	vbody.Init_Spine()
	vbody.Set_Motors(0, 0, -500, 0)
	time.Sleep(time.Second * 1)
	vbody.Set_Motors(0, 0, 0, 0)
	fmt.Println("\nTouch the touch sensor! (CTRL+C to exit)")
	frame := vbody.GetFrame()
	origTouch := frame.Touch
	for {
		frame = vbody.GetFrame()
		if frame.Touch > origTouch+30 {
			vbody.Set_Motors(0, 0, -100, 0)
		} else {
			vbody.Set_Motors(0, 0, 100, 0)
		}
	}
}

func TestBodyRXTX() {
	fmt.Println("\nThis is a demo of the spine. This shows how the wrapper is able to read from the spine while being able to write to it at the same time.")
	fmt.Println("This program will print the touch sensor readout. After a couple seconds, it will tell the body to move the motors and set the LEDs.")
	fmt.Println("Starting in 4 seconds...")
	time.Sleep(time.Second * 4)
	fmt.Println("Initiating spine")
	vbody.Init_Spine()
	screen.Clear()
	shouldStop := false
	go func() {
		time.Sleep(time.Second * 2)
		vbody.Set_LEDs(vbody.LED_BLUE, vbody.LED_BLUE, vbody.LED_BLUE)
		// right wheel, left wheel, lift, head. + = forward/up, - = backward/down
		vbody.Set_Motors(250, -250, 500, 500)
		time.Sleep(time.Second * 2)
		vbody.Set_Motors(-250, 250, -500, -500)
		time.Sleep(time.Second * 2)
		vbody.Set_Motors(0, 0, 0, 0)
		time.Sleep(time.Second * 2)
		shouldStop = true
	}()
	for {
		if shouldStop {
			break
		}
		screen.MoveTopLeft()
		fmt.Println("Touch sensor readout: " + fmt.Sprint(vbody.GetFrame().Touch))
		fmt.Println("Prox sensor readout: " + fmt.Sprint(vbody.GetFrame().ProxMM))
	}
	fmt.Println("Closing spine handle")
	vbody.Close_Spine()
}

// doesn't work
// func CamTest() {
// 	vcam.InitCam()
// 	fmt.Println("initted!")
// 	i := 0
// 	for i < 500 {
// 		i = i + 1
// 		vcam.GetFrame()
// 		time.Sleep(time.Millisecond * 1000 / 24)
// 		fmt.Println(i)
// 	}
// 	fmt.Println("jpged!")
// }

// func MotorTest() {
// 	fmt.Println("motorring")
// 	//rwheel, lwheel,
// 	vbody.Set_Motors(500, 500, 500, 500)
// 	go func() {
// 		time.Sleep(time.Second)
// 		vbody.Set_LEDs(vbody.LED_RED, vbody.LED_GREEN, vbody.LED_BLUE)
// 	}()
// 	time.Sleep(time.Millisecond * 3000)
// 	vbody.Set_Motors(-500, -500, -500, -500)
// 	time.Sleep(time.Millisecond * 1000)
// 	vbody.Set_Motors(0, 0, 0, 0)
// }
