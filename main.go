package main

import (
	"crypto/subtle"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

var (
	openDuration = flag.Int("duration", 5, "how long the valve should be open for in minutes")
	openPeriod   = flag.Int("period", 30, "time between valve open periods in minutes")
	test         = flag.Bool("test", false, "use the fake test relay")
	username     = flag.String("username", "", "username for http authentication")
	password     = flag.String("password", "", "password for http authentication")
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
	if err := p.relay.startPump(); err != nil {
		log.Printf("failed to start pump: %v", err)
	}
	if err := p.relay.openValve(); err != nil {
		log.Printf("failed to open valve: %v", err)
	}
}

func (p *PumpController) Stop() {
	log.Printf("stopping pump")
	//p.enabled = false
	if err := p.relay.stopPump(); err != nil {
		log.Printf("failed to stop pump: %v", err)
	}
	if err := p.relay.closeValve(); err != nil {
		log.Printf("failed to close valve: %v", err)
	}
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

	log.Printf("server started on https://localhost:8443")
	log.Fatal(http.ListenAndServeTLS(":8443", "cert.pem", "key.pem",
		createMux(controller)))
}

var (
	templates = template.Must(template.New("").
		ParseGlob("templates/*.html"))
)

func auth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if ok &&
			subtle.ConstantTimeCompare([]byte(user), []byte(*username)) == 1 &&
			subtle.ConstantTimeCompare([]byte(pass), []byte(*password)) == 1 {
			fn(w, r)
		}
		w.Header().Set("WWW-Authenticate",
			fmt.Sprintf("Basic realm=\"%s\"", "Restricted"))
		http.Error(w, "", 401)
	}
}

// createMux returns an HTTP router that serves HTTP requests for different
// routes.
func createMux(controller *PumpController) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/api/v1/start", auth(func(w http.ResponseWriter, r *http.Request) {
		controller.Start()
	}))
	m.HandleFunc("/api/v1/stop", auth(func(w http.ResponseWriter, r *http.Request) {
		controller.Stop()
	}))
	m.HandleFunc("/api/v1/state", auth(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	m.HandleFunc("/", auth(homeHandler))
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
