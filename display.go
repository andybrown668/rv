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
	c.SetFontSize(12)
	for n, b := range CurrentStats.Batteries {
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
