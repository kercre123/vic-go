package vscreen

// #cgo LDFLAGS: -L${SRCDIR}/.. -lrobot
// #cgo CFLAGS: -I${SRCDIR}/../include
// #include "librobot.h"
// #include "lcd.h"
import "C"
import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os/exec"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var ScreenInitted bool

func InitLCD() {
	exec.Command("/bin/bash", "-c", "chmod 666 /sys/module/spidev/parameters/bufsiz").Run()
	exec.Command("/bin/bash", "-c", "echo 35328 > /sys/module/spidev/parameters/bufsiz").Run()
	exec.Command("/bin/bash", "-c", "chmod 444 /sys/module/spidev/parameters/bufsiz").Run()
	C.init_lcd()
	ScreenInitted = true
	BlackOut()
}

func wrapText(text string, lineWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	lines := words[:1]
	j := 0
	for _, word := range words[1:] {
		if len(lines[j]+" "+word) <= lineWidth {
			lines[j] += " " + word
		} else {
			lines = append(lines, word)
			j++
		}
	}
	return lines
}

func BlackOut() {
	if !ScreenInitted {
		fmt.Println("BlackOut(): init screen first")
	}
	pixels := make([]uint16, 184*96)
	for i := range pixels {
		pixels[i] = 0x000000
	}
	SetScreen(pixels)
}

func CreateTextImage(text string) []uint16 {
	const W, H = 184, 96
	img := image.NewRGBA(image.Rect(0, 0, W, H))
	black := color.RGBA{0, 0, 0, 255}
	white := color.RGBA{255, 255, 255, 255}

	draw.Draw(img, img.Bounds(), &image.Uniform{black}, image.Point{}, draw.Src)

	d := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{white},
		Face: basicfont.Face7x13,
		Dot:  fixed.P(0, 13),
	}

	//fmt.Println(13 * fixed.I(13))

	// Wrap text
	lines := wrapText(text, W/7) // assume each character is ~7px wide
	for _, line := range lines {
		d.Dot.X = 0
		d.DrawString(line)
		d.Dot.Y += fixed.I(13) // move down for the next line
	}

	pixels := make([]uint16, W*H)
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert the color format from RGBA to RGB565
			pixel := (r>>8&0xF8)<<8 | (g>>8&0xFC)<<3 | b>>8>>3
			pixels[y*W+x] = uint16(pixel)
		}
	}

	return pixels
}

func SetScreen(pixels []uint16) {
	if !ScreenInitted {
		fmt.Println("SetScreen(): init screen first")
	}
	C.set_pixels((*C.uint16_t)(&pixels[0]))
}

func StopLCD() {
	ScreenInitted = false
}
