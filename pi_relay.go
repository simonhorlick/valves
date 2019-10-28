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
	err := r.valvePin.Out(gpio.High)
	if err != nil {
		return err
	}

	r.valveOpen = true
	return nil
}

func (r *PiRelay) closeValve() error {
	err := r.valvePin.Out(gpio.Low)
	if err != nil {
		return err
	}

	r.valveOpen = false
	return nil
}

func (r *PiRelay) startPump() error {
	err := r.pumpPin.Out(gpio.High)
	if err != nil {
		return err
	}

	r.pumpOn = true
	return nil
}

func (r *PiRelay) stopPump() error {
	err := r.pumpPin.Out(gpio.Low)
	if err != nil {
		return err
	}

	r.pumpOn = false
	return nil
}

func (r *PiRelay) state() (valveOpen bool, pumpOn bool) {
	return r.valveOpen, r.pumpOn
}
