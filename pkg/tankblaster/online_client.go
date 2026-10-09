package tankblaster

import (
	"context"
	"errors"
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
	ctx               context.Context
	cancel            context.CancelFunc
	serverURL         string
	keepAliveInterval time.Duration
	pingTimeout       time.Duration
	reconnectMinDelay time.Duration
	reconnectMaxDelay time.Duration
	send              chan protocol.Envelope
	recv              chan protocol.Envelope
	errs              chan error
	id                onlineIdentity
}

func newOnlineClient(displayName string) *onlineClient {
	return newOnlineClientWithConfig(displayName, onlineClientConfig{
		serverURL:         core.Config().Online.ServerURL,
		keepAliveInterval: 10 * time.Second,
		pingTimeout:       4 * time.Second,
		reconnectMinDelay: 300 * time.Millisecond,
		reconnectMaxDelay: 5 * time.Second,
	})
}

type onlineClientConfig struct {
	serverURL         string
	keepAliveInterval time.Duration
	pingTimeout       time.Duration
	reconnectMinDelay time.Duration
	reconnectMaxDelay time.Duration
}

func newOnlineClientWithConfig(displayName string, cfg onlineClientConfig) *onlineClient {
	ctx, cancel := context.WithCancel(context.Background())
	if cfg.serverURL == "" {
		cfg.serverURL = core.Config().Online.ServerURL
	}
	if cfg.keepAliveInterval <= 0 {
		cfg.keepAliveInterval = 10 * time.Second
	}
	if cfg.pingTimeout <= 0 {
		cfg.pingTimeout = 4 * time.Second
	}
	if cfg.reconnectMinDelay <= 0 {
		cfg.reconnectMinDelay = 300 * time.Millisecond
	}
	if cfg.reconnectMaxDelay < cfg.reconnectMinDelay {
		cfg.reconnectMaxDelay = cfg.reconnectMinDelay
	}
	c := &onlineClient{
		ctx:               ctx,
		cancel:            cancel,
		serverURL:         cfg.serverURL,
		keepAliveInterval: cfg.keepAliveInterval,
		pingTimeout:       cfg.pingTimeout,
		reconnectMinDelay: cfg.reconnectMinDelay,
		reconnectMaxDelay: cfg.reconnectMaxDelay,
		send:              make(chan protocol.Envelope, 16),
		recv:              make(chan protocol.Envelope, 32),
		errs:              make(chan error, 4),
		id:                loadOnlineIdentity(),
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
	delay := c.reconnectMinDelay
	for {
		if err := c.connectAndRun(); err != nil && !errors.Is(err, context.Canceled) {
			c.report(errors.New("Online-Verbindung unterbrochen, verbinde neu..."))
		}
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(delay):
		}
		delay *= 2
		if delay > c.reconnectMaxDelay {
			delay = c.reconnectMaxDelay
		}
	}
}

func (c *onlineClient) connectAndRun() error {
	conn, _, err := websocket.Dial(c.ctx, c.serverURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "")
	hello, err := protocol.Wrap(protocol.TypeHello, protocol.Hello{
		ProtocolVersion: protocol.ProtocolVersion,
		PlayerID:        c.id.PlayerID,
		PlayerToken:     c.id.PlayerToken,
		DisplayName:     c.id.DisplayName,
	})
	if err != nil {
		return err
	}
	if err := wsjson.Write(c.ctx, conn, hello); err != nil {
		return err
	}
	connCtx, cancel := context.WithCancel(c.ctx)
	defer cancel()
	errs := make(chan error, 2)
	go func() {
		for {
			var env protocol.Envelope
			if err := wsjson.Read(connCtx, conn, &env); err != nil {
				errs <- err
				return
			}
			select {
			case c.recv <- env:
			case <-connCtx.Done():
				return
			}
		}
	}()
	go c.keepAlive(connCtx, conn, errs)
	for {
		select {
		case <-connCtx.Done():
			return connCtx.Err()
		case err := <-errs:
			return err
		case env := <-c.send:
			if err := wsjson.Write(connCtx, conn, env); err != nil {
				return err
			}
		}
	}
}

func (c *onlineClient) keepAlive(ctx context.Context, conn *websocket.Conn, errs chan<- error) {
	ticker := time.NewTicker(c.keepAliveInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, c.pingTimeout)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				select {
				case errs <- err:
				case <-ctx.Done():
				}
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

func truncateRunes(value string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
