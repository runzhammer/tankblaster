package tankblaster

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/runzhammer/tankblaster/pkg/core"
	"github.com/runzhammer/tankblaster/pkg/protocol"
)

type onlineIdentity struct {
	PlayerID    string `json:"player_id" yaml:"player_id,omitempty"`
	PlayerToken string `json:"player_token" yaml:"player_token,omitempty"`
	DisplayName string `json:"display_name" yaml:"display_name,omitempty"`
}

type onlineClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	send   chan protocol.Envelope
	recv   chan protocol.Envelope
	errs   chan error
	id     onlineIdentity
}

func newOnlineClient(displayName string) *onlineClient {
	ctx, cancel := context.WithCancel(context.Background())
	c := &onlineClient{
		ctx: ctx, cancel: cancel,
		send: make(chan protocol.Envelope, 16),
		recv: make(chan protocol.Envelope, 32),
		errs: make(chan error, 4),
		id:   loadOnlineIdentity(),
	}
	if displayName != "" {
		c.id.DisplayName = displayName
	}
	go c.run()
	return c
}

func onlineServerAvailable(ctx context.Context) bool {
	conn, _, err := websocket.Dial(ctx, core.Config().Online.ServerURL, nil)
	if err != nil {
		return false
	}
	_ = conn.Close(websocket.StatusNormalClosure, "")
	return true
}

func (c *onlineClient) Close() {
	c.cancel()
}

func (c *onlineClient) Send(typ protocol.MessageType, payload any) {
	env, err := protocol.Wrap(typ, payload)
	if err != nil {
		select {
		case c.errs <- err:
		default:
		}
		return
	}
	if c.ctx == nil {
		select {
		case c.send <- env:
		case <-time.After(250 * time.Millisecond):
			c.report(context.DeadlineExceeded)
		}
		return
	}
	select {
	case c.send <- env:
	case <-time.After(250 * time.Millisecond):
		c.report(context.DeadlineExceeded)
	case <-c.ctx.Done():
		c.report(c.ctx.Err())
	}
}

func (c *onlineClient) run() {
	conn, _, err := websocket.Dial(c.ctx, core.Config().Online.ServerURL, nil)
	if err != nil {
		c.report(err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")
	c.Send(protocol.TypeHello, protocol.Hello{
		ProtocolVersion: protocol.ProtocolVersion,
		PlayerID:        c.id.PlayerID,
		PlayerToken:     c.id.PlayerToken,
		DisplayName:     c.id.DisplayName,
	})
	go func() {
		for {
			var env protocol.Envelope
			if err := wsjson.Read(c.ctx, conn, &env); err != nil {
				c.report(err)
				return
			}
			select {
			case c.recv <- env:
			case <-c.ctx.Done():
				return
			}
		}
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case env := <-c.send:
			if err := wsjson.Write(c.ctx, conn, env); err != nil {
				c.report(err)
				return
			}
		}
	}
}

func (c *onlineClient) report(err error) {
	select {
	case c.errs <- err:
	default:
	}
}

func loadOnlineIdentity() onlineIdentity {
	path, err := userConfigPath()
	if err != nil {
		return onlineIdentity{DisplayName: texts().GameDefaultPlayerName}
	}
	return loadOnlineIdentityFromPath(path)
}

func loadOnlineIdentityFromPath(path string) onlineIdentity {
	cfg, err := readUserConfigFile(path)
	if err != nil || cfg.OnlineIdentity == nil {
		return onlineIdentity{DisplayName: texts().GameDefaultPlayerName}
	}
	return *cfg.OnlineIdentity
}

func saveOnlineIdentity(id onlineIdentity) {
	path, err := userConfigPath()
	if err != nil {
		return
	}
	_ = saveOnlineIdentityToPath(path, id)
}

func saveOnlineIdentityToPath(path string, id onlineIdentity) error {
	cfg, err := readUserConfigFile(path)
	if err != nil {
		cfg = userConfigFile{}
	}
	cfg.OnlineIdentity = &id
	return writeUserConfigFile(path, cfg)
}

func inviteTokenFromInput(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if u, err := url.Parse(value); err == nil && u.Path != "" {
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) >= 2 && parts[len(parts)-2] == "join" {
			return parts[len(parts)-1]
		}
	}
	if strings.HasPrefix(value, "tankblaster://join/") {
		return strings.TrimPrefix(value, "tankblaster://join/")
	}
	return value
}

func onlineNowString() string {
	return time.Now().Format("15:04:05")
}
