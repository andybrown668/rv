package van

// #include "go_sched.h"
import "C"

import (
	"fmt"
	"image"
	"io/ioutil"
	"time"

	"github.com/goiot/devices/monochromeoled"
	"golang.org/x/exp/io/i2c"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

type Stats struct {
	Temperature int
	Humidity    int
	Batteries   []*Battery
	Adc0        int
	Adc2        int
}

func (this *Stats) initialized() bool {
	return this.Humidity != 0
}

var (
	oled         *monochromeoled.OLED
	dst          = image.NewRGBA(image.Rect(0, 0, 128, 64))
	c            = freetype.NewContext()
	Refresh      = time.Second * 1
	IsOnline     = false
	CurrentStats = Stats{}
)

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

var fnt *truetype.Font

func Initialize() (err error) {
	for {
		oled, err = monochromeoled.Open(&i2c.Devfs{Dev: "/dev/i2c-1"})
		if err == nil {
			oled.On()
			oled.Clear()
			break
		} else {
			fmt.Println("failed to open display:", err)
		}
		time.Sleep(1 * time.Second)
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
