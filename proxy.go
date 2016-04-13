package main

import (
	"net/http"
	"fmt"
	"log"
	"time"
	"strings"
)

var handlers map[string] *Handler = make(map[string] *Handler)

func handle(writer http.ResponseWriter, req *http.Request) {
	key := strings.Split(req.URL.Path, "/")[1]
	if _, ok := handlers[key]; ok {
		server := handlers[key].next()
		log.Printf("proxying %s to %s", req.URL.Path, server.target)
		handlers[key].run(&server, writer, req)

		server.proxy.ServeHTTP(writer, req)
	} else {
		log.Printf("handler not found %s", key)
		http.NotFound(writer, req)
	}
}

func proxy(port int) {
	log.Printf("proxying on :%d\n", port)

	http.HandleFunc("/", func(writer http.ResponseWriter, req *http.Request) {
		handle(writer, req)
	})
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("Fatal: %v", err)
	}
}

func healthCheck() {
	go func() {
		for  {
			for _, handler := range handlers {
				for _, host := range handler.hosts {
					host.ping()
				}
				time.Sleep(30 * time.Second)
			}
		}
	}()
}

func Start(port int) {
	initializeMiddleware()
	listenForConfigChange()
	healthCheck()
	proxy(port)
}

func main() {
	AddHandler("users", []string{"logger"})
	AddHost("users", "http://localhost:8000", "http://localhost:8000/check")
	AddHost("users", "http://localhost:8001", "http://localhost:8001/check")
	AddHost("users", "http://localhost:8002", "http://localhost:8002/check")
	AddHost("users", "http://localhost:8003", "http://localhost:8003/check")
	AddHost("users", "http://localhost:8004", "http://localhost:8004/check")
	AddHost("users", "http://localhost:8005", "http://localhost:8005/check")
	AddHost("users", "http://localhost:8006", "http://localhost:8006/check")

	Start(3000)
}
