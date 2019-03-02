package van

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/golang/freetype"
)

func Display(lines []string) (err error) {
	draw.Draw(dst, dst.Bounds(), image.White, image.ZP, draw.Src)

	//draw text
	c.SetFontSize(15)
	for y, line := range lines {
		if _, err := c.DrawString(line, freetype.Pt(0, 11+y*15)); err != nil {
			return err
		}
	}

	//draw battery meters - each meter indicates discharge with time till dead, and charge with time till full
	fontSize := 20
	switch len(CurrentStats.Batteries) {
	case 2:
		fontSize = 12
	case 3:
		fontSize = 12
	case 4:
		fontSize = 12
	}
	c.SetFontSize(float64(fontSize))
	for n, b := range CurrentStats.Batteries {
		y := 4 + fontSize + n*fontSize
		charge := b.ChargeRatio()

		//battery number and charging/discharging indicator
		indicator := fmt.Sprintf("%s", b.Name)
		ind := ""
		if b.Load < 0 {
			ind = "-"
		} else if b.Load > 0 {
			ind = "+"
		}

		if _, err := c.DrawString(indicator, freetype.Pt(0, y+8)); err != nil {
			return err
		}

		//voltage/s-o-c meter
		//fmt.Println("charge", charge, "from", b.Voltage)
		for x := 0; x <= int(70*charge); x += 4 {
			for dx := 0; dx < 3; dx++ {
				for dy := 0; dy < 3; dy++ {
					dst.Set(16+x+dx, y+dy, color.Black)
				}
			}
		}

		//current meter - range -127 to +128
		// sign indicates direction (discharge/charge)
		load := b.Load
		if load < 0 {
			load = -load
		}
		//base on full load == 127
		pctLoad := float32(load) / 127.0
		//pctLoad = 1
		//fmt.Println("load=", pctLoad, "from", b.Load, load)
		for x := 0; x <= int(70*pctLoad); x += 2 {
			for dy := 4; dy < 7; dy++ {
				dst.Set(16+x, y+dy, color.Black)
			}
		}

		//time when charged/discharged
		//if _, err := c.DrawString(time.Now().Format("15:04"), freetype.Pt(88, y+8)); err != nil {
		//	return err
		//}
		if _, err := c.DrawString(fmt.Sprintf("%2.0f%%%s", 100*charge, ind), freetype.Pt(96, y+8)); err != nil {
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
