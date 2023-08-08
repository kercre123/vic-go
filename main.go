package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/inancgumus/screen"
	"github.com/kercre123/vic-go/vbody"
	"github.com/kercre123/vic-go/vcam"
	"github.com/kercre123/vic-go/vscreen"
)

func main() {
	ProgramDemo()
}

func ProgramDemo() {
	// should include all capabilities of application
	fmt.Println("First, showing camera on screen for 300 frames. Note, setting the CPU and RAM freqs higher will increase framerate shown on screen.")
	CamOnScreen()
	fmt.Println("Now, moving on to body rx/tx demo.")
	TestBodyRXTX()
	fmt.Println("End of the demo. Exiting")
}

func TestBodyRXTX() {
	fmt.Println("\nThis is a demo of the spine. This shows how the wrapper is able to read from the spine while being able to write to it at the same time.")
	fmt.Println("This program will print the touch sensor readout. After a couple seconds, it will tell the body to move the motors and set the LEDs.")
	fmt.Println("Starting in 3 seconds...")
	time.Sleep(time.Second * 3)
	fmt.Println("Initiating spine")
	vbody.Init_Spine()
	fmt.Println("")
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
		fmt.Printf("\rTouch sensor readout: " + fmt.Sprint(vbody.GetFrame().Touch))
	}
	fmt.Printf("\n\nClosing spine handle...\n")
	vbody.Close_Spine()
}

func CamOnScreen() {
	fmt.Println("Initiating LCD and cam")
	vscreen.InitLCD()
	vcam.InitCam(640, 480)
	fmt.Println("")
	i := 0
	for i < 300 {
		i = i + 1
		fmt.Printf("\r" + fmt.Sprint(i))
		vscreen.SetScreen(vcam.ConvertFrameToRGB565(vcam.GetFrame(), 640, 480, 184, 96))
		time.Sleep(time.Millisecond * (1000 / 30))
	}
	fmt.Printf("\n\nCleaning up, stopping camera stream...\n")
	vscreen.BlackOut()
	vscreen.StopLCD()
	vcam.StopCam()
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

// shows journalctl log on screen
func TailLogAndPrint() {
	vscreen.InitLCD()
	// Create the command
	cmd := exec.Command("journalctl", "-f")

	// Get a pipe to read from standard out
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	// Start the command
	if err = cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Create a new scanner
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		var displayLine string
		displayLineRay := strings.Split(scanner.Text(), " ")
		displayLine = displayLineRay[4]
		for i, line := range displayLineRay {
			if i < 5 {
				continue
			}
			displayLine = displayLine + " " + line
		}
		screenText := vscreen.CreateTextImage(displayLine)
		vscreen.SetScreen(screenText)
	}
	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}
	// Wait for the command to finish
	cmd.Wait()
}

func DebugShell() {
	//vbody.Init_Spine()
	for {
		var input string
		fmt.Printf("\n#~ ")
		fmt.Scanln(&input)
		if input == "proxstream" {
			screen.Clear()
			for {
				screen.MoveTopLeft()
				frame := vbody.GetFrame()
				fmt.Println("prox_sigma_mm: " + fmt.Sprint(frame.ProxSigmaMM) + "    ")
				fmt.Println("prox_raw_range_mm: " + fmt.Sprint(frame.ProxRawRangeMM) + "    ")
				fmt.Println("prox_signal_rate_mcps: " + fmt.Sprint(frame.ProxSignalRateMCPS) + "    ")
				fmt.Println("prox_ambient: " + fmt.Sprint(frame.ProxAmbient) + "    ")
				fmt.Println("prox_SPAD_count: " + fmt.Sprint(frame.ProxSPADCount) + "    ")
				fmt.Println("prox_sample_count: " + fmt.Sprint(frame.ProxSampleCount) + "    ")
				fmt.Println("prox_calibration_result: " + fmt.Sprint(frame.ProxCalibResult) + "    ")
				fmt.Println("calctest: " + fmt.Sprint(frame.Test) + "   ")
				time.Sleep(time.Millisecond * 50)
			}
		} else if input == "screen" {
			TailLogAndPrint()
			vscreen.StopLCD()
		} else if input == "exit" {
			vbody.Close_Spine()
			os.Exit(0)
		}

	}
}
