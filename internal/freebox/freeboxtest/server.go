package freeboxtest

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

const (
	AppID     = "test-app"
	AppToken  = "test-app-token"
	challenge = "test-challenge"
	apiPrefix = "/api/v8/"
)

type Server struct {
	*httptest.Server

	mu      sync.Mutex
	epoch   int
	logins  int
	results map[string]string
}

func NewServer(t testing.TB, results map[string]string) *Server {
	t.Helper()
	s := &Server{results: results}
	mux := http.NewServeMux()
	mux.HandleFunc("/api_version", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"api_base_url":"/api/","api_version":"8.0"}`)
	})
	mux.HandleFunc(apiPrefix+"login/", s.handleChallenge)
	mux.HandleFunc(apiPrefix+"login/session/", s.handleSession)
	mux.HandleFunc(apiPrefix, s.handleResource)
	s.Server = httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

func (s *Server) RevokeSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.epoch++
}

func (s *Server) Logins() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.logins
}

func (s *Server) session() string {
	return fmt.Sprintf("session-%d", s.epoch)
}

func (s *Server) handleChallenge(w http.ResponseWriter, _ *http.Request) {
	writeResult(w, http.StatusOK, fmt.Sprintf(`{"logged_in":false,"challenge":%q}`, challenge))
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AppID    string `json:"app_id"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mac := hmac.New(sha1.New, []byte(AppToken))
	mac.Write([]byte(challenge))
	if body.AppID != AppID || body.Password != hex.EncodeToString(mac.Sum(nil)) {
		writeError(w, http.StatusForbidden, "invalid_token")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.logins++
	writeResult(w, http.StatusOK, fmt.Sprintf(`{"session_token":%q}`, s.session()))
}

func (s *Server) handleResource(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	valid := r.Header.Get("X-Fbx-App-Auth") == s.session()
	s.mu.Unlock()
	if !valid {
		writeError(w, http.StatusForbidden, "auth_required")
		return
	}
	result, ok := s.results[strings.TrimPrefix(r.URL.Path, apiPrefix)]
	if !ok {
		writeError(w, http.StatusNotFound, "no_such_resource")
		return
	}
	writeResult(w, http.StatusOK, result)
}

func writeResult(w http.ResponseWriter, status int, result string) {
	w.WriteHeader(status)
	if result == "" {
		_, _ = fmt.Fprint(w, `{"success":true}`)
		return
	}
	_, _ = fmt.Fprintf(w, `{"success":true,"result":%s}`, result)
}

func writeError(w http.ResponseWriter, status int, code string) {
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `{"success":false,"error_code":%q,"msg":%q}`, code, code)
}
