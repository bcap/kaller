package handler

import (
	"fmt"
	"log"
	"time"
)

func (h *handler) logEntry() {
	msg := fmt.Sprintf(
		"> %s %s %s %d",
		h.Request.RemoteAddr,
		h.Request.Method,
		h.Request.URL,
		len(h.RequestBody),
	)
	log.Println(msg)
	h.addTestAccessLogEntry(msg)
}

func (h *handler) logErr(err error) {
	msg := fmt.Sprintf(
		"!! failed to send response to %s %s %s -> %v",
		h.Request.RemoteAddr,
		h.Request.Method,
		h.Request.URL,
		err,
	)
	log.Println(msg)
	h.addTestAccessLogEntry(msg)
}

func (h *handler) logExit() {
	timeTaken := time.Since(h.Start)
	msg := fmt.Sprintf(
		"< %s %s %s %d -> %d %d in %v",
		h.Request.RemoteAddr,
		h.Request.Method,
		h.Request.URL,
		len(h.RequestBody),
		h.ResponseStatusCode,
		len(h.ResponseBody),
		timeTaken,
	)
	log.Println(msg)
	h.addTestAccessLogEntry(msg)
}

func (h *handler) addTestAccessLogEntry(msg string) {
	if !h.testCaptureAccessLog {
		return
	}
	h.testAccessLogMutex.Lock()
	h.testAccessLog = append(h.testAccessLog, msg)
	h.testAccessLogMutex.Unlock()
}
