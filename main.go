package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"time"
)

var (
	openDuration = flag.Int("duration", 5, "how long the valve should be open for in minutes")
	openPeriod   = flag.Int("period", 30, "time between valve open periods in minutes")
	test         = flag.Bool("test", false, "use the fake test relay")
)

type PumpController struct {
	enabled bool
	relay   Relay
}

func (p *PumpController) Run() {
	ticker := time.NewTicker(time.Duration(*openPeriod) * time.Second)
	for {
		select {
		case <-ticker.C:
			if !p.enabled {
				log.Printf("not enabled.")
				continue
			}
			p.relay.openValve()
			p.relay.startPump()
			time.Sleep(time.Duration(*openDuration) * time.Second)
			p.relay.stopPump()
			p.relay.closeValve()
		}
	}
}

func (p *PumpController) Start() {
	log.Printf("starting pump")
	//p.enabled = true
	p.relay.startPump()
	p.relay.openValve()
}

func (p *PumpController) Stop() {
	log.Printf("stopping pump")
	//p.enabled = false
	p.relay.stopPump()
	p.relay.closeValve()
}

func (p *PumpController) RelayState() (valveOpen bool, pumpOn bool) {
	return p.relay.state()
}

func main() {
	flag.Parse()

	var relay Relay
	if *test {
		relay = &fakeRelay{}
	} else {
		var err error
		relay, err = NewRelay()
		if err != nil {
			log.Fatalf("failed to start: %v", err)
		}
	}

	controller := &PumpController{enabled: true, relay: relay}

	//go controller.Run()

	// Log time in microseconds and filenames with log messages.
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	log.Printf("server started on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", createMux(controller)))
}

var (
	templates = template.Must(template.New("").
		ParseGlob("templates/*.html"))
)

// createMux returns an HTTP router that serves HTTP requests for different
// routes.
func createMux(controller *PumpController) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/api/v1/start", func(w http.ResponseWriter, r *http.Request) {
		controller.Start()
	})
	m.HandleFunc("/api/v1/stop", func(w http.ResponseWriter, r *http.Request) {
		controller.Stop()
	})
	m.HandleFunc("/api/v1/state", func(w http.ResponseWriter, r *http.Request) {
		valveOpen, pumpOn := controller.RelayState()
		type Response struct {
			ValveOpen bool `json:"valve_open"`
			PumpOn    bool `json:"pump_on"`
		}

		js, err := json.Marshal(Response{ValveOpen: valveOpen, PumpOn: pumpOn})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})
	m.HandleFunc("/", homeHandler)
	return m
}

// homeHandler is an HTTP request handler for the main page.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "home.html",
		map[string]interface{}{}); err != nil {
		log.Print(err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}
