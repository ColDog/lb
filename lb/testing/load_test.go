package testing

import (
	"github.com/coldog/proxy/lb/lb"
	"github.com/coldog/proxy/lb/router"

	"gopkg.in/tylerb/graceful.v1"

	"net/http"
	"fmt"
	"testing"
	"time"
	"sync"
	"io"
	"io/ioutil"
)

func listen(p int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "HI!")
	})

	graceful.Run(fmt.Sprintf(":%d", p), 30*time.Second, mux)
}

func load(path string, n int, v time.Duration, t time.Duration) {
	requests := 0
	s5xx := 0
	s4xx := 0
	conErr := 0
	times := []int64{}

	wg := sync.WaitGroup{}

	fmt.Printf("beginning load test %v threads for %v seconds\n", n, t.Seconds())

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			start := time.Now()
			httpClient := &http.Client{}

			for {
				if v > 0 {
					time.Sleep(v)
				}

				t1 := time.Now()
				resp, err := httpClient.Get("http://127.0.0.1:9888" + path)
				t2 := time.Now()

				times = append(times, t2.UnixNano() - t1.UnixNano())

				requests++

				if resp == nil {
					fmt.Printf("err: %v\n", err)
					conErr++
				} else {
					if resp.StatusCode >= 500 {
						s5xx++
					} else if resp.StatusCode >= 400 {
						s4xx++
					}
					io.Copy(ioutil.Discard, resp.Body)
					resp.Body.Close()
				}

				if start.Add(t).Before(time.Now()) {
					wg.Done()
					return
				}
			}
		}()
	}

	wg.Wait()

	fmt.Printf("          requests %d\n", requests)
	fmt.Printf("               5xx %d\n", s5xx)
	fmt.Printf("               4xx %d\n", s4xx)
	fmt.Printf(" connection errors %d\n", conErr)
	fmt.Printf("     avg resp time %v\n", time.Duration(avg(times)))
	fmt.Printf("  requests per sec %v\n", float64(requests) / t.Seconds())


	fmt.Println()

}

func avg(arr []int64) float64 {
	t := int64(0)
	for _, i := range arr {
		t += i
	}

	return float64(t) / float64(len(arr))
}

func TestLoad(t *testing.T) {

	for i := 0; i < 5; i++ {
		go listen(3000 + i)
	}

	l := lb.New(lb.DefaultConfig())

	l.PutHandler(&lb.Handler{
		Name: "test",
		Routes: []*router.Route{
			{Path: "/tests/"},
			{Path: "/v1/api/*"},
		},
		Strategy: "rand",
		Targets: []*lb.Target{
			{
				ID: "test-1",
				URL: "http://localhost:3000",
				Weight: 10,
			},
			{
				ID: "test-2",
				URL: "http://localhost:3001",
				Weight: 15,
			},
			{
				ID: "test-3",
				URL: "http://localhost:3002",
				Weight: 20,
			},
			{
				ID: "test-4",
				URL: "http://localhost:3003",
				Weight: 30,
			},
			{
				ID: "test-5",
				URL: "http://localhost:3003",
				Weight: 40,
			},
		},
	})

	go l.Start()

	time.Sleep(2 * time.Second)
	load("/tests/", 2, time.Duration(0), 30 * time.Second)
}
