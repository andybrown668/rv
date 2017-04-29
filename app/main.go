package main

import (
	"fmt"
	"time"

	"github.com/andybrown668/rv"
)

func main() {
	go van.MonitorWebcam()
	//monitorWifi()
	//monitorDHT()
	//monitorBatteries()
	//monitorAdc()
	van.StartHttpApi()

	if err := van.Initialize(); err != nil {
		panic(err)
	}

	//monitor
	blink := false
	for {
		blink = !blink
		title := fmt.Sprintf("C&R %2dc %2d%% ", van.CurrentStats.Temperature, van.CurrentStats.Humidity)
		//add a blinking * if online, - if not
		if blink {
			if van.IsOnline {
				title += "*"
			} else {
				title += "-"
			}

		}

		lines := []string{
			title,
		}
		if err := van.Display(lines); err != nil {
			fmt.Println("failed to display:", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
