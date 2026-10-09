package tankblaster

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/runzhammer/tankblaster/pkg/protocol"
)

func TestInviteTokenFromInputKeepsLongInviteToken(t *testing.T) {
	token := strings.Repeat("a", 64)
	input := "https://tankblaster.example/join/" + token

	got := inviteTokenFromInput(input)
	if got != token {
		t.Fatalf("inviteTokenFromInput() = %q, want %q", got, token)
	}
}

func TestTruncateRunesAllowsConfiguredJoinInputLength(t *testing.T) {
	input := "https://tankblaster.example/join/" + strings.Repeat("a", 64)

	got := truncateRunes(input, 512)
	if got != input {
		t.Fatalf("truncateRunes() shortened a normal invite link to %q", got)
	}
}

func TestOnlineClientReconnectsAndSendsHelloAgain(t *testing.T) {
	var hellos atomic.Int32
	errs := make(chan error, 2)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			errs <- err
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")
		var env protocol.Envelope
		if err := wsjson.Read(r.Context(), conn, &env); err != nil {
			errs <- err
			return
		}
		if env.Type != protocol.TypeHello {
			errs <- errUnexpectedMessage(env.Type)
			return
		}
		n := hellos.Add(1)
		ack, err := protocol.Wrap(protocol.TypeHelloAck, protocol.HelloAck{PlayerID: "p1", DisplayName: "Player", Rating: 1000})
		if err != nil {
			errs <- err
			return
		}
		if err := wsjson.Write(r.Context(), conn, ack); err != nil {
			errs <- err
			return
		}
		if n == 1 {
			_ = conn.Close(websocket.StatusNormalClosure, "force reconnect")
			return
		}
		<-r.Context().Done()
	}))
	defer srv.Close()

	client := newOnlineClientWithConfig("Player", onlineClientConfig{
		serverURL:         "ws" + strings.TrimPrefix(srv.URL, "http"),
		keepAliveInterval: 20 * time.Millisecond,
		pingTimeout:       20 * time.Millisecond,
		reconnectMinDelay: 10 * time.Millisecond,
		reconnectMaxDelay: 20 * time.Millisecond,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	acks := 0
	for acks < 2 {
		select {
		case <-ctx.Done():
			t.Fatalf("received %d hello acknowledgements, want reconnect acknowledgement too", acks)
		case err := <-errs:
			t.Fatalf("server error: %v", err)
		case env := <-client.recv:
			if env.Type == protocol.TypeHelloAck {
				acks++
			}
		}
	}
	if got := hellos.Load(); got < 2 {
		t.Fatalf("server received %d hello messages, want at least 2", got)
	}
}

type errUnexpectedMessage protocol.MessageType

func (e errUnexpectedMessage) Error() string {
	return "unexpected message: " + string(e)
}
