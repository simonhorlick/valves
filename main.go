package main

import (
	"flag"
	"log"
	"time"

	"periph.io/x/periph/host"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/gpio"
)

var (
	openDuration = flag.Int("duration", 5, "how long the valve should be open for in minutes")
	openPeriod = flag.Int("period", 30, "time between valve open periods in minutes")
)

func openValve(p gpio.PinOut) {
	// Set the pin as output High.
	if err := p.Out(gpio.High); err != nil {
	    log.Fatal(err)
	}

	time.Sleep(time.Duration(*openDuration) * time.Second)

	if err := p.Out(gpio.Low); err != nil {
	    log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
	    log.Fatal(err)
	}

	p := gpioreg.ByName("P1_12")
	if p == nil {
	    log.Fatal("Failed to find GPIO pin")
	}

	ticker := time.NewTicker(time.Duration(*openPeriod) * time.Second)
	for {
		select {
		case <-ticker.C:
			openValve(p)
		}
	}
}
