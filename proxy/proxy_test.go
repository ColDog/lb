package proxy

import (
	"testing"
	"time"
	"log"
	"os/exec"
	"os"
	"github.com/coldog/proxy/tools"
	"fmt"
)

func mockServer(seconds, port int) {
	dir, _ := os.Getwd()
	cmd := exec.Command(dir + "/../hello_world.js", string(port))
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(time.Duration(seconds) * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal("failed to kill: ", err)
		}
		log.Println("process killed as timeout reached")
	case err := <-done:
		if err != nil {
			log.Printf("server %d died %v", port, err)
		}
	}
}

func TestSimple(t *testing.T) {
	config := tools.NewMap(map[string] interface{} {
		"ip_hash": true,
		"routes": []string{"test"},
		"key": "test",
		"middleware": []string{"json"},
		"hosts": []map[string] interface{} {
			map[string] interface{} {"target": "http://localhost:3000", "health": "http://localhost:3001"},
		},
	})

	proxy := NewProxyServer(tools.NewMap(map[string] interface{} {
		"binds": "0.0.0.0",
		"port": 3000,
		"access_log": true,
	}))

	proxy.Update(config)

	proxy.Use("test", func(ctx *Context) *Context {
		return ctx
	})
}

func TestServers(t *testing.T) {

	go func() {
		mockServer(5, 8000)
		mockServer(6, 8001)
		mockServer(4, 8003)
	}()

	config := tools.NewMap(map[string] interface{} {
		"ip_hash": false,
		"routes": []string{
			"/test",
			"/test/*",
		},
		"key": "test",
		"middleware": []string{},
		"hosts": []map[string] interface{} {
			map[string] interface{} {"target": "http://localhost:8000", "health": "http://localhost:8000"},
			map[string] interface{} {"target": "http://localhost:8001", "health": "http://localhost:8001"},
			map[string] interface{} {"target": "http://localhost:8002", "health": "http://localhost:8002"},
		},
	})

	proxy := NewProxyServer(tools.NewMap(map[string] interface{} {
		"binds": "0.0.0.0",
		"port": 3000,
		"access_log": false,
	}))

	proxy.Update(config)

	go func() {
		proxy.Start()
	}()

	out, err := exec.Command("wrk", "-c", "20", "-d", "5", "http://localhost:3000/test/123").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", out)
	t.Fail()
}
