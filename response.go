package faas

import (
	"bufio"
	"net"
	"net/http"
)

type ResponseWriter interface {
	http.ResponseWriter

	Status() int
	Size() int
}

var _ ResponseWriter = &Response{}

type Response struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func NewResponse(w http.ResponseWriter) *Response {
	return &Response{
		ResponseWriter: w,
	}
}

func (rw *Response) WriteHeader(statusCode int) {
	if rw.statusCode == 0 {
		rw.statusCode = statusCode
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *Response) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
		rw.ResponseWriter.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}

func (rw *Response) Status() int {
	return rw.statusCode
}

func (rw *Response) Size() int {
	return rw.written
}

func (rw *Response) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (rw *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrHijacked
}

func (rw *Response) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := rw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}
