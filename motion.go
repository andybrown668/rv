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
	//draw.Draw(gray, bounds, &image.Uniform{color.White}, image.ZP, draw.Src)

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
