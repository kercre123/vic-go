package main

import (
	"archive/tar"
	"errors"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kercre123/vic-go/vbody"
	"github.com/kercre123/vic-go/vscreen"
	"gopkg.in/ini.v1"
)

var BootProgressKill bool
var SysProgressKill bool

type Progress struct {
	Size     int64
	Total    int64
	Progress int64
}

func ExecCmds(commands []string) {
	for _, cmd := range commands {
		exec.Command("/bin/bash", "-c", cmd).Run()
	}
}

func SetCPUToPerf() {
	cmds := []string{
		"echo 1267200 > /sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq",
		"echo performance > /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor",
		"echo disabled > /sys/kernel/debug/msm_otg/bus_voting",
		"echo 0 > /sys/kernel/debug/msm-bus-dbg/shell-client/update_request",
		"echo 1 > /sys/kernel/debug/msm-bus-dbg/shell-client/mas",
		"echo 512 > /sys/kernel/debug/msm-bus-dbg/shell-client/slv",
		"echo 0 > /sys/kernel/debug/msm-bus-dbg/shell-client/ab",
		"echo active clk2 0 1 max 800000 > /sys/kernel/debug/rpm_send_msg/message",
		"echo 1 > /sys/kernel/debug/msm-bus-dbg/shell-client/update_request",
	}
	ExecCmds(cmds)
}

func ProgressOnScreen(boot int, sysfs int) {
	linesShow := []vscreen.Line{
		{
			Text:  "Downloading 1.6...",
			Color: color.RGBA{0, 255, 0, 255},
		},
		{
			Text:  " ",
			Color: color.RGBA{0, 255, 0, 255},
		},
	}
	if sysfs != 0 {
		appendLine := vscreen.Line{
			Text:  "Boot partition: DONE",
			Color: color.RGBA{0, 255, 0, 255},
		}
		linesShow = append(linesShow, appendLine)
		if sysfs > 98 {
			appendLine = vscreen.Line{
				Text:  "System partition: DONE",
				Color: color.RGBA{0, 255, 0, 255},
			}
		} else {
			appendLine = vscreen.Line{
				Text:  "System partition: " + fmt.Sprint(sysfs) + "%",
				Color: color.RGBA{255, 255, 255, 255},
			}
		}
		linesShow = append(linesShow, appendLine)
	} else {
		appendLine := vscreen.Line{
			Text:  "Boot partition: " + fmt.Sprint(boot) + "%",
			Color: color.RGBA{255, 255, 255, 255},
		}
		linesShow = append(linesShow, appendLine)
		appendLine = vscreen.Line{
			Text:  "System partition: 0%",
			Color: color.RGBA{255, 255, 255, 255},
		}
		linesShow = append(linesShow, appendLine)
	}

	// include cancel
	linesShow = append(linesShow, []vscreen.Line{
		{
			Text:  " ",
			Color: color.RGBA{255, 255, 255, 255},
		},
		{
			Text:  " ",
			Color: color.RGBA{255, 255, 255, 255},
		},
		{
			Text:  "> Cancel",
			Color: color.RGBA{255, 255, 255, 255},
		},
	}...)
	scrnData := vscreen.CreateTextImageFromLines(linesShow)
	vscreen.SetScreen(scrnData)
}

func trackProgress(p *Progress, partition string) {
	for {
		if strings.Contains(partition, "boot_") && BootProgressKill {
			BootProgressKill = false
			return
		}
		if strings.Contains(partition, "system_") && SysProgressKill {
			SysProgressKill = false
			return
		}
		if p.Total > 0 {
			p.Progress = (p.Size * 100) / p.Total
			if strings.Contains(partition, "boot_") {
				ProgressOnScreen(int(p.Progress), 0)
			} else {
				ProgressOnScreen(100, int(p.Progress))
			}
		}
		time.Sleep(time.Second / 3)
	}
}

const BufferSize = 1024

func StreamOTA(url string) error {
	SetCPUToPerf()
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tr := tar.NewReader(resp.Body)
	var bootSize int64

	stopChan := make(chan bool, 1)
	go func() {
		ticker := time.NewTicker(time.Second / 3)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				frame := vbody.GetFrame()
				if frame.ButtonState {
					stopChan <- true
					SysProgressKill = true
					BootProgressKill = true
					time.Sleep(time.Second / 3)
					return
				}
			case <-stopChan:
				return
			}
		}
	}()

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch hdr.Name {
		case "manifest.ini":
			cfg, err := ini.Load(tr)
			if err != nil {
				return err
			}
			bootSize = cfg.Section("BOOT").Key("bytes").MustInt64(0)
			continue

		case "apq8009-robot-boot.img.gz":
			partition := "/dev/block/bootdevice/by-name/boot_a"
			totalSize := bootSize

			progress := &Progress{Total: totalSize}
			go trackProgress(progress, partition)

			cmdString := "openssl enc -d -aes-256-ctr -pass file:/anki/etc/ota.pas -md md5 | gunzip > /run/boot.img"
			cmd := exec.Command("/bin/bash", "-c", cmdString)

			cmdStdin, _ := cmd.StdinPipe()
			cmdErr := cmd.Start()
			if cmdErr != nil {
				return cmdErr
			}

			buf := make([]byte, BufferSize)
			for {
				n, err := tr.Read(buf)
				if err != nil && err != io.EOF {
					return err
				}
				if n == 0 || err == io.EOF {
					break
				}
				select {
				case <-stopChan:
					cmd.Process.Kill()
					return errors.New("operation stopped: button pressed")
				default:
					progress.Size += int64(n)
					cmdStdin.Write(buf[:n])
				}
			}

			cmdStdin.Close()
			cmd.Wait()
			exec.Command("/bin/bash", "-c", "dd if=/run/boot.img of="+partition).Run()
			os.Remove("/run/boot.img")
			BootProgressKill = true

		case "apq8009-robot-sysfs.img.gz":
			partition := "/dev/block/bootdevice/by-name/system_a"
			totalSize := hdr.Size

			progress := &Progress{Total: totalSize}
			go trackProgress(progress, partition)

			cmdString := fmt.Sprintf("openssl enc -d -aes-256-ctr -pass file:/anki/etc/ota.pas -md md5 | gunzip | dd of=%s", partition)
			cmd := exec.Command("/bin/bash", "-c", cmdString)

			cmdStdin, _ := cmd.StdinPipe()
			cmdErr := cmd.Start()
			if cmdErr != nil {
				return cmdErr
			}

			buf := make([]byte, BufferSize)
			for {
				n, err := tr.Read(buf)
				if err != nil && err != io.EOF {
					return err
				}
				if n == 0 || err == io.EOF {
					break
				}
				select {
				case <-stopChan:
					cmd.Process.Kill()
					return errors.New("operation stopped: button pressed")
				default:
					progress.Size += int64(n)
					cmdStdin.Write(buf[:n])
				}
			}

			cmdStdin.Close()
			cmd.Wait()
			SysProgressKill = true
			exec.Command("/bin/bash", "-c", "bootctl f set_active a")
			stopChan <- true
		}
	}
	return nil
}
