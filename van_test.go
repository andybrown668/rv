package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"testing"

	"github.com/d2r2/go-dht"
	"os"
)

func TestDHT(t *testing.T) {
	tmp, h, err := dht.ReadDHTxx(dht.DHT22, 10, true)
	fmt.Println(tmp, h, err)
}

func BenchDHT(b *testing.B) {
	for n := 0; n < b.N; n++ {
		tmp, h, err := dht.ReadDHTxx(dht.DHT22, 10, true)
		fmt.Println(tmp, h, err)
	}
}

var (
	dhtMarker = []byte{255, 196}
	dhtable   = []byte{1, 162, 0, 0, 1, 5, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 1, 0, 3, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 16, 0, 2, 1, 3, 3, 2, 4, 3, 5, 5, 4, 4, 0, 0, 1, 125, 1, 2, 3, 0, 4, 17, 5, 18, 33, 49, 65, 6, 19, 81, 97, 7, 34, 113, 20, 50, 129, 145, 161, 8, 35, 66, 177, 193, 21, 82, 209, 240, 36, 51, 98, 114, 130, 9, 10, 22, 23, 24, 25, 26, 37, 38, 39, 40, 41, 42, 52, 53, 54, 55, 56, 57, 58, 67, 68, 69, 70, 71, 72, 73, 74, 83, 84, 85, 86, 87, 88, 89, 90, 99, 100, 101, 102, 103, 104, 105, 106, 115, 116, 117, 118, 119, 120, 121, 122, 131, 132, 133, 134, 135, 136, 137, 138, 146, 147, 148, 149, 150, 151, 152, 153, 154, 162, 163, 164, 165, 166, 167, 168, 169, 170, 178, 179, 180, 181, 182, 183, 184, 185, 186, 194, 195, 196, 197, 198, 199, 200, 201, 202, 210, 211, 212, 213, 214, 215, 216, 217, 218, 225, 226, 227, 228, 229, 230, 231, 232, 233, 234, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 17, 0, 2, 1, 2, 4, 4, 3, 4, 7, 5, 4, 4, 0, 1, 2, 119, 0, 1, 2, 3, 17, 4, 5, 33, 49, 6, 18, 65, 81, 7, 97, 113, 19, 34, 50, 129, 8, 20, 66, 145, 161, 177, 193, 9, 35, 51, 82, 240, 21, 98, 114, 209, 10, 22, 36, 52, 225, 37, 241, 23, 24, 25, 26, 38, 39, 40, 41, 42, 53, 54, 55, 56, 57, 58, 67, 68, 69, 70, 71, 72, 73, 74, 83, 84, 85, 86, 87, 88, 89, 90, 99, 100, 101, 102, 103, 104, 105, 106, 115, 116, 117, 118, 119, 120, 121, 122, 130, 131, 132, 133, 134, 135, 136, 137, 138, 146, 147, 148, 149, 150, 151, 152, 153, 154, 162, 163, 164, 165, 166, 167, 168, 169, 170, 178, 179, 180, 181, 182, 183, 184, 185, 186, 194, 195, 196, 197, 198, 199, 200, 201, 202, 210, 211, 212, 213, 214, 215, 216, 217, 218, 226, 227, 228, 229, 230, 231, 232, 233, 234, 242, 243, 244, 245, 246, 247, 248, 249, 250}
	sosMarker = []byte{255, 218}
)

func addMotionDht(frame []byte) []byte {
	jpegParts := bytes.Split(frame, sosMarker)
	return append(jpegParts[0], append(dhtMarker, append(dhtable, append(sosMarker, jpegParts[1]...)...)...)...)
}

func downsample(filename string) (gray *image.Gray, err error) {
	if imgData, err := ioutil.ReadFile(filename); err != nil {
		return nil, err
	} else if img, err := jpeg.Decode(bytes.NewReader(addMotionDht(imgData))); err != nil {
		return nil, err
	} else {
		//downsample all pixels to grayscale
		bounds := img.Bounds()
		gray = image.NewGray(bounds)
		draw.Draw(gray, bounds, &image.Uniform{color.Black}, image.ZP, draw.Src)

		for y := 0; y < bounds.Dy(); y++ {
			for x := 0; x < bounds.Dx(); x++ {
				r, _, _, a := img.At(x, y).RGBA()
				//use red channel only
				r = r >> 8
				bw := color.RGBA{uint8(r), uint8(r), uint8(r), uint8(a)}
				gray.Set(x, y, bw)
			}
		}
		return gray, nil
	}

}

//subtract one image from another
func diff(image1, image2 *image.Gray) (gray *image.Gray) {
	//downsample all pixels to grayscale
	bounds := image1.Bounds()
	gray = image.NewGray(bounds)
	draw.Draw(gray, bounds, &image.Uniform{color.White}, image.ZP, draw.Src)

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			r1, _, _, a1 := image1.At(x, y).RGBA()
			r2, _, _, _ := image2.At(x, y).RGBA()
			//use red channel only
			r1 = r1 >> 8
			r2 = r2 >> 8
			if r2 >= r1 {
				r2 -= r1
			} else {
				r2 = r1 - r2
			}
			if r2 > 128 {
				r2 = 255
			} else {
				r2 = 0
			}
			bw := color.RGBA{uint8(r2), uint8(r2), uint8(r2), uint8(a1)}
			gray.Set(x, y, bw)
		}
	}
	return gray
}

func TestSceneCompaction(t *testing.T) {
	var gray, gray2 *image.Gray
	var err error

	if gray, err = downsample("frame-0.jpg"); err != nil {
		t.Fatal(err)
	} else if out, err := os.Create("./output.jpg"); err != nil {
		t.Fatal(err)
	} else {
		jpeg.Encode(out, gray, &jpeg.Options{Quality: 100})
	}

	if gray2, err = downsample("frame-1.jpg"); err != nil {
		t.Fatal(err)
	} else if out, err := os.Create("./output2.jpg"); err != nil {
		t.Fatal(err)
	} else {
		jpeg.Encode(out, gray2, &jpeg.Options{Quality: 100})
	}

	//subtract one image from another
	if out2, err := os.Create("./output3.jpg"); err != nil {
		t.Fatal(err)
	} else {
		gray3 := diff(gray, gray2)
		jpeg.Encode(out2, gray3, &jpeg.Options{Quality: 100})
	}
}
