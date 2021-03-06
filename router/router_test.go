package router

import (
	"testing"
	"github.com/davecgh/go-spew/spew"
)

func TestRouting(t *testing.T) {
	router := Router()

	router.Default("fail")
	router.Add("/users/:id", "h1")
	router.Add("/users/:id/thing", "h2")
	router.Add("/testing/*", "h3")
	router.Add("/testing", "h3")

	res1, _ := router.Match("/testing/123?name=coldog")
	if res1 != "h3" {
		t.Fatal("Failed to match /testing/123 to /testing/*")
	}

	res2, params := router.Match("/users/123/thing?name=colin")
	if res2 != "h2" || params[":id"] != "123" {
		t.Fatal("Failed to match /users/123/thing to /users/:id/thing")
	}

	res3, _ := router.Match("/users/123")
	if res3 != "h1" {
		t.Fatal("Failed to match /users/123 to /users/:id")
	}

	res4, _ := router.Match("/non-existent")
	if res4 != "fail" {
		t.Fatal("Failed to return default handler")
	}

	res5, _ := router.Match("/testing")
	if res5 != "h3" {
		t.Fatal("Failed to match /testing to /testing/*")
	}

	spew.Println(router.base)
}
