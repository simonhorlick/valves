package main

import "log"

type fakeRelay struct {
	valveOpen bool
	pumpOn    bool
}

func (r *fakeRelay) openValve() error {
	log.Printf("open valve")
	r.valveOpen = true
	return nil
}

func (r *fakeRelay) closeValve() error {
	log.Printf("close valve")
	r.valveOpen = false
	return nil
}

func (r *fakeRelay) startPump() error {
	log.Printf("start pump")
	r.pumpOn = true
	return nil
}

func (r *fakeRelay) stopPump() error {
	log.Printf("stop pump")
	r.pumpOn = false
	return nil
}

func (r *fakeRelay) state() (valveOpen bool, pumpOn bool) {
	return r.valveOpen, r.pumpOn
}
