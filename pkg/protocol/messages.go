package protocol

import (
	"encoding/json"
	"time"

	"github.com/runzhammer/gamedemo/pkg/gamecore"
)

const ProtocolVersion = 1

type MessageType string

const (
	TypeHello                MessageType = "hello"
	TypeQuickMatch           MessageType = "quick_match"
	TypeCancelQuickMatch     MessageType = "cancel_quick_match"
	TypeCreatePublicSession  MessageType = "create_public_session"
	TypeCreatePrivateSession MessageType = "create_private_session"
	TypeListSessions         MessageType = "list_sessions"
	TypeJoinSession          MessageType = "join_session"
	TypeLeaveSession         MessageType = "leave_session"
	TypeReady                MessageType = "ready"
	TypeFire                 MessageType = "fire"
	TypeReconnect            MessageType = "reconnect"
	TypeGetLeaderboard       MessageType = "get_leaderboard"
	TypePing                 MessageType = "ping"

	TypeHelloAck             MessageType = "hello_ack"
	TypeMatchmakingQueued    MessageType = "matchmaking_queued"
	TypeMatchmakingCancelled MessageType = "matchmaking_cancelled"
	TypeSessionCreated       MessageType = "session_created"
	TypeSessionList          MessageType = "session_list"
	TypeSessionJoined        MessageType = "session_joined"
	TypeSessionUpdated       MessageType = "session_updated"
	TypeSessionClosed        MessageType = "session_closed"
	TypeInviteCreated        MessageType = "invite_created"
	TypeMatchFound           MessageType = "match_found"
	TypeGameStart            MessageType = "game_start"
	TypeTurnStart            MessageType = "turn_start"
	TypeShotResult           MessageType = "shot_result"
	TypeStateUpdate          MessageType = "state_update"
	TypeLeaderboard          MessageType = "leaderboard"
	TypePlayerDisconnected   MessageType = "player_disconnected"
	TypePlayerReconnected    MessageType = "player_reconnected"
	TypeGameOver             MessageType = "game_over"
	TypeError                MessageType = "error"
	TypePong                 MessageType = "pong"
)

type Envelope struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

func Wrap[T any](typ MessageType, payload T) (Envelope, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{Type: typ, Data: data}, nil
}

func Decode[T any](env Envelope) (T, error) {
	var out T
	err := json.Unmarshal(env.Data, &out)
	return out, err
}

type Hello struct {
	ProtocolVersion int    `json:"protocol_version"`
	PlayerID        string `json:"player_id,omitempty"`
	PlayerToken     string `json:"player_token,omitempty"`
	DisplayName     string `json:"display_name,omitempty"`
}

type HelloAck struct {
	PlayerID    string `json:"player_id"`
	PlayerToken string `json:"player_token,omitempty"`
	DisplayName string `json:"display_name"`
	Rating      int    `json:"rating"`
	Score       int    `json:"score"`
}

type CreateSession struct {
	DisplayName string `json:"display_name,omitempty"`
}

type JoinSession struct {
	SessionID   string `json:"session_id,omitempty"`
	JoinCode    string `json:"join_code,omitempty"`
	InviteToken string `json:"invite_token,omitempty"`
}

type Ready struct {
	SessionID string `json:"session_id"`
	Ready     bool   `json:"ready"`
}

type FireCommand struct {
	MatchID string  `json:"match_id"`
	Weapon  string  `json:"weapon"`
	Angle   float64 `json:"angle"`
	Power   float64 `json:"power"`
}

type Reconnect struct {
	MatchID string `json:"match_id"`
}

type SessionSummary struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	HostName      string    `json:"host_name"`
	PlayerCount   int       `json:"player_count"`
	MaxPlayers    int       `json:"max_players"`
	AverageRating int       `json:"average_rating"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type SessionList struct {
	Sessions []SessionSummary `json:"sessions"`
}

type SessionCreated struct {
	Session SessionSummary `json:"session"`
}

type InviteCreated struct {
	SessionID string    `json:"session_id"`
	JoinCode  string    `json:"join_code"`
	InviteURL string    `json:"invite_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SessionJoined struct {
	Session SessionSummary `json:"session"`
}

type MatchFound struct {
	SessionID string `json:"session_id"`
	MatchID   string `json:"match_id"`
}

type StateUpdate struct {
	State gamecore.MatchState `json:"state"`
}

type ShotResult struct {
	Result gamecore.ShotResult `json:"result"`
	State  gamecore.MatchState `json:"state"`
}

type LeaderboardRequest struct {
	Limit int `json:"limit,omitempty"`
}

type LeaderboardEntry struct {
	Rank          int    `json:"rank"`
	PlayerID      string `json:"player_id"`
	DisplayName   string `json:"display_name"`
	Score         int    `json:"score"`
	Rating        int    `json:"rating"`
	Wins          int    `json:"wins"`
	Losses        int    `json:"losses"`
	MatchesPlayed int    `json:"matches_played"`
}

type Leaderboard struct {
	Entries []LeaderboardEntry `json:"entries"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
