package main

import (
	"fmt"
	"time"

	"github.com/inancgumus/screen"
	"github.com/kercre123/vic-go/vbody"
)

// 0 = right wheel, 1 = left wheel, 2 = lift, 3 = head

func main() {
	vbody.Init_Spine()
	screen.Clear()
	for {
		screen.MoveTopLeft()
		frame := vbody.GetFrame()
		for i, enc := range frame.Encoders {
			fmt.Println(i)
			fmt.Println("DLT: " + fmt.Sprint(enc.DLT) + "                 ")
			fmt.Println("Pos: " + fmt.Sprint(enc.Pos) + "                 ")
			fmt.Println("TM: " + fmt.Sprint(enc.TM) + "                 ")
		}
		time.Sleep(time.Millisecond * 10)

	}
}
