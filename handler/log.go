package handler

import (
	"fmt"
	"log"
	"time"
)

func (h *handler) logRequestIn(location string) {
	msg := fmt.Sprintf(
		"%-12s > %s %s %s %d",
		location,
		h.Request.RemoteAddr,
		h.Request.Method,
		h.Request.URL,
		len(h.RequestBody),
	)
	log.Println(msg)
	h.addTestAccessLogEntry(msg)
}

func (h *handler) logResponseWriteErr(location string, err error) {
	msg := fmt.Sprintf(
		"%-12s !! failed to send response to %s %s %s -> %v",
		location,
		h.Request.RemoteAddr,
		h.Request.Method,
		h.Request.URL,
		err,
	)
	log.Println(msg)
	h.addTestAccessLogEntry(msg)
}

func (h *handler) logResponseOut(location string) {
	timeTaken := time.Since(h.RequestedAt)
	msg := fmt.Sprintf(
		"%-12s < %s %s %s %d -> %d %d in %v",
		location,
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

func (h *handler) logPostResponseOut(location string) {
	timeTaken := time.Since(h.RespondedAt)
	msg := fmt.Sprintf(
		"%-12s p %s %s %s in %v",
		location,
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
