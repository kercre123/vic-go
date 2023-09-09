package main

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kercre123/vic-go/vbody"
	"github.com/kercre123/vic-go/vscreen"
)

// program which will run in recovery partition

var CurrentList *List
var ScreenInited bool
var BodyInited bool
var MaxTM uint32
var MinTM uint32
var StopListening bool
var HangBody bool

type List struct {
	Info      string
	InfoColor color.Color
	Lines     []vscreen.Line
	// len and position start with 1
	Len       int
	Position  int
	ClickFunc []func()
	inited    bool
}

func (c *List) MoveDown() {
	if c.Len == c.Position {
		c.Position = 1
	} else {
		c.Position = c.Position + 1
	}
	c.UpdateScreen()
}

func (c *List) MoveUp() {
	// i'm not sure how to determine direction from the encoders, so i am doing always down
	fmt.Println("up")
}

func (c *List) UpdateScreen() {
	var linesShow []vscreen.Line
	// if info, have list go to bottom
	// 7 lines fit comfortably on screen
	if c.Info != "" {
		newLine := vscreen.Line{
			Text:  c.Info,
			Color: c.InfoColor,
		}
		linesShow = append(linesShow, newLine)
		numOfSpaces := 7 - c.Len
		if numOfSpaces < 0 {
			panic("too many items in list" + fmt.Sprint(numOfSpaces))
		}
		for i := 1; i < numOfSpaces; i++ {
			newLine = vscreen.Line{
				Text:  " ",
				Color: c.InfoColor,
			}
			linesShow = append(linesShow, newLine)
		}
	}
	for i, line := range c.Lines {
		var newLine vscreen.Line
		if i == c.Position-1 {
			newLine.Text = "> " + line.Text
			newLine.Color = line.Color
		} else {
			newLine.Text = "  " + line.Text
			newLine.Color = line.Color
		}
		linesShow = append(linesShow, newLine)
	}
	scrnData := vscreen.CreateTextImageFromLines(linesShow)
	vscreen.SetScreen(scrnData)
}

func (c *List) Init() {
	c.Position = 1
	c.Len = len(c.Lines)
	if !BodyInited {
		vbody.Init_Spine()
		BodyInited = true
	}
	if !ScreenInited {
		vscreen.InitLCD()
		vscreen.BlackOut()
		ScreenInited = true
	}
	c.UpdateScreen()
	c.inited = true
}

func ListenToBody() {
	if !CurrentList.inited {
		fmt.Println("error: init list before listening dummy")
		os.Exit(1)
	}
	for {
		if StopListening {
			fmt.Println("not listening anymore")
			StopListening = false
			return
		}
		if !CurrentList.inited || HangBody {
			for {
				time.Sleep(time.Second / 5)
				if CurrentList.inited {
					break
				}
			}
		}
		frame := vbody.GetFrame()
		if frame.ButtonState {
			CurrentList.ClickFunc[CurrentList.Position-1]()
			time.Sleep(time.Second / 3)
		}
		for i, enc := range frame.Encoders {
			if i > 1 {
				// only read wheels
				break
			}
			if enc.DLT < -1 {
				stopTimer := false
				stopWatch := false
				go func() {
					timer := 0
					for {
						if StopListening {
							fmt.Println("not listening anymore")
							StopListening = false
							return
						}
						if stopTimer {
							break
						}
						if timer == 30 {
							CurrentList.MoveDown()
							stopWatch = true
							break
						}
						timer = timer + 1
						time.Sleep(time.Millisecond * 10)
					}
				}()
				for {
					if StopListening {
						fmt.Println("not listening anymore")
						StopListening = false
						return
					}
					frame = vbody.GetFrame()
					if stopWatch {
						break
					}
					if frame.Encoders[i].DLT == 0 {
						stopTimer = true
						break
					}
				}
			}
		}
		time.Sleep(time.Millisecond * 10)
	}
}

func StartAnki_Confirm() {
	c := *CurrentList
	CurrentList = Confirm_Create(StartAnki, c)
	CurrentList.Init()
}

func StartAnki() {
	scrnData := vscreen.CreateTextImage("Stopping body...")
	vscreen.SetScreen(scrnData)
	StopListening = true
	time.Sleep(time.Second / 3)
	vbody.Close_Spine()
	scrnData = vscreen.CreateTextImage("Starting anki-robot.target...")
	vscreen.SetScreen(scrnData)
	vscreen.StopLCD()
	time.Sleep(time.Second / 2)
	exec.Command("/bin/bash", "-c", "systemctl start anki-robot.target").Run()
	fmt.Println("exiting")
	os.Exit(0)
}

func StartRescue_Confirm() {
	c := *CurrentList
	CurrentList = Confirm_Create(StartRescue, c)
	CurrentList.Init()
}

func StartRescue() {
	KillButtonDetect := false
	// rescue can crash, often
	HangBody = true
	scrnData := vscreen.CreateTextImage("vic-rescue will start in 3 seconds. Press the button anytime to return to the menu.")
	vscreen.SetScreen(scrnData)
	vscreen.StopLCD()
	ScreenInited = false
	time.Sleep(time.Second * 3)
	cmd := exec.Command("/bin/bash", "-c", "/anki/bin/vic-rescue")
	go func() {
		for {
			frame := vbody.GetFrame()
			if frame.ButtonState || KillButtonDetect {
				break
			}
			time.Sleep(time.Millisecond * 10)
		}
		fmt.Println("killing rescue")
		cmd.Process.Kill()
	}()
	cmd.Run()
	CurrentList = Recovery_Create()
	CurrentList.Init()
	time.Sleep(time.Second / 3)
	HangBody = false
}

func Reboot_Do() {
	exec.Command("/bin/bash", "-c", "bootctl f set_active a")
	scrnData := vscreen.CreateTextImage("Rebooting...")
	vscreen.SetScreen(scrnData)
	StopListening = true
	time.Sleep(time.Second / 2)
	vbody.Close_Spine()
	vscreen.StopLCD()
	exec.Command("/bin/bash", "-c", "reboot").Run()
}

func Reboot_Create() *List {
	// "ARE YOU SURE?"
	var Reboot List

	Reboot.Info = "Reboot?"
	Reboot.InfoColor = color.RGBA{0, 255, 0, 255}
	Reboot.ClickFunc = []func(){Reboot_Do, func() {
		CurrentList = Recovery_Create()
		CurrentList.Init()
	}}

	Reboot.Lines = []vscreen.Line{
		{
			Text:  "Yes",
			Color: color.RGBA{255, 255, 255, 255},
		},
		{
			Text:  "No",
			Color: color.RGBA{255, 255, 255, 255},
		},
	}

	return &Reboot
}

func Restore16() {
	HangBody = true
	err := StreamOTA("http://192.168.1.2:81/3.0.1.9.ota")
	if err != nil {
		if strings.Contains(err.Error(), "button") {
			CurrentList = Recovery_Create()
			CurrentList.Init()
			time.Sleep(time.Second / 3)
			HangBody = false
		} else {
			scrnData := vscreen.CreateTextImage("error downloading OTA: " + err.Error())
			vscreen.SetScreen(scrnData)
			time.Sleep(time.Second * 3)
			CurrentList = Recovery_Create()
			CurrentList.Init()
			HangBody = false
		}
	} else {
		HangBody = false
		CurrentList = Reboot_Create()
		CurrentList.Init()
		time.Sleep(time.Second / 3)
	}
}

func getNet() (ssid string, ip string) {
	out, _ := exec.Command("/bin/bash", "-c", "iwgetid").Output()
	iwcmd := strings.TrimSpace(string(out))
	if !strings.Contains(iwcmd, "ESSID") {
		ssid = "<not connected>"
		ip = "<not connected>"
	} else {
		ssid = strings.Replace(strings.TrimSpace(strings.Split(iwcmd, "ESSID:")[1]), `"`, "", -1)
		out, _ = exec.Command("/bin/bash", "-c", `/sbin/ifconfig wlan0 | grep 'inet addr' | cut -d: -f2 | awk '{print $1}'`).Output()
		ip = strings.TrimSpace(string(out))
	}
	return ssid, ip
}

func DetectButtonPress() {
	// for functions which show on screen, but aren't lists. hangs ListenToBody, returns when button is presed
	for {
		frame := vbody.GetFrame()
		if frame.ButtonState {
			return
		}
		time.Sleep(time.Millisecond * 10)
	}

}

func PrintNetworkInfo() {
	c := *CurrentList
	ssid, ip := getNet()
	lines := []string{"SSID: " + ssid, "IP: " + ip, " ", " ", " ", " ", "> Back"}
	scrnData := vscreen.CreateTextImageFromSlice(lines)
	vscreen.SetScreen(scrnData)
	HangBody = true
	time.Sleep(time.Second / 3)
	DetectButtonPress()
	CurrentList = &c
	CurrentList.Init()
	time.Sleep(time.Second / 3)
	HangBody = false
}

func GetHeadTemp() string {
	sysTemp, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(sysTemp))
}

func PrintBodyInfo() {
	c := *CurrentList
	HangBody = true
	time.Sleep(time.Second / 3)
	for {
		time.Sleep(time.Millisecond * 50)
		frame := vbody.GetFrame()
		if frame.ButtonState {
			break
		}
		lines := []string{
			"CHGR: " + fmt.Sprint(frame.ChargerVoltage) + "mV, BATT: " + fmt.Sprint(frame.BattVoltage) + "mV",
			"TEMP: H-" + GetHeadTemp() + "C, B-" + fmt.Sprint(frame.BodyTemp) + "C",
			"TOUCH: " + fmt.Sprint(frame.Touch),
			"CLIFF: " + fmt.Sprint(frame.Cliffs[0]) + " " + fmt.Sprint(frame.Cliffs[1]) + " " + fmt.Sprint(frame.Cliffs[2]) + " " + fmt.Sprint(frame.Cliffs[3]),
			" ",
			" ",
			"> Back",
		}
		scrnData := vscreen.CreateTextImageFromSlice(lines)
		vscreen.SetScreen(scrnData)
	}
	CurrentList = &c
	CurrentList.Init()
	time.Sleep(time.Second / 3)
	HangBody = false
}

func Rebooter() {
	CurrentList = Reboot_Create()
	CurrentList.Init()
}

func Recovery_Create() *List {
	var Test List

	Test.Info = "Recovery Menu"
	Test.InfoColor = color.RGBA{0, 255, 0, 255}

	Test.ClickFunc = []func(){StartAnki_Confirm, StartRescue_Confirm, PrintNetworkInfo, Rebooter, Restore16}

	Test.Lines = []vscreen.Line{
		{
			Text:  "Start Vector processes",
			Color: color.RGBA{255, 255, 255, 255},
		},
		{
			Text:  "Start vic-rescue",
			Color: color.RGBA{255, 255, 255, 255},
		},
		{
			Text:  "List network info",
			Color: color.RGBA{255, 255, 255, 255},
		},
		{
			Text:  "Reboot to system",
			Color: color.RGBA{255, 255, 255, 255},
		},
		{
			Text:  "Restore to 1.6",
			Color: color.RGBA{255, 255, 255, 255},
		},
	}

	return &Test
}

func Confirm_Create(do func(), origList List) *List {
	// "ARE YOU SURE?"
	var Test List

	Test.Info = "Are you sure?"
	Test.InfoColor = color.RGBA{0, 255, 0, 255}
	Test.ClickFunc = []func(){do, func() {
		CurrentList = &origList
		CurrentList.Init()
	}}

	Test.Lines = []vscreen.Line{
		{
			Text:  "Yes",
			Color: color.RGBA{255, 255, 255, 255},
		},
		{
			Text:  "Go back",
			Color: color.RGBA{255, 255, 255, 255},
		},
	}

	return &Test
}

func main() {
	CurrentList = Recovery_Create()
	CurrentList.Init()
	fmt.Println("started")
	ListenToBody()
}
