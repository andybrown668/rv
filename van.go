package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"github.com/goiot/devices/monochromeoled"
	"golang.org/x/exp/io/i2c"
	"time"
	"image/draw"

	"github.com/golang/freetype"
	"github.com/d2r2/go-dht"
)

func main() {
	if err := initialize(); err != nil {
		panic(err)
	}

	for {
		temperature, humidity, retried, err := dht.ReadDHTxxWithRetry(dht.DHT22, 10, false, 10)
		if err != nil {
			time.Sleep(1*time.Millisecond)
			continue
		}
		fmt.Println(temperature, humidity, retried, err)
		if err == nil {
			lines := []string{fmt.Sprintf("Cut & Run %s", time.Now().Format("15:04")), fmt.Sprintf("Humidity %.0f%%", humidity), fmt.Sprintf("Temp %.0fc", temperature), ""}
			fmt.Println(lines)
			display(lines)
		}
		time.Sleep(1 * time.Second)
	}
}

var (
	oled *monochromeoled.OLED
	dst = image.NewRGBA(image.Rect(0, 0, 128, 64))
	c = freetype.NewContext()
)

func initialize() (err error){
	oled, err = monochromeoled.Open(&i2c.Devfs{Dev: "/dev/i2c-1"})
	if err == nil {
		oled.On()
		oled.Clear()
	}

	data, err := ioutil.ReadFile("./enhanced_dot_digital-7.ttf")
	if err != nil {
		panic(err)
	}
	f, err := freetype.ParseFont(data)
	if err != nil {
		panic(err)
	}

	c.SetDst(dst)
	c.SetClip(dst.Bounds())
	c.SetSrc(image.Black)
	c.SetFont(f)
	c.SetFontSize(15)

	return
}

func display(lines []string) {
	draw.Draw(dst, dst.Bounds(), image.White, image.ZP, draw.Src)

	for y, line := range lines {
		_, err := c.DrawString(line, freetype.Pt(0, 15+y*15))
		if err != nil {
			panic(err)
		}
	}

	for y := 0; y < 64; y++{
		for x := 0; x < 128; x++{
			r, g, b, _ := dst.At(x, y).RGBA()
			if r == 0xffff && g == r && b == r {
				oled.SetPixel(x, y, 0)
			} else {
				oled.SetPixel(x, y, 1)
			}
		}
	}

	oled.Draw()
}
