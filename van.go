package main

import (
	"fmt"
	"image"
	"image/draw"
	"io/ioutil"
	"os/exec"
	"time"

	"github.com/goiot/devices/monochromeoled"
	"golang.org/x/exp/io/i2c"

	"github.com/d2r2/go-dht"
	"github.com/golang/freetype"
)

var (
	oled        *monochromeoled.OLED
	dst         = image.NewRGBA(image.Rect(0, 0, 128, 64))
	c           = freetype.NewContext()
	refresh     = time.Second * 5
	isOnline    = "Offline"
	temperature int
	humidity    int
)

func main() {
	monitorWifi()
	monitorDHT()
	if err := initialize(); err != nil {
		panic(err)
	}

	// wait for first reading...
	for temperature == 0 && humidity == 0 {
		time.Sleep(250 * time.Millisecond)
	}

	//monitor
	for {
		lines := []string{
			fmt.Sprintf("Cut and Run"),
			fmt.Sprintf("Humidity %d%%", humidity),
			fmt.Sprintf("Temp    %dc", temperature),
			fmt.Sprintf("%s  %s", isOnline, time.Now().Format("15:04")),
		}
		for _, line := range lines {
			fmt.Println(line)
		}
		display(lines)
		time.Sleep(refresh)
	}
}

//monitor connectivity state regularly
func monitorWifi() {
	go func() {
		for {
			_, err := exec.Command(`iwgetid`).Output()
			if err == nil {
				isOnline = "net:on"
			} else {
				isOnline = "net:off"
			}
			time.Sleep(refresh)
		}
	}()
	return
}

//monitor temp and humdidty regularly
func monitorDHT() {
	go func() {
		for {
			t, h, err := dht.ReadDHTxx(dht.DHT22, 10, false)
			if err == nil {
				temperature = int(t)
				humidity = int(h)
				time.Sleep(refresh)
			} else {
				time.Sleep(1500 * time.Millisecond)
			}
		}
	}()
}

func initialize() (err error) {
	oled, err = monochromeoled.Open(&i2c.Devfs{Dev: "/dev/i2c-1"})
	if err == nil {
		oled.On()
		oled.Clear()
	}

	data, err := ioutil.ReadFile("./enhanced_dot_digital-7.ttf")
	if err != nil {
		return err
	}
	f, err := freetype.ParseFont(data)
	if err != nil {
		return err
	}

	c.SetDst(dst)
	c.SetClip(dst.Bounds())
	c.SetSrc(image.Black)
	c.SetFont(f)
	c.SetFontSize(15)

	return nil
}

func display(lines []string) (err error) {
	draw.Draw(dst, dst.Bounds(), image.White, image.ZP, draw.Src)

	for y, line := range lines {
		if _, err := c.DrawString(line, freetype.Pt(0, 15+y*15)); err != nil {
			return err
		}
	}

	for y := 0; y < 64; y++ {
		for x := 0; x < 128; x++ {
			r, g, b, _ := dst.At(x, y).RGBA()
			if r == 0xffff && g == r && b == r {
				oled.SetPixel(x, y, 0)
			} else {
				oled.SetPixel(x, y, 1)
			}
		}
	}

	return oled.Draw()
}
