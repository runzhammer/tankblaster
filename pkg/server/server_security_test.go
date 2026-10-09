package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/runzhammer/tankblaster/pkg/protocol"
)

func TestDefaultConfigSecurityLimits(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.WebSocketReadLimit <= 0 {
		t.Fatal("default websocket read limit is not set")
	}
	if cfg.Server.MaxConnections <= 0 {
		t.Fatal("default max connections is not set")
	}
	if cfg.Server.MaxSessions <= 0 {
		t.Fatal("default max sessions is not set")
	}
	if cfg.Server.ReadHeaderTimeout <= 0 || cfg.Server.IdleTimeout <= 0 {
		t.Fatal("default HTTP timeouts are not set")
	}
	if cfg.Server.ReadTimeout != 0 || cfg.Server.WriteTimeout != 0 {
		t.Fatal("long-lived websocket timeouts must stay disabled")
	}
}

func TestBundledServerConfigsAllowAllLobbySlots(t *testing.T) {
	for _, path := range []string{"../../config/server.example.yaml", "../../docker/server.yaml"} {
		cfg, err := LoadConfig(path)
		if err != nil {
			t.Fatalf("LoadConfig(%q) error = %v", path, err)
		}
		if got, want := cfg.Sessions.MaxPlayers, 10; got < want {
			t.Fatalf("%s max players = %d, want at least %d lobby slots", path, got, want)
		}
	}
}

func TestJoinPageEscapesToken(t *testing.T) {
	s := New(DefaultConfig(), nil)
	req := httptest.NewRequest(http.MethodGet, "/join/%3Cscript%3Ealert(1)%3C/script%3E", nil)
	rec := httptest.NewRecorder()

	s.handleJoinPage(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "<script>") {
		t.Fatalf("join page contains unescaped script tag: %s", body)
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Fatalf("join page does not contain escaped token: %s", body)
	}
}

func TestBroadcastSkipsClosedClient(t *testing.T) {
	cfg := DefaultConfig()
	h := New(cfg, nil)
	done := make(chan struct{})
	close(done)
	client := &Client{send: make(chan protocol.Envelope), done: done}
	sess := &Session{Players: []SessionPlayer{{PlayerID: "p1", Client: client}}}

	if err := h.broadcast(sess, protocol.TypePong, struct{}{}); err != nil {
		t.Fatalf("broadcast() error = %v", err)
	}
}
