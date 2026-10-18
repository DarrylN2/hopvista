package httpmeasure

import (
	"net/http/httptrace"
	"sync"
	"time"
)

type traceSnapshot struct {
	connectStart, connectEnd, tlsStart, tlsEnd, written, firstByte time.Time
	connectErr, tlsErr, writeErr                                   error
	gotConn, reused                                                bool
	peer                                                           string
}
type traceRecorder struct {
	mu sync.Mutex
	v  traceSnapshot
}

func (r *traceRecorder) setConnectStart() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.v.connectStart = time.Now().UTC()
}
func (r *traceRecorder) setConnectDone(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.v.connectEnd = time.Now().UTC()
	r.v.connectErr = err
}
func (r *traceRecorder) setTLSStart() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.v.tlsStart = time.Now().UTC()
}
func (r *traceRecorder) setTLSDone(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.v.tlsEnd = time.Now().UTC()
	r.v.tlsErr = err
}
func (r *traceRecorder) setWritten(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.v.written = time.Now().UTC()
	r.v.writeErr = err
}
func (r *traceRecorder) setFirstByte() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.v.firstByte = time.Now().UTC()
}

// Retain the validated TCP peer even when TLS fails before GotConn fires.
func (r *traceRecorder) setPeer(peer string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.v.peer = peer
}
func (r *traceRecorder) setGotConn(info httptrace.GotConnInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.v.gotConn = true
	r.v.reused = info.Reused
	if info.Conn != nil {
		if ip, ok := peerIPv4(info.Conn.RemoteAddr()); ok {
			r.v.peer = ip.String()
		}
	}
}
func (r *traceRecorder) snapshot() traceSnapshot { r.mu.Lock(); defer r.mu.Unlock(); return r.v }
