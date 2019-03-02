package main

import (
	"fmt"
	"time"

	"github.com/andybrown668/rv"
)

func main() {
	//go van.MonitorWebcam()
	van.MonitorWifi()
	van.MonitorDHT()
	van.MonitorBatteries()
	van.MonitorAdc()
	van.StartHttpApi()

	if err := van.Initialize(); err != nil {
		panic(err)
	}

	//monitor
	for {
		title := fmt.Sprintf("C&R %2dc %2d%% ", van.CurrentStats.Temperature, van.CurrentStats.Humidity)
		if van.IsOnline {
			title += "*"
		} else {
			title += "?"
		}

		lines := []string{
			title,
		}
		if err := van.Display(lines); err != nil {
			fmt.Println("failed to display:", err)
		}
		time.Sleep(van.Refresh)
	}
}
