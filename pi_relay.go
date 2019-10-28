package main

import (
	"errors"
	"fmt"
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
	state() (valveOpen bool, pumpOn bool)
}

type PiRelay struct {
	valveOpen bool
	pumpOn    bool

	valvePin gpio.PinOut
	pumpPin  gpio.PinOut
}

func NewRelay() (*PiRelay, error) {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Print(err)
		return nil, fmt.Errorf("failed to initialise periph: %v", err)
	}

	valvePin := gpioreg.ByName("P1_12")
	if valvePin == nil {
		log.Print("Failed to find GPIO pin")
		return nil, errors.New("failed to find GPIO pin")
	}

	pumpPin := gpioreg.ByName("P1_16")
	if pumpPin == nil {
		log.Print("Failed to find GPIO pin")
		return nil, errors.New("failed to find GPIO pin")
	}

	var relay = &PiRelay{
		valvePin: valvePin,
		pumpPin:  pumpPin,
	}

	// Initially in a stopped state.
	relay.stopPump()
	relay.closeValve()

	return relay, nil
}

func (r *PiRelay) openValve() error {
	r.valveOpen = true
	return r.valvePin.Out(gpio.High)
}

func (r *PiRelay) closeValve() error {
	r.valveOpen = false
	return r.valvePin.Out(gpio.Low)
}

func (r *PiRelay) startPump() error {
	r.pumpOn = true
	return r.pumpPin.Out(gpio.High)
}

func (r *PiRelay) stopPump() error {
	r.pumpOn = false
	return r.pumpPin.Out(gpio.Low)
}

func (r *PiRelay) state() (valveOpen bool, pumpOn bool) {
	return r.valveOpen, r.pumpOn
}
