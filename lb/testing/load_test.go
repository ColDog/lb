package testing

import (
	"net/http"
	"fmt"
	"testing"
	"time"
	"github.com/coldog/proxy/lb/lb"
	"github.com/coldog/proxy/lb/router"
	"sync"
)

type serve struct {}

func (s *serve) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello!"))
	w.WriteHeader(200)
}

func listen(p int) {
	http.ListenAndServe(fmt.Sprintf(":%d", p), &serve{})
}

func load(path string, n int, t time.Duration) {
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
				t1 := time.Now()
				resp, _ := httpClient.Get("http://localhost:9888" + path)
				t2 := time.Now()

				times = append(times, t2.UnixNano() - t1.UnixNano())

				requests++

				if resp == nil {
					conErr++
				} else {
					if resp.StatusCode >= 500 {
						s5xx++
					} else if resp.StatusCode >= 400 {
						s4xx++
					}
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
		Strategy: "wrr",
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
		},
	})

	go l.Start()

	load("/test", 20, 30 * time.Second)
}
