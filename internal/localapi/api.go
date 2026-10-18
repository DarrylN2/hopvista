// Package localapi exposes one same-origin, memory-only diagnostic-run API.
package localapi

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"

	"hopvista/internal/run"
	"hopvista/internal/target"
)

//go:embed static/index.html static/app.js static/style.css
var assets embed.FS

const tokenHeader = "X-HopVista-Token"

type Handler struct {
	host   string
	origin string
	token  string
	runs   *run.Manager
}

// NewHandler accepts an IPv4 loopback listener address. The caller must pass
// the actual address returned by net.Listen, as cmd/hopvista does.
func NewHandler(listenerAddr string, runs *run.Manager) (*Handler, error) {
	host, port, err := net.SplitHostPort(listenerAddr)
	ip := net.ParseIP(host)
	if err != nil || ip == nil || !ip.Equal(net.IPv4(127, 0, 0, 1)) || port == "0" || runs == nil {
		return nil, errors.New("an active 127.0.0.1 listener and run manager are required")
	}
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return nil, err
	}
	return &Handler{host: net.JoinHostPort("127.0.0.1", port), origin: "http://" + net.JoinHostPort("127.0.0.1", port), token: hex.EncodeToString(raw[:]), runs: runs}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if r.Host != h.host {
		jsonError(w, http.StatusForbidden, "invalid_host", "This request did not use the local listener address.")
		return
	}
	switch r.URL.Path {
	case "/api/v1/session":
		if r.Method != http.MethodGet {
			methodNotAllowed(w, "GET")
			return
		}
		jsonResponse(w, http.StatusOK, map[string]string{"mutation_token": h.token})
	case "/api/v1/runs":
		if r.Method != http.MethodPost {
			methodNotAllowed(w, "POST")
			return
		}
		if !h.authorizeMutation(w, r) {
			return
		}
		h.start(w, r)
	case "/":
		if r.Method != http.MethodGet {
			methodNotAllowed(w, "GET")
			return
		}
		data, _ := assets.ReadFile("static/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; base-uri 'none'; form-action 'self'")
		w.Write(data)
	case "/app.js":
		if r.Method != http.MethodGet {
			methodNotAllowed(w, "GET")
			return
		}
		data, _ := assets.ReadFile("static/app.js")
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		w.Write(data)
	case "/style.css":
		if r.Method != http.MethodGet {
			methodNotAllowed(w, "GET")
			return
		}
		data, _ := assets.ReadFile("static/style.css")
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Write(data)
	default:
		h.runPath(w, r)
	}
}

func (h *Handler) authorizeMutation(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Origin") != h.origin {
		jsonError(w, http.StatusForbidden, "invalid_origin", "The request origin must match the local application.")
		return false
	}
	provided := r.Header.Get(tokenHeader)
	if len(provided) != len(h.token) || subtle.ConstantTimeCompare([]byte(provided), []byte(h.token)) != 1 {
		jsonError(w, http.StatusForbidden, "invalid_token", "The local mutation token is missing or invalid.")
		return false
	}
	return true
}

type startRequest struct {
	TargetURL string          `json:"target_url"`
	MaxHops   json.RawMessage `json:"max_hops"`
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		jsonError(w, http.StatusUnsupportedMediaType, "json_required", "Send an application/json body.")
		return
	}
	decoder := json.NewDecoder(io.LimitReader(r.Body, 4097))
	decoder.DisallowUnknownFields()
	var payload startRequest
	if err := decoder.Decode(&payload); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid_request", "Provide a JSON object with a target URL and optional max_hops.")
		return
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		jsonError(w, http.StatusBadRequest, "invalid_request", "Only one JSON object is accepted.")
		return
	}
	if len(payload.TargetURL) > 2048 {
		jsonError(w, http.StatusBadRequest, "invalid_target_url", "The target URL is too long.")
		return
	}
	parsed, problem := target.Parse(payload.TargetURL)
	if problem != nil {
		jsonError(w, http.StatusBadRequest, problem.Code, problem.Message)
		return
	}
	hops := 30
	if payload.MaxHops != nil {
		if bytes.Equal(bytes.TrimSpace(payload.MaxHops), []byte("null")) {
			jsonError(w, http.StatusBadRequest, "invalid_max_hops", "max_hops must be an integer from 1 to 64.")
			return
		}
		if err := json.Unmarshal(payload.MaxHops, &hops); err != nil || hops < 1 || hops > 64 {
			jsonError(w, http.StatusBadRequest, "invalid_max_hops", "max_hops must be an integer from 1 to 64.")
			return
		}
	}
	snapshot, err := h.runs.Start(run.Input{TargetURL: parsed.String(), MaxHops: hops})
	if errors.Is(err, run.ErrBusy) {
		jsonError(w, http.StatusConflict, "busy", "A diagnostic run is already active.")
		return
	}
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "start_failed", "The diagnostic run could not start.")
		return
	}
	status := "/api/v1/runs/" + snapshot.RunID
	jsonResponse(w, http.StatusAccepted, map[string]string{"run_id": snapshot.RunID, "status_url": status, "cancel_url": status + "/cancel"})
}

func (h *Handler) runPath(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api/v1/runs/") {
		jsonError(w, http.StatusNotFound, "not_found", "No such endpoint.")
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/runs/")
	cancel := strings.HasSuffix(rest, "/cancel")
	if cancel {
		rest = strings.TrimSuffix(rest, "/cancel")
	}
	if len(rest) != 32 || strings.Contains(rest, "/") {
		jsonError(w, http.StatusNotFound, "not_found", "No such diagnostic run.")
		return
	}
	if _, err := hex.DecodeString(rest); err != nil {
		jsonError(w, http.StatusNotFound, "not_found", "No such diagnostic run.")
		return
	}
	if cancel {
		if r.Method != http.MethodPost {
			methodNotAllowed(w, "POST")
			return
		}
		if !h.authorizeMutation(w, r) {
			return
		}
		snapshot, err := h.runs.Cancel(rest)
		if errors.Is(err, run.ErrNotFound) {
			jsonError(w, http.StatusNotFound, "not_found", "No such diagnostic run.")
			return
		}
		jsonResponse(w, http.StatusOK, snapshot)
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	snapshot, err := h.runs.Get(rest)
	if errors.Is(err, run.ErrNotFound) {
		jsonError(w, http.StatusNotFound, "not_found", "No such diagnostic run.")
		return
	}
	jsonResponse(w, http.StatusOK, snapshot)
}

func jsonResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func jsonError(w http.ResponseWriter, status int, code, message string) {
	jsonResponse(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
func methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	jsonError(w, http.StatusMethodNotAllowed, "method_not_allowed", "This method is not supported for the endpoint.")
}
