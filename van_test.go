package van

import (
	"fmt"
	"image"
	"image/jpeg"
	"testing"

	"os"

	"github.com/d2r2/go-dht"
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

func TestSceneCompaction(t *testing.T) {
	var gray, gray2 *image.Gray
	var err error

	for frame := 1; frame < 67; frame++ {
		//load first frame or copy from prior second frame
		if gray2 == nil {
			if gray, err = downsampleFile(fmt.Sprintf("frame-%d.jpg", frame-1)); err != nil {
				t.Fatal(err)
			}
		} else {
			gray = gray2
		}
		if gray2, err = downsampleFile(fmt.Sprintf("frame-%d.jpg", frame)); err != nil {
			t.Fatal(err)
		}

		//subtract one image from another
		gray3, changes := diff(gray, gray2)
		if changes < 50 {
			continue
		}
		if out2, err := os.Create(fmt.Sprintf("./output-%d-%d.jpg", frame, changes)); err != nil {
			t.Fatal(err)
		} else {
			jpeg.Encode(out2, gray3, &jpeg.Options{Quality: 100})
		}
	}

}

func TestImageAnnotation(t *testing.T) {
	initialize()
	img, err := downsampleFile("./frame2-37-624.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if err := saveImage(img, "./result.jpg"); err != nil{
		t.Fatal(err)
	}
}
