package health

import (
	"net/http"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

type checksHandler struct {
	lock sync.RWMutex

	livenessPath   string
	livenessChecks map[string]Check

	readinessPath   string
	readinessChecks map[string]Check
}

// Checker ...
type Checker interface {
	AddLiveness(name string, check Check)
	AddReadiness(name string, check Check)
	Handler() http.Handler
	RegisterHandler(mux *http.ServeMux)
}

// NewChecksHandler accepts two strings: health and ready paths.
// These paths will be used for liveness and readiness checks.
func NewChecksHandler(healthzPath, readyPath string) Checker {
	if healthzPath[0] != '/' {
		healthzPath = "/" + healthzPath
	}
	if readyPath[0] != '/' {
		readyPath = "/" + readyPath
	}
	ch := &checksHandler{
		livenessPath:    healthzPath,
		livenessChecks:  map[string]Check{},
		readinessPath:   readyPath,
		readinessChecks: map[string]Check{},
	}

	return ch
}

func (ch *checksHandler) AddLiveness(name string, check Check) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	ch.livenessChecks[name] = check
}

func (ch *checksHandler) AddReadiness(name string, check Check) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	ch.readinessChecks[name] = check
}

// Handler returns a new http.Handler for the given health checker
func (ch *checksHandler) Handler() http.Handler {
	mux := http.NewServeMux()
	ch.registerMux(mux)
	return mux
}

// RegisterHandler registers the given health and readiness patterns onto the given http.ServeMux
func (ch *checksHandler) RegisterHandler(mux *http.ServeMux) {
	ch.registerMux(mux)
}

func (ch *checksHandler) registerMux(mux *http.ServeMux) {
	mux.HandleFunc(ch.readinessPath, ch.readyEndpoint)
	mux.HandleFunc(ch.livenessPath, ch.healthEndpoint)
}

func (ch *checksHandler) healthEndpoint(rw http.ResponseWriter, r *http.Request) {
	ch.handle(rw, r, ch.livenessChecks)
}

func (ch *checksHandler) readyEndpoint(rw http.ResponseWriter, r *http.Request) {
	ch.handle(rw, r, ch.readinessChecks)
}

func (ch *checksHandler) handle(rw http.ResponseWriter, r *http.Request, checksSets ...map[string]Check) {
	if r.Method != http.MethodGet {
		http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	errors := map[string]error{}
	status := http.StatusOK
	ch.lock.RLock()
	defer ch.lock.RUnlock()

	for _, checks := range checksSets {
		for name, check := range checks {
			if check == nil {
				continue
			}
			if err := check(); err != nil {
				status = http.StatusServiceUnavailable
				errors[name] = err
			}
		}
	}
	rw.WriteHeader(status)
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	if status == http.StatusOK {
		rw.Write([]byte("{}\n"))
	} else {
		encoder := jsoniter.NewEncoder(rw)
		encoder.SetIndent("", "    ")
		encoder.Encode(errors)
	}
}
