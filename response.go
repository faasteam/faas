package faas

import (
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
