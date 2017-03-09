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
	"math/rand"
	"image/color"
)

type Stats struct {
	Temperature int
	Humidity    int
	Batteries   []*Battery
}

type Battery struct {
	Voltage float32 //draining when < FullChargeVoltage
	Load    float32 //draining when positive, charging when negative
}

const EmptyVoltage = 10.5
const FullChargeVoltage = 13.7
const MaxLoad = 160 //Amps

//charge is the ratio from 0.0 to 1.0
//when <= 1.0 the battery is discharging
//when greater than 1.0 it's charging
func (this *Battery) Charge() float32 {
	//use empty voltage as a base line

	return (this.Voltage - EmptyVoltage) / (FullChargeVoltage - EmptyVoltage)
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
	monitorBatteries()
	monitorAdc()
	startHttpApi()

	if err := initialize(); err != nil {
		panic(err)
	}

	//monitor
	blink := false
	for {
		blink = !blink
		title := fmt.Sprintf("C&R %2dc %2d%% ", stats.Temperature, stats.Humidity)
		//add a blinking * if online, - if not
		if blink {
			if isOnline {
				title += "*"
			} else {
				title += "-"
			}

		}

		lines := []string{
			title,
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

func monitorAdc() {
	go func() {
		//channels
		const LightChannel = 0x00
		const ThermistorChannel = 0x01
		const AIN0 = 0x00
		const AIN1 = 0x01
		const AIN2 = 0x02
		const AIN3 = 0x03
		const PotChannel = 0x03

		//configuration

		//all single input: An->outAn where n = 0..3
		const FOUR_SINGLE = 0x0 << 4

		// three differential. (+A0,-A3)->outA0, (+A1,-A3)->outA1, (+A2,-A3)->outA2
		const THREE_DIFF = 0x1 << 4

		// single: An->outAn where n = 0..1
		// differential (+A2,-A3) -> outA2
		const TWO_SINGLE_ONE_DIFF = 0x2 << 4

		// differential: (+A0,-A1)->outA0 (+A2,-A3)->outA1
		const TWO_DIFF = 0x3 << 4

		//open the adc
		d, err := i2c.Open(&i2c.Devfs{Dev: "/dev/i2c-1"}, 0x48)
		if err != nil {
			fmt.Println("open error", err)
		}
		defer d.Close()

		// write to control register
		if err = d.Write([]byte{TWO_DIFF | AIN0}); err != nil {
			fmt.Println("write error", err)
		}
		twos := func(b byte) int8 {
			if b>>7 == 0 {
				return int8(b)
			} else {
				b = b - (1 << 7)
				return int8(b) - 127 - 1
			}
		}
		read := make([]byte, 0x1)
		for {
			//request to read
			if err = d.Read(read); err != nil {
				fmt.Println("read error", err)
			} else {
				//convert two's complement
				n := twos(read[0])
				fmt.Print(n, " ")
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func monitorBatteries() {
	go func() {
		//fake three batteries
		for n := 0; n < 4; n++ {
			stats.Batteries = append(stats.Batteries, &Battery{})
		}

		for {
			//fake
			for _, battery := range stats.Batteries {
				//voltage in range from dead to full
				battery.Voltage = EmptyVoltage + (FullChargeVoltage-EmptyVoltage)*rand.Float32()
				battery.Load = MaxLoad * rand.Float32()
			}
			time.Sleep(1500 * time.Millisecond)
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

	return nil
}

func display(lines []string) (err error) {
	draw.Draw(dst, dst.Bounds(), image.White, image.ZP, draw.Src)

	//draw text
	c.SetFontSize(15)
	for y, line := range lines {
		if _, err := c.DrawString(line, freetype.Pt(0, 11+y*15)); err != nil {
			return err
		}
	}

	//draw battery meters - each meter indicates discharge with time till dead, and charge with time till full
	c.SetFontSize(12)
	for n, b := range stats.Batteries {
		y := 16 + n*12
		charge := b.Charge()

		//battery number and charging/discharging indicator
		indicator := fmt.Sprintf("%d", n)
		if b.Load < 0 {
			indicator += "+"
		} else {
			indicator += "-"
		}
		if _, err := c.DrawString(indicator, freetype.Pt(0, y+8)); err != nil {
			return err
		}

		//charge meter
		for x := 16; x <= int(80*charge); x += 4 {
			for dx := 0; dx < 3; dx ++ {
				for dy := 0; dy < 3; dy ++ {
					dst.Set(x+dx, y+dy, color.Black)
				}
			}
		}

		//load meter
		load := b.Load / MaxLoad
		for x := 16; x <= int(80*load); x += 2 {
			for dy := 4; dy < 7; dy++ {
				dst.Set(x, y+dy, color.Black)
			}
		}

		//time till charged/discharged
		if _, err := c.DrawString("3h45m", freetype.Pt(88, y+8)); err != nil {
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
