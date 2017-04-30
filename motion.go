package van

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
)

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
	bounds := image1.Bounds()
	gray = image.NewGray(bounds)

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

func surrounded(img *image.Gray, x, y int) int {
	//bloom out from the point looking for all-white pixels, stop when black found
	//return resulting rectangle if area > min
	n := 1
	max := img.Rect.Max
	foundBlack := false
	for !foundBlack {
		for dx := x - n; !foundBlack && dx <= x+n; dx++ {
			//ignore dx out of image bounds
			if dx < 0 || dx >= max.X {
				continue
			}
			for dy := y - 1; !foundBlack && dy < y+n; dy++ {
				//ignore dy out of image bounds
				if dy < 0 || dy > max.Y {
					continue
				}

				if img.GrayAt(dx, dy).Y == 0 {
					foundBlack = true
					break
				}
			}
		}
		n++
	}

	return n
}

func largestChange(img *image.Gray) (at image.Point, size int) {
	max := img.Rect.Max

	// compute bounding rectangle(s) for the changes
	// algorithm: look for first pixel that is surrounded by 5 pixels
	// then bloom out from there to find the boundary
	best := image.Point{}
	bestN := 0
	for y := 0; y < max.Y; y++ {
		for x := 0; x < max.X; x++ {
			c := img.GrayAt(x, y)
			if c.Y == 255 {
				if bound := surrounded(img, x, y); bound > bestN {
					bestN = bound
					best.X = x
					best.Y = y
				}
			}
		}
	}
	return 	best, bestN
}

func annotate(img *image.Gray) {
	at, size := largestChange(img)
	//draw bounding rect on image
	DrawRect(img, color.White, at.X-size, at.Y-size, at.X+size, at.Y+size)
	DrawRect(img, color.White, at.X-size-2, at.Y-size-2, at.X+size+2, at.Y+size+2)

}

// HLine draws a horizontal line
func HLine(img *image.Gray, col color.Color, x1, y, x2 int) {
	for ; x1 <= x2; x1++ {
		img.Set(x1, y, col)
	}
}

// VLine draws a veritcal line
func VLine(img *image.Gray, col color.Color, x, y1, y2 int) {
	for ; y1 <= y2; y1++ {
		img.Set(x, y1, col)
	}
}

// Rect draws a rectangle utilizing HLine() and VLine()
func DrawRect(img *image.Gray, col color.Color, x1, y1, x2, y2 int) {
	HLine(img, col, x1, y1, x2)
	HLine(img, col, x1, y2, x2)
	VLine(img, col, x1, y1, y2)
	VLine(img, col, x2, y1, y2)
}