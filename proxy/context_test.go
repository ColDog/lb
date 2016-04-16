package proxy

import (
	"testing"
	"net/http"
)

type MockWriter struct {
	headers 	http.Header
	status 		int
	body 		[]byte
}

func (w MockWriter) Header() http.Header {
	return http.Header{}
}

func (w MockWriter) WriteHeader(status int) {
	w.status = status
}

func (w MockWriter) Write(body []byte) (int, error) {
	w.body = body
	return len(body), nil
}


func TestName(t *testing.T) {
	ctx := &Context{
		writer: MockWriter{},
		req: &http.Request{},
		finished: false,
		allowProxy: true,
	}

	if ctx.allowProxy == false {
		t.Fail()
	}

	ctx.Finish()

	if ctx.allowProxy == true {
		t.Fail()
	}
}
