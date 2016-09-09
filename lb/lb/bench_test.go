package lb

import (
	"net/http"
	"fmt"
	"testing"
	"time"
	"github.com/coldog/proxy/lb/stats"
)

var s stats.StatsCollector = stats.New(stats.MEMORY)

func listen(p int) {
	http.ListenAndServe(fmt.Sprintf(":%d", p), func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello!"))
		w.WriteHeader(200)
	})
}

func load(path string, n int, t time.Duration)  {
	start := time.Now()

	for i := 0; i < n; i++ {
		go func() {
			for {
				httpClient := &http.Client{}
				resp, err := httpClient.Get("http://localhost:9888" + path)

				if err != nil {
					return err
				}

				s.SetIncrement(statusCodeName(resp), 1)

				if start.Add(t).After(time.Now()) {
					return
				}
			}
		}()
	}
}

func TestLoad(t *testing.T) {

	for i := 0; i < 5; i++ {
		listen(3000 + i)
	}

	load("/test", 20, 30 * time.Second)

}