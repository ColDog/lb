package lb

import (
	"github.com/coldog/proxy/lb/ctx"
	"github.com/coldog/proxy/lb/router"
	"github.com/coldog/proxy/lb/stats"

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Middleware func(c *ctx.Context)

func New(c *Config) *Server {
	return &Server{
		config:     c,
		handlers:   map[string]*Handler{},
		middleware: map[string]Middleware{},
		router:     router.New(),
		lock:       &sync.RWMutex{},
		Stats:      &stats.NoOpStatsCollector{},
	}
}

type Server struct {
	Stats      stats.StatsCollector
	config     *Config
	handlers   map[string]*Handler
	middleware map[string]Middleware
	router     *router.Router
	lock       *sync.RWMutex
}

func (s *Server) Middleware(key string, m Middleware) {
	s.middleware[key] = m
}

func (s *Server) PutHandler(handler *Handler) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if handler.quit == nil {
		handler.quit = make(chan struct{})
		handler.stats = s.Stats
	}

	s.clearHandler(handler.Name)

	s.handlers[handler.Name] = handler
	for _, r := range handler.Routes {
		s.router.Add(handler.Name, r)
	}
	for _, t := range handler.Targets {
		t.stats = s.Stats
	}
}

func (s *Server) HasHandler(name string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, ok := s.handlers[name]
	return ok
}

func (s *Server) AddTarget(name string, target *Target) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	target.stats = s.Stats
	if h, ok := s.handlers[name]; ok {
		h.Targets = append(h.Targets, target)
	}
}

func (s *Server) RemoveHandler(name string) {
	s.lock.Lock()
	h, ok := s.handlers[name]
	s.lock.Unlock()

	if !ok {
		return
	}

	h.draining = true
	time.Sleep(h.ShutdownWait)

	s.lock.Lock()
	defer s.lock.Unlock()

	s.clearHandler(name)
	s.router.Remove(name)
}

func (s *Server) clearHandler(name string) {
	if h, ok := s.handlers[name]; ok {
		h.Close()
		delete(s.handlers, name)

		for _, t := range h.Targets {
			if t.tr != nil {
				t.tr.CloseIdleConnections()
			}
		}
	}
}

func (s *Server) UpdateTargetWeight(name, targetId string, weight int) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if h, ok := s.handlers[name]; ok {
		for _, t := range h.Targets {
			if targetId == t.ID {
				t.Weight = weight
			}
		}
	}
}

func (s *Server) RemoveTarget(name, targetId string) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if h, ok := s.handlers[name]; ok {
		for i, t := range h.Targets {
			if targetId == t.ID {
				if t.tr != nil {
					t.tr.CloseIdleConnections()
				}
				h.Targets = append(h.Targets[:i], h.Targets[i+1:]...)
				break
			}
		}
	}
}

func (s *Server) handler(key string) *Handler {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if handler, ok := s.handlers[key]; ok {
		return handler
	}
	return nil
}

func (s *Server) Start() {
	listen := fmt.Sprintf("%s:%d", s.config.Bind, s.config.Port)
	log.Printf("[INFO] listening %s", listen)
	http.ListenAndServe(listen, s)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/_lb/handlers" {
		data, err := json.Marshal(s.handlers)
		if err != nil {
			log.Printf("[ERROR] failed to print json %v", err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
		return
	}

	key := s.router.Match(r)
	handler := s.handler(key)

	c := ctx.New(w, r)

	if handler == nil {
		c.NoneAvailable()
		return
	}

	err := handler.Process(c)
	if err != nil {
		log.Printf("[ERROR] error processing request %v", err)
	}

	backend := handler.Next(c)
	if backend == nil {
		c.NoneAvailable()
		return
	}

	// run middleware
	for _, mid := range handler.Middleware {
		if m, ok := s.middleware[mid]; ok {
			m(c)
			if c.Finished {
				break
			}
		}
	}

	if c.Finished {
		return
	}

	backend.Proxy(handler, r).ServeHTTP(w, r)
}
