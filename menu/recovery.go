package main

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"time"

	"github.com/kercre123/vic-go/vbody"
	"github.com/kercre123/vic-go/vscreen"
)

// program which will run in recovery partition

var CurrentList List
var MaxTM uint32
var MinTM uint32
var StopListening bool

type List struct {
	Info      string
	InfoColor color.Color
	Lines     []vscreen.Line
	// len and position start with 1
	Len       int
	Position  int
	ClickFunc []func()
	inited    bool
	Stop      bool
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
	vbody.Init_Spine()
	vscreen.InitLCD()
	vscreen.BlackOut()
	c.UpdateScreen()
	c.inited = true
}

func (c *List) ListenToBody() {
	if !c.inited {
		fmt.Println("error: init list before listening dummy")
		os.Exit(1)
	}
	for {
		if StopListening || c.Stop {
			fmt.Println("not listening anymore")
			return
		}
		frame := vbody.GetFrame()
		if frame.ButtonState {
			c.ClickFunc[c.Position-1]()
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
						if stopTimer {
							break
						}
						if timer == 30 {
							c.MoveDown()
							stopWatch = true
							break
						}
						timer = timer + 1
						time.Sleep(time.Millisecond * 10)
					}
				}()
				for {
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
	}
}

func StartAnki() {
	scrnData := vscreen.CreateTextImage("Stopping body...")
	vscreen.SetScreen(scrnData)
	StopListening = true
	time.Sleep(time.Second / 3)
	vbody.Close_Spine()
	scrnData = vscreen.CreateTextImage("Stopping LCD and starting Anki robot target...")
	vscreen.SetScreen(scrnData)
	vscreen.StopLCD()
	time.Sleep(time.Second / 2)
	exec.Command("/bin/bash", "-c", "systemctl start anki-robot.target").Run()
	fmt.Println("exiting")
	os.Exit(0)
}

func Restore16() {
	fmt.Println("no-op")
}

func Test_Create() List {
	var Test List

	Test.Info = "Unbrick Menu"
	Test.InfoColor = color.RGBA{0, 255, 0, 255}

	Test.ClickFunc = []func(){StartAnki, Restore16}

	Test.Lines = []vscreen.Line{
		{
			Text:  "Start Vector processes",
			Color: color.RGBA{255, 255, 255, 255},
		},
		{
			Text:  "Restore to 1.6",
			Color: color.RGBA{255, 255, 255, 255},
		},
	}

	return Test
}

func main() {
	testList := Test_Create()
	testList.Init()
	fmt.Println("started")
	testList.ListenToBody()
}
