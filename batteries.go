package van

import (
	"math/rand"
	"time"
	"golang.org/x/exp/io/i2c"
	"fmt"
)

func MonitorBatteries() {
	go func() {
		//fake three batteries
		for n := 0; n < 4; n++ {
			CurrentStats.Batteries = append(CurrentStats.Batteries, &Battery{})
		}

		for {
			//fake
			for _, battery := range CurrentStats.Batteries {
				//voltage in range from dead to full
				battery.Voltage = EmptyVoltage + (FullChargeVoltage-EmptyVoltage)*rand.Float32()
				battery.Load = MaxLoad * rand.Float32()
			}
			time.Sleep(1500 * time.Millisecond)
		}
	}()
}

type Battery struct {
	Voltage float32 //draining when < FullChargeVoltage
	Load    float32 //draining when positive, charging when negative
}

const EmptyVoltage = 10.5
const FullChargeVoltage = 13.7
const MaxLoad = 160 //Amps

//charge is the ratio from 0.0 to 1.0
//when <= 1.0 the battery is discharging
//when greater than 1.0 it's charging
func (this *Battery) Charge() float32 {
	//use empty voltage as a base line

	return (this.Voltage - EmptyVoltage) / (FullChargeVoltage - EmptyVoltage)
}

func MonitorAdc() {
	go func() {
		//channels
		const LightChannel = 0x00
		const ThermistorChannel = 0x01
		const AIN0 = 0x00
		const AIN1 = 0x01
		const AIN2 = 0x02
		const AIN3 = 0x03
		const PotChannel = 0x03

		//configuration

		//all single input: An->outAn where n = 0..3
		const FOUR_SINGLE = 0x0 << 4

		// three differential. (+A0,-A3)->outA0, (+A1,-A3)->outA1, (+A2,-A3)->outA2
		const THREE_DIFF = 0x1 << 4

		// single: An->outAn where n = 0..1
		// differential (+A2,-A3) -> outA2
		const TWO_SINGLE_ONE_DIFF = 0x2 << 4

		// differential: (+A0,-A1)->outA0 (+A2,-A3)->outA1
		const TWO_DIFF = 0x3 << 4

		//open the adc
		d, err := i2c.Open(&i2c.Devfs{Dev: "/dev/i2c-1"}, 0x48)
		if err != nil {
			fmt.Println("adc open error", err)
		}
		defer d.Close()

		// write to control register
		if err = d.Write([]byte{FOUR_SINGLE | AIN0}); err != nil {
			fmt.Println("adc write error", err)
		}
		//twos := func(b byte) int8 {
		//	if b>>7 == 0 {
		//		return int8(b)
		//	} else {
		//		b = b - (1 << 7)
		//		return int8(b) - 127 - 1
		//	}
		//}
		read := make([]byte, 0x1)
		for {
			//request to read
			if err = d.Read(read); err != nil {
				fmt.Println("adc read error", err)
			} else {
				//convert two's complement
				n := read[0] //twos(read[0])
				fmt.Print(n, " ")
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
}
