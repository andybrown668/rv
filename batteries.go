package van

import (
	"fmt"
	"golang.org/x/exp/io/i2c"
	"time"
)

func MonitorBatteries() {
	go func() {
		//fake three batteries
		CurrentStats.Batteries = append(CurrentStats.Batteries, &Battery{Name: "V", Voltage: EmptyVoltage})
		CurrentStats.Batteries = append(CurrentStats.Batteries, &Battery{Name: "L", Voltage: EmptyVoltage})
		CurrentStats.Batteries = append(CurrentStats.Batteries, &Battery{Name: "R", Voltage: EmptyVoltage})
		//CurrentStats.Batteries = append(CurrentStats.Batteries, &Battery{Name: "R", Voltage:EmptyVoltage+4})

		for {
			//for _, battery := range CurrentStats.Batteries {
			//voltage in range 0-255 from dead to full
			CurrentStats.Batteries[0].Voltage = CurrentStats.Adc0
			CurrentStats.Batteries[0].Load = CurrentStats.Adc2
			//}
			time.Sleep(Refresh)
		}
	}()
}

type Battery struct {
	Name    string
	Voltage int //draining when < FullChargeVoltage
	Load    int //draining when positive, charging when negative
}

// battery voltage indicates state of charge and is measured by the adc
// differential adc is not yet used, so the adc will only indicate 0 if the battery is not connected
// 0 indicates discharged, 150 charged, and above that, charging
const NoVoltage = 0
const EmptyVoltage = 0
const FullChargeVoltage = 220
const MaxLoad = 160 //Amps

func (this *Battery) ChargeRatio() float32 {
	if this.Voltage <= EmptyVoltage {
		return 0
	} else if this.Voltage >= FullChargeVoltage {
		return 1
	} else {
		return float32(this.Voltage-EmptyVoltage) / float32(FullChargeVoltage-EmptyVoltage)
	}
}

func (this *Battery) Disconnected() bool {
	return this.Voltage == 0
}

func MonitorAdc() {
	go func() {
		//channels
		const LightChannel = 0x00
		const ThermistorChannel = 0x01
		const PotChannel = 0x03
		const AIN0 = 0x00
		const AIN1 = 0x01
		const AIN2 = 0x02
		const AIN3 = 0x03
		const AUTO_INC = 0x04

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

		const ENABLE_OUTPUT = 0x1 << 6

		//open the adc
		d, err := i2c.Open(&i2c.Devfs{Dev: "/dev/i2c-1"}, 0x48)
		if err != nil {
			fmt.Println("adc open error", err)
		}
		defer d.Close()

		read := make([]byte, 0x1)
		for {
			// write to control register
			ctl := byte(FOUR_SINGLE | AIN0)
			//fmt.Printf("%08b\n", ctl)
			if err = d.Write([]byte{ctl}); err != nil {
				fmt.Println("adc write ain0 error", err)
			}
			//request to read
			if err = d.Read(read); err != nil {
				fmt.Println("adc read ain0 error", err)
			} else if err = d.Read(read); err != nil {
				fmt.Println("adc read ain0 error", err)
			} else {
				//fmt.Printf("%#x %#x\n",read[1], read[0])
				//convert two's complement
				n := read[0]
				CurrentStats.Adc0 = int(n)
				fmt.Printf("v=%d\t", n)
			}

			time.Sleep(Refresh)
			//// write to control register
			//if err = d.Write([]byte{TWO_SINGLE_ONE_DIFF | AIN2}); err != nil {
			//	fmt.Println("adc write error", err)
			//}
			////request to read
			//if err = d.Read(read); err != nil {
			//	fmt.Println("adc read ain2 error", err)
			//} else if err = d.Read(read); err != nil {
			//	fmt.Println("adc read ain2 error", err)
			//} else {
			//	//convert two's complement
			//	n := read[0]
			//	CurrentStats.Adc2 = twos(n)
			//	fmt.Printf("load=%d\n", n)
			//}
			//time.Sleep(Refresh)
		}
	}()
}

func twos(b byte) int {
	if b>>7 == 0 {
		return int(b)
	} else {
		b = b - (1 << 7)
		return int(b) - 127 - 1
	}
}
