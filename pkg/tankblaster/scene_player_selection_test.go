package tankblaster

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/runzhammer/tankblaster/pkg/buildinfo"
	"github.com/runzhammer/tankblaster/pkg/core"
	"github.com/runzhammer/tankblaster/pkg/protocol"
	"github.com/runzhammer/tankblaster/pkg/tankblaster/computerplayers"
)

func TestColorSwatchClickOpensPaletteForComputerSlot(t *testing.T) {
	scene := &playerSelectionScene{
		focusedName:    3,
		openPaletteFor: -1,
	}
	scene.slots[0] = playerSelectionSlot{
		Kind:       PlayerComputer,
		ComputerID: computerplayers.DoedelID,
	}
	p := colorSwatchRectForSlot(0).Min.Add(image.Pt(1, 1))

	if !scene.handleSlotClick(p.X, p.Y) {
		t.Fatal("handleSlotClick returned false")
	}
	if got, want := scene.openPaletteFor, 0; got != want {
		t.Fatalf("openPaletteFor = %d, want %d", got, want)
	}
	if got, want := scene.focusedName, -1; got != want {
		t.Fatalf("focusedName = %d, want %d", got, want)
	}
	if got, want := scene.slots[0].ComputerID, computerplayers.DoedelID; got != want {
		t.Fatalf("computer id = %v, want %v", got, want)
	}
}

func TestPlayerSelectionVersionLabelUsesVPrefix(t *testing.T) {
	previous := buildinfo.Version
	defer func() {
		buildinfo.Version = previous
	}()

	buildinfo.Version = "1.0.2"
	if got, want := playerSelectionVersionLabel(), "v1.0.2"; got != want {
		t.Fatalf("version label = %q, want %q", got, want)
	}

	buildinfo.Version = "v1.0.2"
	if got, want := playerSelectionVersionLabel(), "v1.0.2"; got != want {
		t.Fatalf("version label = %q, want %q", got, want)
	}

	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "VERSION"), []byte("1.0.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tempDir)
	buildinfo.Version = "dev"
	if got, want := playerSelectionVersionLabel(), "v1.0.9 (dev)"; got != want {
		t.Fatalf("version label = %q, want %q", got, want)
	}
}

func TestPlayerSelectionOnlineClickStaysInSelectionScene(t *testing.T) {
	core.Config().Online.Enabled = true
	scene := &playerSelectionScene{onlineAvailable: true}
	p := onlineSelectionButtonRect().Min.Add(image.Pt(1, 1))

	handled, err := scene.handleOnlineClick(p.X, p.Y)

	if err != nil {
		t.Fatalf("handleOnlineClick() error = %v", err)
	}
	if !handled {
		t.Fatal("handleOnlineClick() handled = false")
	}
	if !scene.onlineMode {
		t.Fatal("online mode was not enabled")
	}
	if scene.g != nil {
		t.Fatal("test unexpectedly needs scene transition")
	}
	scene.leaveEmbeddedOnlineMode()
}

func TestVersionUpdateNoticeDoesNotCoverOnlineButton(t *testing.T) {
	bounds := image.Rect(0, 0, 960, 720)
	if versionUpdateNoticeRect(bounds).Overlaps(onlineSelectionButtonRect()) {
		t.Fatalf("version notice %v overlaps online button %v", versionUpdateNoticeRect(bounds), onlineSelectionButtonRect())
	}
}

func TestVersionUpdateCloseButtonDismissesNotice(t *testing.T) {
	scene := &playerSelectionScene{
		versionUpdate: &versionUpdateResult{available: true, latest: "9.9.9"},
	}
	p := versionUpdateCloseRect(scene.selectionBounds()).Min.Add(image.Pt(1, 1))

	if !scene.handleVersionUpdateClick(p.X, p.Y) {
		t.Fatal("close click was not handled")
	}
	if !scene.versionUpdateDismissed {
		t.Fatal("version update notice was not dismissed")
	}
}

func TestEmbeddedJoinCodeValidation(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"ABCDEFGH", true},
		{"A1B2C3D4E5", true},
		{"ABC1234", false},
		{"A1B2C3D4E5F", false},
		{"ABC-1234", false},
	}
	for _, tt := range tests {
		if got := validEmbeddedJoinCode(tt.value); got != tt.want {
			t.Fatalf("validEmbeddedJoinCode(%q) = %t, want %t", tt.value, got, tt.want)
		}
	}
}

func TestEmbeddedOnlineHostAddsHumanSlotWhenEmpty(t *testing.T) {
	scene := &playerSelectionScene{}
	for i := range scene.slots {
		scene.slots[i].Color = defaultPlayerColors[i%len(defaultPlayerColors)]
	}

	if !scene.ensureEmbeddedOnlineHostPlayer() {
		t.Fatal("ensureEmbeddedOnlineHostPlayer() = false")
	}
	if got, want := scene.slots[0].Kind, PlayerHuman; got != want {
		t.Fatalf("slot kind = %v, want %v", got, want)
	}
	if scene.slots[0].Name == "" {
		t.Fatal("slot name is empty")
	}
}

func TestEmbeddedOnlineHostStatusIsNotUnderCancelButton(t *testing.T) {
	cancel := embeddedOnlineCancelRect()
	pos := embeddedOnlineStatusTextPosition(true)

	if pos.In(cancel) {
		t.Fatalf("host status position %v overlaps cancel button %v", pos, cancel)
	}
	if pos.X <= cancel.Max.X {
		t.Fatalf("host status x = %d, want right of cancel button ending at %d", pos.X, cancel.Max.X)
	}
}

func TestEmbeddedOnlineGuestCanAddOwnHumanSlot(t *testing.T) {
	scene := &playerSelectionScene{
		onlineMode:   true,
		onlineHost:   false,
		onlineClient: &onlineClient{id: onlineIdentity{PlayerID: "guest", DisplayName: "Guest"}},
	}
	for i := range scene.slots {
		scene.slots[i].Color = defaultPlayerColors[i%len(defaultPlayerColors)]
	}

	if !scene.canEditSlotKind(0) {
		t.Fatal("guest could not edit empty slot")
	}
	scene.cycleSlotKind(0)

	if got, want := scene.slots[0].Kind, PlayerHuman; got != want {
		t.Fatalf("slot kind = %v, want %v", got, want)
	}
	if got, want := scene.slots[0].OwnerID, "guest"; got != want {
		t.Fatalf("slot owner = %q, want %q", got, want)
	}
	if got, want := scene.slots[0].Name, "Guest"; got != want {
		t.Fatalf("slot name = %q, want %q", got, want)
	}
}

func TestEmbeddedOnlineGuestCannotEditForeignSlotKind(t *testing.T) {
	scene := &playerSelectionScene{
		onlineMode:   true,
		onlineHost:   false,
		onlineClient: &onlineClient{id: onlineIdentity{PlayerID: "guest"}},
	}
	scene.slots[0] = playerSelectionSlot{Kind: PlayerHuman, OwnerID: "host"}

	if scene.canEditSlotKind(0) {
		t.Fatal("guest can edit foreign human slot")
	}
}

func TestEmbeddedOnlineGuestCanConvertOwnHumanSlotToComputer(t *testing.T) {
	scene := &playerSelectionScene{
		onlineMode:   true,
		onlineHost:   false,
		onlineClient: &onlineClient{id: onlineIdentity{PlayerID: "guest"}},
	}
	scene.slots[0] = playerSelectionSlot{Kind: PlayerHuman, OwnerID: "guest", PlayerID: "guest:slot:0", Name: "Guest"}

	if !scene.canEditSlotKind(0) {
		t.Fatal("guest could not edit own slot")
	}
	scene.cycleSlotKind(0)

	if got, want := scene.slots[0].Kind, PlayerComputer; got != want {
		t.Fatalf("slot kind = %v, want %v", got, want)
	}
	if got, want := scene.slots[0].OwnerID, "guest"; got != want {
		t.Fatalf("slot owner = %q, want %q", got, want)
	}
}

func TestEmbeddedOnlineClearedOwnedSlotSendsNoneUntilConfirmed(t *testing.T) {
	scene := &playerSelectionScene{
		onlineMode:   true,
		onlineHost:   false,
		onlineClient: &onlineClient{id: onlineIdentity{PlayerID: "guest"}},
	}
	scene.slots[0] = playerSelectionSlot{Kind: PlayerComputer, OwnerID: "guest", PlayerID: "guest:slot:0", Name: "CPU"}

	scene.cycleSlotKind(0)
	slots := scene.embeddedLobbySlots()

	if got, want := scene.slots[0].Kind, PlayerNone; got != want {
		t.Fatalf("slot kind = %v, want %v", got, want)
	}
	if got, want := len(slots), 1; got != want {
		t.Fatalf("embedded slots = %d, want %d", got, want)
	}
	if got, want := slots[0].Kind, "none"; got != want {
		t.Fatalf("embedded slot kind = %q, want %q", got, want)
	}
	if got, want := slots[0].OwnerID, "guest"; got != want {
		t.Fatalf("embedded slot owner = %q, want %q", got, want)
	}
}

func TestEmbeddedOnlineLobbyUpdateAppliesMasterRoundsAndOptions(t *testing.T) {
	scene := &playerSelectionScene{
		g:           &GameLoop{options: defaultGameOptions()},
		onlineMode:  true,
		onlineHost:  false,
		optionsOpen: true,
	}

	scene.applyEmbeddedLobbyUpdate(protocol.LobbyUpdate{
		Rounds: 7,
		Options: protocol.LobbyOptions{
			ProjectileReentry: 2,
			PalmCount:         -1,
			CloudAggression:   80,
			QuickRoundStart:   false,
		},
	})

	if got, want := scene.rounds, 7; got != want {
		t.Fatalf("rounds = %d, want %d", got, want)
	}
	if got, want := scene.g.options.projectileReentry, 2; got != want {
		t.Fatalf("projectile reentry = %d, want %d", got, want)
	}
	if got, want := scene.g.options.palmCount, -1; got != want {
		t.Fatalf("palm count = %d, want %d", got, want)
	}
	if got, want := scene.g.options.cloudAggression, 80; got != want {
		t.Fatalf("cloud aggression = %d, want %d", got, want)
	}
	if scene.g.options.quickRoundStart {
		t.Fatal("quick round start = true, want false")
	}
	if scene.optionsDraft != scene.g.options {
		t.Fatalf("options draft = %+v, want %+v", scene.optionsDraft, scene.g.options)
	}
}

func TestEmbeddedOnlinePendingHostSlotRejectsUnconfirmedLobbyUpdate(t *testing.T) {
	scene := &playerSelectionScene{
		g:               &GameLoop{options: defaultGameOptions()},
		onlineMode:      true,
		onlineHost:      true,
		onlineClient:    &onlineClient{id: onlineIdentity{PlayerID: "host"}},
		onlineSessionID: "session-1",
	}
	scene.slots[0] = playerSelectionSlot{
		Kind:     PlayerHuman,
		OwnerID:  "host",
		PlayerID: "host:slot:0",
		Name:     "Host",
		Color:    defaultPlayerColors[0],
	}
	scene.slots[1] = playerSelectionSlot{
		Kind:     PlayerHuman,
		OwnerID:  "host",
		PlayerID: "host:slot:1",
		Name:     "Spieler 2",
		Color:    defaultPlayerColors[1],
	}
	scene.onlinePendingLobby = cloneLobbyUpdate(protocol.LobbyUpdate{
		SessionID: "session-1",
		Slots:     scene.embeddedLobbySlots(),
		Rounds:    5,
		Options:   lobbyOptionsFromGameOptions(scene.g.options),
	})

	unconfirmed := protocol.LobbyUpdate{
		SessionID: "session-1",
		Revision:  2,
		Slots: []protocol.LobbySlot{
			{Index: 0, Kind: "human", OwnerID: "host", PlayerID: "host:slot:0", Name: "Host", Color: rgbaFromColor(defaultPlayerColors[0])},
		},
		Rounds:  5,
		Options: lobbyOptionsFromGameOptions(scene.g.options),
	}

	if scene.acceptEmbeddedLobbyUpdate(unconfirmed) {
		t.Fatal("unconfirmed lobby update was accepted")
	}
	if got, want := scene.slots[1].Kind, PlayerHuman; got != want {
		t.Fatalf("pending slot kind = %v, want still %v", got, want)
	}
	if scene.onlinePendingLobby == nil {
		t.Fatal("pending lobby was cleared without confirmation")
	}
}

func TestEmbeddedOnlineIgnoresStaleLobbyRevision(t *testing.T) {
	scene := &playerSelectionScene{
		g:                   &GameLoop{options: defaultGameOptions()},
		onlineMode:          true,
		onlineHost:          true,
		onlineClient:        &onlineClient{id: onlineIdentity{PlayerID: "host"}},
		onlineSessionID:     "session-1",
		onlineLobbyRevision: 4,
	}
	scene.slots[1] = playerSelectionSlot{
		Kind:     PlayerHuman,
		OwnerID:  "host",
		PlayerID: "host:slot:1",
		Name:     "Spieler 2",
		Color:    defaultPlayerColors[1],
	}

	stale := protocol.LobbyUpdate{
		SessionID: "session-1",
		Revision:  4,
		Slots: []protocol.LobbySlot{
			{Index: 0, Kind: "human", OwnerID: "host", PlayerID: "host:slot:0", Name: "Host", Color: rgbaFromColor(defaultPlayerColors[0])},
		},
	}

	if scene.acceptEmbeddedLobbyUpdate(stale) {
		t.Fatal("stale lobby update was accepted")
	}
	if got, want := scene.slots[1].Kind, PlayerHuman; got != want {
		t.Fatalf("slot kind = %v, want unchanged %v", got, want)
	}
}

func TestEmbeddedOnlinePendingGuestSlotAcceptsConfirmedLobbyUpdate(t *testing.T) {
	scene := &playerSelectionScene{
		g:               &GameLoop{options: defaultGameOptions()},
		onlineMode:      true,
		onlineHost:      false,
		onlineClient:    &onlineClient{id: onlineIdentity{PlayerID: "guest"}},
		onlineSessionID: "session-1",
	}
	pendingSlot := protocol.LobbySlot{
		Index:    2,
		Kind:     "human",
		OwnerID:  "guest",
		PlayerID: "guest:slot:2",
		Name:     "Guest 2",
		Color:    rgbaFromColor(defaultPlayerColors[2]),
	}
	scene.onlinePendingLobby = cloneLobbyUpdate(protocol.LobbyUpdate{SessionID: "session-1", Slots: []protocol.LobbySlot{pendingSlot}})

	confirmed := protocol.LobbyUpdate{
		SessionID: "session-1",
		Slots: []protocol.LobbySlot{
			{Index: 0, Kind: "human", OwnerID: "host", PlayerID: "host:slot:0", Name: "Host", Color: rgbaFromColor(defaultPlayerColors[0])},
			pendingSlot,
		},
	}

	if !scene.acceptEmbeddedLobbyUpdate(confirmed) {
		t.Fatal("confirmed lobby update was rejected")
	}
	if scene.onlinePendingLobby != nil {
		t.Fatal("pending lobby was not cleared")
	}
}

func TestEmbeddedOnlinePendingClearedSlotAcceptsMissingServerSlot(t *testing.T) {
	scene := &playerSelectionScene{
		g:               &GameLoop{options: defaultGameOptions()},
		onlineMode:      true,
		onlineHost:      false,
		onlineClient:    &onlineClient{id: onlineIdentity{PlayerID: "guest"}},
		onlineSessionID: "session-1",
	}
	scene.onlinePendingLobby = cloneLobbyUpdate(protocol.LobbyUpdate{
		SessionID: "session-1",
		Slots: []protocol.LobbySlot{
			{Index: 2, Kind: "none", OwnerID: "guest"},
		},
	})

	confirmed := protocol.LobbyUpdate{
		SessionID: "session-1",
		Slots: []protocol.LobbySlot{
			{Index: 0, Kind: "human", OwnerID: "host", PlayerID: "host:slot:0", Name: "Host", Color: rgbaFromColor(defaultPlayerColors[0])},
		},
	}

	if !scene.acceptEmbeddedLobbyUpdate(confirmed) {
		t.Fatal("cleared slot confirmation was rejected")
	}
	if scene.onlinePendingLobby != nil {
		t.Fatal("pending lobby was not cleared")
	}
}

func TestEmbeddedOnlineGuestCannotEditLobbySettings(t *testing.T) {
	scene := &playerSelectionScene{
		g:          &GameLoop{options: defaultGameOptions()},
		rounds:     5,
		onlineMode: true,
		onlineHost: false,
	}

	if scene.canEditLobbySettings() {
		t.Fatal("guest can edit lobby settings")
	}
	minus := image.Rect(434, 72, 466, 103).Min.Add(image.Pt(1, 1))
	if !scene.handleRoundsClick(minus.X, minus.Y) {
		t.Fatal("rounds click was not handled")
	}
	if got, want := scene.rounds, 5; got != want {
		t.Fatalf("rounds = %d, want unchanged %d", got, want)
	}
}

func rgbaFromColor(c color.RGBA) protocol.RGBA {
	return protocol.RGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}
