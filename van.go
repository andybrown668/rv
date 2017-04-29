package main

// #include "go_sched.h"
import "C"

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

	"image/color"
	"image/png"
	"math/rand"

	"bytes"
	"image/jpeg"
	"os"

	"github.com/blackjack/webcam"
	"github.com/d2r2/go-dht"
	"github.com/golang/freetype"
	"sync"
	"github.com/golang/freetype/truetype"
	"io"
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
	go monitorWebcam()
	//monitorWifi()
	//monitorDHT()
	//monitorBatteries()
	//monitorAdc()
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
		if err := display(lines); err != nil {
			fmt.Println("failed to display:", err)
		}
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
	//webcam
	http.HandleFunc("/webcamraw", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Refresh", "15;url=webcamraw")
		mu.Lock()
		defer mu.Unlock()
		//if frame1 != nil {
		//	png.Encode(w, frame1)
		//}
		if file, err := os.Open("now.jpg"); err == nil {
			io.Copy(w, file)
			file.Close()
		} else {
			fmt.Println("failed to read now.jpg:", err)
		}
	})
	http.HandleFunc("/webcam", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Refresh", "15;url=webcam")
		mu.Lock()
		defer mu.Unlock()
		if gray3 != nil {
			png.Encode(w, gray3)
		}
	})
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println(err)
		}
	}()
}

var (
	dhtMarker = []byte{255, 196}
	dhtable   = []byte{1, 162, 0, 0, 1, 5, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 1, 0, 3, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 16, 0, 2, 1, 3, 3, 2, 4, 3, 5, 5, 4, 4, 0, 0, 1, 125, 1, 2, 3, 0, 4, 17, 5, 18, 33, 49, 65, 6, 19, 81, 97, 7, 34, 113, 20, 50, 129, 145, 161, 8, 35, 66, 177, 193, 21, 82, 209, 240, 36, 51, 98, 114, 130, 9, 10, 22, 23, 24, 25, 26, 37, 38, 39, 40, 41, 42, 52, 53, 54, 55, 56, 57, 58, 67, 68, 69, 70, 71, 72, 73, 74, 83, 84, 85, 86, 87, 88, 89, 90, 99, 100, 101, 102, 103, 104, 105, 106, 115, 116, 117, 118, 119, 120, 121, 122, 131, 132, 133, 134, 135, 136, 137, 138, 146, 147, 148, 149, 150, 151, 152, 153, 154, 162, 163, 164, 165, 166, 167, 168, 169, 170, 178, 179, 180, 181, 182, 183, 184, 185, 186, 194, 195, 196, 197, 198, 199, 200, 201, 202, 210, 211, 212, 213, 214, 215, 216, 217, 218, 225, 226, 227, 228, 229, 230, 231, 232, 233, 234, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 17, 0, 2, 1, 2, 4, 4, 3, 4, 7, 5, 4, 4, 0, 1, 2, 119, 0, 1, 2, 3, 17, 4, 5, 33, 49, 6, 18, 65, 81, 7, 97, 113, 19, 34, 50, 129, 8, 20, 66, 145, 161, 177, 193, 9, 35, 51, 82, 240, 21, 98, 114, 209, 10, 22, 36, 52, 225, 37, 241, 23, 24, 25, 26, 38, 39, 40, 41, 42, 53, 54, 55, 56, 57, 58, 67, 68, 69, 70, 71, 72, 73, 74, 83, 84, 85, 86, 87, 88, 89, 90, 99, 100, 101, 102, 103, 104, 105, 106, 115, 116, 117, 118, 119, 120, 121, 122, 130, 131, 132, 133, 134, 135, 136, 137, 138, 146, 147, 148, 149, 150, 151, 152, 153, 154, 162, 163, 164, 165, 166, 167, 168, 169, 170, 178, 179, 180, 181, 182, 183, 184, 185, 186, 194, 195, 196, 197, 198, 199, 200, 201, 202, 210, 211, 212, 213, 214, 215, 216, 217, 218, 226, 227, 228, 229, 230, 231, 232, 233, 234, 242, 243, 244, 245, 246, 247, 248, 249, 250}
	sosMarker = []byte{255, 218}
)

func addMotionDht(frame []byte) []byte {
	jpegParts := bytes.Split(frame, sosMarker)
	if len(jpegParts) < 2 {
		return frame
	}
	return append(jpegParts[0], append(dhtMarker, append(dhtable, append(sosMarker, jpegParts[1]...)...)...)...)
}

func downsampleFile(filename string) (gray *image.Gray, err error) {
	if imgData, err := ioutil.ReadFile(filename); err != nil {
		return nil, err
	} else if img, err := jpeg.Decode(bytes.NewReader(addMotionDht(imgData))); err != nil {
		return nil, err
	} else {
		return downsampleImage(img), nil
	}
}

func downsampleImage(img image.Image) (gray *image.Gray) {
	//downsampleFile all pixels to grayscale
	bounds := img.Bounds()
	gray = image.NewGray(bounds)
	draw.Draw(gray, bounds, &image.Uniform{color.Black}, image.ZP, draw.Src)

	for y := 0; y < bounds.Dy(); y += 1 {
		for x := 0; x < bounds.Dx(); x += 1 {
			r, _, _, _ := img.At(x, y).RGBA()
			//use red channel only
			r = r >> 8
			bw := color.Gray{uint8(r)}
			gray.SetGray(x, y, bw)
		}
	}
	return gray
}

//subtract one image from another
func diff(image1, image2 *image.Gray) (gray *image.Gray, changes int) {
	//downsampleFile all pixels to grayscale
	bounds := image1.Bounds()
	gray = image.NewGray(bounds)
	draw.Draw(gray, bounds, &image.Uniform{color.White}, image.ZP, draw.Src)

	total := bounds.Dx() * bounds.Dy()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			c1 := image1.GrayAt(x, y)
			c2 := image2.GrayAt(x, y)
			r1 := c1.Y
			r2 := c2.Y - r1
			if r2 < 0 {
				r2 = 0 - r2
			}

			if r2 > 60 && r2 < 200 {
				r2 = 255
				changes++
			} else {
				r2 = 0
			}
			gray.Set(x, y, color.Gray{r2})
		}
	}

	return gray, (10000 * changes) / total
}

func saveImage(img *image.Gray, filename string) error {
	// copy the image as rgb so we can annotate it and add a timestamp/info strip below
	bounds := img.Bounds()
	bounds.Max.Y += 20
	i := image.NewRGBA(bounds)
	draw.Draw(i, bounds, img, image.ZP, draw.Src)

	// make the info strip white
	strip := bounds
	strip.Min.Y = strip.Max.Y - 20
	draw.Draw(i, strip, &image.Uniform{color.White}, image.ZP, draw.Src)
	c := freetype.NewContext()
	c.SetDst(i)
	c.SetClip(i.Bounds())
	c.SetSrc(image.NewUniform(color.RGBA{R:0x00, B:0xff, A:0xff}))
	c.SetFont(fnt)
	c.SetFontSize(18)
	if _, err := c.DrawString(time.Now().Format("2006-01-02 15:04:05"), freetype.Pt(5, bounds.Max.Y-4)); err != nil {
		return err
	}
	if f, err := os.Create(filename); err != nil {
		return err
	} else {
		jpeg.Encode(f, i, &jpeg.Options{Quality: 100})
	}
	return nil
}

var mu sync.Mutex
var frame1, gray3 *image.Gray

//get a downsampled frame or nil on error
func getFrame(cam *webcam.Webcam) *image.Gray {

	if err := cam.WaitForFrame(5); err != nil {
		fmt.Println("Failed waiting for frame", err)
		return nil
	}

	var img image.Image

	if frame, err := cam.ReadFrame(); err != nil {
		fmt.Println("failed to get frame", err)
	} else if img, err = jpeg.Decode(bytes.NewReader(addMotionDht(frame))); err != nil {
		fmt.Println("failed to decode frame", err)
	}

	if img != nil {
		return downsampleImage(img)
	}

	return nil
}

func monitorWebcam() {
	cam, err := webcam.Open("/dev/video0")
	if err != nil {
		fmt.Println("failed to open webcam", err)
	}
	defer cam.Close()

	var format webcam.PixelFormat
	formats := cam.GetSupportedFormats()
	for k, v := range formats {
		fmt.Println(k, v)
		if v == "Motion-JPEG" {
			format = k
		}
	}

	sizes := cam.GetSupportedFrameSizes(format)
	fmt.Println(sizes)
	var size webcam.FrameSize
	size = sizes[1]
	size.MaxWidth = 320
	size.MaxHeight = 240
	if _, _, _, err := cam.SetImageFormat(format, size.MaxWidth, size.MaxHeight); err != nil {
		fmt.Println("failed to set format and size: ", err)
	}

	n := 0

	//get and compare consecutive frames
	var frame2 *image.Gray
	for {
		time.Sleep(250 * time.Millisecond)
		if n == 0 {
			if err := cam.StartStreaming(); err != nil {
				fmt.Println("error starting streaming:", err)
				break
			}

		}
		//read frame 1 or use prior frame2
		if frame2 == nil {
			if frame := getFrame(cam); frame == nil {
				continue
			} else {
				mu.Lock()
				frame1 = frame
				mu.Unlock()
			}
		} else {
			mu.Lock()
			frame1 = frame2
			mu.Unlock()
		}

		//always save current frame
		mu.Lock()
		if err := saveImage(frame1, "now.jpg"); err != nil {
			fmt.Println("error saving current:", err)
		}
		mu.Unlock()

		if frame2 = getFrame(cam); frame2 == nil {
			continue
		}

		//subtract one image from another
		mu.Lock()
		var changes int
		gray3, changes = diff(frame1, frame2)
		mu.Unlock()
		if changes < 50 {
			continue
		}

		//save diff frame and two input frames
		if err := saveImage(frame1, fmt.Sprintf("./%d-before-%d.jpg", n, changes)); err != nil {
			fmt.Println("failed to write frame 1")
		}
		if err := saveImage(frame2, fmt.Sprintf("./%d-after-%d.jpg", n, changes)); err != nil {
			fmt.Println("failed to write frame 2")
		}
		if err := saveImage(gray3, fmt.Sprintf("./%d-changes%d.jpg", n, changes)); err != nil {
			fmt.Println("failed to write frame diff")
		} else {
			fmt.Println("motion at", n)
			n++
		}

	}

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
type Reading struct {
	value    C.int
	duration time.Duration
	ticks    uint32
}

func (this Reading) String() string {
	//return fmt.Sprintf("%c %s,", this.value, this.duration)
	if this.value == 38 {
		return "."
	} else if this.value == 39 {
		if this.ticks == 3 {
			return "1"
		} else {
			return "0"
		}
	} else {
		return fmt.Sprintf("%c", this.value)
	}
}

func monitorDHT() {
	go func() {
		for {
			t, h, err := dht.ReadDHTxx(dht.DHT22, 10, true)
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
			fmt.Println("adc open error", err)
		}
		defer d.Close()

		// write to control register
		if err = d.Write([]byte{FOUR_SINGLE | AIN0}); err != nil {
			fmt.Println("adc write error", err)
		}
		//twos := func(b byte) int8 {
		//	if b>>7 == 0 {
		//		return int8(b)
		//	} else {
		//		b = b - (1 << 7)
		//		return int8(b) - 127 - 1
		//	}
		//}
		read := make([]byte, 0x1)
		for {
			//request to read
			if err = d.Read(read); err != nil {
				fmt.Println("adc read error", err)
			} else {
				//convert two's complement
				n := read[0] //twos(read[0])
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

var fnt *truetype.Font
func initialize() (err error) {
	oled, err = monochromeoled.Open(&i2c.Devfs{Dev: "/dev/i2c-1"})
	if err == nil {
		oled.On()
		oled.Clear()
	} else {
		fmt.Println("failed to open display")
	}

	data, err := ioutil.ReadFile("./enhanced_dot_digital-7.ttf")
	if err != nil {
		return err
	}
	fnt, err = freetype.ParseFont(data)
	if err != nil {
		return err
	}

	c.SetDst(dst)
	c.SetClip(dst.Bounds())
	c.SetSrc(image.Black)
	c.SetFont(fnt)

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
			for dx := 0; dx < 3; dx++ {
				for dy := 0; dy < 3; dy++ {
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
