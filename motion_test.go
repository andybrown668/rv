package van

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"testing"
)

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

//using two images with changes, verify the annotated resultant image
func TestImageAnnotation(t *testing.T) {
	Initialize()
	imgBefore, err := downsampleFile("test-data/motion/localized/nomouse.jpg")
	if err != nil {
		t.Fatal(err)
	}
	imgAfter, err := downsampleFile("test-data/motion/localized/mouse.jpg")
	if err != nil {
		t.Fatal(err)
	}

	imgChanges, _ := diff(imgBefore, imgAfter)

	annotate(imgChanges)

	if err := saveImage(imgChanges, "./result.jpg"); err != nil {
		t.Fatal(err)
	}
}
