package tools

import (
	"testing"
)

func TestMap(t *testing.T) {
	m := Map{json: []byte(`{"name": 123}`)}
	if m.Int64("name") != int64(123) {
		t.Fail()
	}
}
