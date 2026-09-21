package api

import (
	"fmt"
	"net/http"

	"rift/internal/model"
)

// handleStream implements the Real-Time Telemetry Explorer / Control Tower's
// live feed via Server-Sent Events: a single, dependency-free, browser-native
// transport (EventSource) carrying telemetry, events and alerts for one
// twin. The same fan-out points (Bus.Subscribe, Pipeline.Subscribe) are what
// a WebSocket/gRPC-streaming adapter would attach to in a deployment that
// needs bidirectional streaming.
func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	twinID := r.PathValue("id")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	readingCh := s.Pipeline.Subscribe(256)
	eventCh := make(chan model.Event, 256)
	s.Bus.Subscribe("*", func(evt model.Event) {
		if evt.TwinID == twinID {
			select {
			case eventCh <- evt:
			default:
			}
		}
	})

	ctx := r.Context()
	fmt.Fprintf(w, "retry: 2000\n\n")
	flusher.Flush()
	for {
		select {
		case <-ctx.Done():
			return
		case reading := <-readingCh:
			if reading.TwinID != twinID {
				continue
			}
			writeSSE(w, "telemetry", reading)
			flusher.Flush()
		case evt := <-eventCh:
			writeSSE(w, "event", evt)
			flusher.Flush()
		}
	}
}

func writeSSE(w http.ResponseWriter, event string, v interface{}) {
	data, err := jsonCompact(v)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
}
