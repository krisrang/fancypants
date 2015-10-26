package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// 0 s "GET /iris/DSC05007.JPG HTTP/1.1" - "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.80 Safari/537.36"
const (
	ApacheFormatPattern = "%v \033[0m\033[36;1m%fs\033[0m \"%v\" %v %v \033[1;30m\"%v\"\033[0m"
)

type ApacheLogRecord struct {
	http.ResponseWriter

	ip, userAgent         string
	time                  time.Time
	method, uri, protocol string
	status                int
	responseBytes         int64
	elapsedTime           time.Duration
}

func (r *ApacheLogRecord) Log(out *log.Logger) {
	requestLine := fmt.Sprintf("%s %s %s", r.method, r.uri, r.protocol)
	out.Printf(ApacheFormatPattern, r.status, r.elapsedTime.Seconds(), requestLine, r.responseBytes, r.ip, r.userAgent)
}

func (r *ApacheLogRecord) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.responseBytes += int64(written)
	return written, err
}

func (r *ApacheLogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

type ApacheLoggingHandler struct {
	handler http.Handler
	out     *log.Logger
}

func NewApacheLoggingHandler(handler http.Handler, out *log.Logger) http.Handler {
	return &ApacheLoggingHandler{
		handler: handler,
		out:     out,
	}
}

func (h *ApacheLoggingHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}

	ua := r.UserAgent()

	record := &ApacheLogRecord{
		ResponseWriter: rw,
		ip:             clientIP,
		userAgent:      ua,
		time:           time.Time{},
		method:         r.Method,
		uri:            r.RequestURI,
		protocol:       r.Proto,
		status:         http.StatusOK,
		elapsedTime:    time.Duration(0),
	}

	startTime := time.Now()
	h.handler.ServeHTTP(record, r)
	finishTime := time.Now()

	record.time = finishTime.UTC()
	record.elapsedTime = finishTime.Sub(startTime)

	record.Log(h.out)
}
