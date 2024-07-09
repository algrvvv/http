package http

import "net/http"

type Request struct {
	Method string
	URL    string
	Body   []byte
	Header http.Header
}
