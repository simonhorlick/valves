package main

import (
	"log"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

type Relay interface {
	openValve() error
	closeValve() error
	startPump() error
	stopPump() error
}

type PiRelay struct {
	valvePin gpio.PinOut
	pumpPin  gpio.PinOut
}

func NewRelay() *PiRelay {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	valvePin := gpioreg.ByName("P1_12")
	if valvePin == nil {
		log.Fatal("Failed to find GPIO pin")
	}

	// if pumpPin := gpioreg.ByName("P1_12"); pumpPin == nil {
	// 	log.Fatal("Failed to find GPIO pin")
	// }

	return &PiRelay{
		valvePin: valvePin,
		// pumpPin:  gpio.PinOut,
	}
}

func (r *PiRelay) openValve() error {
	return r.valvePin.Out(gpio.High)
}

func (r *PiRelay) closeValve() error {
	return r.valvePin.Out(gpio.Low)
}

func (r *PiRelay) startPump() error {
	return r.pumpPin.Out(gpio.High)
}

func (r *PiRelay) stopPump() error {
	return r.pumpPin.Out(gpio.Low)
}
