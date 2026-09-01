package tankblaster

import (
	"strings"
	"testing"
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

	got := truncateRunes(input, onlineJoinInputMaxRunes)
	if got != input {
		t.Fatalf("truncateRunes() shortened a normal invite link to %q", got)
	}
}
