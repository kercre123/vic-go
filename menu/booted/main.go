package main

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kercre123/vic-go/vbody"
	"github.com/kercre123/vic-go/vscreen"
)

var screenText []string = []string{
	"The OS is booted.",
}

func getNet() (ssid string, ip string) {
	out, _ := exec.Command("/bin/bash", "-c", "iwgetid").Output()
	iwcmd := strings.TrimSpace(string(out))
	ssid = strings.Replace(strings.TrimSpace(strings.Split(iwcmd, "ESSID:")[1]), `"`, "", -1)
	out, _ = exec.Command("/bin/bash", "-c", `/sbin/ifconfig wlan0 | grep 'inet addr' | cut -d: -f2 | awk '{print $1}'`).Output()
	ip = strings.TrimSpace(string(out))
	return ssid, ip
}

func main() {
	time.Sleep(time.Second * 4)
	exec.Command("/bin/bash", "-c", "systemctl stop vic-bootAnim").Run()
	time.Sleep(time.Second * 1)
	vscreen.InitLCD()
	vbody.Init_Spine()
	vscreen.BlackOut()
	ssid, ip := getNet()
	screenText = append(screenText, "SSID: "+ssid)
	screenText = append(screenText, "IP: "+ip)
	scrnBytes := vscreen.CreateTextImageFromSlice(screenText)
	vscreen.SetScreen(scrnBytes)
	vbody.Set_LEDs(vbody.LED_BLUE, vbody.LED_BLUE, vbody.LED_BLUE)
	vbody.Set_Motors(0, 0, 0, 100)
	time.Sleep(time.Second * 2)
	vbody.Set_Motors(0, 0, 0, 0)
	time.Sleep(time.Second / 3)
	vscreen.StopLCD()
	vbody.Close_Spine()
	os.Exit(0)
}
