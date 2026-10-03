package tankblaster

import (
	"hash/fnv"
	"strconv"

	"github.com/runzhammer/tankblaster/pkg/core"
	"github.com/runzhammer/tankblaster/pkg/protocol"
)

func (s *GameScene) consumeOnlineGameCommands() {
	if s.g.online == nil || s.g.online.client == nil {
		return
	}
	for {
		select {
		case env := <-s.g.online.client.recv:
			s.handleOnlineGameEnvelope(env)
		case err := <-s.g.online.client.errs:
			if err != nil {
				// Keep the offline scene alive; network errors should not crash rendering.
			}
		default:
			return
		}
	}
}

func (s *GameScene) handleOnlineGameEnvelope(env protocol.Envelope) {
	switch env.Type {
	case protocol.TypeOnlineGameCommand:
		cmd, err := protocol.Decode[protocol.OnlineGameCommand](env)
		if err == nil {
			s.applyOnlineGameCommand(cmd)
		}
	case protocol.TypeStateUpdate, protocol.TypeGameStart, protocol.TypeTurnStart, protocol.TypeGameOver:
		msg, err := protocol.Decode[protocol.StateUpdate](env)
		if err == nil {
			s.g.online.state = msg.State
		}
	}
}

func (s *GameScene) sendOnlineGameCommand(cmd protocol.OnlineGameCommand) {
	if s.g.online == nil || s.g.online.client == nil {
		return
	}
	cmd.MatchID = s.g.online.state.MatchID
	cmd.PlayerID = s.g.online.playerID
	s.g.online.client.Send(protocol.TypeOnlineGameCommand, cmd)
}

func (s *GameScene) applyOnlineGameCommand(cmd protocol.OnlineGameCommand) {
	if s.g.online == nil || cmd.MatchID != s.g.online.state.MatchID {
		return
	}
	switch cmd.Kind {
	case "aim":
		s.applyOnlineAimCommand(cmd)
	case "fire":
		s.applyOnlineFireCommand(cmd)
	case "xm_v12":
		s.applyOnlineXMV12Command(cmd)
	case "turn":
		s.applyOnlineTurnCommand(cmd)
	case "shop_hover":
		s.applyOnlineShopHoverCommand(cmd)
	case "shop_mode":
		s.applyOnlineShopModeCommand(cmd)
	case "shop_select":
		s.applyOnlineShopSelectCommand(cmd)
	case "shop_buy":
		s.applyOnlineShopBuyCommand(cmd)
	case "shop_back":
		s.applyOnlineShopBackCommand(cmd)
	case "shop_continue":
		s.applyOnlineShopContinueCommand(cmd)
	}
}

func (s *GameScene) applyOnlineShopHoverCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineShopCommandAllowed(cmd) {
		return
	}
	s.shopHoverClass = maxInt(0, minInt(2, cmd.ShopHoverClass))
}

func (s *GameScene) applyOnlineShopModeCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineShopCommandAllowed(cmd) {
		return
	}
	s.withRemoteOnlineCommand(func() {
		mode := shopMode(cmd.ShopMode)
		if mode != shopModeClassA && mode != shopModeClassB {
			return
		}
		s.setShopMode(mode)
	})
}

func (s *GameScene) applyOnlineShopSelectCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineShopCommandAllowed(cmd) {
		return
	}
	s.withRemoteOnlineCommand(func() {
		s.setShopSelectedIndex(cmd.ShopListIndex)
	})
}

func (s *GameScene) applyOnlineShopBuyCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineShopCommandAllowed(cmd) {
		return
	}
	s.withRemoteOnlineCommand(func() {
		s.buySelectedShopItem()
	})
}

func (s *GameScene) applyOnlineShopBackCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineShopCommandAllowed(cmd) {
		return
	}
	s.withRemoteOnlineCommand(func() {
		s.shopMode = shopModeEntry
		s.shopHoverClass = 0
	})
}

func (s *GameScene) applyOnlineShopContinueCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineShopCommandAllowed(cmd) {
		return
	}
	s.withRemoteOnlineCommand(func() {
		s.continueShop()
	})
}

func (s *GameScene) onlineShopCommandAllowed(cmd protocol.OnlineGameCommand) bool {
	if s.g.online == nil || cmd.MatchID != s.g.online.state.MatchID || s.phase != phaseShop {
		return false
	}
	if cmd.PlayerID == s.g.online.playerID {
		s.shopAwaitingOnline = false
	}
	return s.onlinePlayerIndexForCommand(cmd) == s.currentShopPlayerIndex()
}

func (s *GameScene) withRemoteOnlineCommand(apply func()) {
	s.g.online.applyingRemoteCmd = true
	defer func() {
		s.g.online.applyingRemoteCmd = false
	}()
	apply()
}

func (s *GameScene) applyOnlineFireCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineCommandIsCurrentTurn(cmd) {
		return
	}
	s.g.online.applyingRemoteCmd = true
	defer func() {
		s.g.online.applyingRemoteCmd = false
	}()
	playerIndex := s.onlinePlayerIndexForCommand(cmd)
	if playerIndex < 0 || playerIndex >= len(s.tanks) {
		return
	}
	tank := s.tanks[playerIndex]
	if tank == nil || tank.cannon == nil {
		return
	}
	s.activePlayerIndex = playerIndex
	if cmd.WeaponSlot >= 0 && cmd.WeaponSlot < s.weaponSlotCount() {
		tank.selectedWeapon = cmd.WeaponSlot
	}
	tank.shotStrength = maxInt(0, minInt(cmd.ShotStrength, s.maxShotStrength()))
	tank.cannon.Rot = cmd.CannonRotation
	s.cameraX = s.clampOnlineCameraX(cmd.CameraX)
	s.fireActiveWeapon()
}

func (s *GameScene) applyOnlineAimCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineCommandIsCurrentTurn(cmd) {
		return
	}
	s.g.online.applyingRemoteCmd = true
	defer func() {
		s.g.online.applyingRemoteCmd = false
	}()
	tank := s.tankForOnlineCommand(cmd)
	if tank == nil || tank.cannon == nil {
		return
	}
	if s.activePlayerIndex >= 0 && tank.playerIndex != s.activePlayerIndex {
		return
	}
	if s.activePlayerIndex < 0 {
		s.activePlayerIndex = tank.playerIndex
	}
	if cmd.WeaponSlot >= 0 && cmd.WeaponSlot < s.weaponSlotCount() {
		tank.selectedWeapon = cmd.WeaponSlot
	}
	tank.shotStrength = maxInt(s.minShotStrength(), minInt(cmd.ShotStrength, s.maxShotStrength()))
	tank.cannon.Rot = cmd.CannonRotation
	s.clampCannonRotationToTank(tank.cannon, tank.body)
	s.updateXMV12FacingFromCannon(tank)
	s.cameraX = s.clampOnlineCameraX(cmd.CameraX)
}

func (s *GameScene) applyOnlineXMV12Command(cmd protocol.OnlineGameCommand) {
	if !s.onlineCommandIsCurrentTurn(cmd) {
		return
	}
	s.g.online.applyingRemoteCmd = true
	defer func() {
		s.g.online.applyingRemoteCmd = false
	}()
	tank := s.tankForOnlineCommand(cmd)
	if tank == nil {
		return
	}
	s.activePlayerIndex = tank.playerIndex
	s.xmV12DriveMode = cmd.XMV12Mode
	s.xmV12DriveDirection = cmd.XMV12Direction
	if cmd.XMV12MotorOff {
		s.playEventSound(soundEventXMV12MotorOff)
		s.stopXMV12LoopSounds()
		s.xmV12DriveMode = true
		s.xmV12DriveDirection = 0
		s.xmV12EngineOffDelay = xmV12EngineOffDelayFrames
		return
	}
	if cmd.XMV12Mode {
		s.xmV12EngineOffDelay = 0
		s.playEventSound(soundEventXMV12Ignition)
	}
}

func (s *GameScene) applyOnlineTurnCommand(cmd protocol.OnlineGameCommand) {
	if cmd.TurnSequence <= s.g.online.turnSequence {
		return
	}
	if s.onlineTurnAdvanceInProgress() {
		return
	}
	s.g.online.applyingRemoteCmd = true
	defer func() {
		s.g.online.applyingRemoteCmd = false
	}()
	playerIndex := cmd.PlayerIndex
	if playerIndex < 0 || playerIndex >= len(s.tanks) {
		return
	}
	s.g.online.turnSequence = cmd.TurnSequence
	s.activePlayerIndex = playerIndex
	s.resetComputerTurnPlans()
	s.clampActiveShotStrength()
	s.cameraGoal = s.cameraTargetForTank(s.activePlayerIndex)
	s.cameraGoalY = 0
	s.cameraX = s.clampOnlineCameraX(cmd.CameraX)
	s.cameraY = 0
}

func (s *GameScene) syncOnlineAim(tank *battleTank) {
	if s.g.online == nil || s.g.online.applyingRemoteCmd || tank == nil || tank.cannon == nil || !s.onlineCanControlPlayer(tank.playerIndex) {
		return
	}
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:           "aim",
		PlayerIndex:    tank.playerIndex,
		TurnSequence:   s.g.online.turnSequence,
		WeaponSlot:     tank.selectedWeapon,
		ShotStrength:   tank.shotStrength,
		CannonRotation: tank.cannon.Rot,
		CameraX:        s.cameraX,
	})
}

func (s *GameScene) requestXMV12Start(tank *battleTank) {
	if tank == nil {
		return
	}
	s.playEventSound(soundEventXMV12Ignition)
	s.xmV12DriveMode = true
	s.xmV12DriveDirection = 0
	s.syncOnlineXMV12(tank, false)
}

func (s *GameScene) requestXMV12Direction(tank *battleTank, direction int) {
	if tank == nil {
		return
	}
	s.xmV12DriveDirection = direction
	s.syncOnlineXMV12(tank, false)
}

func (s *GameScene) requestXMV12MotorOff(tank *battleTank) {
	if tank == nil {
		return
	}
	s.playEventSound(soundEventXMV12MotorOff)
	s.stopXMV12LoopSounds()
	s.xmV12DriveDirection = 0
	s.xmV12EngineOffDelay = xmV12EngineOffDelayFrames
	s.syncOnlineXMV12(tank, true)
}

func (s *GameScene) syncOnlineXMV12(tank *battleTank, motorOff bool) {
	if s.g.online == nil || s.g.online.applyingRemoteCmd || tank == nil || !s.onlineCanControlPlayer(tank.playerIndex) {
		return
	}
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:           "xm_v12",
		PlayerIndex:    tank.playerIndex,
		TurnSequence:   s.g.online.turnSequence,
		CameraX:        s.cameraX,
		XMV12Mode:      s.xmV12DriveMode,
		XMV12Direction: s.xmV12DriveDirection,
		XMV12MotorOff:  motorOff,
	})
}

func (s *GameScene) syncOnlineTurn(previousPlayerIndex int) {
	if s.g.online == nil || s.g.online.applyingRemoteCmd {
		return
	}
	s.g.online.turnSequence++
	if !s.onlineCanControlPlayer(previousPlayerIndex) {
		return
	}
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:         "turn",
		PlayerIndex:  s.activePlayerIndex,
		TurnSequence: s.g.online.turnSequence,
		CameraX:      s.cameraX,
	})
}

func (s *GameScene) syncOnlineShopHover(hoverClass int) {
	if s.g.online == nil || s.g.online.applyingRemoteCmd || !s.onlineCanControlPlayer(s.currentShopPlayerIndex()) {
		return
	}
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:           "shop_hover",
		PlayerIndex:    s.currentShopPlayerIndex(),
		ShopHoverClass: maxInt(0, minInt(2, hoverClass)),
	})
}

func (s *GameScene) syncOnlineShopMode(mode shopMode) {
	if !s.canSendOnlineShopCommand() {
		return
	}
	s.shopAwaitingOnline = true
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:        "shop_mode",
		PlayerIndex: s.currentShopPlayerIndex(),
		ShopMode:    int(mode),
	})
}

func (s *GameScene) syncOnlineShopSelect(index int) {
	if !s.canSendOnlineShopCommand() {
		return
	}
	s.shopAwaitingOnline = true
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:          "shop_select",
		PlayerIndex:   s.currentShopPlayerIndex(),
		ShopListIndex: index,
	})
}

func (s *GameScene) syncOnlineShopBuy() {
	if !s.canSendOnlineShopCommand() {
		return
	}
	s.shopAwaitingOnline = true
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:        "shop_buy",
		PlayerIndex: s.currentShopPlayerIndex(),
	})
}

func (s *GameScene) syncOnlineShopBack() {
	if !s.canSendOnlineShopCommand() {
		return
	}
	s.shopAwaitingOnline = true
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:        "shop_back",
		PlayerIndex: s.currentShopPlayerIndex(),
	})
}

func (s *GameScene) syncOnlineShopContinue() {
	if !s.canSendOnlineShopCommand() {
		return
	}
	s.shopAwaitingOnline = true
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:        "shop_continue",
		PlayerIndex: s.currentShopPlayerIndex(),
	})
}

func (s *GameScene) canSendOnlineShopCommand() bool {
	return s.g.online != nil &&
		!s.g.online.applyingRemoteCmd &&
		!s.shopAwaitingOnline &&
		s.phase == phaseShop &&
		s.onlineCanControlPlayer(s.currentShopPlayerIndex())
}

func (s *GameScene) onlineCommandIsCurrentTurn(cmd protocol.OnlineGameCommand) bool {
	return s.g.online == nil || cmd.TurnSequence == s.g.online.turnSequence
}

func (s *GameScene) onlineTurnAdvanceInProgress() bool {
	return s.projectilesActive() || s.turnAdvanceDelay > 0 || s.turnAdvanceBlocked() || s.xmV12DriveMode
}

func (s *GameScene) clampOnlineCameraX(cameraX float64) float64 {
	maxCameraX := s.worldWidth - core.Config().Screen.Width
	if maxCameraX < 0 {
		maxCameraX = 0
	}
	return maxFloat64(0, minFloat(cameraX, maxCameraX))
}

func (s *GameScene) onlineComputerDecisionSeed(tank *battleTank) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(strconv.FormatInt(s.g.online.state.Seed, 10)))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(s.roundNumber)))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(s.activePlayerIndex)))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(int(s.time / 30))))
	if tank != nil {
		_, _ = hasher.Write([]byte(":"))
		_, _ = hasher.Write([]byte(tank.player.Name))
	}
	return int64(hasher.Sum64())
}

func (s *GameScene) reportOnlineMatchComplete() {
	if s.g.online == nil || s.g.online.client == nil || s.g.online.matchResultSent {
		return
	}
	winnerIndex, ok := s.onlineWinnerByScore()
	if !ok || winnerIndex < 0 || winnerIndex >= len(s.g.online.state.Players) {
		return
	}
	if s.g.online.state.Players[winnerIndex].ID != s.g.online.playerID {
		return
	}
	scores := make(map[string]int, len(s.g.online.state.Players))
	for index, player := range s.g.online.state.Players {
		scores[player.ID] = s.scoreForPlayer(index)
	}
	s.g.online.matchResultSent = true
	s.g.online.client.Send(protocol.TypeMatchComplete, protocol.MatchComplete{
		MatchID: s.g.online.state.MatchID,
		Scores:  scores,
	})
}

func (s *GameScene) onlineWinnerByScore() (int, bool) {
	if s.g.online == nil || len(s.g.online.state.Players) == 0 {
		return -1, false
	}
	winnerIndex := -1
	winnerScore := 0
	tied := false
	for index := range s.g.online.state.Players {
		score := s.scoreForPlayer(index)
		if winnerIndex < 0 || score > winnerScore {
			winnerIndex = index
			winnerScore = score
			tied = false
			continue
		}
		if score == winnerScore {
			tied = true
		}
	}
	return winnerIndex, winnerIndex >= 0 && !tied
}

func (s *GameScene) tankForOnlineCommand(cmd protocol.OnlineGameCommand) *battleTank {
	playerIndex := s.onlinePlayerIndexForCommand(cmd)
	if playerIndex < 0 || playerIndex >= len(s.tanks) {
		return nil
	}
	return s.tanks[playerIndex]
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func (s *GameScene) onlinePlayerIndexForCommand(cmd protocol.OnlineGameCommand) int {
	if cmd.PlayerIndex >= 0 &&
		cmd.PlayerIndex < len(s.g.online.state.Players) &&
		s.g.online.state.Players[cmd.PlayerIndex].ID == cmd.PlayerID {
		return cmd.PlayerIndex
	}
	for index, player := range s.g.online.state.Players {
		if player.ID == cmd.PlayerID {
			return index
		}
	}
	return -1
}
