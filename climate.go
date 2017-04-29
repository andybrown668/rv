package van

import (
	"time"

	"github.com/d2r2/go-dht"
)

func MonitorDHT() {
	go func() {
		for {
			t, h, err := dht.ReadDHTxx(dht.DHT22, 10, true)
			if err == nil {
				CurrentStats.Temperature = int(t)
				CurrentStats.Humidity = int(h)
				time.Sleep(refresh)
			} else {
				time.Sleep(1500 * time.Millisecond)
			}
		}
	}()
}
