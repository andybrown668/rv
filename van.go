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

	"encoding/json"
	"net/http"

	"github.com/d2r2/go-dht"
	"github.com/golang/freetype"
	"image/png"
)

type Stats struct {
	Temperature int
	Humidity    int
}

func (this *Stats) initialized() bool {
	return this.Humidity != 0
}

var (
	oled     *monochromeoled.OLED
	dst      = image.NewRGBA(image.Rect(0, 0, 128, 64))
	c        = freetype.NewContext()
	refresh  = time.Second * 5
	isOnline = false
	stats    = Stats{}
)

func main() {
	monitorWifi()
	monitorDHT()
	startHttpApi()

	if err := initialize(); err != nil {
		panic(err)
	}

	//monitor
	blink := false
	for {
		blink = !blink
		title := "Cut and Run   "
		//add a blinking * if online, - if not
		if blink {
			if isOnline{
			title += "*"
			} else{
				title += "-"
			}

		}

		lines := []string{
			title,
			fmt.Sprintf("Humidity    %2d%%", stats.Humidity),
			fmt.Sprintf("Temp       %2dc", stats.Temperature),
			time.Now().Format("15:04"),
		}
		for _, line := range lines {
			fmt.Println(line)
		}
		display(lines)
		time.Sleep(500 * time.Millisecond)
	}
}

func startHttpApi() {
	//json stats
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if jsonStats, err := json.MarshalIndent(stats, "", "    "); err == nil {
			fmt.Fprint(w, string(jsonStats))
		} else {
			fmt.Fprint(w, err.Error())
		}
	})

	//display image
	http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, dst)
	})
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println(err)
		}
	}()
}

//monitor connectivity state regularly
func monitorWifi() {
	go func() {
		for {
			_, err := exec.Command(`iwgetid`).Output()
			if err == nil {
				isOnline = true
			} else {
				isOnline = false
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
				stats.Temperature = int(t)
				stats.Humidity = int(h)
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
		if _, err := c.DrawString(line, freetype.Pt(0, 11+y*15)); err != nil {
			return err
		}
	}

	//only output if the oled device is available
	if oled == nil {
		return
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
