package handler

import (
	"fmt"
	"log"
	"time"
)

func (h *handler) logRequestIn() {
	msg := fmt.Sprintf(
		"%s > %s %s %s %d",
		h.RequestID,
		h.Request.RemoteAddr,
		h.Request.Method,
		h.Request.URL,
		len(h.RequestBody),
	)
	log.Println(msg)
	h.addTestAccessLogEntry(msg)
}

func (h *handler) logResponseWriteErr(err error) {
	msg := fmt.Sprintf(
		"%s !! failed to send response to %s %s %s -> %v",
		h.RequestID,
		h.Request.RemoteAddr,
		h.Request.Method,
		h.Request.URL,
		err,
	)
	log.Println(msg)
	h.addTestAccessLogEntry(msg)
}

func (h *handler) logResponseOut() {
	timeTaken := time.Since(h.RequestedAt)
	msg := fmt.Sprintf(
		"%s < %s %s %s %d -> %d %d in %v",
		h.RequestID,
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

func (h *handler) logPostResponseOut() {
	timeTaken := time.Since(h.RespondedAt)
	msg := fmt.Sprintf(
		"%s p %s %s %s in %v",
		h.RequestID,
		h.Request.RemoteAddr,
		h.Request.Method,
		h.Request.URL,
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
