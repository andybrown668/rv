package van

import (
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func StartHttpApi() {
	//json stats
	http.HandleFunc("/live/stats", func(w http.ResponseWriter, r *http.Request) {
		if jsonStats, err := json.MarshalIndent(CurrentStats, "", "    "); err == nil {
			fmt.Fprint(w, string(jsonStats))
		} else {
			fmt.Fprint(w, err.Error())
		}
	})

	fs := http.FileServer(http.Dir(ImagesFolder))
	http.Handle("/", fs)

	http.HandleFunc("/all", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		//emit simple html of all images
		writeFilmStrip(w)
	})

	//display image
	http.HandleFunc("/live/display", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, dst)
	})
	//webcam
	http.HandleFunc("/live/webcamraw", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Refresh", "15;url=webcamraw")
		mu.Lock()
		defer mu.Unlock()
		//if frame1 != nil {
		//	png.Encode(w, frame1)
		//}
		if file, err := os.Open(ImagesFolder + "now.jpg"); err == nil {
			io.Copy(w, file)
			file.Close()
		} else {
			fmt.Println("failed to read now.jpg:", err)
		}
	})
	http.HandleFunc("/live/webcam", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Refresh", "15;url=webcam")
		mu.Lock()
		defer mu.Unlock()
		if gray3 != nil {
			png.Encode(w, gray3)
		}
	})
	go func() {
		if err := http.ListenAndServe(":80", nil); err != nil {
			fmt.Println(err)
		}
	}()
}

func writeFilmStrip(w io.Writer) {
	fmt.Fprint(w, "<html><body>")
	filepath.Walk(ImagesFolder, func(path string, info os.FileInfo, err error) error {
		//ignore folders
		if info.IsDir() {
			if info.Name() == "images" {
				return nil
			}
			return filepath.SkipDir
		}
		fmt.Fprintf(w, "<img src=%s />", info.Name())
		return nil
	})
	fmt.Fprint(w, "</body></html")

}
