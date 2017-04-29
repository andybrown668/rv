package van

import (
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

//monitor connectivity state regularly
func MonitorWifi() {
	go func() {
		for {
			_, err := exec.Command(`iwgetid`).Output()
			if err == nil {
				IsOnline = true
			} else {
				IsOnline = false
			}
			time.Sleep(refresh)
		}
	}()
	return
}

func StartHttpApi() {
	//json stats
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if jsonStats, err := json.MarshalIndent(CurrentStats, "", "    "); err == nil {
			fmt.Fprint(w, string(jsonStats))
		} else {
			fmt.Fprint(w, err.Error())
		}
	})

	//display image
	http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, dst)
	})
	//webcam
	http.HandleFunc("/webcamraw", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Refresh", "15;url=webcamraw")
		mu.Lock()
		defer mu.Unlock()
		//if frame1 != nil {
		//	png.Encode(w, frame1)
		//}
		if file, err := os.Open("now.jpg"); err == nil {
			io.Copy(w, file)
			file.Close()
		} else {
			fmt.Println("failed to read now.jpg:", err)
		}
	})
	http.HandleFunc("/webcam", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Refresh", "15;url=webcam")
		mu.Lock()
		defer mu.Unlock()
		if gray3 != nil {
			png.Encode(w, gray3)
		}
	})
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println(err)
		}
	}()
}
