package van

import (
	"os/exec"
	"time"
)

//monitor connectivity state regularly
func MonitorWifi() {
	go func() {
		for {
			_, err := exec.Command(`/sbin/iwgetid`).Output()
			if err == nil {
				IsOnline = true
				//fmt.Println("Wifi is up")
			} else {
				IsOnline = false
				//fmt.Println("Wifi is down: ", err)
			}
			time.Sleep(Refresh)
		}
	}()
	return
}
