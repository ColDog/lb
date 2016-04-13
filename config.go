package main

import (
	"log"
	"errors"
	"net/url"
	"net/http/httputil"
	"time"
)

var ConfigMethod = "FILE"

// loads in the configuration for a host from a generic map in json format.
// Each handler
// if a configuration engine is being used, it will load the configuration
// from this engine.
func loadConfig() {
	if ConfigMethod == "FILE" {

	}
}

// listens to the configuration engine for a change in the
// configuration for a single host. If using file only, then
// a worker runs and re-reads the config file every minute.
func listenForConfigChange() {
	if ConfigMethod == "FILE" {
		go func() {
			time.Sleep(1 * time.Minute)
			loadConfig()
		}()
	}
}

func AddHandler(path string, middleware []string) error {
	handlers[path] = &Handler{
		path: path,
		hosts: make([]Host, 0),
		middleware: middleware,
		lastHost: 0,
	}

	log.Printf("Adding Handler: %s", path)

	return nil
}

func AddHost(path, target, health string) error {
	if _, ok := handlers[path]; !ok {
		return errors.New("Could not find a handler to add this host to")
	}

	if target == "" {
		return errors.New("Improperly defined host")
	}

	dest, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(dest)

	hostStruct := &Host{
		health: health,
		target: target,
		checked: 0,
		healthy: true,
		proxy: proxy,
	}


	handlers[path].hosts = append(handlers[path].hosts, *hostStruct)
	return nil
}
