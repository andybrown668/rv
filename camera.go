package van

import (
	"fmt"
	"image"
	"path/filepath"
	"time"

	"bytes"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"sync"

	"github.com/blackjack/webcam"
	"github.com/golang/freetype"
)

//put all images into a subfolder named from first and last image dates
func archiveImages() {
	earliest := time.Now()
	var latest time.Time
	fileNames := []string{}
	filepath.Walk(ImagesFolder, func(path string, info os.FileInfo, err error) error {
		//ignore folders
		if info.IsDir() {
			if info.Name() == "images" {
				return nil
			}
			return filepath.SkipDir
		}
		fileNames = append(fileNames, info.Name())
		//retain earliest and latest creation dates
		modified := info.ModTime()
		if earliest.Sub(modified) > 0 {
			earliest = modified
		}
		if modified.Sub(latest) > 0 {
			latest = modified
		}
		return nil
	})

	//done if there's no files
	if len(fileNames) == 0 {
		fmt.Println("info: no images to archive")
		return
	}

	timeFormat := "1-2-1504"
	archiveFolderName := fmt.Sprintf("%simages-%s-%s-%d/", ImagesFolder, earliest.Format(timeFormat), latest.Format(timeFormat), len(fileNames))

	fmt.Printf("Archiving %d old images to %s\n", len(fileNames), archiveFolderName)
	if err := os.Mkdir(archiveFolderName, os.ModeDir+0744); err != nil {
		if !os.IsExist(err) {
			fmt.Printf("failed to create archive folder: %v\n", err)
			return
		}
	}

	//move all files to archive folder
	for _, fileName := range fileNames {
		if err := os.Rename(ImagesFolder+fileName, archiveFolderName+fileName); err != nil {
			fmt.Printf("failed to archive %s: %s\n", fileName, err)
		}
	}

}

func MonitorWebcam() {
	//archive prior run's images - based on the time range
	archiveImages()
	cam, err := webcam.Open("/dev/video0")
	if err != nil {
		fmt.Println("failed to open webcam", err)
	}
	defer cam.Close()

	var format webcam.PixelFormat
	formats := cam.GetSupportedFormats()
	for k, v := range formats {
		if v == "Motion-JPEG" {
			format = k
		}
	}

	sizes := cam.GetSupportedFrameSizes(format)
	var size webcam.FrameSize
	size = sizes[1]
	size.MaxWidth = 320
	size.MaxHeight = 240
	if _, _, _, err := cam.SetImageFormat(format, size.MaxWidth, size.MaxHeight); err != nil {
		fmt.Println("failed to set format and size: ", err)
	}

	//try to start streaming
	isStreaming := false
	for tries := 0; tries < 5; tries++ {
		if err := cam.StartStreaming(); err == nil {
			isStreaming = true
			break
		} else {
			fmt.Println("error starting streaming:", err)
		}
		fmt.Println("retry start streaming:", tries)
		time.Sleep(250 * time.Millisecond)
	}

	if !isStreaming {
		fmt.Println("Failed  to start streaming")
		return
	}

	notify("Monitoring camera feed")
	n := 0
	//get and compare consecutive frames
	var frame2 *image.Gray
	for {
		time.Sleep(500 * time.Millisecond)
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
		if changes > 50 {
			annotate(gray3)
		}
		mu.Unlock()
		if changes < 50 {
			continue
		}

		notify(fmt.Sprintf("Movement: http://192.168.1.51/%d-after.jpg", n))

		//save diff frame and two input frames
		if err := saveImage(frame1, fmt.Sprintf("./%d-before.jpg", n)); err != nil {
			fmt.Println("failed to write frame 1")
		}
		if err := saveImage(frame2, fmt.Sprintf("./%d-after.jpg", n)); err != nil {
			fmt.Println("failed to write frame 2")
		}

		if err := saveImage(gray3, fmt.Sprintf("./%d-changes.jpg", n)); err != nil {
			fmt.Println("failed to write frame diff")
		} else {
			fmt.Println("motion at", n)
			n++
		}

	}

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

var ImagesFolder = "images/"

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
	c.SetSrc(image.NewUniform(color.RGBA{R: 0x00, B: 0xff, A: 0xff}))
	c.SetFont(fnt)
	c.SetFontSize(18)
	if _, err := c.DrawString(time.Now().Format("2006-01-02 15:04:05"), freetype.Pt(5, bounds.Max.Y-4)); err != nil {
		return err
	}

	//ensure output folder exists
	if _, err := os.Stat(ImagesFolder); os.IsNotExist(err) {
		os.Mkdir(ImagesFolder, os.ModePerm)
	}

	if f, err := os.Create(ImagesFolder + filename); err != nil {
		return err
	} else {
		jpeg.Encode(f, i, &jpeg.Options{Quality: 100})
	}
	return nil
}
