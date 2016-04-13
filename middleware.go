package main

import (
	"log"
	"net/http"
)

type Request struct {
	handler 	*Handler
	req 		*http.Request
	Host		*Host
}

type Response struct {
	writer 		http.ResponseWriter
}

type BaseMiddleware func(req *Request, res *Response) (*Request, *Response)

var middleware map[string] BaseMiddleware = make(map[string] BaseMiddleware)

func initializeMiddleware()  {
	middleware["logger"] = logger
}

func logger(req *Request, res *Response) (*Request, *Response) {
	res.writer.Header().Add("X-Test", "hello there")
	log.Printf("I am some logging middleware! request %s\n", req.handler.path)
	return req, res
}
